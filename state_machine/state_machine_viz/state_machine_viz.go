package state_machine_viz

import (
	"fmt"
	"strings"

	"github.com/xhd2015/data-driven-testing/state_machine"
)

// Color constants for terminal output
const (
	ColorReset  = "\033[0m"
	ColorGreen  = "\033[32m"
	ColorRed    = "\033[31m"
	ColorGray   = "\033[90m"
	ColorYellow = "\033[33m"
)

// Symbol constants for visualization
const (
	SymbolPassed    = "🟢"
	SymbolFailed    = "🔴"
	SymbolUnvisited = "⚪"
	SymbolFailedTx  = "❌"
)

// VisualizationOptions contains configuration options for the visualization
type VisualizationOptions struct {
	UseUnicode bool // Whether to use Unicode symbols (true) or ASCII (false)
	UseColors  bool // Whether to use ANSI color codes
	Compact    bool // Whether to use a compact representation
}

// DefaultOptions returns the default visualization options
func DefaultOptions() VisualizationOptions {
	return VisualizationOptions{
		UseUnicode: true,
		UseColors:  true,
		Compact:    false,
	}
}

// VisualizeStateMachine generates a text-based visualization of a state machine
// with color-coded states based on their status.
//
// Parameters:
// - sm: The state machine to visualize
// - passedStates: Slice of state names that have been successfully traversed
// - failedStates: Slice of state names that failed during traversal
// - failedTransition: Optional name of the transition/event that failed
// - options: Optional visualization options
//
// Returns:
// - string: A formatted string representation of the state machine diagram
func VisualizeStateMachine(
	sm *state_machine.AnyMapStateMachine,
	passedStates []string,
	failedStates []string,
	failedTransition string,
	options VisualizationOptions,
) string {
	var sb strings.Builder

	// Add header
	sb.WriteString("STATE MACHINE VISUALIZATION\n")
	sb.WriteString("===========================\n\n")

	// Create a map for quick lookup of state status
	stateStatus := make(map[string]string)

	// Create maps to convert between state names and IDs
	nameToID := make(map[string]string)
	idToName := make(map[string]string)

	// Populate the name/ID maps
	for id, state := range sm.States {
		nameToID[state.Name] = id
		idToName[id] = state.Name
	}

	// Mark passed states
	for _, stateName := range passedStates {
		if id, ok := nameToID[stateName]; ok {
			stateStatus[id] = "passed"
		} else {
			// Try using the name as an ID directly
			stateStatus[stateName] = "passed"
		}
	}

	// Mark failed states
	for _, stateName := range failedStates {
		if id, ok := nameToID[stateName]; ok {
			stateStatus[id] = "failed"
		} else {
			// Try using the name as an ID directly
			stateStatus[stateName] = "failed"
		}
	}

	// Build a simple graph representation for easier traversal
	graph := buildGraph(sm)

	// Generate the visualization
	if options.UseUnicode {
		generateUnicodeVisualization(&sb, sm, graph, stateStatus, failedTransition, options)
	} else {
		generateASCIIVisualization(&sb, sm, graph, stateStatus, failedTransition, options)
	}

	// Add legend
	sb.WriteString("\nLegend:\n")
	if options.UseUnicode {
		if options.UseColors {
			sb.WriteString(fmt.Sprintf("%s%s%s Passed States\n", ColorGreen, SymbolPassed, ColorReset))
			sb.WriteString(fmt.Sprintf("%s%s%s Failed States\n", ColorRed, SymbolFailed, ColorReset))
			sb.WriteString(fmt.Sprintf("%s%s%s Unvisited States\n", ColorGray, SymbolUnvisited, ColorReset))
			sb.WriteString(fmt.Sprintf("%s%s%s Failed Transition\n", ColorRed, SymbolFailedTx, ColorReset))
		} else {
			sb.WriteString(fmt.Sprintf("%s Passed States\n", SymbolPassed))
			sb.WriteString(fmt.Sprintf("%s Failed States\n", SymbolFailed))
			sb.WriteString(fmt.Sprintf("%s Unvisited States\n", SymbolUnvisited))
			sb.WriteString(fmt.Sprintf("%s Failed Transition\n", SymbolFailedTx))
		}
	} else {
		if options.UseColors {
			sb.WriteString(fmt.Sprintf("[%sPASSED%s] - Successfully traversed states\n", ColorGreen, ColorReset))
			sb.WriteString(fmt.Sprintf("[%sFAILED%s] - States that encountered errors\n", ColorRed, ColorReset))
			sb.WriteString(fmt.Sprintf("[%s      %s] - Unvisited states\n", ColorGray, ColorReset))
			sb.WriteString(fmt.Sprintf("(%sFAILED%s) - Failed transition\n", ColorRed, ColorReset))
		} else {
			sb.WriteString("[PASSED] - Successfully traversed states\n")
			sb.WriteString("[FAILED] - States that encountered errors\n")
			sb.WriteString("[      ] - Unvisited states\n")
			sb.WriteString("(FAILED) - Failed transition\n")
		}
	}

	return sb.String()
}

