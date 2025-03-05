package layout

import (
	"math"

	"github.com/xhd2015/data-driven-testing/decision_tree"
)

// LayoutNode represents a node with layout information
type LayoutNode struct {
	Node         *decision_tree.Node
	X, Y         float64 // Final coordinates
	Width        float64 // Node width based on text
	Height       float64 // Node height
	Level        int     // Tree depth level
	SubtreeWidth float64 // Width of entire subtree
	Parent       *LayoutNode
	Children     []*LayoutNode // Layout-specific children
	Style        *decision_tree.NodeStyle
}

// Engine handles the layout calculation
type Engine struct {
	config       *decision_tree.Config
	centerParent bool // when true, centers parent nodes over their children
}

// NewEngine creates a new layout engine
func NewEngine(config *decision_tree.Config) *Engine {
	return &Engine{
		config:       config,
		centerParent: true, // default to true for backward compatibility
	}
}

// SetCenterParent sets whether parents should be centered over their children
func (e *Engine) SetCenterParent(center bool) {
	e.centerParent = center
}

// CalculateLayout computes the layout for the entire tree
func (e *Engine) CalculateLayout(root *decision_tree.Node) *LayoutNode {
	// Create layout tree
	layoutRoot := e.createLayoutTree(root, nil, 0)

	// First pass: calculate sizes and subtree widths
	e.calculateSizes(layoutRoot)

	// Second pass: assign coordinates starting from left
	initialX := 50.0 // Fixed left margin
	e.assignCoordinates(layoutRoot, initialX, 0)

	return layoutRoot
}

// createLayoutTree creates a layout tree from the input tree
func (e *Engine) createLayoutTree(node *decision_tree.Node, parent *LayoutNode, level int) *LayoutNode {
	if node == nil {
		return nil
	}

	// Calculate node height based on content
	height := 30.0 // Reduced base height
	if len(node.Conditions) > 0 {
		if len(node.Children) == 0 {
			// For terminal nodes, conditions go below
			height += 20 // Reduced extra space for conditions below
		} else {
			// For non-terminal nodes, conditions go above
			height += 15 // Reduced extra space for conditions above
		}
	}

	layoutNode := &LayoutNode{
		Node:   node,
		Level:  level,
		Parent: parent,
		Height: height,
		Style:  node.Style,
	}

	// Create layout nodes for children
	if len(node.Children) > 0 {
		layoutNode.Children = make([]*LayoutNode, len(node.Children))
		for i, child := range node.Children {
			layoutNode.Children[i] = e.createLayoutTree(child, layoutNode, level+1)
		}
	}

	return layoutNode
}

// calculateSizes computes node and subtree dimensions
func (e *Engine) calculateSizes(node *LayoutNode) {
	if node == nil {
		return
	}

	// Calculate node width based on content with max width limit
	const maxChars = 20 // Maximum characters per line
	labelLen := len(node.Node.Label)
	var labelWidth float64
	if labelLen > maxChars {
		// If label is longer than maxChars, we'll need multiple lines
		lines := (labelLen + maxChars - 1) / maxChars // Round up division
		labelWidth = float64(maxChars) * 8            // Reduced character width
		node.Height += float64(lines-1) * 15          // Reduced line height
	} else {
		labelWidth = float64(labelLen) * 8 // Reduced character width
	}

	// Calculate width needed for conditions
	var conditionsWidth float64
	if len(node.Node.Conditions) > 0 {
		// Estimate conditions text width
		condText := 0
		for k, v := range node.Node.Conditions {
			// Rough estimate of characters needed
			condText += len(k) + 5 // 5 for "=value"
			switch v := v.(type) {
			case string:
				condText += len(v)
			default:
				condText += 5 // Assume 5 chars for numbers etc
			}
		}
		// Apply max width to conditions as well
		if condText > maxChars {
			conditionsWidth = float64(maxChars) * 6 // Reduced font size for conditions
			node.Height += 15                       // Reduced extra height for wrapped conditions
		} else {
			conditionsWidth = float64(condText) * 6
		}
	}

	// Use the larger of label width, conditions width, or base width
	node.Width = math.Max(math.Max(labelWidth, conditionsWidth), e.config.BaseNodeWidth)
	node.Width = math.Min(node.Width+e.config.NodePadding*2, float64(maxChars)*8+e.config.NodePadding*2) // Add padding but respect max width

	// Process children
	var totalChildrenWidth float64
	var maxChildWidth float64

	for _, child := range node.Children {
		e.calculateSizes(child)
		totalChildrenWidth += child.SubtreeWidth
		if child.SubtreeWidth > maxChildWidth {
			maxChildWidth = child.SubtreeWidth
		}
	}

	// Add spacing between children
	if len(node.Children) > 1 {
		totalChildrenWidth += e.config.NodeSpacing * float64(len(node.Children)-1)
	}

	// Subtree width is the max of node width and total children width
	node.SubtreeWidth = math.Max(node.Width, totalChildrenWidth)

	// Add extra padding for better separation between subtrees
	if len(node.Children) > 0 {
		node.SubtreeWidth += e.config.NodeSpacing * 0.5 // Reduced extra padding
	}
}

// assignCoordinates assigns final x,y coordinates to nodes
func (e *Engine) assignCoordinates(node *LayoutNode, x, y float64) {
	if node == nil {
		return
	}

	if len(node.Children) == 0 {
		node.X = x
		node.Y = y
		return
	}

	// First position all children from left to right
	startX := x // Start from parent's x position
	for _, child := range node.Children {
		e.assignCoordinates(child, startX, y+e.config.LevelHeight)
		startX += child.SubtreeWidth + e.config.NodeSpacing
	}

	// Then position the parent node
	if e.centerParent && len(node.Children) > 0 {
		// Center parent over its children span
		firstChild := node.Children[0]
		lastChild := node.Children[len(node.Children)-1]
		node.X = firstChild.X + (lastChild.X-firstChild.X)/2
	} else {
		node.X = x
	}
	node.Y = y
}
