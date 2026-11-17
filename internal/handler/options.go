package handler

import "go.uber.org/zap"

type daemonTopicHandlerOption func(*handlerCore)

func WithLogger(zl *zap.Logger) daemonTopicHandlerOption {
	return func(h *handlerCore) {
		h.zl = zl
	}
}
