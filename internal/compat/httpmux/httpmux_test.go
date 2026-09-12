package httpmux

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// newProbeMux builds a mux whose handlers record which route fired (by writing
// a label to the body) and echo any captured path values, so tests can assert
// both routing precedence and PathValue capture in one place.
func labelHandler(label string, pathVars ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(label))
		for _, name := range pathVars {
			_, _ = w.Write([]byte("|" + name + "=" + PathValue(r, name)))
		}
	}
}

func do(t *testing.T, m *Mux, method, path string) (int, string) {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, req)
	return rec.Code, rec.Body.String()
}

func TestMethodMatchAnd405(t *testing.T) {
	m := New()
	m.HandleFunc("GET /widgets", labelHandler("get-widgets"))
	m.HandleFunc("POST /widgets", labelHandler("post-widgets"))

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantBody   string
	}{
		{"GET matches", http.MethodGet, "/widgets", http.StatusOK, "get-widgets"},
		{"POST matches", http.MethodPost, "/widgets", http.StatusOK, "post-widgets"},
		{"DELETE -> 405", http.MethodDelete, "/widgets", http.StatusMethodNotAllowed, ""},
		{"PUT -> 405", http.MethodPut, "/widgets", http.StatusMethodNotAllowed, ""},
		{"unknown path -> 404", http.MethodGet, "/nope", http.StatusNotFound, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			code, body := do(t, m, tc.method, tc.path)
			if code != tc.wantStatus {
				t.Fatalf("%s %s: status = %d, want %d", tc.method, tc.path, code, tc.wantStatus)
			}
			if tc.wantBody != "" && body != tc.wantBody {
				t.Fatalf("%s %s: body = %q, want %q", tc.method, tc.path, body, tc.wantBody)
			}
		})
	}
}

func TestWildcardCaptureAndPathValue(t *testing.T) {
	m := New()
	m.HandleFunc("GET /inbox/items/{id}", labelHandler("item", "id"))
	m.HandleFunc("GET /users/{uid}/posts/{pid}", labelHandler("post", "uid", "pid"))

	tests := []struct {
		name     string
		path     string
		wantBody string
	}{
		{"single wildcard", "/inbox/items/abc123", "item|id=abc123"},
		{"two wildcards", "/users/42/posts/99", "post|uid=42|pid=99"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			code, body := do(t, m, http.MethodGet, tc.path)
			if code != http.StatusOK {
				t.Fatalf("%s: status = %d, want 200", tc.path, code)
			}
			if body != tc.wantBody {
				t.Fatalf("%s: body = %q, want %q", tc.path, body, tc.wantBody)
			}
		})
	}

	// A wildcard matches exactly one segment: /inbox/items with no id, or with
	// an extra segment, must not match the single-wildcard route.
	if code, _ := do(t, m, http.MethodGet, "/inbox/items"); code != http.StatusNotFound {
		t.Errorf("/inbox/items (missing id) status = %d, want 404", code)
	}
	if code, _ := do(t, m, http.MethodGet, "/inbox/items/a/b"); code != http.StatusNotFound {
		t.Errorf("/inbox/items/a/b (extra segment) status = %d, want 404", code)
	}
}

func TestSubtreePrecedenceVsLiteral(t *testing.T) {
	m := New()
	// A subtree and a more specific literal/wildcard under it. The most
	// specific match must win, mirroring the standard mux.
	m.HandleFunc("GET /api/", labelHandler("subtree"))
	m.HandleFunc("GET /api/health", labelHandler("literal-health"))
	m.HandleFunc("GET /api/users/{id}", labelHandler("user", "id"))

	tests := []struct {
		name     string
		path     string
		wantBody string
	}{
		{"literal beats subtree", "/api/health", "literal-health"},
		{"wildcard beats subtree", "/api/users/7", "user|id=7"},
		{"subtree catches the rest", "/api/anything/else", "subtree"},
		{"subtree catches its own prefix", "/api/", "subtree"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, body := do(t, m, http.MethodGet, tc.path)
			if body != tc.wantBody {
				t.Fatalf("%s: body = %q, want %q", tc.path, body, tc.wantBody)
			}
		})
	}
}

// TestRootCatchAll guards the exact regression the review calls out: `GET /` is
// a subtree with segments=nil scoring negative specificity, and it must still
// win for `/` (and for any path not matched by a more specific route) rather
// than 404. This test fails if the "best == nil" first-match selection is
// reverted to a score-seeded comparison.
func TestRootCatchAll(t *testing.T) {
	m := New()
	m.HandleFunc("GET /", labelHandler("root"))
	m.HandleFunc("GET /status", labelHandler("status"))

	tests := []struct {
		name     string
		path     string
		wantBody string
	}{
		{"root serves index", "/", "root"},
		{"root catches unknown", "/whatever/deep/path", "root"},
		{"specific route still wins", "/status", "status"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			code, body := do(t, m, http.MethodGet, tc.path)
			if code != http.StatusOK {
				t.Fatalf("%s: status = %d, want 200", tc.path, code)
			}
			if body != tc.wantBody {
				t.Fatalf("%s: body = %q, want %q", tc.path, body, tc.wantBody)
			}
		})
	}
}

