package t_tree

import (
	"strings"
	"testing"
)

func TestDecisionTree(t *testing.T) {
	// Create a sample tree for testing
	root := &Node[any, any, any]{
		ID:          "root",
		Description: "Root Node",
		Tags:        []string{"root", "test"},
		Children: []*Node[any, any, any]{
			{
				ID:          "child1",
				Description: "Child 1",
				Tags:        []string{"child", "leaf"},
			},
			{
				ID:          "child2",
				Description: "Child 2",
				Tags:        []string{"child", "branch"},
				Children: []*Node[any, any, any]{
					{
						ID:          "grandchild1",
						Description: "Grandchild 1",
						Tags:        []string{"grandchild", "leaf"},
					},
				},
			},
		},
	}

	tree := &Tree[any, any, any]{Root: root}

	t.Run("ToDecisionTree", func(t *testing.T) {
		dt := tree.ToDecisionTree()
		if dt == nil {
			t.Fatal("expected non-nil decision tree")
		}

		// Verify root node
		if dt.ID != "root" {
			t.Errorf("expected root ID 'root', got %q", dt.ID)
		}
		if dt.Label != "Root Node" {
			t.Errorf("expected root Label 'Root Node', got %q", dt.Label)
		}
		if len(dt.Children) != 2 {
			t.Errorf("expected 2 children, got %d", len(dt.Children))
		}

		// Verify tags are converted to conditions
		if dt.Conditions == nil {
			t.Error("expected non-nil conditions for root")
		} else if tags, ok := dt.Conditions["tags"].([]string); !ok {
			t.Error("expected tags condition to be []string")
		} else if len(tags) != 2 {
			t.Errorf("expected 2 tags, got %d", len(tags))
		}
	})

	t.Run("ToSVG", func(t *testing.T) {
		svg := tree.ToSVG()
		if svg == "" {
			t.Error("expected non-empty SVG")
		}

		// Basic SVG validation
		if !strings.HasPrefix(svg, "<svg") {
			t.Error("expected SVG to start with <svg")
		}
		if !strings.HasSuffix(svg, "</svg>") {
			t.Error("expected SVG to end with </svg>")
		}

		// Check if nodes are present
		if !strings.Contains(svg, "Root Node") {
			t.Error("expected SVG to contain root node label")
		}
		if !strings.Contains(svg, "Child 1") {
			t.Error("expected SVG to contain Child 1 label")
		}
		if !strings.Contains(svg, "Child 2") {
			t.Error("expected SVG to contain Child 2 label")
		}
		if !strings.Contains(svg, "Grandchild 1") {
			t.Error("expected SVG to contain Grandchild 1 label")
		}
	})

	t.Run("ToMermaid", func(t *testing.T) {
		mermaid := tree.ToMermaid()
		if mermaid == "" {
			t.Error("expected non-empty Mermaid diagram")
		}

		// Basic Mermaid validation
		if !strings.HasPrefix(mermaid, "graph TD;") {
			t.Error("expected Mermaid to start with graph TD;")
		}

		// Check if nodes and connections are present
		if !strings.Contains(mermaid, "Root Node") {
			t.Error("expected Mermaid to contain root node label")
		}
		if !strings.Contains(mermaid, "Child 1") {
			t.Error("expected Mermaid to contain Child 1 label")
		}
		if !strings.Contains(mermaid, "-->") {
			t.Error("expected Mermaid to contain node connections")
		}
	})
}
