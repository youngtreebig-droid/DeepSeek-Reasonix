//go:build !win7

package builtin

import udiff "github.com/aymanbagabas/go-udiff"

// lineEdit is the minimal view of a diff hunk the evidence preview needs: the
// byte offsets in the old text that a change spans. The full build derives it
// from go-udiff's line-boundary edits; the Win7 build computes an equivalent
// via a self-contained line diff (see evidence_diff_win7.go) so it can pin
// go-udiff to a Go-1.20-compatible release that lacks udiff.Lines.
type lineEdit struct {
	Start int
	End   int
}

// builtinDiffLineEdits returns the byte-offset spans of line-boundary edits
// between oldText and newText.
func builtinDiffLineEdits(oldText, newText string) []lineEdit {
	edits := udiff.Lines(oldText, newText)
	out := make([]lineEdit, 0, len(edits))
	for _, e := range edits {
		out = append(out, lineEdit{Start: e.Start, End: e.End})
	}
	return out
}
