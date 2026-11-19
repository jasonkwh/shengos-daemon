package handler

import (
	"context"

	"github.com/jasonkwh/shengos-daemon/internal/adapter"
	"github.com/jasonkwh/shengos-daemon/internal/message"
	"go.uber.org/zap"
)

var _ adapter.DaemonTopicHandler = (*clipboardTopicHandler)(nil)

type clipboardTopicHandler struct {
	clipboard adapter.Clipboard
	core      *handlerCore
}

func NewClipboardTopicHandler(
	clipboard adapter.Clipboard,
	opts ...daemonTopicHandlerOption,
) adapter.DaemonTopicHandler {
	dh := &clipboardTopicHandler{
		clipboard: clipboard,
		core:      &handlerCore{zl: zap.NewNop()},
	}

	for _, opt := range opts {
		opt(dh.core)
	}

	return dh
}

func (h *clipboardTopicHandler) Handle(
	ctx context.Context,
	payload []byte,
) error {
	topic := message.ClipboardTopic

	h.core.zl.Debug("processing clipboard message",
		zap.String("topic", topic))

	var msg message.ClipboardMessage

	if err := msg.Unmarshal(payload); err != nil {
		h.core.zl.Error("failed to unmarshal clipboard message",
			zap.Error(err))

		return errUnmarshalFailed(err)
	}

	if err := h.clipboard.Apply(ctx, msg); err != nil {
		h.core.zl.Error("failed to apply clipboard message",
			zap.String("origin", msg.Origin),
			zap.Error(err))

		return err
	}

	return nil
}
