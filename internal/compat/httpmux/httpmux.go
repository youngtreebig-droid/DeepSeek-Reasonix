// Package httpmux is a minimal router that backports the subset of the Go 1.22
// enhanced http.ServeMux that Reasonix's serve layer relies on: method-scoped
// patterns ("GET /path"), single-segment wildcards ("/inbox/items/{id}"),
// subtree patterns ending in "/", and per-request path values retrieved via
// PathValue. The go1.20.14 toolchain used for the Windows 7 build has none of
// this in net/http, so serve.go uses this router on every build to keep one
// code path. Behavior matches the standard mux closely enough for the fixed,
// known route table in internal/serve; it is not a general drop-in.
//
// Intentional divergence from net/http.ServeMux (Go 1.22): this router does NOT
// perform request-path cleaning or the sanitizing 301 redirects the standard
// mux issues. The standard mux, on a request to an unclean path, responds with
// a 301 to the cleaned path (collapsing "//", resolving "." and ".."). Here,
// splitPath only trims the leading/trailing "/" and splits on "/"; it does not
// resolve "." / ".." and it does not collapse repeated slashes. An empty
// segment produced by a doubled slash is therefore kept as a distinct (empty)
// segment, so an unclean path like "/inbox//items" does NOT match the route
// registered for "/inbox/items" (it splits to ["inbox","","items"], three
// segments) and, absent a matching subtree, yields 404 rather than the
// standard mux's 301 redirect to the cleaned path. This is safe and intended
// for internal/serve: every route is a fixed, first-party API or UI path
// invoked by Reasonix's own clients (CLI, desktop shell, the bundled web UI),
// none of which construct unclean paths or depend on the redirect. If this
// router is ever reused for a surface where the redirect matters (e.g.
// canonical-URL SEO, or clients that special-case 301), path cleaning must be
// added.
package httpmux

import (
	"context"
	"net/http"
	"strings"
)

type ctxKey struct{}

// route is one registered pattern.
type route struct {
	method   string   // "" means any method
	segments []string // path split on "/"; a segment "{name}" is a wildcard
	wildcard bool     // pattern ended in "/" (subtree match)
	handler  http.HandlerFunc
}

// Mux is a method+wildcard aware HTTP request multiplexer.
type Mux struct {
	routes []route
}

// New returns an empty Mux.
func New() *Mux { return &Mux{} }

// HandleFunc registers handler for pattern. pattern is "[METHOD ]/path", where
// a path segment written as "{name}" matches any single segment and a pattern
// ending in "/" matches that subtree. This mirrors the Go 1.22 ServeMux syntax
// used by internal/serve.
func (m *Mux) HandleFunc(pattern string, handler http.HandlerFunc) {
	method := ""
	path := pattern
	if i := strings.IndexByte(pattern, ' '); i >= 0 {
		method = pattern[:i]
		path = strings.TrimSpace(pattern[i+1:])
	}
	r := route{
		method:   method,
		handler:  handler,
		wildcard: strings.HasSuffix(path, "/"),
	}
	r.segments = splitPath(path)
	m.routes = append(m.routes, r)
}

// Handle registers a handler (http.Handler) for pattern.
func (m *Mux) Handle(pattern string, handler http.Handler) {
	m.HandleFunc(pattern, handler.ServeHTTP)
}

// splitPath splits a URL path into segments, dropping the leading empty element
// and any trailing empty element (so "/a/b" -> ["a","b"], "/" -> []).
func splitPath(p string) []string {
	p = strings.Trim(p, "/")
	if p == "" {
		return nil
	}
	return strings.Split(p, "/")
}

// match reports whether r matches the request path segments and, if so, returns
// the captured wildcard values.
func (r route) match(reqSegs []string) (map[string]string, bool) {
	if r.wildcard {
		// Subtree match: the request must be at least as long as the pattern
		// prefix, and every non-wildcard prefix segment must be equal.
		if len(reqSegs) < len(r.segments) {
			return nil, false
		}
		vals := map[string]string{}
		for i, seg := range r.segments {
			if name, ok := wildcardName(seg); ok {
				vals[name] = reqSegs[i]
				continue
			}
			if reqSegs[i] != seg {
				return nil, false
			}
		}
		return vals, true
	}
	if len(reqSegs) != len(r.segments) {
		return nil, false
	}
	vals := map[string]string{}
	for i, seg := range r.segments {
		if name, ok := wildcardName(seg); ok {
			vals[name] = reqSegs[i]
			continue
		}
		if reqSegs[i] != seg {
			return nil, false
		}
	}
	return vals, true
}

func wildcardName(seg string) (string, bool) {
	if len(seg) >= 2 && seg[0] == '{' && seg[len(seg)-1] == '}' {
		return seg[1 : len(seg)-1], true
	}
	return "", false
}

// specificity ranks a route so the most specific match wins, mirroring the
// standard mux preferring exact literals and longer patterns over subtree "/".
func (r route) specificity() int {
	score := len(r.segments) * 10
	if r.wildcard {
		score -= 5
	}
	for _, seg := range r.segments {
		if _, ok := wildcardName(seg); !ok {
			score++
		}
	}
	return score
}

// ServeHTTP routes req to the best-matching handler.
func (m *Mux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	reqSegs := splitPath(req.URL.Path)
	var best *route
	var bestVals map[string]string
	bestScore := 0
	methodMismatch := false
	for i := range m.routes {
		r := &m.routes[i]
		vals, ok := r.match(reqSegs)
		if !ok {
			continue
		}
		if r.method != "" && r.method != req.Method {
			methodMismatch = true
			continue
		}
		if s := r.specificity(); best == nil || s > bestScore {
			bestScore = s
			best = r
			bestVals = vals
		}
	}
	if best == nil {
		if methodMismatch {
			http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
			return
		}
		http.NotFound(w, req)
		return
	}
	if len(bestVals) > 0 {
		req = req.WithContext(context.WithValue(req.Context(), ctxKey{}, bestVals))
	}
	best.handler(w, req)
}

// PathValue returns the value captured for the named wildcard segment of the
// route that matched req, or "" if there is none. It replaces
// (*http.Request).PathValue (Go 1.22).
func PathValue(req *http.Request, name string) string {
	vals, _ := req.Context().Value(ctxKey{}).(map[string]string)
	return vals[name]
}
