//go:build integration
// +build integration

package integration

import (
	"slices"

	"go.uber.org/multierr"
)

func (s *integrationSuite) gracefulClose() error {
	s.T().Helper()

	var errs error

	for _, svc := range slices.Backward(s.clPool) {
		if err := svc.Close(); err != nil {
			errs = multierr.Append(errs, err)
		}
	}

	return errs
}
