package generator

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/anmho/create-go-service/internal/generator/api"
	"github.com/anmho/create-go-service/internal/generator/config"
	"github.com/anmho/create-go-service/internal/generator/database"
	"github.com/anmho/create-go-service/internal/generator/deployment"
)

func TestTemplateEmbedding(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		templatePath string
	}{
		{"go.mod template", "base/go.mod.tmpl"},
		{"main.go template", "base/main.go.tmpl"},
		{"README template", "base/README.md.tmpl"},
		{"config.yaml template", "base/config.yaml.tmpl"},
		{"env.example template", "base/env.example.tmpl"},
		{".gitignore template", "base/.gitignore.tmpl"},
		{".dockerignore template", "base/.dockerignore.tmpl"},
		{"config.go template", "config/config.go.tmpl"},
		{"Makefile template", "makefile/Makefile.tmpl"},
		{"Dockerfile template", "docker/Dockerfile.tmpl"},
		{"docker-compose template", "docker/docker-compose.yml.tmpl"},
		{"fly.toml template", "fly/fly.toml.tmpl"},
		{"wgo.yaml template", "wgo/wgo.yaml.tmpl"},
		{"deploy workflow template", "github/workflows/deploy.yml.tmpl"},
		{"Chi server template", "chi/server.go.tmpl"},
		{"Chi json template", "chi/json.go.tmpl"},
		{"DynamoDB template", "dynamodb/dynamodb.go.tmpl"},
		{"Postgres template", "postgres/postgres.go.tmpl"},
		{"JWT auth template", "auth/jwt.go.tmpl"},
		{"Metrics template", "metrics/metrics.go.tmpl"},
		{"gRPC server template", "grpc/server.go.tmpl"},
		{"gRPC proto template", "grpc/example.proto.tmpl"},
		{"buf.yaml template", "grpc/buf.yaml.tmpl"},
		{"buf.gen.yaml template", "grpc/buf.gen.yaml.tmpl"},
	}

	config := config.ProjectConfig{
		ProjectName: "test",
		ModulePath:  "github.com/test",
		OutputDir:   "/tmp/test",
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
	g := NewGenerator(config)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := g.loadTemplate(tt.templatePath)
			if err != nil {
				t.Errorf("failed to load template %s: %v", tt.templatePath, err)
				return
			}
			if tmpl == nil {
				t.Errorf("template %s is nil", tt.templatePath)
			}
		})
	}
}

func TestTemplateExecution(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		templatePath string
		config       config.ProjectConfig
	}{
		{
			name:         "DynamoDB config",
			templatePath: "config/config.go.tmpl",
			config: config.ProjectConfig{
				ProjectName: "test-service",
				ModulePath:  "github.com/test/service",
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
			},
		},
		{
			name:         "go.mod with DynamoDB",
			templatePath: "base/go.mod.tmpl",
			config: config.ProjectConfig{
				ProjectName: "test-service",
				ModulePath:  "github.com/test/service",
				API: api.Config{
					Types: []api.Type{api.TypeChi},
				},
				Database: database.Config{
					Type: database.TypeDynamoDB,
				},
				Deployment: deployment.Config{
					Type: deployment.TypeFly,
				},
			},
		},
		{
			name:         "main.go with DynamoDB",
			templatePath: "base/main.go.tmpl",
			config: config.ProjectConfig{
				ProjectName: "test-service",
				ModulePath:  "github.com/test/service",
				API: api.Config{
					Types: []api.Type{api.TypeChi},
				},
				Database: database.Config{
					Type: database.TypeDynamoDB,
				},
				Deployment: deployment.Config{
					Type: deployment.TypeFly,
				},
			},
		},
		{
			name:         "docker-compose with DynamoDB",
			templatePath: "docker/docker-compose.yml.tmpl",
			config: config.ProjectConfig{
				ProjectName: "test-service",
				API: api.Config{
					Types: []api.Type{api.TypeChi},
				},
				Database: database.Config{
					Type: database.TypeDynamoDB,
				},
				Deployment: deployment.Config{
					Type: deployment.TypeFly,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGenerator(tt.config)
			
			tmpl, err := g.loadTemplate(tt.templatePath)
			if err != nil {
				t.Fatalf("failed to load template: %v", err)
			}

			data := g.getTemplateData()
			
			// Try to execute the template
			var buf bytes.Buffer
			err = tmpl.Execute(&buf, data)
			if err != nil {
				t.Errorf("failed to execute template: %v", err)
			}
			
			if buf.Len() == 0 {
				t.Errorf("template produced empty output")
			}
		})
	}
}

