package t_tree

import (
	"strings"
	"testing"
)

func TestToMermaid(t *testing.T) {
	// Create a simple tree for testing
	root := &Node[string, string, string]{
		ID:          "root",
		Description: "Root Node",
		Children: []*Node[string, string, string]{
			{
				ID:          "child1",
				Description: "Child 1",
				Children: []*Node[string, string, string]{
					{
						ID:          "grandchild1",
						Description: "Grandchild 1",
					},
				},
			},
			{
				ID:          "child2",
				Description: "Child 2",
			},
			{
				// Node with only ID, no description
				ID: "child3",
			},
			{
				// Node with same ID and description
				ID:          "child4",
				Description: "child4",
			},
		},
	}

	tree := &Tree[string, string, string]{
		Root: root,
	}
	tree.init() // Initialize the tree

	// Generate Mermaid diagram
	mermaid := tree.ToMermaid()

	// Verify the output contains expected elements
	t.Logf("Generated Mermaid diagram:\n%s", mermaid)

	// Check that the diagram starts with the correct header
	if !strings.HasPrefix(mermaid, "graph TD;") {
		t.Errorf("Mermaid diagram should start with 'graph TD;', got: %s", mermaid)
	}

	// Check that all nodes are included with correct formatting
	// Root node should be formatted as a rounded rectangle with parentheses
	if !strings.Contains(mermaid, "root(\"root<br><i>Root Node</i>\")") {
		t.Errorf("Root node not formatted correctly in Mermaid diagram")
	}

	// Internal nodes (with children) should be formatted with curly braces
	if !strings.Contains(mermaid, "child1{\"child1<br><i>Child 1</i>\"}") {
		t.Errorf("Internal node not formatted correctly in Mermaid diagram")
	}

	// Leaf nodes should be formatted with square brackets
	if !strings.Contains(mermaid, "grandchild1[\"grandchild1<br><i>Grandchild 1</i>\"]") {
		t.Errorf("Leaf node not formatted correctly in Mermaid diagram")
	}
	if !strings.Contains(mermaid, "child2[\"child2<br><i>Child 2</i>\"]") {
		t.Errorf("Leaf node not formatted correctly in Mermaid diagram")
	}
	if !strings.Contains(mermaid, "child3[\"child3\"]") {
		t.Errorf("Leaf node with only ID not formatted correctly in Mermaid diagram")
	}
	if !strings.Contains(mermaid, "child4[\"child4<br><i>child4</i>\"]") {
		t.Errorf("Leaf node with same ID and description not formatted correctly in Mermaid diagram")
	}

	// Check that connections are included
	expectedConnections := []string{
		"root --> child1",
		"root --> child2",
		"root --> child3",
		"root --> child4",
		"child1 --> grandchild1",
	}
	for _, conn := range expectedConnections {
		if !strings.Contains(mermaid, conn) {
			t.Errorf("Mermaid diagram should contain connection '%s'", conn)
		}
	}

	// Test empty tree
	emptyTree := &Tree[string, string, string]{
		Root: nil,
	}
	emptyMermaid := emptyTree.ToMermaid()
	if emptyMermaid != "graph TD;\n" {
		t.Errorf("Empty tree should generate 'graph TD;\\n', got: %s", emptyMermaid)
	}
}
