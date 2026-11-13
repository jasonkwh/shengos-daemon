package clipboard

import (
	"context"

	"github.com/jasonkwh/shengos-daemon/config"
	"github.com/jasonkwh/shengos-daemon/internal/adapter"
	"github.com/jasonkwh/shengos-daemon/internal/message"
	"go.uber.org/zap"
	"golang.design/x/clipboard"
)

var _ adapter.Clipboard = (*clboard)(nil)

type clboard struct {
	cfg      config.ClipboardConfig
	bus      adapter.DaemonBus
	zl       *zap.Logger
	hostname string
}

func NewClipboard(
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

	// TODO: Implement clipboard watching

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

	var err error

	switch msg.Format {
	case message.ClipboardFormatText:
		_, err = clipboard.Write(ctx, clipboard.FmtText, msg.Data)
	case message.ClipboardFormatImage:
		_, err = clipboard.Write(ctx, clipboard.FmtImage, msg.Data)
	case message.ClipboardFormatFile:
		var uriList []byte

		uriList, err = c.sftpURIList(msg.Origin, msg.Data)
		if err != nil {
			break
		}

		_, err = clipboard.Write(ctx, clipboard.FmtFiles, uriList)
	default:
		err = errUnknownFormat(msg.Format)
	}

	if err != nil {
		return errApplyFailed(err)
	}

	c.zl.Debug("clipboard message applied",
		zap.String("origin", msg.Origin),
		zap.String("format", msg.Format.String()))

	return nil
}
