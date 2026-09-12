package cli

import (
	"strings"

	"reasonix/internal/control"
	"reasonix/internal/event"
	"reasonix/internal/recovery"
)

// planApprovalTool is the Tool name the controller puts on the ApprovalRequest it
// emits to gate a plan (mirrors control's constant). The banner, status line, and
// approval handler key on it to render the plan-specific prompt and to keep the
// [plan] tag in sync when the user starts execution or exits without executing.
//
// These approval helpers are plain decision logic with no TUI dependency, so
// they live in an untagged file shared by both the full and Win7 builds (the
// approval-choice code path is non-interactive-safe).
const planApprovalTool = "exit_plan_mode"

func isRecoveryApprovalEvent(a *event.Approval) bool {
	return a != nil && (a.Kind == recovery.ApprovalKindRecovery || a.Recovery != nil)
}

func isRecoveryPlanChangeApproval(a *event.Approval) bool {
	if !isRecoveryApprovalEvent(a) || a.Recovery == nil {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(a.Recovery.ChangeKind)) {
	case string(recovery.ChangeStrategy), string(recovery.ChangeScope):
		return true
	default:
		return false
	}
}

func freshApprovalAllowsSession(toolName string) bool {
	return toolName == control.SandboxEscapeApprovalTool || toolName == control.ManagedConfigWriteApprovalTool
}
