package cmd

import (
	"context"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jasonkwh/perf"
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

	// TODO: add other initializations here

	zl.Info("daemon started")

	<-ctx.Done()

	err = gracefulClose(clPool)
	if err != nil {
		log.Fatalf("failed to close the daemon: %v", err)
	}
}
