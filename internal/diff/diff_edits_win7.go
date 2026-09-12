//go:build win7

package diff

import udiff "github.com/aymanbagabas/go-udiff"

// diffLineEdits computes edits between old and new text for the Win7 build.
//
// The Win7 build pins github.com/aymanbagabas/go-udiff to v0.2.0 (the newest
// release that still compiles under Go 1.20; v0.3.0+ declares go 1.23 and
// imports std slices). That release has no Lines helper, so the Win7 diff uses
// the rune-granularity Strings edit. The unified-diff output is still correct;
// only the internal edit granularity differs, which the unified renderer and
// the +N/-M line tallies normalize away.
func diffLineEdits(oldText, newText string) []udiff.Edit {
	return udiff.Strings(oldText, newText)
}
