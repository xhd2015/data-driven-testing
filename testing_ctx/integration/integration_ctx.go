package integration

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/xhd2015/data-driven-testing/testing_ctx"
)

var errFatal = errors.New("FATAL")

type IntegrationContext struct {
	name    string
	isError bool
	indent  string
	isLast  bool // tracks if this is the last test in its group
	parent  *IntegrationContext
	isSkip  bool

	infoWriter io.Writer
	errWriter  io.Writer

	startedAt  time.Time
	finishedAt time.Time

	context context.Context
}

var _ testing_ctx.T = &IntegrationContext{}
var _ testing_ctx.ContextAware = &IntegrationContext{}

func New() *IntegrationContext {
	return &IntegrationContext{
		infoWriter: os.Stdout,
		errWriter:  os.Stdout,
	}
}

func WithOptions(opts Options) *IntegrationContext {
	t := New()
	if opts.InfoWriter != nil {
		t.infoWriter = opts.InfoWriter
	}
	if opts.ErrWriter != nil {
		t.errWriter = opts.ErrWriter
	}
	return t
}

func (t *IntegrationContext) getPrefix() string {
	if t.parent == nil {
		return ""
	}

	prefix := ""
	for p := t.parent; p != nil && p.indent != ""; p = p.parent {
		if p.isLast {
			prefix = "    " + prefix
		} else {
			prefix = "│   " + prefix
		}
	}

	if t.isLast {
		return prefix + "└── "
	}
	return prefix + "├── "
}

// Errorf implements testing_ctx.T.
func (t *IntegrationContext) Errorf(format string, args ...interface{}) {
	t.isError = true
	// get file and line
	_, file, line, _ := runtime.Caller(1)
	fmt.Fprint(t.errWriter, t.getPrefix())
	fmt.Fprint(t.errWriter, filepath.Base(file))
	fmt.Fprint(t.errWriter, ":")
	fmt.Fprint(t.errWriter, line)
	fmt.Fprint(t.errWriter, ": ")
	fmt.Fprintf(t.errWriter, format, args...)
	fmt.Fprintln(t.errWriter)
}

// Logf implements testing_ctx.T.
func (t *IntegrationContext) Logf(format string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	fmt.Fprint(t.infoWriter, t.getPrefix())
	fmt.Fprint(t.infoWriter, filepath.Base(file))
	fmt.Fprint(t.infoWriter, ":")
	fmt.Fprint(t.infoWriter, line)
	fmt.Fprint(t.infoWriter, ": ")
	fmt.Fprintf(t.infoWriter, format, args...)
	fmt.Fprintln(t.infoWriter)
}

// Fatalf implements testing_ctx.T.
func (t *IntegrationContext) Fatalf(format string, args ...interface{}) {
	t.isError = true
	_, file, line, _ := runtime.Caller(1)
	fmt.Fprint(t.errWriter, t.getPrefix())
	fmt.Fprint(t.errWriter, filepath.Base(file))
	fmt.Fprint(t.errWriter, ":")
	fmt.Fprint(t.errWriter, line)
	fmt.Fprint(t.errWriter, ": ")
	fmt.Fprintf(t.errWriter, format, args...)
	fmt.Fprintln(t.errWriter)
	panic(errFatal)
}

// Error implements testing_ctx.T.
func (t *IntegrationContext) Error(args ...interface{}) {
	t.isError = true
	_, file, line, _ := runtime.Caller(1)
	fmt.Fprint(t.errWriter, t.getPrefix())
	fmt.Fprint(t.errWriter, filepath.Base(file))
	fmt.Fprint(t.errWriter, ":")
	fmt.Fprint(t.errWriter, line)
	fmt.Fprint(t.errWriter, ": ")
	fmt.Fprintln(t.errWriter, args...)
}

