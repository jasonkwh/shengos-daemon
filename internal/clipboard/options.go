package clipboard

import (
	"context"

	"go.uber.org/zap"
	"golang.design/x/clipboard"
)

type ClipboardOption func(*clboard)

func WithLogger(zl *zap.Logger) ClipboardOption {
	return func(c *clboard) {
		c.zl = zl
	}
}

func WithWatchFunc(
	watchFunc func(ctx context.Context, opts ...clipboard.Option) <-chan clipboard.Data,
) ClipboardOption {
	return func(c *clboard) {
		c.watchFunc = watchFunc
	}
}
