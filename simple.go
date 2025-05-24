// Copyright © 2025 Prabhjot Singh Sethi, All Rights reserved
// Author: Prabhjot Singh Sethi <prabhjot.sethi@gmail.com>

package patricia

// stringKey is a wrapper for string keys used in the Patricia tree.
// It implements the required interface for key operations, allowing
// the Patricia tree to work efficiently with string keys.
type stringKey struct {
	val string
}

// BitLength returns the length of the key in bits.
// This is used by the Patricia tree to determine the key length for traversal
// and bitwise operations.
func (k stringKey) BitLength() int {
	return (len(k.val) << 3)
}

// ByteValue returns the byte value at the specified position in the key.
// If the position is out of bounds, it returns 0. This method is used
// internally by the Patricia tree for key comparison and traversal.
func (k stringKey) ByteValue(pos int) byte {
	if pos < 0 || pos >= len(k.val) {
		return 0
	}
	return k.val[pos]
}

// SimpleTree provides a simplified Patricia tree interface for string keys.
// It is a generic interface that allows storing values of any type D.
// The interface exposes basic operations for insertion, removal, prefix search,
// and in-order traversal.
//
// Example usage:
//
//	tree := patricia.NewSimpleTree[int]()
//	tree.Insert("foo", 123)
//	tree.Insert("bar", 456)
//	val, ok := tree.PrefixSearch("foo") // returns 123, true
//	val, ok := tree.PrefixSearch("foobar") // returns 123, true
//	val, ok := tree.PrefixSearch("bar") // returns 456, true
//	val, ok := tree.PrefixSearch("xyz") // returns 0, false
//	tree.Remove("bar")
type SimpleTree[D any] interface {
	// Insert adds a key-value pair to the Patricia tree.
	// Returns true if the key was newly inserted, false if it was updated.
	Insert(key string, data D) bool

	// Remove deletes the specified key from the Patricia tree.
	// Returns true if the key was present and removed, false otherwise.
	Remove(key string) bool

	// PrefixSearch finds the value associated with the longest prefix match for the given key.
	// Returns the value for the best matching prefix. returns found as false if no match is found.
	PrefixSearch(key string) (D, bool)

	// All returns a function that iterates over all key-value pairs in the tree in order.
	// The provided yield function is called for each key-value pair.
	// If yield returns false, iteration stops early.
	// Example:
	//
	// for k, v := range tree.All() {
	//     fmt.Println(k, v)
	// }
	All() func(func(string, D) bool)
}

// simpleTree provides a concrete implementation of the SimpleTree interface for string keys.
// It wraps the generic treeImpl type, allowing users to work directly with string keys
// and values of any type D.
type simpleTree[D any] struct {
	tree *treeImpl[stringKey, D]
}

// Insert adds a key-value pair to the Patricia tree.
// Returns true if the key was newly inserted, false if it was updated.
func (t *simpleTree[D]) Insert(key string, data D) bool {
	return t.tree.Insert(&stringKey{val: key}, &data)
}

// Remove deletes the specified key from the Patricia tree.
// Returns true if the key was present and removed, false otherwise.
func (t *simpleTree[D]) Remove(key string) bool {
	return t.tree.Remove(&stringKey{val: key})
}

// PrefixSearch finds the value associated with the longest prefix match for the given key.
// Returns the value for the best matching prefix. returns found as false if no match is found.
func (t *simpleTree[D]) PrefixSearch(key string) (D, bool) {
	var zero D
	val := t.tree.LPMFind(&stringKey{val: key})
	if val != nil {
		return *val, true
	}
	return zero, false
}

// All returns a function that iterates over all key-value pairs in the tree in order.
// The provided yield function is called for each key-value pair.
// If yield returns false, iteration stops early.
//
// Example:
//
//	for k, v := range tree.All() {
//	    fmt.Println(k, v)
//	}
func (t *simpleTree[D]) All() func(func(string, D) bool) {
	return func(yield func(string, D) bool) {
		if t.tree.root_ == nil {
			// If the tree is empty, nothing to yield.
			return
		}
		// Start with the leftmost node (smallest key).
		n := t.tree.getNextNode(nil)
		for n != nil {
			// Only yield nodes that have both key and data set (i.e., leaf nodes).
			if n.data_ != nil && n.key_ != nil {
				if !yield(n.key_.val, *n.data_) {
					// If yield returns false, stop iteration early.
					return
				}
			}
			// Move to the next node in order.
			n = t.tree.getNextNode(n)
		}
	}
}

// NewSimpleTree creates and returns a new SimpleTree instance for string keys and values of type D.
//
// Example:
//
//	tree := patricia.NewSimpleTree[int]()
func NewSimpleTree[D any]() SimpleTree[D] {
	return &simpleTree[D]{
		tree: &treeImpl[stringKey, D]{},
	}
}
