package cli

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"reasonix/internal/agent"
	"reasonix/internal/i18n"
)

const resumeListCap = 10

// recentSessions returns the newest saved sessions under dir. It keeps recovery
// groups intact at the display cap (a single group may make the result slightly
// larger) so the 1-based indices match /resume <n> and its completion without
// orphaning a conflict copy from its parent. A read error yields an empty list.
func recentSessions(dir string) []agent.SessionInfo {
	if dir == "" {
		return nil
	}
	sessions, err := agent.ListSessions(dir)
	if err != nil {
		return nil
	}
	sessions = orderResumeSessions(sessions)
	return capResumeSessionGroups(sessions, resumeListCap)
}

// resumeEntry is one picker row: a session plus, for cross-project rows, the
// project it belongs to. The current directory's sessions keep project empty
// so existing labels are unchanged.
type resumeEntry struct {
	session agent.SessionInfo
	project string
}

const resumeOtherProjectsCap = 5

// resumeEntries lists the current directory's recent sessions, then the
// newest session of other known projects — a user who worked on this machine
// over SSH resumes from any directory, not only the workspace root (#9477).
func resumeEntries(dir string) []resumeEntry {
	base := recentSessions(dir)
	out := make([]resumeEntry, 0, len(base)+resumeOtherProjectsCap)
	for _, s := range base {
		out = append(out, resumeEntry{session: s})
	}
	out = append(out, otherProjectResumeEntries(dir)...)
	return out
}

func otherProjectResumeEntries(excludeDir string) []resumeEntry {
	type target struct {
		path string
		root string
	}
	var targets []target
	for _, t := range defaultSessionCatalogTargets() {
		if t.Scope != "project" || t.Path == "" {
			continue
		}
		targets = append(targets, target{path: t.Path, root: t.WorkspaceRoot})
	}
	exclude := filepath.Clean(excludeDir)
	var out []resumeEntry
	for _, t := range targets {
		if filepath.Clean(t.path) == exclude {
			continue
		}
		sessions, err := agent.ListSessions(t.path)
		if err != nil || len(sessions) == 0 {
			continue
		}
		name := filepath.Base(strings.TrimRight(t.root, string(filepath.Separator)))
		if name == "" || name == "." {
			name = t.root
		}
		out = append(out, resumeEntry{session: sessions[0], project: name})
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].session.ModTime.After(out[j].session.ModTime)
	})
	if len(out) > resumeOtherProjectsCap {
		out = out[:resumeOtherProjectsCap]
	}
	return out
}

// mostRecentSession returns the chronologically newest saved session for
// --continue. Interactive resume surfaces deliberately group recovery families
// and prefer visible leaves, but --continue promises the most recent session and
// must not let that presentation ordering select an older recovery copy.
func mostRecentSession(dir string) (agent.SessionInfo, bool) {
	if dir == "" {
		return agent.SessionInfo{}, false
	}
	sessions, err := agent.ListSessions(dir)
	if err != nil || len(sessions) == 0 {
		return agent.SessionInfo{}, false
	}
	return sessions[0], true
}

func capResumeSessionGroups(sessions []agent.SessionInfo, limit int) []agent.SessionInfo {
	if limit <= 0 || len(sessions) <= limit {
		return sessions
	}
	byID := make(map[string]agent.SessionInfo, len(sessions))
	for _, session := range sessions {
		byID[agent.BranchID(session.Path)] = session
	}
	out := make([]agent.SessionInfo, 0, limit)
	for start := 0; start < len(sessions); {
		key := recoveryResumeGroupKey(sessions[start], byID)
		end := start + 1
		for end < len(sessions) && recoveryResumeGroupKey(sessions[end], byID) == key {
			end++
		}
		if len(out) > 0 && len(out)+(end-start) > limit {
			break
		}
		out = append(out, sessions[start:end]...)
		start = end
		if len(out) >= limit {
			break
		}
	}
	return out
}

// orderResumeSessions keeps conflict-recovery copies next to the session they
// came from. Groups remain newest-first, while the newest visible leaf is first
// within each group so interactive picker and numbered resume surfaces present
// the most likely writable continuation before its ancestors.
func orderResumeSessions(sessions []agent.SessionInfo) []agent.SessionInfo {
	if len(sessions) < 2 {
		return sessions
	}
	byID := make(map[string]agent.SessionInfo, len(sessions))
	for _, session := range sessions {
		byID[agent.BranchID(session.Path)] = session
	}
	type resumeGroup struct {
		items    []agent.SessionInfo
		newest   int
		activity int64
	}
	groups := make(map[string]*resumeGroup, len(sessions))
	order := make([]*resumeGroup, 0, len(sessions))
	for i, session := range sessions {
		key := recoveryResumeGroupKey(session, byID)
		group := groups[key]
		if group == nil {
			group = &resumeGroup{newest: i}
			groups[key] = group
			order = append(order, group)
		}
		group.items = append(group.items, session)
		if stamp := session.ModTime.UnixNano(); stamp > group.activity {
			group.activity = stamp
		}
	}
	sort.SliceStable(order, func(i, j int) bool {
		if order[i].activity == order[j].activity {
			return order[i].newest < order[j].newest
		}
		return order[i].activity > order[j].activity
	})

	out := make([]agent.SessionInfo, 0, len(sessions))
	for _, group := range order {
		children := make(map[string]bool, len(group.items))
		members := make(map[string]bool, len(group.items))
		for _, session := range group.items {
			members[agent.BranchID(session.Path)] = true
		}
		for _, session := range group.items {
			parentID := strings.TrimSpace(session.ParentID)
			if members[parentID] {
				children[parentID] = true
			}
		}
		sort.SliceStable(group.items, func(i, j int) bool {
			iLeaf := !children[agent.BranchID(group.items[i].Path)]
			jLeaf := !children[agent.BranchID(group.items[j].Path)]
			if iLeaf != jLeaf {
				return iLeaf
			}
			return group.items[i].ModTime.After(group.items[j].ModTime)
		})
		out = append(out, group.items...)
	}
	return out
}

func recoveryResumeGroupKey(session agent.SessionInfo, byID map[string]agent.SessionInfo) string {
	id := agent.BranchID(session.Path)
	if !session.Recovered {
		return id
	}
	seen := map[string]bool{id: true}
	current := session
	for {
		parentID := strings.TrimSpace(current.ParentID)
		if parentID == "" {
			return agent.BranchID(current.Path)
		}
		if seen[parentID] {
			return "recovery-cycle:" + parentID
		}
		seen[parentID] = true
		parent, ok := byID[parentID]
		if !ok {
			return "recovery-parent:" + parentID
		}
		if !parent.Recovered {
			return parentID
		}
		current = parent
	}
}


// sessionSummary is the "N turns · display title" line shared by the /resume
// list and its argument completion. Explicit session renames win, then topic
// titles, then the raw preview so the user can identify sessions at a glance.
func sessionSummary(s agent.SessionInfo) string {
	preview := s.CustomTitle
	if preview == "" {
		preview = s.TopicTitle
	}
	if preview == "" {
		preview = s.Preview
	}
	if preview == "" {
		preview = "(no user message yet)"
	}
	return recoverySessionBadge(s) + fmt.Sprintf("%d turns · %s", s.Turns, preview)
}

func recoverySessionBadge(s agent.SessionInfo) string {
	if !s.Recovered {
		return ""
	}
	parent := strings.TrimSpace(s.ParentID)
	if len(parent) > 8 {
		parent = parent[:8]
	}
	if parent == "" {
		parent = "?"
	}
	return fmt.Sprintf(i18n.M.ResumeRecoveryBadgeFmt, parent) + " "
}
