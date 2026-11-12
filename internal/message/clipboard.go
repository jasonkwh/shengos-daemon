package message

import (
	"fmt"
)

const (
	clipboardVersion     = 1
	clipboardQoS         = QoSAtLeastOnce
	clipboardTopicFormat = "shengos/v%d/clipboard"
)

type ClipboardMessage struct {
	Origin string
	Format ClipboardFormat
	Data   []byte
}

func (m *ClipboardMessage) Payload() ([]byte, error) {
	return serialize(m)
}

func (m *ClipboardMessage) Topic() string {
	return fmt.Sprintf(clipboardTopicFormat, clipboardVersion)
}

func (m *ClipboardMessage) QoS() byte {
	return byte(clipboardQoS)
}

type ClipboardFormat uint8

const (
	ClipboardFormatUnknown ClipboardFormat = iota
	ClipboardFormatText
	ClipboardFormatImage
)
