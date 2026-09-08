package server

import (
	"io/fs"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ServerOption = func(server *Server)

func WithProdMode(prodMode bool) ServerOption {
	return func(server *Server) {
		server.prodMode = prodMode
	}
}

func WithDebugRoute(path string) ServerOption {
	return func(server *Server) {
		server.debugRoute = path
	}
}

func WithSwagger(path string, openApiSchema []byte) ServerOption {
	return func(server *Server) {
		server.swaggerRoute = path
		server.openAPISchema = openApiSchema
	}
}

func WithGatewayPrefixes(gatewayPrefixes ...string) ServerOption {
	return func(server *Server) {
		server.gatewayPrefixes = gatewayPrefixes
	}
}

func WithProtosDownload(protosFs fs.FS, httpPath, protosBasepath string) ServerOption {
	return func(server *Server) {
		zipFile, err := createProtoZipFile(protosFs, protosBasepath)
		if err != nil {
			log.Fatal(err)
		}
		server.protosZip = zipFile
		server.protosZipRoute = httpPath
	}
}

func WithServices(servicesFns []RegisterServiceFn) ServerOption {
	return func(server *Server) {
		server.registerServiceFns = servicesFns
	}
}

func WithRouterGroup(routerGroup func(r chi.Router)) ServerOption {
	return func(server *Server) {
		server.routerGroup = routerGroup
	}
}

func WithChiMiddlewares(middlewares ...func(http.Handler) http.Handler) ServerOption {
	return func(server *Server) {
		server.chiMiddlewares = middlewares
	}
}
