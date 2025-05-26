// Copyright © 2025 Prabhjot Singh Sethi, All Rights reserved
// Author: Prabhjot Singh Sethi <prabhjot.sethi@gmail.com>

package main

import (
	"fmt"

	"github.com/go-core-stack/patricia"
)

func main() {
	// Create a new UrlTree for int values
	tree := patricia.NewUrlTree[int]()

	// Insert URL patterns
	tree.Insert("/api/service/v1/id/{id}", 42)
	tree.Insert("/api/service/v1/id/{id}/name/{name}", 99)

	// Match a URL
	keys, values, handle, ok := tree.Match("/api/service/v1/id/123")
	if ok {
		fmt.Printf("Matched! keys=%v, values=%v, handle=%v\n", keys, values, handle)
		// Output: Matched! keys=[id], values=[123], handle=42
	}

	keys, values, handle, ok = tree.Match("/api/service/v1/id/123/name/alice")
	if ok {
		fmt.Printf("Matched! keys=%v, values=%v, handle=%v\n", keys, values, handle)
		// Output: Matched! keys=[id name], values=[123 alice], handle=99
	}
}
