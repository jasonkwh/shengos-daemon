package adapter

import "context"

type DaemonMessageHandler interface {
	ProcessMessage(
		ctx context.Context,
		topic string,
		payload []byte,
	) error
}
