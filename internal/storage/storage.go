package storage

import "errors"

var (
	ErrURLNotFound   = errors.New("url not found")
	ErrURLExists     = errors.New("url exists")
	ErrEventNotFound = errors.New("no new events")
)
