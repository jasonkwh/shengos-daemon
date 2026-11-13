package message

import (
	"fmt"
)

const (
	clipboardVersion = 1
	clipboardRetain  = true
)

var (
	ClipboardTopic = fmt.Sprintf("shengos/v%d/clipboard",
		clipboardVersion)

	ClipboardSubscription = Subscription{
		Topic:             ClipboardTopic,
		QoS:               QoSAtLeastOnce,
		NoLocal:           true,
		RetainAsPublished: false,
		RetainHandling:    RetainHandlingOnSubscribe,
	}
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
	ClipboardFormatFile
)

func (f ClipboardFormat) String() string {
	switch f {
	case ClipboardFormatText:
		return "text"
	case ClipboardFormatImage:
		return "image"
	case ClipboardFormatFile:
		return "file"
	default:
		return "unknown"
	}
}
