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

	if tree.nodes != 1 {
		t.Errorf("Node count Expected 1, got %d", tree.nodes)
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

	// Ensure both nodes and intNodes are zero after removals
	if tree.nodes != 0 {
		t.Errorf("Node count Expected 0, got %d", tree.nodes)
	}
	if tree.intNodes != 0 {
		t.Errorf("Internal node count Expected 0, got %d", tree.intNodes)
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

	if tree.nodes != 6 {
		t.Errorf("Node count Expected 6, got %d", tree.nodes)
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
	if tree.nodes != 5 {
		t.Errorf("Node count Expected 5, got %d", tree.nodes)
	}
	tree.Remove(&key{val: "aabc"})
	tree.Remove(&key{val: "aabbc"})
	tree.Remove(&key{val: "aabbcc"})
	tree.Remove(&key{val: "aabbccd"})
	tree.Remove(&key{val: "aabbccdd"})
	if tree.nodes != 0 {
		t.Errorf("Node count Expected 0, got %d", tree.nodes)
	}
	if tree.intNodes != 0 {
		t.Errorf("Internal node count Expected 0, got %d", tree.intNodes)
	}
}

// Test_EmptyKeyAndValue tests inserting and removing an empty key and value.
func Test_EmptyKeyAndValue(t *testing.T) {
	initializeTree()
	ret := tree.Insert(&key{val: ""}, &data{""})
	if !ret {
		t.Errorf("Failed to insert empty key/value")
	}
	if tree.nodes != 1 {
		t.Errorf("Node count Expected 1, got %d", tree.nodes)
	}
	ret = tree.Remove(&key{val: ""})
	if !ret {
		t.Errorf("Failed to remove empty key")
	}
	// Ensure both nodes and intNodes are zero after removals
	if tree.nodes != 0 {
		t.Errorf("Node count Expected 0, got %d", tree.nodes)
	}
	if tree.intNodes != 0 {
		t.Errorf("Internal node count Expected 0, got %d", tree.intNodes)
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
	if tree.nodes != 0 {
		t.Errorf("Node count Expected 0, got %d", tree.nodes)
	}
	if tree.intNodes != 0 {
		t.Errorf("Internal node count Expected 0, got %d", tree.intNodes)
	}
}

// testKey is a special key type for testing non-byte-aligned bit lengths
type testKey struct {
	bytes     []byte
	bitLength int
}

func (k testKey) BitLength() int {
	return k.bitLength
}

func (k testKey) ByteValue(pos int) byte {
	if pos < 0 || pos >= len(k.bytes) {
		return 0
	}
	return k.bytes[pos]
}

func Test_InsertFindWithNonByteAligned(t *testing.T) {
	tree := &treeImpl[testKey, data]{}

	// Both keys have 13 bits (1 full byte + 5 remaining bits)
	// Same first byte (0xFF) but differ in remaining 5 bits
	// k1: 0xFF (11111111) + first 5 bits of 0xF8 (11111000) = 11111111 11111
	// k2: 0xFF (11111111) + first 5 bits of 0x00 (00000000) = 11111111 00000
	k1 := testKey{bytes: []byte{0xFF, 0xF8}, bitLength: 13}
	k2 := testKey{bytes: []byte{0xFF, 0x00}, bitLength: 13}

	d1 := &data{val: "data1"}

	// Insert k1 first
	if !tree.Insert(&k1, d1) {
		t.Error("Failed to insert k1")
	}

	// FindNode should find k1 and return data1
	result := tree.FindNode(&k1)
	if result == nil || result.val != "data1" {
		t.Error("Failed to find k1")
	}

	// FindNode with k2 should return nil (not inserted yet)
	// But with the bug in compareNodes, it might incorrectly return data1
	result = tree.FindNode(&k2)
	if result != nil {
		t.Errorf("BUG: FindNode(&k2) returned %v, expected nil (compareNodes bug: remaining bits not compared)", result.val)
	}

	if !tree.Remove(&k1) {
		t.Errorf("Failed to remove key k1")
	}
}

// Test_CompareNodesWithNonByteAligned tests compareNodes with non-byte-aligned keys.
// This test ensures the bug fix for comparing remaining bits works correctly.
// The bug was: the loop compared pos (in bits) against byteLen (in bytes),
// so remaining bits after full bytes were never compared.
func Test_CompareNodesWithNonByteAligned(t *testing.T) {
	tree := &treeImpl[testKey, data]{}

	// Test Case 1: Keys with 13 bits (1 full byte + 5 remaining bits)
	// Both keys have same first byte (0xFF) but differ in the remaining 5 bits
	// Key 1: 0xFF (11111111) + first 5 bits of 0xF8 (11111000) = 11111111 11111
	// Key 2: 0xFF (11111111) + first 5 bits of 0x00 (00000000) = 11111111 00000
	k1 := testKey{bytes: []byte{0xFF, 0xF8}, bitLength: 13}
	k2 := testKey{bytes: []byte{0xFF, 0x00}, bitLength: 13}

	n1 := &node[testKey, data]{key: &k1}
	n2 := &node[testKey, data]{key: &k2}

	// These keys differ only in the remaining 5 bits (bits 8-12)
	// Without the fix, compareNodes would return true (bug!)
	// With the fix, it should return false
	if tree.compareNodes(n1, n2) {
		t.Error("Expected nodes with different remaining bits to not be equal (BUG: remaining bits not compared)")
	}

	// Test Case 2: Keys with 13 bits that are identical
	k3 := testKey{bytes: []byte{0xFF, 0xF8}, bitLength: 13}
	k4 := testKey{bytes: []byte{0xFF, 0xF8}, bitLength: 13}

	n3 := &node[testKey, data]{key: &k3}
	n4 := &node[testKey, data]{key: &k4}

	if !tree.compareNodes(n3, n4) {
		t.Error("Expected nodes with same keys to be equal")
	}

	// Test Case 3: Keys with 10 bits (1 full byte + 2 remaining bits)
	// Key 5: 0xAA (10101010) + first 2 bits of 0xC0 (11000000) = 10101010 11
	// Key 6: 0xAA (10101010) + first 2 bits of 0x00 (00000000) = 10101010 00
	k5 := testKey{bytes: []byte{0xAA, 0xC0}, bitLength: 10}
	k6 := testKey{bytes: []byte{0xAA, 0x00}, bitLength: 10}

	n5 := &node[testKey, data]{key: &k5}
	n6 := &node[testKey, data]{key: &k6}

	// Differ in bits 8-9
	if tree.compareNodes(n5, n6) {
		t.Error("Expected nodes with different remaining bits to not be equal (BUG: remaining bits not compared)")
	}

	// Test nil cases
	if tree.compareNodes(nil, n1) {
		t.Error("Expected compareNodes with nil to return false")
	}
	if tree.compareNodes(n1, nil) {
		t.Error("Expected compareNodes with nil to return false")
	}
}
