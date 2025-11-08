package generator

import (
	"fmt"
	"path/filepath"

	"github.com/anmho/create-go-service/internal/generator/api"
	"github.com/anmho/create-go-service/internal/generator/config"
	"github.com/anmho/create-go-service/internal/generator/database"
	"github.com/anmho/create-go-service/internal/generator/deployment"
)

type ProjectConfig = config.ProjectConfig

type Generator struct {
	config         ProjectConfig
	fs             FileSystem
	templateLoader TemplateLoader
}

// NewGenerator creates a new generator with default dependencies
func NewGenerator(config ProjectConfig) *Generator {
	return &Generator{
		config:         config,
		fs:             &OSFileSystem{},
		templateLoader: NewEmbeddedTemplateLoader(),
	}
}

// NewGeneratorWithDeps creates a new generator with custom dependencies (for testing)
func NewGeneratorWithDeps(config ProjectConfig, fs FileSystem, loader TemplateLoader) *Generator {
	return &Generator{
		config:         config,
		fs:             fs,
		templateLoader: loader,
	}
}

func (g *Generator) Generate() error {
	// Create output directory
	if err := g.fs.MkdirAll(g.config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create directory structure
	if err := g.createDirectoryStructure(); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}

	// Generate base files
	if err := g.generateBaseFiles(); err != nil {
		return fmt.Errorf("failed to generate base files: %w", err)
	}

	// Generate API files based on selected API types
	for _, apiType := range g.config.API.Types {
		if err := g.generateAPIFiles(apiType); err != nil {
			return fmt.Errorf("failed to generate %s files: %w", apiType, err)
		}
	}

	// Generate database files
	if err := g.generateDatabaseFiles(); err != nil {
		return fmt.Errorf("failed to generate database files: %w", err)
	}

	// Generate posts feature (domain layer)
	if err := g.generatePostsFiles(); err != nil {
		return fmt.Errorf("failed to generate posts files: %w", err)
	}

	// Generate feature files
	if err := g.generateFeatureFiles(); err != nil {
		return fmt.Errorf("failed to generate feature files: %w", err)
	}

	data := g.getTemplateData()

	// Generate main.go
	if err := g.generateFile("cmd/api/main.go", "base/main.go.tmpl", data); err != nil {
		return fmt.Errorf("failed to generate main.go: %w", err)
	}

	// Generate config files
	if err := g.generateConfigFiles(); err != nil {
		return fmt.Errorf("failed to generate config files: %w", err)
	}

	// Generate deployment files
	if err := g.generateDeploymentFiles(); err != nil {
		return fmt.Errorf("failed to generate deployment files: %w", err)
	}

	// Generate development files
	if err := g.generateDevelopmentFiles(); err != nil {
		return fmt.Errorf("failed to generate development files: %w", err)
	}

	// Generate Terraform files (if DynamoDB)
	if g.config.Database.Type == database.TypeDynamoDB {
		if err := g.generateTerraformFiles(); err != nil {
			return fmt.Errorf("failed to generate terraform files: %w", err)
		}
	}

	if g.config.Database.Type == database.TypePostgres {
		if err := g.generateAtlasFiles(); err != nil {
			return fmt.Errorf("failed to generate atlas files: %w", err)
		}
	}

	// Generate CLI files
	if err := g.generateCLIFiles(); err != nil {
		return fmt.Errorf("failed to generate CLI files: %w", err)
	}

	return nil
}

