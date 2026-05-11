#  pg_rewind

Stolon can use [pg_rewind](http://www.postgresql.org/docs/current/static/app-pgrewind.html) to speedup instance resynchronization (for example resyncing an old master or a slave ahead of the current master) without the need to copy all the new master data.

## Enabling

It can be enabled setting to true the cluster specification option `usePgrewind` (defaults to false):

``` bash
stolonctl [cluster options] update --patch '{ "usePgrewind" : true }'
```

This will also enable the `wal_log_hints` postgresql parameter. If previously `wal_log_hints` wasn't enabled you should restart the postgresql instances (you can do so restarting the `stolon-keeper`)

pg_rewind needs to connect to the master database with a superuser role (see the [Stolon Architecture and Requirements](architecture.md)).

## Retries

When `usePgrewind` is enabled, a keeper that must **resync** normally runs `pg_rewind` once; if it fails, Stolon falls back to a full copy using `pg_basebackup`. You can retry `pg_rewind` a configurable number of times before that fallback by setting `pgRewindRetry` on the cluster specification (see [cluster_spec.md](cluster_spec.md)).

**Default (omitted `pgRewindRetry`):** one `pg_rewind` attempt only (`maxAttempts` 1), no waits between attempts (`interval` 0), backoff multiplier 1. This matches the historical behaviour of trying `pg_rewind` once then `pg_basebackup`.

Fields:

* **maxAttempts** — total `pg_rewind` runs per resync, including the first; default **1** (no retries).
* **interval** — base sleep after a failed run, before the next attempt; default **0s** (no delay; meaningful when `maxAttempts` is greater than 1).
* **backoffMultiplier** — each wait is multiplied by this factor relative to the previous wait’s base slot; must be ≥ **1**; default **1** (same `interval` between retries).

Example: up to 4 attempts, 5s before the second run, then 10s, then 20s (`backoffMultiplier` 2):

```bash
stolonctl [cluster options] update --patch '{
  "usePgrewind": true,
  "pgRewindRetry": {
    "maxAttempts": 4,
    "interval": "5s",
    "backoffMultiplier": 2
  }
}'
```
