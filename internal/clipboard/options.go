package clipboard

import "go.uber.org/zap"

type clboardOption func(*clboard)

func WithLogger(zl *zap.Logger) clboardOption {
	return func(c *clboard) {
		c.zl = zl
	}
}
