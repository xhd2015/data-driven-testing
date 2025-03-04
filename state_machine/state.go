package state_machine

import (
	"errors"
	"fmt"
	"strings"
)

// StateRef is an interface that represents a reference to a state in a state machine
type StateRef[T any] interface {
	// ID returns the unique identifier of the state
	ID() string

	// Name returns the human-readable name of the state
	Name() string

	// Description returns the human-readable description of the state
	Description() string

	// IsInitial returns whether this is an initial state
	IsInitial() bool

	// IsFinal returns whether this is a final state
	IsFinal() bool

	// GetData returns the data associated with the state
	GetData() map[string]interface{}

	// SetData sets data for the state
	SetData(key string, value interface{})
}

// stateRefImpl is the internal implementation of StateRef
type stateRefImpl[T any] struct {
	stateMachine *StateMachine[T]
	state        *State
}

// ID returns the unique identifier of the state
func (sr *stateRefImpl[T]) ID() string {
	return sr.state.ID
}

// Name returns the human-readable name of the state
func (sr *stateRefImpl[T]) Name() string {
	return sr.state.Name
}

// Description returns the human-readable description of the state
func (sr *stateRefImpl[T]) Description() string {
	return sr.state.Description
}

// IsInitial returns whether this is an initial state
func (sr *stateRefImpl[T]) IsInitial() bool {
	return sr.state.IsInitial
}

// IsFinal returns whether this is a final state
func (sr *stateRefImpl[T]) IsFinal() bool {
	return sr.state.IsFinal
}

// GetData returns the data associated with the state
func (sr *stateRefImpl[T]) GetData() map[string]interface{} {
	return sr.state.Data
}

// SetData sets data for the state
func (sr *stateRefImpl[T]) SetData(key string, value interface{}) {
	sr.state.Data[key] = value
}

// State represents a single state in a state machine
type State struct {
	ID          string
	Name        string // Human-readable name
	Description string
	IsInitial   bool
	IsFinal     bool
	Data        map[string]interface{} // Additional metadata
}

// Transition represents a transition between states
type Transition[T any] struct {
	From      string
	To        string
	Event     string
	Condition func(*T) bool  // Optional condition function
	Action    func(*T) error // Optional action to execute during transition
}

// StateMachine represents a complete state machine
type StateMachine[T any] struct {
	Name         string
	States       map[string]*State
	Transitions  []Transition[T]
	CurrentState *State
	Context      T // Shared context for the state machine
}

// NewStateMachine creates a new state machine with the given name, states, and transitions
func NewStateMachine[T any](name string, states map[string]*State, transitions []Transition[T]) *StateMachine[T] {
	sm := &StateMachine[T]{
		Name:        name,
		States:      states,
		Transitions: transitions,
	}

	// Set initial state if one exists
	for _, state := range states {
		if state.IsInitial {
			sm.CurrentState = state
			break
		}
	}

	return sm
}

// GetCurrentState returns the current state of the state machine
func (sm *StateMachine[T]) GetCurrentState() *State {
	return sm.CurrentState
}

// Reset resets the state machine to its initial state
func (sm *StateMachine[T]) Reset() error {
	// Find initial state
	var initialState *State
	for _, state := range sm.States {
		if state.IsInitial {
			initialState = state
			break
		}
	}

	if initialState == nil {
		return errors.New("no initial state defined")
	}

	sm.CurrentState = initialState

	// Clear the context if it's a map, but preserve the map itself
	if contextMap, ok := any(sm.Context).(map[string]interface{}); ok {
		for k := range contextMap {
			delete(contextMap, k)
		}
	}

	return nil
}

// Trigger attempts to trigger a transition with the given event
func (sm *StateMachine[T]) Trigger(event string) error {
	if sm.CurrentState == nil {
		return errors.New("state machine has no current state")
	}

	// Find valid transitions for the current state and event
	var validTransitions []Transition[T]
	for _, t := range sm.Transitions {
		if t.From == sm.CurrentState.ID && t.Event == event {
			validTransitions = append(validTransitions, t)
		}
	}

	if len(validTransitions) == 0 {
		return fmt.Errorf("no valid transition for event %s from state %s", event, sm.CurrentState.ID)
	}

	// Find the first transition with a satisfied condition
	for _, t := range validTransitions {
		if t.Condition == nil || t.Condition(&sm.Context) {
			// Execute the action if defined
			if t.Action != nil {
				if err := t.Action(&sm.Context); err != nil {
					return fmt.Errorf("action execution failed: %w", err)
				}
			}

			// Transition to the new state
			sm.CurrentState = sm.States[t.To]
			return nil
		}
	}

	return fmt.Errorf("no transition conditions satisfied for event %s from state %s", event, sm.CurrentState.ID)
}

// Validate checks if the state machine is valid
func (sm *StateMachine[T]) Validate() error {
	// Check if there's at least one state
	if len(sm.States) == 0 {
		return errors.New("state machine has no states")
	}

	// Check if there's exactly one initial state
	initialStates := 0
	for _, state := range sm.States {
		if state.IsInitial {
			initialStates++
		}
	}
	if initialStates == 0 {
		return errors.New("state machine has no initial state")
	}
	if initialStates > 1 {
		return errors.New("state machine has multiple initial states")
	}

	// Check if all transitions reference valid states
	for _, t := range sm.Transitions {
		if _, exists := sm.States[t.From]; !exists {
			return fmt.Errorf("transition references non-existent from state: %s", t.From)
		}
		if _, exists := sm.States[t.To]; !exists {
			return fmt.Errorf("transition references non-existent to state: %s", t.To)
		}
	}

	return nil
}

