//go:build !win7

package sessioninbox

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FreezeRefs snapshots explicit workspace paths without following symlinks.
// Controller-owned enqueue uses its typed resolver; this remains for old/API
// callers that provide an explicit path list.
//
// It uses os.Root (Go 1.24) so every path resolves inside the workspace with
// kernel-enforced traversal safety. The Win7 build (Go 1.20) has no os.Root; it
// uses an equivalent implementation in refs_root_win7.go that keeps the same
// scopedRelativePath validation and symlink rejection.
func FreezeRefs(ctx context.Context, workspace string, paths []string) ([]RefSnapshot, error) {
	_ = ctx
	rootPath, err := filepath.Abs(strings.TrimSpace(workspace))
	if err != nil || strings.TrimSpace(workspace) == "" {
		return nil, fmt.Errorf("path requires workspace root")
	}
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return nil, fmt.Errorf("open workspace root: %w", err)
	}
	defer root.Close()

	out := make([]RefSnapshot, 0, len(paths))
	seen := map[string]struct{}{}
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		snapshot, freezeErr := freezeScopedRef(root, rootPath, path)
		if freezeErr != nil {
			snapshot = RefSnapshot{
				Kind:        "frozen",
				Path:        path,
				DisplayPath: path,
				Content:     fmt.Appendf(nil, "/* ref freeze failed: %v */", freezeErr),
			}
		}
		out = append(out, snapshot)
	}
	return out, nil
}

func freezeScopedRef(root *os.Root, rootPath, path string) (RefSnapshot, error) {
	rel, err := scopedRelativePath(rootPath, path)
	if err != nil {
		return RefSnapshot{}, err
	}
	info, err := root.Lstat(rel)
	if err != nil {
		return RefSnapshot{}, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return RefSnapshot{}, fmt.Errorf("path escapes workspace through symlink: %s", path)
	}
	if info.IsDir() {
		content, err := freezeDirectory(root, rel)
		if err != nil {
			return RefSnapshot{}, err
		}
		return frozenSnapshot(rel, content, false), nil
	}

	file, err := root.Open(rel)
	if err != nil {
		return RefSnapshot{}, err
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil {
		return RefSnapshot{}, err
	}
	if !os.SameFile(info, opened) {
		return RefSnapshot{}, fmt.Errorf("reference changed while opening: %s", path)
	}
	data, err := io.ReadAll(io.LimitReader(file, DefaultMaxItemBytes+1))
	if err != nil {
		return RefSnapshot{}, err
	}
	truncated := len(data) > DefaultMaxItemBytes
	if truncated {
		data = data[:DefaultMaxItemBytes]
	}
	after, err := file.Stat()
	if err != nil || !os.SameFile(opened, after) || after.Size() != opened.Size() || !after.ModTime().Equal(opened.ModTime()) {
		return RefSnapshot{}, fmt.Errorf("reference changed while reading: %s", path)
	}
	return frozenSnapshot(rel, data, truncated), nil
}

func freezeDirectory(root *os.Root, rel string) ([]byte, error) {
	var entries []string
	if err := walkFrozenDirectory(root, rel, rel, &entries); err != nil {
		return nil, err
	}
	sort.Strings(entries)
	if len(entries) > maxFrozenDirEntries {
		entries = append(entries[:maxFrozenDirEntries], "…[truncated; directory has more entries]…")
	}
	return []byte(strings.Join(entries, "\n")), nil
}

func walkFrozenDirectory(root *os.Root, dir, base string, entries *[]string) error {
	if len(*entries) > maxFrozenDirEntries {
		return nil
	}
	before, err := root.Lstat(dir)
	if err != nil {
		return err
	}
	if before.Mode()&os.ModeSymlink != 0 || !before.IsDir() {
		return fmt.Errorf("directory reference changed while opening: %s", dir)
	}
	opened, err := root.Open(dir)
	if err != nil {
		return err
	}
	after, err := opened.Stat()
	if err != nil || !os.SameFile(before, after) {
		opened.Close()
		return fmt.Errorf("directory reference changed while opening: %s", dir)
	}
	children, err := opened.ReadDir(-1)
	opened.Close()
	if err != nil {
		return err
	}
	for _, child := range children {
		path := filepath.Join(dir, child.Name())
		info, err := root.Lstat(path)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 || skipFrozenDirEntry(child.Name(), info.IsDir()) {
			continue
		}
		display, relErr := filepath.Rel(base, path)
		if relErr != nil || !filepath.IsLocal(display) {
			return fmt.Errorf("directory entry escapes reference root")
		}
		display = filepath.ToSlash(display)
		if info.IsDir() {
			*entries = append(*entries, display+"/")
			if err := walkFrozenDirectory(root, path, base, entries); err != nil {
				return err
			}
		} else {
			*entries = append(*entries, display)
		}
		if len(*entries) > maxFrozenDirEntries {
			return nil
		}
	}
	return nil
}
