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

func TestGenerateBaseFiles(t *testing.T) {
	t.Parallel()
	mockFS := mocks.NewFileSystem(t)
	mockLoader := NewMockTemplateLoader()
	
	// Create templates for base files
	baseTemplates := []string{
		"base/go.mod.tmpl",
		"base/README.md.tmpl",
		"base/.gitignore.tmpl",
		"base/.dockerignore.tmpl",
		"base/local.yaml.tmpl",
		"base/development.yaml.tmpl",
		"base/production.yaml.tmpl",
		"makefile/Makefile.tmpl",
	}
	for _, tmplPath := range baseTemplates {
		tmpl := template.Must(template.New(tmplPath).Parse("{{.ProjectName}}"))
		mockLoader.Templates[tmplPath] = tmpl
	}

	config := config.ProjectConfig{
		ProjectName: "test-service",
		ModulePath:  "github.com/test/service",
		OutputDir:   "/tmp/test",
	}
	gen := NewGeneratorWithDeps(config, mockFS, mockLoader)

	// Set up expectations for all base files
	expectedFiles := []string{
		"/tmp/test/go.mod",
		"/tmp/test/README.md",
		"/tmp/test/.gitignore",
		"/tmp/test/.dockerignore",
		"/tmp/test/local.yaml",
		"/tmp/test/development.yaml",
		"/tmp/test/production.yaml",
		"/tmp/test/Makefile",
	}

	// MkdirAll will be called for each file's directory (all in /tmp/test)
	mockFS.On("MkdirAll", "/tmp/test", mock.Anything).Return(nil).Times(len(expectedFiles))
	for _, file := range expectedFiles {
		mockFS.On("WriteFile", file, mock.Anything, mock.Anything).Return(nil)
	}

	err := gen.generateBaseFiles()
	if err != nil {
		t.Fatalf("generateBaseFiles failed: %v", err)
	}

	mockFS.AssertExpectations(t)
}

func TestGenerateAPIFiles(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		apiType      api.Type
		expectedFiles []string
	}{
		{
			name:    "Chi",
			apiType: api.TypeChi,
			expectedFiles: []string{
				"/tmp/test/internal/api/server.go",
				"/tmp/test/internal/json/json.go",
			},
		},
		{
			name:    "gRPC",
			apiType: api.TypeGRPC,
			expectedFiles: []string{
				"/tmp/test/internal/api/server.go",
				"/tmp/test/internal/api/posts_handler.go",
				"/tmp/test/protos/posts/v1/posts.proto",
				"/tmp/test/buf.yaml",
				"/tmp/test/buf.gen.yaml",
				"/tmp/test/cmd/api/main.go",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := mocks.NewFileSystem(t)
			mockLoader := NewMockTemplateLoader()
			
			// Create templates
			templates := map[string]string{
				"chi/server.go.tmpl":           "/tmp/test/internal/api/server.go",
				"chi/json.go.tmpl":             "/tmp/test/internal/json/json.go",
				"grpc/server.go.tmpl":          "/tmp/test/internal/api/server.go",
				"grpc/posts_handler.go.tmpl":    "/tmp/test/internal/api/posts_handler.go",
				"grpc/posts.proto.tmpl":         "/tmp/test/protos/posts/v1/posts.proto",
				"grpc/buf.yaml.tmpl":            "/tmp/test/buf.yaml",
				"grpc/buf.gen.yaml.tmpl":        "/tmp/test/buf.gen.yaml",
				"grpc/main.go.tmpl":             "/tmp/test/cmd/api/main.go",
			}
			
			for tmplPath := range templates {
				tmpl := template.Must(template.New(tmplPath).Parse("{{.ProjectName}}"))
				mockLoader.Templates[tmplPath] = tmpl
			}

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

			// Set up expectations for all expected files
			// MkdirAll will be called for each file's directory
			for _, file := range tt.expectedFiles {
				dir := filepath.Dir(file)
				mockFS.On("MkdirAll", dir, mock.Anything).Return(nil)
				mockFS.On("WriteFile", file, mock.Anything, mock.Anything).Return(nil)
			}

			err := gen.generateAPIFiles(tt.apiType)
			if err != nil {
				t.Fatalf("generateAPIFiles failed: %v", err)
			}

			mockFS.AssertExpectations(t)
		})
	}
}

