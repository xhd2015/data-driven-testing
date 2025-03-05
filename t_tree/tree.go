package t_tree

import (
	"fmt"

	"github.com/xhd2015/data-driven-testing/testing_ctx"
)

type ITestingAware interface {
	OnTestingInit(t testing_ctx.T)
}

type Tree[Q any, R any, TC any] struct {
	Root *Node[Q, R, TC]

	buildingNodeToInternalNode map[*Node[Q, R, TC]]*Node[Q, R, TC]
	childToParent              map[*Node[Q, R, TC]]*Node[Q, R, TC]
	idToNode                   map[string]*Node[Q, R, TC]
}

func MustBuild[Q any, R any, TC any](root *Node[Q, R, TC], nodes []*Node[Q, R, TC]) *Tree[Q, R, TC] {
	tree, err := Build[Q, R, TC](root, nodes)
	if err != nil {
		panic(err)
	}
	return tree
}

// Build builds a tree from a list of nodes.
// The root node is the node without parent
// If multiple nodes defined no
func Build[Q any, R any, TC any](root *Node[Q, R, TC], nodes []*Node[Q, R, TC]) (*Tree[Q, R, TC], error) {
	if root == nil {
		return nil, fmt.Errorf("root is nil")
	}
	allNodes := append([]*Node[Q, R, TC]{root}, nodes...)
	nameMapping := make(map[string]*Node[Q, R, TC], len(allNodes))

	// build id and check conflict
	var buildIDMapping func(node *Node[Q, R, TC]) error
	buildIDMapping = func(node *Node[Q, R, TC]) error {
		if node.ID != "" {
			if _, ok := nameMapping[node.ID]; ok {
				// TODO: add code source information
				return fmt.Errorf("duplicate node: %s", node.ID)
			}
			nameMapping[node.ID] = node
		}
		for _, child := range node.Children {
			if err := buildIDMapping(child); err != nil {
				return err
			}
		}
		return nil
	}

	// deep copy all nodes for modification
	buildingNodeToInternalNode := make(map[*Node[Q, R, TC]]*Node[Q, R, TC])
	var copyNode func(node *Node[Q, R, TC]) *Node[Q, R, TC]
	copyNode = func(node *Node[Q, R, TC]) *Node[Q, R, TC] {
		cpNode := *node
		buildingNodeToInternalNode[node] = &cpNode
		cpNode.Children = make([]*Node[Q, R, TC], len(node.Children))
		for i, child := range node.Children {
			cpNode.Children[i] = copyNode(child)
		}
		return &cpNode
	}

	copiedNodes := copyNode(&Node[Q, R, TC]{Children: allNodes})
	err := buildIDMapping(copiedNodes)
	if err != nil {
		return nil, err
	}

	buildingRoot := copiedNodes.Children[0]

	// build parent-child relation on flat nodes
	for _, node := range copiedNodes.Children[1:] {
		var nodeParent *Node[Q, R, TC]
		if node.ParentID == "" && node.ParentNode == nil {
			nodeParent = buildingRoot
		} else {
			var nodeParentByID *Node[Q, R, TC]
			var nodeParentByNode *Node[Q, R, TC]
			if node.ParentID != "" {
				var ok bool
				nodeParentByID, ok = nameMapping[node.ParentID]
				if !ok {
					return nil, fmt.Errorf("missing parent for: %s(%s), parentID: %s", node.ID, node.Description, node.ParentID)
				}
				nodeParent = nodeParentByID
			}
			if node.ParentNode != nil {
				nodeParentByNode = buildingNodeToInternalNode[node.ParentNode]
				if nodeParentByNode == nil {
					return nil, fmt.Errorf("missing parent for: %s(%s)", node.ID, node.Description)
				}
				nodeParent = nodeParentByNode
			}
			// if both are set, they must much
			if nodeParentByID != nil && nodeParentByNode != nil && nodeParentByID != nodeParentByNode {
				return nil, fmt.Errorf("parent mismatch for: %s(%s), parentID: %s, parentNode: %s", node.ID, node.Description, node.ParentID, node.ParentNode.ID)
			}
		}
		nodeParent.Children = append(nodeParent.Children, node)
	}

	tree := &Tree[Q, R, TC]{Root: buildingRoot}
	tree.buildingNodeToInternalNode = buildingNodeToInternalNode
	tree.init()
	return tree, nil
}

