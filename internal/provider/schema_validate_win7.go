//go:build win7

package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// Win7 build stub for ValidateToolSchema.
//
// The full implementation (schema_validate.go, `//go:build !win7`) compiles
// the schema with github.com/santhosh-tekuri/jsonschema/v6, which requires
// Go >= 1.21 and cannot be built by the go1.20.14 toolchain used for the
// Windows 7 binary. This stub keeps the cheap structural checks that catch
// the schemas most likely to break a provider request (the root must be a
// JSON object with type "object"), but skips full JSON Schema compilation.
func ValidateToolSchema(raw json.RawMessage) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var doc any
	if err := decoder.Decode(&doc); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return fmt.Errorf("invalid JSON: multiple values")
		}
		return fmt.Errorf("invalid JSON: %w", err)
	}
	obj, ok := doc.(map[string]any)
	if !ok {
		return fmt.Errorf("root must be an object")
	}
	switch typ := obj["type"].(type) {
	case string:
		if typ != "object" {
			return fmt.Errorf("root type must be %q, got %q", "object", typ)
		}
	case nil:
		return fmt.Errorf("root schema must declare type %q", "object")
	default:
		return fmt.Errorf("root type must be %q, got %s", "object", schemaJSONString(typ))
	}
	return nil
}
