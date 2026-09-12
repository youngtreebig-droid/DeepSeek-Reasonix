//go:build !win7

// Interactive session-lease handoff: chatTUI methods that move the single-CLI
// session lease as the user switches sessions in the interactive TUI. Excluded
// from the Win7 build (no interactive TUI). The non-interactive lease helpers
// stay in session_lease.go for both builds.

package cli

import (
	"fmt"

	"reasonix/internal/agent"
	"reasonix/internal/control"
)

// rebindSessionLease moves the chat TUI's session lease to path before the
// controller binds it for writing. A nil keeper (tests, persistence disabled)
// gates nothing. On error the keeper still guards the previous session.
func (m *chatTUI) rebindSessionLease(path string) error {
	if m.leases == nil {
		return nil
	}
	handled := false
	var err error
	if m.takeover != nil {
		handled, err = m.takeover.RebindAway(path)
	}
	if err != nil {
		return err
	}
	if !handled {
		err = m.leases.Rebind(path)
	}
	if err != nil {
		return err
	}
	return bindChatTUIAuthority(m)
}

// commitSessionSwitch acquires the target lease before loading its transcript
// while retaining the source keeper. This is the ordinary counterpart of
// /takeover's targeted transaction and also lets a mirrored CLI leave for a
// free session without dropping its source before the candidate is authorized.
func (m *chatTUI) commitSessionSwitch(path string) error {
	return m.commitSessionSwitchWithLoader(path, loadResumableSession)
}

func (m *chatTUI) commitSessionSwitchWithLoader(path string, load func(string) (*agent.Session, error)) error {
	if m == nil {
		return fmt.Errorf("resume candidate unavailable")
	}
	if validator, ok := m.ctrl.(interface{ ValidateSessionModel(string) error }); ok {
		if err := validator.ValidateSessionModel(path); err != nil {
			return err
		}
	}
	binding, err := cliAcquireFreeSession(path, m.leases, m.takeover)
	if err != nil {
		return err
	}
	loaded, err := load(path)
	if err != nil {
		_ = cliReturnFailedTakeover(binding, m.leases, m.takeover)
		return err
	}
	if m.leases != nil {
		if err := m.leases.BindSessionAuthority(loaded); err != nil {
			_ = cliReturnFailedTakeover(binding, m.leases, m.takeover)
			return err
		}
	}
	if err := binding.commitPrevious(m.takeover); err != nil {
		_ = cliReturnFailedTakeover(binding, m.leases, m.takeover)
		return err
	}
	m.ctrl.Resume(loaded, path)
	return bindChatTUIAuthority(m)
}

// restoreSessionLease re-points the lease at the controller's current session
// after a switch attempt moved it but the switch itself then failed.
// Best-effort: the old lease was released during the rebind, so in the
// (unlikely) case another runtime grabbed it in between this stays silent and
// the next write surfaces the conflict.
func (m *chatTUI) restoreSessionLease() {
	if m.leases == nil {
		return
	}
	_ = m.leases.Rebind(m.ctrl.SessionPath())
	_ = bindChatTUIAuthority(m)
}

// followSessionLease re-points the TUI's session lease at the controller's
// current session file after an operation that rotated it to a fresh path
// (/new, /clear, /branch, fork). A fresh path cannot be held by anyone else,
// so failure is theoretical — but never silent.
func (m *chatTUI) followSessionLease() {
	if m.leases == nil {
		return
	}
	if err := m.leases.Rebind(m.ctrl.SessionPath()); err != nil {
		m.notice(sessionLeaseHeldNotice(err))
		return
	}
	if err := bindChatTUIAuthority(m); err != nil {
		m.notice(fmt.Sprintf("session write authority: %v", err))
	}
}

// cliSessionRecoveredHandler moves the single-session CLI lease during the
// controller's recovery commit. The callback runs before Controller changes its
// session path, closing the unguarded interval that event-driven follow-up
// calls left after ordinary turn-end and mid-turn autosaves.


func bindChatTUIAuthority(m *chatTUI) error {
	if m == nil || m.leases == nil {
		return nil
	}
	c, ok := m.ctrl.(*control.Controller)
	if !ok || c == nil {
		return nil
	}
	return m.leases.BindControllerAuthority(c)
}
