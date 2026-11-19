package message

import (
	"fmt"
)

const (
	clipboardVersion = 2
	clipboardRetain  = false
)

var (
	ClipboardTopic = fmt.Sprintf("%s/clipboard",
		topicPrefix(clipboardVersion))

	ClipboardSubscription = Subscription{
		Topic:             ClipboardTopic,
		QoS:               QoSAtLeastOnce,
		NoLocal:           true,
		RetainAsPublished: false,
		RetainHandling:    RetainDoNotSend,
	}
)

type ClipboardItem struct {
	Format ClipboardFormat
	Data   []byte
}

type ClipboardMessage struct {
	Origin string
	Items  []ClipboardItem
}

func (msg *ClipboardMessage) Payload() ([]byte, error) {
	return serialize(msg)
}

func (msg *ClipboardMessage) Unmarshal(data []byte) error {
	return deserialize(data, msg)
}

func (msg *ClipboardMessage) Topic() string {
	return ClipboardTopic
}

func (msg *ClipboardMessage) QoS() byte {
	return byte(ClipboardSubscription.QoS)
}

func (msg *ClipboardMessage) Retain() bool {
	return clipboardRetain
}

type ClipboardFormat uint8

const (
	ClipboardFormatUnknown ClipboardFormat = iota
	ClipboardFormatText
	ClipboardFormatImage
	ClipboardFormatFiles
)

func (f ClipboardFormat) String() string {
	switch f {
	case ClipboardFormatText:
		return "text"
	case ClipboardFormatImage:
		return "image"
	case ClipboardFormatFiles:
		return "files"
	default:
		return "unknown"
	}
}
