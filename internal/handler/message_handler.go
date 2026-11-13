package handler

import (
	"context"
	"fmt"

	"github.com/jasonkwh/shengos-daemon/internal/adapter"
	"github.com/jasonkwh/shengos-daemon/internal/message"
	"go.uber.org/zap"
)

var _ adapter.DaemonMessageHandler = (*daemonMessageHandler)(nil)

type daemonMessageHandler struct {
	clipboard adapter.Clipboard
	zl        *zap.Logger
}

func NewDaemonMessageHandler(
	clipboard adapter.Clipboard,
	opts ...daemonMessageHandlerOption,
) adapter.DaemonMessageHandler {
	dh := &daemonMessageHandler{
		clipboard: clipboard,
		zl:        zap.NewNop(),
	}

	for _, opt := range opts {
		opt(dh)
	}

	return dh
}

func (h *daemonMessageHandler) ProcessMessage(
	ctx context.Context,
	topic string,
	payload []byte,
) error {
	h.zl.Debug("processing message",
		zap.String("topic", topic))

	switch topic {
	case message.ClipboardTopic:
		return h.processClipboardMessage(ctx, payload)
	default:
		h.zl.Error("unknown topic",
			zap.String("topic", topic))

		return errUnknownTopic(topic)
	}
}

func (h *daemonMessageHandler) processClipboardMessage(
	ctx context.Context,
	payload []byte,
) error {
	var msg message.ClipboardMessage

	if err := msg.Unmarshal(payload); err != nil {
		h.zl.Error("failed to unmarshal clipboard message",
			zap.Error(err))

		return errUnmarshalFailed(err)
	}

	if err := h.clipboard.Apply(ctx, msg); err != nil {
		h.zl.Error("failed to apply clipboard message",
			zap.String("origin", msg.Origin),
			zap.String("format", msg.Format.String()),
			zap.Error(err))

		return err
	}

	h.zl.Info("clipboard message applied",
		zap.String("origin", msg.Origin),
		zap.String("format", msg.Format.String()),
		zap.String("data_len", fmt.Sprintf("%d", len(msg.Data))))

	return nil
}
