package testing_tree

import (
	"strings"
	"testing"
)

func (c *Case[Q, R, TestingContext]) RunAll(t *testing.T) {
	cases := c.GetAllCases()

	for _, tt := range cases {
		tt := tt
		t.Run(strings.Join(tt.GetPath(), "/"), func(t *testing.T) {
			tt.Run(t)
		})
	}
}

func (c *Case[Q, R, TestingContext]) RunPath(t *testing.T, path []string) {
	tt, err := c.FindPath(path)
	if err != nil {
		t.Error(err)
		return
	}
	tt.Run(t)
}

func (tt CasePath[Q, R, TestingContext]) Run(t *testing.T) {
	runner := tt.GetRunner()
	if runner == nil {
		t.Errorf("Runner not set")
		return
	}
	var req *Q
	var tc TestingContext
	tctx := &tc
	var itctx interface{} = tctx
	if tctx, ok := itctx.(ITestingAware); ok {
		tctx.OnTestingInit(t)
	}

	for _, tt := range tt {
		if tt.Setup != nil {
			tctx, req = tt.Setup(tctx, req)
		}
	}

	resp, err := runner(tctx, req)
	for _, tt := range tt {
		if tt.Assert != nil {
			tt.Assert(t, tctx, req, resp, err)
		}
	}
}