func TestGenerateDatabaseFiles(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		database     database.Type
		expectedFile string
	}{
		{
			name:         "DynamoDB",
			database:      database.TypeDynamoDB,
			expectedFile: "/tmp/test/internal/database/dynamodb.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := mocks.NewFileSystem(t)
			mockLoader := NewMockTemplateLoader()
			
			tmplPath := "dynamodb/dynamodb.go.tmpl"
			if tt.database == database.TypeDynamoDB {
				tmplPath = "postgres/postgres.go.tmpl"
			}
			
			tmpl := template.Must(template.New(tmplPath).Parse("{{.ProjectName}}"))
			mockLoader.Templates[tmplPath] = tmpl

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

			dir := filepath.Dir(tt.expectedFile)
			mockFS.On("MkdirAll", dir, mock.Anything).Return(nil)
			mockFS.On("WriteFile", tt.expectedFile, mock.Anything, mock.Anything).Return(nil)

			err := gen.generateDatabaseFiles()
			if err != nil {
				t.Fatalf("generateDatabaseFiles failed: %v", err)
			}

			mockFS.AssertExpectations(t)
		})
	}
}

func TestGeneratePostsFiles(t *testing.T) {
	t.Parallel()
	mockFS := mocks.NewFileSystem(t)
	mockLoader := NewMockTemplateLoader()
	
	// Create templates
	templates := []string{
		"posts/post.go.tmpl",
		"posts/service.go.tmpl",
		"posts/dynamodb_table.go.tmpl",
		"posts/dynamodb_table_test.go.tmpl",
		"posts/handlers.go.tmpl",
	}
	for _, tmplPath := range templates {
		tmpl := template.Must(template.New(tmplPath).Parse("{{.ProjectName}}"))
		mockLoader.Templates[tmplPath] = tmpl
	}

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

	// Set up expectations for all expected files
	expectedFiles := []string{
		"/tmp/test/internal/posts/post.go",
		"/tmp/test/internal/posts/service.go",
		"/tmp/test/internal/posts/dynamodb_table.go",
		"/tmp/test/internal/posts/post_table_test.go",
		"/tmp/test/internal/posts/handlers.go",
	}

	for _, file := range expectedFiles {
		dir := filepath.Dir(file)
		mockFS.On("MkdirAll", dir, mock.Anything).Return(nil)
		mockFS.On("WriteFile", file, mock.Anything, mock.Anything).Return(nil)
	}

	err := gen.generatePostsFiles()
	if err != nil {
		t.Fatalf("generatePostsFiles failed: %v", err)
	}

	mockFS.AssertExpectations(t)
}

func TestGenerateFeatureFiles(t *testing.T) {
	t.Parallel()
	mockFS := mocks.NewFileSystem(t)
	mockLoader := NewMockTemplateLoader()
	
	// Create templates
	metricsTmpl := template.Must(template.New("metrics/metrics.go.tmpl").Parse("{{.ProjectName}}"))
	mockLoader.Templates["metrics/metrics.go.tmpl"] = metricsTmpl
	
	posthogTmpl := template.Must(template.New("posthog/posthog.go.tmpl").Parse("{{.ProjectName}}"))
	mockLoader.Templates["posthog/posthog.go.tmpl"] = posthogTmpl
	
	authTmpl := template.Must(template.New("auth/jwt.go.tmpl").Parse("{{.ProjectName}}"))
	mockLoader.Templates["auth/jwt.go.tmpl"] = authTmpl

	config := config.ProjectConfig{
		ProjectName: "test-service",
		ModulePath:  "github.com/test/service",
		OutputDir:   "/tmp/test",
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
	gen := NewGeneratorWithDeps(config, mockFS, mockLoader)

	// Set up expectations
	// Metrics is always generated
	mockFS.On("MkdirAll", "/tmp/test/internal/metrics", mock.Anything).Return(nil)
	mockFS.On("WriteFile", "/tmp/test/internal/metrics/metrics.go", mock.Anything, mock.Anything).Return(nil)
	// Auth is generated when FeatureAuth is enabled
	mockFS.On("MkdirAll", "/tmp/test/internal/auth", mock.Anything).Return(nil)
	mockFS.On("WriteFile", "/tmp/test/internal/auth/jwt.go", mock.Anything, mock.Anything).Return(nil)

	err := gen.generateFeatureFiles()
	if err != nil {
		t.Fatalf("generateFeatureFiles failed: %v", err)
	}

	mockFS.AssertExpectations(t)
}

