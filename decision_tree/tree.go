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
	LevelHeight   float64 // Vertical distance between levels
	NodeSpacing   float64 // Minimum horizontal space between nodes
	NodePadding   float64 // Text padding inside nodes
	BaseNodeWidth float64 // Minimum node width

	// Default styles
	DefaultStyle *NodeStyle
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		LevelHeight:   60,  // Reduced vertical spacing
		NodeSpacing:   20,  // Reduced horizontal spacing
		NodePadding:   10,  // Reduced node padding
		BaseNodeWidth: 140, // Reduced minimum node width
		DefaultStyle: &NodeStyle{
			Shape:       "rectangle",
			Fill:        "#ffffff",
			Stroke:      "#000000",
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