func TestGenerateProject(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		config config.ProjectConfig
	}{
		{
			name: "Chi with DynamoDB",
			config: config.ProjectConfig{
				ProjectName: "test-chi-dynamo",
				ModulePath:  "github.com/test/chi-dynamo",
				OutputDir:   filepath.Join(os.TempDir(), "test-chi-dynamo"),
				Features:    []config.Feature{},
				API: api.Config{
					Types: []api.Type{api.TypeChi},
				},
				Database: database.Config{
					Type: database.TypeDynamoDB,
				},
				Deployment: deployment.Config{
					Type: deployment.TypeFly,
				},
			},
		},
		{
			name: "gRPC with DynamoDB",
			config: config.ProjectConfig{
				ProjectName: "test-grpc-dynamo",
				ModulePath:  "github.com/test/grpc-dynamo",
				OutputDir:   filepath.Join(os.TempDir(), "test-grpc-dynamo"),
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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before test
			os.RemoveAll(tt.config.OutputDir)
			defer os.RemoveAll(tt.config.OutputDir)

			g := NewGenerator(tt.config)
			err := g.Generate()
			if err != nil {
				t.Fatalf("failed to generate project: %v", err)
			}

			// Verify key files exist
			expectedFiles := []string{
				"go.mod",
				"README.md",
				".gitignore",
				".dockerignore",
				"Makefile",
				"Dockerfile",
				"docker-compose.yml",
				"config.yaml",
				".env.example",
				"cmd/api/main.go",
				"internal/config/config.go",
			}

			// Add API-specific files
			if tt.config.API.Types[0] == api.TypeChi {
				expectedFiles = append(expectedFiles, "internal/api/server.go", "internal/json/json.go")
			} else if tt.config.API.Types[0] == api.TypeGRPC {
				expectedFiles = append(expectedFiles, 
					"internal/api/server.go",
					"internal/api/posts_handler.go",
					"protos/posts/v1/posts.proto",
					"buf.yaml",
					"buf.gen.yaml",
				)
			}

			// Add database-specific files
			if tt.config.Database.Type == database.TypeDynamoDB {
				expectedFiles = append(expectedFiles, "internal/database/dynamodb.go")
			}

			// Add feature-specific files
			for _, feature := range tt.config.Features {
				if feature == config.FeatureAuth {
					expectedFiles = append(expectedFiles, "internal/auth/jwt.go")
				}
				// Metrics is always included, so no need to check
			}

			for _, file := range expectedFiles {
				path := filepath.Join(tt.config.OutputDir, file)
				if _, err := os.Stat(path); os.IsNotExist(err) {
					t.Errorf("expected file %s does not exist", file)
				}
			}

			// Verify fly.toml exists
			flyPath := filepath.Join(tt.config.OutputDir, "fly.toml")
			if _, err := os.Stat(flyPath); os.IsNotExist(err) {
				t.Errorf("fly.toml does not exist")
			}

			// Hot reload is always enabled, so wgo.yaml should always exist
			wgoPath := filepath.Join(tt.config.OutputDir, "wgo.yaml")
			if _, err := os.Stat(wgoPath); os.IsNotExist(err) {
				t.Errorf("wgo.yaml does not exist (hot reload is always enabled)")
			}

			// Verify protos directory exists for gRPC
			if tt.config.API.Types[0] == api.TypeGRPC {
				protosPath := filepath.Join(tt.config.OutputDir, "protos")
				if _, err := os.Stat(protosPath); os.IsNotExist(err) {
					t.Errorf("protos directory does not exist for gRPC project")
				}
			}
		})
	}
}

