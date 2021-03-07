# TODO List

- Try to load catalog first from catalog server, and fallback
    to catalog file if not successful, or catalog server
    address is empty

- Add `shutdown` chan to the CatalogUpdatesPoller

- Add configuration option for the API header (currently
    hardcoded as `X-Api-Key`)

- Implement exponential backof int `pkg/catalog/updater.go`'s
    updatesPoller

- Implement `checkUpdateFromRedis`: we can fetch the current
    catalog from redis. This way, a sin


