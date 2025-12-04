// Copyright © 2025 Prabhjot Singh Sethi, All Rights reserved
// Author: Prabhjot Singh Sethi <prabhjot.sethi@gmail.com>
//
// Initial logic was contributed by Prabhjot Singh Sethi in
// 2013 to Contail Systems (Juniper Networks) as part of the
// Linux Foundation open source project Tungsten Fabric, while
// originally it was developed for networks centric usecases
// using C++ as the programming language
//
// This repo repurpose the core logic and is rewritten in golang
// for different extensions and use cases.
// Over the time it is expected to significantly diverge from the
// original implementation, but the core logic remains the same.
// Original source:
// https://github.com/tungstenfabric/tf-common/blob/master/base/patricia.h

package patricia

// Key is an interface that must be implemented by any key type used in the Patricia tree.
// It provides methods to get the bit length of the key and to retrieve a byte at a given position.
type Key interface {
	// BitLength returns the number of bits in the key.
	BitLength() int
	// ByteValue returns the byte value at the specified position in the key.
	ByteValue(int) byte
}

// node represents a node in the Patricia tree.
// Each node may be an internal node (used for branching) or a leaf node (storing actual data).
type node[K Key, D any] struct {
	left    *node[K, D] // Left child node
	right   *node[K, D] // Right child node
	intnode bool        // True if this is an internal node, false if it is a leaf
	bitpos  int         // Bit position used for branching at this node
	key     *K          // Pointer to the key (only set for leaf nodes)
	data    *D          // Pointer to the data (only set for leaf nodes)
}

// BitLength returns the bit length of the key stored in this node.
// If the key is nil, it returns 0.
func (n *node[K, D]) BitLength() int {
	if n.key == nil {
		return 0
	}
	return (*n.key).BitLength()
}

// ByteValue returns the byte at the given position in the key stored in this node.
// If the key is nil, it returns 0.
func (n *node[K, D]) ByteValue(pos int) byte {
	if n.key == nil {
		return 0
	}
	return (*n.key).ByteValue(pos)
}

// treeImpl represents a Patricia tree (Practical Algorithm to Retrieve Information Coded in Alphanumeric).
// It is a space-optimized trie data structure where each node with only one child is merged with its child.
// The Patricia tree supports efficient insert, remove, and search operations for variable-length keys.
type treeImpl[K Key, D any] struct {
	nodes    int         // Number of leaf nodes (actual data entries)
	intNodes int         // Number of internal nodes (used for branching)
	root     *node[K, D] // Pointer to the root node of the tree
}

// createNew creates a new node of type node[K, D].
// This is a simple allocation function that returns a zero-initialized node.
func createNew[K Key, D any]() *node[K, D] {
	return &node[K, D]{}
}

// getBit returns the value of the bit at position pos in the given node's key.
// Returns false if the node is nil or the position is out of bounds.
func (t *treeImpl[K, D]) getBit(node *node[K, D], pos int) bool {
	if node == nil {
		return false
	}
	if pos < 0 || pos >= node.BitLength() {
		return false
	}
	return (node.ByteValue(pos>>3) & (0x80 >> (pos & 7))) != 0
}

// compareNodes compares two nodes for key equality.
// Returns true if the keys are equal in both length and content.
func (t *treeImpl[K, D]) compareNodes(n_left, n_right *node[K, D]) bool {
	if n_left == nil || n_right == nil {
		return false
	}
	if n_left.BitLength() != n_right.BitLength() {
		return false
	}

	bitLength := n_left.BitLength()
	// Compare byte by byte for the full bytes.
	byteLen := bitLength >> 3
	pos := 0
	for ; pos < byteLen; pos++ {
		if n_left.ByteValue(pos) != n_right.ByteValue(pos) {
			return false
		}
	}

	// Compare remaining bits if any.
	for pos <<= 3; pos < bitLength; pos++ {
		if t.getBit(n_left, pos) != t.getBit(n_right, pos) {
			return false
		}
	}
	return true
}

// compare compares two nodes' keys starting from a given bit position.
// Returns the position of the first differing bit and whether the keys are equal up to the shortest length.
func (t *treeImpl[K, D]) compare(node_left, node_right *node[K, D], start int) (pos int, isEqual bool) {
	isEqual = false
	if node_left == nil || node_right == nil {
		return
	}

	shortLen := 0
	if node_left.BitLength() < node_right.BitLength() {
		shortLen = node_left.BitLength()
		isEqual = false
	} else {
		shortLen = node_right.BitLength()
		isEqual = (node_left.BitLength() == node_right.BitLength())
	}

	byteLen := shortLen >> 3

	// Compare bytes first.
	for pos = start >> 3; pos < byteLen; pos++ {
		if node_left.ByteValue(pos) != node_right.ByteValue(pos) {
			break
		}
	}
	pos <<= 3
	if pos < start {
		pos = start
	}
	// Compare remaining bits.
	for ; pos < shortLen; pos++ {
		if t.getBit(node_left, pos) != t.getBit(node_right, pos) {
			isEqual = false
			return
		}
	}
	return
}

