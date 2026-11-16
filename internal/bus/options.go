package bus

import (
	"github.com/jasonkwh/shengos-daemon/internal/adapter"
	"go.uber.org/zap"
)

type daemonBusOption func(*daemonBus)

func WithLogger(zl *zap.Logger) daemonBusOption {
	return func(b *daemonBus) {
		b.zl = zl
	}
}

func WithMQTTClient(mqttCl adapter.MQTTClient) daemonBusOption {
	return func(b *daemonBus) {
		b.mqttCl = mqttCl
	}
}
