package cmd

import (
	"context"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jasonkwh/perf"
	"github.com/jasonkwh/shengos-daemon/internal/bus"
	"github.com/jasonkwh/shengos-daemon/internal/clipboard"
	"github.com/jasonkwh/shengos-daemon/internal/handler"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "execute serve command",
	Run:   serve,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func serve(cmd *cobra.Command, args []string) {
	zl, err := initZapLogger()
	if err != nil {
		log.Fatal("unable to start zap logger")
	}

	var clPool []io.Closer

	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	go func(logger *zap.Logger) {
		<-c
		logger.Info("cancelling")

		cancel()
	}(zl)

	if cfg.PProfEnabled {
		pprofSrv, perr := perf.NewPProfServer(zl)
		if perr != nil {
			zl.Fatal("failed to create pprof server", zap.Error(perr))
		}

		clPool = append(clPool, pprofSrv)
	}

	hostname, err := os.Hostname()
	if err != nil {
		zl.Fatal("failed to get hostname", zap.Error(err))
	}

	clb, err := clipboard.New(
		cfg.Clipboard,
		hostname,
		clipboard.WithLogger(zl))
	if err != nil {
		zl.Fatal("failed to create clipboard", zap.Error(err))
	}

	dbus, err := bus.NewDaemonBus(ctx,
		cfg.MQTTConn,
		handler.NewDaemonMessageHandler(clb,
			handler.WithMessageLogger(zl)),
		hostname,
		bus.WithLogger(zl))
	if err != nil {
		zl.Fatal("failed to create daemon bus", zap.Error(err))
	}

	clPool = append(clPool, dbus)

	go func() {
		if cerr := clb.Watch(ctx, dbus.Publish); cerr != nil {
			zl.Fatal("failed to watch clipboard", zap.Error(cerr))
		}
	}()

	zl.Info("daemon started")

	<-ctx.Done()

	err = gracefulClose(clPool)
	if err != nil {
		log.Fatalf("failed to close the daemon: %v", err)
	}
}