// rewireRightMost rewires the rightmost child of x to point to p.
// Used during node removal and tree restructuring.
func (t *treeImpl[K, D]) rewireRightMost(p *node[K, D], x *node[K, D]) *node[K, D] {
	if x == nil {
		return nil
	}

	for x.right != nil && x.right.bitpos > x.bitpos {
		x = x.right
	}
	pRight := x.right
	x.right = p
	return pRight
}

// GetLastNode returns the last (rightmost) node in the tree.
// Deprecated: will be removed in future versions. Kept for backward compatibility.
func (t *treeImpl[K, D]) GetLastNode() *node[K, D] {
	if t.root == nil {
		return nil
	}

	x := t.root
	for x != nil {
		if x.right != nil {
			if x.right.bitpos < x.bitpos {
				if x.left == nil {
					return x
				}
				x = x.left
			} else {
				x = x.right
			}
		} else {
			if x.left == nil {
				return x
			}
			x = x.left
		}
	}

	return x
}

// GetPrevNode returns the previous node in the tree before node n.
// Deprecated: will be removed in future versions. Kept for backward compatibility.
func (t *treeImpl[K, D]) GetPrevNode(n *node[K, D]) *node[K, D] {
	if n == nil || t.root == nil {
		return nil
	}

	var l *node[K, D]
	p := t.root
	x := p
	rightTurn := x
	greatestPartial := x

	for x != nil {
		if x.bitpos > n.BitLength() {
			x = nil
			break
		} else if x.bitpos == n.BitLength() && !x.intnode {
			break
		}
		p = x
		if t.getBit(n, x.bitpos) {
			rightTurn = x
			x = x.right
		} else {
			if !x.intnode {
				greatestPartial = x
			}
			x = x.left
		}
		if x != nil && p.bitpos >= x.bitpos {
			x = nil
			break
		}
	}

	if x == nil || !t.compareNodes(n, x) {
		return nil
	}

	if rightTurn != nil && greatestPartial != nil {
		if greatestPartial.bitpos > rightTurn.bitpos {
			return greatestPartial
		}
	}

	if rightTurn == nil {
		return greatestPartial
	}

	x = rightTurn.left
	for x != nil {
		l = x.left
		r := x.right
		if r != nil && r.bitpos > x.bitpos {
			x = r
		} else if l != nil {
			x = l
		} else {
			return x
		}
	}

	return x
}

// getNextNode returns the next node in the tree after node n.
// If n is nil, it returns the leftmost node (the first node in order).
func (t *treeImpl[K, D]) getNextNode(n *node[K, D]) *node[K, D] {
	if t.root == nil {
		return nil
	}

	x := t.root
	if n != nil || x.intnode {
		if n != nil {
			x = n
		}

		l := x

		for x != nil {
			if x.bitpos < l.bitpos {
				l = x
				x = l.right
			} else {
				l = x
				if l.left != nil {
					x = l.left
				} else {
					x = l.right
				}
			}
			if x != nil && x.bitpos > l.bitpos && !x.intnode {
				break
			}
		}
	}

	return x
}

// LPMFind performs a longest prefix match search for the given key.
// Returns the data pointer for the node with the longest matching prefix, or nil if no match is found.
// Logic: Traverses the tree, comparing bits of the search key with each node.
// If a node matches up to its bit position, it is a candidate for the longest prefix match.
func (t *treeImpl[K, D]) LPMFind(key *K) *D {
	if t.root == nil {
		return nil
	}

	n := createNew[K, D]()
	n.key = key

	var p, x, l *node[K, D]
	x = t.root
	i := 0

	for x != nil {
		if !x.intnode {
			var ok bool
			if i, ok = t.compare(n, x, i); ok {
				return x.data
			}
			if i == x.bitpos {
				l = x
			}
		}
		if x.bitpos > n.BitLength() {
			break
		}
		p = x
		if t.getBit(n, x.bitpos) {
			x = x.right
		} else {
			x = x.left
		}
		if x != nil && p.bitpos >= x.bitpos {
			break
		}
	}

	if l != nil {
		return l.data
	}

	return nil
}

