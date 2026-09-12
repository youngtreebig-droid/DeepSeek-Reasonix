//go:build !win7

package productdocs

import "github.com/yuin/goldmark/ast"

// headingStartOffset returns the byte offset of a heading in the source. It
// uses ast.Node.Pos, available in the goldmark version the default build pins.
// The Win7 build pins the older goldmark v1.7.8 (the newest release buildable
// under Go 1.20), which predates Pos; see pos_win7.go.
func headingStartOffset(heading *ast.Heading) int {
	return heading.Pos()
}
