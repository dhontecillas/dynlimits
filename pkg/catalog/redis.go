package catalog

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dhontecillas/dynlimits/pkg/ratelimit"
	"github.com/gomodule/redigo/redis"
)

const (
	RedisKeyLimitsVersion          string = "dynlimits_limits_version"
	RedisKeyCatalogVersion         string = "dynlimits_catalog_version"
	RedisKeyUpdating               string = "dynlimits_updating"
	RedisKeyUpdatingLimitsVersion  string = "dynlimits_updating_limits_version"
	RedisKeyUpdatingCatalogVersion string = "dynlimits_updating_catalog_version"
	RedisKeyUpdateStarted          string = "dynlimits_update_started"
	RedisKeyUpdateFinished         string = "dynlimits_update_finished"

	UpdateTimeoutSeconds int64 = 5 * 60 // 5 min to update the catalog
)

/* RedisStartUpdate sets a redis key with the time starting to update
the redis catalog if it is not set (no other process is updatating the
catalog) returning the given timestamp. In case it cannot set the
timestamp it returns a nil time, and nil error.
*/
func RedisStartUpdate(conn redis.Conn, tm time.Time) (*time.Time, error) {
	// we have 5 minutes to complete a full update of the redis catalog
	res, err := conn.Do("SET", RedisKeyUpdating, tm.Format(time.RFC3339), "NX",
		"EX", UpdateTimeoutSeconds)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}
	// se set the start date, as the previous key behaves like a lock
	conn.Do("SET", RedisKeyUpdateStarted, tm.Unix())
	return &tm, nil
}

func RedisFinishUpdate(conn redis.Conn, started time.Time,
	catalogVersion string, limitsVersion string) {
	// check if we timedout
	now := time.Now()
	// we do not check for errors this data as is only for reference for now
	// TODO: review if we want to set the expiration date of RedisStartUpdate
	// based on how long took to perform the last update (something like:
	// last_update_time * 2 + 2 min for example), or if the last try to update
	// took to long (start is after end) just adjust
	if now.Sub(started) > time.Second*time.Duration(UpdateTimeoutSeconds-5) {
		// if there are only 5 secs left (should not happend), we just
		// let it expire.
		return
	}
	conn.Send("SET", RedisKeyUpdateFinished, now.Unix())
	conn.Send("SET", RedisKeyCatalogVersion, catalogVersion)
	conn.Send("SET", RedisKeyLimitsVersion, limitsVersion)
	conn.Send("DEL", RedisKeyUpdating)
	conn.Do("EXEC")
	// TODO: log the result of the DEL command, and the time it took
	// to complete all the update
}

func RedisUpdateLimits(conn redis.Conn, ail *APIIndexedLimits) {
	for _, akil := range ail.APILimits {
		apiKey := akil.APIKey
		for _, lim := range akil.Limits {
			if lim.EndpointIdx < 0 || lim.EndpointIdx >= len(ail.Endpoints) {
				// TODO: log malformed data
				continue
			}
			eidx := ail.Endpoints[lim.EndpointIdx]
			if eidx.MethodIdx < 0 || eidx.MethodIdx >= len(ail.Methods) ||
				eidx.PathIdx < 0 || eidx.PathIdx >= len(ail.Paths) {
				// TODO: log malformed data
				continue
			}
			mn := strings.ToUpper(ail.Methods[eidx.MethodIdx])
			key := fmt.Sprintf("%s_%s_%s", apiKey, mn,
				ail.Paths[eidx.PathIdx])
			fmt.Printf("set limit for %s to %d\n", key, lim.RateLimit)
			// TODO: we can optimize this by checking if the ratelimit
			// has changed.
			ratelimit.SetRedisRateLimit(conn, key, lim.RateLimit)
		}
	}
}

// RedisUpdate loads a catalog of per endpoint and api key
// into Redis
func RedisUpdate(conn redis.Conn, ail *APIIndexedLimits) {
	// TODO: Put the catalog version into the APIIndexedLimits
	// TODO: Split the catalog vs limits versions
	n := time.Now()
	updateTime, err := RedisStartUpdate(conn, n)
	if err != nil {
		fmt.Printf("cannot start redis update: %s\n", err.Error())
		return
	}
	if updateTime == nil {
		// another process already started updateing the catalog
		return
	}

	RedisUpdateLimits(conn, ail)

	// TODO: Put here the catalog version
	RedisFinishUpdate(conn, n, "v0.0.1", ail.Version.HashVer)
}

type RedisCatalogStatus struct {
	LimitsVersion  string
	CatalogVersion string
	UpdateStarted  time.Time
	UpdateFinished time.Time
}

// RedisGetVersions returns the catalogVersion, and limitsVersions
func RedisGetCatalogStatus(conn redis.Conn) (*RedisCatalogStatus, error) {
	var rcs RedisCatalogStatus
	if err := conn.Send("GET", RedisKeyLimitsVersion); err != nil {
		return nil, err
	}
	if err := conn.Send("GET", RedisKeyUpdatingCatalogVersion); err != nil {
		return nil, err
	}
	if err := conn.Send("GET", RedisKeyUpdateStarted); err != nil {
		return nil, err
	}
	if err := conn.Send("GET", RedisKeyUpdateFinished); err != nil {
		return nil, err
	}

	res, err := redis.Values(conn.Do("EXEC"))
	if err != nil {
		return nil, err
	}

	if res[0] != nil {
		bs, ok := res[0].([]byte)
		if ok {
			rcs.LimitsVersion = string(bs)
		}
	}

	if res[1] != nil {
		bs, ok := res[1].([]byte)
		if ok {
			rcs.CatalogVersion = string(bs)
		}
	}

	if res[2] != nil {
		bs, ok := res[2].([]byte)
		if ok {
			ts, err := strconv.ParseInt(string(bs), 10, 64)
			if err != nil {
				rcs.UpdateStarted = time.Unix(ts, 0)
			}
		}
	}

	if res[3] != nil {
		bs, ok := res[3].([]byte)
		if ok {
			ts, err := strconv.ParseInt(string(bs), 10, 64)
			if err != nil {
				rcs.UpdateFinished = time.Unix(ts, 0)
			}
		}
	}

	return &rcs, nil
}
