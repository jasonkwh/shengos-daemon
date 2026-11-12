package bus

import (
	"fmt"
	"os"

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

func clientID() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s-%s", clientIDPrefix, hostname), nil
}
