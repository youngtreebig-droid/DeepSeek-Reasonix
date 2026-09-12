//go:build go1.21

// Package xslog re-exports a structured logging API that resolves to the
// standard log/slog package on the native toolchain (Go 1.21+) and to
// golang.org/x/exp/slog under the go1.20.14 Win7 toolchain, where log/slog does
// not exist. Only the symbols used by Reasonix are re-exported.
package xslog

import "log/slog"

type (
	// Logger mirrors slog.Logger.
	Logger = slog.Logger
	// Handler mirrors slog.Handler.
	Handler = slog.Handler
	// HandlerOptions mirrors slog.HandlerOptions.
	HandlerOptions = slog.HandlerOptions
	// Level mirrors slog.Level.
	Level = slog.Level
	// Attr mirrors slog.Attr.
	Attr = slog.Attr
)

// Level constants.
const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

// Constructors and package-level helpers.
var (
	New            = slog.New
	NewTextHandler = slog.NewTextHandler
	NewJSONHandler = slog.NewJSONHandler
	Default        = slog.Default
	SetDefault     = slog.SetDefault
	Debug          = slog.Debug
	Info           = slog.Info
	Warn           = slog.Warn
	Error          = slog.Error
	String         = slog.String
	Int            = slog.Int
	Any            = slog.Any
)
