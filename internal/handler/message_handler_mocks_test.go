package handler_test

import (
	"testing"

	"github.com/jasonkwh/shengos-daemon/internal/adapter"
	"github.com/jasonkwh/shengos-daemon/internal/handler"
	mock_adapter "github.com/jasonkwh/shengos-daemon/mocks/adapter"
	"go.uber.org/zap"
)

type testMocks struct {
	mockClipboard *mock_adapter.MockClipboard
}

func newMockMessageHandler(
	t *testing.T,
) (adapter.DaemonMessageHandler, *testMocks) {
	tm := &testMocks{
		mockClipboard: mock_adapter.NewMockClipboard(t),
	}

	return handler.NewDaemonMessageHandler(
		tm.mockClipboard,
		handler.WithMessageLogger(zap.NewNop()),
	), tm
}
