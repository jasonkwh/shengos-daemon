package clipboard

import (
	"context"
	"os"

	"github.com/jasonkwh/shengos-daemon/config"
	"github.com/jasonkwh/shengos-daemon/internal/adapter"
	"github.com/jasonkwh/shengos-daemon/internal/message"
	"go.uber.org/zap"
	"golang.design/x/clipboard"
)

var _ adapter.Clipboard = (*clboard)(nil)

type clboard struct {
	cfg config.ClipboardConfig
	bus adapter.DaemonBus

	zl *zap.Logger
}

func NewClipboard(
	cfg config.ClipboardConfig,
	bus adapter.DaemonBus,
	opts ...clboardOption,
) (adapter.Clipboard, error) {
	clb := &clboard{
		cfg: cfg,
		bus: bus,
		zl:  zap.NewNop(),
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

	hostname, err := os.Hostname()
	if err != nil {
		return errApplyFailed(err)
	}

	if msg.Origin == hostname {
		c.zl.Debug("clipboard message origin is the same as the hostname, skipping",
			zap.String("origin", msg.Origin),
			zap.String("hostname", hostname))

		return nil
	}

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
