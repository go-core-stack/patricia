// Copyright © 2025 Prabhjot Singh Sethi, All Rights reserved
// Author: Prabhjot Singh Sethi <prabhjot.sethi@gmail.com>

package patricia

import (
	"testing"
)

// key represents a string key for the Patricia tree.
type key struct {
	val string
	len int
}

// BitLength returns the bit length of the key.
func (k key) BitLength() int {
	if k.len != 0 {
		return k.len << 3
	}
	return (len(k.val) << 3)
}

// ByteValue returns the byte at the given position in the key.
func (k key) ByteValue(pos int) byte {
	if pos < 0 || pos >= len(k.val) {
		return 0
	}
	return k.val[pos]
}

// data represents the value stored in the Patricia tree.
type data struct {
	val string
}

// tree is the global Patricia tree instance used in tests.
var tree *treeImpl[key, data]

// initializeTree resets the global tree before each test.
func initializeTree() {
	if tree == nil {
		tree = &treeImpl[key, data]{}
	}
}

// Test_BasicInsertRemoves tests basic insert and remove operations.
func Test_BasicInsertRemoves(t *testing.T) {
	initializeTree()

	ret := tree.Insert(&key{val: "abc"}, &data{"abc"})
	if !ret {
		t.Errorf("Failed to insert entry, Expected true, got false")
	}

	ret = tree.Insert(&key{val: "abc"}, &data{"abc2"})
	if ret {
		t.Errorf("it should have failed to insert the entry, Expected false, got true")
	}

	if tree.nodes_ != 1 {
		t.Errorf("Node count Expected 1, got %d", tree.nodes_)
	}

	ret = tree.Remove(&key{val: "abc"})
	if !ret {
		t.Errorf("Failed to remove entry, Expected true, got false")
	}

	// Remove again, should fail
	ret = tree.Remove(&key{val: "abc"})
	if ret {
		t.Errorf("Should not remove non-existent entry, Expected false, got true")
	}

	// Ensure both nodes_ and intNodes_ are zero after removals
	if tree.nodes_ != 0 {
		t.Errorf("Node count Expected 0, got %d", tree.nodes_)
	}
	if tree.intNodes_ != 0 {
		t.Errorf("Internal node count Expected 0, got %d", tree.intNodes_)
	}
}

// Test_All tests insertion of multiple keys and the All() method.
func Test_All(t *testing.T) {
	initializeTree()
	ret := tree.Insert(&key{val: "abc"}, &data{"abc"})
	if !ret {
		t.Errorf("Failed to insert entry, Expected true, got false")
	}
	ret = tree.Insert(&key{val: "abc"}, &data{"abc2"})
	if ret {
		t.Errorf("it should have failed to insert the entry, Expected false, got true")
	}
	tree.Insert(&key{val: "aabc"}, &data{"aabc"})
	tree.Insert(&key{val: "aabbc"}, &data{"aabbc"})
	tree.Insert(&key{val: "aabbcc"}, &data{"aabbcc"})
	tree.Insert(&key{val: "aabbccd"}, &data{"aabbccd"})
	tree.Insert(&key{val: "aabbccdd"}, &data{"aabbccdd"})

	if tree.nodes_ != 6 {
		t.Errorf("Node count Expected 6, got %d", tree.nodes_)
	}

	expected := map[string]string{
		"abc":      "abc",
		"aabc":     "aabc",
		"aabbc":    "aabbc",
		"aabbcc":   "aabbcc",
		"aabbccd":  "aabbccd",
		"aabbccdd": "aabbccdd",
	}

	count := 0
	for k, v := range tree.All() {
		if expected[k.val] != v.val {
			t.Errorf("Expected value %s for key %s, got %s", expected[k.val], k.val, v.val)
		}
		count++
	}

	if count != len(expected) {
		t.Errorf("Expected %d keys, got %d", len(expected), count)
	}

	// Remove all keys and check node counts
	tree.Remove(&key{val: "abc"})
	if tree.nodes_ != 5 {
		t.Errorf("Node count Expected 5, got %d", tree.nodes_)
	}
	tree.Remove(&key{val: "aabc"})
	tree.Remove(&key{val: "aabbc"})
	tree.Remove(&key{val: "aabbcc"})
	tree.Remove(&key{val: "aabbccd"})
	tree.Remove(&key{val: "aabbccdd"})
	if tree.nodes_ != 0 {
		t.Errorf("Node count Expected 0, got %d", tree.nodes_)
	}
	if tree.intNodes_ != 0 {
		t.Errorf("Internal node count Expected 0, got %d", tree.intNodes_)
	}
}

// Test_EmptyKeyAndValue tests inserting and removing an empty key and value.
func Test_EmptyKeyAndValue(t *testing.T) {
	initializeTree()
	ret := tree.Insert(&key{val: ""}, &data{""})
	if !ret {
		t.Errorf("Failed to insert empty key/value")
	}
	if tree.nodes_ != 1 {
		t.Errorf("Node count Expected 1, got %d", tree.nodes_)
	}
	ret = tree.Remove(&key{val: ""})
	if !ret {
		t.Errorf("Failed to remove empty key")
	}
	// Ensure both nodes_ and intNodes_ are zero after removals
	if tree.nodes_ != 0 {
		t.Errorf("Node count Expected 0, got %d", tree.nodes_)
	}
	if tree.intNodes_ != 0 {
		t.Errorf("Internal node count Expected 0, got %d", tree.intNodes_)
	}
}

// Test_RemoveNonExistentKey tests removing a key that does not exist.
func Test_RemoveNonExistentKey(t *testing.T) {
	initializeTree()
	tree.Insert(&key{val: "abc"}, &data{"abc"})
	ret := tree.Remove(&key{val: "notfound"})
	if ret {
		t.Errorf("Should not remove non-existent key")
	}
	tree.Remove(&key{val: "abc"})
}

// Test_LPMFind tests LPMFind for missing and partial matches.
func Test_LPMFind(t *testing.T) {
	initializeTree()
	keys := []string{"a", "ab", "abc", "abcd", "aabc", "aabcd"}
	for _, k := range keys {
		tree.Insert(&key{val: k}, &data{k})
	}

	// Should return nil for a completely missing key
	node := tree.LPMFind(&key{val: "xyz"})
	if node != nil {
		t.Errorf("Expected nil for non-existent best match")
	}

	// Should return the best prefix match ("a")
	node = tree.LPMFind(&key{val: "adfgd"})
	if node == nil {
		t.Errorf("failed to get best match, got nil")
	} else if node.val != "a" {
		t.Errorf("failed to get best match: %v", node)
	}

	// Remove all keys after test
	for _, k := range keys {
		tree.Remove(&key{val: k})
	}
}

// Test_AllAfterRemovals tests that All() returns nothing after all keys are removed.
func Test_AllAfterRemovals(t *testing.T) {
	initializeTree()
	keys := []string{"a", "ab", "abc"}
	for _, k := range keys {
		tree.Insert(&key{val: k}, &data{k})
	}
	for _, k := range keys {
		tree.Remove(&key{val: k})
	}
	if tree.nodes_ != 0 {
		t.Errorf("Node count Expected 0, got %d", tree.nodes_)
	}
	if tree.intNodes_ != 0 {
		t.Errorf("Internal node count Expected 0, got %d", tree.intNodes_)
	}
}
