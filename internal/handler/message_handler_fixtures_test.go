package handler_test

import "github.com/jasonkwh/shengos-daemon/internal/message"

var (
	sampleClipboardMessage = message.ClipboardMessage{
		Origin: "jasonkwh-test-rig",
		Format: message.ClipboardFormatText,
		Data:   []byte("test"),
	}
)
