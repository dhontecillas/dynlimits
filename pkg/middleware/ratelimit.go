package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/dhontecillas/dynlimits/pkg/catalog"
	"github.com/dhontecillas/dynlimits/pkg/pathmatcher"
	"github.com/dhontecillas/dynlimits/pkg/ratelimit"
)

// RateLimitMiddleware contains the data required
// to implement per endpoint rate limiting using
// a catalog of valid api keys.
//
// For unknow endpoints we can choose if we let them
// pass or if we deny the access
type RateLimitMiddleware struct {
	next              http.Handler
	apiKeyHeader      string
	apiKeyCatalog     catalog.APIKeys
	redisPool         *redis.Pool
	matcher           pathmatcher.Matcher
	allowUnknownPaths bool
}

// NewRateLimitMiddleware creates a new RateLimitMiddleware
func NewRateLimitMiddleware(next http.Handler, apiKeyHeader string,
	apiKeyCatalog catalog.APIKeys, redisPool *redis.Pool,
	matcher pathmatcher.Matcher) *RateLimitMiddleware {

	return &RateLimitMiddleware{
		next:              next,
		apiKeyHeader:      apiKeyHeader,
		apiKeyCatalog:     apiKeyCatalog,
		redisPool:         redisPool,
		matcher:           matcher,
		allowUnknownPaths: false,
	}
}

// ServeHTTP
// https://tools.ietf.org/id/draft-polli-ratelimit-headers-00.html
func (rlm *RateLimitMiddleware) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	apiKey, ok := req.Header[rlm.apiKeyHeader]
	if !ok || len(apiKey) == 0 || len(apiKey[0]) == 0 {
		// no key, no request :)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	ak := apiKey[0]
	now := time.Now().Unix()

	// TODO: check locally the existence of the API key

	pm := rlm.matcher.LookupRoute(req.Method, req.URL.Path)
	if pm == nil {
		if rlm.allowUnknownPaths {
			rlm.next.ServeHTTP(rw, req)
		} else {
			rw.WriteHeader(http.StatusNotFound)
			return
		}
	}

	key := fmt.Sprintf("%s_%s", ak, pm.RedisKey)

	conn := rlm.redisPool.Get()
	if conn == nil {
		// TODO: review what to do here, and if we want to put a flag
		// to select behaviour
		// if we cannot connect to redis, we let the request pass for now
		rlm.next.ServeHTTP(rw, req)
		return
	}
	// WARNING !!!
	//
	// we want to close the connection as soon as possible, so we do not
	// rely on defer. It is here only for safety, if more code is added,
	// try to close the connection as soon as possible
	defer conn.Close()
	//
	wnd, err := ratelimit.GetRedisSlidingCountersWindow(conn, key, now)
	if err != nil {
		// TODO: review what to do here, and if we want to put a flag
		// to select behaviour
		// error fetching from redis the window, we let the request pass
		conn.Close()
		rlm.next.ServeHTTP(rw, req)
		return
	}

	header := rw.Header()
	header.Add("RateLimit-Limit", strconv.FormatInt(wnd.ReqPerMin, 10))
	header.Add("RateLimit-Reset", strconv.Itoa(wnd.NumEmptySlotsAtStart()))
	if wnd.Sum >= wnd.ReqPerMin {
		// TODO: here we can save and optimize later to not have to
		// go to redis on the next request
		conn.Close()
		header.Add("RateLimit-Remaining", "0")
		rw.WriteHeader(http.StatusTooManyRequests)
		return
	}
	header.Add("RateLimit-Remaining", strconv.FormatInt(wnd.ReqPerMin-wnd.Sum-1, 10))
	err = ratelimit.AddToRedisSlidingCountersWindow(conn, key, now)
	conn.Close()
	rlm.next.ServeHTTP(rw, req)
}
