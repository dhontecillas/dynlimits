package main

import (
	"encoding/json"
	"fmt"

	"github.com/dhontecillas/dynlimits/pkg/catalog"
	"github.com/dhontecillas/dynlimits/pkg/config"
	"github.com/dhontecillas/dynlimits/pkg/middleware"
	"github.com/dhontecillas/dynlimits/pkg/pathmatcher"
	"github.com/dhontecillas/dynlimits/pkg/proxy"
	"github.com/dhontecillas/dynlimits/pkg/ratelimit"
	"github.com/dhontecillas/dynlimits/pkg/server"
	"github.com/gomodule/redigo/redis"
)

var (
	globalSharedPathMatcher *pathmatcher.SharedPathMatcher
)

func main() {
	fmt.Println("DynLimits proxy")

	pool := ratelimit.NewRedisPool(
		ratelimit.NewRedisPoolConf("localhost:6379"))

	conn := pool.Get()
	defer conn.Close()

	conf := config.LoadConf()

	// TODO: Load from control backend, and fallback to local file
	// if control backend is not available.

	// TODO: check if we should let pass `OPTIONS` verb

	tmpConfig := `
{
	"methods": ["GET", "POST", "DELETE", "PUT", "PATCH", "HEAD", "OPTIONS"],
	"paths": [
		"\/api\/filipid\/recipients\/{recipient_id}\/preferences",
		"\/api\/filipid\/recipients"
	],
	"endpoints": [
		{"p": 0, "m": 0},
		{"p": 1, "m": 0}
	],
	"apilimits": [
		{
			"key": "7H6AMB0FXQKQBG3JKPW1PXTTNW",
		  	"limits": [
				{ "ep": 0, "rl": 20 },
				{ "ep": 1, "rl": 5 }
		 	]
		}
	]
}
`
	var indexedLimits catalog.APIIndexedLimits
	if err := json.Unmarshal([]byte(tmpConfig), &indexedLimits); err != nil {
		fmt.Printf("cannot load the config: %s\n", err.Error())
		return
	}

	indexedLimitsErrs := indexedLimits.Validate()
	if len(indexedLimitsErrs) > 0 {
		fmt.Printf("BAD indexed limits\n")
		for _, e := range indexedLimitsErrs {
			fmt.Printf("- %s\n", e.Error())
		}
		fmt.Printf("---------- Aborting\n")
		return
	}

	// crate the shared path matcher
	globalSharedPathMatcher = pathmatcher.NewSharedPathMatcher(
		pathmatcher.NewPathMatcher())

	catalog.UpdateSharedMatcher(&indexedLimits, globalSharedPathMatcher)

	// checking that the route was added
	rtest := globalSharedPathMatcher.LookupRoute("GET",
		"/api/filipid/recipients/foooo/preferences")
	if rtest == nil {
		fmt.Printf("the routes were not well loaded\nIMPLEMENT ROUTE LOADING\n")
		return
	}

	// now update all api keys in the redis server
	catalog.RedisUpdate(conn, &indexedLimits)

	// shutdownChan
	_, err := catalog.LaunchUpdatesPoller(
		pool, "http://localhost:7887/api/dynlimits/v1",
		"FAKE_API_KEY", globalSharedPathMatcher, 3, 5)
	if err != nil {
		// TODO: log the error and decide what to do with it
		fmt.Printf("cannot launch the policy updater: %s\n", err.Error())
		return
	}

	proxyH := proxy.NewProxyHandler(conf.ForwardToScheme, conf.ForwardAddr())

	defApiKeyCatalog := catalog.DefaultAPIKeys{}
	rateLimitH := middleware.NewRateLimitMiddleware(proxyH,
		"X-Api-Key", &defApiKeyCatalog, pool, globalSharedPathMatcher)

	// server.LaunchBlockingServer(proxyH)
	server.LaunchBlockingServer(conf.ListenAddr(), rateLimitH)
	//testRedisSlidingCounterWindow(conn)
}

func testRedisSlidingCounterWindow(conn redis.Conn) {

	keyPrefix := "kk"

	var err error
	err = ratelimit.SetRedisRateLimit(conn, keyPrefix, 122)
	if err != nil {
		fmt.Printf("cannot set rate limit\n")
	}
	err = ratelimit.AddToRedisSlidingCountersWindow(conn, keyPrefix, 1999)
	if err != nil {
		fmt.Printf("cannot add to ratelimit %s\n", err.Error())
		return
	}
	err = ratelimit.AddToRedisSlidingCountersWindow(conn, keyPrefix, 1998)
	if err != nil {
		fmt.Printf("cannot add to ratelimit %s\n", err.Error())
		return
	}

	rrl, err := ratelimit.GetRedisSlidingCountersWindow(conn, keyPrefix, 2000)
	if err != nil {
		fmt.Printf("RRL Err: %s\n", err.Error())
		return
	}

	rrl.Print()
}
