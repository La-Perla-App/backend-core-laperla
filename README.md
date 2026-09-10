# La Perla Backend Core

Framework/base para crear backends en Go con infraestructura lista.

## Características

- **API Server** - HTTP/gRPC con chi router y connectrpc
- **Bases de datos** - PostgreSQL (sqlx) y Cassandra (gocql)
- **Cache** - Redis
- **Messaging** - NATS
- **Auth** - JWT
- **CLI** - Herramienta para generar código desde proto

## Estructura del Proyecto Core

```
backend-core-laperla/
├── cmd/                           # Punto de entrada y CLI
│   ├── backend-core-laperla-cli/   # CLI para generar código
│   └── protoc-gen-.../            # Generador de código protobuf
├── pkg/
│   ├── api/                       # Manejo de requests, auth, helpers
│   ├── broker/nats/               # Cliente NATS
│   ├── cache/                     # Redis cache
│   ├── config/                    # Configuración centralizada
│   ├── db/                        # PostgreSQL y Cassandra
│   ├── events/                    # Dispatcher de eventos
│   ├── jsonparser/                # Parsing JSON/YAML eficiente
│   ├── modules/                   # Módulos de acciones
│   ├── security/                  # Keys y seguridad
│   ├── server/                    # Servidor HTTP, middlewares
│   └── utils/                     # Utilidades
```

## Uso - Crear un Microservicio

### 1. Instalar CLI

```bash
go install github.com/La-Perla-App/backend-core-laperla/cmd/backend-core-laperla-cli@main
```

### 2. Inicializar Proyecto

```bash
# Crear nuevo microservicio
backend-core-laperla-cli init my-service -o ./my-service -p my-service

# Ir al directorio
cd my-service

# Inicializar go module
go mod init my-service

# Descargar dependencias protobuf
buf dep update

# Generar código desde proto
buf generate

# Descargar dependencias Go
go mod tidy
```

Esto genera la siguiente estructura:

```
my-service/
├── main.go              # Punto de entrada
├── catalogs/
│   └── catalogs.go      # Configuraciones del servicio
├── handlers/
│   ├── handlers.go      # Registro de handlers
│   └── example/
│       └── v1/
│           ├── handler.go
│           └── register.go
├── proto/
│   ├── services/
│   │   └── example/
│   │       └── v1/
│   │           ├── service.proto
│   │           └── types.proto
│   └── generated/       # Código generado
├── buf.yaml
└── buf.gen.yaml
```

### 3. Definir Proto

Edita `proto/services/<service>/v1/service.proto`:

```protobuf
syntax = "proto3";

package services.example.v1;

import "google/api/annotations.proto";
import "services/example/v1/types.proto";

service ExampleService {
  rpc GetUser(GetUserRequest) returns (GetUserResponse) {
    option (google.api.http) = {
      get: "/api/example/v1/users/{id}"
    };
  }
  
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse) {
    option (google.api.http) = {
      post: "/api/example/v1/users"
      body: "*"
    };
  }
}
```

Edita `proto/services/<service>/v1/types.proto`:

```protobuf
syntax = "proto3";

package services.example.v1;

message GetUserRequest {
  string id = 1;
}

message GetUserResponse {
  string id = 1;
  string name = 2;
  string email = 3;
}
```

### 4. Regenerar Código

```bash
buf dep update
buf generate
```

### 5. Implementar Handler

Edita `handlers/<service>/v1/handler.go`:

```go
package examplev1handler

import (
    "context"
    "buf.build/go/protovalidate"
    "connectrpc.com/connect"
    examplev1 "my-service/proto/generated/services/example/v1"
    "my-service/proto/generated/services/example/v1/examplev1connect"
)

var handler examplev1connect.ExampleServiceHandler = &handlerImpl{}

type handlerImpl{}

func (h *handlerImpl) GetUser(ctx context.Context, req *connect.Request[examplev1.GetUserRequest]) (*connect.Response[examplev1.GetUserResponse], error) {
    if err := protovalidate.Validate(req.Msg); err != nil {
        return nil, connect.NewError(connect.CodeInvalidArgument, err)
    }

    response := &examplev1.GetUserResponse{
        Id:    req.Msg.Id,
        Name:  "John Doe",
        Email: "john@example.com",
    }

    return connect.NewResponse(response), nil
}
```

