package handler

import (
	"errors"
	"fmt"
)

var (
	ErrUnmarshalFailed = errors.New("failed to unmarshal message")
)

func errUnmarshalFailed(err error) error {
	return fmt.Errorf("%w: %w", ErrUnmarshalFailed, err)
}
