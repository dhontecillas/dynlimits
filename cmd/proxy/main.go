package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime/pprof"

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

var profFile *os.File

func startProfiling() {
	var err error
	profFile, err = os.Create("cpuprofile")
	if err != nil {
		fmt.Printf("cannot start cpuprofiling %s\n", err.Error())
		return
	}
	if err := pprof.StartCPUProfile(profFile); err != nil {
		fmt.Printf("cannot start cpuprofiling %s\n", err.Error())
		return
	}

	fmt.Printf("started profiling\n")
}

func endProfiling() {
	fmt.Printf("end profile called\n")
	if profFile != nil {
		fmt.Printf("recording profile\n")
		pprof.StopCPUProfile()
		profFile.Close()
	} else {
		fmt.Printf("was not opened the os.File\n")
	}
}

func main() {
	fmt.Println("DynLimits proxy")

	startProfiling()
	defer endProfiling()

	conf := config.LoadConf()

	pool := ratelimit.NewRedisPool(
		ratelimit.NewRedisPoolConf(conf.RedisAddress))
	conn := pool.Get()
	defer conn.Close()

	var indexedLimits catalog.APIIndexedLimits
	if len(conf.CatalogFile) > 0 {
		initialCatalog, err := ioutil.ReadFile(conf.CatalogFile)
		if err != nil {
			fmt.Printf("cannot read the config file: %s\n", err.Error())
			return
		}
		if err := json.Unmarshal([]byte(initialCatalog), &indexedLimits); err != nil {
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
		fmt.Printf("initial catalog file:\n %#v\n", conf)
	} else {
		fmt.Printf("no initial catalog file\n: %#v\n", conf)
	}

	// crate the shared path matcher
	globalSharedPathMatcher = pathmatcher.NewSharedPathMatcher(
		pathmatcher.NewPathMatcher())

	catalog.UpdateSharedMatcher(&indexedLimits, globalSharedPathMatcher)

	// TODO: move this to a unit test case:
	// checking that the route was added
	/*
		rtest := globalSharedPathMatcher.LookupRoute("GET",
			"/api/filipid/recipients/foooo/preferences")
		if rtest == nil {
			fmt.Printf("the routes were not well loaded\nIMPLEMENT ROUTE LOADING\n")
			return
		}
	*/

	// now update all api keys in the redis server
	catalog.RedisUpdate(conn, &indexedLimits)

	if len(conf.CatalogServerURL) > 0 {
		_, err := catalog.LaunchUpdatesPoller(
			pool, conf.CatalogServerURL, conf.CatalogServerAPIKey,
			globalSharedPathMatcher, conf.CatalogRedisPollSecs,
			conf.CatalogServerPollSecs)
		if err != nil {
			// TODO: log the error and decide what to do with it
			fmt.Printf("cannot launch the policy updater: %s\n", err.Error())
			return
		}
	}

	proxyH := proxy.NewProxyHandler(conf.ForwardToScheme, conf.ForwardAddr())

	defApiKeyCatalog := catalog.DefaultAPIKeys{}
	rateLimitH := middleware.NewRateLimitMiddleware(proxyH,
		"X-Api-Key", &defApiKeyCatalog, pool, globalSharedPathMatcher)

	// server.LaunchBlockingServer(proxyH)
	fmt.Printf("conf.ListenAddr: %s\n", conf.ListenAddr())

	// srv := server.LaunchBlockingServer(conf.ListenAddr(), rateLimitH)
	//testRedisSlidingCounterWindow(conn)

	srv := server.LaunchBackgroundServer(conf.ListenAddr(), rateLimitH)
	// install a signal handler for shutdown
	sigkillchan := make(chan os.Signal)
	signal.Notify(sigkillchan, os.Interrupt)
	_ = <-sigkillchan
	server.ShutdownServer(srv)
	endProfiling()

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
