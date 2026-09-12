package tool

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

const maxArgumentViolations = 8

// ArgumentValidator adds call-dependent checks that cannot be represented by
// one provider-visible JSON Schema. JSON Schema validation always runs first.
type ArgumentValidator interface {
	ValidateArguments(json.RawMessage) []ArgumentViolation
}

// CapabilityArgumentContract is the effective inner contract exposed when a
// stable proxy injects target-specific fields such as a skill name.
type CapabilityArgumentContract struct {
	Schema  json.RawMessage `json:"input_schema"`
	Example json.RawMessage `json:"call_example,omitempty"`
}

// CapabilityArgumentProvider lets an inspect action show the actual inner
// contract instead of the proxy's generic arguments object.
type CapabilityArgumentProvider interface {
	CapabilityArguments(capabilityID string) (CapabilityArgumentContract, bool)
}

// ArgumentViolation is a value-free description of one invalid argument. It
// intentionally contains schema expectations, never the supplied value.
type ArgumentViolation struct {
	Path     string `json:"path"`
	Keyword  string `json:"keyword"`
	Expected string `json:"expected"`
}

// ArgumentValidationResult is the host-side result for one concrete target.
// Skipped is only safe for third-party MCP schemas that cannot be compiled;
// built-in schema failures are returned in CompileErr.
type ArgumentValidationResult struct {
	Fingerprint string
	Violations  []ArgumentViolation
	Skipped     bool
	CompileErr  error
}

// argumentSchemaCache stores compiled schema validators keyed by fingerprint.
// The concrete cached type is defined alongside the compiler that populates it
// (see arguments_schema.go); the Win7 build skips schema compilation entirely
// and never touches this cache.
var argumentSchemaCache sync.Map

// NormalizeArguments preserves the historical empty/null-to-object
// compatibility without guessing fields, coercing types, or rewriting values.
func NormalizeArguments(raw json.RawMessage) json.RawMessage {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return json.RawMessage(`{}`)
	}
	return append(json.RawMessage(nil), trimmed...)
}

// ValidateArguments validates args against the concrete tool's real schema and
// then runs its optional conditional validator. Compiled validators are shared
// by schema fingerprint and never resolve filesystem or network references.
func ValidateArguments(target Tool, raw json.RawMessage) ArgumentValidationResult {
	if target == nil {
		return ArgumentValidationResult{CompileErr: fmt.Errorf("argument validation target is nil")}
	}
	result := ValidateJSONSchemaValue(target.Schema(), NormalizeArguments(raw))
	if result.CompileErr != nil {
		if _, thirdParty := target.(MCPMetadata); thirdParty {
			result.Skipped = true
			result.CompileErr = nil
			return result
		}
		return result
	}

	if conditional, ok := target.(ArgumentValidator); ok && len(result.Violations) < maxArgumentViolations {
		remaining := maxArgumentViolations - len(result.Violations)
		extra := conditional.ValidateArguments(NormalizeArguments(raw))
		if len(extra) > remaining {
			extra = extra[:remaining]
		}
		result.Violations = append(result.Violations, extra...)
	}
	return result
}

func schemaFingerprint(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

// SchemaFingerprint returns a deterministic digest for diagnostics and cache
// invalidation without returning the schema itself.
func SchemaFingerprint(raw json.RawMessage) string {
	return schemaFingerprint(bytes.TrimSpace(raw))
}

// InvalidateArgumentSchemas drops compiled validators after a catalog change.
func InvalidateArgumentSchemas(fingerprints []string) {
	for _, fingerprint := range fingerprints {
		if fingerprint != "" {
			argumentSchemaCache.Delete(fingerprint)
		}
	}
}

func jsonPointer(tokens []string) string {
	var b strings.Builder
	for _, token := range tokens {
		b.WriteByte('/')
		b.WriteString(strings.ReplaceAll(strings.ReplaceAll(token, "~", "~0"), "/", "~1"))
	}
	return b.String()
}

func truncateASCII(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	if limit <= 3 {
		return value[:limit]
	}
	return value[:limit-3] + "..."
}
