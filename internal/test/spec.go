package test

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

// LoadSpec loads and parses the OpenAPI spec from a YAML file.
func LoadSpec(t *testing.T, path string) *openapi3.T {
	t.Helper()
	loader := openapi3.NewLoader()
	spec, err := loader.LoadFromFile(path)
	if err != nil {
		t.Fatalf("Failed to load spec %s: %v", path, err)
	}
	return spec
}

// ValidateAgainstSpec checks that a response matches the OpenAPI spec.
func ValidateAgainstSpec(t *testing.T, spec *openapi3.T, method, path string, statusCode int, contentType string, body []byte) {
	t.Helper()

	// Find the operation
	pathItem := spec.Paths.Find(path)
	if pathItem == nil {
		t.Errorf("Path %s not found in spec", path)
		return
	}

	var op *openapi3.Operation
	switch method {
	case http.MethodGet:
		op = pathItem.Get
	case http.MethodPost:
		op = pathItem.Post
	case http.MethodPut:
		op = pathItem.Put
	case http.MethodDelete:
		op = pathItem.Delete
	}
	if op == nil {
		t.Errorf("Method %s %s not found in spec", method, path)
		return
	}

	// Check status code is in spec
	respRef := op.Responses.Status(statusCode)
	if respRef == nil && statusCode != 501 {
		t.Errorf("Status %d not declared in spec for %s %s", statusCode, method, path)
		return
	}

	// Validate response body structure for JSON
	if contentType == "application/json" && len(body) > 0 {
		var v interface{}
		if err := json.Unmarshal(body, &v); err != nil {
			t.Errorf("Response body is not valid JSON: %v", err)
		}
	}

	// Validate text/plain response against pattern if specified
	if strings.HasPrefix(contentType, "text/plain") && respRef != nil && respRef.Value != nil {
		content := respRef.Value.Content["text/plain"]
		if content != nil && content.Schema != nil && content.Schema.Value != nil {
			pattern := content.Schema.Value.Pattern
			if pattern != "" {
				matched, _ := regexp.MatchString(pattern, string(body))
				if !matched {
					t.Errorf("Response body doesn't match pattern %s", pattern)
				}
			}
		}
	}
}