func (g *Generator) createDirectoryStructure() error {
	dirs := []string{
		"cmd/api",
		"cmd/postctl",
		"internal/config",
		"internal/cli",
		"internal/database",
		"internal/posts",
		"internal/metrics",
		"internal/auth",
	}

	// Add feature-specific directories
	for _, feature := range g.config.Features {
		switch feature {
		case config.FeaturePostHog:
			dirs = append(dirs, "internal/posthog")
		}
	}

	// Add API-specific directories
	for _, apiType := range g.config.API.Types {
		if apiType == api.TypeChi {
			dirs = append(dirs, "internal/api", "internal/json")
		} else if apiType == api.TypeGRPC {
			dirs = append(dirs, "internal/api", "protos/posts/v1")
		}
	}

	// Add terraform directory if using DynamoDB
	if g.config.Database.Type == database.TypeDynamoDB {
		dirs = append(dirs, "terraform")
	}

	// Add migrations directory if using PostgreSQL
	if g.config.Database.Type == database.TypePostgres {
		dirs = append(dirs, "migrations")
	}


	for _, dir := range dirs {
		path := filepath.Join(g.config.OutputDir, dir)
		if err := g.fs.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

func (g *Generator) generateBaseFiles() error {
	data := g.getTemplateData()

	// Generate go.mod
	if err := g.generateFile("go.mod", "base/go.mod.tmpl", data); err != nil {
		return err
	}

	// Generate README.md
	if err := g.generateFile("README.md", "base/README.md.tmpl", data); err != nil {
		return err
	}

	// Generate .gitignore
	if err := g.generateFile(".gitignore", "base/.gitignore.tmpl", data); err != nil {
		return err
	}

	// Generate .dockerignore
	if err := g.generateFile(".dockerignore", "base/.dockerignore.tmpl", data); err != nil {
		return err
	}

	// Generate stage-specific YAML config files
	if err := g.generateFile("local.yaml", "base/local.yaml.tmpl", data); err != nil {
		return err
	}
	if err := g.generateFile("development.yaml", "base/development.yaml.tmpl", data); err != nil {
		return err
	}
	if err := g.generateFile("production.yaml", "base/production.yaml.tmpl", data); err != nil {
		return err
	}

	// Generate Makefile
	if err := g.generateFile("Makefile", "makefile/Makefile.tmpl", data); err != nil {
		return err
	}

	return nil
}

func (g *Generator) generateAPIFiles(apiType api.Type) error {
	data := g.getTemplateData()

	switch apiType {
	case api.TypeChi:
		// Generate Chi server files
		if err := g.generateFile("internal/api/server.go", "chi/server.go.tmpl", data); err != nil {
			return err
		}
		// Generate JSON utilities (shared package to avoid import cycles)
		if err := g.generateFile("internal/json/json.go", "chi/json.go.tmpl", data); err != nil {
			return err
		}
	case api.TypeGRPC:
		// Generate gRPC/ConnectRPC server files
		if err := g.generateFile("internal/api/server.go", "grpc/server.go.tmpl", data); err != nil {
			return err
		}
		if err := g.generateFile("internal/api/posts_handler.go", "grpc/posts_handler.go.tmpl", data); err != nil {
			return err
		}
		// Generate proto files and buf config
		if err := g.generateFile("protos/posts/v1/posts.proto", "grpc/posts.proto.tmpl", data); err != nil {
			return err
		}
		if err := g.generateFile("buf.yaml", "grpc/buf.yaml.tmpl", data); err != nil {
			return err
		}
		if err := g.generateFile("buf.gen.yaml", "grpc/buf.gen.yaml.tmpl", data); err != nil {
			return err
		}
		// Generate gRPC-specific main.go
		if err := g.generateFile("cmd/api/main.go", "grpc/main.go.tmpl", data); err != nil {
			return err
		}
	case api.TypeHuma:
		// Huma not implemented yet - skip for now
		return nil
	}

	return nil
}

func (g *Generator) generateDatabaseFiles() error {
	data := g.getTemplateData()

	switch g.config.Database.Type {
	case database.TypeDynamoDB:
		if err := g.generateFile("internal/database/dynamodb.go", "dynamodb/dynamodb.go.tmpl", data); err != nil {
			return err
		}
	case database.TypePostgres:
		if err := g.generateFile("internal/database/postgres.go", "postgres/postgres.go.tmpl", data); err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) generatePostsFiles() error {
	data := g.getTemplateData()

	// Generate posts domain files
	if err := g.generateFile("internal/posts/post.go", "posts/post.go.tmpl", data); err != nil {
		return err
	}
	if err := g.generateFile("internal/posts/service.go", "posts/service.go.tmpl", data); err != nil {
		return err
	}

	// Generate database-specific repository
	switch g.config.Database.Type {
	case database.TypeDynamoDB:
		if err := g.generateFile("internal/posts/dynamodb_table.go", "posts/dynamodb_table.go.tmpl", data); err != nil {
			return err
		}
		// Generate tests for DynamoDB repository (outputs to post_table_test.go)
		if err := g.generateFile("internal/posts/post_table_test.go", "posts/dynamodb_table_test.go.tmpl", data); err != nil {
			return err
		}
	case database.TypePostgres:
		if err := g.generateFile("internal/posts/postgres_table.go", "posts/postgres_table.go.tmpl", data); err != nil {
			return err
		}
		// Generate tests for PostgreSQL repository (outputs to post_table_test.go)
		if err := g.generateFile("internal/posts/post_table_test.go", "posts/postgres_table_test.go.tmpl", data); err != nil {
			return err
		}
	}

	// Generate API-specific handlers (folder-by-feature: handlers stay in posts package)
	for _, apiType := range g.config.API.Types {
		if apiType == api.TypeChi {
			if err := g.generateFile("internal/posts/handlers.go", "posts/handlers.go.tmpl", data); err != nil {
				return err
			}
		}
		// gRPC handlers are generated in generateAPIFiles
	}

	return nil
}

func (g *Generator) generateAtlasFiles() error {
	data := g.getTemplateData()

	// Generate Atlas configuration file (reproducible)
	if err := g.generateFile("atlas.hcl", "atlas/atlas.hcl.tmpl", data); err != nil {
		return err
	}

	// Generate initial migration files
	if err := g.generateFile("migrations/001_initial.up.sql", "atlas/migrations/001_initial.up.sql.tmpl", data); err != nil {
		return err
	}
	if err := g.generateFile("migrations/001_initial.down.sql", "atlas/migrations/001_initial.down.sql.tmpl", data); err != nil {
		return err
	}

	return nil
}

func (g *Generator) generateFeatureFiles() error {
	data := g.getTemplateData()

	// Always generate metrics files (mandatory)
	if err := g.generateFile("internal/metrics/metrics.go", "metrics/metrics.go.tmpl", data); err != nil {
		return err
	}

	// Generate feature files if enabled
	for _, feature := range g.config.Features {
		switch feature {
		case config.FeatureAuth:
			if err := g.generateFile("internal/auth/jwt.go", "auth/jwt.go.tmpl", data); err != nil {
				return err
			}
		case config.FeaturePostHog:
			if err := g.generateFile("internal/posthog/posthog.go", "posthog/posthog.go.tmpl", data); err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *Generator) generateTerraformFiles() error {
	data := g.getTemplateData()

	// Generate Terraform configuration files
	if err := g.generateFile("terraform/main.tf", "terraform/main.tf.tmpl", data); err != nil {
		return err
	}
	if err := g.generateFile("terraform/variables.tf", "terraform/variables.tf.tmpl", data); err != nil {
		return err
	}
	if err := g.generateFile("terraform/.gitignore", "terraform/.gitignore.tmpl", data); err != nil {
		return err
	}
	if err := g.generateFile("terraform/README.md", "terraform/README.md.tmpl", data); err != nil {
		return err
	}

	return nil
}

func (g *Generator) generateCLIFiles() error {
	data := g.getTemplateData()

	// Generate CLI main entry point
	if err := g.generateFile("cmd/postctl/main.go", "cli/main.go.tmpl", data); err != nil {
		return err
	}

	// Generate CLI command files
	if err := g.generateFile("internal/cli/root.go", "cli/root.go.tmpl", data); err != nil {
		return err
	}
	if err := g.generateFile("internal/cli/server.go", "cli/server.go.tmpl", data); err != nil {
		return err
	}
	if err := g.generateFile("internal/cli/posts.go", "cli/posts.go.tmpl", data); err != nil {
		return err
	}
	if err := g.generateFile("internal/cli/version.go", "cli/version.go.tmpl", data); err != nil {
		return err
	}

	return nil
}

func (g *Generator) generateConfigFiles() error {
	data := g.getTemplateData()

	// Generate config.go
	if err := g.generateFile("internal/config/config.go", "config/config.go.tmpl", data); err != nil {
		return err
	}

	// Generate config.yaml
	if err := g.generateFile("config.yaml", "base/config.yaml.tmpl", data); err != nil {
		return err
	}

	// Generate .env.example
	if err := g.generateFile(".env.example", "base/env.example.tmpl", data); err != nil {
		return err
	}

	return nil
}

func (g *Generator) generateDeploymentFiles() error {
	data := g.getTemplateData()

	// Generate fly.toml
	if g.config.Deployment.Type == deployment.TypeFly {
		if err := g.generateFile("fly.toml", "fly/fly.toml.tmpl", data); err != nil {
			return err
		}
	}

	// Generate GitHub Actions workflow
	if err := g.fs.MkdirAll(filepath.Join(g.config.OutputDir, ".github", "workflows"), 0755); err != nil {
		return fmt.Errorf("failed to create .github/workflows directory: %w", err)
	}

	if err := g.generateFile(".github/workflows/deploy.yml", "github/workflows/deploy.yml.tmpl", data); err != nil {
		return err
	}

	return nil
}

func (g *Generator) generateDevelopmentFiles() error {
	data := g.getTemplateData()

	// Generate wgo.yaml (hot reload is always enabled)
	if err := g.generateFile("wgo.yaml", "wgo/wgo.yaml.tmpl", data); err != nil {
		return err
	}

	// Generate Dockerfile
	if err := g.generateFile("Dockerfile", "docker/Dockerfile.tmpl", data); err != nil {
		return err
	}

	// Generate docker-compose.yml
	if err := g.generateFile("docker-compose.yml", "docker/docker-compose.yml.tmpl", data); err != nil {
		return err
	}


	return nil
}

