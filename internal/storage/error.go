package storage

import (
	"errors"
)

var (
	ErrUnauthenticated = errors.New("unauthenticated access")
	ErrNotFound        = errors.New("resource not found")
)
