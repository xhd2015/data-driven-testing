package testing_tree_v2

type CasePath[Q any, R any, TestingContext any, V any] []*Case[Q, R, TestingContext, V]

func (c CasePath[Q, R, TestingContext, V]) GetRunner() func(tctx *TestingContext, req *Q, variant V) (*R, error) {
	n := len(c)
	for i := n - 1; i >= 0; i-- {
		if c[i].Run != nil {
			return c[i].Run
		}
	}
	return nil
}

func (c CasePath[Q, R, TestingContext, V]) GetPath() []string {
	path := make([]string, 0, len(c))
	for _, tt := range c {
		path = append(path, tt.Name)
	}
	return path
}

func (c CasePath[Q, R, TestingContext, V]) GetVariants() []V {
	n := len(c)
	for i := n - 1; i >= 0; i-- {
		if len(c[i].Variants) > 0 {
			return c[i].Variants
		}
	}
	return nil
}
