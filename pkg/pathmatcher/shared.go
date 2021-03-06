package pathmatcher

import (
	"sync"

	"github.com/go-openapi/runtime/middleware/denco"
)

type SharedPathMatcher struct {
	matcher       *PathMatcher
	recordsAccess sync.RWMutex
	routerAccess  sync.RWMutex
}

type ChangeSet struct {
	finished bool
	spm      *SharedPathMatcher
}

func NewSharedPathMatcher(matcher *PathMatcher) *SharedPathMatcher {
	return &SharedPathMatcher{
		matcher: matcher,
	}
}

func (spm *SharedPathMatcher) LookupRoute(method, pathWithParams string) *PathMatched {
	spm.routerAccess.RLock()
	defer spm.routerAccess.RUnlock()
	return spm.matcher.LookupRoute(method, pathWithParams)
}

func (spm *SharedPathMatcher) StartChangeSet() *ChangeSet {
	spm.recordsAccess.Lock()
	return &ChangeSet{
		finished: false,
		spm:      spm,
	}
}

func (cs *ChangeSet) AddRoute(method, path string) {
	if cs.finished {
		return
	}
	pm := NewPathMatched(method, path)
	byMethod, ok := cs.spm.matcher.records[pm.Method]
	if ok {
		// checks that the entry do not exists
		for _, r := range byMethod {
			if r.Key == pm.RouterPath {
				return
			}
		}
	} else {
		byMethod = make([]denco.Record, 0, 4)
	}
	record := denco.NewRecord(pm.RouterPath, pm)
	byMethod = append(byMethod, record)
	cs.spm.matcher.records[pm.Method] = byMethod
}

func (cs *ChangeSet) RemoveRoute(method, path string) {
	if cs.finished {
		return
	}
	// by creating a new path matches  we make sure
	// method is upper case, and we have a path in
	// the router format (the one that is stored in r.Key)
	pm := NewPathMatched(method, path)
	byMethod, ok := cs.spm.matcher.records[pm.Method]
	if !ok {
		return
	}

	for idx, r := range byMethod {
		if r.Key == pm.RouterPath {
			if idx < len(byMethod)-1 {
				// denco router, does not care about order of records
				// https://github.com/naoina/denco#url-pattern
				// so we can swap last record with found one
				byMethod[idx] = byMethod[len(byMethod)-1]
			}
			// just shorten the slice
			cs.spm.matcher.records[pm.Method] = byMethod[:len(byMethod)-1]
		}
	}
}

func (cs *ChangeSet) RemoveAll() {
	if cs.finished {
		return
	}
	// clean the underlying dengo record by reseting all the slices
	// but  maintaining their capacities
	for k, _ := range cs.spm.matcher.records {
		if len(cs.spm.matcher.records[k]) > 0 {
			cs.spm.matcher.records[k] = cs.spm.matcher.records[k][:0]
		}
	}
}

func (cs *ChangeSet) Commit() {
	if cs.finished {
		return
	}
	cs.finished = true
	newRouters := cs.spm.matcher.buildRouter(cs.spm.matcher.records)
	cs.spm.routerAccess.Lock()
	cs.spm.matcher.routers = newRouters
	cs.spm.routerAccess.Unlock()
	cs.spm.recordsAccess.Unlock()
}
