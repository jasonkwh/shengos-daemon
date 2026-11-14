package adapter

import (
	"context"

	"github.com/jasonkwh/shengos-daemon/internal/message"
)

type Clipboard interface {
	Watch(
		ctx context.Context,
		publishFunc func(
			ctx context.Context,
			message DaemonMessage,
		) error,
	) error

	Apply(
		ctx context.Context,
		msg message.ClipboardMessage,
	) error
}
