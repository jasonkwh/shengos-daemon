package bus

import (
	"github.com/eclipse/paho.golang/paho"
	"github.com/jasonkwh/shengos-daemon/internal/message"
)

func toPahoSubscribeOptions(
	subscription message.Subscription,
) paho.SubscribeOptions {
	return paho.SubscribeOptions{
		Topic:             subscription.Topic,
		QoS:               byte(subscription.QoS),
		NoLocal:           subscription.NoLocal,
		RetainAsPublished: subscription.RetainAsPublished,
		RetainHandling:    byte(subscription.RetainHandling),
	}
}
