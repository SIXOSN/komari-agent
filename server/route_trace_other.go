//go:build !linux

package server

import (
	"context"
	"errors"
)

func traceTCPRoute(_ context.Context, _ string, _ int) ([]string, error) {
	return nil, errors.New("TCP route tracing is currently supported on Linux agents")
}
