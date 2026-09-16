# Agentes — backend-core-laperla

Librería compartida de los backends Go de La Perla. **No es un microservicio de producto.**

## Qué tocas aquí

`pkg/cache` (Redis), `pkg/search` (OpenSearch), `pkg/server`, `pkg/api` + JWT, `pkg/db/postgres`, `pkg/broker/nats`, `pkg/security`, `pkg/config`, CLI `protoc-gen-backend-core-laperla-cli`.

## Reglas

- **Redis:** solo este `pkg/cache`. Si un servicio necesita un comando nuevo, **extiende** `Cache` + `redis_impl.go` + wrappers en `provider.go`. No dejes que los servicios hablen `go-redis` directo.
- **OpenSearch:** solo `pkg/search`. Extiende `Searcher` + `client.go` + wrappers en `provider.go`. No importes `opensearch-go` en servicios de producto.
- **InitOpenSearch:** `opensearch.addrs` / `opensearch.url` o `OPENSEARCH_URL`; user/password desde config o `OPENSEARCH_USER` / `OPENSEARCH_PASSWORD`. HTTPS self-signed: `opensearch.insecureSkipVerify` o `OPENSEARCH_INSECURE_SKIP_VERIFY`.
- **Cluster:** multi-key solo con el mismo hash tag. Helpers: `HashTag`, `ClusterKey`, `SameSlot`. `IncrExpire` es Lua de **una** key. No uses `Keys()` en código de producción.
- **InitRedis:** `cache.redis.addrs` o `REDIS_ADDR`. `isCluster` / `db` desde config.
- **Auth JWT:** cookie `laperla-access-token`. Sin `tenant_id` en `SessionData`.
- **Sin tenant** en APIs nuevas. Multi-destino es `destination_id` en los servicios de producto.
- **Object store (`pkg/storage/minio`):** MinIO **y** AWS S3 vía el mismo cliente (`minio-go`). Config `storage.*` o legacy `minio.*`. Firmadas: `PresignedPut` / `PresignedGet`. Dos clientes solo si `publicHost` ≠ `endpoint` (MinIO cluster vs Ingress); en S3 suele bastar uno.
- **NATS `AddWorker`:** un callback que devuelve false reentrega **con espera** (`consumerConfig.BackOff`, o `DefaultRetryBackoff` si el consumidor no lo declara). Antes era `Nak()` a secas: reentregaba al instante y un consumidor con `MaxDeliver: 8` se comía los ocho intentos en menos de un segundo, justo mientras la dependencia caída seguía caída. Reintento inmediato a propósito: `BackOff: []time.Duration{0}`. `AddFanOutTaskWorker` no entra aquí: es tarea única con un reintento por diseño.
- Tag/release: los servicios consumen un **módulo publicado** (`v0.1.0+`). No asumas `replace` en repos remotos.
- Humo Redis (Telepresence): `go test -tags smoke ./pkg/cache/`
- Humo NATS (Telepresence): `SMOKE_NATS_URL=… go test -tags smoke ./pkg/broker/nats/` — crea y borra su propio stream; comprueba que una reentrega espera.
- No commitees secretos. Commits: La Perla, no marcas ajenas.

## Docs

README de este repo. Skills del workspace (si estás en `/home/jvillacorta/laperla`): `skills/laperla-data`, `skills/laperla-platform`.
