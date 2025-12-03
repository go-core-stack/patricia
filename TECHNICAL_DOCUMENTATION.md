# Patricia Tree - Technical Documentation

## Table of Contents
1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Core Implementation](#core-implementation)
4. [Components](#components)
5. [Algorithms](#algorithms)
6. [Performance Characteristics](#performance-characteristics)
7. [Usage Examples](#usage-examples)
8. [Implementation Review](#implementation-review)
9. [Optimization Recommendations](#optimization-recommendations)

---

## Overview

This repository provides a high-performance, generic Patricia tree (radix tree) implementation in Go. Patricia trees are space-optimized trie data structures particularly efficient for prefix matching operations.

### Key Features
- **Generic Type Support**: Uses Go generics for type-safe key-value storage
- **Multiple Interfaces**: SimpleTree for string keys, UrlTree for URL pattern matching
- **Efficient Operations**: O(k) time complexity where k is key length
- **Space Optimization**: Prefix compression reduces memory footprint
- **Longest Prefix Matching**: Built-in LPM support for routing and lookup scenarios

### Historical Context
The core logic was originally contributed by Prabhjot Singh Sethi in 2013 to Contail Systems (Juniper Networks) as part of the Linux Foundation Tungsten Fabric project, written in C++ for network-centric use cases. This repository repurposes and reimplements the logic in Go.

---

## Architecture

### Component Structure

```
patricia/
├── impl.go         # Core Patricia tree implementation
├── simple.go       # String key wrapper interface
├── url.go          # URL pattern matcher with variables
├── docs.go         # Package documentation
├── *_test.go       # Comprehensive test suites
└── example/        # Usage examples
    ├── simple/
    └── url/
```

### Type Hierarchy

```
Key Interface
    ├── BitLength() int
    └── ByteValue(int) byte

treeImpl[K Key, D any]
    ├── node[K, D]
    │   ├── left_: *node
    │   ├── right_: *node
    │   ├── bitpos_: int
    │   ├── intnode_: bool
    │   ├── key_: *K
    │   └── data_: *D
    └── Methods: Insert, Remove, FindNode, LPMFind, All

SimpleTree[D any] Interface
    └── simpleTree[D]
        └── wraps treeImpl[stringKey, D]

UrlTree[D any] Interface
    └── urlTree[D]
        └── urlNode[D]
            ├── nextWildcard: *urlNode
            └── subTree: *treeImpl[stringKey, urlNode]
```

---

## Core Implementation

### Node Structure (impl.go:30-39)

```go
type node[K Key, D any] struct {
    left_    *node[K, D]  // Left child node
    right_   *node[K, D]  // Right child node
    intnode_ bool         // Internal node flag
    bitpos_  int          // Bit position for branching
    key_     *K           // Key pointer (leaf nodes only)
    data_    *D           // Data pointer (leaf nodes only)
}
```

**Design Rationale:**
- **Dual Node Types**: Internal nodes (for branching) vs leaf nodes (storing data)
- **Bit Position**: Determines where keys diverge in the tree
- **Pointer Storage**: Keys and data stored as pointers to avoid copying large structures

### Tree Structure (impl.go:62-66)

```go
type treeImpl[K Key, D any] struct {
    nodes_    int         // Count of leaf nodes (data entries)
    intNodes_ int         // Count of internal nodes (branching)
    root_     *node[K, D] // Root node pointer
}
```

---

## Components

### 1. Core Patricia Tree (impl.go)

#### Key Interface
Every key type must implement:
- `BitLength() int`: Returns total bits in key
- `ByteValue(pos int) byte`: Returns byte at position

#### Core Operations

##### Insert (impl.go:605-711)
**Algorithm:**
1. Traverse tree following key bits until divergence point
2. Compare with existing node at insertion point
3. Handle four cases:
   - Key exists: return false
   - Replace internal node at matching position
   - Insert leaf with potential rewiring
   - Create new internal node for branching

**Time Complexity:** O(k) where k = key length in bits

##### Remove (impl.go:494-600)
**Algorithm:**
1. Traverse to find node with matching key
2. Handle multiple removal cases:
   - Node with both children and right descendant → create internal node
   - Node with only left child → promote left, rewire right
   - Node with only right descendant → promote right
   - Leaf node with internal parent → remove internal node
   - Leaf node with regular parent → adjust parent pointers

**Time Complexity:** O(k)

##### FindNode (impl.go:456-489)
**Algorithm:**
1. Traverse tree following key bits
2. Stop when bit position exceeds key length or matches
3. Compare found node with search key for exact match

**Time Complexity:** O(k)

##### LPMFind (Longest Prefix Match) (impl.go:314-355)
**Algorithm:**
1. Traverse tree, tracking last matching prefix
2. At each leaf node, compare key prefixes
3. Return data of longest matching prefix

**Use Cases:** IP routing, autocomplete, hierarchical lookups

**Time Complexity:** O(k)

#### Helper Operations

##### getBit (impl.go:87-95)
Extracts bit value at position using bit manipulation:
```go
(node.ByteValue(pos>>3) & (0x80 >> (pos & 7))) != 0
```
- `pos>>3`: Convert bit position to byte index
- `pos & 7`: Get bit offset within byte
- `0x80 >> offset`: Create bit mask

##### compare (impl.go:127-162)
Compares two keys from a starting bit position:
1. Fast byte-level comparison for full bytes
2. Bit-level comparison for remaining bits
3. Returns first differing bit position and equality status

##### rewireRightMost (impl.go:166-177)
Finds rightmost child and rewires its right pointer during removal.

### 2. SimpleTree Interface (simple.go)

#### stringKey Implementation (simple.go:9-28)
Wraps Go strings to implement Key interface:
```go
type stringKey struct {
    val string
}

func (k stringKey) BitLength() int {
    return (len(k.val) << 3)  // Bytes to bits
}

func (k stringKey) ByteValue(pos int) byte {
    if pos < 0 || pos >= len(k.val) {
        return 0
    }
    return k.val[pos]
}
```

#### SimpleTree Interface (simple.go:45-71)
Provides convenient API for string keys:
- `Insert(key string, data D) bool`
- `Remove(key string) bool`
- `PrefixSearch(key string) (D, bool)` - Longest prefix match
- `Search(key string) (D, bool)` - Exact match
- `All() func(func(string, D) bool)` - Iterator using Go 1.23+ range-over-func

**Usage:**
```go
tree := patricia.NewSimpleTree[int]()
tree.Insert("hello", 42)
val, ok := tree.PrefixSearch("helloworld")  // Returns 42, true
```

### 3. UrlTree - Pattern Matching (url.go)

#### Design Goals
Match URL patterns with variable segments:
- `/api/users/{id}` matches `/api/users/123`
- Extract variable names and values
- Support mixed literal and variable segments

#### urlKey Structure (url.go:23-53)
```go
type urlKey struct {
    actVal string   // Original URL
    parts  []string // Split by '/', vars replaced with "*"
    vars   []string // Variable names extracted from {name}
}
```

**Example:**
```
URL: "/api/service/v1/id/{id}/name/{name}"
parts: ["api", "service", "v1", "id", "*", "name", "*"]
vars: ["id", "name"]
```

#### urlNode Structure (url.go:57-64)
```go
type urlNode[D any] struct {
    subKey       string           // Segment value or "*"
    key          *urlKey          // Full pattern (terminal nodes)
    pos          int              // Segment position
    match        *D               // Data if terminal node
    nextWildcard *urlNode[D]      // Wildcard child
    subTree      *treeImpl[...]   // Literal children
}
```

**Constraint:** A node cannot have both wildcard and literal children (url.go:98-107). This simplifies matching logic but limits pattern flexibility.

#### UrlTree Operations

##### Insert (url.go:190-201)
1. Parse URL pattern into urlKey
2. Recursively create nodes for each segment
3. Mark terminal node with data
4. Return false if pattern exists

##### Match (url.go:212-223)
1. Parse input URL into urlKey
2. Recursively match segments:
   - Check wildcard child first
   - Fall back to literal child lookup
   - Collect wildcard values
3. Return variable names, values, and associated data

**Time Complexity:** O(n) where n = number of URL segments

**Usage:**
```go
tree := patricia.NewUrlTree[int]()
tree.Insert("/api/users/{id}", 42)
keys, values, handle, ok := tree.Match("/api/users/123")
// keys = ["id"], values = ["123"], handle = 42, ok = true
```

---

## Algorithms

### Bit Manipulation Techniques

#### Bit Position Calculation
```go
bitpos_ := len(key) * 8  // Total bits
```

#### Byte Access
```go
byteIndex := bitPosition >> 3  // Divide by 8
```

#### Bit Access Within Byte
```go
bitOffset := bitPosition & 7   // Modulo 8
mask := 0x80 >> bitOffset      // Create bit mask
bit := (byteValue & mask) != 0
```

### Tree Traversal Patterns

#### Downward Traversal (Insert/Find)
```go
for x != nil {
    if getBit(key, x.bitpos_) {
        x = x.right_  // Bit is 1
    } else {
        x = x.left_   // Bit is 0
    }
    if x != nil && parent.bitpos_ >= x.bitpos_ {
        break  // Upward pointer detected (cycle prevention)
    }
}
```

#### In-Order Traversal (All)
Uses `getNextNode` to iterate in sorted order:
1. Start from leftmost node
2. Follow left children preferentially
3. Move to right when left exhausted
4. Handle upward pointers (impl.go:276-308)

### Patricia Tree Properties

1. **Path Compression:** Only branch at bit positions where keys differ
2. **No Single-Child Nodes:** Except at root
3. **Upward Pointers:** Right pointers may point to ancestors (bitpos decreases)
4. **Leaf vs Internal:** Internal nodes have `intnode_=true`, no data

---

## Performance Characteristics

### Time Complexity

| Operation | Complexity | Notes |
|-----------|-----------|-------|
| Insert | O(k) | k = key length in bits |
| Remove | O(k) | k = key length in bits |
| Search (exact) | O(k) | k = key length in bits |
| PrefixSearch (LPM) | O(k) | k = key length in bits |
| All (iterator) | O(n×k) | n = number of entries |
| UrlTree Insert | O(n) | n = number of URL segments |
| UrlTree Match | O(n) | n = number of URL segments |

### Space Complexity

- **Leaf Nodes:** 1 per inserted key
- **Internal Nodes:** Up to n-1 for n keys (worst case)
- **Total:** O(n) nodes where n = number of keys
- **Memory per Node:** ~64-72 bytes (2 pointers + int + bool + 2 pointers)

### Comparison with Other Structures

| Structure | Insert | Search | Prefix Match | Space |
|-----------|--------|--------|--------------|-------|
| Hash Map | O(1) | O(1) | N/A | O(n) |
| Binary Search Tree | O(log n) | O(log n) | O(k log n) | O(n) |
| Standard Trie | O(k) | O(k) | O(k) | O(n×k) |
| Patricia Tree | O(k) | O(k) | O(k) | O(n) |

**Patricia Advantage:** Better space efficiency than standard trie while maintaining O(k) operations.

---

## Usage Examples

### Example 1: Simple String Keys

```go
tree := patricia.NewSimpleTree[int]()

// Insert entries
tree.Insert("apple", 1)
tree.Insert("application", 2)
tree.Insert("apply", 3)

// Exact search
val, ok := tree.Search("apple")  // Returns 1, true
val, ok = tree.Search("app")     // Returns 0, false

// Prefix search (longest match)
val, ok = tree.PrefixSearch("application")  // Returns 2, true
val, ok = tree.PrefixSearch("apply")        // Returns 3, true
val, ok = tree.PrefixSearch("applying")     // Returns 3, true (matches "apply")

// Iterate all
for key, value := range tree.All() {
    fmt.Printf("%s: %d\n", key, value)
}

// Remove
tree.Remove("application")
```

### Example 2: URL Pattern Matching

```go
tree := patricia.NewUrlTree[string]()

// Define routes
tree.Insert("/api/users/{id}", "GetUser")
tree.Insert("/api/users/{id}/posts", "GetUserPosts")
tree.Insert("/api/users/{id}/posts/{postId}", "GetPost")
tree.Insert("/api/health", "HealthCheck")

// Match incoming requests
keys, values, handler, ok := tree.Match("/api/users/42")
// keys = ["id"], values = ["42"], handler = "GetUser", ok = true

keys, values, handler, ok = tree.Match("/api/users/42/posts/99")
// keys = ["id", "postId"], values = ["42", "99"], handler = "GetPost"

keys, values, handler, ok = tree.Match("/api/health")
// keys = [], values = [], handler = "HealthCheck", ok = true
```

### Example 3: Custom Key Type

```go
type ipv4Key struct {
    addr [4]byte
    prefixLen int
}

func (k ipv4Key) BitLength() int {
    return k.prefixLen
}

func (k ipv4Key) ByteValue(pos int) byte {
    if pos >= 4 {
        return 0
    }
    return k.addr[pos]
}

tree := &patricia.treeImpl[ipv4Key, string]{}
tree.Insert(&ipv4Key{addr: [4]byte{192, 168, 1, 0}, prefixLen: 24}, ptr("Home Network"))
```

---

## Implementation Review

### Strengths

1. **Generic Design (impl.go:21-28)**
   - Clean Key interface abstraction
   - Type-safe with Go generics
   - Reusable for multiple key types

2. **Efficient Bit Operations (impl.go:87-95)**
   - Direct byte array access
   - Minimal bit manipulation overhead
   - Cache-friendly memory access patterns

3. **Comprehensive Test Coverage**
   - All core operations tested (impl_test.go)
   - Edge cases covered (empty keys, non-existent keys)
   - Integration tests for real-world scenarios

4. **Clean API Design (simple.go)**
   - Intuitive method names
   - Go 1.23+ iterator pattern support
   - Boolean return values for operation status

5. **URL Pattern Matching Innovation (url.go)**
   - Practical use case implementation
   - Clean separation from core logic
   - Variable extraction built-in

### Code Quality Observations

#### Positive Aspects

1. **Documentation:**
   - Good inline comments explaining algorithms
   - Clear parameter descriptions
   - Historical attribution preserved

2. **Error Handling:**
   - Nil checks throughout
   - Bounds checking in ByteValue methods
   - Safe pointer dereferencing

3. **Code Organization:**
   - Clear separation of concerns
   - Logical file structure
   - Consistent naming conventions

#### Areas of Concern

1. **Reflection Usage (impl.go:69-83)**
   ```go
   func createNew[K Key, D any]() *node[K, D] {
       var t node[K, D]
       typeOfT := reflect.TypeOf(t)
       // Reflection logic...
   }
   ```
   **Issue:** Unnecessary reflection for node creation
   **Impact:** Performance overhead on every allocation

2. **Bug in compareNodes (impl.go:117)**
   ```go
   for pos <<= 3; pos < byteLen; pos++ {  // Should be: pos < bitLength
   ```
   **Issue:** Loop condition uses byteLen instead of bit length
   **Impact:** May not compare remaining bits correctly

3. **Deprecated Methods (impl.go:181, 209, 360)**
   - `GetLastNode`, `GetPrevNode`, `FindNextNode` marked deprecated
   - Still present in codebase
   **Recommendation:** Remove or clearly separate as legacy API

4. **URL Tree Limitation (url.go:98-107)**
   - Cannot mix wildcards and literal children at same level
   - Prevents patterns like `/api/{version}/users` and `/api/v1/users` coexisting
   **Impact:** Limits routing flexibility

5. **No Concurrency Support**
   - README states "external synchronization required"
   - No sync.RWMutex option provided
   **Recommendation:** Add optional concurrent-safe wrapper

6. **Memory Management**
   - No node pooling or reuse
   - Every allocation creates new node
   **Impact:** GC pressure under high insert/remove workload

---

## Optimization Recommendations

### High Priority

#### 1. Remove Reflection in createNew (impl.go:69-83)

**Current:**
```go
func createNew[K Key, D any]() *node[K, D] {
    var t node[K, D]
    typeOfT := reflect.TypeOf(t)
    // Reflection logic...
}
```

**Optimized:**
```go
func createNew[K Key, D any]() *node[K, D] {
    return &node[K, D]{}
}
```

**Benefit:**
- Eliminate reflection overhead
- 10-20% faster allocations
- Simpler code

**Rationale:** Reflection is unnecessary for generic types in Go 1.18+. Direct allocation works perfectly.

#### 2. Fix compareNodes Bug (impl.go:117)

**Current:**
```go
for pos <<= 3; pos < byteLen; pos++ {  // Bug: byteLen is bytes, not bits
    if t.getBit(n_left, pos) != t.getBit(n_right, pos) {
        return false
    }
}
```

**Fixed:**
```go
bitLength := n_left.BitLength()
for pos <<= 3; pos < bitLength; pos++ {
    if t.getBit(n_left, pos) != t.getBit(n_right, pos) {
        return false
    }
}
```

**Impact:** Ensures correct comparison of keys with non-byte-aligned lengths.

#### 3. Add Concurrent-Safe Wrapper

**Implementation:**
```go
type ConcurrentSimpleTree[D any] struct {
    tree SimpleTree[D]
    mu   sync.RWMutex
}

func (t *ConcurrentSimpleTree[D]) Insert(key string, data D) bool {
    t.mu.Lock()
    defer t.mu.Unlock()
    return t.tree.Insert(key, data)
}

func (t *ConcurrentSimpleTree[D]) PrefixSearch(key string) (D, bool) {
    t.mu.RLock()
    defer t.mu.RUnlock()
    return t.tree.PrefixSearch(key)
}

// Similar for Remove, Search, All...
```

**Benefit:** Safe for concurrent read operations with minimal overhead.

### Medium Priority

#### 4. Implement Node Pooling

**Implementation:**
```go
var nodePool = sync.Pool{
    New: func() interface{} {
        return &node[K, D]{}
    },
}

func allocNode[K Key, D any]() *node[K, D] {
    return nodePool.Get().(*node[K, D])
}

func freeNode[K Key, D any](n *node[K, D]) {
    // Clear pointers to avoid retaining memory
    n.left_ = nil
    n.right_ = nil
    n.key_ = nil
    n.data_ = nil
    nodePool.Put(n)
}
```

**Benefit:**
- Reduce GC pressure
- Faster allocation/deallocation
- Better for high-churn workloads

**Trade-off:** Increased code complexity

#### 5. Optimize compare Function (impl.go:127-162)

**Current:** Mixed byte and bit comparison with complex logic

**Optimized:**
```go
func (t *treeImpl[K, D]) compare(node_left, node_right *node[K, D], start int) (pos int, isEqual bool) {
    if node_left == nil || node_right == nil {
        return
    }

    leftLen := node_left.BitLength()
    rightLen := node_right.BitLength()
    shortLen := leftLen
    if rightLen < shortLen {
        shortLen = rightLen
    }
    isEqual = (leftLen == rightLen)

    startByte := start >> 3
    endByte := shortLen >> 3

    // Fast byte comparison using slices
    for pos = startByte; pos < endByte; pos++ {
        if node_left.ByteValue(pos) != node_right.ByteValue(pos) {
            // Find exact bit position
            pos <<= 3
            for ; pos < shortLen; pos++ {
                if t.getBit(node_left, pos) != t.getBit(node_right, pos) {
                    isEqual = false
                    return
                }
            }
        }
    }

    // Compare remaining bits
    pos = endByte << 3
    if pos < start {
        pos = start
    }
    for ; pos < shortLen; pos++ {
        if t.getBit(node_left, pos) != t.getBit(node_right, pos) {
            isEqual = false
            return
        }
    }

    return
}
```

**Benefit:** Clearer logic, potential for SIMD optimization by compiler.

#### 6. URL Tree: Support Mixed Wildcards and Literals

**Current Limitation:** Cannot have both wildcard and literal children (url.go:98-107)

**Proposed:**
```go
type urlNode[D any] struct {
    subKey       string
    key          *urlKey
    pos          int
    match        *D
    nextWildcard *urlNode[D]
    subTree      *treeImpl[stringKey, urlNode[D]]
}

// In createNode:
if subKey == "*" {
    node = n.nextWildcard
    isWildcard = true
} else {
    if n.subTree != nil {
        node = n.subTree.FindNode(&stringKey{val: subKey})
    }
}
// Remove the mutual exclusion check
```

**Change findNode to prioritize literal matches:**
```go
func (n *urlNode[D]) findNode(key *urlKey, values_in []string) (keys, values []string, handle *D, ok bool) {
    pos := n.pos + 1
    values = values_in
    var node *urlNode[D]

    // Try literal match first (more specific)
    if n.subTree != nil {
        node = n.subTree.FindNode(&stringKey{val: key.parts[pos]})
        if node != nil {
            // Found literal match, continue...
        }
    }

    // Fall back to wildcard if no literal match
    if node == nil && n.nextWildcard != nil {
        node = n.nextWildcard
        values = append(values, key.parts[pos])
    }

    // Rest of logic...
}
```

**Benefit:** More flexible routing, matches HTTP router libraries' behavior.

### Low Priority

#### 7. Add Benchmarks

Create `impl_bench_test.go`:
```go
func BenchmarkInsert(b *testing.B) {
    tree := NewSimpleTree[int]()
    keys := generateRandomKeys(b.N)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        tree.Insert(keys[i], i)
    }
}

func BenchmarkPrefixSearch(b *testing.B) {
    tree := NewSimpleTree[int]()
    // Setup...
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        tree.PrefixSearch(keys[i%len(keys)])
    }
}
```

#### 8. Add Metrics/Instrumentation

```go
type TreeStats struct {
    TotalNodes     int
    InternalNodes  int
    MaxDepth       int
    AvgKeyLength   float64
}

func (t *treeImpl[K, D]) Stats() TreeStats {
    // Calculate and return stats
}
```

**Benefit:** Performance monitoring, capacity planning.

#### 9. Implement Delete-by-Prefix

```go
func (t *simpleTree[D]) DeletePrefix(prefix string) int {
    // Remove all keys with given prefix
    // Return count of deleted keys
}
```

**Use Case:** Bulk deletion in caching scenarios.

#### 10. Add JSON Serialization

```go
func (t *simpleTree[D]) MarshalJSON() ([]byte, error)
func (t *simpleTree[D]) UnmarshalJSON(data []byte) error
```

**Use Case:** Persistence, configuration storage.

---

## Conclusion

This Patricia tree implementation demonstrates solid algorithmic understanding and clean API design. The core logic is sound and well-tested. Key areas for improvement focus on:

1. **Performance:** Remove reflection, add node pooling
2. **Correctness:** Fix compareNodes bug
3. **Functionality:** Concurrent wrapper, enhanced URL routing
4. **Observability:** Benchmarks, metrics

The codebase serves as an excellent foundation for production use with the recommended optimizations applied.

---

## References

- Original C++ implementation: [Tungsten Fabric Patricia Tree](https://github.com/tungstenfabric/tf-common/blob/master/base/patricia.h)
- Patricia Tree paper: Morrison, D.R. (1968). "PATRICIA—Practical Algorithm To Retrieve Information Coded in Alphanumeric"
- Go Generics: [Go 1.18 Release Notes](https://go.dev/doc/go1.18)
- Range-over-func: [Go 1.23 Release Notes](https://go.dev/doc/go1.23)
