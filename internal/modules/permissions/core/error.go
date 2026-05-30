package core

import "errors"

var (
	ErrNotFound        = errors.New("permission not found")
	ErrConflict        = errors.New("permission conflict")
	ErrSystemImmutable = errors.New("system permission is immutable")
)
