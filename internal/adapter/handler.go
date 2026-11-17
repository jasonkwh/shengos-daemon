package adapter

import "context"

type DaemonTopicHandler interface {
	Handle(ctx context.Context, payload []byte) error
}
