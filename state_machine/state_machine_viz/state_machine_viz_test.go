package state_machine_viz

import (
	"strings"
	"testing"

	"github.com/xhd2015/data-driven-testing/state_machine"
)

// TestVisualizeStateMachine tests the state machine visualization
func TestVisualizeStateMachine(t *testing.T) {
	// Create a simple state machine for testing
	states := map[string]*state_machine.State{
		"initial": {
			ID:          "initial",
			Name:        "Initial",
			Description: "Initial state",
			IsInitial:   true,
			Data:        make(map[string]interface{}),
		},
		"processing": {
			ID:          "processing",
			Name:        "Processing",
			Description: "Processing state",
			Data:        make(map[string]interface{}),
		},
		"completed": {
			ID:          "completed",
			Name:        "Completed",
			Description: "Completed state",
			IsFinal:     true,
			Data:        make(map[string]interface{}),
		},
		"failed": {
			ID:          "failed",
			Name:        "Failed",
			Description: "Failed state",
			IsFinal:     true,
			Data:        make(map[string]interface{}),
		},
	}

	transitions := []state_machine.Transition[map[string]interface{}]{
		{
			From:  "initial",
			To:    "processing",
			Event: "start_processing",
		},
		{
			From:  "processing",
			To:    "completed",
			Event: "complete",
		},
		{
			From:  "processing",
			To:    "failed",
			Event: "fail",
		},
	}

	sm := state_machine.NewStateMachine("TestWorkflow", states, transitions)

	// Test cases
	testCases := []struct {
		name             string
		passedStates     []string
		failedStates     []string
		failedTransition string
		options          VisualizationOptions
		expectedContains []string
		notExpected      []string
	}{
		{
			name:             "Success Path",
			passedStates:     []string{"Initial", "Processing", "Completed"},
			failedStates:     []string{},
			failedTransition: "",
			options:          DefaultOptions(),
			expectedContains: []string{
				"Initial",
				"Processing",
				"Completed",
				"start_processing",
			},
			notExpected: []string{
				"fail(", // No failed transition
			},
		},
		{
			name:             "Failure Path",
			passedStates:     []string{"Initial", "Processing"},
			failedStates:     []string{"Failed"},
			failedTransition: "fail",
			options:          DefaultOptions(),
			expectedContains: []string{
				"Initial",
				"Processing",
				"Failed",
				"fail", // The event name
				"❌",    // The failed symbol (might be in different places)
			},
			notExpected: []string{},
		},
		{
			name:             "ASCII Mode",
			passedStates:     []string{"Initial", "Processing"},
			failedStates:     []string{"Failed"},
			failedTransition: "fail",
			options: VisualizationOptions{
				UseUnicode: false,
				UseColors:  false,
				Compact:    false,
			},
			expectedContains: []string{
				"[PASSED] Initial",
				"[PASSED] Processing",
				"[FAILED] Failed",
				"fail",     // The event name
				"(FAILED)", // The failed marker (might be in different places)
			},
			notExpected: []string{
				"🟢", "🔴", "⚪", "❌", // Unicode symbols should not be present
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Render(sm, tc.passedStates, tc.failedStates, tc.failedTransition, tc.options)

			// Check that all expected states are present
			for _, expected := range tc.expectedContains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected visualization to contain '%s', but it didn't.\nVisualization:\n%s", expected, result)
				}
			}

			// Check that unexpected strings are not present
			for _, notExpected := range tc.notExpected {
				if strings.Contains(result, notExpected) {
					t.Errorf("Expected visualization to NOT contain '%s', but it did.\nVisualization:\n%s", notExpected, result)
				}
			}

			// Print the visualization for debugging
			t.Logf("Visualization:\n%s", result)
		})
	}
}

