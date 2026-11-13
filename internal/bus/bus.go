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
	mqttCl   *autopaho.ConnectionManager
	zl       *zap.Logger
	hostname string
}

func NewDaemonBus(
	ctx context.Context,
	cfg config.MQTTConnConfig,
	hdl adapter.DaemonMessageHandler,
	hostname string,
	opts ...daemonBusOption,
) (adapter.DaemonBus, error) {
	dbus := &daemonBus{
		zl:       zap.NewNop(),
		hostname: hostname,
	}

	for _, opt := range opts {
		opt(dbus)
	}

	if err := dbus.initMQTTClient(ctx, cfg, hdl); err != nil {
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
	if err := b.subscribeTopics(ctx, b.mqttCl); err != nil {
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
	hdl adapter.DaemonMessageHandler,
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
			if serr := b.subscribeTopics(ctx, cm); serr != nil {
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
			b.onPublishReceived(ctx, hdl),
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

	if err := cm.AwaitConnection(ctx); err != nil {
		return errConnectFailed(err)
	}

	b.zl.Info("connected to mqtt broker",
		zap.String("url", srvURL))

	b.mqttCl = cm

	return nil
}

func (b *daemonBus) clientID() string {
	return fmt.Sprintf("%s-%s", clientIDPrefix, b.hostname)
}

func (b *daemonBus) subscribeTopics(
	ctx context.Context,
	cm *autopaho.ConnectionManager,
) error {
	b.zl.Debug("subscribing to mqtt broker")

	subs := []paho.SubscribeOptions{
		toPahoSubscribeOptions(message.ClipboardSubscription),
	}

	sub, err := cm.Subscribe(ctx, &paho.Subscribe{Subscriptions: subs})
	if err != nil {
		return errSubscribeFailed(err)
	}

	if len(sub.Reasons) != len(subs) {
		return errSubscribeRejected(fmt.Sprintf("%d reasons for %d subscriptions", len(sub.Reasons), len(subs)))
	}

	rejected := make([]string, 0)

	for i, opt := range subs {
		if i >= len(sub.Reasons) {
			continue
		}

		reason := sub.Reasons[i]
		if reason >= packets.SubackUnspecifiederror {
			rejected = append(rejected, fmt.Sprintf("%s: 0x%02x", opt.Topic, reason))
		}
	}

	if len(rejected) > 0 {
		return errSubscribeRejected(strings.Join(rejected, ", "))
	}

	b.zl.Debug("subscribed", zap.Int("topic_count", len(subs)))

	return nil
}

func (b *daemonBus) onPublishReceived(
	ctx context.Context,
	hdl adapter.DaemonMessageHandler,
) func(paho.PublishReceived) (bool, error) {
	return func(pr paho.PublishReceived) (bool, error) {
		if err := hdl.ProcessMessage(ctx, pr.Packet.Topic, pr.Packet.Payload); err != nil {
			b.zl.Error("handler failed",
				zap.String("topic", pr.Packet.Topic),
				zap.Error(err))

			return true, err
		}

		if err := pr.Client.Ack(pr.Packet); err != nil {
			b.zl.Error("failed to ack message",
				zap.String("topic", pr.Packet.Topic),
				zap.Error(err))

			return true, err
		}

		return true, nil
	}
}