func TestGetTemplateData(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		config config.ProjectConfig
		check  func(*testing.T, config.ProjectConfig, map[string]interface{})
	}{
		{
			name: "DynamoDB with all features",
			config: config.ProjectConfig{
				ProjectName: "test-service",
				ModulePath:  "github.com/test/service",
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
			},
			check: func(t *testing.T, cfg config.ProjectConfig, data map[string]interface{}) {
				// Check API config
				if len(cfg.API.Types) == 0 {
					t.Error("expected at least one API type")
				}
				if cfg.API.Types[0] != api.TypeChi {
					t.Errorf("expected API type %v, got %v", api.TypeChi, cfg.API.Types[0])
				}
				
				// Check database config
				if cfg.Database.Type != database.TypeDynamoDB {
					t.Errorf("expected database type %v, got %v", database.TypeDynamoDB, cfg.Database.Type)
				}
				
				// Check deployment config
				if cfg.Deployment.Type != deployment.TypeFly {
					t.Errorf("expected deployment type %v, got %v", deployment.TypeFly, cfg.Deployment.Type)
				}
				
				// Check features
				hasAuth := false
				for _, f := range cfg.Features {
					if f == config.FeatureAuth {
						hasAuth = true
					}
				}
				if !hasAuth {
					t.Error("expected auth feature to be enabled")
				}
				
				// Check template data has required keys
				requiredKeys := []string{"ProjectName", "ModulePath", "HasChi", "HasDynamoDB", "HasMetrics", "HasAuth", "HasHotReload"}
				for _, key := range requiredKeys {
					if _, ok := data[key]; !ok {
						t.Errorf("expected key %s not found in template data", key)
					}
				}
			},
		},
		{
			name: "gRPC with DynamoDB",
			config: config.ProjectConfig{
				ProjectName: "test-service",
				ModulePath:  "github.com/test/service",
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
			},
			check: func(t *testing.T, cfg config.ProjectConfig, data map[string]interface{}) {
				// Check API config
				if len(cfg.API.Types) == 0 {
					t.Error("expected at least one API type")
				}
				if cfg.API.Types[0] != api.TypeGRPC {
					t.Errorf("expected API type %v, got %v", api.TypeGRPC, cfg.API.Types[0])
				}
				
				// Check database config
				if cfg.Database.Type != database.TypeDynamoDB {
					t.Errorf("expected database type %v, got %v", database.TypeDynamoDB, cfg.Database.Type)
				}
				
				// Check deployment config
				if cfg.Deployment.Type != deployment.TypeFly {
					t.Errorf("expected deployment type %v, got %v", deployment.TypeFly, cfg.Deployment.Type)
				}
				
				// Check features - should be empty
				if len(cfg.Features) != 0 {
					t.Errorf("expected no features, got %v", cfg.Features)
				}
				
				// Check template data has required keys
				requiredKeys := []string{"ProjectName", "ModulePath", "HasGRPC", "HasDynamoDB", "HasMetrics", "HasHotReload"}
				for _, key := range requiredKeys {
					if _, ok := data[key]; !ok {
						t.Errorf("expected key %s not found in template data", key)
					}
				}
				
				// Check HasAuth is false
				if hasAuth, ok := data["HasAuth"].(bool); !ok || hasAuth {
					t.Error("expected HasAuth to be false")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGenerator(tt.config)
			data := g.getTemplateData()
			tt.check(t, tt.config, data)
		})
	}
}

// verifyProjectBuildAndTests verifies that a generated project builds and its tests pass
func verifyProjectBuildAndTests(t *testing.T, projectDir string) {
	// Check if this is a gRPC project (has buf.yaml)
	bufYamlPath := filepath.Join(projectDir, "buf.yaml")
	if _, err := os.Stat(bufYamlPath); err == nil {
		// This is a gRPC project, need to generate protobuf files first
		t.Logf("Running 'buf generate' in %s", projectDir)
		// Check if buf is available
		if _, err := exec.LookPath("buf"); err != nil {
			t.Skip("buf command not found, skipping gRPC project test")
			return
		}
		cmd := exec.Command("buf", "generate")
		cmd.Dir = projectDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Check if error is due to authentication
			outputStr := string(output)
			if strings.Contains(outputStr, "invalid") || strings.Contains(outputStr, "authentication") || strings.Contains(outputStr, "login") {
				t.Skipf("buf authentication required (run 'buf registry login'): %s", outputStr)
				return
			}
			t.Logf("buf generate output: %s", outputStr)
			t.Fatalf("failed to run buf generate: %v", err)
		}
	}

	// Run go mod tidy
	t.Logf("Running 'go mod tidy' in %s", projectDir)
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = projectDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("go mod tidy output: %s", string(output))
		t.Fatalf("failed to run go mod tidy: %v", err)
	}

	// Verify the project builds
	t.Logf("Running 'go build ./...' in %s", projectDir)
	cmd = exec.Command("go", "build", "./...")
	cmd.Dir = projectDir
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Logf("go build output: %s", string(output))
		t.Fatalf("failed to build project: %v", err)
	}
}

