package handler

import (
	"errors"
	"fmt"
)

var (
	ErrUnknownTopic    = errors.New("unknown topic")
	ErrUnmarshalFailed = errors.New("failed to unmarshal message")
)

func errUnknownTopic(topic string) error {
	return fmt.Errorf("%w: %s", ErrUnknownTopic, topic)
}

func errUnmarshalFailed(err error) error {
	return fmt.Errorf("%w: %w", ErrUnmarshalFailed, err)
}
