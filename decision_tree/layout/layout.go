package layout

import (
	"math"

	"github.com/xhd2015/data-driven-testing/decision_tree"
)

// LayoutNode represents a node with layout information
type LayoutNode struct {
	Node      *decision_tree.Node
	X, Y      float64 // Final coordinates
	Width     float64 // Node width based on text
	Height    float64 // Node height
	Level     int     // Tree depth level
	LeafCount int     // Number of leaf nodes under this node
	Parent    *LayoutNode
	Children  []*LayoutNode // Layout-specific children
	Style     *decision_tree.NodeStyle
	IsLeaf    bool // Whether this is a leaf node
	LeafIndex int  // Sequential index for leaf nodes
}

// LevelInfo holds information about nodes at a specific level
type LevelInfo struct {
	nodes     []*LayoutNode
	leafCount int     // Total number of leaf nodes at this level
	width     float64 // Total width required for this level
}

// Engine handles the layout calculation
type Engine struct {
	config       *decision_tree.Config
	centerParent bool
	leafNodes    []*LayoutNode // All leaf nodes in left-to-right order
}

// NewEngine creates a new layout engine
func NewEngine(config *decision_tree.Config) *Engine {
	return &Engine{
		config:       config,
		centerParent: true,
	}
}

// SetCenterParent sets whether parents should be centered over their children
func (e *Engine) SetCenterParent(center bool) {
	e.centerParent = center
}

// CalculateLayout computes the layout for the entire tree
func (e *Engine) CalculateLayout(root *decision_tree.Node) *LayoutNode {
	if root == nil {
		return nil
	}

	e.leafNodes = nil // Reset leaf nodes

	// Create initial layout tree
	layoutRoot := e.createBasicLayoutTree(root, nil, 0)

	// Collect and index leaf nodes
	e.collectLeafNodes(layoutRoot)

	// Group nodes by level
	levels := e.groupNodesByLevel(layoutRoot)

	// Calculate dimensions
	e.calculateDimensions(levels)

	// Assign coordinates with sequential leaf distribution
	e.assignCoordinatesWithLeafOrder(levels)

	return layoutRoot
}

// createBasicLayoutTree creates the initial layout tree with basic properties
func (e *Engine) createBasicLayoutTree(node *decision_tree.Node, parent *LayoutNode, level int) *LayoutNode {
	if node == nil {
		return nil
	}

	layoutNode := &LayoutNode{
		Node:   node,
		Level:  level,
		Parent: parent,
		Style:  node.Style,
	}

	// Create layout nodes for children
	if len(node.Children) > 0 {
		layoutNode.Children = make([]*LayoutNode, len(node.Children))
		for i, child := range node.Children {
			layoutNode.Children[i] = e.createBasicLayoutTree(child, layoutNode, level+1)
		}
	}

	return layoutNode
}

// groupNodesByLevel groups nodes by their level using DFS
func (e *Engine) groupNodesByLevel(root *LayoutNode) map[int]*LevelInfo {
	levels := make(map[int]*LevelInfo)
	if root == nil {
		return levels
	}

	var visit func(*LayoutNode)
	visit = func(node *LayoutNode) {
		if levels[node.Level] == nil {
			levels[node.Level] = &LevelInfo{}
		}
		levels[node.Level].nodes = append(levels[node.Level].nodes, node)

		for _, child := range node.Children {
			visit(child)
		}
	}
	visit(root)

	return levels
}

// calculateDimensions calculates node dimensions and leaf counts bottom-up
func (e *Engine) calculateDimensions(levels map[int]*LevelInfo) {
	maxLevel := 0
	for level := range levels {
		if level > maxLevel {
			maxLevel = level
		}
	}

	// Bottom-up calculation
	for level := maxLevel; level >= 0; level-- {
		info := levels[level]
		for _, node := range info.nodes {
			// Calculate node dimensions
			e.calculateNodeDimensions(node)

			// Calculate leaf count
			if len(node.Children) == 0 {
				node.LeafCount = 1
				info.leafCount++
			} else {
				for _, child := range node.Children {
					node.LeafCount += child.LeafCount
				}
			}
		}

		// Calculate total width required for this level
		info.width = e.calculateLevelWidth(info.nodes)
	}
}

// calculateNodeDimensions calculates dimensions for a single node
func (e *Engine) calculateNodeDimensions(node *LayoutNode) {
	const maxChars = 20

	// Base height
	height := 40.0

	// Add space for conditions
	if len(node.Node.Conditions) > 0 {
		conditionLines := (len(node.Node.Conditions) + 1) / 2
		height += float64(conditionLines) * 15
	}

	// Add height for wrapped label
	labelLines := (len(node.Node.Label) + maxChars - 1) / maxChars
	if labelLines > 1 {
		height += float64(labelLines-1) * 15
	}

	node.Height = height

	// Calculate width with more space
	labelWidth := math.Min(float64(len(node.Node.Label))*9, float64(maxChars)*9)
	node.Width = math.Max(labelWidth+e.config.NodePadding*2, e.config.BaseNodeWidth)
}