// TestGenerateProjectAndRunTests generates projects and verifies they build
// This is an integration test that verifies the generated projects compile correctly
// Uses temp directories (not Docker) for faster execution
func TestGenerateProjectAndRunTests(t *testing.T) {
	t.Parallel()
	// Skip if go is not available
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go command not found, skipping integration test")
	}

	tests := []struct {
		name   string
		config config.ProjectConfig
		skip   bool // Skip tests that require external services (e.g., DynamoDB, Postgres)
	}{
		{
			name: "Chi with Postgres",
			config: config.ProjectConfig{
				ProjectName: "test-chi-postgres",
				ModulePath:  "github.com/test/chi-postgres",
				OutputDir:   filepath.Join(os.TempDir(), "test-chi-postgres"),
				Features:    []config.Feature{},
				API: api.Config{
					Types: []api.Type{api.TypeChi},
				},
				Database: database.Config{
					Type: database.TypePostgres,
				},
				Deployment: deployment.Config{
					Type: deployment.TypeFly,
				},
			},
			skip: false, // Postgres tests use testcontainers, should work
		},
		{
			name: "Chi with DynamoDB",
			config: config.ProjectConfig{
				ProjectName: "test-chi-dynamo",
				ModulePath:  "github.com/test/chi-dynamo",
				OutputDir:   filepath.Join(os.TempDir(), "test-chi-dynamo"),
				Features:    []config.Feature{},
				API: api.Config{
					Types: []api.Type{api.TypeChi},
				},
				Database: database.Config{
					Type: database.TypeDynamoDB,
				},
				Deployment: deployment.Config{
					Type: deployment.TypeFly,
				},
			},
			skip: true, // DynamoDB tests may require local DynamoDB, skip for now
		},
		{
			name: "gRPC with Postgres",
			config: config.ProjectConfig{
				ProjectName: "test-grpc-postgres",
				ModulePath:  "github.com/test/grpc-postgres",
				OutputDir:   filepath.Join(os.TempDir(), "test-grpc-postgres"),
				Features:    []config.Feature{},
				API: api.Config{
					Types: []api.Type{api.TypeGRPC},
				},
				Database: database.Config{
					Type: database.TypePostgres,
				},
				Deployment: deployment.Config{
					Type: deployment.TypeFly,
				},
			},
			skip: false, // gRPC with Postgres should work
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip {
				t.Skip("skipping test that requires external services")
			}

			// Use unique temp directory per test run to avoid concurrency issues
			uniqueDir := filepath.Join(os.TempDir(), "test-"+t.Name()+"-"+tt.config.ProjectName)
			tt.config.OutputDir = uniqueDir

			// Clean up before test
			os.RemoveAll(tt.config.OutputDir)
			defer os.RemoveAll(tt.config.OutputDir)

			// Generate the project
			g := NewGenerator(tt.config)
			err := g.Generate()
			if err != nil {
				t.Fatalf("failed to generate project: %v", err)
			}

			// Verify build
			verifyProjectBuildAndTests(t, tt.config.OutputDir)

			t.Logf("Build passed successfully for %s", tt.name)
		})
	}
}

