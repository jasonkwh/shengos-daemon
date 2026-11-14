package clipboard_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"github.com/go-openapi/testify/v2/require"
	"github.com/jasonkwh/shengos-daemon/internal/adapter"
	clboard "github.com/jasonkwh/shengos-daemon/internal/clipboard"
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
			rs := require.New(t)
			ctx := t.Context()

			clb, tm, err := newMockClipboard(t,
				clboard.WithWatchFunc(func(context.Context, ...clipboard.Option) <-chan clipboard.Data {
					changes := make(chan clipboard.Data, 1)
					changes <- clipboard.Data{Format: tt.format, Bytes: tt.data}

					close(changes)

					return changes
				}), clboard.WithSkipInit(true))
			rs.NoError(err)

			tm.mockDaemonBus.EXPECT().
				Publish(ctx, mock.MatchedBy(func(msg adapter.DaemonMessage) bool {
					got, ok := msg.(*message.ClipboardMessage)

					return ok &&
						got.Origin == tt.want.Origin &&
						got.Format == tt.want.Format &&
						bytes.Equal(got.Data, tt.want.Data)
				})).Return(nil)

			as.NoError(clb.Watch(ctx, tm.mockDaemonBus.Publish))
		})
	}
}

func Test_clipboard_Apply(t *testing.T) {
	testOrigin := fmt.Sprintf("%s-1", sampleHostname)

	tests := []struct {
		name       string
		msg        message.ClipboardMessage
		wantErr    error
		wantFormat clipboard.Format
		wantWrite  bool
	}{
		{
			name: "skipping own message",
			msg: message.ClipboardMessage{
				Origin: sampleHostname,
				Format: message.ClipboardFormatText,
				Data:   []byte("hello"),
			},
		},
		{
			name: "unknown clipboard format 1",
			msg: message.ClipboardMessage{
				Origin: testOrigin,
				Format: message.ClipboardFormatUnknown,
				Data:   []byte("hello"),
			},
			wantErr: clboard.ErrUnknownFormat,
		},
		{
			name: "unknown clipboard format 2",
			msg: message.ClipboardMessage{
				Origin: testOrigin,
				Format: message.ClipboardFormatUnknown,
				Data:   []byte("hello"),
			},
			wantErr: clboard.ErrApplyFailed,
		},
		{
			name: "text",
			msg: message.ClipboardMessage{
				Origin: testOrigin,
				Format: message.ClipboardFormatText,
				Data:   []byte("hello"),
			},
			wantFormat: clipboard.FmtText,
			wantWrite:  true,
		},
		{
			name: "image",
			msg: message.ClipboardMessage{
				Origin: testOrigin,
				Format: message.ClipboardFormatImage,
				Data:   []byte("image.png"),
			},
			wantFormat: clipboard.FmtImage,
			wantWrite:  true,
		},
		{
			name: "files",
			msg: message.ClipboardMessage{
				Origin: testOrigin,
				Format: message.ClipboardFormatFiles,
				Data:   []byte("file:///home/jasonkwh/test.txt"),
			},
			wantFormat: clipboard.FmtFiles,
			wantWrite:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			as := assert.New(t)
			rs := require.New(t)
			ctx := t.Context()

			var (
				gotFormat clipboard.Format
				gotData   []byte
				wrote     bool
			)

			clb, _, err := newMockClipboard(t,
				clboard.WithWriteFunc(func(
					_ context.Context,
					format clipboard.Format,
					data []byte,
					_ ...clipboard.Option,
				) (<-chan struct{}, error) {
					wrote = true
					gotFormat = format
					gotData = data

					return nil, nil
				}),
				clboard.WithSkipInit(true),
			)
			rs.NoError(err)

			err = clb.Apply(ctx, tt.msg)
			as.ErrorIs(err, tt.wantErr)
			as.Equal(tt.wantWrite, wrote)

			if !tt.wantWrite {
				return
			}

			as.Equal(tt.wantFormat, gotFormat)
			as.Equal(tt.msg.Data, gotData)
		})
	}
}
