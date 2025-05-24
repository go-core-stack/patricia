package patricia

import "reflect"

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
	left_    *node[K, D] // Left child node
	right_   *node[K, D] // Right child node
	intnode_ bool        // True if this is an internal node, false if it is a leaf
	bitpos_  int         // Bit position used for branching at this node
	key_     *K          // Pointer to the key (only set for leaf nodes)
	data_    *D          // Pointer to the data (only set for leaf nodes)
}

// BitLength returns the bit length of the key stored in this node.
// If the key is nil, it returns 0.
func (n *node[K, D]) BitLength() int {
	if n.key_ == nil {
		return 0
	}
	return (*n.key_).BitLength()
}

// ByteValue returns the byte at the given position in the key stored in this node.
// If the key is nil, it returns 0.
func (n *node[K, D]) ByteValue(pos int) byte {
	if n.key_ == nil {
		return 0
	}
	return (*n.key_).ByteValue(pos)
}

// Tree represents a Patricia tree (Practical Algorithm to Retrieve Information Coded in Alphanumeric).
// It is a space-optimized trie data structure where each node with only one child is merged with its child.
// The Patricia tree supports efficient insert, remove, and search operations for variable-length keys.
type Tree[K Key, D any] struct {
	nodes_    int         // Number of leaf nodes (actual data entries)
	intNodes_ int         // Number of internal nodes (used for branching)
	root_     *node[K, D] // Pointer to the root node of the tree
}

// createNew creates a new node of type node[K, D] using reflection.
// This is used to generically allocate nodes for any key/data type.
func createNew[K Key, D any]() *node[K, D] {
	var t node[K, D]
	typeOfT := reflect.TypeOf(t)

	// If the type is not a pointer, create a pointer to it.
	if typeOfT.Kind() != reflect.Ptr {
		ptrToT := reflect.New(typeOfT)
		val := ptrToT.Elem().Interface().(node[K, D])
		return &val
	}

	val := reflect.New(typeOfT.Elem()).Interface().(node[K, D])
	return &val
}

