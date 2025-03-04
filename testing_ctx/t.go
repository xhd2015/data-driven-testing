package testing_ctx

type T interface {
	Run(name string, f func(t T))
	Logf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Log(args ...interface{})
	Error(args ...interface{})
	Skip(args ...interface{})
}