// TestRenderWithFailedTransition tests the automatic state determination
// based on a failed transition
func TestRenderWithFailedTransition(t *testing.T) {
	// Create a simple linear state machine for testing
	linearStates := map[string]*state_machine.State{
		"A": {
			ID:        "A",
			Name:      "State A",
			IsInitial: true,
			Data:      make(map[string]interface{}),
		},
		"B": {
			ID:   "B",
			Name: "State B",
			Data: make(map[string]interface{}),
		},
		"C": {
			ID:   "C",
			Name: "State C",
			Data: make(map[string]interface{}),
		},
		"D": {
			ID:      "D",
			Name:    "State D",
			IsFinal: true,
			Data:    make(map[string]interface{}),
		},
	}

	linearTransitions := []state_machine.Transition[map[string]interface{}]{
		{
			From:  "A",
			To:    "B",
			Event: "A_to_B",
		},
		{
			From:  "B",
			To:    "C",
			Event: "B_to_C",
		},
		{
			From:  "C",
			To:    "D",
			Event: "C_to_D",
		},
	}

	linearSM := state_machine.NewStateMachine("LinearWorkflow", linearStates, linearTransitions)

	// Create a more complex state machine with branches
	branchingStates := map[string]*state_machine.State{
		"start": {
			ID:        "start",
			Name:      "Start",
			IsInitial: true,
			Data:      make(map[string]interface{}),
		},
		"branch1": {
			ID:   "branch1",
			Name: "Branch 1",
			Data: make(map[string]interface{}),
		},
		"branch2": {
			ID:   "branch2",
			Name: "Branch 2",
			Data: make(map[string]interface{}),
		},
		"merge": {
			ID:   "merge",
			Name: "Merge Point",
			Data: make(map[string]interface{}),
		},
		"end": {
			ID:      "end",
			Name:    "End",
			IsFinal: true,
			Data:    make(map[string]interface{}),
		},
	}

	branchingTransitions := []state_machine.Transition[map[string]interface{}]{
		{
			From:  "start",
			To:    "branch1",
			Event: "take_branch1",
		},
		{
			From:  "start",
			To:    "branch2",
			Event: "take_branch2",
		},
		{
			From:  "branch1",
			To:    "merge",
			Event: "branch1_complete",
		},
		{
			From:  "branch2",
			To:    "merge",
			Event: "branch2_complete",
		},
		{
			From:  "merge",
			To:    "end",
			Event: "finish",
		},
	}

	branchingSM := state_machine.NewStateMachine("BranchingWorkflow", branchingStates, branchingTransitions)

	// Test cases
	testCases := []struct {
		name             string
		sm               *state_machine.StateMachine[map[string]interface{}]
		failedTransition string
		options          VisualizationOptions
		expectedPassed   []string
		expectedFailed   []string
	}{
		{
			name:             "Linear Path Basic Failure",
			sm:               linearSM,
			failedTransition: "B_to_C",
			options:          DefaultOptions(),
			expectedPassed:   []string{"State A", "State B"},
			expectedFailed:   []string{"State C"},
		},
		{
			name:             "Linear Path First Transition Failure",
			sm:               linearSM,
			failedTransition: "A_to_B",
			options:          DefaultOptions(),
			expectedPassed:   []string{"State A"},
			expectedFailed:   []string{"State B"},
		},
		{
			name:             "Linear Path Last Transition Failure",
			sm:               linearSM,
			failedTransition: "C_to_D",
			options:          DefaultOptions(),
			expectedPassed:   []string{"State A", "State B", "State C"},
			expectedFailed:   []string{"State D"},
		},
		{
			name:             "Branching Path Failure on Branch 1",
			sm:               branchingSM,
			failedTransition: "branch1_complete",
			options:          DefaultOptions(),
			expectedPassed:   []string{"Start", "Branch 1"},
			expectedFailed:   []string{"Merge Point"},
		},
		{
			name:             "Branching Path Failure on Final Transition",
			sm:               branchingSM,
			failedTransition: "finish",
			options:          DefaultOptions(),
			// The actual path could be either branch, but we should see the merge point
			expectedPassed: []string{"Merge Point"},
			expectedFailed: []string{"End"},
		},
		{
			name:             "Non-existent Transition",
			sm:               linearSM,
			failedTransition: "nonexistent",
			options:          DefaultOptions(),
			// Should mark all states as passed when transition doesn't exist
			expectedPassed: []string{"State A", "State B", "State C", "State D"},
			expectedFailed: []string{},
		},
		{
			name:             "Empty Transition",
			sm:               linearSM,
			failedTransition: "",
			options:          DefaultOptions(),
			// Should mark all states as passed when no transition is specified
			expectedPassed: []string{"State A", "State B", "State C", "State D"},
			expectedFailed: []string{},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := RenderFailedOptions(tc.sm, tc.failedTransition, tc.options)

			// Print the visualization for debugging
			t.Logf("Visualization:\n%s", result)

			// Check each expected passed state
			for _, stateName := range tc.expectedPassed {
				// Use a more flexible approach that just looks for the state name and the passed symbol nearby
				if tc.options.UseUnicode {
					// For unicode mode, check if both the state name and the passed symbol are in the output
					if !strings.Contains(result, stateName) || !strings.Contains(result, SymbolPassed) {
						t.Errorf("Expected state '%s' to be marked as passed with symbol '%s', but couldn't find both in output",
							stateName, SymbolPassed)
					}
				} else {
					// For ASCII mode
					if !strings.Contains(result, stateName) || !strings.Contains(result, "PASSED") {
						t.Errorf("Expected state '%s' to be marked as PASSED, but couldn't find both in output", stateName)
					}
				}
			}

			// Check each expected failed state
			for _, stateName := range tc.expectedFailed {
				if tc.options.UseUnicode {
					// For unicode mode, check if both the state name and the failed symbol are in the output
					if !strings.Contains(result, stateName) || !strings.Contains(result, SymbolFailed) {
						t.Errorf("Expected state '%s' to be marked as failed with symbol '%s', but couldn't find both in output",
							stateName, SymbolFailed)
					}
				} else {
					// For ASCII mode
					if !strings.Contains(result, stateName) || !strings.Contains(result, "FAILED") {
						t.Errorf("Expected state '%s' to be marked as FAILED, but couldn't find both in output", stateName)
					}
				}
			}

			// Check that the failed transition is marked correctly if it exists
			if tc.failedTransition != "" && tc.failedTransition != "nonexistent" {
				if !strings.Contains(result, tc.failedTransition) {
					t.Errorf("Expected failed transition '%s' to be included in the visualization, but it wasn't",
						tc.failedTransition)
				}
			}
		})
	}
}

