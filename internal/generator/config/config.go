package config

import (
	"github.com/anmho/create-go-service/internal/generator/api"
	"github.com/anmho/create-go-service/internal/generator/database"
	"github.com/anmho/create-go-service/internal/generator/deployment"
)

// Feature represents an optional feature
type Feature string

const (
	FeatureAuth    Feature = "auth"    // Optional: JWT authentication
	FeaturePostHog Feature = "posthog" // Optional: PostHog event tracking
	// Note: Metrics and hot reload are always enabled, not optional features
)

// ProjectConfig holds all project configuration, grouped by function
type ProjectConfig struct {
	ProjectName string
	ModulePath  string
	OutputDir   string
	Features    []Feature // Optional features (e.g., auth)
	
	// Optional feature configurations
	Auth    AuthConfig
	PostHog PostHogConfig
	
	// Grouped configurations
	API        api.Config
	Database   database.Config
	Deployment deployment.Config
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret string // JWT secret for decoding JWTs (e.g., from Supabase Auth)
}

// PostHogConfig holds PostHog configuration
type PostHogConfig struct {
	APIKey string
	Host   string
}