// calculateLevelWidth calculates total width required for a level
func (e *Engine) calculateLevelWidth(nodes []*LayoutNode) float64 {
	if len(nodes) == 0 {
		return 0
	}

	totalWidth := 0.0
	for _, node := range nodes {
		totalWidth += node.Width
	}

	// Add spacing between nodes
	totalWidth += e.config.NodeSpacing * float64(len(nodes)-1)

	return totalWidth
}

// collectLeafNodes collects all leaf nodes in left-to-right order
func (e *Engine) collectLeafNodes(node *LayoutNode) {
	if node == nil {
		return
	}

	if len(node.Children) == 0 {
		node.IsLeaf = true
		node.LeafIndex = len(e.leafNodes)
		e.leafNodes = append(e.leafNodes, node)
	}

	for _, child := range node.Children {
		e.collectLeafNodes(child)
	}
}

// assignCoordinatesWithLeafOrder assigns coordinates with sequential leaf distribution
func (e *Engine) assignCoordinatesWithLeafOrder(levels map[int]*LevelInfo) {
	startX := 50.0
	startY := 50.0
	maxLevel := 0
	for level := range levels {
		if level > maxLevel {
			maxLevel = level
		}
	}

	// Calculate Y coordinates for each level first
	levelY := make(map[int]float64)
	levelY[0] = startY

	// First pass: position leaf nodes sequentially to calculate their positions
	leafSpacing := e.config.LeafNodeSpacing
	if leafSpacing == 0 {
		leafSpacing = e.config.NodeSpacing * 1.5 // Fallback to default multiplier
	}
	currentLeafX := startX
	for _, leaf := range e.leafNodes {
		leaf.X = currentLeafX + leaf.Width/2
		currentLeafX = leaf.X + leaf.Width/2 + leafSpacing
	}

	// Second pass: position non-leaf nodes bottom-up and calculate their positions
	for level := maxLevel - 1; level >= 0; level-- {
		info, exists := levels[level]
		if !exists {
			continue
		}

		for _, node := range info.nodes {
			if !node.IsLeaf && len(node.Children) > 0 {
				// Center parent over its children's span
				leftMost := node.Children[0].X - node.Children[0].Width/2
				rightMost := node.Children[len(node.Children)-1].X + node.Children[len(node.Children)-1].Width/2
				node.X = leftMost + (rightMost-leftMost)/2
			}
		}
	}

	// Third pass: calculate Y coordinates with dynamic spacing
	for level := 1; level <= maxLevel; level++ {
		prevLevelMaxHeight := 0.0
		if prevInfo, exists := levels[level-1]; exists {
			for _, node := range prevInfo.nodes {
				if node.Height > prevLevelMaxHeight {
					prevLevelMaxHeight = node.Height
				}
			}
		}

		// Calculate spacing based on parent nodes' children span
		maxSpacing := e.config.ParentChildSpacing // minimum spacing
		if prevInfo, exists := levels[level-1]; exists {
			for _, node := range prevInfo.nodes {
				if len(node.Children) > 0 {
					// Calculate children's horizontal span
					leftMost := node.Children[0].X - node.Children[0].Width/2
					rightMost := node.Children[len(node.Children)-1].X + node.Children[len(node.Children)-1].Width/2
					span := rightMost - leftMost

					// Calculate total spacing based on span and coefficient
					spacing := e.config.ParentChildSpacing * (1 + span*e.config.VerticalSpanCoeff)
					if spacing > maxSpacing {
						maxSpacing = spacing
					}
				}
			}
		}

		// Apply the spacing
		levelY[level] = levelY[level-1] + prevLevelMaxHeight + maxSpacing
	}

	// Final pass: set Y coordinates and adjust horizontal spacing
	for level := 0; level <= maxLevel; level++ {
		info, exists := levels[level]
		if !exists {
			continue
		}

		// Set Y coordinates for all nodes at this level
		for _, node := range info.nodes {
			node.Y = levelY[level]
		}

		// Adjust horizontal spacing
		nodes := info.nodes
		for i := 1; i < len(nodes); i++ {
			prev := nodes[i-1]
			curr := nodes[i]
			minSpacing := (prev.Width+curr.Width)/2 + e.config.NodeSpacing
			if curr.X-prev.X < minSpacing {
				shift := minSpacing - (curr.X - prev.X)
				e.shiftSubtree(curr, shift)
			}
		}
	}
}

func (e *Engine) shiftSubtree(node *LayoutNode, shift float64) {
	if node == nil {
		return
	}
	node.X += shift
	for _, child := range node.Children {
		e.shiftSubtree(child, shift)
	}
}