// Log implements testing_ctx.T.
func (t *IntegrationContext) Log(args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	fmt.Fprint(t.infoWriter, t.getPrefix())
	fmt.Fprint(t.infoWriter, filepath.Base(file))
	fmt.Fprint(t.infoWriter, ":")
	fmt.Fprint(t.infoWriter, line)
	fmt.Fprint(t.infoWriter, ": ")
	fmt.Fprintln(t.infoWriter, args...)
}

// Fatal implements testing_ctx.T.
func (t *IntegrationContext) Fatal(args ...interface{}) {
	t.isError = true
	_, file, line, _ := runtime.Caller(1)
	fmt.Fprint(t.errWriter, t.getPrefix())
	fmt.Fprint(t.errWriter, filepath.Base(file))
	fmt.Fprint(t.errWriter, ":")
	fmt.Fprint(t.errWriter, line)
	fmt.Fprint(t.errWriter, ": ")
	fmt.Fprintln(t.errWriter, args...)
	panic(errFatal)
}

// Skip implements testing_ctx.T.
func (t *IntegrationContext) Skip(args ...interface{}) {
	t.isSkip = true
	_, file, line, _ := runtime.Caller(1)
	fmt.Fprint(t.infoWriter, t.getPrefix())
	fmt.Fprint(t.infoWriter, filepath.Base(file))
	fmt.Fprint(t.infoWriter, ":")
	fmt.Fprint(t.infoWriter, line)
	fmt.Fprint(t.infoWriter, ": SKIP ")
	fmt.Fprintln(t.infoWriter, args...)
}

func (t *IntegrationContext) Status() testing_ctx.Status {
	if t.startedAt.IsZero() {
		return testing_ctx.StatusNone
	}
	if t.isSkip {
		return testing_ctx.StatusSkip
	}
	if t.isError {
		return testing_ctx.StatusFail
	}
	if t.finishedAt.IsZero() {
		return testing_ctx.StatusRunning
	}
	return testing_ctx.StatusPass
}

func (t *IntegrationContext) Run(name string, f func(t testing_ctx.T)) {
	startTime := time.Now()
	defer func() {
		t.finishedAt = time.Now()
	}()
	t.startedAt = startTime
	subT := &IntegrationContext{
		name:       name,
		indent:     t.indent + "  ",
		parent:     t,
		isLast:     true, // assume it's last by default
		infoWriter: t.infoWriter,
		errWriter:  t.errWriter,
		context:    t.context,
	}

	// If there's a previous test running at this level, mark it as not last
	if t.parent != nil {
		prevSibling := t
		prevSibling.isLast = false
	}

	defer func() {
		if e := recover(); e != nil {
			subT.isError = true
			if e == errFatal {
				return
			}
			fmt.Fprint(t.errWriter, t.getPrefix())
			fmt.Fprintf(t.errWriter, "panic: %v\n", e)
			stack := debug.Stack()
			fmt.Fprint(t.errWriter, string(stack))
		}
		pass := "PASS"
		if subT.isError {
			pass = "FAIL"
			t.isError = true
		} else if subT.isSkip {
			pass = "SKIP"
		}
		fmt.Fprint(t.infoWriter, t.getPrefix())
		fmt.Fprintf(t.infoWriter, "%s %s (%s)\n", pass, name, fmtTime(time.Since(startTime)))
	}()
	fmt.Fprint(t.infoWriter, t.getPrefix())
	fmt.Fprintf(t.infoWriter, "RUN %s\n", name)
	f(subT)
}

func fmtTime(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", int(d.Microseconds()))
	}
	if d < 1*time.Second {
		return fmt.Sprintf("%dms", int(d.Milliseconds()))
	}
	if d < 1*time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
}

// Context implements testing_ctx.ContextAware.
func (t *IntegrationContext) Context() context.Context {
	return t.context
}

// SetContext implements testing_ctx.ContextAware.
func (t *IntegrationContext) SetContext(ctx context.Context) {
	t.context = ctx
}
