package pathmatcher

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-openapi/runtime/middleware/denco"
)

// PathMatcher holds the information of a matched
// path in different formats
type PathMatched struct {
	Method      string
	OpenAPIPath string
	RouterPath  string
	RedisKey    string
}

func NewPathMatched(method, openApiPath string) *PathMatched {
	pathConverter := regexp.MustCompile(`{(.+?)}([^/]*)`)
	mn := strings.ToUpper(method)
	conv := pathConverter.ReplaceAllString(openApiPath, ":$1")
	redisKeyPath := fmt.Sprintf("%s_%s", mn, openApiPath)
	return &PathMatched{
		Method:      mn,
		OpenAPIPath: openApiPath,
		RouterPath:  conv,
		RedisKey:    redisKeyPath,
	}
}

// Matcher defines the interface to lookup a a path
type Matcher interface {
	LookupRoute(method, pathWithParams string) *PathMatched
}

type PathMatcher struct {
	records map[string][]denco.Record
	routers map[string]*denco.Router
}

func NewPathMatcher() *PathMatcher {
	return &PathMatcher{
		records: make(map[string][]denco.Record),
		routers: make(map[string]*denco.Router),
	}
}

func (pm *PathMatcher) AddRoute(method, path string) {
	pathMatched := NewPathMatched(method, path)
	record := denco.NewRecord(pathMatched.RouterPath, pathMatched)
	pm.records[pathMatched.Method] = append(
		pm.records[pathMatched.Method], record)
}

func (pm *PathMatcher) LookupRoute(method, pathWithParams string) *PathMatched {
	method = strings.ToUpper(method)
	r, ok := pm.routers[method]
	if !ok {
		// fmt.Printf("routers %#v\n", pm.routers)
		return nil
	}
	// we ignore the params
	res, _, found := r.Lookup(pathWithParams)
	if !found {
		/*
			fmt.Printf("Lookup NOT found: %s -> %s, %#v, %t\n%#v\n",
				pathWithParams, res, params, found, pm.routers[method])
		*/
		return nil
	}
	// fmt.Printf("Lookup Params:\n%#v\n", params)
	p, ok := res.(*PathMatched)
	if !ok {
		return nil
	}
	return p
}

func (pm *PathMatcher) buildRouter(records map[string][]denco.Record) map[string]*denco.Router {
	routers := make(map[string]*denco.Router)
	for method, records := range pm.records {
		router := denco.New()
		_ = router.Build(records)
		routers[method] = router
	}
	return routers
}

func (pm *PathMatcher) Build() {
	pm.routers = pm.buildRouter(pm.records)
}
