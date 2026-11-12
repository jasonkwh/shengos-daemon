package message

import (
	"fmt"
)

const (
	clipboardVersion = 1
	clipboardRetain  = true
)

var (
	ClipboardSubscription = Subscription{
		Topic: fmt.Sprintf("shengos/v%d/clipboard",
			clipboardVersion),
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
	return ClipboardSubscription.Topic
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
)
