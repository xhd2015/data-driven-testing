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

	// Calculate base node height
	height := 30.0 // Base height

	// Add space for conditions
	if len(node.Conditions) > 0 {
		conditionLines := (len(node.Conditions) + 1) / 2 // Estimate lines needed for conditions
		if len(node.Children) == 0 {
			// For terminal nodes, conditions go below
			height += float64(conditionLines) * 15
		} else {
			// For non-terminal nodes, conditions go above
			height += float64(conditionLines) * 12
		}
	}

	// Add extra height for long labels
	labelLines := (len(node.Label) + 19) / 20 // 20 chars per line
	if labelLines > 1 {
		height += float64(labelLines-1) * 15
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

	// Calculate adaptive level height based on multiple factors
	adaptiveLevelHeight := e.config.LevelHeight

	// 1. Sibling density factor - more moderate adjustment for nodes with many siblings
	if node.Parent != nil {
		siblingCount := len(node.Parent.Children)
		if siblingCount > 1 {
			// Linear growth with smaller factor, capped at 1.5x
			siblingFactor := math.Min(1.5, 1.0+float64(siblingCount-1)*0.15)
			adaptiveLevelHeight *= siblingFactor
		}
	}

	// 2. Subtree complexity factor - smaller increase for deeper subtrees
	maxDepth := e.calculateSubtreeDepth(node)
	if maxDepth > 1 {
		// Reduced space increase for complex subtrees
		depthFactor := 1.0 + float64(maxDepth-1)*0.1
		adaptiveLevelHeight *= depthFactor
	}

	// 3. Level-based adjustment - gentler reduction
	levelFactor := math.Max(0.9, 1.0-float64(node.Level)*0.03)
	adaptiveLevelHeight *= levelFactor

	// 4. Content-based adjustment - smaller increase for nodes with conditions
	if len(node.Node.Conditions) > 0 {
		adaptiveLevelHeight *= 1.1 // Reduced extra space for conditions
	}

	// Position all children
	startX := x
	for _, child := range node.Children {
		e.assignCoordinates(child, startX, y+adaptiveLevelHeight)
		startX += child.SubtreeWidth + e.config.NodeSpacing
	}

	// Position parent node
	if e.centerParent && len(node.Children) > 0 {
		firstChild := node.Children[0]
		lastChild := node.Children[len(node.Children)-1]
		node.X = firstChild.X + (lastChild.X-firstChild.X)/2
	} else {
		node.X = x
	}
	node.Y = y
}

// calculateSubtreeDepth returns the maximum depth of the subtree
func (e *Engine) calculateSubtreeDepth(node *LayoutNode) int {
	if len(node.Children) == 0 {
		return 1
	}
	maxChildDepth := 0
	for _, child := range node.Children {
		childDepth := e.calculateSubtreeDepth(child)
		if childDepth > maxChildDepth {
			maxChildDepth = childDepth
		}
	}
	return maxChildDepth + 1
}
