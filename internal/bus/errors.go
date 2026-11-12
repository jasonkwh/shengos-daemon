package bus

import (
	"errors"
	"fmt"
)

var (
	ErrConnectFailed     = errors.New("connect failed")
	ErrPublishFailed     = errors.New("publish failed")
	ErrSubscribeFailed   = errors.New("subscribe failed")
	ErrSubscribeRejected = errors.New("subscribe rejected")
	ErrCloseFailed       = errors.New("close failed")
)

func errConnectFailed(err error) error {
	return fmt.Errorf("%w: %w", ErrConnectFailed, err)
}

func errPublishFailed(err error) error {
	return fmt.Errorf("%w: %w", ErrPublishFailed, err)
}

func errSubscribeFailed(err error) error {
	return fmt.Errorf("%w: %w", ErrSubscribeFailed, err)
}

func errSubscribeRejected(reasons string) error {
	return fmt.Errorf("%w: %s", ErrSubscribeRejected, reasons)
}

func errCloseFailed(err error) error {
	return fmt.Errorf("%w: %w", ErrCloseFailed, err)
}
