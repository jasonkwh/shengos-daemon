package handler_test

import "github.com/jasonkwh/shengos-daemon/internal/message"

var (
	sampleClipboardMessage = message.ClipboardMessage{
		Origin: "jasonkwh-test-rig",
		Items: []message.ClipboardItem{
			{
				Format: message.ClipboardFormatText,
				Data:   []byte("test"),
			},
		},
	}
)
