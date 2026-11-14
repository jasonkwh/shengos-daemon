package clipboard

import (
	"context"
	"errors"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"golang.design/x/clipboard"
)

func Test_clipboard_writeRemembered(t *testing.T) {
	format := clipboard.FmtText
	data := []byte("hello")

	t.Run("stores the hash before write", func(t *testing.T) {
		as := assert.New(t)
		c := &clboard{}
		c.writeFunc = func(context.Context, clipboard.Format, []byte, ...clipboard.Option) (<-chan struct{}, error) {
			as.True(c.ownWrite(format, data))

			return nil, nil
		}

		as.NoError(c.writeRemembered(t.Context(), format, data))
		as.False(c.ownWrite(format, data))
	})

	t.Run("clears the hash when write fails", func(t *testing.T) {
		as := assert.New(t)
		c := &clboard{
			writeFunc: func(context.Context, clipboard.Format, []byte, ...clipboard.Option) (<-chan struct{}, error) {
				return nil, errors.New("write failed")
			},
		}

		as.Error(c.writeRemembered(t.Context(), format, data))
		as.False(c.ownWrite(format, data))
	})
}

func Test_clipboard_toSFTPURIList(t *testing.T) {
	sampleSSHUser := "jasonkwh"

	tests := []struct {
		name    string
		origin  string
		data    []byte
		want    []byte
		wantErr error
	}{
		{
			name:   "success",
			origin: "jasonkwh-test-rig",
			data:   []byte("file:///home/jasonkwh/test.txt"),
			want:   []byte("sftp://jasonkwh@jasonkwh-test-rig/home/jasonkwh/test.txt\r\n"),
		},
		{
			name:   "success with multiple files",
			origin: "jasonkwh-test-rig",
			data:   []byte("file:///home/jasonkwh/test.txt\nfile:///home/jasonkwh/test2.txt"),
			want:   []byte("sftp://jasonkwh@jasonkwh-test-rig/home/jasonkwh/test.txt\r\nsftp://jasonkwh@jasonkwh-test-rig/home/jasonkwh/test2.txt\r\n"),
		},
		{
			name:   "success with multiple files and comments",
			origin: "jasonkwh-test-rig",
			data:   []byte("file:///home/jasonkwh/test.txt\n# comment\nfile:///home/jasonkwh/test2.txt"),
			want:   []byte("sftp://jasonkwh@jasonkwh-test-rig/home/jasonkwh/test.txt\r\nsftp://jasonkwh@jasonkwh-test-rig/home/jasonkwh/test2.txt\r\n"),
		},
		{
			name:   "success with folder",
			origin: "jasonkwh-test-rig",
			data:   []byte("file:///home/jasonkwh/test/"),
			want:   []byte("sftp://jasonkwh@jasonkwh-test-rig/home/jasonkwh/test/\r\n"),
		},
		{
			name:    "error with empty data",
			origin:  "jasonkwh-test-rig",
			data:    []byte(""),
			want:    nil,
			wantErr: ErrEmptyPathList,
		},
		{
			name:    "error with invalid path",
			origin:  "jasonkwh-test-rig",
			data:    []byte("http://www.google.com"),
			want:    nil,
			wantErr: ErrEmptyPathList,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			as := assert.New(t)

			got, err := toSFTPURIList(tt.origin, tt.data, sampleSSHUser)
			as.Equal(tt.want, got)
			as.ErrorIs(err, tt.wantErr)
		})
	}
}
