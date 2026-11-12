package bus

import "go.uber.org/zap"

type daemonBusOption func(*daemonBus)

func WithLogger(zl *zap.Logger) daemonBusOption {
	return func(b *daemonBus) {
		b.zl = zl
	}
}
