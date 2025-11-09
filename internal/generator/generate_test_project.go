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
		ProjectName: "test-chi-service",
		ModulePath:  "github.com/test/chi-service",
		OutputDir:   "/tmp/test-chi-service",
		Features:    []config.Feature{config.FeatureAuth},
		API: api.Config{
			Types: []api.Type{api.TypeChi},
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

	fmt.Println("✓ Project generated successfully at:", cfg.OutputDir)
}
