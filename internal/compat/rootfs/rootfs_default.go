//go:build !win7

// Package rootfs provides a small traversal-safe directory handle used by the
// legacy archive readers. On the default build it is a thin wrapper over
// os.Root (Go 1.24), which enforces confinement in the kernel. The go1.20.14
// Win7 toolchain has no os.Root, so rootfs_win7.go reimplements the same
// surface using absolute-path operations plus per-component symlink rejection.
package rootfs

import "os"

// Root is a handle to a directory that confines all relative operations to that
// directory subtree.
type Root struct {
	root *os.Root
}

// OpenRoot opens the named directory as a confinement root.
func OpenRoot(name string) (*Root, error) {
	r, err := os.OpenRoot(name)
	if err != nil {
		return nil, err
	}
	return &Root{root: r}, nil
}

// Open opens name (relative to the root) for reading, without following any
// symlink that would leave the root.
func (r *Root) Open(name string) (*os.File, error) { return r.root.Open(name) }

// Lstat returns file info for name (relative to the root) without following a
// terminal symlink.
func (r *Root) Lstat(name string) (os.FileInfo, error) { return r.root.Lstat(name) }

// Stat returns file info for name (relative to the root), following symlinks
// that stay within the root.
func (r *Root) Stat(name string) (os.FileInfo, error) { return r.root.Stat(name) }

// OpenRoot opens name (relative to the root) as a nested confinement root.
func (r *Root) OpenRoot(name string) (*Root, error) {
	sub, err := r.root.OpenRoot(name)
	if err != nil {
		return nil, err
	}
	return &Root{root: sub}, nil
}

// MkdirAll creates name and any missing parents (relative to the root).
func (r *Root) MkdirAll(name string, perm os.FileMode) error { return r.root.MkdirAll(name, perm) }

// Rename renames oldName to newName, both relative to the root.
func (r *Root) Rename(oldName, newName string) error { return r.root.Rename(oldName, newName) }

// Chmod changes the mode of name (relative to the root).
func (r *Root) Chmod(name string, mode os.FileMode) error { return r.root.Chmod(name, mode) }

// Close releases the underlying directory handle.
func (r *Root) Close() error { return r.root.Close() }
