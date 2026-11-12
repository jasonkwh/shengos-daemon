package adapter

import (
	"context"
	"io"
)

type DaemonBus interface {
	Publish(
		ctx context.Context,
		message DaemonMessage,
	) error

	Subscribe(
		ctx context.Context,
	) error

	io.Closer
}
