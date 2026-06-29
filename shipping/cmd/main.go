package main

import (
	"log"

	"github.com/sheilallee/microservices/shipping/config"
	"github.com/sheilallee/microservices/shipping/internal/adapters/db"
	grpc_adapter "github.com/sheilallee/microservices/shipping/internal/adapters/grpc"
	"github.com/sheilallee/microservices/shipping/internal/application/core/api"
)

func main() {
	dbAdapter, err := db.NewAdapter(config.GetDataSourceURL())
	if err != nil {
		log.Fatalf("Failed to connect to database. Error: %v", err)
	}

	application := api.NewApplication(dbAdapter)
	grpcAdapter := grpc_adapter.NewAdapter(application, config.GetApplicationPort())
	grpcAdapter.Run()
}
