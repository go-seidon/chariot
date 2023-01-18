package restapp

import (
	"context"
)

type Server interface {
	Start(address string) error
	Shutdown(ctx context.Context) error
}
