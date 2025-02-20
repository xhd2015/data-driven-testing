package main

import (
	"testing"
)

func TestGetTestFuncName(t *testing.T) {
	tests := []struct {
		name     string
		path     []string
		variant  *Variant
		expected string
	}{
		{
			name:     "basic path",
			path:     []string{"root", "child", "leaf"},
			expected: "TestRoot_Child_Leaf",
		},
		{
			name:     "empty name becomes Unamed",
			path:     []string{"root", "", "leaf"},
			expected: "TestRoot_Unamed_Leaf",
		},
		{
			name:     "special characters are stripped",
			path:     []string{"root$", "child?", "leaf#"},
			expected: "TestRoot_Child_Leaf",
		},
		{
			name:     "spaces and hyphens are removed",
			path:     []string{"root test", "child-name", "leaf_case"},
			expected: "TestRoottest_Childname_Leafcase",
		},
		{
			name:     "numbers are preserved except at start",
			path:     []string{"1root", "child2", "3leaf"},
			expected: "Test1root_Child2_3leaf",
		},
		{
			name: "with variant",
			path: []string{"root", "child"},
			variant: &Variant{
				ShortestName: "var1",
			},
			expected: "TestRoot_Child_Var1",
		},
		{
			name: "variant with special chars",
			path: []string{"root"},
			variant: &Variant{
				ShortestName: "var$1@test",
			},
			expected: "TestRoot_Var1test",
		},
		{
			name:     "single element path",
			path:     []string{"root"},
			expected: "TestRoot",
		},
		{
			name:     "all special chars become underscore",
			path:     []string{"!@#"},
			expected: "Test",
		},
		{
			name:     "greater than with percentage",
			path:     []string{"A > 20%"},
			expected: "TestAᐳ20ᵖ",
		},
		{
			name:     "less than with percentage",
			path:     []string{"value < 50%"},
			expected: "TestValueᐸ50ᵖ",
		},
		{
			name:     "arithmetic operations",
			path:     []string{"x + y", "result * 2", "a / b"},
			expected: "TestXᐩy_Resultˣ2_Aᐟb",
		},
		{
			name:     "complex comparison",
			path:     []string{"age >= 18", "score = 100%"},
			expected: "TestAgeᐳᗕ18_Scoreᗕ100ᵖ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTestFuncName(tt.path, tt.variant)
			if got != tt.expected {
				t.Errorf("GetTestFuncName() = %v, want %v", got, tt.expected)
			}
		})
	}
}
