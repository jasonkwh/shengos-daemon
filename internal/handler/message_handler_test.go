package handler_test

import (
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
	"github.com/jasonkwh/shengos-daemon/internal/adapter"
	"github.com/jasonkwh/shengos-daemon/internal/handler"
	"github.com/jasonkwh/shengos-daemon/internal/message"
	mock_adapter "github.com/jasonkwh/shengos-daemon/mocks/adapter"
)

func Test_daemonMessageHandler_ProcessMessage(t *testing.T) {
	tests := []struct {
		name             string
		topic            string
		payload          []byte
		enabledClipboard bool
		want             *message.ClipboardMessage
		wantErr          error
	}{
		{
			name:             "success",
			topic:            message.ClipboardTopic,
			payload:          messagePayload(t, &sampleClipboardMessage),
			enabledClipboard: true,
			want:             &sampleClipboardMessage,
		},
		{
			name:             "unknown topic",
			topic:            "shengos/v1/unknown",
			payload:          []byte{},
			enabledClipboard: true,
			wantErr:          handler.ErrUnknownTopic,
		},
		{
			name:             "failed to unmarshal message",
			topic:            message.ClipboardTopic,
			payload:          []byte("invalid"),
			enabledClipboard: true,
			wantErr:          handler.ErrUnmarshalFailed,
		},
		{
			name:    "clipboard is not enabled",
			topic:   message.ClipboardTopic,
			payload: messagePayload(t, &sampleClipboardMessage),
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			as := assert.New(t)
			ctx := t.Context()

			var clb adapter.Clipboard

			if tt.enabledClipboard {
				mockClipboard := mock_adapter.NewMockClipboard(t)
				clb = mockClipboard

				if tt.want != nil {
					mockClipboard.EXPECT().
						Apply(ctx, *tt.want).
						Return(nil)
				}
			}

			hdl := handler.NewDaemonMessageHandler(clb)

			err := hdl.ProcessMessage(ctx, tt.topic, tt.payload)
			as.ErrorIs(err, tt.wantErr)
		})
	}
}

func messagePayload(t *testing.T, msg adapter.DaemonMessage) []byte {
	t.Helper()
	rs := require.New(t)

	payload, err := msg.Payload()
	rs.NoError(err)

	return payload
}
