package agent

import (
	"reasonix/internal/provider"
)

func (s *Session) updateBatchGuardResults(calls []provider.ToolCall, results []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	end := len(s.Messages) - 1
	changed, durableChange := false, false
	_rev1 := calls
	for i := len(_rev1) - 1; i >= 0; i-- {
		call := _rev1[i]
		for j := end; j >= 0; j-- {
			m := s.Messages[j]
			if m.Role != provider.RoleTool || m.ToolCallID != call.ID || m.Name != call.Name {
				continue
			}
			end = j - 1
			if m.Content != results[i] {
				s.Messages[j].Content = results[i]
				changed = true
				durableChange = durableChange || j < len(s.persistedMessages)
			}
			break
		}
	}
	if changed {
		s.version++
	}
	if durableChange {
		s.rewriteVersion++
		s.pendingContentReasons = append(s.pendingContentReasons, "tool_guard_before_sampling")
	}
}
