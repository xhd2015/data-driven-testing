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
