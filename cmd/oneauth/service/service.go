package service

import "errors"

var (
	ErrNotInstalled   = errors.New("oneauth service is not installed")
	ErrNotImplemented = errors.New("not yet implemented for your OS")
)
