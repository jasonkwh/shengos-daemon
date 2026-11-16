package adapter

import (
	"context"

	"github.com/eclipse/paho.golang/paho"
)

type MQTTClient interface {
	Publish(
		ctx context.Context,
		p *paho.Publish,
	) (*paho.PublishResponse, error)

	Subscribe(
		ctx context.Context,
		s *paho.Subscribe,
	) (*paho.Suback, error)

	Disconnect(ctx context.Context) error
}

type MQTTPublishAcker interface {
	Ack(pb *paho.Publish) error
}
