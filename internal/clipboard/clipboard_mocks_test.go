package clipboard_test

import (
	"testing"

	"github.com/jasonkwh/shengos-daemon/internal/adapter"
	"github.com/jasonkwh/shengos-daemon/internal/clipboard"
	mock_adapter "github.com/jasonkwh/shengos-daemon/mocks/adapter"
)

type testMocks struct {
	mockDaemonBus *mock_adapter.MockDaemonBus
}

func newMockClipboard(
	t *testing.T,
	opts ...clipboard.ClipboardOption,
) (adapter.Clipboard, *testMocks, error) {
	tm := &testMocks{
		mockDaemonBus: mock_adapter.NewMockDaemonBus(t),
	}

	clb, err := clipboard.New(
		sampleClipboardConfig,
		tm.mockDaemonBus,
		sampleHostname,
		opts...)

	return clb, tm, err
}
