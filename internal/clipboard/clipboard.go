package clipboard

import (
	"context"
	"sync"

	"github.com/jasonkwh/shengos-daemon/config"
	"github.com/jasonkwh/shengos-daemon/internal/adapter"
	"github.com/jasonkwh/shengos-daemon/internal/message"
	"go.uber.org/zap"
	"golang.design/x/clipboard"
)

var (
	_              adapter.Clipboard = (*clboard)(nil)
	watchClipboard                   = clipboard.Watch
)

type clboard struct {
	cfg      config.ClipboardConfig
	bus      adapter.DaemonBus
	zl       *zap.Logger
	hostname string

	written *[32]byte
	mu      sync.Mutex
}

func New(
	cfg config.ClipboardConfig,
	bus adapter.DaemonBus,
	hostname string,
	opts ...clboardOption,
) (adapter.Clipboard, error) {
	clb := &clboard{
		cfg:      cfg,
		bus:      bus,
		zl:       zap.NewNop(),
		hostname: hostname,
	}

	for _, opt := range opts {
		opt(clb)
	}

	if err := clipboard.Init(); err != nil {
		return nil, errInitFailed(err)
	}

	return clb, nil
}

func (c *clboard) Watch(
	ctx context.Context,
) error {
	c.zl.Debug("watching clipboard message")

	for data := range watchClipboard(ctx) {
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

		if err := c.bus.Publish(ctx, &message.ClipboardMessage{
			Origin: c.hostname,
			Format: msgFormat,
			Data:   data.Bytes,
		}); err != nil {
			c.zl.Error("failed to publish clipboard message",
				zap.Error(err))

			return errWatchFailed(err)
		}

		c.zl.Debug("clipboard message published",
			zap.String("origin", c.hostname),
			zap.String("format", msgFormat.String()))
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
