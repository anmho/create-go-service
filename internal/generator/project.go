package generator

import (
	"fmt"
	"path/filepath"

	"github.com/anmho/create-go-service/internal/generator/api"
	"github.com/anmho/create-go-service/internal/generator/config"
	"github.com/anmho/create-go-service/internal/generator/database"
	"github.com/anmho/create-go-service/internal/generator/deployment"
)

// fileMapping represents a source template to output file mapping
type fileMapping struct {
	outputPath   string
	templatePath string
}

// fileGenerationRule defines when and what files to generate
type fileGenerationRule struct {
	files     []fileMapping
	condition func(*Generator) bool
}

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

	// Get all file generation rules based on project configuration
	rules := g.getFileGenerationRules()
	data := g.getTemplateData()

	// Generate all files based on rules
	for _, rule := range rules {
		if rule.condition == nil || rule.condition(g) {
			if err := g.generateFiles(rule.files, data); err != nil {
				return err
			}
		}
	}

	return nil
}

// getFileGenerationRules returns all file generation rules based on project configuration
func (g *Generator) getFileGenerationRules() []fileGenerationRule {
	var rules []fileGenerationRule

	// Base files (always generated)
	rules = append(rules, fileGenerationRule{
		files: []fileMapping{
			{"go.mod", "base/go.mod.tmpl"},
			{"README.md", "base/README.md.tmpl"},
			{".gitignore", "base/.gitignore.tmpl"},
			{".dockerignore", "base/.dockerignore.tmpl"},
			{"internal/config/local.yaml", "base/local.yaml.tmpl"},
			{"internal/config/production.yaml", "base/production.yaml.tmpl"},
			{"local.yaml", "base/local.yaml.tmpl"},
			{"production.yaml", "base/production.yaml.tmpl"},
			{".env.local.example", "base/env.local.example.tmpl"},
			{".env.production.example", "base/env.production.example.tmpl"},
			{"Makefile", "makefile/Makefile.tmpl"},
		},
	})

	// Config files (always generated)
	rules = append(rules, fileGenerationRule{
		files: []fileMapping{
			{"internal/config/config.go", "config/config.go.tmpl"},
			{".env.example", "base/env.example.tmpl"},
			{".env", "base/env.tmpl"},
			{".env.local", "base/env.local.example.tmpl"},
		},
	})

	// Development files (always generated)
	rules = append(rules, fileGenerationRule{
		files: []fileMapping{
			{"wgo.yaml", "wgo/wgo.yaml.tmpl"},
			{"Dockerfile", "docker/Dockerfile.tmpl"},
			{"docker-compose.yml", "docker/docker-compose.yml.tmpl"},
		},
	})

	// CLI files (always generated)
	rules = append(rules, fileGenerationRule{
		files: []fileMapping{
			{"cmd/postctl/main.go", "cli/main.go.tmpl"},
			{"internal/cli/root.go", "cli/root.go.tmpl"},
			{"internal/cli/server.go", "cli/server.go.tmpl"},
			{"internal/cli/posts.go", "cli/posts.go.tmpl"},
			{"internal/cli/version.go", "cli/version.go.tmpl"},
		},
	})

	// Metrics files (always generated)
	rules = append(rules, fileGenerationRule{
		files: []fileMapping{
			{"internal/metrics/metrics.go", "metrics/metrics.go.tmpl"},
		},
	})

	// API type-specific files
	for _, apiType := range g.config.API.Types {
		switch apiType {
		case api.TypeChi:
			rules = append(rules, fileGenerationRule{
				files: []fileMapping{
					{"internal/api/server.go", "chi/server.go.tmpl"},
					{"internal/json/json.go", "chi/json.go.tmpl"},
					{"internal/posts/handlers.go", "posts/handlers.go.tmpl"},
				},
			})
			// Generate main.go only if not gRPC
			if !g.hasAPIType(api.TypeGRPC) {
				rules = append(rules, fileGenerationRule{
					files: []fileMapping{
						{"cmd/api/main.go", "base/main.go.tmpl"},
					},
				})
			}
		case api.TypeGRPC:
			rules = append(rules, fileGenerationRule{
				files: []fileMapping{
					{"internal/api/server.go", "grpc/server.go.tmpl"},
					{"internal/api/posts_handler.go", "grpc/posts_handler.go.tmpl"},
					{"protos/posts/v1/posts.proto", "grpc/posts.proto.tmpl"},
					{"buf.yaml", "grpc/buf.yaml.tmpl"},
					{"buf.gen.yaml", "grpc/buf.gen.yaml.tmpl"},
					{"cmd/api/main.go", "grpc/main.go.tmpl"},
				},
			})
		}
	}

	// Database type-specific files
	switch g.config.Database.Type {
	case database.TypeDynamoDB:
		rules = append(rules, fileGenerationRule{
			files: []fileMapping{
				{"internal/database/dynamodb.go", "dynamodb/dynamodb.go.tmpl"},
				{"internal/posts/dynamodb_table.go", "posts/dynamodb_table.go.tmpl"},
				{"internal/posts/post_table_test.go", "posts/dynamodb_table_test.go.tmpl"},
			},
		})
		// Terraform files for DynamoDB
		rules = append(rules, fileGenerationRule{
			files: []fileMapping{
				{"terraform/main.tf", "terraform/main.tf.tmpl"},
				{"terraform/variables.tf", "terraform/variables.tf.tmpl"},
				{"terraform/.gitignore", "terraform/.gitignore.tmpl"},
				{"terraform/README.md", "terraform/README.md.tmpl"},
			},
		})
	case database.TypePostgres:
		rules = append(rules, fileGenerationRule{
			files: []fileMapping{
				{"internal/database/postgres.go", "postgres/postgres.go.tmpl"},
				{"internal/posts/postgres_table.go", "posts/postgres_table.go.tmpl"},
				{"internal/posts/post_table_test.go", "posts/postgres_table_test.go.tmpl"},
				{"atlas.hcl", "atlas/atlas.hcl.tmpl"},
				{"migrations/001_initial.up.sql", "atlas/migrations/001_initial.up.sql.tmpl"},
				{"migrations/001_initial.down.sql", "atlas/migrations/001_initial.down.sql.tmpl"},
			},
		})
	}

	// Posts domain files (always generated)
	rules = append(rules, fileGenerationRule{
		files: []fileMapping{
			{"internal/posts/post.go", "posts/post.go.tmpl"},
			{"internal/posts/service.go", "posts/service.go.tmpl"},
			{"internal/posts/converters.go", "posts/converters.go.tmpl"},
			{"internal/posts/converters_test.go", "posts/converters_test.go.tmpl"},
		},
	})

	// Feature-specific files
	for _, feature := range g.config.Features {
		switch feature {
		case config.FeatureAuth:
			rules = append(rules, fileGenerationRule{
				files: []fileMapping{
					{"internal/auth/jwt.go", "auth/jwt.go.tmpl"},
				},
			})
		case config.FeaturePostHog:
			rules = append(rules, fileGenerationRule{
				files: []fileMapping{
					{"internal/posthog/posthog.go", "posthog/posthog.go.tmpl"},
				},
			})
		}
	}

	// Deployment type-specific files
	switch g.config.Deployment.Type {
	case deployment.TypeFly:
		rules = append(rules, fileGenerationRule{
			files: []fileMapping{
				{"fly.toml", "fly/fly.toml.tmpl"},
				{".github/workflows/deploy.yml", "github/workflows/deploy.yml.tmpl"},
			},
			condition: func(g *Generator) bool {
				// Create .github/workflows directory if needed
				_ = g.fs.MkdirAll(filepath.Join(g.config.OutputDir, ".github", "workflows"), 0755)
				return true
			},
		})
	}

	return rules
}

// hasAPIType checks if the project has a specific API type
func (g *Generator) hasAPIType(apiType api.Type) bool {
	for _, t := range g.config.API.Types {
		if t == apiType {
			return true
		}
	}
	return false
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

// generateFiles generates multiple files from a list of mappings
func (g *Generator) generateFiles(files []fileMapping, data interface{}) error {
	for _, file := range files {
		if err := g.generateFile(file.outputPath, file.templatePath, data); err != nil {
			return fmt.Errorf("failed to generate %s: %w", file.outputPath, err)
		}
	}
	return nil
}