// FindNextNode is deprecated. Use getNextNode instead.
// Returns the next node in the tree after node n, or nil if there is none.
// Logic: Traverses the tree to find the next node in key order after n.
func (t *treeImpl[K, D]) FindNextNode(n *node[K, D]) *node[K, D] {
	if n == nil || t.root == nil {
		return nil
	}
	p := t.root
	l := p
	x := p
	i := 0
	for x != nil {
		if !x.intnode {
			var ok bool
			if i, ok = t.compare(n, x, i); ok {
				return t.getNextNode(x)
			}
			if x.bitpos > n.BitLength() || i != x.bitpos {
				break
			}
			l = x
		}
		p = x
		if t.getBit(n, x.bitpos) {
			x = x.right
		} else {
			x = x.left
		}
		if x != nil && p.bitpos >= x.bitpos {
			break
		}
	}
	if l != nil {
		x = l
		for x != nil && x.bitpos <= i {
			l = x
			if t.getBit(n, x.bitpos) {
				x = x.right
			} else {
				x = x.left
			}
			if x != nil && l.bitpos >= x.bitpos {
				break
			}
		}
		if n.BitLength() != l.bitpos {
			if t.getBit(n, l.bitpos) {
				if x == nil {
					return nil
				}
				if l.bitpos > x.bitpos {
					for x != nil && l.bitpos > x.bitpos {
						l = x
						x = x.right
					}
					l = x
				} else if t.getBit(n, i) {
					for x.right != nil && x.bitpos < x.right.bitpos {
						x = x.right
					}
					l = x
					x = x.right
					for x != nil && l.bitpos > x.bitpos {
						l = x
						x = x.right
					}
					l = x
				} else {
					l = x
				}
			} else {
				if x == nil || t.getBit(n, i) {
					x = l.right
					for x != nil && l.bitpos > x.bitpos {
						l = x
						x = x.right
					}
					l = x
				} else {
					l = x
				}
			}
			if x != nil && !x.intnode {
				return x
			}
		}
		return t.getNextNode(l)
	} else {
		if !t.getBit(n, i) {
			// all elements of the tree are on right
			return x
		}
	}
	return nil
}

// FindNode searches for a node with the exact key and returns its data pointer.
// Returns nil if the key is not found.
// Logic: Traverses the tree, following the bits of the search key, and compares the found node for exact match.
func (t *treeImpl[K, D]) FindNode(key *K) *D {
	if t.root == nil {
		return nil
	}

	n := createNew[K, D]()
	n.key = key

	p := t.root
	x := p
	for x != nil {
		if x.bitpos > n.BitLength() {
			x = nil
			break
		} else if x.bitpos == n.BitLength() && !x.intnode {
			break
		}
		p = x
		if t.getBit(n, x.bitpos) {
			x = x.right
		} else {
			x = x.left
		}
		if x != nil && p.bitpos >= x.bitpos {
			break
		}
	}

	if x == nil || !t.compareNodes(n, x) {
		return nil
	}

	return x.data
}

// Remove deletes the node with the given key from the tree.
// Returns true if the node was found and removed, false otherwise.
// Logic: Traverses the tree to find the node, then restructures the tree to maintain the Patricia property.
func (t *treeImpl[K, D]) Remove(key *K) bool {
	if t.root == nil {
		return false
	}

	n := createNew[K, D]()
	n.key = key

	var pPrev, p *node[K, D]
	x := t.root

	// Traverse the tree to find the node to remove.
	for x != nil {
		if x.bitpos > n.BitLength() {
			x = nil
			break
		} else if x.bitpos == n.BitLength() && !x.intnode {
			break
		}
		pPrev = p
		p = x
		if t.getBit(n, x.bitpos) {
			x = x.right
		} else {
			x = x.left
		}
		if x != nil && p.bitpos >= x.bitpos {
			/* no x to deal with */
			x = nil
			break
		}
	}

	// If the node was not found or does not match, return false.
	if x == nil || !t.compareNodes(n, x) {
		return false
	}

	var a *node[K, D]
	// Case 1: Node has both left and right children, and right child is a descendant.
	if x.left != nil && x.right != nil && x.bitpos < x.right.bitpos {
		// Replace the node with a new internal node.
		a = createNew[K, D]()
		a.bitpos = x.bitpos
		a.intnode = true
		t.intNodes++
		a.left = x.left
		a.right = x.right
		t.rewireRightMost(a, x.left)
		if p == nil {
			t.root = a
		} else if t.getBit(x, p.bitpos) {
			p.right = a
		} else {
			p.left = a
		}
		// Case 2: Node has only a left child.
	} else if x.left != nil {
		if p == nil {
			t.root = x.left
		} else if t.getBit(x, p.bitpos) {
			p.right = x.left
		} else {
			p.left = x.left
		}
		t.rewireRightMost(x.right, x.left)
		// Case 3: Node has only a right child, and right child is a descendant.
	} else if x.right != nil && x.bitpos < x.right.bitpos {
		if p == nil {
			t.root = x.right
		} else if t.getBit(x, p.bitpos) {
			p.right = x.right
		} else {
			p.left = x.right
		}
		// Case 4: Node is a leaf or has no children.
	} else {
		if p == nil {
			t.root = nil
		} else if p.intnode {
			if t.getBit(x, p.bitpos) {
				a = p.left
				// RewireRightMost((pPrev.left == p) ? pPrev : nil, a)
				t.rewireRightMost(x.right, a)
			} else {
				a = p.right
			}
			if pPrev == nil {
				t.root = a
			} else if t.getBit(x, pPrev.bitpos) {
				pPrev.right = a
			} else {
				pPrev.left = a
			}
			t.intNodes--
		} else {
			if t.getBit(x, p.bitpos) {
				p.right = x.right
			} else {
				p.left = nil
			}
		}
	}

	t.nodes--
	return true
}