// TestRenderFailed tests the convenience function for rendering a failed transition
func TestRenderFailed(t *testing.T) {
	// Create a simple state machine for testing
	states := map[string]*state_machine.State{
		"initial": {
			ID:          "initial",
			Name:        "Initial",
			Description: "Initial state",
			IsInitial:   true,
			Data:        make(map[string]interface{}),
		},
		"processing": {
			ID:          "processing",
			Name:        "Processing",
			Description: "Processing state",
			Data:        make(map[string]interface{}),
		},
		"completed": {
			ID:          "completed",
			Name:        "Completed",
			Description: "Completed state",
			IsFinal:     true,
			Data:        make(map[string]interface{}),
		},
		"failed": {
			ID:          "failed",
			Name:        "Failed",
			Description: "Failed state",
			IsFinal:     true,
			Data:        make(map[string]interface{}),
		},
	}

	transitions := []state_machine.Transition[map[string]interface{}]{
		{
			From:  "initial",
			To:    "processing",
			Event: "start_processing",
		},
		{
			From:  "processing",
			To:    "completed",
			Event: "complete",
		},
		{
			From:  "processing",
			To:    "failed",
			Event: "fail",
		},
	}

	sm := state_machine.NewStateMachine("TestWorkflow", states, transitions)

	// Test the convenience function with a failed transition
	result := RenderFailed(sm, "fail")

	// Print the visualization for debugging
	t.Logf("Visualization:\n%s", result)

	// Check that the initial state is marked as passed
	if !strings.Contains(result, "Initial") || !strings.Contains(result, SymbolPassed) {
		t.Errorf("Expected 'Initial' state to be marked as passed, but it wasn't")
	}

	// Check that the processing state is marked as passed
	if !strings.Contains(result, "Processing") || !strings.Contains(result, SymbolPassed) {
		t.Errorf("Expected 'Processing' state to be marked as passed, but it wasn't")
	}

	// Check that the failed state is marked as failed
	if !strings.Contains(result, "Failed") || !strings.Contains(result, SymbolFailed) {
		t.Errorf("Expected 'Failed' state to be marked as failed, but it wasn't")
	}

	// Check that the failed transition is marked
	if !strings.Contains(result, "fail") {
		t.Errorf("Expected failed transition 'fail' to be included in the visualization, but it wasn't")
	}
}
