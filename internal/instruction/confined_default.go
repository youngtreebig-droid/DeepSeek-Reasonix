//go:build !win7

package instruction

import (
	"os"
	"path/filepath"
)

// readConfinedDocument reads path only if it resolves inside boundary, using
// os.Root (Go 1.24) for kernel-enforced confinement. The Win7 build (Go 1.20)
// has no os.Root; confined_win7.go reproduces the same confinement with
// EvalSymlinks + filepath.IsLocal checks.
func readConfinedDocument(path, boundary, escapeCode string) (string, os.FileInfo, bool, string) {
	boundary = realDirectory(boundary)
	root, err := os.OpenRoot(boundary)
	if err != nil {
		return "", nil, false, ""
	}
	defer root.Close()

	rel, err := filepath.Rel(boundary, absolutePath(path))
	if err == nil && filepath.IsLocal(rel) {
		if f, openErr := root.Open(rel); openErr == nil {
			body, info, ok := readOpenedDocument(f)
			return body, info, ok, ""
		}
	}

	// Root.Open deliberately rejects absolute symlinks, including ones whose
	// target remains inside the root. Resolve those for compatibility, then open
	// the resolved relative path through the same root handle. The second open
	// remains confined if any component changes after EvalSymlinks.
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
	f, err := root.Open(rel)
	if err != nil {
		return "", nil, false, ""
	}
	body, info, ok := readOpenedDocument(f)
	return body, info, ok, ""
}
