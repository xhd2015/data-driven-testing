package t_tree

import "github.com/xhd2015/data-driven-testing/testing_ctx"

// Node defines a node in the tree of testing cases
// Q: request
// R: response
// TC: testing context
type Node[Q any, R any, TC any] struct {
	ID          string
	ParentID    string // only required in detached mode
	Description string

	Run    func(tctx *TC, req *Q) (*R, error)
	Setup  func(tctx *TC, req *Q) (*TC, *Q)
	Assert func(t testing_ctx.T, tctx *TC, req *Q, res *R, err error)

	Children []*Node[Q, R, TC]
}
