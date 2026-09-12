module reasonix

// Win7 build modfile (used via `go build -modfile=go.win7.mod -tags win7`).
// The go1.20.14 toolchain is the last Go that produces binaries which run on
// Windows 7 (golang/go#64622); it cannot parse the main go.mod's `go 1.26.0`
// directive, and several dependencies pinned there require Go >= 1.21/1.23.
// This modfile lowers the language version and pins the golang.org/x/* stack
// plus modernc.org/sqlite to their newest go-1.20-compatible releases. The
// TUI (charm.land/*), MCP (modelcontextprotocol/go-sdk) and pi catalog
// (sky-valley/pi) subsystems are excluded from the win7 build via build tags,
// so their non-go-1.20 requirements never reach the compiler.
go 1.20

require (
	charm.land/bubbles/v2 v2.2.1
	charm.land/bubbletea/v2 v2.0.9
	charm.land/lipgloss/v2 v2.0.6
	git.sr.ht/~jackmordaunt/go-toast/v2 v2.0.3
	github.com/BurntSushi/toml v1.6.0
	github.com/alecthomas/chroma/v2 v2.27.0
	github.com/atotto/clipboard v0.1.4
	github.com/aymanbagabas/go-udiff v0.4.1
	github.com/bmatcuk/doublestar/v4 v4.10.0
	github.com/charmbracelet/colorprofile v0.4.3
	github.com/charmbracelet/x/ansi v0.11.8
	github.com/godbus/dbus/v5 v5.2.2
	github.com/gorilla/websocket v1.5.3
	github.com/joho/godotenv v1.5.1
	github.com/kevinburke/ssh_config v1.6.0
	github.com/larksuite/oapi-sdk-go/v3 v3.11.0
	github.com/mattn/go-runewidth v0.0.29
	github.com/modelcontextprotocol/go-sdk v1.7.0
	github.com/pkg/sftp v1.13.11
	github.com/rivo/uniseg v0.4.7
	github.com/sabhiram/go-gitignore v0.0.0-20210923224102-525f6e181f06
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.3
	github.com/sky-valley/pi v0.85.21
	github.com/spf13/pflag v1.0.10
	github.com/tree-sitter/go-tree-sitter v0.25.0
	github.com/tree-sitter/tree-sitter-javascript v0.25.0
	github.com/tree-sitter/tree-sitter-python v0.25.0
	github.com/tree-sitter/tree-sitter-rust v0.24.2
	github.com/tree-sitter/tree-sitter-typescript v0.23.2
	github.com/yuin/goldmark v1.8.6
	github.com/zalando/go-keyring v0.2.8
	go.uber.org/goleak v1.3.0
	golang.org/x/crypto v0.54.0
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d
	golang.org/x/image v0.41.0
	golang.org/x/mod v0.37.0
	golang.org/x/net v0.56.0
	golang.org/x/oauth2 v0.35.0
	golang.org/x/sys v0.47.0
	golang.org/x/term v0.45.0
	golang.org/x/text v0.40.0
	gopkg.in/yaml.v3 v3.0.1
	modernc.org/sqlite v1.33.0
	mvdan.cc/sh/v3 v3.7.0
)

// The golang.org/x/* stack and a few transitive deps release newer versions
// that declare `go 1.21`+ and import the Go 1.21/1.23/1.25 std packages
// (slices, cmp, iter, maps, log/slog, math/rand/v2). MVS would otherwise
// upgrade to those and break the go1.20.14 build. These `replace` directives
// pin the newest release of each that still compiles under Go 1.20.
replace (
	github.com/aymanbagabas/go-udiff => github.com/aymanbagabas/go-udiff v0.2.0
	github.com/mattn/go-runewidth => github.com/mattn/go-runewidth v0.0.25
	github.com/pkg/sftp => github.com/pkg/sftp v1.13.9
	github.com/yuin/goldmark => github.com/yuin/goldmark v1.7.8
	modernc.org/libc => modernc.org/libc v1.41.0
	modernc.org/memory => modernc.org/memory v1.7.2
	modernc.org/sqlite => modernc.org/sqlite v1.29.0
	golang.org/x/crypto => golang.org/x/crypto v0.31.0
	golang.org/x/exp => golang.org/x/exp v0.0.0-20231006140011-7918f672742d
	golang.org/x/image => golang.org/x/image v0.23.0
	golang.org/x/mod => golang.org/x/mod v0.17.0
	golang.org/x/net => golang.org/x/net v0.33.0
	golang.org/x/oauth2 => golang.org/x/oauth2 v0.24.0
	golang.org/x/sync => golang.org/x/sync v0.10.0
	golang.org/x/sys => golang.org/x/sys v0.28.0
	golang.org/x/term => golang.org/x/term v0.27.0
	golang.org/x/text => golang.org/x/text v0.21.0
	golang.org/x/time => golang.org/x/time v0.8.0
	golang.org/x/tools => golang.org/x/tools v0.21.0
	mvdan.cc/sh/v3 => mvdan.cc/sh/v3 v3.7.0
)

require (
	github.com/charmbracelet/ultraviolet v0.0.0-20260811164956-006e29f97886 // indirect
	github.com/charmbracelet/x/term v0.2.2 // indirect
	github.com/charmbracelet/x/windows v0.2.2 // indirect
	github.com/clipperhouse/displaywidth v0.11.0 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/danieljoos/wincred v1.2.3 // indirect
	github.com/dlclark/regexp2/v2 v2.2.1 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/jsonschema-go v0.4.3 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/lucasb-eyer/go-colorful v1.4.1 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/mattn/go-pointer v0.0.1 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/segmentio/asm v1.1.3 // indirect
	github.com/segmentio/encoding v0.5.4 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/time v0.15.0 // indirect
	modernc.org/libc v1.61.0 // indirect
	modernc.org/mathutil v1.6.0 // indirect
	modernc.org/memory v1.8.0 // indirect
)
