//go:build win7

package instruction

import (
	"os"
	"path/filepath"
)

// readConfinedDocument reads path only if it resolves inside boundary. The
// default build uses os.Root (Go 1.24) for kernel-enforced confinement; the
// go1.20.14 Win7 toolchain has no os.Root, so this variant confines by fully
// resolving symlinks with EvalSymlinks and requiring the resolved path to be a
// local (no "..") path under the boundary. A path that escapes the boundary
// through a symlink or "../" is rejected with escapeCode, matching the default
// behaviour.
func readConfinedDocument(path, boundary, escapeCode string) (string, os.FileInfo, bool, string) {
	boundary = realDirectory(boundary)

	// Fast path: a plain relative-within-boundary path with no symlink games.
	rel, err := filepath.Rel(boundary, absolutePath(path))
	if err == nil && filepath.IsLocal(rel) {
		candidate := filepath.Join(boundary, rel)
		if realPath, evalErr := filepath.EvalSymlinks(candidate); evalErr == nil && pathWithin(realPath, boundary) {
			if f, openErr := os.Open(candidate); openErr == nil {
				body, info, ok := readOpenedDocument(f)
				return body, info, ok, ""
			}
		}
	}

	// Slow path: resolve symlinks and re-check confinement, mirroring the
	// os.Root variant's symlink-compatibility branch.
	realPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", nil, false, ""
	}
	if !pathWithin(realPath, boundary) {
		return "", nil, false, escapeCode
	}
	rel, err = filepath.Rel(boundary, realPath)
	if err != nil || !filepath.IsLocal(rel) {
		return "", nil, false, escapeCode
	}
	f, err := os.Open(realPath)
	if err != nil {
		return "", nil, false, ""
	}
	body, info, ok := readOpenedDocument(f)
	return body, info, ok, ""
}
