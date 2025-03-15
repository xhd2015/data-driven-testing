package integration

import "io"

type Options struct {
	InfoWriter io.Writer
	ErrWriter  io.Writer
}
