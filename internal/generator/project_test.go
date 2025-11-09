package generator

import (
	"errors"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/anmho/create-go-service/internal/generator/api"
	"github.com/anmho/create-go-service/internal/generator/config"
	"github.com/anmho/create-go-service/internal/generator/database"
	"github.com/anmho/create-go-service/internal/generator/deployment"
	"github.com/anmho/create-go-service/internal/generator/mocks"
	"github.com/stretchr/testify/mock"
)

func TestNewGenerator(t *testing.T) {
	t.Parallel()
	cfg := config.ProjectConfig{
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

	gen := NewGenerator(cfg)
	if gen == nil {
		t.Fatal("NewGenerator returned nil")
	}
	if gen.config.ProjectName != "test" {
		t.Errorf("expected ProjectName 'test', got %s", gen.config.ProjectName)
	}
	if gen.fs == nil {
		t.Error("FileSystem should not be nil")
	}
	if gen.templateLoader == nil {
		t.Error("TemplateLoader should not be nil")
	}
}

func TestNewGeneratorWithDeps(t *testing.T) {
	t.Parallel()
	cfg := config.ProjectConfig{
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
	mockFS := mocks.NewFileSystem(t)
	mockLoader := NewMockTemplateLoader()

	gen := NewGeneratorWithDeps(cfg, mockFS, mockLoader)
	if gen == nil {
		t.Fatal("NewGeneratorWithDeps returned nil")
	}
	if gen.fs != mockFS {
		t.Error("FileSystem should be the provided mock")
	}
	if gen.templateLoader != mockLoader {
		t.Error("TemplateLoader should be the provided mock")
	}
}

func TestCreateDirectoryStructure(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		config   config.ProjectConfig
		expected []string
	}{
		{
			name: "Chi with DynamoDB",
			config: config.ProjectConfig{
				OutputDir: "/tmp/test",
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
			expected: []string{
				"/tmp/test/cmd/api",
				"/tmp/test/cmd/postctl",
				"/tmp/test/internal/config",
				"/tmp/test/internal/cli",
				"/tmp/test/internal/database",
				"/tmp/test/internal/posts",
				"/tmp/test/internal/metrics",
				"/tmp/test/internal/auth",
				"/tmp/test/internal/api",
				"/tmp/test/internal/json",
				"/tmp/test/terraform",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := mocks.NewFileSystem(t)
			// Set up expectations for all expected directories
			for _, expected := range tt.expected {
				mockFS.On("MkdirAll", expected, mock.Anything).Return(nil)
			}
			mockLoader := NewMockTemplateLoader()
			gen := NewGeneratorWithDeps(tt.config, mockFS, mockLoader)

			err := gen.createDirectoryStructure()
			if err != nil {
				t.Fatalf("createDirectoryStructure failed: %v", err)
			}

			mockFS.AssertExpectations(t)
		})
	}
}

func TestCreateDirectoryStructureError(t *testing.T) {
	t.Parallel()
	mockFS := mocks.NewFileSystem(t)
	mockFS.On("MkdirAll", mock.Anything, mock.Anything).Return(errors.New("mkdir failed"))
	mockLoader := NewMockTemplateLoader()
	config := config.ProjectConfig{
		OutputDir: "/tmp/test",
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

	gen := NewGeneratorWithDeps(config, mockFS, mockLoader)
	err := gen.createDirectoryStructure()
	if err == nil {
		t.Error("expected error from createDirectoryStructure, got nil")
	}
}

func TestGenerateFile(t *testing.T) {
	t.Parallel()
	mockFS := mocks.NewFileSystem(t)
	mockLoader := NewMockTemplateLoader()

	// Create a simple template
	tmpl := template.Must(template.New("test.tmpl").Parse("Hello {{.ProjectName}}"))
	mockLoader.Templates["test.tmpl"] = tmpl

	config := config.ProjectConfig{
		OutputDir:   "/tmp/test",
		ProjectName: "test-service",
	}
	gen := NewGeneratorWithDeps(config, mockFS, mockLoader)

	data := map[string]interface{}{
		"ProjectName": "test-service",
	}

	expectedPath := filepath.Join("/tmp/test", "output.txt")
	expectedContent := "Hello test-service"

	// Set up expectations
	mockFS.On("MkdirAll", filepath.Dir(expectedPath), mock.Anything).Return(nil)
	mockFS.On("WriteFile", expectedPath, []byte(expectedContent), mock.Anything).Return(nil)

	err := gen.generateFile("output.txt", "test.tmpl", data)
	if err != nil {
		t.Fatalf("generateFile failed: %v", err)
	}

	mockFS.AssertExpectations(t)
}

func TestGenerateFileTemplateError(t *testing.T) {
	t.Parallel()
	mockFS := mocks.NewFileSystem(t)
	mockLoader := NewMockTemplateLoader()
	mockLoader.LoadError = errors.New("template not found")

	config := config.ProjectConfig{
		OutputDir: "/tmp/test",
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
	gen := NewGeneratorWithDeps(config, mockFS, mockLoader)

	err := gen.generateFile("output.txt", "nonexistent.tmpl", nil)
	if err == nil {
		t.Error("expected error from generateFile, got nil")
	}
}

func TestGenerateFileWriteError(t *testing.T) {
	t.Parallel()
	mockFS := mocks.NewFileSystem(t)
	mockLoader := NewMockTemplateLoader()

	tmpl := template.Must(template.New("test.tmpl").Parse("test"))
	mockLoader.Templates["test.tmpl"] = tmpl

	config := config.ProjectConfig{
		OutputDir: "/tmp/test",
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
	gen := NewGeneratorWithDeps(config, mockFS, mockLoader)

	expectedPath := filepath.Join("/tmp/test", "output.txt")
	mockFS.On("MkdirAll", filepath.Dir(expectedPath), mock.Anything).Return(nil)
	mockFS.On("WriteFile", expectedPath, mock.Anything, mock.Anything).Return(errors.New("write failed"))

	err := gen.generateFile("output.txt", "test.tmpl", nil)
	if err == nil {
		t.Error("expected error from generateFile, got nil")
	}
}

func TestGetFileGenerationRules(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		config           config.ProjectConfig
		expectedMinFiles int // Minimum number of files we expect
	}{
		{
			name: "Chi with DynamoDB",
			config: config.ProjectConfig{
				ProjectName: "test-service",
				ModulePath:  "github.com/test/service",
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
			},
			expectedMinFiles: 20, // Base files + config + dev + CLI + metrics + Chi + DynamoDB + posts + terraform
		},
		{
			name: "gRPC with Postgres and Auth",
			config: config.ProjectConfig{
				ProjectName: "test-service",
				ModulePath:  "github.com/test/service",
				OutputDir:   "/tmp/test",
				API: api.Config{
					Types: []api.Type{api.TypeGRPC},
				},
				Database: database.Config{
					Type: database.TypePostgres,
				},
				Deployment: deployment.Config{
					Type: deployment.TypeFly,
				},
				Features: []config.Feature{config.FeatureAuth},
			},
			expectedMinFiles: 20, // Base files + config + dev + CLI + metrics + gRPC + Postgres + posts + atlas + auth
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := mocks.NewFileSystem(t)
			mockLoader := NewMockTemplateLoader()
			gen := NewGeneratorWithDeps(tt.config, mockFS, mockLoader)

			// Mock MkdirAll for .github/workflows (called by deployment condition)
			mockFS.On("MkdirAll", filepath.Join("/tmp/test", ".github", "workflows"), mock.Anything).Return(nil)

			rules := gen.getFileGenerationRules()

			// Count total files
			totalFiles := 0
			for _, rule := range rules {
				if rule.condition == nil || rule.condition(gen) {
					totalFiles += len(rule.files)
				}
			}

			if totalFiles < tt.expectedMinFiles {
				t.Errorf("expected at least %d files, got %d", tt.expectedMinFiles, totalFiles)
			}
		})
	}
}

func TestGenerateAPIFilesInRules(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		apiType       api.Type
		expectedFiles []string
	}{
		{
			name:    "Chi",
			apiType: api.TypeChi,
			expectedFiles: []string{
				"internal/api/server.go",
				"internal/json/json.go",
				"internal/posts/handlers.go",
				"cmd/api/main.go", // Only if not gRPC
			},
		},
		{
			name:    "gRPC",
			apiType: api.TypeGRPC,
			expectedFiles: []string{
				"internal/api/server.go",
				"internal/api/posts_handler.go",
				"protos/posts/v1/posts.proto",
				"buf.yaml",
				"buf.gen.yaml",
				"cmd/api/main.go",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := mocks.NewFileSystem(t)
			mockLoader := NewMockTemplateLoader()

			config := config.ProjectConfig{
				ProjectName: "test-service",
				ModulePath:  "github.com/test/service",
				OutputDir:   "/tmp/test",
				API: api.Config{
					Types: []api.Type{tt.apiType},
				},
				Database: database.Config{
					Type: database.TypeDynamoDB,
				},
				Deployment: deployment.Config{
					Type: deployment.TypeFly,
				},
			}
			gen := NewGeneratorWithDeps(config, mockFS, mockLoader)

			// Mock MkdirAll for .github/workflows (called by deployment condition)
			mockFS.On("MkdirAll", filepath.Join("/tmp/test", ".github", "workflows"), mock.Anything).Return(nil)

			rules := gen.getFileGenerationRules()

			// Find API-related files in rules
			foundFiles := make(map[string]bool)
			for _, rule := range rules {
				if rule.condition == nil || rule.condition(gen) {
					for _, file := range rule.files {
						foundFiles[file.outputPath] = true
					}
				}
			}

			// Check that expected files are present
			for _, expectedFile := range tt.expectedFiles {
				if !foundFiles[expectedFile] {
					t.Errorf("expected file %s not found in generation rules", expectedFile)
				}
			}
		})
	}
}

func TestGenerateDatabaseFilesInRules(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		database      database.Type
		expectedFiles []string
	}{
		{
			name:     "DynamoDB",
			database: database.TypeDynamoDB,
			expectedFiles: []string{
				"internal/database/dynamodb.go",
				"internal/posts/dynamodb_table.go",
				"internal/posts/post_table_test.go",
				"terraform/main.tf",
				"terraform/variables.tf",
				"terraform/.gitignore",
				"terraform/README.md",
			},
		},
		{
			name:     "Postgres",
			database: database.TypePostgres,
			expectedFiles: []string{
				"internal/database/postgres.go",
				"internal/posts/postgres_table.go",
				"internal/posts/post_table_test.go",
				"atlas.hcl",
				"migrations/001_initial.up.sql",
				"migrations/001_initial.down.sql",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := mocks.NewFileSystem(t)
			mockLoader := NewMockTemplateLoader()

			config := config.ProjectConfig{
				ProjectName: "test-service",
				ModulePath:  "github.com/test/service",
				OutputDir:   "/tmp/test",
				API: api.Config{
					Types: []api.Type{api.TypeChi},
				},
				Database: database.Config{
					Type: tt.database,
				},
				Deployment: deployment.Config{
					Type: deployment.TypeFly,
				},
			}
			gen := NewGeneratorWithDeps(config, mockFS, mockLoader)

			// Mock MkdirAll for .github/workflows (called by deployment condition)
			mockFS.On("MkdirAll", filepath.Join("/tmp/test", ".github", "workflows"), mock.Anything).Return(nil)

			rules := gen.getFileGenerationRules()

			// Find database-related files in rules
			foundFiles := make(map[string]bool)
			for _, rule := range rules {
				if rule.condition == nil || rule.condition(gen) {
					for _, file := range rule.files {
						foundFiles[file.outputPath] = true
					}
				}
			}

			// Check that expected files are present
			for _, expectedFile := range tt.expectedFiles {
				if !foundFiles[expectedFile] {
					t.Errorf("expected file %s not found in generation rules", expectedFile)
				}
			}
		})
	}
}

func TestGeneratePostsFilesInRules(t *testing.T) {
	t.Parallel()
	mockFS := mocks.NewFileSystem(t)
	mockLoader := NewMockTemplateLoader()

	config := config.ProjectConfig{
		ProjectName: "test-service",
		ModulePath:  "github.com/test/service",
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
	gen := NewGeneratorWithDeps(config, mockFS, mockLoader)

	// Mock MkdirAll for .github/workflows (called by deployment condition)
	mockFS.On("MkdirAll", filepath.Join("/tmp/test", ".github", "workflows"), mock.Anything).Return(nil)

	rules := gen.getFileGenerationRules()

	// Find posts-related files in rules
	foundFiles := make(map[string]bool)
	for _, rule := range rules {
		if rule.condition == nil || rule.condition(gen) {
			for _, file := range rule.files {
				if filepath.Dir(file.outputPath) == "internal/posts" || filepath.Base(file.outputPath) == "handlers.go" {
					foundFiles[file.outputPath] = true
				}
			}
		}
	}

	// Check that expected posts files are present
	expectedFiles := []string{
		"internal/posts/post.go",
		"internal/posts/service.go",
		"internal/posts/converters.go",
		"internal/posts/converters_test.go",
		"internal/posts/dynamodb_table.go",
		"internal/posts/post_table_test.go",
		"internal/posts/handlers.go",
	}

	for _, expectedFile := range expectedFiles {
		if !foundFiles[expectedFile] {
			t.Errorf("expected file %s not found in generation rules", expectedFile)
		}
	}
}

func TestGenerateFeatureFilesInRules(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		features      []config.Feature
		expectedFiles []string
	}{
		{
			name:     "Auth feature",
			features: []config.Feature{config.FeatureAuth},
			expectedFiles: []string{
				"internal/metrics/metrics.go", // Always generated
				"internal/auth/jwt.go",
			},
		},
		{
			name:     "PostHog feature",
			features: []config.Feature{config.FeaturePostHog},
			expectedFiles: []string{
				"internal/metrics/metrics.go", // Always generated
				"internal/posthog/posthog.go",
			},
		},
		{
			name:     "Both features",
			features: []config.Feature{config.FeatureAuth, config.FeaturePostHog},
			expectedFiles: []string{
				"internal/metrics/metrics.go", // Always generated
				"internal/auth/jwt.go",
				"internal/posthog/posthog.go",
			},
		},
		{
			name:     "No features",
			features: []config.Feature{},
			expectedFiles: []string{
				"internal/metrics/metrics.go", // Always generated
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := mocks.NewFileSystem(t)
			mockLoader := NewMockTemplateLoader()

			config := config.ProjectConfig{
				ProjectName: "test-service",
				ModulePath:  "github.com/test/service",
				OutputDir:   "/tmp/test",
				Features:    tt.features,
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
			gen := NewGeneratorWithDeps(config, mockFS, mockLoader)

			// Mock MkdirAll for .github/workflows (called by deployment condition)
			mockFS.On("MkdirAll", filepath.Join("/tmp/test", ".github", "workflows"), mock.Anything).Return(nil)

			rules := gen.getFileGenerationRules()

			// Find feature-related files in rules
			foundFiles := make(map[string]bool)
			for _, rule := range rules {
				if rule.condition == nil || rule.condition(gen) {
					for _, file := range rule.files {
						if filepath.Dir(file.outputPath) == "internal/metrics" ||
							filepath.Dir(file.outputPath) == "internal/auth" ||
							filepath.Dir(file.outputPath) == "internal/posthog" {
							foundFiles[file.outputPath] = true
						}
					}
				}
			}

			// Check that expected files are present
			for _, expectedFile := range tt.expectedFiles {
				if !foundFiles[expectedFile] {
					t.Errorf("expected file %s not found in generation rules", expectedFile)
				}
			}

			// Check that unexpected files are not present
			allFeatureFiles := []string{
				"internal/metrics/metrics.go",
				"internal/auth/jwt.go",
				"internal/posthog/posthog.go",
			}
			for _, file := range allFeatureFiles {
				shouldExist := false
				for _, expected := range tt.expectedFiles {
					if expected == file {
						shouldExist = true
						break
					}
				}
				if foundFiles[file] != shouldExist {
					if shouldExist {
						t.Errorf("expected file %s to exist but it doesn't", file)
					} else {
						t.Errorf("unexpected file %s found in generation rules", file)
					}
				}
			}
		})
	}
}