// TestSpecificityOrderingIndependentOfRegistration verifies that a
// less-specific route registered AFTER a more-specific one cannot displace it,
// which is the property the GET / fix relies on (`s > bestScore`, strictly
// greater). Registration order must not affect which handler wins.
func TestSpecificityOrderingIndependentOfRegistration(t *testing.T) {
	// Order A: subtree first, literal second.
	a := New()
	a.HandleFunc("GET /files/", labelHandler("subtree"))
	a.HandleFunc("GET /files/readme", labelHandler("literal"))

	// Order B: literal first, subtree second.
	b := New()
	b.HandleFunc("GET /files/readme", labelHandler("literal"))
	b.HandleFunc("GET /files/", labelHandler("subtree"))

	for name, m := range map[string]*Mux{"subtree-first": a, "literal-first": b} {
		if _, body := do(t, m, http.MethodGet, "/files/readme"); body != "literal" {
			t.Errorf("%s: /files/readme served %q, want literal", name, body)
		}
		if _, body := do(t, m, http.MethodGet, "/files/other"); body != "subtree" {
			t.Errorf("%s: /files/other served %q, want subtree", name, body)
		}
	}
}

// TestMethodlessRouteMatchesAnyMethod checks that a pattern registered without
// a method (method == "") matches every HTTP method.
func TestMethodlessRouteMatchesAnyMethod(t *testing.T) {
	m := New()
	m.HandleFunc("/ping", labelHandler("ping"))

	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodDelete} {
		code, body := do(t, m, method, "/ping")
		if code != http.StatusOK || body != "ping" {
			t.Errorf("%s /ping: code=%d body=%q, want 200 ping", method, code, body)
		}
	}
}

// TestPathNormalizationDivergence locks in the documented, intentional
// divergence from net/http.ServeMux: unclean paths are neither cleaned nor
// 301-redirected. A doubled slash produces an empty segment that is kept, so
// the unclean path does not match the clean route and (with no matching
// subtree) yields 404 rather than the standard mux's 301. If someone later
// adds path cleaning, this test flags the behavioral change so the package doc
// can be reconciled.
func TestPathNormalizationDivergence(t *testing.T) {
	m := New()
	m.HandleFunc("GET /inbox/items", labelHandler("items"))

	// Clean path matches.
	if code, body := do(t, m, http.MethodGet, "/inbox/items"); code != http.StatusOK || body != "items" {
		t.Fatalf("/inbox/items: code=%d body=%q, want 200 items", code, body)
	}

	// Doubled slash is NOT collapsed and NOT redirected: it does not match the
	// clean route, so it 404s (the standard mux would 301 to the cleaned path).
	code, _ := do(t, m, http.MethodGet, "/inbox//items")
	if code == http.StatusMovedPermanently {
		t.Fatalf("/inbox//items unexpectedly redirected (301); package doc says paths are not cleaned")
	}
	if code != http.StatusNotFound {
		t.Fatalf("/inbox//items: code=%d, want 404 (unclean path not matched, per package doc)", code)
	}
}

// TestPathNormalizationSubtreeStillMatches documents that when a subtree route
// covers the prefix, an unclean path under it still resolves via the subtree
// (the trailing empty/extra segments are simply captured as part of the
// subtree match), never via a redirect.
func TestPathNormalizationSubtreeStillMatches(t *testing.T) {
	m := New()
	m.HandleFunc("GET /inbox/", labelHandler("inbox-subtree"))

	if code, body := do(t, m, http.MethodGet, "/inbox//items"); code != http.StatusOK || body != "inbox-subtree" {
		t.Fatalf("/inbox//items under subtree: code=%d body=%q, want 200 inbox-subtree", code, body)
	}
}

// TestMethodMismatchDoesNotMask404 verifies that a 405 is only returned when a
// path matched but the method did not, and that the root catch-all (if present)
// takes precedence over a sibling 405 — mirroring the standard mux.
func TestMethodMismatchVsCatchAll(t *testing.T) {
	m := New()
	m.HandleFunc("POST /submit", labelHandler("submit"))
	m.HandleFunc("GET /", labelHandler("root"))

	// GET /submit: the POST /submit route mismatches method, but GET / catches
	// it, so we serve root (200), not 405.
	if code, body := do(t, m, http.MethodGet, "/submit"); code != http.StatusOK || body != "root" {
		t.Errorf("GET /submit: code=%d body=%q, want 200 root (catch-all wins over 405)", code, body)
	}
}

func TestMethodMismatch405WithoutCatchAll(t *testing.T) {
	m := New()
	m.HandleFunc("POST /submit", labelHandler("submit"))

	if code, _ := do(t, m, http.MethodGet, "/submit"); code != http.StatusMethodNotAllowed {
		t.Errorf("GET /submit without catch-all: code=%d, want 405", code)
	}
}

func TestHandleWrapsHandler(t *testing.T) {
	m := New()
	m.Handle("GET /h", http.HandlerFunc(labelHandler("handler")))
	if code, body := do(t, m, http.MethodGet, "/h"); code != http.StatusOK || body != "handler" {
		t.Errorf("Handle: code=%d body=%q, want 200 handler", code, body)
	}
}
