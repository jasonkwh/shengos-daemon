package message

import (
	"fmt"
)

const (
	clipboardVersion     = 1
	clipboardQoS         = QoSAtLeastOnce
	clipboardTopicFormat = "shengos/v%d/clipboard"
	clipboardRetain      = true
)

type ClipboardMessage struct {
	Origin string
	Format ClipboardFormat
	Data   []byte
}

func (msg *ClipboardMessage) Payload() ([]byte, error) {
	return serialize(msg)
}

func (msg *ClipboardMessage) Topic() string {
	return fmt.Sprintf(clipboardTopicFormat, clipboardVersion)
}

func (msg *ClipboardMessage) QoS() byte {
	return byte(clipboardQoS)
}

func (msg *ClipboardMessage) Retain() bool {
	return clipboardRetain
}

type ClipboardFormat uint8

const (
	ClipboardFormatUnknown ClipboardFormat = iota
	ClipboardFormatText
	ClipboardFormatImage
)
