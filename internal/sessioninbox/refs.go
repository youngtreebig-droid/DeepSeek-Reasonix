package sessioninbox

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

const maxFrozenDirEntries = 100

func scopedRelativePath(rootPath, path string) (string, error) {
	candidate := filepath.Clean(path)
	if filepath.IsAbs(candidate) {
		rel, err := filepath.Rel(rootPath, candidate)
		if err != nil || !filepath.IsLocal(rel) {
			return "", fmt.Errorf("path outside workspace: %s", path)
		}
		candidate = rel
	}
	if !filepath.IsLocal(candidate) {
		return "", fmt.Errorf("path outside workspace: %s", path)
	}
	return candidate, nil
}

func skipFrozenDirEntry(name string, isDir bool) bool {
	if name == ".DS_Store" || name == "Thumbs.db" {
		return true
	}
	if !isDir {
		return false
	}
	switch name {
	case ".git", ".idea", ".vscode", "build", "dist", "node_modules", "__pycache__":
		return true
	default:
		return false
	}
}

func frozenSnapshot(path string, content []byte, truncated bool) RefSnapshot {
	sum := sha256.Sum256(content)
	path = filepath.ToSlash(path)
	return RefSnapshot{
		Kind:        "frozen",
		Path:        path,
		DisplayPath: path,
		Content:     content,
		ContentSHA:  hex.EncodeToString(sum[:]),
		Truncated:   truncated,
	}
}

// ApplyFrozenRefs keeps legacy version-1 blobs readable. New entries persist
// the exact typed reference block in PromptEnvelope.FrozenRefBlock.
func ApplyFrozenRefs(submit string, bodies map[string]string) string {
	if len(bodies) == 0 {
		return submit
	}
	paths := make([]string, 0, len(bodies))
	for path := range bodies {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	var b strings.Builder
	b.WriteString(submit)
	b.WriteString("\n\n<!-- frozen inbox references (enqueue-time snapshot) -->\n")
	for _, path := range paths {
		fmt.Fprintf(&b, "\n### @%s\n```\n%s\n```\n", path, bodies[path])
	}
	return b.String()
}

// MaterializeRefs validates stored legacy snapshots without reading live paths.
// Clean-git entries from pre-fix blobs are paused until the user refreshes them.
func MaterializeRefs(_ context.Context, _ string, refs []RefSnapshot) (string, map[string]string, error) {
	bodies := make(map[string]string, len(refs))
	for _, ref := range refs {
		if ref.Kind == "clean_git" {
			return fmt.Sprintf("legacy clean-git reference %s requires refresh", ref.Path), bodies, nil
		}
		if ref.ContentSHA != "" {
			sum := sha256.Sum256(ref.Content)
			if hex.EncodeToString(sum[:]) != ref.ContentSHA {
				return fmt.Sprintf("checksum mismatch for frozen ref %s", ref.Path), bodies, nil
			}
		}
		bodies[firstNonEmpty(ref.DisplayPath, ref.Path)] = string(ref.Content)
	}
	return "", bodies, nil
}
