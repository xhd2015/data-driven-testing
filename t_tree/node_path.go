package t_tree

import (
	"fmt"
	"runtime/debug"

	"github.com/xhd2015/data-driven-testing/testing_ctx"
)

type PanicError struct {
	Arg   interface{}
	Stack []byte
}

func (c *PanicError) Error() string {
	if pe, ok := c.Arg.(error); ok {
		return pe.Error()
	} else {
		return fmt.Sprintf("panic: %v", c.Arg)
	}
}

func (c *PanicError) Unwrap() error {
	if pe, ok := c.Arg.(error); ok {
		return pe
	} else {
		return fmt.Errorf("panic: %v", c.Arg)
	}
}

type NodePath[Q any, R any, TC any] []*Node[Q, R, TC]

func (c NodePath[Q, R, TC]) Run(t testing_ctx.T) {
	if len(c) == 0 {
		t.Error("node path is empty")
		return
	}
	runner := c.Runner()
	if runner == nil {
		t.Errorf("missing runner: %s", c[len(c)-1].ID)
		return
	}
	req, tctx := c.Setup(t)

	var resp *R
	var err error
	func() {
		defer func() {
			if e := recover(); e != nil {
				err = &PanicError{
					Arg:   e,
					Stack: debug.Stack(),
				}
			}
		}()
		resp, err = runner(t, tctx, req)
	}()

	c.Assert(t, tctx, req, resp, err)
}

func (c NodePath[Q, R, TC]) Runner() func(t testing_ctx.T, tctx *TC, req *Q) (*R, error) {
	n := len(c)
	for i := n - 1; i >= 0; i-- {
		if c[i].Run != nil {
			return c[i].Run
		}
	}
	return nil
}

func (c NodePath[Q, R, TC]) Setup(t testing_ctx.T) (*Q, *TC) {
	var req *Q
	var tc TC
	tctx := &tc
	if t != nil {
		var itctx interface{} = tctx
		if tctx, ok := itctx.(ITestingAware); ok {
			tctx.OnTestingInit(t)
		}
	}
	n := len(c)
	for i := 0; i < n; i++ {
		if c[i].Setup != nil {
			tctx, req = c[i].Setup(t, tctx, req)
		}
	}
	return req, tctx
}

func (c NodePath[Q, R, TC]) Assert(t testing_ctx.T, tctx *TC, req *Q, resp *R, err error) {
	// check the assert chain until inherit is false
	var asserts []func(t testing_ctx.T, tctx *TC, req *Q, resp *R, err error)
	n := len(c)
	for i := n - 1; i >= 0; i-- {
		nd := c[i]
		if nd.Assert != nil {
			asserts = append(asserts, nd.Assert)
		}
		if !nd.InheritAssert {
			break
		}
	}

	for i := len(asserts) - 1; i >= 0; i-- {
		asserts[i](t, tctx, req, resp, err)
	}
}

func (c NodePath[Q, R, TC]) Parent() NodePath[Q, R, TC] {
	n := len(c)
	if n == 0 {
		return nil
	}
	return c[:n-1]
}
