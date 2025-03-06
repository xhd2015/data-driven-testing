package decision_tree

// Node represents a single node in the decision tree
type Node struct {
	ID         string         `json:"id"`
	Label      string         `json:"label"`
	Conditions map[string]any `json:"conditions,omitempty"`
	Style      *NodeStyle     `json:"style,omitempty"`
	Children   []*Node        `json:"children,omitempty"`
}

// NodeStyle defines visual properties for a node
type NodeStyle struct {
	Shape       string `json:"shape,omitempty"`       // rectangle, diamond
	Fill        string `json:"fill,omitempty"`        // CSS color
	Stroke      string `json:"stroke,omitempty"`      // CSS color
	StrokeWidth int    `json:"strokeWidth,omitempty"` // line width
}

// Config holds configuration for tree rendering
type Config struct {
	// Layout configuration
	LevelHeight        float64 // Vertical distance between levels
	NodeSpacing        float64 // Minimum horizontal space between nodes
	NodePadding        float64 // Text padding inside nodes
	BaseNodeWidth      float64 // Minimum node width
	LeafNodeSpacing    float64 // Horizontal spacing between leaf nodes (default: 1.5 * NodeSpacing)
	ParentChildSpacing float64 // Minimum vertical space between parent and child nodes
	VerticalSpanCoeff  float64 // Coefficient for vertical spacing scaling with children span (0.0-1.0)

	// Default styles
	DefaultStyle *NodeStyle
}

// layout algorithm
// 1. BFS to get all leaf nodes, this determins span of the tree
// 2. for each node, the larger children span, the more vertical space it gets
// 3. nodes at same level is top aligned

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		LevelHeight:        20,     // Base height between levels
		NodeSpacing:        20,     // Increased horizontal spacing for better readability
		NodePadding:        2,      // Slightly reduced padding
		BaseNodeWidth:      120,    // Reduced base width for better compactness
		LeafNodeSpacing:    10,     // 1.5 * NodeSpacing for leaf nodes
		ParentChildSpacing: 10,     // Minimum space between parent and child nodes
		VerticalSpanCoeff:  0.0009, // 15% of children's horizontal span added to vertical spacing
		DefaultStyle: &NodeStyle{
			Shape:       "rectangle",
			Fill:        "url(#nodeGradient)",
			Stroke:      "#666666",
			StrokeWidth: 1,
		},
	}
}

// Clone creates a deep copy of the node and its children
func (n *Node) Clone() *Node {
	if n == nil {
		return nil
	}

	clone := &Node{
		ID:         n.ID,
		Label:      n.Label,
		Conditions: make(map[string]any, len(n.Conditions)),
	}

	// Copy conditions
	for k, v := range n.Conditions {
		clone.Conditions[k] = v
	}

	// Copy style if exists
	if n.Style != nil {
		clone.Style = &NodeStyle{
			Shape:       n.Style.Shape,
			Fill:        n.Style.Fill,
			Stroke:      n.Style.Stroke,
			StrokeWidth: n.Style.StrokeWidth,
		}
	}

	// Clone children
	if len(n.Children) > 0 {
		clone.Children = make([]*Node, len(n.Children))
		for i, child := range n.Children {
			clone.Children[i] = child.Clone()
		}
	}

	return clone
}
