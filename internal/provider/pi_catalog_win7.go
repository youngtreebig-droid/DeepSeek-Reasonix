//go:build win7

package provider

// Win7 build stub for the sky-valley/pi model catalog.
//
// The upstream github.com/sky-valley/pi module targets Go 1.26 and uses the
// Go 1.23 range-over-func (iter) language feature, neither of which the
// go1.20.14 toolchain that produces Win7-runnable binaries can compile. The
// real catalog lives in pi_catalog.go behind `//go:build !win7`; this file
// provides API-compatible stubs so the rest of the provider package (and its
// callers in internal/config) compile without pulling in pi.
//
// Behaviour: the embedded catalog is reported as empty. Callers already treat
// a miss as "no catalog facts" and fall back to the built-in/static model
// metadata, so the Win7 build simply never inherits pi catalog data.

// piCatalogModel mirrors the subset of github.com/sky-valley/pi/ai.Model that
// callers in this package read (see opencode_go.go). Kept unexported and only
// used by the stub helpers below.
type piCatalogModel struct {
	ID            string
	ContextWindow int
	MaxTokens     int
}

// PiCatalogModelInfo returns no catalog match on the Win7 build.
func PiCatalogModelInfo(kind, baseURL, model string) (ModelInfo, bool) {
	return ModelInfo{}, false
}

// PiCatalogModelInfoForProvider returns no catalog match on the Win7 build.
func PiCatalogModelInfoForProvider(providerID, kind, baseURL, model string) (ModelInfo, bool) {
	return ModelInfo{}, false
}

// PiCatalogModelInfos returns an empty catalog on the Win7 build.
func PiCatalogModelInfos(providerID string) []ModelInfo {
	return nil
}

// PiCatalogOpenCodeGoModelIDs returns no catalog IDs on the Win7 build.
func PiCatalogOpenCodeGoModelIDs(route string) []string {
	return nil
}

// PiCatalogOpenCodeGoReasoning reports no catalog reasoning facts on Win7.
func PiCatalogOpenCodeGoReasoning(route, id string) (ReasoningCapability, bool) {
	return ReasoningCapability{}, false
}

// PiCatalogOpenCodeGoVisionModelIDs returns no catalog vision IDs on Win7.
func PiCatalogOpenCodeGoVisionModelIDs(route string) []string {
	return nil
}

// piCatalogOpenCodeGoModels returns an empty catalog on the Win7 build. The
// element type carries the same field names opencode_go.go reads.
func piCatalogOpenCodeGoModels(route string) []*piCatalogModel {
	return nil
}
