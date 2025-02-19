package testing_tree_v2

import (
	"fmt"
	"strings"
	"testing"
)

func (c *Case[Q, R, TestingContext, V]) RunAll(t *testing.T) {
	cases := c.GetAllCases()

	for _, tt := range cases {
		tt := tt
		t.Run(strings.Join(tt.GetPath(), "/"), func(t *testing.T) {
			tt.Run(t)
		})
	}
}

func (c *Case[Q, R, TestingContext, V]) RunPath(t *testing.T, path []string) {
	tt, err := c.FindPath(path)
	if err != nil {
		t.Error(err)
		return
	}
	tt.Run(t)
}

func (c *Case[Q, R, TestingContext, V]) RunPathVariant(t *testing.T, path []string, v V) {
	tt, err := c.FindPath(path)
	if err != nil {
		t.Error(err)
		return
	}
	tt.runWithVariants(t, []V{v})
}

func (tt CasePath[Q, R, TestingContext, V]) Run(t *testing.T) {
	variants := tt.GetVariants()
	if len(variants) == 0 {
		// default to single default variant
		var v V
		variants = []V{v}
	}
	tt.runWithVariants(t, variants)
}

func (tt CasePath[Q, R, TestingContext, V]) runWithVariants(t *testing.T, variants []V) {
	runner := tt.GetRunner()
	if runner == nil {
		t.Errorf("Runner not set")
		return
	}

	runWithT := func(t *testing.T, v V) {
		var req *Q
		var tc TestingContext
		tctx := &tc
		var itctx interface{} = tctx
		if tctx, ok := itctx.(ITestingAware); ok {
			tctx.OnTestingInit(t)
		}

		for _, tt := range tt {
			if tt.Setup != nil {
				tctx, req = tt.Setup(tctx, req, v)
			}
		}

		resp, err := runner(tctx, req, v)
		for _, tt := range tt {
			if tt.Assert != nil {
				tt.Assert(t, tctx, req, v, resp, err)
			}
		}
	}

	if len(variants) == 1 {
		// single variant
		runWithT(t, variants[0])
	} else {
		for _, v := range variants {
			t.Run(fmt.Sprintf("%v", v), func(t *testing.T) {
				runWithT(t, v)
			})
		}
	}
}
