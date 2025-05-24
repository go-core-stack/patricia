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

## Implementation Details

- **Node Structure:** Each node represents a common prefix and may store a value. Child nodes are indexed by the next character(s) in the key.
- **Path Compression:** Shared prefixes are stored only once, reducing memory usage and improving cache locality.
- **Concurrency:** The implementation is safe for concurrent read operations. For concurrent writes, external synchronization is required.
- **API Design:** The library exposes a simple and idiomatic Go API for insertion, lookup, deletion, and traversal.

## Usage

### Installation

```sh
go get github.com/go-core-stack/patricia
```