package adapter

import "context"

type DaemonHandler interface {
	ProcessMessage(
		ctx context.Context,
		topic string,
		payload []byte,
	) error
}
