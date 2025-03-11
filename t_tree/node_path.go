package t_tree

import (
	"github.com/xhd2015/data-driven-testing/testing_ctx"
)

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
	req, tctx := c.SetupTesting(t)
	resp, err := runner(tctx, req)
	c.Assert(t, tctx, req, resp, err)
}

func (c NodePath[Q, R, TC]) Runner() func(tctx *TC, req *Q) (*R, error) {
	n := len(c)
	for i := n - 1; i >= 0; i-- {
		if c[i].Run != nil {
			return c[i].Run
		}
	}
	return nil
}

func (c NodePath[Q, R, TC]) SetupTesting(t testing_ctx.T) (*Q, *TC) {
	return c.setup(t)
}

func (c NodePath[Q, R, TC]) Setup() (*Q, *TC) {
	return c.setup(nil)
}

func (c NodePath[Q, R, TC]) setup(t testing_ctx.T) (*Q, *TC) {
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
			tctx, req = c[i].Setup(tctx, req)
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
	if c[n-1].AssertSelf != nil {
		c[n-1].AssertSelf(t, tctx, req, resp, err)
	}
}
