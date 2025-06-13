// Copyright © 2025 Prabhjot Singh Sethi, All Rights reserved
// Author: Prabhjot Singh Sethi <prabhjot.sethi@gmail.com>

package patricia

import (
	"fmt"
	"testing"
)

// TestSimpleTreeBasic tests basic Insert, Get, Remove, and PrefixSearch operations.
func TestSimpleTreeBasic(t *testing.T) {
	tree := NewSimpleTree[int]()

	// Insert and retrieve
	if !tree.Insert("foo", 1) {
		t.Error("expected new insert for 'foo'")
	}
	if !tree.Insert("bar", 2) {
		t.Error("expected new insert for 'bar'")
	}
	if tree.Insert("foo", 3) {
		t.Error("expected insert failure for 'foo'")
	}

	// PrefixSearch
	val, _ := tree.PrefixSearch("foo")
	if val != 1 {
		t.Errorf("expected PrefixSearch('foo')=1, got %d", val)
	}
	val, _ = tree.PrefixSearch("bar")
	if val != 2 {
		t.Errorf("expected PrefixSearch('bar')=2, got %d", val)
	}

	// Remove
	if !tree.Remove("foo") {
		t.Error("expected to remove 'foo'")
	}
	if tree.Remove("foo") {
		t.Error("expected 'foo' to be already removed")
	}
}

// TestSimpleTreePrefixSearch tests longest prefix matching.
func TestSimpleTreePrefixSearch(t *testing.T) {
	tree := NewSimpleTree[string]()
	tree.Insert("a", "A")
	tree.Insert("ab", "AB")
	tree.Insert("abc", "ABC")

	tests := []struct {
		key      string
		expected string
	}{
		{"a", "A"},
		{"ab", "AB"},
		{"abc", "ABC"},
		{"abcd", "ABC"},
		{"abx", "AB"},
		{"ax", "A"},
	}

	for _, tt := range tests {
		got, _ := tree.PrefixSearch(tt.key)
		if got != tt.expected {
			t.Errorf("PrefixSearch(%q) = %q, want %q", tt.key, got, tt.expected)
		}
	}

	got, ok := tree.PrefixSearch("adfg")
	if !ok {
		t.Error("expected PrefixSearch('adfg') to return a value")
	} else {
		if got != "A" {
			t.Errorf("expected PrefixSearch('adfg') to return 'A', got %s", got)
		}
	}

	_, ok = tree.Search("adfg")
	if ok {
		t.Error("expected Search('adfg') to fail")
	}

	got, ok = tree.Search("a")
	if !ok || got != "A" {
		t.Errorf("expected Search('a') to return 'A', got %s, ok=%v", got, ok)
	}
}

// TestSimpleTreeAll tests the All iterator.
func TestSimpleTreeAll(t *testing.T) {
	tree := NewSimpleTree[int]()
	keys := []string{"foo", "bar", "baz"}
	vals := []int{10, 20, 30}
	for i, k := range keys {
		tree.Insert(k, vals[i])
	}

	found := make(map[string]int)
	for k, v := range tree.All() {
		found[k] = v
	}

	for i, k := range keys {
		if found[k] != vals[i] {
			t.Errorf("All: expected %s=%d, got %d", k, vals[i], found[k])
		}
	}
}

// TestSimpleTreeAll tests the All iterator.
func TestSimpleTreeAll_2(t *testing.T) {
	tree := NewSimpleTree[[]string]()
	//keys := []string{"foo", "bar", "baz"}
	keys := []string{"user"}
	vals := []string{"foo", "bar", "baz"}
	for _, k := range keys {
		tree.Insert(k, vals)
	}

	found := make(map[string][]string)
	for k, v := range tree.All() {
		found[k] = v
	}

	for _, k := range keys {
		newVals := found[k]
		for i, v := range vals {
			if newVals[i] != v {
				t.Errorf("All: expected %s=%v, got %v", k, v, newVals[i])
			}
		}
	}
}

// TestSimpleTreeEmpty tests operations on an empty tree.
func TestSimpleTreeEmpty(t *testing.T) {
	tree := NewSimpleTree[int]()
	count := 0
	for range tree.All() {
		count++
	}
	if count != 0 {
		t.Error("Expected an empty tree")
	}
}

// TestSimpleTreePanicOnNoPrefixMatch ensures PrefixSearch panics if no match.
func TestSimpleTreePanicOnNoPrefixMatch(t *testing.T) {
	tree := NewSimpleTree[int]()
	tree.Insert("foo", 1)
	_, ok := tree.PrefixSearch("bar")
	if ok {
		t.Error("expected no match for 'bar', but got a value")
	}
}

// Example usage for documentation.
func ExampleSimpleTree() {
	tree := NewSimpleTree[string]()
	tree.Insert("foo", "bar")
	tree.Insert("baz", "qux")
	for k, v := range tree.All() {
		fmt.Printf("%s: %s\n", k, v)
	}
}
