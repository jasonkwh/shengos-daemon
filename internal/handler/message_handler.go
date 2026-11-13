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
	zl *zap.Logger
}

func NewDaemonMessageHandler(
	opts ...daemonMessageHandlerOption,
) adapter.DaemonMessageHandler {
	dh := &daemonMessageHandler{
		zl: zap.NewNop(),
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
	h.zl.Info("processing message",
		zap.String("topic", topic))

	switch topic {
	case message.ClipboardTopic:
		var msg message.ClipboardMessage

		if err := message.Deserialize(payload, &msg); err != nil {
			return fmt.Errorf("failed to deserialize clipboard message: %w", err)
		}

		h.zl.Debug("clipboard message",
			zap.String("origin", msg.Origin),
			zap.String("format", msg.Format.String()),
			zap.String("data_len", fmt.Sprintf("%d", len(msg.Data))))

		// TODO: Implement clipboard message processing

		return nil
	default:
		return fmt.Errorf("unknown topic: %s", topic)
	}
}
