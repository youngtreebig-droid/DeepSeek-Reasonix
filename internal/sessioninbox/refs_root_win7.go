//go:build win7

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
//
// Win7 build note: Go 1.20 has no os.Root (added in Go 1.24), so this variant
// reproduces the same guarantees without it. scopedRelativePath already rejects
// any path that is absolute-outside-root or contains ".." (filepath.IsLocal),
// and every path component is Lstat-checked and rejected if it is a symlink, so
// a reference can neither escape the workspace nor be redirected through a
// symlink. The default build uses the kernel-enforced os.Root in refs_root.go.
func FreezeRefs(ctx context.Context, workspace string, paths []string) ([]RefSnapshot, error) {
	_ = ctx
	rootPath, err := filepath.Abs(strings.TrimSpace(workspace))
	if err != nil || strings.TrimSpace(workspace) == "" {
		return nil, fmt.Errorf("path requires workspace root")
	}

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
		snapshot, freezeErr := freezeScopedRef(rootPath, path)
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

// verifyNoSymlinkComponents rejects rel if any component between rootPath and
// the leaf (inclusive) is a symlink, mirroring os.Root's refusal to traverse
// symlinks. rel must already be a local (no "..") path.
func verifyNoSymlinkComponents(rootPath, rel string) error {
	current := rootPath
	for _, part := range strings.Split(filepath.ToSlash(rel), "/") {
		if part == "" || part == "." {
			continue
		}
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("path escapes workspace through symlink: %s", rel)
		}
	}
	return nil
}

func freezeScopedRef(rootPath, path string) (RefSnapshot, error) {
	rel, err := scopedRelativePath(rootPath, path)
	if err != nil {
		return RefSnapshot{}, err
	}
	if err := verifyNoSymlinkComponents(rootPath, rel); err != nil {
		return RefSnapshot{}, err
	}
	abs := filepath.Join(rootPath, rel)
	info, err := os.Lstat(abs)
	if err != nil {
		return RefSnapshot{}, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return RefSnapshot{}, fmt.Errorf("path escapes workspace through symlink: %s", path)
	}
	if info.IsDir() {
		content, err := freezeDirectory(rootPath, rel)
		if err != nil {
			return RefSnapshot{}, err
		}
		return frozenSnapshot(rel, content, false), nil
	}

	file, err := os.Open(abs)
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

func freezeDirectory(rootPath, rel string) ([]byte, error) {
	var entries []string
	if err := walkFrozenDirectory(rootPath, rel, rel, &entries); err != nil {
		return nil, err
	}
	sort.Strings(entries)
	if len(entries) > maxFrozenDirEntries {
		entries = append(entries[:maxFrozenDirEntries], "…[truncated; directory has more entries]…")
	}
	return []byte(strings.Join(entries, "\n")), nil
}

func walkFrozenDirectory(rootPath, dir, base string, entries *[]string) error {
	if len(*entries) > maxFrozenDirEntries {
		return nil
	}
	absDir := filepath.Join(rootPath, dir)
	before, err := os.Lstat(absDir)
	if err != nil {
		return err
	}
	if before.Mode()&os.ModeSymlink != 0 || !before.IsDir() {
		return fmt.Errorf("directory reference changed while opening: %s", dir)
	}
	opened, err := os.Open(absDir)
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
		info, err := os.Lstat(filepath.Join(rootPath, path))
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
			if err := walkFrozenDirectory(rootPath, path, base, entries); err != nil {
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
