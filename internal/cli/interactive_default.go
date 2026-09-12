//go:build !win7

package cli

// defaultInteractiveSession is the interactive entry for the full build: the
// charm.land/v2 TUI chat REPL. The reduced Win7 build replaces this with a stub
// (see interactive_win7.go) because the TUI stack requires Go >= 1.23.
func defaultInteractiveSession(args []string, version string) int {
	return chatREPL(args, version)
}
