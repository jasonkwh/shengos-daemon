package clipboard

import (
	"context"
	"iter"
	"sync"

	"github.com/jasonkwh/shengos-daemon/config"
	"github.com/jasonkwh/shengos-daemon/internal/adapter"
	"github.com/jasonkwh/shengos-daemon/internal/message"
	"go.uber.org/zap"
	"golang.design/x/clipboard"
)

var _ adapter.Clipboard = (*clboard)(nil)

type clboard struct {
	cfg       config.ClipboardConfig
	bus       adapter.DaemonBus
	zl        *zap.Logger
	hostname  string
	watchFunc func(ctx context.Context, opts ...clipboard.Option) <-chan clipboard.Data
	skipInit  bool

	written *[32]byte
	mu      sync.Mutex
}

func New(
	cfg config.ClipboardConfig,
	bus adapter.DaemonBus,
	hostname string,
	opts ...ClipboardOption,
) (adapter.Clipboard, error) {
	clb := &clboard{
		cfg:       cfg,
		bus:       bus,
		zl:        zap.NewNop(),
		hostname:  hostname,
		watchFunc: clipboard.Watch,
	}

	for _, opt := range opts {
		opt(clb)
	}

	if !clb.skipInit {
		if err := clipboard.Init(); err != nil {
			return nil, errInitFailed(err)
		}
	}

	return clb, nil
}

func (c *clboard) Watch(
	ctx context.Context,
) error {
	c.zl.Debug("watching clipboard message")

	for msg := range c.clipboardMessages(ctx) {
		if err := c.bus.Publish(ctx, &msg); err != nil {
			c.zl.Error("failed to publish clipboard message",
				zap.Error(err))

			return errWatchFailed(err)
		}

		c.zl.Debug("clipboard message published",
			zap.String("origin", c.hostname),
			zap.String("format", msg.Format.String()))
	}

	c.zl.Debug("ended watching clipboard message")

	return nil
}

func (c *clboard) Apply(
	ctx context.Context,
	msg message.ClipboardMessage,
) error {
	c.zl.Debug("applying clipboard message",
		zap.String("origin", msg.Origin),
		zap.String("format", msg.Format.String()))

	if msg.Origin == c.hostname {
		c.zl.Debug("clipboard message origin is the same as the hostname, skipping",
			zap.String("origin", msg.Origin),
			zap.String("hostname", c.hostname))

		return nil
	}

	format, err := toClipboardFormat(msg.Format)
	if err != nil {
		return errApplyFailed(err)
	}

	if err := c.writeRemembered(ctx, format, msg.Data); err != nil {
		return errApplyFailed(err)
	}

	c.zl.Debug("clipboard message applied",
		zap.String("origin", msg.Origin),
		zap.String("format", msg.Format.String()))

	return nil
}

func (c *clboard) clipboardMessages(
	ctx context.Context,
) iter.Seq[message.ClipboardMessage] {
	return func(yield func(message.ClipboardMessage) bool) {
		for data := range c.watchFunc(ctx) {
			if data.Sensitive || c.ownWrite(data.Format, data.Bytes) {
				c.zl.Debug("clipboard data is sensitive or own write, skipping")

				continue
			}

			msgFormat := toMessageClipboardFormat(data.Format)

			if msgFormat == message.ClipboardFormatUnknown {
				c.zl.Debug("clipboard message format is unknown, skipping")

				continue
			}

			if msgFormat == message.ClipboardFormatFiles {
				uriList, err := toSFTPURIList(c.hostname, data.Bytes, c.cfg.SSHUser)
				if err != nil {
					c.zl.Debug("failed to convert clipboard message to sftp uri list",
						zap.Error(err))

					continue
				}

				data.Bytes = uriList
			}

			if !yield(message.ClipboardMessage{
				Origin: c.hostname,
				Format: msgFormat,
				Data:   data.Bytes,
			}) { //nolint:whitespace
				return
			}
		}
	}
}
