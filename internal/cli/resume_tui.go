//go:build !win7

// Interactive /resume and /takeover command handlers plus their completion.
// These are chatTUI methods driving the interactive picker and are excluded
// from the reduced Win7 build (no interactive TUI). The non-interactive resume
// helpers (recentSessions, mostRecentSession, ...) stay in resume.go for both
// builds.

package cli

import (
	"fmt"
	"strconv"
	"strings"

	"reasonix/internal/i18n"
)

// runResumeCommand handles "/resume": with no argument it opens the recent
// session picker; "/resume <n>" loads that
// session into the running controller in place — keeping the current model and
// replaying the transcript into scrollback.
func (m *chatTUI) runResumeCommand(input string) {
	args := tokenizeArgs(input) // args[0] == "/resume"
	if len(args) < 2 {
		m.openResumePicker()
		return
	}
	// Do not run recovery GC between displaying/completing a numeric index and
	// resolving it here. Removing an earlier row would silently retarget the
	// user's already-selected number. Bare /resume performs cleanup before it
	// builds the picker, and startup performs the ordinary background sweep.
	entries := resumeEntries(m.ctrl.SessionDir())
	if len(entries) == 0 {
		m.notice(i18n.M.NoSessionToResume)
		return
	}
	if m.ctrl.Running() {
		m.notice(i18n.M.ResumeBusy)
		return
	}
	idx, err := strconv.Atoi(strings.TrimSpace(args[1]))
	if err != nil || idx < 1 || idx > len(entries) {
		m.notice(fmt.Sprintf(i18n.M.ResumeBadIndexFmt, len(entries)))
		return
	}
	target := entries[idx-1]
	if target.session.Path == m.ctrl.SessionPath() {
		m.notice(i18n.M.ResumeAlreadyActive)
		return
	}
	// Persist the conversation we're leaving so switching back later restores it.
	// Snapshot before moving the lease: the outgoing session must be written
	// while this process still owns it.
	if err := m.ctrl.Snapshot(); err != nil {
		m.notice("resume: snapshot current session: " + err.Error())
		return
	}
	m.followSessionLease()
	if err := m.commitSessionSwitch(target.session.Path); err != nil {
		m.notice("resume: " + sessionLeaseHeldNotice(err))
		if cliSessionTakeoverCandidate(err) {
			m.pendingTakeoverPath = target.session.Path
			m.notice("run /takeover to take this session over from the resident serve")
		}
		return
	}
	m.replayActiveBranch(i18n.M.ResumedTitle)
}

// runTakeoverCommand handles "/takeover": it force-takes the last refused
// resume target (or an explicit index/path argument) from the resident serve
// on this machine, then resumes it.
func (m *chatTUI) runTakeoverCommand(input string) {
	m.echoLocalCommand(input)
	args := tokenizeArgs(input) // args[0] == "/takeover"
	target := strings.TrimSpace(m.pendingTakeoverPath)
	if len(args) >= 2 {
		target = strings.TrimSpace(args[1])
		if idx, err := strconv.Atoi(target); err == nil {
			entries := resumeEntries(m.ctrl.SessionDir())
			if idx < 1 || idx > len(entries) {
				m.notice(fmt.Sprintf(i18n.M.ResumeBadIndexFmt, len(entries)))
				return
			}
			target = entries[idx-1].session.Path
		}
	}
	if target == "" {
		m.notice("takeover: no refused session; run /resume <n> first or pass an index")
		return
	}
	if m.ctrl.Running() {
		m.notice(i18n.M.ResumeBusy)
		return
	}
	_, err := loadResumableSession(target)
	if err != nil {
		m.notice("takeover: " + err.Error())
		return
	}
	if err := m.ctrl.Snapshot(); err != nil {
		m.notice("takeover: snapshot current session: " + err.Error())
		return
	}
	m.followSessionLease()
	binding, bindErr := cliAcquireFreeSession(target, m.leases, m.takeover)
	if bindErr != nil {
		if !cliSessionTakeoverCandidate(bindErr) {
			m.notice("takeover: " + sessionLeaseHeldNotice(bindErr))
			return
		}
		m.notice("taking the session over from the resident serve…")
		binding, err = cliTakeoverHeldSession(target, bindErr, m.leases, m.takeover)
		if err != nil {
			m.notice("takeover: " + err.Error())
			return
		}
	}
	loaded, err := cliPrepareTakeoverCandidate(binding, m.leases)
	if err != nil {
		_ = cliReturnFailedTakeover(binding, m.leases, m.takeover)
		m.notice("takeover: " + err.Error())
		return
	}
	if err := binding.commitPrevious(m.takeover); err != nil {
		_ = cliReturnFailedTakeover(binding, m.leases, m.takeover)
		m.notice("takeover: " + err.Error())
		return
	}
	m.ctrl.Resume(loaded, target)
	if err := bindChatTUIAuthority(m); err != nil {
		m.notice("takeover: " + err.Error())
		return
	}
	m.pendingTakeoverPath = ""
	if m.takeover != nil && binding.grant.MirrorID != "" {
		m.takeover.AttachController(m.ctrl)
		m.takeover.Activate(binding)
	}
	m.replayActiveBranch(i18n.M.ResumedTitle)
	m.notice("session taken over; the remote side is now read-only and can take it back")
}

// resumeArgItems completes the index argument of "/resume <n>": once past the
// command word it lists recent sessions, inserting the 1-based index and
// showing timestamp + turn count + preview as the hint. Indices match
// the picker because both window through recentSessions.
func (m *chatTUI) resumeArgItems(val string) ([]compItem, int, bool) {
	cmdEnd := strings.IndexAny(val, " \t")
	if cmdEnd < 0 || val[:cmdEnd] != "/resume" {
		return nil, 0, false
	}
	from := strings.LastIndexAny(val, " \t") + 1
	if len(strings.Fields(val[:from])) != 1 || m.ctrl == nil {
		return nil, from, true
	}
	cur := val[from:]
	var out []compItem
	for i, entry := range resumeEntries(m.ctrl.SessionDir()) {
		idx := strconv.Itoa(i + 1)
		if cur != "" && !strings.HasPrefix(idx, cur) {
			continue
		}
		hint := fmt.Sprintf("%s · %s", entry.session.ModTime.Local().Format("01-02 15:04"), sessionSummary(entry.session))
		if entry.project != "" {
			hint = fmt.Sprintf("[%s] %s", entry.project, hint)
		}
		out = append(out, compItem{label: idx, insert: idx, hint: hint})
	}
	return out, from, true
}