// Insert adds a new key/data pair to the tree.
// Returns false if the key already exists, true if the insertion was successful.
// Logic: Traverses the tree to find the correct insertion point, then inserts the new node and restructures as needed.
func (t *treeImpl[K, D]) Insert(key *K, data *D) bool {
	n := createNew[K, D]()
	n.key = key
	n.data = data
	x := t.root
	var p *node[K, D]
	// Traverse the tree to find the correct insertion point.
	for x != nil {
		if x.bitpos >= n.BitLength() && !x.intnode {
			break
		}
		p = x
		if t.getBit(n, x.bitpos) {
			x = x.right
		} else {
			x = x.left
		}
		if x != nil && p.bitpos >= x.bitpos {
			x = nil
			break
		}
	}
	i := 0
	l := x
	if x == nil {
		l = p
	}
	if l != nil {
		var ok bool
		if i, ok = t.compare(n, l, 0); ok {
			// Key already exists.
			return false
		}
		if i != n.BitLength() || i != l.bitpos {
			p = nil
			x = t.root
			for x != nil && x.bitpos <= i && x.bitpos < n.BitLength() {
				p = x
				if t.getBit(n, x.bitpos) {
					x = x.right
				} else {
					x = x.left
				}
				if x != nil && p.bitpos >= x.bitpos {
					x = nil
					break
				}
			}
		}
	}
	t.nodes++
	n.left = nil
	n.right = nil
	n.bitpos = n.BitLength()
	if x != nil {
		if x.bitpos == i {
			n.right = x.right
			n.left = x.left
			t.rewireRightMost(n, x.left)
			t.intNodes--
			l = n
		} else {
			if i == n.BitLength() {
				if t.getBit(l, i) {
					n.right = x
				} else {
					n.left = x
					n.right = t.rewireRightMost(n, x)
				}
				l = n
			} else {
				l = createNew[K, D]()
				t.intNodes++
				l.bitpos = i
				l.intnode = true
				if t.getBit(n, i) {
					l.left = x
					l.right = n
					n.right = t.rewireRightMost(l, x)
				} else {
					l.left = n
					l.right = x
					n.right = l
				}
			}
		}
	} else {
		if p != nil {
			if t.getBit(n, p.bitpos) {
				n.right = p.right
			} else {
				n.right = p
			}
		}
		l = n
	}
	if p != nil {
		if t.getBit(n, p.bitpos) {
			p.right = l
		} else {
			p.left = l
		}
	} else {
		t.root = l
	}
	return true
}

// All returns a function that iterates over all key/data pairs in the tree in order.
// The provided yield function is called for each key/data pair. If yield returns false, iteration stops.
// Logic: Uses getNextNode to traverse the tree in key order, yielding each leaf node's key and data.
func (t *treeImpl[K, D]) All() func(func(K, D) bool) {
	return func(yield func(K, D) bool) {
		if t.root == nil {
			// If the tree is empty, nothing to yield.
			return
		}
		// Start with the leftmost node (smallest key).
		n := t.getNextNode(nil)
		for n != nil {
			// Only yield nodes that have both key and data set (i.e., leaf nodes).
			if n.data != nil && n.key != nil {
				if !yield(*n.key, *n.data) {
					// If yield returns false, stop iteration early.
					return
				}
			}
			// Move to the next node in order.
			n = t.getNextNode(n)
		}
	}
}
