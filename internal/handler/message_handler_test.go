package handler_test

import (
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
	"github.com/jasonkwh/shengos-daemon/internal/adapter"
	"github.com/jasonkwh/shengos-daemon/internal/handler"
	"github.com/jasonkwh/shengos-daemon/internal/message"
)

func Test_daemonMessageHandler_ProcessMessage(t *testing.T) {
	tests := []struct {
		name    string
		topic   string
		payload []byte
		want    *message.ClipboardMessage
		wantErr error
	}{
		{
			name:    "success",
			topic:   message.ClipboardTopic,
			payload: messagePayload(t, &sampleClipboardMessage),
			want:    &sampleClipboardMessage,
		},
		{
			name:    "unknown topic",
			topic:   "shengos/v1/unknown",
			payload: []byte{},
			wantErr: handler.ErrUnknownTopic,
		},
		{
			name:    "failed to unmarshal message",
			topic:   message.ClipboardTopic,
			payload: []byte("invalid"),
			wantErr: handler.ErrUnmarshalFailed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			as := assert.New(t)
			ctx := t.Context()
			hdl, tm := newMockMessageHandler(t)

			if tt.want != nil {
				tm.mockClipboard.EXPECT().Apply(ctx,
					*tt.want).Return(nil)
			}

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
