package t_tree

import "github.com/xhd2015/data-driven-testing/testing_ctx"

// Node defines a node in the tree of testing cases
// Q: request
// R: response
// TC: testing context
type Node[Q any, R any, TC any] struct {
	ID            string
	ParentID      string          // only required in detached mode
	ParentNode    *Node[Q, R, TC] // optional with ParentID. If both set, they must match
	InheritAssert bool            // by default assert is not inherited
	Description   string
	Tags          []string // for grouping

	Run    func(t testing_ctx.T, tctx *TC, req *Q) (*R, error)
	Setup  func(t testing_ctx.T, tctx *TC, req *Q) (*TC, *Q)
	Assert func(t testing_ctx.T, tctx *TC, req *Q, res *R, err error)

	Children []*Node[Q, R, TC]
}
