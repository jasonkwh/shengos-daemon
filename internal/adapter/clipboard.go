package adapter

import (
	"context"

	"github.com/jasonkwh/shengos-daemon/internal/message"
)

type Clipboard interface {
	Watch(
		ctx context.Context,
	) error

	Apply(
		ctx context.Context,
		msg message.ClipboardMessage,
	) error
}
