package testing_ctx

import "context"

type Status int

const (
	StatusNone Status = iota
	StatusRunning
	StatusPass
	StatusFail
	StatusSkip
)

type T interface {
	Run(name string, f func(t T))
	Logf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Log(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Skip(args ...interface{})
	Status() Status
}

// ContextAware is additional interface for T that
// allows setting and getting the context
type ContextAware interface {
	SetContext(ctx context.Context)
	Context() context.Context
}
