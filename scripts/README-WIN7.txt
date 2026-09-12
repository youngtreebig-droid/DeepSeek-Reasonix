Reasonix for Windows 7 (amd64)
==============================

This is a REDUCED build of the Reasonix CLI intended for Windows 7 (64-bit).

Why a special build?
--------------------
It is compiled with the Go 1.20 toolchain (go1.20.14), which is the last Go
release that produces binaries that RUN on Windows 7. Binaries built with
Go 1.21.5 and later crash on launch (0xc0000005) on Windows 7
(see golang/go issue #64622). The full-featured Reasonix build targets
Windows 10 and later and is built with the current Go toolchain.

What WORKS in this build
------------------------
The non-interactive command-line surface of Reasonix:
  - reasonix --version / reasonix version --verbose / --json
  - one-shot / scripted agent runs and the non-interactive REPL
  - configuration, provider setup, and the built-in tools
  - the built-in providers (anthropic, openai, responses)
  - sqlite-backed session storage

What is EXCLUDED from this build
--------------------------------
To stay compatible with Go 1.20 / Windows 7, the following subsystems are
compiled out of this build:
  - The interactive chat TUI (the full-screen Bubble Tea / Lip Gloss
    interface and syntax-highlighted rendering). A non-interactive REPL /
    plain output is used instead.
  - MCP (Model Context Protocol) servers and their plugin wiring. Commands
    that require MCP report a clear "not available in this build" message.
  - The "pi" model catalog (sky-valley/pi). A static/empty catalog is used.

For the interactive TUI, MCP servers, and the pi model catalog, use the
full-featured Reasonix build on Windows 10 or later.

Install
-------
This is a portable, single-file executable. No installer is required.
  1. Extract reasonix-win7-amd64.exe from this zip.
  2. (Optional) copy it somewhere on your PATH, e.g. C:\Program Files\Reasonix\.
  3. Run it from a Command Prompt or PowerShell:
       reasonix-win7-amd64.exe --version

Notes
-----
  - 32-bit (386) Windows 7 is not provided: the sqlite backend
    (modernc.org/sqlite) only ships 64-bit architecture support in the
    pinned version, so a 386 build does not link.
  - This build could not be launched-tested in the build sandbox (Linux, no
    Windows 7 VM). Compatibility rests on the Go 1.20 toolchain choice plus a
    successful windows/amd64 compile.
