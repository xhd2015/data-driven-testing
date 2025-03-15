package testing_ctx

import "context"

type T interface {
	Run(name string, f func(t T))
	Logf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Log(args ...interface{})
	Error(args ...interface{})
	Skip(args ...interface{})
}

// ContextAware is additional interface for T that
// allows setting and getting the context
type ContextAware interface {
	SetContext(ctx context.Context)
	Context() context.Context
}
