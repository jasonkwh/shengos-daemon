package clipboard

import (
	"testing"

	"github.com/go-openapi/testify/v2/assert"
)

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
