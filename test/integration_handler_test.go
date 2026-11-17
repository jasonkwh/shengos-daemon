//go:build integration
// +build integration

package integration

import (
	"context"

	"github.com/jasonkwh/shengos-daemon/internal/message"
)

type suiteClipboardHandler struct {
	s *integrationSuite
}

func (h suiteClipboardHandler) Handle(ctx context.Context, payload []byte) error {
	return h.s.testHandler(ctx, payload)
}

func (s *integrationSuite) testHandler(
	_ context.Context,
	payload []byte,
) error {
	var msg message.ClipboardMessage

	if err := msg.Unmarshal(payload); err != nil {
		return err
	}

	s.T().Logf("mqtt received: origin=%q format=%s data=%q",
		msg.Origin, msg.Format, string(msg.Data))

	delivered := &message.ClipboardMessage{
		Origin: msg.Origin,
		Format: msg.Format,
		Data:   append([]byte(nil), msg.Data...),
	}

	select {
	case s.received <- delivered:
	default:
		s.Fail("received channel full")
	}

	return nil
}
