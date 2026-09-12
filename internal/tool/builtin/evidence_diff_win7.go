//go:build win7

package builtin

// lineEdit is the minimal view of a diff hunk the evidence preview needs: the
// byte offsets in the old text that a change spans. The Win7 build cannot use
// go-udiff's udiff.Lines (that API is only in releases requiring Go >= 1.21),
// so it computes the same line-boundary spans with a self-contained LCS line
// diff below.
type lineEdit struct {
	Start int
	End   int
}

// builtinDiffLineEdits returns the byte-offset spans (into oldText) of the line
// regions that differ between oldText and newText. The output mirrors
// go-udiff's Lines: edits fall on line boundaries and cover only the changed
// old-text lines, which is all the evidence preview consumes.
func builtinDiffLineEdits(oldText, newText string) []lineEdit {
	oldLines, oldOffsets := splitLinesKeepEnds(oldText)
	newLines, _ := splitLinesKeepEnds(newText)

	// Longest common subsequence over whole lines.
	n, m := len(oldLines), len(newLines)
	lcs := make([][]int, n+1)
	for i := range lcs {
		lcs[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if oldLines[i] == newLines[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else if lcs[i+1][j] >= lcs[i][j+1] {
				lcs[i][j] = lcs[i+1][j]
			} else {
				lcs[i][j] = lcs[i][j+1]
			}
		}
	}

	var edits []lineEdit
	i, j := 0, 0
	for i < n && j < m {
		if oldLines[i] == newLines[j] {
			i++
			j++
			continue
		}
		// Start of a divergent region in the old text.
		startLine := i
		for i < n && j < m && oldLines[i] != newLines[j] {
			if lcs[i+1][j] >= lcs[i][j+1] {
				i++
			} else {
				j++
			}
		}
		if i > startLine {
			edits = append(edits, lineEdit{Start: oldOffsets[startLine], End: oldOffsets[i]})
		} else {
			// Pure insertion at this boundary: represent as a zero-width edit at
			// the boundary offset, matching go-udiff's line-boundary edits.
			edits = append(edits, lineEdit{Start: oldOffsets[startLine], End: oldOffsets[startLine]})
		}
	}
	if i < n {
		edits = append(edits, lineEdit{Start: oldOffsets[i], End: oldOffsets[n]})
	} else if j < m {
		edits = append(edits, lineEdit{Start: oldOffsets[n], End: oldOffsets[n]})
	}
	return edits
}

// splitLinesKeepEnds splits text into lines that retain their trailing newline
// (matching go-udiff's splitLines) and returns the byte offset at which each
// line starts, plus a final sentinel offset equal to len(text).
func splitLinesKeepEnds(text string) ([]string, []int) {
	var lines []string
	offsets := []int{0}
	start := 0
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			lines = append(lines, text[start:i+1])
			start = i + 1
			offsets = append(offsets, start)
		}
	}
	if start < len(text) {
		lines = append(lines, text[start:])
		offsets = append(offsets, len(text))
	}
	return lines, offsets
}
