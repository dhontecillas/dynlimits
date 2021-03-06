package catalog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dhontecillas/dynlimits/pkg/pathmatcher"
	"github.com/gomodule/redigo/redis"
)

// CatalogURLs contains all the endpoints to get
// information from the control server
//
// 	- GetLatestVersion: an url to get the latest hash for the
//		most recent version of the catalog. (We could get the
// 		catalog with an ETag)
//	- GetDiff: an url to get the difference between two versions
//		with from and to params. If to is not provided, it is
//		assumed to be the most recent one. The diff is always "forward"
// 		meaning that from is always before in time than to
//  - GetIndexedCatalog: to retrieve the full catalog (to be used
//		if there is no existing version in the redis ? or we
//		want a full rebuild)
//
type CatalogURLs struct {
	GetLatestVersion  string
	GetDiff           string
	GetIndexedCatalog string
}

// CatalogUpdater holds the information required to keep
// the catalago updated from the command and control server
//
//	- urls: the command and control URL server
// 	- catalogAPIKey: the api to use when requesting updates
// 	- serverCheckSeconds: how often we need to poll the server
//		for new updates
//	- redisCheckSeconds: we check in redis if some other proxy
//		has already updated to a new version the data in Redis
//  - redisPool: a pool of connections for redis
//
//  - RequestOnDemandUpdate: a channel to be used by the client
//		code to force an update
//	- RequestShutdown: a channel to notify that we want the poller
//		to stop
type CatalogUpdater struct {
	urls               CatalogURLs
	catalogAPIKey      string
	redisCheckSeconds  int64
	serverCheckSeconds int64
	redisPool          *redis.Pool
	matcher            *pathmatcher.SharedPathMatcher

	RequestOnDemandUpdate chan bool
	RequestShutdown       chan bool
}

// check server for new updated since the latest version
func (cu *CatalogUpdater) getLatestVersion() (*APICatalogVersion, error) {
	nr, err := http.NewRequest("GET", cu.urls.GetLatestVersion, nil)
	if err != nil {
		return nil, err
	}
	nr.Header.Add("X-Api-Key", cu.catalogAPIKey)

	client := &http.Client{}
	res, err := client.Do(nr)
	if err != nil {
		// TODO: log the error, is important if we
		// cannot connect to the control server !!!
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		// TODO: log the error, is important if we
		// cannot connect to the control server or
		// if we are providing bad login credentials
		return nil, err
	}
	// read the response body into bytes
	var buf []byte
	if res.ContentLength >= 0 {
		buf = make([]byte, 0, res.ContentLength)
	}
	b := bytes.NewBuffer(buf)
	b.ReadFrom(res.Body)

	// parse the server response
	var serverVer APICatalogVersion
	err = json.Unmarshal(b.Bytes(), &serverVer)
	if err != nil {
		// TODO: log the error, IMPORTANT !! if the
		// server is sending a bad format
		return nil, err
	}

	return &serverVer, nil
}

// getCatalogFromServer requests the latest full catalog
// from the server
func (cu *CatalogUpdater) getCatalogFromServer() (*APIIndexedLimits, error) {
	nr, err := http.NewRequest("GET", cu.urls.GetIndexedCatalog, nil)
	if err != nil {
		return nil, err
	}
	nr.Header.Add("X-Api-Key", cu.catalogAPIKey)

	client := &http.Client{}
	res, err := client.Do(nr)
	if err != nil {
		// TODO: log the error, is important if we
		// cannot connect to the control server !!!
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		// TODO: log the error, is important if we
		// cannot connect to the control server or
		// if we are providing bad login credentials
		return nil, fmt.Errorf("status code %d", res.StatusCode)
	}
	// read the response body into bytes
	var buf []byte
	if res.ContentLength >= 0 {
		buf = make([]byte, 0, res.ContentLength)
	} else {
		fmt.Printf("no content length")
		buf = make([]byte, 0, 1024)
	}
	b := bytes.NewBuffer(buf)
	b.ReadFrom(res.Body)

	// parse the server response
	var c APIIndexedLimits
	err = json.Unmarshal(b.Bytes(), &c)
	if err != nil {
		// TODO: log the error, IMPORTANT !! if the
		// server is sending a bad format
		return nil, err
	}
	return &c, nil
}

// getCatalogFromRedis gets the current list of limits
// for each api key and endpoint from the redis server
func (cu *CatalogUpdater) getCatalogFromRedis() (*APIIndexedLimits, error) {
	return nil, fmt.Errorf("not implemented")
}

func (cu *CatalogUpdater) checkUpdateFromServer() {
	indexedCatalog, err := cu.getCatalogFromServer()
	if err != nil {
		fmt.Printf("Err checkUpdateFromServer: %s\n", err.Error())
		return
	}
	fmt.Printf("Found checkUPdateFromServer: %s\n",
		indexedCatalog.Version.SemVer)

	errs := indexedCatalog.Validate()
	for _, e := range errs {
		fmt.Printf("--> err: %s\n", e.Error())
	}
	UpdateSharedMatcher(indexedCatalog, cu.matcher)

	rc := cu.redisPool.Get()
	defer rc.Close()
	RedisUpdate(rc, indexedCatalog)
}

func (cu *CatalogUpdater) checkUpdateFromRedis() {

}

// updatesPoller keeps mkaing requests at the poll intervals,
// but also listend for a signal for an "on demand" update
// in case we want to trigger an update from an outside event
func (cu *CatalogUpdater) updatesPoller() {
	var redisC, serverC <-chan time.Time
	if cu.redisCheckSeconds > 0 {
		redisTicker := time.NewTicker(time.Duration(cu.redisCheckSeconds) * time.Second)
		redisC = redisTicker.C
	} else {
		redisC = make(chan time.Time)
	}

	if cu.serverCheckSeconds > 0 {
		serverTicker := time.NewTicker(time.Duration(cu.serverCheckSeconds) * time.Second)
		serverC = serverTicker.C
	} else {
		serverC = make(chan time.Time)
	}

	// TODO: Implement exponential backoff
	for {
		select {
		case <-cu.RequestShutdown: // stop the poller
			return
		case <-redisC:
			cu.checkUpdateFromRedis()
		case <-cu.RequestOnDemandUpdate: // forced update from server
			cu.checkUpdateFromServer()
		case <-serverC:
			cu.checkUpdateFromServer()
		}
	}
}

// LaunchUpdatesPoller returns a CatalogUpdater
func LaunchUpdatesPoller(redisPool *redis.Pool, updateBaseURL string,
	catalogApiKey string, matcher *pathmatcher.SharedPathMatcher,
	redisCheckSeconds int64, serverCheckSeconds int64) (*CatalogUpdater, error) {

	if serverCheckSeconds < redisCheckSeconds && serverCheckSeconds > 0 {
		// makes no sense to check the server more often than the server
		redisCheckSeconds = serverCheckSeconds
	}

	catalogUpdater := CatalogUpdater{
		urls: CatalogURLs{
			GetLatestVersion:  updateBaseURL + "/latest",
			GetDiff:           updateBaseURL + "/catalog_diff",
			GetIndexedCatalog: updateBaseURL + "/indexed_limits",
		},
		catalogAPIKey:         catalogApiKey,
		redisCheckSeconds:     redisCheckSeconds,
		serverCheckSeconds:    serverCheckSeconds,
		redisPool:             redisPool,
		matcher:               matcher,
		RequestOnDemandUpdate: make(chan bool),
		RequestShutdown:       make(chan bool),
	}

	go catalogUpdater.updatesPoller()

	return &catalogUpdater, nil
}