// getBit returns the value of the bit at position pos in the given node's key.
// Returns false if the node is nil or the position is out of bounds.
func (t *Tree[K, D]) getBit(node *node[K, D], pos int) bool {
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
func (t *Tree[K, D]) compareNodes(n_left, n_right *node[K, D]) bool {
	if n_left == nil || n_right == nil {
		return false
	}
	if n_left.BitLength() != n_right.BitLength() {
		return false
	}

	// Compare byte by byte for the full bytes.
	byteLen := n_left.BitLength() >> 3
	pos := 0
	for ; pos < byteLen; pos++ {
		if n_left.ByteValue(pos) != n_right.ByteValue(pos) {
			return false
		}
	}

	// Compare remaining bits if any.
	for pos <<= 3; pos < byteLen; pos++ {
		if t.getBit(n_left, pos) != t.getBit(n_right, pos) {
			return false
		}
	}
	return true
}

// compare compares two nodes' keys starting from a given bit position.
// Returns the position of the first differing bit and whether the keys are equal up to the shortest length.
func (t *Tree[K, D]) compare(node_left, node_right *node[K, D], start int) (pos int, isEqual bool) {
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
func (t *Tree[K, D]) rewireRightMost(p *node[K, D], x *node[K, D]) *node[K, D] {
	if x == nil {
		return nil
	}

	for x.right_ != nil && x.right_.bitpos_ > x.bitpos_ {
		x = x.right_
	}
	pRight := x.right_
	x.right_ = p
	return pRight
}

// GetLastNode returns the last (rightmost) node in the tree.
// Deprecated: will be removed in future versions. Kept for backward compatibility.
func (t *Tree[K, D]) GetLastNode() *node[K, D] {
	if t.root_ == nil {
		return nil
	}

	x := t.root_
	for x != nil {
		if x.right_ != nil {
			if x.right_.bitpos_ < x.bitpos_ {
				if x.left_ == nil {
					return x
				}
				x = x.left_
			} else {
				x = x.right_
			}
		} else {
			if x.left_ == nil {
				return x
			}
			x = x.left_
		}
	}

	return x
}

// GetPrevNode returns the previous node in the tree before node n.
// Deprecated: will be removed in future versions. Kept for backward compatibility.
func (t *Tree[K, D]) GetPrevNode(n *node[K, D]) *node[K, D] {
	if n == nil || t.root_ == nil {
		return nil
	}

	var l *node[K, D]
	p := t.root_
	x := p
	rightTurn := x
	greatestPartial := x

	for x != nil {
		if x.bitpos_ > n.BitLength() {
			x = nil
			break
		} else if x.bitpos_ == n.BitLength() && !x.intnode_ {
			break
		}
		p = x
		if t.getBit(n, x.bitpos_) {
			rightTurn = x
			x = x.right_
		} else {
			if !x.intnode_ {
				greatestPartial = x
			}
			x = x.left_
		}
		if x != nil && p.bitpos_ >= x.bitpos_ {
			x = nil
			break
		}
	}

	if x == nil || !t.compareNodes(n, x) {
		return nil
	}

	if rightTurn != nil && greatestPartial != nil {
		if greatestPartial.bitpos_ > rightTurn.bitpos_ {
			return greatestPartial
		}
	}

	if rightTurn == nil {
		return greatestPartial
	}

	x = rightTurn.left_
	for x != nil {
		l = x.left_
		r := x.right_
		if r != nil && r.bitpos_ > x.bitpos_ {
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
func (t *Tree[K, D]) getNextNode(n *node[K, D]) *node[K, D] {
	if t.root_ == nil {
		return nil
	}

	x := t.root_
	if n != nil {
		x = n
	}
	l := x

	for x != nil {
		if x.bitpos_ < l.bitpos_ {
			l = x
			x = l.right_
		} else {
			l = x
			if l.left_ != nil {
				x = l.left_
			} else {
				x = l.right_
			}
		}
		if x != nil && x.bitpos_ > l.bitpos_ && !x.intnode_ {
			break
		}
	}

	return x
}

// LPMFind performs a longest prefix match search for the given key.
// Returns the data pointer for the node with the longest matching prefix, or nil if no match is found.
// Logic: Traverses the tree, comparing bits of the search key with each node.
// If a node matches up to its bit position, it is a candidate for the longest prefix match.
func (t *Tree[K, D]) LPMFind(key *K) *D {
	if t.root_ == nil {
		return nil
	}

	n := createNew[K, D]()
	n.key_ = key

	var p, x, l *node[K, D]
	x = t.root_
	i := 0

	for x != nil {
		if !x.intnode_ {
			var ok bool
			if i, ok = t.compare(n, x, i); ok {
				return x.data_
			}
			if i == x.bitpos_ {
				l = x
			}
		}
		if x.bitpos_ > n.BitLength() {
			break
		}
		p = x
		if t.getBit(n, x.bitpos_) {
			x = x.right_
		} else {
			x = x.left_
		}
		if x != nil && p.bitpos_ >= x.bitpos_ {
			break
		}
	}

	if l != nil {
		return l.data_
	}

	return nil
}

// FindNextNode is deprecated. Use getNextNode instead.
// Returns the next node in the tree after node n, or nil if there is none.
// Logic: Traverses the tree to find the next node in key order after n.
func (t *Tree[K, D]) FindNextNode(n *node[K, D]) *node[K, D] {
	if n == nil || t.root_ == nil {
		return nil
	}
	p := t.root_
	l := p
	x := p
	i := 0
	for x != nil {
		if !x.intnode_ {
			var ok bool
			if i, ok = t.compare(n, x, i); ok {
				return t.getNextNode(x)
			}
			if x.bitpos_ > n.BitLength() || i != x.bitpos_ {
				break
			}
			l = x
		}
		p = x
		if t.getBit(n, x.bitpos_) {
			x = x.right_
		} else {
			x = x.left_
		}
		if x != nil && p.bitpos_ >= x.bitpos_ {
			break
		}
	}
	if l != nil {
		x = l
		for x != nil && x.bitpos_ <= i {
			l = x
			if t.getBit(n, x.bitpos_) {
				x = x.right_
			} else {
				x = x.left_
			}
			if x != nil && l.bitpos_ >= x.bitpos_ {
				break
			}
		}
		if n.BitLength() != l.bitpos_ {
			if t.getBit(n, l.bitpos_) {
				if x == nil {
					return nil
				}
				if l.bitpos_ > x.bitpos_ {
					for x != nil && l.bitpos_ > x.bitpos_ {
						l = x
						x = x.right_
					}
					l = x
				} else if t.getBit(n, i) {
					for x.right_ != nil && x.bitpos_ < x.right_.bitpos_ {
						x = x.right_
					}
					l = x
					x = x.right_
					for x != nil && l.bitpos_ > x.bitpos_ {
						l = x
						x = x.right_
					}
					l = x
				} else {
					l = x
				}
			} else {
				if x == nil || t.getBit(n, i) {
					x = l.right_
					for x != nil && l.bitpos_ > x.bitpos_ {
						l = x
						x = x.right_
					}
					l = x
				} else {
					l = x
				}
			}
			if x != nil && !x.intnode_ {
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
func (t *Tree[K, D]) FindNode(key *K) *D {
	if t.root_ == nil {
		return nil
	}

	n := createNew[K, D]()
	n.key_ = key

	p := t.root_
	x := p
	for x != nil {
		if x.bitpos_ > n.BitLength() {
			x = nil
			break
		} else if x.bitpos_ == n.BitLength() && !x.intnode_ {
			break
		}
		p = x
		if t.getBit(n, x.bitpos_) {
			x = x.right_
		} else {
			x = x.left_
		}
		if x != nil && p.bitpos_ >= x.bitpos_ {
			break
		}
	}

	if x == nil || !t.compareNodes(n, x) {
		return nil
	}

	return x.data_
}

// Remove deletes the node with the given key from the tree.
// Returns true if the node was found and removed, false otherwise.
// Logic: Traverses the tree to find the node, then restructures the tree to maintain the Patricia property.
func (t *Tree[K, D]) Remove(key *K) bool {
	if t.root_ == nil {
		return false
	}

	n := createNew[K, D]()
	n.key_ = key

	var pPrev, p *node[K, D]
	x := t.root_

	// Traverse the tree to find the node to remove.
	for x != nil {
		if x.bitpos_ > n.BitLength() {
			x = nil
			break
		} else if x.bitpos_ == n.BitLength() && !x.intnode_ {
			break
		}
		pPrev = p
		p = x
		if t.getBit(n, x.bitpos_) {
			x = x.right_
		} else {
			x = x.left_
		}
		if x != nil && p.bitpos_ >= x.bitpos_ {
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
	if x.left_ != nil && x.right_ != nil && x.bitpos_ < x.right_.bitpos_ {
		// Replace the node with a new internal node.
		a = createNew[K, D]()
		a.bitpos_ = x.bitpos_
		a.intnode_ = true
		t.intNodes_++
		a.left_ = x.left_
		a.right_ = x.right_
		t.rewireRightMost(a, x.left_)
		if p == nil {
			t.root_ = a
		} else if t.getBit(x, p.bitpos_) {
			p.right_ = a
		} else {
			p.left_ = a
		}
		// Case 2: Node has only a left child.
	} else if x.left_ != nil {
		if p == nil {
			t.root_ = x.left_
		} else if t.getBit(x, p.bitpos_) {
			p.right_ = x.left_
		} else {
			p.left_ = x.left_
		}
		t.rewireRightMost(x.right_, x.left_)
		// Case 3: Node has only a right child, and right child is a descendant.
	} else if x.right_ != nil && x.bitpos_ < x.right_.bitpos_ {
		if p == nil {
			t.root_ = x.right_
		} else if t.getBit(x, p.bitpos_) {
			p.right_ = x.right_
		} else {
			p.left_ = x.right_
		}
		// Case 4: Node is a leaf or has no children.
	} else {
		if p == nil {
			t.root_ = nil
		} else if p.intnode_ {
			if t.getBit(x, p.bitpos_) {
				a = p.left_
				// RewireRightMost((pPrev.left_ == p) ? pPrev : nil, a)
				t.rewireRightMost(x.right_, a)
			} else {
				a = p.right_
			}
			if pPrev == nil {
				t.root_ = a
			} else if t.getBit(x, pPrev.bitpos_) {
				pPrev.right_ = a
			} else {
				pPrev.left_ = a
			}
			t.intNodes_--
		} else {
			if t.getBit(x, p.bitpos_) {
				p.right_ = x.right_
			} else {
				p.left_ = nil
			}
		}
	}

	t.nodes_--
	return true
}

// Insert adds a new key/data pair to the tree.
// Returns false if the key already exists, true if the insertion was successful.
// Logic: Traverses the tree to find the correct insertion point, then inserts the new node and restructures as needed.
func (t *Tree[K, D]) Insert(key *K, data *D) bool {
	n := createNew[K, D]()
	n.key_ = key
	n.data_ = data
	x := t.root_
	var p *node[K, D]
	// Traverse the tree to find the correct insertion point.
	for x != nil {
		if x.bitpos_ >= n.BitLength() && !x.intnode_ {
			break
		}
		p = x
		if t.getBit(n, x.bitpos_) {
			x = x.right_
		} else {
			x = x.left_
		}
		if x != nil && p.bitpos_ >= x.bitpos_ {
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
		if i != n.BitLength() || i != l.bitpos_ {
			p = nil
			x = t.root_
			for x != nil && x.bitpos_ <= i && x.bitpos_ < n.BitLength() {
				p = x
				if t.getBit(n, x.bitpos_) {
					x = x.right_
				} else {
					x = x.left_
				}
				if x != nil && p.bitpos_ >= x.bitpos_ {
					x = nil
					break
				}
			}
		}
	}
	t.nodes_++
	n.left_ = nil
	n.right_ = nil
	n.bitpos_ = n.BitLength()
	if x != nil {
		if x.bitpos_ == i {
			n.right_ = x.right_
			n.left_ = x.left_
			t.rewireRightMost(n, x.left_)
			t.intNodes_--
			l = n
		} else {
			if i == n.BitLength() {
				if t.getBit(l, i) {
					n.right_ = x
				} else {
					n.left_ = x
					n.right_ = t.rewireRightMost(n, x)
				}
				l = n
			} else {
				l = createNew[K, D]()
				t.intNodes_++
				l.bitpos_ = i
				l.intnode_ = true
				if t.getBit(n, i) {
					l.left_ = x
					l.right_ = n
					n.right_ = t.rewireRightMost(l, x)
				} else {
					l.left_ = n
					l.right_ = x
					n.right_ = l
				}
			}
		}
	} else {
		if p != nil {
			if t.getBit(n, p.bitpos_) {
				n.right_ = p.right_
			} else {
				n.right_ = p
			}
		}
		l = n
	}
	if p != nil {
		if t.getBit(n, p.bitpos_) {
			p.right_ = l
		} else {
			p.left_ = l
		}
	} else {
		t.root_ = l
	}
	return true
}

// All returns a function that iterates over all key/data pairs in the tree in order.
// The provided yield function is called for each key/data pair. If yield returns false, iteration stops.
// Logic: Uses getNextNode to traverse the tree in key order, yielding each leaf node's key and data.
func (t *Tree[K, D]) All() func(func(K, D) bool) {
	return func(yield func(K, D) bool) {
		if t.root_ == nil {
			// If the tree is empty, nothing to yield.
			return
		}
		// Start with the leftmost node (smallest key).
		n := t.getNextNode(nil)
		for n != nil {
			// Only yield nodes that have both key and data set (i.e., leaf nodes).
			if n.data_ != nil && n.key_ != nil {
				if !yield(*n.key_, *n.data_) {
					// If yield returns false, stop iteration early.
					return
				}
			}
			// Move to the next node in order.
			n = t.getNextNode(n)
		}
	}
}
