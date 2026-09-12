//go:build !win7

// The interactive chat REPL is the charm.land/v2 TUI entry point. It is gated
// behind !win7 because the bubbletea/bubbles/lipgloss stack requires Go >= 1.23
// and is excluded from the reduced go1.20.14 Win7 build. The Win7 build wires a
// non-interactive stub instead (see interactive_win7.go).

package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"reasonix/internal/agent"
	"reasonix/internal/boot"
	"reasonix/internal/config"
	"reasonix/internal/control"
	"reasonix/internal/event"
	"reasonix/internal/extension/providerext"
	"reasonix/internal/i18n"
	"reasonix/internal/provider"
	"reasonix/internal/telemetry"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/pflag"
	"golang.org/x/term"
)

// chatREPL is an interactive session: a single persistent agent/session and a
// prompt loop that keeps conversation context across turns. Exit with
// 'exit'/'quit' or Ctrl-D.
func chatREPL(args []string, version string) int {
	args, deprecatedMode, err := consumeDeprecatedModeFlags(args, "profile", "preset")
	if err != nil {
		fmt.Fprintln(os.Stderr, i18n.M.ErrorPrefix, err)
		return 2
	}
	fs := pflag.NewFlagSet("reasonix", pflag.ContinueOnError)
	fs.SetInterspersed(true)
	model := fs.String("model", "", "provider name (default: config default_model)")
	maxSteps := fs.Int("max-steps", 0, "one-off max tool-call rounds (0 = automatic)")
	cont := registerContinueFlag(fs)
	resume := fs.StringP("resume", "r", "", "resume by session ID/query, or open the picker when no value is given")
	fs.Lookup("resume").NoOptDefVal = resumePickerSentinel
	copySession := fs.Bool("copy", false, "with --resume/--continue: duplicate the selected session and continue in the copy (escape hatch when the original is held by another Reasonix process)")
	yolo := fs.Bool("dangerously-skip-permissions", false, "YOLO: auto-approve approval-gated tool calls this session; same runtime mode as Ctrl+Y")
	fs.BoolVar(yolo, "yolo", false, "alias for --dangerously-skip-permissions")
	dir := fs.String("dir", "", "change to this directory first (project root); config, sandbox and file tools resolve from here")
	effort := fs.String("effort", "", "session reasoning effort override")
	permissionMode := fs.String("permission-mode", "ask", "permission mode: manual | ask | auto | acceptEdits | dontAsk | plan | bypassPermissions")
	var additionalDirs []string
	fs.StringArrayVar(&additionalDirs, "add-dir", nil, "allow tool access to an additional directory (repeatable)")
	var allowedToolValues []string
	fs.StringArrayVar(&allowedToolValues, "allowed-tools", nil, "comma or space-separated permission rules to allow")
	fs.StringArrayVar(&allowedToolValues, "allowedTools", nil, "alias for --allowed-tools")
	if code, ok := parseCommandFlags(fs, normalizeOptionalResumeArg(args)); !ok {
		return code
	}
	allowedTools, err := splitAllowedToolRules(allowedToolValues)
	if err != nil {
		fmt.Fprintln(os.Stderr, i18n.M.ErrorPrefix, err)
		return 2
	}
	if err := acceptDeprecatedModeFlag(deprecatedMode); err != nil {
		fmt.Fprintln(os.Stderr, i18n.M.ErrorPrefix, err)
		return 2
	}
	permissions, err := parsePermissionMode(*permissionMode)
	if err != nil {
		fmt.Fprintln(os.Stderr, i18n.M.ErrorPrefix, err)
		return 2
	}
	allowedTools = uniqueStrings(append(allowedTools, permissions.allow...))
	if rc := chdirTo(*dir); rc != 0 {
		return rc
	}
	workspaceRoot, err := workspaceRootForDir(*dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, i18n.M.ErrorPrefix, err)
		return 1
	}
	// Bubble Tea owns the terminal from the resume picker through controller
	// shutdown. Start diagnostics before config/controller work so hangs leave a
	// non-zero log with milestones (#7435, #7507).
	diagnostics := startTUIDiagnostics(config.ReasonixHomeDir())
	defer diagnostics.Close()
	diagnostics.Milestone("config_load_begin")
	cfg, err := config.Load()
	if err == nil {
		configureCLIThemeWithStyle(cfg.UITheme(), cfg.UIThemeStyle())
		cliCursorShape = cfg.UICursorShape()
	}
	diagnostics.Milestone("config_load_done")

	// Decide whether we're starting fresh or resuming. --resume opens an
	// interactive picker; --continue / -c jumps straight into the newest.
	var resumePath string
	resumeValue := strings.TrimSpace(*resume)
	switch strings.ToLower(resumeValue) {
	case "true":
		resumeValue = resumePickerSentinel
	case "false":
		resumeValue = ""
	}
	switch {
	case resumeValue == resumePickerSentinel:
		path, rc := pickSessionToResume()
		if rc != 0 {
			return rc
		}
		resumePath = path
	case resumeValue != "":
		path, err := resolveSessionQuery(resolveCLISessionDir(), resumeValue)
		if err != nil {
			fmt.Fprintln(os.Stderr, i18n.M.ErrorPrefix, err)
			return 1
		}
		resumePath = path
	case *cont:
		sessionDir := resolveCLISessionDir()
		reclaimCLIRecoveryBranches(sessionDir)
		session, ok := mostRecentSession(sessionDir)
		if !ok {
			fmt.Fprintln(os.Stderr, i18n.M.NoSessionToResume)
			return 1
		}
		resumePath = session.Path
	}
	if *copySession && resumePath == "" {
		fmt.Fprintln(os.Stderr, i18n.M.ErrorPrefix, "--copy requires --resume or --continue")
		return 2
	}
	if *copySession {
		copied, err := copyResumableSession(*model, resumePath, cfg)
		if err != nil {
			fmt.Fprintln(os.Stderr, i18n.M.ErrorPrefix, err)
			return 1
		}
		fmt.Printf("continuing in a session copy: %s\n", copied)
		resumePath = copied
	}
	sessionMode := cliTelemetrySessionMode(*cont, resumeValue != "", *copySession)
	reporter := startCLITelemetry(cfg, telemetry.Options{
		Version: version, Interactive: isInteractive(), CLIMode: "tui",
		PermissionMode: *permissionMode, SessionMode: sessionMode,
	})

	// Own the active session file for the TUI's lifetime; in-TUI switches
	// (/resume, /switch, /new, ...) move the lease with the active path.
	// Refusing a held resume target up front is what keeps a desktop window
	// and this chat from silently double-writing one transcript.
	leases := control.NewSessionLeaseKeeper()
	defer leases.Release()
	takeoverManager := newCLITakeoverManager(nil, leases)
	defer func() {
		if err := takeoverManager.Close(); err != nil {
			fmt.Fprintln(os.Stderr, i18n.M.ErrorPrefix, err)
		}
	}()
	var takeoverBinding *cliTakeoverBinding
	var startupResumeSession *agent.Session
	if resumePath != "" {
		startupResumeSession, err = bindAndLoadCLIResume(leases, resumePath, loadResumableSession)
		if errors.Is(err, agent.ErrSessionLeaseHeld) && cliSessionTakeoverCandidate(err) && promptSessionTakeover(err) {
			takeoverBinding, err = cliTakeoverHeldSession(resumePath, err, leases, takeoverManager)
			if err != nil {
				fmt.Fprintln(os.Stderr, i18n.M.ErrorPrefix, err)
				return 1
			}
			startupResumeSession, err = cliPrepareTakeoverCandidate(takeoverBinding, leases)
			if err != nil {
				_ = cliReturnFailedTakeover(takeoverBinding, leases, takeoverManager)
				fmt.Fprintln(os.Stderr, i18n.M.ErrorPrefix, err)
				return 1
			}
		}
		if err != nil {
			if errors.Is(err, agent.ErrSessionLeaseHeld) {
				fmt.Fprintln(os.Stderr, i18n.M.ErrorPrefix, sessionLeaseResumeRefusal(err))
			} else {
				fmt.Fprintln(os.Stderr, i18n.M.ErrorPrefix, err)
			}
			return 1
		}
	}

	ctx := context.Background()
	if err := applyResumeModel(model, resumePath, cfg); err != nil {
		return cliTakeoverFailure(takeoverBinding, leases, takeoverManager, err)
	}

	// Plumb the controller's typed event stream through a channel so each event
	// can become a tea.Msg inside the TUI's update loop. Buffered generously:
	// streaming bursts (tool results, long answers) shouldn't backpressure the
	// agent goroutine.
	eventCh := make(chan event.Event, 1024)

	var sink event.Sink = &eventSink{ch: eventCh}
	sink = withNotifications(sink, cfg)
	sink = reporter.Wrap(sink)
	takeoverManager.SetInner(sink)
	sink = takeoverManager
	var effortOverride *string
	if strings.TrimSpace(*effort) != "" {
		effortOverride = effort
	}
	overrides := cliBuildOverrides{
		Preset:             deprecatedMode,
		Effort:             effortOverride,
		PermissionAllow:    allowedTools,
		AdditionalDirs:     additionalDirs,
		WorkspaceRoot:      workspaceRoot,
		InteractiveHost:    true,
		Stderr:             diagnostics.Writer(),
		OnSessionRecovered: cliSessionRecoveredHandler(leases),
	}
	diagnostics.Milestone("controller_build_begin")
	ctrl, err := setupProfileWithOverrides(ctx, *model, *maxSteps, false, sink, overrides)
	if err != nil && errors.Is(err, boot.ErrUnknownModel) && isInteractive() && config.SourcePath() == "" {
		// True first run whose default model can't resolve: guide setup, then retry.
		// With a config present, fall through to the descriptive error — re-running
		// the wizard would overwrite the user's config (#2856).
		fmt.Fprintln(os.Stderr, i18n.M.ReconfigureOnUnknownModel)
		if rc := interactiveSetup(defaultConfigTarget(), defaultEnvTarget()); rc != 0 {
			_ = cliReturnFailedTakeover(takeoverBinding, leases, takeoverManager)
			return rc
		}
		ctrl, err = setupProfileWithOverrides(ctx, *model, *maxSteps, false, sink, overrides)
	}
	if err != nil {
		_ = cliReturnFailedTakeover(takeoverBinding, leases, takeoverManager)
		fmt.Fprintln(os.Stderr, i18n.M.ErrorPrefix, err)
		return 1
	}
	diagnostics.Milestone("controller_build_done")

	// Decide where this conversation's auto-save lands. A resume reuses the
	// file so closing/reopening keeps appending to the same history; a fresh
	// session lands in a new file stamped with the model name.
	if err := commitResumedSession(takeoverBinding, takeoverManager, ctrl, startupResumeSession, resumePath); err != nil {
		return cliTakeoverFailure(takeoverBinding, leases, takeoverManager, err)
	}
	ctrl.EnsureSessionPath()
	// Fresh sessions take the lease too (defensive: the path is brand new); a
	// resumed path is already held, making this a no-op.
	if err := rebindCLIControllerAuthority(leases, ctrl); err != nil {
		_ = cliReturnFailedTakeover(takeoverBinding, leases, takeoverManager)
		fmt.Fprintln(os.Stderr, i18n.M.ErrorPrefix, control.SessionInUseMessage(err)+"; "+control.SessionLeaseCloseHint)
		return 1
	}
	reclaimCLIRecoveryBranches(ctrl.SessionDir())

	// Surface a missing-key warning inside the TUI banner so the first message
	// failing is at least pre-announced; the user can still enter chat.
	// resolveModelForCLI transparently falls through a keyless default to the
	// next configured provider (issue #6996). Validating the final ref is a
	// no-op for that configured fallback and preserves the warning when every
	// eligible chat provider is still keyless.
	missing := ""
	if cfg, loadErr := config.Load(); loadErr == nil {
		name, _, err := resolveModelForCLI(*model, cfg)
		switch {
		case err != nil:
			missing = err.Error()
		case name != "" && providerext.PluginRefOwner(name) != "":
			// Plugin-namespaced refs hold no config credential; boot's merged
			// resolver already gated them, and there is no key env to warn about.
		case name != "":
			if vErr := cfg.Validate(name); vErr != nil {
				missing = vErr.Error()
			}
		}
	}

	// Initial terminal width — the TUI re-flows on every WindowSizeMsg so
	// this is just a starting estimate before the first resize event lands.
	termW := 80
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		termW = w
	}

	// Route "ask" decisions to the TUI: the controller emits an ApprovalRequest
	// event and blocks until the user answers via ctrl.Approve. Sub-agents (the
	// task tool) keep their headless gate from setup — no UI to prompt through.
	ctrl.EnableInteractiveApproval()
	applyPermissionMode(ctrl, permissions)
	// YOLO: skip ordinary tool approval requests for the session (deny rules and
	// fresh reviews still apply; ask questions and plan approvals still wait).
	if *yolo {
		ctrl.SetAutoApproveTools(true)
	}

	m := newChatTUI(ctrl, missing, eventCh, termW)
	m.diagnostics = diagnostics
	m.updateWatchdogStatusProvider()
	m.planMode = permissions.plan
	m.leases = leases
	m.takeover = takeoverManager
	takeoverManager.AttachController(ctrl)
	if takeoverBinding != nil {
		takeoverManager.Activate(takeoverBinding)
	}
	if cfg != nil {
		m.outputStyle = cfg.Agent.OutputStyle    // shown as the active entry in /output-style
		m.statuslineCmd = cfg.Statusline.Command // custom status-line command, "" = built-in row
		m.showReasoning = cfg.UI.ShowReasoning   // /verbose persistence: start with config default
		m.showTurnUsage = cfg.UI.ShowTurnUsage   // retain usage accounting even when transcript receipts are hidden
		m.cfg = cfg
	}

	// /model support: a pure builder the TUI calls to rebuild on a different
	// model (carrying the conversation). It must NOT touch the running model —
	// runModelSubcommand performs the swap on the live copy. The same stable sink
	// feeds the new controller, so events keep flowing to this TUI.
	m.buildController = func(spec controllerBuildSpec, carry []provider.Message, resumePath string, oldCtrl control.SessionAPI) (*control.Controller, error) {
		effectiveOverrides := overrides
		if spec.EffortOverride != nil {
			effectiveOverrides.Effort = spec.EffortOverride
		}
		// Keep the logical-session private temporary directory across model /
		// profile switches (Issue #7575).
		effectiveOverrides.SessionTemp = sessionTempFromCLIController(oldCtrl)
		c, err := setupQuietProfile(ctx, spec.ModelRef, *maxSteps, false, sink, effectiveOverrides)
		if err != nil {
			return nil, err
		}
		if spec.EffortOverride != nil {
			overrides.Effort = spec.EffortOverride
		}
		// Keep the carried conversation in its existing file so the switch doesn't
		// orphan a duplicate (#2807).
		path := agent.ContinueSessionPath(resumePath, c.SessionDir(), c.Label())
		if err := adoptCarriedHistoryPreservingProfileAndGrants(c, carry, path, oldCtrl); err != nil {
			c.Close()
			return nil, err
		}
		c.EnableInteractiveApproval()
		c.SetPlanMode(spec.PlanMode)
		if spec.ToolApprovalMode != "" {
			c.SetToolApprovalMode(spec.ToolApprovalMode)
		}
		return c, nil
	}
	// /reload support: rebuild the runtime through boot.Rebuild so tools,
	// skills, commands, hooks, MCP servers, and providers are discovered fresh
	// while the boot layer migrates the session (history, approval grants,
	// goal/recovery state, lifecycle). Same construction inputs as
	// buildController so the replacement matches this session's launch wiring;
	// the CLI holds no SharedHost, so each rebuild owns its plugin host.
	m.bindRuntimeRebuilder(*maxSteps, sink, *yolo, overrides, cliProfileBuildOptions)
	if effortOverride != nil {
		m.effortLevel = *effortOverride
	}
	if effortOverride == nil {
		m.refreshEffortStatus()
	}

	if m.nativeScrollback {
		prepareNativeScrollback(os.Stdout, m.bottomRows())
	}

	// Non-Termux terminals use an alt-screen transcript viewport. Termux stays
	// in the normal buffer so native touch scrollback and soft-keyboard focus
	// keep working; finalized transcript lines are emitted via tea.Println.
	diagnostics.Milestone("terminal_takeover_begin")
	p := tea.NewProgram(m)
	takeoverManager.SetYieldCallback(func() { p.Send(tuiShutdownMsg{}) })
	diagnostics.StartWatchdog(p)
	// SSH drop (SIGHUP) or service stop (SIGTERM): persist the conversation
	// before the terminal goes away, then unwind through the normal close path
	// so resume picks up the interrupted session (#3772).
	hangup := make(chan os.Signal, 1)
	signal.Notify(hangup, syscall.SIGHUP, syscall.SIGTERM)
	go func() {
		for range hangup {
			p.Send(tuiShutdownMsg{})
		}
	}()
	final, runErr := p.Run()
	signal.Stop(hangup)
	diagnostics.Milestone("terminal_released")
	if err := takeoverManager.Close(); err != nil {
		fmt.Fprintln(os.Stderr, i18n.M.ErrorPrefix, err)
		if runErr == nil {
			runErr = err
		}
	}
	// Close the active controller plus any retired ones from /model switches.
	// Retired controllers were stashed rather than closed at switch time
	// because Controller.Close() runs SessionEnd hooks and kills plugin
	// subprocesses — operations that corrupt bubbletea's terminal raw mode
	// when executed while the TUI is alive.
	var launchWeb bool
	var launchWebPath, launchWebSessionID, launchWebModelRef string
	if fm, ok := final.(chatTUI); ok {
		reportShutdownFailure(fm.shutdownErr)
		launchWeb = fm.launchWebOnExit
		for _, oc := range fm.oldControllers {
			if c, ok := oc.(*control.Controller); ok {
				reporter.RecordRecovery(c.DrainRecoveryMetrics())
			}
			oc.Close()
		}
		if fm.ctrl != nil {
			launchWebPath = fm.launchWebResumePath
			launchWebSessionID = fm.launchWebSessionID
			launchWebModelRef = fm.launchWebModelRef
			if c, ok := fm.ctrl.(*control.Controller); ok {
				reporter.RecordRecovery(c.DrainRecoveryMetrics())
			}
			fm.ctrl.Close()
		} else {
			reporter.RecordRecovery(ctrl.DrainRecoveryMetrics())
			ctrl.Close()
		}
	} else {
		reporter.RecordRecovery(ctrl.DrainRecoveryMetrics())
		ctrl.Close()
	}
	if runErr != nil {
		fmt.Fprintln(os.Stderr, i18n.M.ErrorPrefix, runErr)
		return 1
	}
	if launchWeb {
		// The Web runtime resumes a materialized TUI transcript or binds the exact
		// reserved identity for a never-used session. Release the TUI lease before
		// rebuilding the controller or the handoff would correctly reject its own
		// session as already in use. The deferred Release remains as a harmless
		// final guard for every other return path.
		leases.Release()
		return runWebCommand(webHandoffArgs(launchWebPath, launchWebSessionID, launchWebModelRef))
	}
	return 0
}
