//go:build !win7

package diff

import udiff "github.com/aymanbagabas/go-udiff"

// diffLineEdits computes line-granularity edits between old and new text.
// go-udiff v0.3+ exposes this as Lines; the Win7 build pins the older v0.2.0
// (the newest release still buildable under Go 1.20) and approximates it with
// the rune-granularity Strings edit in diff_edits_win7.go.
func diffLineEdits(oldText, newText string) []udiff.Edit {
	return udiff.Lines(oldText, newText)
}
