package testing_tree_v2

import (
	"fmt"
	"strings"
	"testing"
)

type ITestingAware interface {
	OnTestingInit(t *testing.T)
}

type Case[Q any, R any, TestingContext any, V any] struct {
	Name     string
	Variants []V
	Run      func(tctx *TestingContext, req *Q, variant V) (*R, error)
	Setup    func(tctx *TestingContext, req *Q, variant V) (*TestingContext, *Q)
	Assert   func(t *testing.T, tctx *TestingContext, req *Q, variant V, res *R, err error)
	SubCases []*Case[Q, R, TestingContext, V]
}

func (c *Case[Q, R, TestingContext, V]) FindSubCase(name string) *Case[Q, R, TestingContext, V] {
	for _, subCase := range c.SubCases {
		if subCase.Name == name {
			return subCase
		}
	}
	return nil
}

func (c *Case[Q, R, TestingContext, V]) GetAllCases() []CasePath[Q, R, TestingContext, V] {
	if c == nil {
		return nil
	}
	return c.getAllCases(nil)
}

func (c *Case[Q, R, TestingContext, V]) FindPath(path []string) (CasePath[Q, R, TestingContext, V], error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("invalid path")
	}
	if c == nil {
		return nil, fmt.Errorf("root case not found: %s", path[0])
	}
	if c.Name != path[0] {
		return nil, fmt.Errorf("expecting root case: %s, actual: %s", path[0], c.Name)
	}
	casePath := make(CasePath[Q, R, TestingContext, V], 0, len(path)+1)
	casePath = append(casePath, c)
	tt := c
	for i, name := range path[1:] {
		tt = tt.FindSubCase(name)
		if tt == nil {
			return nil, fmt.Errorf("case not found: %s", strings.Join(path[:i+1], "-"))
		}
		casePath = append(casePath, tt)
	}
	return casePath, nil
}

func (c *Case[Q, R, TestingContext, V]) getAllCases(parents CasePath[Q, R, TestingContext, V]) []CasePath[Q, R, TestingContext, V] {
	cases := make([]CasePath[Q, R, TestingContext, V], 0, len(c.SubCases)+1)

	// copy
	casePath := make(CasePath[Q, R, TestingContext, V], len(parents)+1)
	copy(casePath, parents)
	casePath[len(parents)] = c

	// only case with assert
	if c.Assert != nil {
		cases = append(cases, casePath)
	}

	for _, subCase := range c.SubCases {
		subPrefixCases := subCase.getAllCases(casePath)
		cases = append(cases, subPrefixCases...)
	}
	return cases
}
