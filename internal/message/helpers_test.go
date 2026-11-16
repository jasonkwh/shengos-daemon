package message

import (
	"reflect"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
)

func Test_message_serialize(t *testing.T) {
	tests := []struct {
		name string
		msg  any
		want []byte
	}{
		{
			name: "ClipboardMessage",
			msg: &ClipboardMessage{
				Origin: "test",
				Format: ClipboardFormatText,
				Data:   []byte("test"),
			},
			want: []byte("\x83\xa6Origin\xa4test\xa6Format\xcc\x01\xa4Data\xc4\x04test"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			as := assert.New(t)

			got, gotErr := serialize(tt.msg)
			as.NoError(gotErr)
			as.Equal(tt.want, got)
		})
	}
}

func Test_message_deserialize(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want any
	}{
		{
			name: "ClipboardMessage",
			data: []byte("\x83\xa6Origin\xa4test\xa6Format\xcc\x01\xa4Data\xc4\x04test"),
			want: &ClipboardMessage{
				Origin: "test",
				Format: ClipboardFormatText,
				Data:   []byte("test"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			as := assert.New(t)

			got := reflect.New(reflect.TypeOf(tt.want).Elem()).Interface()
			gotErr := deserialize(tt.data, got)
			as.NoError(gotErr)
			as.Equal(tt.want, got)
		})
	}
}
