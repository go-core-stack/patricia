// Copyright © 2025 Prabhjot Singh Sethi, All Rights reserved
// Author: Prabhjot Singh Sethi <prabhjot.sethi@gmail.com>

package patricia

import (
	"fmt"
	"testing"
)

func TestUrlKeyFormation(t *testing.T) {
	key := newurlKey("/api/service1/v1/scope/{scopeId}/id/{uid}/name/{name}/")
	fmt.Printf("Key: %s, Parts: %v, Vars: %v\n", key.actVal, key.parts, key.vars)
}

func TestUrlSearch(t *testing.T) {
	tree := urlTree[int]{}
	tree.Insert("/api/service1/v1/scope/{scopeId}/id/{uid}/name/{name}", 1)
	tree.Insert("/api/service2/v1/scope/{scopeId}/id/{uid}", 2)
	tree.Insert("/api/sample-service/v1/scope/{scopeId}/disable", 3)
	tree.Insert("/api/public/v1/download", 4)
	tree.Insert("/api/service1/v1/scope/{scopeId}/id/{uid}/table/{tbl}", 1)

	keys, val, h, found := tree.Match("/api/service1/v1/scope/123/id/456/name/test")
	if !found {
		t.Errorf("Expected to find the URL, but it was not found")
	}
	fmt.Printf("Found URL: %v, Value: %v, Hierarchy: %v\n", keys, val, h)

	keys, val, h, found = tree.Match("/api/service1/v1/scope/123/id/456/table/test")
	if !found {
		t.Errorf("Expected to find the URL, but it was not found")
	}
	fmt.Printf("Found URL: %v, Value: %v, Hierarchy: %v\n", keys, val, h)

	keys, val, h, found = tree.Match("/api/public/v1/download")
	if !found {
		t.Errorf("Expected to find the URL, but it was not found")
	}
	fmt.Printf("Found URL: %v, Value: %v, Hierarchy: %v\n", keys, val, h)
}
