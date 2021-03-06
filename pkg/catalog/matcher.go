package catalog

import (
	"github.com/dhontecillas/dynlimits/pkg/pathmatcher"
)

// UpdateSharedMatcher updates the valid paths where rate limits can
// be applied to
func UpdateSharedMatcher(ail *APIIndexedLimits, pm *pathmatcher.SharedPathMatcher) {
	cs := pm.StartChangeSet()

	cs.RemoveAll()
	for _, ep := range ail.Endpoints {
		mIdx := ep.MethodIdx
		pIdx := ep.PathIdx
		if mIdx < 0 || mIdx >= len(ail.Methods) || pIdx < 0 || pIdx > len(ail.Paths) {
			continue
		}
		cs.AddRoute(ail.Methods[mIdx], ail.Paths[pIdx])
	}
	cs.Commit()
}
