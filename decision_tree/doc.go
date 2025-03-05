// Package decision_tree provides functionality to create and render decision trees as SVG.
//
// The package supports:
// - JSON-based tree definition
// - Automatic layout calculation
// - SVG rendering with proper spacing and connections
// - Custom node styling
//
// Example usage:
//
//	tree := &Node{
//		ID:    "root",
//		Label: "Start",
//		Children: []*Node{
//			{ID: "child1", Label: "Option 1"},
//			{ID: "child2", Label: "Option 2"},
//		},
//	}
//
//	renderer := NewRenderer(DefaultConfig())
//	svg := renderer.RenderTree(tree)
//
// RoadMap:
// - [ ] Add background color for terminal nodes
// - [ ] Add extra legend
// - [ ] Adjust font size for conditions
// - [ ] Make terminal node size more appropriate
package decision_tree
