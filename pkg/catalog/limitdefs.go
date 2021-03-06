package catalog

import (
	"fmt"
	"time"
)

// EndpointIndexedDef contains the index
// to the path definition, and an index
// to the http verb to define an endpoint
type EndpointIndexedDef struct {
	PathIdx   int `json:"p"`
	MethodIdx int `json:"m"`
}

// EndpointIndexedLimits contains an index
// to the list of endpoints definitions and
// the limit to apply to this endpoint.
type EndpointIndexedLimits struct {
	EndpointIdx int   `json:"ep"`
	RateLimit   int64 `json:"rl"`
}

// APIKeyIndexedLimits has the limits to be
// applied to a given API Key. See
// `EndpointsIndexedLimits` to see how to
// reference tha endpoints.
type APIKeyIndexedLimits struct {
	APIKey string                  `json:"key"`
	Limits []EndpointIndexedLimits `json:"limits"`
}

// APICatalogVersion contains the version information
// for an API Catalog
//
// - SemVer: a semantic version v3.22.15 (not useful
//		for machine consumption)
// - HashVer: a hash generated from the catalog data,
// 		the semver, and the release date
// - Released: the time at what this catalog was created
// 		or updated
type APICatalogVersion struct {
	SemVer   string    `json:"semver"`
	HashVer  string    `json:"hash"`
	Released time.Time `json:"released"`
}

// APIIndexedLimits contains all the information required
// to perform per endpoint and api key rate limits
type APIIndexedLimits struct {
	Version   APICatalogVersion     `json:"version"`
	Methods   []string              `json:"methods"`
	Paths     []string              `json:"paths"`
	Endpoints []EndpointIndexedDef  `json:"endpoints"`
	APILimits []APIKeyIndexedLimits `json:"apilimits"`
}

// Validate checks that all indices point to valid positions
// in each of the lists.
func (ail *APIIndexedLimits) Validate() []error {
	errs := []error{}
	for idx, ep := range ail.Endpoints {
		if ep.PathIdx < 0 || ep.PathIdx >= len(ail.Paths) {
			errs = append(errs,
				fmt.Errorf("Bad PathIdx in Endpoint %d (%#v)",
					idx, ep))
		}
		if ep.MethodIdx < 0 || ep.MethodIdx >= len(ail.Methods) {
			errs = append(errs,
				fmt.Errorf("Bad MethodIdx in Endpoint %d (%#v)",
					idx, ep))
		}
	}

	for apiLimIdx, akil := range ail.APILimits {
		for limIdx, lim := range akil.Limits {
			if lim.EndpointIdx < 0 || lim.EndpointIdx >= len(ail.Endpoints) {
				errs = append(errs,
					fmt.Errorf("Bad EndpointIdx in APILim %d, limIdx: %d (%#v)",
						apiLimIdx, limIdx, lim))
			}
		}
	}
	return errs
}
