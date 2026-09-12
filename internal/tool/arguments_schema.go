//go:build !win7

package tool

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/santhosh-tekuri/jsonschema/v6/kind"
)

type compiledArgumentSchema struct {
	schema *jsonschema.Schema
	err    error
}

// ValidateJSONSchemaValue validates an arbitrary JSON value against a schema.
// It is used for third-party MCP outputSchema telemetry as well as argument
// contracts; callers decide whether a compile failure is fatal or advisory.
func ValidateJSONSchemaValue(schemaRaw, raw json.RawMessage) ArgumentValidationResult {
	schemaRaw = bytes.TrimSpace(schemaRaw)
	fingerprint := schemaFingerprint(schemaRaw)
	result := ArgumentValidationResult{Fingerprint: fingerprint}
	compiled := loadCompiledArgumentSchema(fingerprint, schemaRaw)
	if compiled.err != nil {
		result.CompileErr = compiled.err
		return result
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		result.Violations = []ArgumentViolation{{Path: "", Keyword: "json", Expected: "one valid JSON object"}}
		return result
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		result.Violations = []ArgumentViolation{{Path: "", Keyword: "json", Expected: "one valid JSON object"}}
		return result
	}
	if err := compiled.schema.Validate(value); err != nil {
		result.Violations = validationViolations(err)
	}
	return result
}

func loadCompiledArgumentSchema(fingerprint string, raw []byte) compiledArgumentSchema {
	if cached, ok := argumentSchemaCache.Load(fingerprint); ok {
		return cached.(compiledArgumentSchema)
	}
	compiled := compileArgumentSchema(fingerprint, raw)
	actual, _ := argumentSchemaCache.LoadOrStore(fingerprint, compiled)
	return actual.(compiledArgumentSchema)
}

func compileArgumentSchema(fingerprint string, raw []byte) compiledArgumentSchema {
	var doc any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&doc); err != nil {
		return compiledArgumentSchema{err: fmt.Errorf("invalid JSON schema: %w", err)}
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return compiledArgumentSchema{err: fmt.Errorf("invalid JSON schema: multiple values")}
	}
	obj, ok := doc.(map[string]any)
	if !ok {
		return compiledArgumentSchema{err: fmt.Errorf("JSON schema root must be an object")}
	}
	_, explicitDialect := obj["$schema"]
	compile := func(draft *jsonschema.Draft) (*jsonschema.Schema, error) {
		compiler := jsonschema.NewCompiler()
		compiler.UseLoader(nil)
		compiler.DefaultDraft(draft)
		resource := "urn:reasonix:argument-schema:" + fingerprint
		if err := compiler.AddResource(resource, doc); err != nil {
			return nil, err
		}
		return compiler.Compile(resource)
	}
	compiled, err := compile(jsonschema.Draft2020)
	if err != nil && !explicitDialect {
		compiled, err = compile(jsonschema.Draft7)
	}
	if err != nil {
		return compiledArgumentSchema{err: fmt.Errorf("compile JSON schema: %w", err)}
	}
	return compiledArgumentSchema{schema: compiled}
}

func validationViolations(err error) []ArgumentViolation {
	var validationErr *jsonschema.ValidationError
	if !errors.As(err, &validationErr) {
		return []ArgumentViolation{{Path: "", Keyword: "schema", Expected: "arguments satisfying the tool schema"}}
	}
	leaves := make([]*jsonschema.ValidationError, 0, maxArgumentViolations)
	collectValidationLeaves(validationErr, &leaves)
	violations := make([]ArgumentViolation, 0, len(leaves))
	for _, leaf := range leaves {
		keyword := "schema"
		if path := leaf.ErrorKind.KeywordPath(); len(path) > 0 {
			keyword = path[len(path)-1]
		}
		violations = append(violations, ArgumentViolation{
			Path:     jsonPointer(leaf.InstanceLocation),
			Keyword:  keyword,
			Expected: expectedForErrorKind(leaf.ErrorKind),
		})
	}
	if len(violations) == 0 {
		violations = append(violations, ArgumentViolation{Path: "", Keyword: "schema", Expected: "arguments satisfying the tool schema"})
	}
	return violations
}

func collectValidationLeaves(err *jsonschema.ValidationError, out *[]*jsonschema.ValidationError) {
	if err == nil || len(*out) >= maxArgumentViolations {
		return
	}
	if len(err.Causes) == 0 {
		*out = append(*out, err)
		return
	}
	for _, cause := range err.Causes {
		collectValidationLeaves(cause, out)
		if len(*out) >= maxArgumentViolations {
			return
		}
	}
}

func expectedForErrorKind(errorKind jsonschema.ErrorKind) string {
	switch k := errorKind.(type) {
	case *kind.Type:
		return strings.Join(k.Want, " or ")
	case *kind.Required:
		return "required properties: " + strings.Join(k.Missing, ", ")
	case *kind.AdditionalProperties:
		return "only declared properties; remove: " + strings.Join(k.Properties, ", ")
	case *kind.Enum:
		return "one of: " + boundedSchemaValues(k.Want)
	case *kind.Const:
		return "constant: " + boundedSchemaValue(k.Want)
	case *kind.MinProperties:
		return "at least " + strconv.Itoa(k.Want) + " properties"
	case *kind.MaxProperties:
		return "at most " + strconv.Itoa(k.Want) + " properties"
	case *kind.MinItems:
		return "at least " + strconv.Itoa(k.Want) + " items"
	case *kind.MaxItems:
		return "at most " + strconv.Itoa(k.Want) + " items"
	default:
		return "value satisfying " + lastKeyword(errorKind.KeywordPath())
	}
}

func boundedSchemaValues(values []any) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, boundedSchemaValue(value))
		if len(strings.Join(parts, ", ")) >= 512 {
			break
		}
	}
	return truncateASCII(strings.Join(parts, ", "), 512)
}

func boundedSchemaValue(value any) string {
	b, err := json.Marshal(value)
	if err != nil {
		return "declared schema value"
	}
	return truncateASCII(string(b), 256)
}

func lastKeyword(path []string) string {
	if len(path) == 0 {
		return "the schema"
	}
	return path[len(path)-1]
}
