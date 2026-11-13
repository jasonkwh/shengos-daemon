package clipboard

import (
	"bytes"
	"context"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/jasonkwh/shengos-daemon/internal/adapter"
	"github.com/jasonkwh/shengos-daemon/internal/message"
	"github.com/stretchr/testify/mock"
	"golang.design/x/clipboard"
)

func Test_clipboard_Watch(t *testing.T) {
	tests := []struct {
		name   string
		format clipboard.Format
		data   []byte
		want   *message.ClipboardMessage
	}{
		{
			name:   "text",
			format: clipboard.FmtText,
			data:   []byte("hello"),
			want: &message.ClipboardMessage{
				Origin: sampleHostname,
				Format: message.ClipboardFormatText,
				Data:   []byte("hello"),
			},
		},
		{
			name:   "image",
			format: clipboard.FmtImage,
			data:   []byte("image.png"),
			want: &message.ClipboardMessage{
				Origin: sampleHostname,
				Format: message.ClipboardFormatImage,
				Data:   []byte("image.png"),
			},
		},
		{
			name:   "files",
			format: clipboard.FmtFiles,
			data:   []byte("file:///home/jasonkwh/test.txt"),
			want: &message.ClipboardMessage{
				Origin: sampleHostname,
				Format: message.ClipboardFormatFiles,
				Data:   []byte("sftp://jasonkwh@jasonkwh-test-rig/home/jasonkwh/test.txt\r\n"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			as := assert.New(t)
			ctx := t.Context()

			clb, tm, err := newMockClipboard(t)
			as.NoError(err)

			previous := watchClipboard
			t.Cleanup(func() { watchClipboard = previous })

			watchClipboard = func(context.Context, ...clipboard.Option) <-chan clipboard.Data {
				changes := make(chan clipboard.Data, 1)
				changes <- clipboard.Data{Format: tt.format, Bytes: tt.data}

				close(changes)

				return changes
			}

			tm.mockDaemonBus.EXPECT().
				Publish(ctx, mock.MatchedBy(func(msg adapter.DaemonMessage) bool {
					got, ok := msg.(*message.ClipboardMessage)

					return ok &&
						got.Origin == tt.want.Origin &&
						got.Format == tt.want.Format &&
						bytes.Equal(got.Data, tt.want.Data)
				})).Return(nil)

			as.NoError(clb.Watch(ctx))
		})
	}
}
