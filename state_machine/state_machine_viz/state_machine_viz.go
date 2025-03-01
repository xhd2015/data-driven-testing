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

// RenderAllPassed generates a visualization where all states are marked as passed,
// using default visualization options. This is the simplest API for state machine visualization.
//
// Parameters:
// - sm: The state machine to visualize. Can be any type of StateMachine[T].
//
// Returns:
// - string: A formatted string representation of the state machine diagram with all states marked as passed
func RenderAllPassed[T any](sm *state_machine.StateMachine[T]) string {
	return RenderAllPassedOptions(sm, DefaultOptions())
}

// RenderAllPassedOptions generates a visualization where all states are marked as passed.
// This is a simplified API for VisualizeStateMachine that marks all states as passed.
//
// Parameters:
// - sm: The state machine to visualize. Can be any type of StateMachine[T].
// - options: Visualization options
//
// Returns:
// - string: A formatted string representation of the state machine diagram with all states marked as passed
func RenderAllPassedOptions[T any](sm *state_machine.StateMachine[T], options VisualizationOptions) string {
	// Extract all state names from the state machine
	allStateNames := make([]string, 0, len(sm.States))
	for _, state := range sm.States {
		allStateNames = append(allStateNames, state.Name)
	}

	// Call the main visualization function with all states marked as passed
	return Render(
		sm,
		allStateNames, // All states are passed
		[]string{},    // No failed states
		"",            // No failed transition
		options,
	)
}

// Render generates a text-based visualization of a state machine
// with color-coded states based on their status.
//
// Parameters:
// - sm: The state machine to visualize. Can be any type of StateMachine[T].
// - passedStates: Slice of state names that have been successfully traversed
// - failedStates: Slice of state names that failed during traversal
// - failedTransition: Optional name of the transition/event that failed
// - options: Optional visualization options
//
// Returns:
// - string: A formatted string representation of the state machine diagram
func Render[T any](sm *state_machine.StateMachine[T], passedStates []string, failedStates []string, failedTransition string, options VisualizationOptions) string {
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
func buildGraph[T any](sm *state_machine.StateMachine[T]) map[string]map[string]string {
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
func getSortedStateIDs[T any](sm *state_machine.StateMachine[T], graph map[string]map[string]string, failedTransition string) []string {
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
func generateUnicodeVisualization[T any](
	sb *strings.Builder,
	sm *state_machine.StateMachine[T],
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
func generateASCIIVisualization[T any](
	sb *strings.Builder,
	sm *state_machine.StateMachine[T],
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

// RenderFailedOptions generates a visualization by automatically determining
// passed states and failed states based on a failed transition.
//
// Parameters:
// - sm: The state machine to visualize
// - failedTransition: The name of the transition that failed
// - options: Optional visualization options
//
// Returns:
// - string: A formatted string representation of the state machine diagram
func RenderFailedOptions[T any](
	sm *state_machine.StateMachine[T],
	failedTransition string,
	options VisualizationOptions,
) string {
	// Use default options if not specified
	if (options == VisualizationOptions{}) {
		options = DefaultOptions()
	}

	// If no failed transition specified, just render all states as passed
	if failedTransition == "" {
		return RenderAllPassedOptions(sm, options)
	}

	// Find the failed transition in the state machine
	var fromStateID, toStateID string
	found := false

	for _, t := range sm.Transitions {
		if t.Event == failedTransition {
			fromStateID = t.From
			toStateID = t.To
			found = true
			break
		}
	}

	// If the transition wasn't found, just render all states as passed
	if !found {
		return RenderAllPassedOptions(sm, options)
	}

	// Find the initial state
	var initialStateID string
	for id, state := range sm.States {
		if state.IsInitial {
			initialStateID = id
			break
		}
	}

	// If no initial state found, just render with the failed transition
	if initialStateID == "" {
		return Render(
			sm,
			[]string{},          // No passed states
			[]string{toStateID}, // Mark the target state as failed
			failedTransition,
			options,
		)
	}

	// Trace the execution path from the initial state to the source of the failed transition
	passedStateIDs := traceExecutionPath(sm, initialStateID, fromStateID)

	// Convert state IDs to state names for the Render function
	passedStateNames := make([]string, 0, len(passedStateIDs))
	for _, id := range passedStateIDs {
		if state, exists := sm.States[id]; exists {
			passedStateNames = append(passedStateNames, state.Name)
		}
	}

	// Get the target state name
	var failedStateName string
	if targetState, exists := sm.States[toStateID]; exists {
		failedStateName = targetState.Name
	}

	// Call the existing Render function with the computed states
	return Render(
		sm,
		passedStateNames,
		[]string{failedStateName},
		failedTransition,
		options,
	)
}

// RenderFailed is a convenience function that visualizes a state machine with a failed transition
// using default visualization options. This is the simplest API for visualizing a failed transition.
//
// Parameters:
// - sm: The state machine to visualize
// - failedTransition: The name of the transition that failed
//
// Returns:
// - string: A formatted string representation of the state machine diagram
func RenderFailed[T any](sm *state_machine.StateMachine[T], failedTransition string) string {
	return RenderFailedOptions(sm, failedTransition, DefaultOptions())
}

// traceExecutionPath performs a breadth-first search from the initial state to the target state,
// returning a slice of state IDs that would be traversed in the path.
func traceExecutionPath[T any](
	sm *state_machine.StateMachine[T],
	initialStateID string,
	targetStateID string,
) []string {
	// If the initial state is the target, just return it
	if initialStateID == targetStateID {
		return []string{initialStateID}
	}

	// Build a graph representation for traversal
	graph := make(map[string]map[string]bool)
	for _, t := range sm.Transitions {
		if _, exists := graph[t.From]; !exists {
			graph[t.From] = make(map[string]bool)
		}
		graph[t.From][t.To] = true
	}

	// Queue for BFS traversal
	queue := []string{initialStateID}

	// Track visited states and paths
	visited := make(map[string]bool)
	visited[initialStateID] = true

	// Track parent states to reconstruct the path
	parent := make(map[string]string)

	// Perform BFS
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// If we found the target state, reconstruct the path
		if current == targetStateID {
			path := []string{current}
			for state := parent[current]; state != ""; state = parent[state] {
				path = append([]string{state}, path...)
			}
			return path
		}

		// Explore neighbors
		for neighbor := range graph[current] {
			if !visited[neighbor] {
				visited[neighbor] = true
				parent[neighbor] = current
				queue = append(queue, neighbor)
			}
		}
	}

	// If no path was found, just return the initial state
	return []string{initialStateID}
}
