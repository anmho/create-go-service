//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"

	"github.com/anmho/create-go-service/internal/generator"
	"github.com/anmho/create-go-service/internal/generator/api"
	"github.com/anmho/create-go-service/internal/generator/config"
	"github.com/anmho/create-go-service/internal/generator/database"
	"github.com/anmho/create-go-service/internal/generator/deployment"
)

func main() {
	cfg := config.ProjectConfig{
		ProjectName: "test-grpc-service",
		ModulePath:  "github.com/test/grpc-service",
		OutputDir:   "/tmp/test-grpc-service",
		Features:    []config.Feature{},
		API: api.Config{
			Types: []api.Type{api.TypeGRPC},
		},
		Database: database.Config{
			Type: database.TypeDynamoDB,
		},
		Deployment: deployment.Config{
			Type: deployment.TypeFly,
		},
	}

	gen := generator.NewGenerator(cfg)
	if err := gen.Generate(); err != nil {
		log.Fatalf("Failed to generate project: %v", err)
	}

	fmt.Println("✓ gRPC project generated successfully at:", cfg.OutputDir)
}

