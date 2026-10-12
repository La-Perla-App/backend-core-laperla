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
- **El `message` de un error sale de su código, no se supone.** Si nadie lo
  fijó (`UpdateResponseInfo*`), el interceptor usa `api.ErrorMessageFor`: el
  nombre del catálogo (`invalid_request_data`, `not_found`…) o el del código
  Connect (`permission_denied`…). Lo que no llega a un handler (ruta
  inexistente, cuerpo ilegible) lo nombra el middleware REST por el estado
  HTTP. Antes todo salía `internal_server_error`, también un 400 por un
  parámetro que faltaba. Sólo afecta al envoltorio REST: por Connect,
  `Response-Info` no viaja en los errores y el cliente recibe el mensaje del
  handler tal cual.
- No commitees secretos. Commits: La Perla, no marcas ajenas.

## Connect v1, y por qué todavía no v2

**El plugin `buf.build/connectrpc/go` va fijado a `v1.18.1` en los nueve
repos de servicio.** Sin el pin, el remoto genera para
`connectrpc.com/connect/v2` y nada compila. No lo quites sin leer esto.

v2 no es un renombrado, es otra arquitectura:

| | v1 (lo que usamos) | v2 |
|---|---|---|
| Registrar | `NewXHandler(svc, opts) → (string, http.Handler)` | `RegisterXHandler(*connect.Server, svc)` |
| Firma del RPC | `func(ctx, *connect.Request[T]) (*connect.Response[R], error)` | `func(ctx, *T) (*R, error)` |
| Servir | un `http.Handler` por servicio | `connecthttp.Mount(mux, server)` |

El paquete `connectrpc.com/connect/v2` **no tiene** `Request`, `Response` ni
`Handler`: es sólo cliente y transporte.

**Lo que bloquea la migración es `pkg/server/services.go`.** Ahí se monta un
transcodificador de vanguard sobre todos los servicios, y de él sale la
superficie REST entera (`/api/auth`, `/api/directory`, `/api/benefits`…), la
que publica el Ingress y consume la app. Vanguard necesita **un `http.Handler`
por servicio**, y en v2 eso no existe. Vanguard se quedó en **v0.4.0**: no hay
v2 ni prerelease, y connect v2 no trae transcodificación REST propia — no
entiende `google.api.http`. Migrar hoy deja los `/api/...` en 404 sin
sustituto.

Y hay un segundo coste que no es menor: al desaparecer `connect.Request[T]`
desaparece `req.Header()`, que es de donde sale la sesión en **cada RPC de los
diez servicios**. No es un ajuste de firmas, es reescribir la autorización.

**Cuándo retomarlo:** cuando vanguard publique una versión compatible con
connect v2. Entonces el orden es core (`pkg/api`, `pkg/server`,
`interceptors`, `middlewares`, y el plugin
`protoc-gen-backend-core-laperla-cli` que genera los clientes) y después los
servicios. La alternativa —renunciar a REST y dejar sólo Connect— no es una
decisión técnica: afecta a la app y a cualquier integración futura.

Comprobado el 08/10/2026 generando con el plugin de v2 y leyendo el módulo.

## Docs

README de este repo. Skills del workspace (si estás en `/home/jvillacorta/laperla`): `skills/laperla-data`, `skills/laperla-platform`.
