package bus

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/packets"
	"github.com/eclipse/paho.golang/paho"
	"github.com/jasonkwh/shengos-daemon/config"
	"github.com/jasonkwh/shengos-daemon/internal/adapter"
	"github.com/jasonkwh/shengos-daemon/internal/message"
	"go.uber.org/zap"
)

var _ adapter.DaemonBus = (*daemonBus)(nil)

type daemonBus struct {
	mqttCl       adapter.MQTTClient
	publishAcker adapter.MQTTPublishAcker
	zl           *zap.Logger
	hostname     string
	subOpts      []paho.SubscribeOptions
}

func NewDaemonBus(
	ctx context.Context,
	cfg config.MQTTConnConfig,
	hostname string,
	topicHdlrs map[string]adapter.DaemonTopicHandler,
	subs []message.Subscription,
	opts ...daemonBusOption,
) (adapter.DaemonBus, error) {
	dbus := &daemonBus{
		zl:       zap.NewNop(),
		hostname: hostname,
		subOpts:  ToPahoSubscribeOptions(subs),
	}

	for _, opt := range opts {
		opt(dbus)
	}

	if dbus.mqttCl != nil {
		return dbus, nil
	}

	if err := dbus.initMQTTClient(ctx, cfg, topicHdlrs); err != nil {
		return nil, err
	}

	return dbus, nil
}

func (b *daemonBus) Publish(
	ctx context.Context,
	message adapter.DaemonMessage,
) error {
	b.zl.Debug("publishing message",
		zap.String("topic", message.Topic()))

	payload, err := message.Payload()
	if err != nil {
		b.zl.Error("failed to get payload",
			zap.String("topic", message.Topic()),
			zap.Error(err))

		return errPublishFailed(err)
	}

	if _, err := b.mqttCl.Publish(ctx, &paho.Publish{
		QoS:     message.QoS(),
		Retain:  message.Retain(),
		Topic:   message.Topic(),
		Payload: payload,
	}); err != nil {
		b.zl.Error("failed to publish message",
			zap.String("topic", message.Topic()),
			zap.Error(err))

		return errPublishFailed(err)
	}

	b.zl.Debug("published message",
		zap.String("topic", message.Topic()))

	return nil
}

func (b *daemonBus) Subscribe(ctx context.Context) error {
	if err := b.subscribeTopics(ctx); err != nil {
		b.zl.Error("failed to subscribe to mqtt broker", zap.Error(err))

		return err
	}

	return nil
}

func (b *daemonBus) Close() error {
	if b.mqttCl != nil {
		if err := b.mqttCl.Disconnect(context.Background()); err != nil {
			b.zl.Error("failed to disconnect from MQTT broker", zap.Error(err))

			return errCloseFailed(err)
		}
	}

	b.zl.Info("disconnected from MQTT broker")

	return nil
}

func (b *daemonBus) initMQTTClient(
	ctx context.Context,
	cfg config.MQTTConnConfig,
	topicHdlrs map[string]adapter.DaemonTopicHandler,
) error {
	srvURL := net.JoinHostPort(cfg.Host, cfg.Port)

	autoCfg := autopaho.ClientConfig{
		ServerUrls: []*url.URL{
			{
				Scheme: "mqtt",
				Host:   srvURL,
			},
		},
		KeepAlive:                     cfg.KeepAlive,
		CleanStartOnInitialConnection: true,
		SessionExpiryInterval:         sessionExpiryInterval,
		OnConnectionUp: func(cm *autopaho.ConnectionManager, _ *paho.Connack) {
			b.zl.Info("connected to mqtt broker",
				zap.String("url", srvURL))

			if serr := b.subscribeTopics(ctx); serr != nil {
				b.zl.Error("failed to subscribe to mqtt broker", zap.Error(serr))
			}
		},
		OnConnectError: func(err error) {
			b.zl.Error("mqtt connect failed", zap.Error(err))
		},
		ClientID:                   b.clientID(),
		EnableManualAcknowledgment: true,
		OnClientError: func(err error) {
			b.zl.Error("mqtt client error", zap.Error(err))
		},
		OnPublishReceived: []func(paho.PublishReceived) (bool, error){
			b.onPublishReceived(ctx, topicHdlrs),
		},
	}

	if cfg.Username != nil {
		autoCfg.ConnectUsername = *cfg.Username
	}

	if cfg.Password != nil {
		autoCfg.ConnectPassword = []byte(*cfg.Password)
	}

	cm, err := autopaho.NewConnection(ctx, autoCfg)
	if err != nil {
		return errConnectFailed(err)
	}

	b.mqttCl = cm

	if err := cm.AwaitConnection(ctx); err != nil {
		b.zl.Error("failed to await connection", zap.Error(err))

		if derr := cm.Disconnect(context.Background()); derr != nil {
			b.zl.Error("failed to disconnect from MQTT broker",
				zap.Error(derr))
		}

		return errConnectFailed(err)
	}

	return nil
}

func (b *daemonBus) clientID() string {
	return fmt.Sprintf("%s-%s", clientIDPrefix, b.hostname)
}

func (b *daemonBus) subscribeTopics(
	ctx context.Context,
) error {
	if len(b.subOpts) == 0 {
		b.zl.Debug("no subscriptions, skipping")

		return nil
	}

	b.zl.Debug("subscribing to mqtt broker")

	sub, err := b.mqttCl.Subscribe(ctx, &paho.Subscribe{Subscriptions: b.subOpts})
	if err != nil {
		return errSubscribeFailed(err)
	}

	if len(sub.Reasons) != len(b.subOpts) {
		return errSubscribeRejected(fmt.Sprintf("%d reasons for %d subscriptions",
			len(sub.Reasons), len(b.subOpts)))
	}

	rejected := make([]string, 0)

	for index, opt := range b.subOpts {
		if index >= len(sub.Reasons) {
			continue
		}

		reason := sub.Reasons[index]
		if reason >= packets.SubackUnspecifiederror {
			rejected = append(rejected, fmt.Sprintf("%s: 0x%02x", opt.Topic, reason))
		}
	}

	if len(rejected) > 0 {
		return errSubscribeRejected(strings.Join(rejected, ", "))
	}

	b.zl.Debug("subscribed", zap.Int("topic_count", len(b.subOpts)))

	return nil
}

func (b *daemonBus) onPublishReceived(
	ctx context.Context,
	topicHdlrs map[string]adapter.DaemonTopicHandler,
) func(paho.PublishReceived) (bool, error) {
	return func(pr paho.PublishReceived) (bool, error) {
		hdl, ok := topicHdlrs[pr.Packet.Topic]
		if !ok {
			err := errUnknownTopic(pr.Packet.Topic)
			b.zl.Error("handler not found",
				zap.String("topic", pr.Packet.Topic),
				zap.Error(err))

			return true, err
		}

		if err := hdl.Handle(ctx, pr.Packet.Payload); err != nil {
			b.zl.Error("handler failed",
				zap.String("topic", pr.Packet.Topic),
				zap.Error(err))

			return true, err
		}

		if err := b.ackMessage(pr); err != nil {
			b.zl.Error("failed to ack message",
				zap.String("topic", pr.Packet.Topic),
				zap.Error(err))

			return true, err
		}

		return true, nil
	}
}

func (b *daemonBus) ackMessage(pr paho.PublishReceived) error {
	if b.publishAcker != nil {
		// for unit-testing purposes
		return b.publishAcker.Ack(pr.Packet)
	}

	return pr.Client.Ack(pr.Packet)
}
