package svg

import (
	"fmt"
	"strings"

	"github.com/xhd2015/data-driven-testing/decision_tree"
	"github.com/xhd2015/data-driven-testing/decision_tree/layout"
)

// Renderer handles SVG generation
type Renderer struct {
	config *decision_tree.Config
	layout *layout.Engine
}

// NewRenderer creates a new SVG renderer
func NewRenderer(config *decision_tree.Config) *Renderer {
	return &Renderer{
		config: config,
		layout: layout.NewEngine(config),
	}
}

// SetCenterParent sets whether parent nodes should be centered over their children
func (r *Renderer) SetCenterParent(center bool) {
	r.layout.SetCenterParent(center)
}

// RenderTree generates SVG for the entire tree
func (r *Renderer) RenderTree(root *decision_tree.Node) string {
	// Calculate layout
	layoutRoot := r.layout.CalculateLayout(root)

	// Find bounds
	minX, minY, maxX, maxY := r.findBounds(layoutRoot)

	// Add asymmetric padding for left alignment
	leftPadding := 20.0   // smaller padding on left
	rightPadding := 100.0 // larger padding on right
	topPadding := 50.0
	bottomPadding := 50.0

	minX -= leftPadding
	minY -= topPadding
	maxX += rightPadding
	maxY += bottomPadding

	width := maxX - minX
	height := maxY - minY

	// Generate SVG with viewBox for proper scaling
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`<svg width="%f" height="%f" viewBox="%f %f %f %f" xmlns="http://www.w3.org/2000/svg">`,
		width, height, minX, minY, width, height))

	// Add definitions for markers (arrow heads)
	sb.WriteString(`
	<defs>
		<marker id="arrowhead" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
			<polygon points="0 0, 10 3.5, 0 7" fill="#000"/>
		</marker>
	</defs>`)

	// Add background for better visibility (optional)
	sb.WriteString(fmt.Sprintf(`<rect x="%f" y="%f" width="%f" height="%f" fill="white"/>`,
		minX, minY, width, height))

	// Render all edges first (so they appear behind nodes)
	r.renderEdges(&sb, layoutRoot)

	// Render all nodes
	r.renderNodes(&sb, layoutRoot)

	sb.WriteString("</svg>")
	return sb.String()
}

// findBounds calculates the bounding box of the tree
func (r *Renderer) findBounds(node *layout.LayoutNode) (minX, minY, maxX, maxY float64) {
	if node == nil {
		return 0, 0, 0, 0
	}

	minX = node.X - node.Width/2
	maxX = node.X + node.Width/2
	minY = node.Y
	maxY = node.Y + node.Height

	// Include space for conditions text
	if len(node.Node.Conditions) > 0 {
		if len(node.Children) == 0 {
			// For terminal nodes, add space below
			maxY += 30 // Space for conditions text
		} else {
			// For non-terminal nodes, add space above
			minY -= 20
		}
	}

	for _, child := range node.Children {
		childMinX, childMinY, childMaxX, childMaxY := r.findBounds(child)
		if childMinX < minX {
			minX = childMinX
		}
		if childMinY < minY {
			minY = childMinY
		}
		if childMaxX > maxX {
			maxX = childMaxX
		}
		if childMaxY > maxY {
			maxY = childMaxY
		}
	}

	return minX, minY, maxX, maxY
}

// wrapText splits text into lines of maximum length
func (r *Renderer) wrapText(text string, maxChars int) []string {
	if len(text) <= maxChars {
		return []string{text}
	}

	var lines []string
	words := strings.Fields(text)
	currentLine := words[0]

	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) <= maxChars {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	lines = append(lines, currentLine)
	return lines
}

// renderNodes renders all nodes in the tree
func (r *Renderer) renderNodes(sb *strings.Builder, node *layout.LayoutNode) {
	if node == nil {
		return
	}

	// Get node style
	style := node.Style
	if style == nil {
		style = r.config.DefaultStyle
	}

	// Render node
	x := node.X - node.Width/2
	y := node.Y

	// Node shape - use rounded corners only for terminal nodes
	if len(node.Children) == 0 {
		// Terminal nodes get rounded corners
		sb.WriteString(fmt.Sprintf(`<rect x="%f" y="%f" width="%f" height="%f" 
			fill="%s" stroke="%s" stroke-width="%d" rx="3" ry="3"/>`,
			x, y, node.Width, node.Height,
			style.Fill, style.Stroke, style.StrokeWidth))
	} else {
		// Non-terminal nodes get sharp corners
		sb.WriteString(fmt.Sprintf(`<rect x="%f" y="%f" width="%f" height="%f" 
			fill="%s" stroke="%s" stroke-width="%d"/>`,
			x, y, node.Width, node.Height,
			style.Fill, style.Stroke, style.StrokeWidth))
	}

	// Wrap and render node text
	lines := r.wrapText(node.Node.Label, 20)
	lineHeight := 15.0 // Reduced line height
	startY := node.Y + (node.Height-float64(len(lines))*lineHeight)/2 + lineHeight/2
	for i, line := range lines {
		sb.WriteString(fmt.Sprintf(`<text x="%f" y="%f" text-anchor="middle" dominant-baseline="middle" font-family="Arial" font-size="12">%s</text>`,
			node.X, startY+float64(i)*lineHeight, line))
	}

	// Render conditions if any
	if len(node.Node.Conditions) > 0 {
		conditions := make([]string, 0, len(node.Node.Conditions))
		for k, v := range node.Node.Conditions {
			conditions = append(conditions, fmt.Sprintf("%s=%v", k, v))
		}
		condText := strings.Join(conditions, ", ")
		// Wrap conditions text
		condLines := r.wrapText(condText, 20)

		if len(node.Children) == 0 {
			// For terminal nodes, render conditions below the node
			for i, line := range condLines {
				sb.WriteString(fmt.Sprintf(`<text x="%f" y="%f" text-anchor="middle" font-size="10" fill="#666" font-family="Arial">%s</text>`,
					node.X, node.Y+node.Height+12+float64(i)*12, line))
			}
		} else {
			// For non-terminal nodes, render conditions above the node
			for i, line := range condLines {
				sb.WriteString(fmt.Sprintf(`<text x="%f" y="%f" text-anchor="middle" font-size="10" fill="#666" font-family="Arial">%s</text>`,
					node.X, node.Y-3-float64(len(condLines)-1-i)*12, line))
			}
		}
	}

	// Render children
	for _, child := range node.Children {
		r.renderNodes(sb, child)
	}
}

// renderEdges renders all edges in the tree
func (r *Renderer) renderEdges(sb *strings.Builder, node *layout.LayoutNode) {
	if node == nil {
		return
	}

	for _, child := range node.Children {
		// Calculate edge points
		startX := node.X
		startY := node.Y + node.Height
		endX := child.X
		endY := child.Y

		// Draw straight edge with arrow
		sb.WriteString(fmt.Sprintf(`<path d="M %f %f L %f %f" stroke="black" stroke-width="1" 
			fill="none" marker-end="url(#arrowhead)"/>`,
			startX, startY, endX, endY))

		r.renderEdges(sb, child)
	}
}
