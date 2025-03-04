package t_tree

import (
	"fmt"
	"html"
	"strings"
)

// ToMermaid generates a Mermaid flowchart diagram representation of the tree.
// The diagram uses TD (top-down) layout with nodes and connections.
// Returns a string in Mermaid syntax that can be rendered in Markdown.
func (c *Tree[Q, R, TC]) ToMermaid() string {
	var sb strings.Builder
	sb.WriteString("graph TD;\n")

	// Handle empty tree case
	if c.Root == nil {
		return sb.String()
	}

	// Process nodes recursively
	nodeIDs := make(map[*Node[Q, R, TC]]string)
	processNode(&sb, c.Root, "", nodeIDs)

	return sb.String()
}

// processNode recursively processes a node and its children to build the Mermaid diagram
func processNode[Q, R, TC any](sb *strings.Builder, node *Node[Q, R, TC], parentID string, nodeIDs map[*Node[Q, R, TC]]string) string {
	// Generate or retrieve node ID
	nodeID := getNodeID(node, nodeIDs)

	// Create node label (use description if available, otherwise ID or "Node")
	nodeLabel := getNodeLabel(node)

	// Add node to diagram with appropriate styling
	// Use different node shapes based on node characteristics:
	// - Root nodes (no parent): rounded rectangle
	// - Leaf nodes (no children): stadium shape
	// - Other nodes: default rectangle
	if parentID == "" {
		// Root node - rounded rectangle
		fmt.Fprintf(sb, "    %s(\"%s\");\n", nodeID, escapeLabel(nodeLabel))
	} else if len(node.Children) == 0 {
		// Leaf node - stadium shape
		fmt.Fprintf(sb, "    %s[\"%s\"];\n", nodeID, escapeLabel(nodeLabel))
	} else {
		// Internal node - rectangle with rounded edges
		fmt.Fprintf(sb, "    %s{\"%s\"};\n", nodeID, escapeLabel(nodeLabel))
	}

	// Add connection to parent if exists
	if parentID != "" {
		fmt.Fprintf(sb, "    %s --> %s;\n", parentID, nodeID)
	}

	// Process children recursively
	for _, child := range node.Children {
		processNode(sb, child, nodeID, nodeIDs)
	}

	return nodeID
}

// getNodeID returns a unique ID for the node, using the node's ID if available
// or generating a unique one based on memory address
func getNodeID[Q, R, TC any](node *Node[Q, R, TC], nodeIDs map[*Node[Q, R, TC]]string) string {
	// Check if we already assigned an ID to this node
	if id, exists := nodeIDs[node]; exists {
		return id
	}

	// Use node's ID if available
	if node.ID != "" {
		// Replace spaces and special characters for Mermaid compatibility
		id := strings.ReplaceAll(node.ID, " ", "_")
		id = strings.ReplaceAll(id, "-", "_")
		nodeIDs[node] = id
		return id
	}

	// Generate a unique ID based on memory address
	id := fmt.Sprintf("node_%p", node)
	nodeIDs[node] = id
	return id
}

// getNodeLabel returns a formatted label for the node, including both description and ID
func getNodeLabel[Q, R, TC any](node *Node[Q, R, TC]) string {
	var label string

	if node.Description == "" && node.ID == "" {
		label = "Node"
	} else if node.Description == "" {
		label = node.ID
	} else if node.ID == "" {
		label = node.Description
	} else {
		// html escape description
		label = fmt.Sprintf("%s<br><i>%s</i>", node.ID, html.EscapeString(node.Description))
	}

	return label
}

// escapeLabel escapes special characters in Mermaid labels
func escapeLabel(label string) string {
	// Escape quotes and other special characters
	label = strings.ReplaceAll(label, "\"", "\\\"")
	// Don't escape HTML tags as Mermaid supports them for formatting
	return label
}
