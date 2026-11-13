package clipboard

import (
	"errors"
	"fmt"

	"github.com/jasonkwh/shengos-daemon/internal/message"
)

var (
	ErrInitFailed    = errors.New("init failed")
	ErrWatchFailed   = errors.New("watch failed")
	ErrApplyFailed   = errors.New("apply failed")
	ErrUnknownFormat = errors.New("unknown format")
	ErrRelativePath  = errors.New("relative path")
	ErrEmptyPathList = errors.New("empty path list")
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

func errUnknownFormat(format message.ClipboardFormat) error {
	return fmt.Errorf("%w: %s", ErrUnknownFormat, format)
}

func errRelativePath(path []byte) error {
	return fmt.Errorf("%w: %s", ErrRelativePath, path)
}
