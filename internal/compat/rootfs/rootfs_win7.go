//go:build win7

// Package rootfs provides a small traversal-safe directory handle. This Win7
// variant reimplements the os.Root surface used by the legacy archive readers
// without os.Root (absent from the go1.20.14 toolchain). Confinement is
// enforced by rejecting any non-local ("..") path and any path that traverses
// a symlinked component, which matches how os.Root refuses to leave the root.
package rootfs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Root is a handle to a directory that confines all relative operations to that
// directory subtree.
type Root struct {
	base string
}

// OpenRoot opens the named directory as a confinement root.
func OpenRoot(name string) (*Root, error) {
	abs, err := filepath.Abs(name)
	if err != nil {
		return nil, err
	}
	info, err := os.Lstat(abs)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("rootfs: root %q is a symlink", name)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("rootfs: root %q is not a directory", name)
	}
	return &Root{base: abs}, nil
}

// resolve validates name is local and traverses no symlinked component, then
// returns the absolute path under the root.
func (r *Root) resolve(name string) (string, error) {
	clean := filepath.Clean(name)
	if clean == "." {
		return r.base, nil
	}
	if !filepath.IsLocal(clean) {
		return "", fmt.Errorf("rootfs: path %q escapes root", name)
	}
	current := r.base
	for _, part := range strings.Split(filepath.ToSlash(clean), "/") {
		if part == "" || part == "." {
			continue
		}
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if err != nil {
			return "", err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("rootfs: path %q traverses a symlink", name)
		}
	}
	return current, nil
}

// Open opens name (relative to the root) for reading.
func (r *Root) Open(name string) (*os.File, error) {
	abs, err := r.resolve(name)
	if err != nil {
		return nil, err
	}
	return os.Open(abs)
}

// Lstat returns file info for name (relative to the root) without following a
// terminal symlink.
func (r *Root) Lstat(name string) (os.FileInfo, error) {
	// Lstat must still report a terminal symlink rather than rejecting it, so
	// resolve only the parent components and Lstat the leaf directly.
	clean := filepath.Clean(name)
	if clean == "." {
		return os.Lstat(r.base)
	}
	if !filepath.IsLocal(clean) {
		return nil, fmt.Errorf("rootfs: path %q escapes root", name)
	}
	dir := filepath.Dir(clean)
	if dir != "." {
		if _, err := r.resolve(dir); err != nil {
			return nil, err
		}
	}
	return os.Lstat(filepath.Join(r.base, clean))
}

// Stat returns file info for name (relative to the root).
func (r *Root) Stat(name string) (os.FileInfo, error) {
	abs, err := r.resolve(name)
	if err != nil {
		return nil, err
	}
	return os.Stat(abs)
}

// OpenRoot opens name (relative to the root) as a nested confinement root.
func (r *Root) OpenRoot(name string) (*Root, error) {
	abs, err := r.resolve(name)
	if err != nil {
		return nil, err
	}
	return &Root{base: abs}, nil
}

// resolveForWrite validates that name is local and that every EXISTING
// component up to the leaf traverses no symlink, then returns the absolute
// path under the root. Unlike resolve it tolerates missing components (needed
// for MkdirAll/Rename destinations), rejecting only symlinked ones that exist.
func (r *Root) resolveForWrite(name string) (string, error) {
	clean := filepath.Clean(name)
	if clean == "." {
		return r.base, nil
	}
	if !filepath.IsLocal(clean) {
		return "", fmt.Errorf("rootfs: path %q escapes root", name)
	}
	current := r.base
	for _, part := range strings.Split(filepath.ToSlash(clean), "/") {
		if part == "" || part == "." {
			continue
		}
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("rootfs: path %q traverses a symlink", name)
		}
	}
	return current, nil
}

// MkdirAll creates name and any missing parents (relative to the root).
func (r *Root) MkdirAll(name string, perm os.FileMode) error {
	abs, err := r.resolveForWrite(name)
	if err != nil {
		return err
	}
	return os.MkdirAll(abs, perm)
}

// Rename renames oldName to newName, both relative to the root.
func (r *Root) Rename(oldName, newName string) error {
	src, err := r.resolveForWrite(oldName)
	if err != nil {
		return err
	}
	dst, err := r.resolveForWrite(newName)
	if err != nil {
		return err
	}
	return os.Rename(src, dst)
}

// Chmod changes the mode of name (relative to the root).
func (r *Root) Chmod(name string, mode os.FileMode) error {
	abs, err := r.resolveForWrite(name)
	if err != nil {
		return err
	}
	return os.Chmod(abs, mode)
}

// Close is a no-op; this variant holds no file descriptor.
func (r *Root) Close() error { return nil }
