package handler

import "go.uber.org/zap"

type daemonHandlerOption func(*daemonHandler)

func WithLogger(zl *zap.Logger) daemonHandlerOption {
	return func(h *daemonHandler) {
		h.zl = zl
	}
}
