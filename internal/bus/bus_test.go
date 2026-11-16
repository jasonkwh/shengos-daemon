package bus_test

import (
	"context"
	"errors"
	"testing"

	"github.com/eclipse/paho.golang/packets"
	"github.com/eclipse/paho.golang/paho"
	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
	"github.com/jasonkwh/shengos-daemon/config"
	"github.com/jasonkwh/shengos-daemon/internal/adapter"
	"github.com/jasonkwh/shengos-daemon/internal/bus"
	"github.com/jasonkwh/shengos-daemon/internal/message"
	mock_adapter "github.com/jasonkwh/shengos-daemon/mocks/adapter"
)

var sampleMQTTConnConfig = config.MQTTConnConfig{
	Host:      "localhost",
	Port:      "1883",
	KeepAlive: 10,
}

const sampleHostname = "jasonkwh-test-rig"

func Test_daemonBus_Publish(t *testing.T) {
	subs := []message.Subscription{message.ClipboardSubscription}

	sampleMessage := &message.ClipboardMessage{
		Origin: sampleHostname,
		Format: message.ClipboardFormatText,
		Data:   []byte("hello"),
	}

	sampleMessagePayload, err := sampleMessage.Payload()
	if err != nil {
		t.Fatalf("failed to get sample message payload: %v", err)
	}

	tests := []struct {
		name       string
		publishErr error
	}{
		{
			name: "success",
		},
		{
			name:       "publish error",
			publishErr: errors.New("publish error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			as := assert.New(t)
			rs := require.New(t)
			ctx := t.Context()

			dbus, tm, err := newMockDaemonBus(t, ctx, subs)
			rs.NoError(err)

			tm.mockMQTTClient.EXPECT().Publish(ctx, &paho.Publish{
				QoS:     sampleMessage.QoS(),
				Retain:  sampleMessage.Retain(),
				Topic:   sampleMessage.Topic(),
				Payload: sampleMessagePayload,
			}).Return(&paho.PublishResponse{}, tt.publishErr)

			err = dbus.Publish(ctx, sampleMessage)
			if tt.publishErr != nil {
				as.ErrorIs(err, bus.ErrPublishFailed)
				as.ErrorIs(err, tt.publishErr)
				return
			}

			as.NoError(err)
		})
	}
}

func Test_daemonBus_Subscribe(t *testing.T) {
	tests := []struct {
		name         string
		reasons      []byte
		subscribeErr error
		wantErr      error
		wantDetail   string
		subs         []message.Subscription
	}{
		{
			name:    "success",
			reasons: []byte{packets.SubackGrantedQoS1},
			subs:    []message.Subscription{message.ClipboardSubscription},
		},
		{
			name:         "subscribe error",
			subscribeErr: errors.New("subscribe error"),
			wantErr:      bus.ErrSubscribeFailed,
			subs:         []message.Subscription{message.ClipboardSubscription},
		},
		{
			name:       "reason count mismatch",
			wantErr:    bus.ErrSubscribeRejected,
			wantDetail: "0 reasons for 1 subscriptions",
			subs:       []message.Subscription{message.ClipboardSubscription},
		},
		{
			name:       "subscription rejected",
			reasons:    []byte{packets.SubackNotauthorized},
			wantErr:    bus.ErrSubscribeRejected,
			wantDetail: message.ClipboardTopic + ": 0x87",
			subs:       []message.Subscription{message.ClipboardSubscription},
		},
		{
			name: "no subscriptions",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			as := assert.New(t)
			rs := require.New(t)
			ctx := t.Context()

			dbus, tm, err := newMockDaemonBus(t, ctx, tt.subs)
			rs.NoError(err)

			var suback *paho.Suback
			if tt.subscribeErr == nil {
				suback = &paho.Suback{Reasons: tt.reasons}
			}

			if len(tt.subs) > 0 {
				tm.mockMQTTClient.EXPECT().Subscribe(ctx, &paho.Subscribe{
					Subscriptions: bus.ToPahoSubscribeOptions(tt.subs),
				}).Return(suback, tt.subscribeErr)
			}

			err = dbus.Subscribe(ctx)
			if tt.subscribeErr != nil {
				as.ErrorIs(err, bus.ErrSubscribeFailed)
				as.ErrorIs(err, tt.subscribeErr)
				return
			}

			if tt.wantErr != nil {
				as.ErrorIs(err, tt.wantErr)
				as.ErrorContains(err, tt.wantDetail)
				return
			}

			as.NoError(err)
		})
	}
}

type testMocks struct {
	mockMQTTClient     *mock_adapter.MockMQTTClient
	mockMessageHandler *mock_adapter.MockDaemonMessageHandler
}

func newMockDaemonBus(
	t *testing.T,
	ctx context.Context,
	subs []message.Subscription,
) (adapter.DaemonBus, *testMocks, error) {
	tm := &testMocks{
		mockMQTTClient:     mock_adapter.NewMockMQTTClient(t),
		mockMessageHandler: mock_adapter.NewMockDaemonMessageHandler(t),
	}

	dbus, err := bus.NewDaemonBus(
		ctx,
		sampleMQTTConnConfig,
		tm.mockMessageHandler,
		sampleHostname,
		subs,
		bus.WithMQTTClient(tm.mockMQTTClient))

	return dbus, tm, err
}
