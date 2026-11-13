package handler

import "go.uber.org/zap"

type daemonMessageHandlerOption func(*daemonMessageHandler)

func WithMessageLogger(zl *zap.Logger) daemonMessageHandlerOption {
	return func(h *daemonMessageHandler) {
		h.zl = zl
	}
}
