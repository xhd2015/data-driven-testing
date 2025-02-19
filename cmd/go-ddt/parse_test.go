package main

import (
	"reflect"
	"testing"
)

func TestShortestUncommonName(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "empty input",
			input:    []string{},
			expected: nil,
		},
		{
			name:     "single name",
			input:    []string{"TestFoo"},
			expected: []string{"TestFoo"},
		},
		{
			name:     "common prefix",
			input:    []string{"TestFoo", "TestBar", "TestBaz"},
			expected: []string{"Foo", "Bar", "Baz"},
		},
		{
			name:     "common suffix",
			input:    []string{"aTest", "bTest", "cTest"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "common prefix and suffix",
			input:    []string{"TestFooBar", "TestBazBar", "TestQuxBar"},
			expected: []string{"Foo", "Baz", "Qux"},
		},
		{
			name:     "no common parts",
			input:    []string{"Foo", "Bar", "Baz"},
			expected: []string{"Foo", "Bar", "Baz"},
		},
		{
			name:     "identical names",
			input:    []string{"Test", "Test", "Test"},
			expected: []string{"", "", ""},
		},
		{
			name:     "names with different lengths",
			input:    []string{"TestLongName", "TestShort", "TestMediumName"},
			expected: []string{"LongName", "Short", "MediumName"},
		},
		{
			name:     "partial common parts",
			input:    []string{"TestAFoo", "TestBFoo", "TestCBar"},
			expected: []string{"AFoo", "BFoo", "CBar"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shortestUncommonName(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("shortestUncommonName() = %v, want %v", got, tt.expected)
			}
		})
	}
}
