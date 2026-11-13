package handler

import (
	"errors"
	"fmt"
)

var (
	ErrUnknownTopic         = errors.New("unknown topic")
	ErrUnmarshalFailed      = errors.New("failed to unmarshal message")
	ErrClipboardApplyFailed = errors.New("failed to apply clipboard message")
)

func errUnknownTopic(topic string) error {
	return fmt.Errorf("%w: %s", ErrUnknownTopic, topic)
}

func errUnmarshalFailed(err error) error {
	return fmt.Errorf("%w: %w", ErrUnmarshalFailed, err)
}

func errClipboardApplyFailed(err error) error {
	return fmt.Errorf("%w: %w", ErrClipboardApplyFailed, err)
}
