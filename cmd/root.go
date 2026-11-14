package cmd

import (
	"io"
	"log"
	"slices"

	"github.com/jasonkwh/envconfig"
	"github.com/jasonkwh/shengos-daemon/config"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

func initZapLogger() (*zap.Logger, error) {
	zlCfg := zap.NewProductionConfig()

	// set the internal logger to INFO because we need all internal logs
	zlCfg.Level.SetLevel(zapcore.Level(cfg.LogLevel))

	return zlCfg.Build()
}

func gracefulClose(services []io.Closer) error {
	var errs error

	for _, svc := range slices.Backward(services) {
		if err := svc.Close(); err != nil {
			errs = multierr.Append(errs, err)
		}
	}

	return errs
}
