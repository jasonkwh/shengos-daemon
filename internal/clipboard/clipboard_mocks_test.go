package clipboard

import (
	"testing"

	"github.com/jasonkwh/shengos-daemon/internal/adapter"
	mock_adapter "github.com/jasonkwh/shengos-daemon/mocks/adapter"
)

type testMocks struct {
	mockDaemonBus *mock_adapter.MockDaemonBus
}

func newMockClipboard(
	t *testing.T,
) (adapter.Clipboard, *testMocks, error) {
	tm := &testMocks{
		mockDaemonBus: mock_adapter.NewMockDaemonBus(t),
	}

	clb, err := New(
		sampleClipboardConfig,
		tm.mockDaemonBus,
		sampleHostname,
	)

	return clb, tm, err
}