// buildGraph creates a graph representation of the state machine for easier traversal
func buildGraph(sm *state_machine.AnyMapStateMachine) map[string]map[string]string {
	graph := make(map[string]map[string]string)

	// Initialize the adjacency list for each state
	for id := range sm.States {
		graph[id] = make(map[string]string)
	}

	// Add transitions to the graph
	for _, t := range sm.Transitions {
		graph[t.From][t.To] = t.Event
	}

	return graph
}

// getSortedStateIDs returns state IDs sorted in a logical order for visualization
func getSortedStateIDs(sm *state_machine.AnyMapStateMachine, graph map[string]map[string]string, failedTransition string) []string {
	// Find the initial state
	var initialStateID string
	for id, state := range sm.States {
		if state.IsInitial {
			initialStateID = id
			break
		}
	}

	// If no initial state found, just return all states
	if initialStateID == "" {
		stateIDs := make([]string, 0, len(sm.States))
		for id := range sm.States {
			stateIDs = append(stateIDs, id)
		}
		return stateIDs
	}

	// Find the states connected by the failed transition, if any
	var fromFailedID, toFailedID string
	if failedTransition != "" {
		for _, t := range sm.Transitions {
			if t.Event == failedTransition {
				fromFailedID = t.From
				toFailedID = t.To
				break
			}
		}
	}

	// Perform a topological sort starting from the initial state
	visited := make(map[string]bool)
	ordered := make([]string, 0, len(sm.States))

	var visit func(string)
	visit = func(id string) {
		if visited[id] {
			return
		}
		visited[id] = true
		ordered = append(ordered, id)

		// If this is the source of a failed transition, visit the target next
		if id == fromFailedID && toFailedID != "" {
			if !visited[toFailedID] {
				visit(toFailedID)
			}
			return
		}

		// Visit all neighbors
		for neighbor := range graph[id] {
			visit(neighbor)
		}
	}

	visit(initialStateID)

	// Add any remaining states that weren't reachable
	for id := range sm.States {
		if !visited[id] {
			ordered = append(ordered, id)
		}
	}

	return ordered
}

// generateUnicodeVisualization generates a Unicode-based visualization
func generateUnicodeVisualization(
	sb *strings.Builder,
	sm *state_machine.AnyMapStateMachine,
	graph map[string]map[string]string,
	stateStatus map[string]string,
	failedTransition string,
	options VisualizationOptions,
) {
	// Get all states and sort them by their relationships
	stateIDs := getSortedStateIDs(sm, graph, failedTransition)

	// Create a simple linear representation for now
	// This could be enhanced with a more sophisticated layout algorithm
	for i, stateID := range stateIDs {
		state := sm.States[stateID]

		// Format the state with appropriate color and symbol
		stateStr := formatState(stateID, state.Name, stateStatus[stateID], options)
		sb.WriteString(stateStr)

		// Add transitions if not the last state
		if i < len(stateIDs)-1 {
			nextStateID := stateIDs[i+1]

			// Find the event that connects these states
			event := ""
			for to, evt := range graph[stateID] {
				if to == nextStateID {
					event = evt
					break
				}
			}

			// If no direct connection, check if this is the source of the failed transition
			if event == "" && failedTransition != "" {
				for _, t := range sm.Transitions {
					if t.From == stateID && t.Event == failedTransition {
						event = failedTransition
						break
					}
				}
			}

			// If still no event found, use a placeholder
			if event == "" {
				event = "..."
			}

			// Format the transition
			transitionStr := formatTransition(event, event == failedTransition, options)
			sb.WriteString(transitionStr)
		}
	}
}

