package clipboard

import (
	"errors"
	"fmt"
)

var (
	ErrInitFailed    = errors.New("init failed")
	ErrWatchFailed   = errors.New("watch failed")
	ErrApplyFailed   = errors.New("apply failed")
	ErrEmptyPathList = errors.New("empty path list")
	ErrUnknownFormat = errors.New("unknown format")
)

func errInitFailed(err error) error {
	return fmt.Errorf("%w: %w", ErrInitFailed, err)
}

func errWatchFailed(err error) error {
	return fmt.Errorf("%w: %w", ErrWatchFailed, err)
}

func errApplyFailed(err error) error {
	return fmt.Errorf("%w: %w", ErrApplyFailed, err)
}
