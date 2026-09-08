package main

import (
	"embed"

	"__package__/catalogs"
	"__package__/handlers"
	"__package__/proto"
	"github.com/La-Perla-App/backend-core-laperla/pkg/config"
	"github.com/La-Perla-App/backend-core-laperla/pkg/server"
	"log"
)

var (
	//go:embed proto
	protoFilesFs embed.FS

	address = config.GetString("server.address")
)

func main() {
	server.InitEnvironment()

	srv := server.NewServer(
		server.WithProdMode(catalogs.IsProd),
		server.WithDebugRoute(catalogs.SpecialRoutes.DebugRoute),
		server.WithSwagger(catalogs.SpecialRoutes.SwaggerRoute, proto.SwaggerJsonDoc),
		server.WithServices(handlers.RegisterServicesFns),
		server.WithProtosDownload(protoFilesFs, catalogs.SpecialRoutes.ProtosDownload, "proto"),
	)

	log.Printf("Initializing gRPC server on address: %s\n", address)
	if err := srv.Listen(address); err != nil {
		log.Fatal(err)
	}
}
