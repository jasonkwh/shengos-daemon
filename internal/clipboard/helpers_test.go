package clipboard

import (
	"context"
	"errors"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"golang.design/x/clipboard"
)

func Test_clipboard_writeRemembered(t *testing.T) {
	items := []clipboard.Item{
		{Format: clipboard.FmtFiles, Bytes: []byte("sftp://jasonkwh@jasonkwh-test-rig/home/jasonkwh/test.txt")},
		{Format: clipboard.FmtText, Bytes: []byte("file:///home/jasonkwh/test.txt")},
	}

	t.Run("stores all hashes before a single write and retains them on success", func(t *testing.T) {
		as := assert.New(t)
		c := &clboard{
			written: make(map[clipboard.Format][32]byte),
		}
		calls := 0
		c.writeFunc = func(_ context.Context, opts ...clipboard.Option) (<-chan struct{}, error) {
			calls++
			as.Equal([]clipboard.Option{items[0], items[1]}, opts)
			for _, item := range items {
				savedHash, ok := c.written[item.Format]
				as.True(ok)
				as.Equal(clipboardHash(item.Format, item.Bytes), savedHash)
			}

			return nil, nil
		}

		as.NoError(c.writeRemembered(t.Context(), items))
		as.Equal(1, calls)
		for _, item := range items {
			as.True(c.ownWrite(item.Format, item.Bytes))
			as.False(c.ownWrite(item.Format, item.Bytes))
		}
	})

	t.Run("clears all hashes before returning a write error", func(t *testing.T) {
		as := assert.New(t)
		writeErr := errors.New("write failed")
		calls := 0
		c := &clboard{
			written: make(map[clipboard.Format][32]byte),
			writeFunc: func(_ context.Context, opts ...clipboard.Option) (<-chan struct{}, error) {
				calls++
				as.Equal([]clipboard.Option{items[0], items[1]}, opts)

				return nil, writeErr
			},
		}

		as.ErrorIs(c.writeRemembered(t.Context(), items), writeErr)
		as.Equal(1, calls)
		as.Empty(c.written)
		for _, item := range items {
			as.False(c.ownWrite(item.Format, item.Bytes))
		}
	})

	t.Run("rejects empty items without writing", func(t *testing.T) {
		as := assert.New(t)
		calls := 0
		c := &clboard{
			written: make(map[clipboard.Format][32]byte),
			writeFunc: func(context.Context, ...clipboard.Option) (<-chan struct{}, error) {
				calls++

				return nil, nil
			},
		}

		as.ErrorIs(c.writeRemembered(t.Context(), nil), ErrEmptyItems)
		as.ErrorIs(c.writeRemembered(t.Context(), []clipboard.Item{}), ErrEmptyItems)
		as.Equal(0, calls)
		as.Empty(c.written)
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

func Test_clipboard_ownWrite(t *testing.T) {
	format := clipboard.FmtText
	data := []byte("hello")

	t.Run("returns true for a remembered hash", func(t *testing.T) {
		as := assert.New(t)
		c := &clboard{
			written: make(map[clipboard.Format][32]byte),
		}

		c.remember(format, data)
		as.True(c.ownWrite(format, data))
	})

	t.Run("returns false after the hash is forgotten", func(t *testing.T) {
		as := assert.New(t)
		c := &clboard{
			written: make(map[clipboard.Format][32]byte),
		}

		c.remember(format, data)
		c.forget(format, data)
		as.False(c.ownWrite(format, data))
	})

	t.Run("returns false when nothing was remembered", func(t *testing.T) {
		as := assert.New(t)
		c := &clboard{
			written: make(map[clipboard.Format][32]byte),
		}

		as.False(c.ownWrite(format, data))
	})

	t.Run("consumes each format's remembered hash independently", func(t *testing.T) {
		data1 := []byte("file:///home/jasonkwh/test.txt")
		data2 := []byte("sftp://jasonkwh@jasonkwh-test-rig/home/jasonkwh/test.txt")

		as := assert.New(t)
		c := &clboard{
			written: make(map[clipboard.Format][32]byte),
		}

		c.remember(clipboard.FmtText, data1)
		c.remember(clipboard.FmtFiles, data2)

		ownWrite := c.ownWrite(clipboard.FmtText, data1)
		as.True(ownWrite)
		ownWrite = c.ownWrite(clipboard.FmtText, data1)
		as.False(ownWrite)
		ownWrite = c.ownWrite(clipboard.FmtFiles, data2)
		as.True(ownWrite)
		ownWrite = c.ownWrite(clipboard.FmtFiles, data2)
		as.False(ownWrite)
	})
}
