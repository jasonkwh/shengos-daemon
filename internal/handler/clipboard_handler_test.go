package handler_test

import (
	"errors"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/jasonkwh/shengos-daemon/internal/adapter"
	"github.com/jasonkwh/shengos-daemon/internal/handler"
	"github.com/jasonkwh/shengos-daemon/internal/message"
	mock_adapter "github.com/jasonkwh/shengos-daemon/mocks/adapter"
)

func Test_clipboardTopicHandler_Handle(t *testing.T) {
	applyErr := errors.New("apply failed")

	tests := []struct {
		name     string
		payload  []byte
		want     *message.ClipboardMessage
		applyErr error
		wantErr  error
	}{
		{
			name:    "success",
			payload: messagePayload(t, &sampleClipboardMessage),
			want:    &sampleClipboardMessage,
		},
		{
			name:    "failed to unmarshal clipboard message",
			payload: []byte("invalid"),
			wantErr: handler.ErrUnmarshalFailed,
		},
		{
			name:     "failed to apply clipboard message",
			payload:  messagePayload(t, &sampleClipboardMessage),
			want:     &sampleClipboardMessage,
			applyErr: applyErr,
			wantErr:  applyErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			as := assert.New(t)
			ctx := t.Context()
			hdl, mockClipboard := newMockClipboardTopicHandler(t)

			if tt.want != nil {
				mockClipboard.EXPECT().
					Apply(ctx, *tt.want).
					Return(tt.applyErr)
			}

			err := hdl.Handle(ctx, tt.payload)
			as.ErrorIs(err, tt.wantErr)
		})
	}
}

func newMockClipboardTopicHandler(
	t *testing.T,
) (adapter.DaemonTopicHandler, *mock_adapter.MockClipboard) {
	mockClipboard := mock_adapter.NewMockClipboard(t)

	return handler.NewClipboardTopicHandler(
		mockClipboard,
	), mockClipboard
}
