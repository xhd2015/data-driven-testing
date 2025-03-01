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
			result := VisualizeStateMachine(sm, tc.passedStates, tc.failedStates, tc.failedTransition, tc.options)

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
