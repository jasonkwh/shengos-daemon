package handler_test

import (
	"testing"

	"github.com/go-openapi/testify/v2/require"
	"github.com/jasonkwh/shengos-daemon/internal/adapter"
)

func messagePayload(t *testing.T, msg adapter.DaemonMessage) []byte {
	t.Helper()
	rs := require.New(t)

	payload, err := msg.Payload()
	rs.NoError(err)

	return payload
}
