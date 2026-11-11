package cmd

import (
	"io"
	"log"

	"github.com/jasonkwh/envconfig"
	"github.com/jasonkwh/perf"
	"github.com/jasonkwh/shengos-daemon/config"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

var (
	rootCmd = &cobra.Command{
		Use:   "serve",
		Short: "starts the service",
	}
	cfgFile string
	cfg     config.Config
)

func init() {
	cfgr := envconfig.New(&cfgFile, &cfg)
	cobra.OnInitialize(cfgr.Load)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "config.yaml", "config file")
}

func initPProfServer(zl *zap.Logger) []io.Closer {
	var clPool []io.Closer

	if cfg.PProfEnabled {
		pprofSrv, perr := perf.NewPProfServer(zl)
		if perr != nil {
			zl.Fatal("failed to create pprof server", zap.Error(perr))
		}

		clPool = append(clPool, pprofSrv)
	}

	return clPool
}

func gracefulClose(services []io.Closer) error {
	var errs error

	for i := len(services) - 1; i >= 0; i-- {
		if err := services[i].Close(); err != nil {
			errs = multierr.Append(errs, err)
		}
	}

	return errs
}
