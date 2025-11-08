package generator

import (
	"text/template"

	"github.com/anmho/create-go-service/internal/generator/api"
	"github.com/anmho/create-go-service/internal/generator/config"
	"github.com/anmho/create-go-service/internal/generator/database"
	"github.com/anmho/create-go-service/internal/generator/deployment"
)

// loadTemplate is deprecated - use templateLoader.LoadTemplate instead
// Kept for backward compatibility with existing tests
func (g *Generator) loadTemplate(path string) (*template.Template, error) {
	return g.templateLoader.LoadTemplate(path)
}

func (g *Generator) getTemplateData() map[string]interface{} {
	// Metrics and hot reload are always enabled (mandatory)
	hasMetrics := true
	hasHotReload := true
	hasAuth := false
	hasPostHog := false

	for _, feature := range g.config.Features {
		switch feature {
		case config.FeatureAuth:
			hasAuth = true
		case config.FeaturePostHog:
			hasPostHog = true
		}
	}

	// Determine API types
	hasChi := false
	hasHuma := false
	hasGRPC := false
	if len(g.config.API.Types) > 0 {
		switch g.config.API.Types[0] {
		case api.TypeChi:
			hasChi = true
		case api.TypeHuma:
			hasHuma = true
		case api.TypeGRPC:
			hasGRPC = true
		}
	}

	return map[string]interface{}{
		"ProjectName":         g.config.ProjectName,
		"ModulePath":          g.config.ModulePath,
		"OutputDir":           g.config.OutputDir,
		"APITypes":            g.config.API.Types,
		"Database":            string(g.config.Database.Type),
		"Features":            g.config.Features,
		"Deployment":          string(g.config.Deployment.Type),
		"HasChi":              hasChi,
		"HasHuma":             hasHuma,
		"HasGRPC":             hasGRPC,
		"HasDynamoDB":         g.config.Database.Type == database.TypeDynamoDB,
		"HasPostgres":         g.config.Database.Type == database.TypePostgres,
		"HasMetrics":          hasMetrics,
		"HasPostHog":          hasPostHog,
		"HasAuth":             hasAuth,
		"HasHotReload":        hasHotReload,
		"HasFly":              g.config.Deployment.Type == deployment.TypeFly,
		"JWTSecret":           g.config.Auth.JWTSecret,
		"PostHogAPIKey":      g.config.PostHog.APIKey,
		"PostHogHost":        g.config.PostHog.Host,
	}
}
