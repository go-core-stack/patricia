// Copyright © 2025 Prabhjot Singh Sethi, All Rights reserved
// Author: Prabhjot Singh Sethi <prabhjot.sethi@gmail.com>

package main

import (
	"fmt"

	"github.com/go-core-stack/patricia"
)

func main() {
	// Create a new Patricia tree for int values
	tree := patricia.NewSimpleTree[int]()

	// Insert key-value pairs
	tree.Insert("foo", 123)
	tree.Insert("bar", 456)

	// Longest prefix search
	if val, ok := tree.PrefixSearch("foo"); ok {
		fmt.Println("foo:", val) // Output: foo: 123
	}
	if val, ok := tree.PrefixSearch("foobar"); ok {
		fmt.Println("foobar:", val) // Output: foobar: 123
	}
	if val, ok := tree.PrefixSearch("bar"); ok {
		fmt.Println("bar:", val) // Output: bar: 456
	}
	if _, ok := tree.PrefixSearch("xyz"); !ok {
		fmt.Println("xyz not found")
	}

	// Remove a key
	tree.Remove("bar")

	// In-order traversal
	for key, value := range tree.All() {
		fmt.Printf("%s: %d\n", key, value)
	}
}