// ToDOT generates a DOT representation of the state machine for visualization with Graphviz
func (sm *StateMachine[T]) ToDOT() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("digraph %s {\n", sm.Name))
	sb.WriteString("  rankdir=LR;\n")
	sb.WriteString("  node [shape=circle];\n")

	// Add states using sorted order for consistent rendering
	sortedStates := sm.SortedStates()
	for _, state := range sortedStates {
		shape := "circle"
		if state.IsInitial {
			shape = "doublecircle"
		}
		if state.IsFinal {
			shape = "doublecircle"
		}
		sb.WriteString(fmt.Sprintf("  %s [shape=%s, label=\"%s\"];\n", state.ID, shape, state.Name))
	}

	// Add transitions
	for _, t := range sm.Transitions {
		sb.WriteString(fmt.Sprintf("  %s -> %s [label=\"%s\"];\n", t.From, t.To, t.Event))
	}

	sb.WriteString("}\n")
	return sb.String()
}

// ToMermaid generates a Mermaid.js representation of the state machine
func (sm *StateMachine[T]) ToMermaid() string {
	var sb strings.Builder

	sb.WriteString("stateDiagram-v2\n")

	// Add states using the sorted order
	sortedStates := sm.SortedStates()
	for _, state := range sortedStates {
		// Add a comment with the state name for clarity
		sb.WriteString(fmt.Sprintf("  %s: %s\n", state.ID, state.Name))

		if state.IsInitial {
			sb.WriteString(fmt.Sprintf("  [*] --> %s\n", state.ID))
		}
		if state.IsFinal {
			sb.WriteString(fmt.Sprintf("  %s --> [*]\n", state.ID))
		}
	}

	// Add transitions
	for _, t := range sm.Transitions {
		sb.WriteString(fmt.Sprintf("  %s --> %s : %s\n", t.From, t.To, t.Event))
	}

	return sb.String()
}

// SortedStates returns states in a topological order starting from initial states.
// This ensures consistent visualization output regardless of Go's map iteration order.
func (sm *StateMachine[T]) SortedStates() []*State {
	// Prepare result slice
	sorted := make([]*State, 0, len(sm.States))
	visited := make(map[string]bool)

	// Build adjacency list from transitions
	adjacencyList := make(map[string][]string)
	for _, t := range sm.Transitions {
		adjacencyList[t.From] = append(adjacencyList[t.From], t.To)
	}

	// First collect initial states
	initialStates := make([]string, 0)
	for id, state := range sm.States {
		if state.IsInitial {
			initialStates = append(initialStates, id)
			sorted = append(sorted, state)
			visited[id] = true
		}
	}

	// Breadth-first traversal from initial states
	queue := initialStates
	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		// Add all unvisited neighbors to queue
		for _, neighborID := range adjacencyList[currentID] {
			if !visited[neighborID] {
				if state, exists := sm.States[neighborID]; exists {
					sorted = append(sorted, state)
					visited[neighborID] = true
					queue = append(queue, neighborID)
				}
			}
		}
	}

	// Add any remaining states that weren't reachable from initial states
	for id, state := range sm.States {
		if !visited[id] {
			sorted = append(sorted, state)
		}
	}

	return sorted
}

// ToPlantUML generates a PlantUML representation of the state machine
func (sm *StateMachine[T]) ToPlantUML() string {
	var sb strings.Builder

	sb.WriteString("@startuml\n")

	// Add title based on state machine name
	sb.WriteString(fmt.Sprintf("title %s\n\n", sm.Name))

	// Use sorted states for consistent rendering
	sortedStates := sm.SortedStates()

	// Add initial states
	for _, state := range sortedStates {
		if state.IsInitial {
			sb.WriteString(fmt.Sprintf("[*] --> %s\n", state.ID))
		}
	}

	// Add all states with their descriptions
	for _, state := range sortedStates {
		// Only add description if the state has a name
		if state.Name != "" {
			sb.WriteString(fmt.Sprintf("state \"%s\" as %s", state.Name, state.ID))
			if state.Description != "" {
				sb.WriteString(fmt.Sprintf(" : %s", state.Description))
			}
			sb.WriteString("\n")
		}

		// Mark final states
		if state.IsFinal {
			sb.WriteString(fmt.Sprintf("%s --> [*]\n", state.ID))
		}
	}

	// Add all transitions
	for _, t := range sm.Transitions {
		sb.WriteString(fmt.Sprintf("%s --> %s : %s\n", t.From, t.To, t.Event))
	}

	sb.WriteString("@enduml\n")
	return sb.String()
}

// For backward compatibility
type AnyMapStateMachine = StateMachine[map[string]interface{}]

// CreateAnyMapStateMachine creates a new state machine with map[string]interface{} context
// for backward compatibility with code that expects the old untyped approach
func CreateAnyMapStateMachine(name string) *AnyMapStateMachine {
	sm := &AnyMapStateMachine{
		Name:        name,
		States:      make(map[string]*State),
		Transitions: []Transition[map[string]interface{}]{},
		Context:     make(map[string]interface{}),
	}
	return sm
}
