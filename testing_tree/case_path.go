package testing_tree

type CasePath[Q any, R any, TestingContext any] []*Case[Q, R, TestingContext]

func (c CasePath[Q, R, TestingContext]) GetRunner() func(tctx *TestingContext, req *Q) (*R, error) {
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