func (c *Tree[Q, R, TC]) init() {
	c.childToParent = make(map[*Node[Q, R, TC]]*Node[Q, R, TC])
	c.idToNode = make(map[string]*Node[Q, R, TC])
	var buildChildToParent func(node *Node[Q, R, TC])
	buildChildToParent = func(node *Node[Q, R, TC]) {
		if node.ID != "" {
			c.idToNode[node.ID] = node
		}
		for _, child := range node.Children {
			c.childToParent[child] = node
			buildChildToParent(child)
		}
	}
	buildChildToParent(c.Root)
}

func (c *Tree[Q, R, TC]) Run(t testing_ctx.T) {
	c.run(t, []*Node[Q, R, TC]{c.Root})
}

func (c *Tree[Q, R, TC]) run(t testing_ctx.T, nodePath []*Node[Q, R, TC]) {
	node := nodePath[len(nodePath)-1]
	id := node.ID
	t.Run(id, func(t testing_ctx.T) {
		if node.Assert != nil || node.AssertSelf != nil {
			c.runPath(t, nodePath)
		}
		for _, child := range node.Children {
			c.run(t, append(nodePath, child))
		}
	})
}

func (c *Tree[Q, R, TC]) FindNode(id string) *Node[Q, R, TC] {
	return c.idToNode[id]
}

func (c *Tree[Q, R, TC]) RunNode(t testing_ctx.T, node *Node[Q, R, TC]) {
	var nodePathReversed []*Node[Q, R, TC]

	internalNode := c.buildingNodeToInternalNode[node]
	if internalNode == nil {
		internalNode = node
	}
	p := internalNode
	for p != c.Root {
		nodePathReversed = append(nodePathReversed, p)
		next := c.childToParent[p]
		if next == nil {
			t.Errorf("missing parent: %s", p.ID)
			return
		}
		p = next
	}
	// last always be the root
	nodePathReversed = append(nodePathReversed, c.Root)

	n := len(nodePathReversed)
	nodePath := make([]*Node[Q, R, TC], n)
	for i := 0; i < n; i++ {
		nodePath[i] = nodePathReversed[n-1-i]
	}
	c.runPath(t, nodePath)
}

func (c *Tree[Q, R, TC]) runPath(t testing_ctx.T, nodePath []*Node[Q, R, TC]) {
	if len(nodePath) == 0 {
		t.Errorf("node path is empty")
		return
	}

	n := len(nodePath)
	var runner func(tctx *TC, req *Q) (*R, error)
	for i := n - 1; i >= 0; i-- {
		if nodePath[i].Run != nil {
			runner = nodePath[i].Run
			break
		}
	}
	if runner == nil {
		t.Errorf("missing runner: %s", nodePath[n-1].ID)
		return
	}

	var req *Q
	var tc TC
	tctx := &tc
	var itctx interface{} = tctx
	if tctx, ok := itctx.(ITestingAware); ok {
		tctx.OnTestingInit(t)
	}
	for i := 0; i < n; i++ {
		if nodePath[i].Setup != nil {
			tctx, req = nodePath[i].Setup(tctx, req)
		}
	}

	resp, err := runner(tctx, req)
	// check the assert chain until inherit is false
	var asserts []func(t testing_ctx.T, tctx *TC, req *Q, resp *R, err error)
	for i := n - 1; i >= 0; i-- {
		nd := nodePath[i]
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
	if nodePath[n-1].AssertSelf != nil {
		nodePath[n-1].AssertSelf(t, tctx, req, resp, err)
	}
}
