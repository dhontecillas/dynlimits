package catalog

/*
The API Catalog has a fast lookup for API keys that are limited at
least until a well known time.

Should act as a local caching limit, so we do not need to hit
the redis cache
*/

import (
	"time"
)

// TODO: Create Default Limits, to apply just in case some API key
// misses its endpoints

type APILimits struct {
	RateLimitsKeyPrefix string
	BlockedUntil        time.Time
}

type APIKeys interface {
	GetLimits(apiKey string, method string, endpoint string) APILimits
	BlockUntil(apiKey string, until time.Time)
}

/*

type RedisBackedAPIKeys struct {

}

func NewRedisBackedAPIKeys() {

}

func (rbak *RedisBackedAPIKeys) GetLimitsFromComposed() {

}

*/

type DefaultAPIKeys struct {
}

func NewDefaultAPIKeys() *DefaultAPIKeys {
	return &DefaultAPIKeys{}
}

func (dak *DefaultAPIKeys) GetLimits(apiKey string, method string, path string) APILimits {
	return APILimits{
		RateLimitsKeyPrefix: "tt",
	}
}

func (dak *DefaultAPIKeys) BlockUntil(apiKey string, until time.Time) {
}

func (dak *DefaultAPIKeys) toPrefix(apiKey string, method string, pathDef string) {
}

func (dak *DefaultAPIKeys) SetLimit(apiKey string, method string, pathDef string) {
}