### 6. Registrar Handler

Edita `handlers/handlers.go`:

```go
package handlers

import (
    examplev1handler "my-service/handlers/example/v1"
    "github.com/La-Perla-App/backend-core-laperla/pkg/server"
)

var RegisterServicesFns = []server.RegisterServiceFn{
    examplev1handler.RegisterServiceHandler,
}
```

### 7. Crear Nuevos Handlers

```bash
# Crear nuevo handler
backend-core-laperla-cli create handler user -v 1 -o .

# Regenerar proto
buf dep update
buf generate
go mod tidy
```

## Configuración

### Acceso a Config

```go
import "github.com/La-Perla-App/backend-core-laperla/pkg/config"

// Acceso directo a valores
serverAddr := config.GetString("server.address")
dbConnStr := config.GetString("database.connectionString")
redisAddrs := config.GetArrayStrings("cache.redis.addrs")
```

### Archivo de Configuración

El proyecto soporta JSON y YAML (`config.FileName` y `config.FilePath`):

```yaml
# config.yaml
server:
  address: "0.0.0.0:8080"

database:
  connectionString: "postgres://user:password@localhost:5432/mydb"
  maxIdleConnections: 10
  maxOpenConnections: 100

cassandra:
  hosts:
    - "localhost"
  keyspace: "mykeyspace"

cache:
  redis:
    addrs:
      - "localhost:6379"
    password: ""
    isCluster: false

nats:
  addrs:
    - "nats://localhost:4222"

security:
  rsa:
    privateKey: ""
    publicKey: ""

grpc:
  clientAddresses:
    catalogService: "localhost:50051"
```

Para especificar el archivo de configuración usa:
- Variable `config.FileName` y `config.FilePath`
- Flag `--config-file=path/to/config.yaml`

## Inicialización de Servicios

El core provee funciones para inicializar los servicios automáticamente desde el config:

```go
package main

import (
    "github.com/La-Perla-App/backend-core-laperla/pkg/config"
    "github.com/La-Perla-App/backend-core-laperla/pkg/server"
)

func main() {
    server.InitEnvironment()
    server.InitRedis()
    server.InitNats()

    addr := config.GetString("server.address")

    srv := server.NewServer(
        server.WithProdMode(catalogs.IsProd),
        server.WithDebugRoute(catalogs.SpecialRoutes.DebugRoute),
        server.WithSwagger(catalogs.SpecialRoutes.SwaggerRoute, proto.SwaggerJsonDoc),
        server.WithServices(handlers.RegisterServicesFns),
        server.WithProtosDownload(protoFilesFs, catalogs.SpecialRoutes.ProtosDownload, "proto"),
    )

    srv.Listen(addr)
}
```

## Rutas Automáticas

- `/health` - Health check (compat)
- `/healthz` - Liveness (Helm / Kubernetes)
- `/readyz` - Readiness (Helm / Kubernetes)
- `/api/<service>/debug` - Profiling (development only)
- `/api/<service>/swagger` - Swagger UI (development only)
- `/api/<service>/protos_download` - Descargar protos (development only)

## Requisitos

- Go 1.23+
- buf CLI (`go install github.com/bufbuild/buf/cmd/buf@latest`)
- PostgreSQL (opcional)
- Cassandra (opcional)
- Redis (opcional)
- NATS (opcional)

## Estructura de un Microservicio

```
my-service/
├── main.go
├── catalogs/
│   └── catalogs.go       # Config del servicio
├── handlers/
│   ├── handlers.go       # Registro de servicios
│   ├── example/
│   │   └── v1/
│   │       ├── handler.go
│   │       ├── handler_test.go
│   │       └── register.go
│   └── user/
│       └── v1/
├── proto/
│   ├── services/
│   │   ├── example/
│   │   │   └── v1/
│   │   │       ├── service.proto
│   │   │       └── types.proto
│   │   └── user/
│   │       └── v1/
│   ├── generated/
│   ├── embed.go
│   ├── buf.yaml
│   └── buf.gen.yaml
├── buf.lock
├── go.mod
├── go.sum
├── Dockerfile
└── config.yaml
```

## Licencia

Proprietary - La Perla
