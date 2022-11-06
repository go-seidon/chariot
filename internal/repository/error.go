package repository

import (
	"errors"
)

var (
	ErrInvalidParam = errors.New("invalid param")
	ErrNotFound     = errors.New("resource not found")
	ErrDeleted      = errors.New("resource deleted")
	ErrExists       = errors.New("resource already exists")
)
