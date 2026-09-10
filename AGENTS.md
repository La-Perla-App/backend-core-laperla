# Agentes — backend-core-laperla

Librería compartida de los backends Go de La Perla. **No es un microservicio de producto.**

## Qué tocas aquí

`pkg/cache` (Redis), `pkg/server`, `pkg/api` + JWT, `pkg/db/postgres`, `pkg/broker/nats`, `pkg/security`, `pkg/config`, CLI `protoc-gen-backend-core-laperla-cli`.

## Reglas

- **Redis:** solo este `pkg/cache`. Si un servicio necesita un comando nuevo, **extiende** `Cache` + `redis_impl.go` + wrappers en `provider.go`. No dejes que los servicios hablen `go-redis` directo.
- **Cluster:** multi-key solo con el mismo hash tag. Helpers: `HashTag`, `ClusterKey`, `SameSlot`. `IncrExpire` es Lua de **una** key. No uses `Keys()` en código de producción.
- **InitRedis:** `cache.redis.addrs` o `REDIS_ADDR`. `isCluster` / `db` desde config.
- **Auth JWT:** cookie `laperla-access-token`. Sin `tenant_id` en `SessionData`.
- **Sin tenant** en APIs nuevas. Multi-destino es `destination_id` en los servicios de producto.
- Tag/release: los servicios consumen un **módulo publicado** (`v0.1.0+`). No asumas `replace` en repos remotos.
- Humo Redis (Telepresence): `go test -tags smoke ./pkg/cache/`
- No commitees secretos. Commits: La Perla, no marcas ajenas.

## Docs

README de este repo. Skills del workspace (si estás en `/home/jvillacorta/laperla`): `skills/laperla-data`, `skills/laperla-platform`.
