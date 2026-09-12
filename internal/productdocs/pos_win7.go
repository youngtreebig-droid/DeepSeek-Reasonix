//go:build win7

package productdocs

import "github.com/yuin/goldmark/ast"

// headingStartOffset returns the byte offset of a heading in the source for the
// Win7 build, which pins goldmark v1.7.8 (the newest release buildable under
// Go 1.20). That release has no ast.Node.Pos, so the offset is taken from the
// heading's first line segment, which is the same start offset Pos reports for
// a heading. Headings with no line segments report -1, matching how the
// default path treats a missing position.
func headingStartOffset(heading *ast.Heading) int {
	if lines := heading.Lines(); lines != nil && lines.Len() > 0 {
		return lines.At(0).Start
	}
	return -1
}