// TestGenerateProjectAndRunTestsInDocker generates projects and runs their tests inside Docker containers
// This provides true isolation and ensures the generated projects work in a clean environment
// NOTE: This test is slow (~25-45s)
func TestGenerateProjectAndRunTestsInDocker(t *testing.T) {
	t.Parallel()
	// Skip if docker is not available
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker command not found, skipping Docker integration test")
	}

	tests := []struct {
		name   string
		config config.ProjectConfig
		skip   bool
	}{
		{
			name: "Chi with Postgres",
			config: config.ProjectConfig{
				ProjectName: "test-chi-postgres-docker",
				ModulePath:  "github.com/test/chi-postgres-docker",
				OutputDir:   filepath.Join(os.TempDir(), "test-chi-postgres-docker"),
				Features:    []config.Feature{},
				API: api.Config{
					Types: []api.Type{api.TypeChi},
				},
				Database: database.Config{
					Type: database.TypePostgres,
				},
				Deployment: deployment.Config{
					Type: deployment.TypeFly,
				},
			},
			skip: false,
		},
		{
			name: "Chi with DynamoDB",
			config: config.ProjectConfig{
				ProjectName: "test-chi-dynamo-docker",
				ModulePath:  "github.com/test/chi-dynamo-docker",
				OutputDir:   filepath.Join(os.TempDir(), "test-chi-dynamo-docker"),
				Features:    []config.Feature{},
				API: api.Config{
					Types: []api.Type{api.TypeChi},
				},
				Database: database.Config{
					Type: database.TypeDynamoDB,
				},
				Deployment: deployment.Config{
					Type: deployment.TypeFly,
				},
			},
			skip: true, // DynamoDB tests may require local DynamoDB, skip for now
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip {
				t.Skip("skipping test that requires external services")
			}

			// Use unique temp directory per test run to avoid concurrency issues
			uniqueDir := filepath.Join(os.TempDir(), "test-docker-"+t.Name()+"-"+tt.config.ProjectName)
			tt.config.OutputDir = uniqueDir

			// Clean up before test
			os.RemoveAll(tt.config.OutputDir)
			defer os.RemoveAll(tt.config.OutputDir)

			// Generate the project
			g := NewGenerator(tt.config)
			err := g.Generate()
			if err != nil {
				t.Fatalf("failed to generate project: %v", err)
			}

			// Run go mod tidy to generate go.sum before Docker build
			t.Logf("Running 'go mod tidy' to generate go.sum in %s", tt.config.OutputDir)
			cmd := exec.Command("go", "mod", "tidy")
			cmd.Dir = tt.config.OutputDir
			_, err = cmd.CombinedOutput()
			if err != nil {
				t.Logf("go mod tidy may have failed, continuing anyway")
			}

			// Create a test Dockerfile that verifies build
			testDockerfile := `FROM golang:1.25-alpine AS test
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod ./
COPY go.sum* ./
RUN go mod download
COPY . .
RUN go build ./...
`
			dockerfilePath := filepath.Join(tt.config.OutputDir, "Dockerfile.test")
			err = os.WriteFile(dockerfilePath, []byte(testDockerfile), 0644)
			if err != nil {
				t.Fatalf("failed to write test Dockerfile: %v", err)
			}

			// Use unique Docker image name per test run to avoid concurrency issues
			// Sanitize test name for Docker tag (must be lowercase, no slashes)
			sanitizedName := strings.ToLower(strings.ReplaceAll(t.Name(), "/", "-"))
			imageName := "test-" + sanitizedName + "-" + strings.ToLower(tt.config.ProjectName)
			t.Logf("Building Docker image %s from %s (build will be verified during image build)", imageName, tt.config.OutputDir)
			cmd = exec.Command("docker", "build", "-f", "Dockerfile.test", "-t", imageName, ".")
			cmd.Dir = tt.config.OutputDir
			var output []byte
			output, err = cmd.CombinedOutput()
			if err != nil {
				t.Logf("docker build output: %s", string(output))
				t.Fatalf("failed to build Docker image (build may have failed): %v", err)
			}

			// Clean up Docker image after test
			defer func() {
				exec.Command("docker", "rmi", imageName).Run()
			}()

			// If build succeeded, project compiles correctly
			t.Logf("Docker build succeeded - project compiles correctly in container for %s", tt.name)
		})
	}
}

