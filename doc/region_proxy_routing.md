# Region-aware proxy routing

Multi-region deployments can expose PostgreSQL on **two addresses**: a **public** address (NAT/load balancer) used across regions and an optional **internal** address on a private network used when the **stolon-proxy** runs in the **same region** as the primary **keeper**.

Stolon compares an opaque **region** string on the proxy (`--region`) with the primary keeper’s region from cluster data (`Keeper.Status.Region`, sourced from keeper `--region`). When they match (both non-empty and equal) **and** an internal advertise address is published in `DB.Status`, the proxy connects via **internal** host/port; otherwise it uses the usual advertised address (`ListenAddress` / `Port`), unchanged from older releases.

## Operational checklist (dual-homed PostgreSQL)

1. Configure Postgres to listen on the interfaces that serve both paths (`listen_addresses`, `pg_hba.conf`). Validate connectivity on **both** addresses before relying on routing.
2. On each keeper, set **`--pg-advertise-address`** / **`--pg-advertise-port`** to what other regions and clients should use (external path).
3. Optionally set **`--pg-advertise-internal-address`** and, if needed, **`--pg-advertise-internal-port`** for the private path. If the internal port is omitted, the advertised external port is reused for internal routing decisions.
4. Set **`--region`** on keepers and proxies to the **same string** within a region (e.g. `eu-west-1`).
5. Deploy **sentinel** and **proxy** binaries that understand the new cluster-data fields before relying on internal routing (see upgrade order below).

## Flags summary

### stolon-keeper

| Flag | Meaning |
|------|---------|
| `--region string` | Opaque region id stored in keeper info and mirrored into cluster data. |
| `--pg-advertise-internal-address string` | Private network host/IP advertised for same-region proxies. |
| `--pg-advertise-internal-port string` | Optional; defaults to the effective advertised external port when unset. |

Environment (via existing `STKEEPER_*` convention): e.g. `STKEEPER_REGION`, `STKEEPER_PG_ADVERTISE_INTERNAL_ADDRESS`, `STKEEPER_PG_ADVERTISE_INTERNAL_PORT`.

### stolon-proxy

| Flag | Meaning |
|------|---------|
| `--region string` | Region for this proxy; compared to the master keeper’s region to choose internal vs external endpoint. |

Environment: `STPROXY_REGION`.

## Upgrade order

Rolling upgrades are backward compatible when new fields are unset:

1. Upgrade **sentinel** first so it persists `Region` and internal listen fields into cluster data when present.
2. Upgrade **keepers** and configure `--region` / internal advertise as needed.
3. Upgrade **proxies** and set `--region` where internal routing should apply.

Older proxies ignore unknown JSON fields on read and simply use external addresses until upgraded.

## Backward compatibility

| Setting | Behavior |
|---------|----------|
| Proxy `--region` empty | Always use external `ListenAddress` / `Port` (legacy behavior). |
| Keeper `--region` empty | Master keeper region empty → internal path never selected. |
| Internal listen address empty | Internal path disabled; external used. |
| Region mismatch or failover to another region | Proxy uses external path automatically. |

No changes to **cluster spec** are required for region routing.
