// Copyright © 2025 Prabhjot Singh Sethi, All Rights reserved
// Author: Prabhjot Singh Sethi <prabhjot.sethi@gmail.com>

package patricia

import "strings"

// UrlTree is a Patricia tree variant designed for matching URL patterns with variable path segments.
// This is useful for routing, API gateways, or any scenario where URLs may contain variable parts.
//
// Example patterns supported:
//   /api/service1/v1/scope/{scope}/id/{id}/name/{name}
//   /api/service2/v1/scope/{scope}/id/{id}
//   /api/sample-service/v1/scope/{scope}/disable
//   /api/public/v1/download
//
// In these patterns, {xyz} denotes a variable segment. When matching, any value in that position is accepted,
// and the variable name is reported back along with the matched value. This requires a different matching
// strategy and tree structure compared to a basic Patricia tree.

// urlKey represents a parsed URL pattern, split into parts.
// Variable segments (e.g., {id}) are replaced with "*" in parts, and their names are stored in vars.
type urlKey struct {
	actVal string   // The original URL pattern string
	parts  []string // Parts of the URL split by '/'
	vars   []string // Names of the variable parts in the URL
}

// newurlKey parses a URL pattern into a urlKey, identifying variable segments.
// Variable segments are replaced with "*" in parts, and their names are stored in vars.
func newurlKey(url string) *urlKey {
	parts := strings.Split(url, "/")

	key := &urlKey{
		actVal: url,
	}

	for i, part := range parts {
		if part == "" && i == 0 {
			// Skip leading empty part (from leading slash)
			continue
		}
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			varName := part[1 : len(part)-1] // remove { and }
			key.vars = append(key.vars, varName)
			key.parts = append(key.parts, "*") // wildcard for matching
		} else {
			key.parts = append(key.parts, part)
		}
	}

	return key
}

// urlNode represents a node in the UrlTree.
// Each node corresponds to a segment of the URL path, which may be a literal or a wildcard.
type urlNode[D any] struct {
	subKey       string                           // The segment for this node (literal or "*")
	key          *urlKey                          // The urlKey for this node (only set for terminal nodes)
	pos          int                              // Position in the URL path matched so far; -1 for root
	match        *D                               // Pointer to the value if this node is a terminal match
	nextWildcard *urlNode[D]                      // Child node for wildcard ("*") segment, if any
	subTree      *treeImpl[stringKey, urlNode[D]] // Subtree for literal children
}

// UrlTree is the public interface for the URL-matching Patricia tree.
// Insert adds a pattern with associated data.
// Match matches a URL against the patterns, returning variable names, values, and the associated data.
type UrlTree[D any] interface {
	// Insert adds a URL pattern and its associated data to the tree.
	// Returns true if the pattern was newly inserted, false if it already existed.
	Insert(url string, data D) bool

	// Match attempts to match the given URL against the patterns in the tree.
	// Returns:
	//   - keys:   the variable names in the matched pattern (e.g., ["scope", "id"])
	//   - values: the values matched from the URL (e.g., ["foo", "123"])
	//   - handle: the data associated with the matched pattern
	//   - ok:     true if a match was found, false otherwise
	Match(url string) (keys, values []string, handle D, ok bool)
}

// urlTree is the concrete implementation of UrlTree.
type urlTree[D any] struct {
	root *urlNode[D] // Root node of the tree
}

// createNode recursively creates or finds nodes for each segment of the pattern.
// Returns the terminal node if insertion is successful, or nil if the pattern already exists.
func (n *urlNode[D]) createNode(key *urlKey, data *D) *urlNode[D] {
	pos := n.pos + 1
	subKey := key.parts[pos]
	var node *urlNode[D]
	isWildcard := false
	// Find if node exists already in the tree
	if subKey == "*" {
		if n.subTree != nil {
			// currently we do not support wild card and subtree together
			return nil // Cannot have wildcard and subtree together
		}
		node = n.nextWildcard
		isWildcard = true
	} else {
		if n.nextWildcard != nil {
			// currently we do not support wild card and subtree together
			return nil // Cannot have wildcard and subtree together
		}
		if n.subTree != nil {
			node = n.subTree.FindNode(&stringKey{val: subKey})
		}
	}

	if node == nil {
		// Create a new node for this segment
		node = &urlNode[D]{
			subKey: subKey,
			pos:    n.pos + 1,
			match:  nil,
		}
		if isWildcard {
			n.nextWildcard = node
		} else {
			if n.subTree == nil {
				n.subTree = &treeImpl[stringKey, urlNode[D]]{}
			}
			n.subTree.Insert(&stringKey{val: subKey}, node)
		}
	}

	// If this is the last segment in the pattern, set the match value
	if pos == len(key.parts)-1 {
		if node.match != nil {
			// Pattern already exists
			return nil
		} else {
			node.match = data // Set the match value
			node.key = key    // Store the pattern key
			return node
		}
	}

	// Recurse to the next segment
	return node.createNode(key, data)
}

// findNode recursively matches a URL against the tree, collecting variable values.
// Returns variable names, matched values, pointer to data, and match status.
func (n *urlNode[D]) findNode(key *urlKey, values_in []string) (keys, values []string, handle *D, ok bool) {
	pos := n.pos + 1
	values = values_in
	var node *urlNode[D]
	if n.nextWildcard != nil {
		// If there is a wildcard node, check it first
		node = n.nextWildcard
	} else {
		// Otherwise, look for a literal match in the subtree
		node = n.subTree.FindNode(&stringKey{val: key.parts[pos]})
	}

	if node == nil {
		return
	}

	if node.subKey == "*" {
		// If this is a wildcard node, collect the value for this variable segment
		values = append(values, key.parts[pos])
	}

	// If this is the last segment, check for a match
	if pos == len(key.parts)-1 {
		if node.match != nil {
			return node.key.vars, values, node.match, true
		}
		return
	}

	// Continue matching the next segment
	return node.findNode(key, values)
}

// Insert adds a URL pattern and its associated data to the tree.
// Returns true if the pattern was newly inserted, false if it already existed.
//
// Example:
//
//	tree := patricia.NewUrlTree[int]()
//	ok := tree.Insert("/api/service/v1/id/{id}", 42)
func (t *urlTree[D]) Insert(url string, data D) bool {
	uKey := newurlKey(url)
	if t.root == nil {
		t.root = &urlNode[D]{
			pos:     -1,
			subTree: &treeImpl[stringKey, urlNode[D]]{},
		}
	}

	n := t.root.createNode(uKey, &data)
	return n != nil
}

// Match attempts to match the given URL against the patterns in the tree.
// Returns variable names, matched values, the associated data, and whether a match was found.
//
// Example:
//
//	tree := patricia.NewUrlTree[int]()
//	tree.Insert("/api/service/v1/id/{id}", 42)
//	keys, values, handle, ok := tree.Match("/api/service/v1/id/123")
//	// keys = ["id"], values = ["123"], handle = 42, ok = true
func (t *urlTree[D]) Match(url string) (keys, values []string, handle D, ok bool) {
	var zero D
	if t.root == nil {
		return nil, nil, zero, false
	}
	uKey := newurlKey(url)
	keys, values, hPtr, ok := t.root.findNode(uKey, []string{})
	if hPtr == nil {
		return nil, nil, zero, false
	}
	return keys, values, *hPtr, ok
}

// NewUrlTree creates and returns a new UrlTree instance for URL pattern matching.
//
// Example:
//
//	tree := patricia.NewUrlTree[int]()
func NewUrlTree[D any]() UrlTree[D] {
	return &urlTree[D]{
		root: nil,
	}
}
