package bus

import (
	"errors"
	"testing"

	"github.com/eclipse/paho.golang/paho"
	"github.com/go-openapi/testify/v2/assert"
	"github.com/jasonkwh/shengos-daemon/internal/adapter"
	"github.com/jasonkwh/shengos-daemon/internal/message"
	mock_adapter "github.com/jasonkwh/shengos-daemon/mocks/adapter"
	"go.uber.org/zap"
)

const sampleHostname = "jasonkwh-test-rig"

func Test_daemonBus_onPublishReceived(t *testing.T) {
	processErr := errors.New("process error")
	ackErr := errors.New("ack error")

	tests := []struct {
		name         string
		packet       *paho.Publish
		processedErr error
		ackErr       error
		wantErr      error
	}{
		{
			name: "success",
			packet: &paho.Publish{
				Topic:   message.ClipboardTopic,
				Payload: []byte("payload"),
			},
		},
		{
			name: "process error",
			packet: &paho.Publish{
				Topic:   message.ClipboardTopic,
				Payload: []byte("payload"),
			},
			processedErr: processErr,
			wantErr:      processErr,
		},
		{
			name: "ack error",
			packet: &paho.Publish{
				Topic:   message.ClipboardTopic,
				Payload: []byte("payload"),
			},
			ackErr:  ackErr,
			wantErr: ackErr,
		},
		{
			name: "unknown topic",
			packet: &paho.Publish{
				Topic:   "shengos/v1/unknown",
				Payload: []byte("payload"),
			},
			wantErr: ErrUnknownTopic,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			as := assert.New(t)
			ctx := t.Context()

			dbus, tm := newMockDaemonBus(t)

			topicHandlers := map[string]adapter.DaemonTopicHandler{
				message.ClipboardTopic: tm.mockTopicHandler,
			}

			if _, ok := topicHandlers[tt.packet.Topic]; ok {
				tm.mockTopicHandler.EXPECT().
					Handle(ctx, tt.packet.Payload).
					Return(tt.processedErr)

				if tt.processedErr == nil {
					tm.mockMQTTPublishAcker.EXPECT().
						Ack(tt.packet).
						Return(tt.ackErr)
				}
			}

			handled, err := dbus.onPublishReceived(ctx, topicHandlers)(paho.PublishReceived{
				Packet: tt.packet,
			})

			as.True(handled)
			as.ErrorIs(err, tt.wantErr)
		})
	}
}

type testMocks struct {
	mockMQTTClient       *mock_adapter.MockMQTTClient
	mockMQTTPublishAcker *mock_adapter.MockMQTTPublishAcker
	mockTopicHandler     *mock_adapter.MockDaemonTopicHandler
}

func newMockDaemonBus(t *testing.T) (*daemonBus, *testMocks) {
	tm := &testMocks{
		mockMQTTClient:       mock_adapter.NewMockMQTTClient(t),
		mockMQTTPublishAcker: mock_adapter.NewMockMQTTPublishAcker(t),
		mockTopicHandler:     mock_adapter.NewMockDaemonTopicHandler(t),
	}

	return &daemonBus{
		mqttCl:       tm.mockMQTTClient,
		publishAcker: tm.mockMQTTPublishAcker,
		zl:           zap.NewNop(),
		hostname:     sampleHostname,
	}, tm
}
