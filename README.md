# Patricia Tree for Go

A high-performance, space-efficient [Patricia tree](https://en.wikipedia.org/wiki/Radix_tree)
(Practical Algorithm to Retrieve Information Coded in Alphanumeric) implementation in Go.

## Overview

Patricia trees, also known as radix trees or prefix trees, are a specialized data structure
for storing associative arrays where the keys are usually strings. They are particularly
efficient for scenarios involving prefix matching, such as IP routing tables, autocomplete,
and dictionary implementations.

This repository provides a robust and idiomatic Go implementation of Patricia trees,
suitable for use in production systems.

## Features

- **Efficient Prefix Matching:** Quickly find all keys sharing a common prefix.
- **Space Optimization:** Compresses common prefixes to minimize memory usage.
- **Fast Lookup, Insert, and Delete:** Operations are performed in O(k) time, where k is
  the length of the key.
- **Ordered Traversal:** Supports in-order traversal of keys.
- **Generic API:** The library provides a generic, type-safe interface for storing values of any type.
- **SimpleTree Interface:** A simplified interface for string keys, supporting insertion, removal, longest prefix search, and in-order traversal.

## Implementation Details

- **Node Structure:** Each node represents a common prefix and may store a value. Child nodes are indexed by the next character(s) in the key.
- **Path Compression:** Shared prefixes are stored only once, reducing memory usage and improving cache locality.
- **Concurrency:** The implementation is safe for concurrent read operations. For concurrent writes, external synchronization is required.
- **API Design:** The library exposes a simple and idiomatic Go API for insertion, lookup, deletion, and traversal, including a `SimpleTree` interface for ease of use with string keys.

## Usage

### Installation

```sh
go get github.com/go-core-stack/patricia
```

### Basic Example

```go
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
    for key, value := range tree.All(){
        fmt.Printf("%s: %d\n", key, value)
    }
}
```

### API Overview

- `Insert(key string, data D) bool`: Adds or updates a key-value pair. Returns true if inserted, false if updated.
- `Remove(key string) bool`: Deletes a key. Returns true if removed, false if not found.
- `PrefixSearch(key string) (D, bool)`: Finds the value for the longest prefix match. Returns value and whether found.
- `All() func(func(string, D) bool)`: Iterates over all key-value pairs in order. Stops early if the callback returns false.
