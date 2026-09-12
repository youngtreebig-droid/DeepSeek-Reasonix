//go:build win7

package tool

import (
	"bytes"
	"encoding/json"
)

// Win7 build stub for JSON Schema argument validation.
//
// The full validator (arguments_schema.go, `//go:build !win7`) relies on
// github.com/santhosh-tekuri/jsonschema/v6, which requires Go >= 1.21 and
// imports the Go 1.21 std `slices` package. The go1.20.14 toolchain that
// produces Windows 7-runnable binaries cannot compile it, and there is no
// go-1.20-compatible v6 release. Rather than pull the whole tool subsystem
// out of the Win7 build, this stub skips schema validation: arguments are
// normalized and passed through without structural checks. Conditional
// ArgumentValidator hooks (see ValidateArguments in arguments.go) still run.
func ValidateJSONSchemaValue(schemaRaw, raw json.RawMessage) ArgumentValidationResult {
	return ArgumentValidationResult{
		Fingerprint: schemaFingerprint(bytes.TrimSpace(schemaRaw)),
	}
}
