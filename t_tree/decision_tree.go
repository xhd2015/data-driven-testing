package t_tree

import (
	"errors"

	"github.com/xhd2015/data-driven-testing/decision_tree"
	"github.com/xhd2015/data-driven-testing/decision_tree/svg"
)

// ToDecisionTree converts a t_tree.Tree to a decision_tree.Node
// This allows us to leverage the visualization capabilities of the decision_tree package
func (t *Tree[Q, R, TC]) ToDecisionTree() *decision_tree.Node {
	if t == nil || t.Root == nil {
		return nil
	}
	return convertNode(t.Root)
}

// convertNode converts a t_tree.Node to a decision_tree.Node
func convertNode[Q, R, TC any](node *Node[Q, R, TC]) *decision_tree.Node {
	if node == nil {
		return nil
	}

	// Choose the best label, fallback to ID if description is empty
	label := node.Description
	if label == "" {
		label = node.ID
	}

	dt := &decision_tree.Node{
		ID:    node.ID,
		Label: label,
	}

	// Build structured conditions
	conditions := make(map[string]any)
	if len(node.Tags) > 0 {
		conditions["tags"] = node.Tags
	}
	// Add any additional node metadata if present
	if len(conditions) > 0 {
		dt.Conditions = conditions
	}

	// Convert children with proper validation
	if len(node.Children) > 0 {
		children := make([]*decision_tree.Node, 0, len(node.Children))
		for _, child := range node.Children {
			if dtChild := convertNode(child); dtChild != nil {
				children = append(children, dtChild)
			}
		}
		if len(children) > 0 {
			dt.Children = children
		}
	}

	return dt
}

// ToSVG generates an SVG representation of the tree
func (t *Tree[Q, R, TC]) ToSVG() string {
	dt := t.ToDecisionTree()
	if dt == nil {
		return ""
	}

	renderer := svg.NewRenderer(decision_tree.DefaultConfig())
	return renderer.RenderTree(dt)
}

// ServeSVG generates an SVG representation of the tree and serves it on a local server
func (t *Tree[Q, R, TC]) ServeSVG() error {
	dt := t.ToDecisionTree()
	if dt == nil {
		return errors.New("tree is nil")
	}

	server := svg.NewServer(svg.NewRenderer(decision_tree.DefaultConfig()))
	return server.Serve(dt)
}
