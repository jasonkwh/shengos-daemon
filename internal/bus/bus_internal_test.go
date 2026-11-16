package bus

import (
	"errors"
	"testing"

	"github.com/eclipse/paho.golang/paho"
	"github.com/go-openapi/testify/v2/assert"
	mock_adapter "github.com/jasonkwh/shengos-daemon/mocks/adapter"
	"go.uber.org/zap"
)

const sampleHostname = "jasonkwh-test-rig"

func Test_daemonBus_onPublishReceived(t *testing.T) {
	packet := &paho.Publish{
		Topic:   "shengos/v1/clipboard",
		Payload: []byte("payload"),
	}

	processErr := errors.New("process error")
	ackErr := errors.New("ack error")

	tests := []struct {
		name         string
		processedErr error
		ackErr       error
		wantErr      error
	}{
		{
			name: "success",
		},
		{
			name:         "process error",
			processedErr: processErr,
			wantErr:      processErr,
		},
		{
			name:    "ack error",
			ackErr:  ackErr,
			wantErr: ackErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			as := assert.New(t)
			ctx := t.Context()

			dbus, tm := newMockDaemonBus(t)

			tm.mockMessageHandler.EXPECT().
				ProcessMessage(ctx, packet.Topic, packet.Payload).
				Return(tt.processedErr)

			if tt.processedErr == nil {
				tm.mockMQTTPublishAcker.EXPECT().
					Ack(packet).
					Return(tt.ackErr)
			}

			handled, err := dbus.onPublishReceived(ctx, tm.mockMessageHandler)(paho.PublishReceived{
				Packet: packet,
			})

			as.True(handled)
			as.ErrorIs(err, tt.wantErr)
		})
	}
}

type testMocks struct {
	mockMQTTClient       *mock_adapter.MockMQTTClient
	mockMQTTPublishAcker *mock_adapter.MockMQTTPublishAcker
	mockMessageHandler   *mock_adapter.MockDaemonMessageHandler
}

func newMockDaemonBus(t *testing.T) (*daemonBus, *testMocks) {
	tm := &testMocks{
		mockMQTTClient:       mock_adapter.NewMockMQTTClient(t),
		mockMQTTPublishAcker: mock_adapter.NewMockMQTTPublishAcker(t),
		mockMessageHandler:   mock_adapter.NewMockDaemonMessageHandler(t),
	}

	return &daemonBus{
		mqttCl:       tm.mockMQTTClient,
		publishAcker: tm.mockMQTTPublishAcker,
		zl:           zap.NewNop(),
		hostname:     sampleHostname,
	}, tm
}
