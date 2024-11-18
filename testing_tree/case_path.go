package testing_tree

import (
	"testing"
)

type CasePath[Q any, R any, TestingContext ITestingContext] []*Case[Q, R, TestingContext]

func (c CasePath[Q, R, TestingContext]) GetRunner() func(tctx TestingContext, req *Q) (*R, error) {
	n := len(c)
	for i := n - 1; i >= 0; i-- {
		if c[i].Run != nil {
			return c[i].Run
		}
	}
	return nil
}

func (c CasePath[Q, R, TestingContext]) GetPath() []string {
	path := make([]string, 0, len(c))
	for _, tt := range c {
		path = append(path, tt.Name)
	}
	return path
}

func (tt CasePath[Q, R, TestingContext]) Run(t *testing.T) {
	runner := tt.GetRunner()
	if runner == nil {
		t.Errorf("Runner not set")
		return
	}
	var req *Q
	var tctx TestingContext
	tctx.OnInit(t)

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