// generateASCIIVisualization generates an ASCII-based visualization
func generateASCIIVisualization(
	sb *strings.Builder,
	sm *state_machine.AnyMapStateMachine,
	graph map[string]map[string]string,
	stateStatus map[string]string,
	failedTransition string,
	options VisualizationOptions,
) {
	// Get all states and sort them by their relationships
	stateIDs := getSortedStateIDs(sm, graph, failedTransition)

	// Create a simple linear representation for now
	for i, stateID := range stateIDs {
		state := sm.States[stateID]

		// Format the state with appropriate color
		stateStr := formatStateASCII(stateID, state.Name, stateStatus[stateID], options)
		sb.WriteString(stateStr)

		// Add transitions if not the last state
		if i < len(stateIDs)-1 {
			nextStateID := stateIDs[i+1]

			// Find the event that connects these states
			event := ""
			for to, evt := range graph[stateID] {
				if to == nextStateID {
					event = evt
					break
				}
			}

			// If no direct connection, check if this is the source of the failed transition
			if event == "" && failedTransition != "" {
				for _, t := range sm.Transitions {
					if t.From == stateID && t.Event == failedTransition {
						event = failedTransition
						break
					}
				}
			}

			// If still no event found, use a placeholder
			if event == "" {
				event = "..."
			}

			// Format the transition
			transitionStr := formatTransitionASCII(event, event == failedTransition, options)
			sb.WriteString(transitionStr)
		}
	}
}

// formatState formats a state with appropriate color and symbol
func formatState(id, name, status string, options VisualizationOptions) string {
	symbol := SymbolUnvisited
	color := ColorGray

	switch status {
	case "passed":
		symbol = SymbolPassed
		color = ColorGreen
	case "failed":
		symbol = SymbolFailed
		color = ColorRed
	}

	if options.UseColors {
		return fmt.Sprintf("[%s%s%s %s] ", color, symbol, ColorReset, name)
	}
	return fmt.Sprintf("[%s %s] ", symbol, name)
}

// formatTransition formats a transition with appropriate styling
func formatTransition(event string, isFailed bool, options VisualizationOptions) string {
	if isFailed {
		if options.UseColors {
			return fmt.Sprintf("--%s(%s%s%s)--> ", event, ColorRed, SymbolFailedTx, ColorReset)
		}
		return fmt.Sprintf("--%s(%s)--> ", event, SymbolFailedTx)
	}
	return fmt.Sprintf("--%s--> ", event)
}

// formatStateASCII formats a state with appropriate color for ASCII output
func formatStateASCII(id, name, status string, options VisualizationOptions) string {
	var stateStr string
	color := ColorGray

	switch status {
	case "passed":
		stateStr = "PASSED"
		color = ColorGreen
	case "failed":
		stateStr = "FAILED"
		color = ColorRed
	default:
		stateStr = "      "
	}

	if options.UseColors {
		return fmt.Sprintf("[%s%s%s] %s ", color, stateStr, ColorReset, name)
	}
	return fmt.Sprintf("[%s] %s ", stateStr, name)
}

// formatTransitionASCII formats a transition with appropriate styling for ASCII output
func formatTransitionASCII(event string, isFailed bool, options VisualizationOptions) string {
	if isFailed {
		if options.UseColors {
			return fmt.Sprintf("--%s(%s%s%s)--> ", event, ColorRed, "FAILED", ColorReset)
		}
		return fmt.Sprintf("--%s(FAILED)--> ", event)
	}
	return fmt.Sprintf("--%s--> ", event)
}
