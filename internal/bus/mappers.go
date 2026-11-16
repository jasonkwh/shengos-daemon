package bus

import (
	"github.com/eclipse/paho.golang/paho"
	"github.com/jasonkwh/shengos-daemon/internal/message"
)

func toPahoSubscribeOption(
	sub message.Subscription,
) paho.SubscribeOptions {
	return paho.SubscribeOptions{
		Topic:             sub.Topic,
		QoS:               byte(sub.QoS),
		NoLocal:           sub.NoLocal,
		RetainAsPublished: sub.RetainAsPublished,
		RetainHandling:    byte(sub.RetainHandling),
	}
}

func ToPahoSubscribeOptions(
	subs []message.Subscription,
) []paho.SubscribeOptions {
	opts := make([]paho.SubscribeOptions, len(subs))

	for index, sub := range subs {
		opts[index] = toPahoSubscribeOption(sub)
	}

	return opts
}
