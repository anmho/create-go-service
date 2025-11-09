package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/anmho/create-go-service/internal/generator"
	"github.com/anmho/create-go-service/internal/generator/api"
	"github.com/anmho/create-go-service/internal/generator/config"
	"github.com/anmho/create-go-service/internal/generator/database"
	"github.com/anmho/create-go-service/internal/generator/deployment"
	"github.com/anmho/create-go-service/internal/tui"
	"github.com/spf13/cobra"
)

var (
	// Version is set during build via ldflags
	Version = "dev"
	// Commit is set during build via ldflags
	Commit = "unknown"
	// BuildTime is set during build via ldflags
	BuildTime = "unknown"
)

// Execute runs the CLI application
func Execute() error {
	var (
		projectName    string
		modulePath     string
		outputDir      string
		apiType        string
		databaseType   string
		features       string
		jwtSecret      string
		posthogAPIKey  string
		posthogHost    string
		deploymentType string
	)

	rootCmd := &cobra.Command{
		Use:   "create-go-service",
		Short: "Generate a new Go service project",
		Long:  "Generate a new Go service project with customizable API, database, and features",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if any flags were provided
			flagsProvided := projectName != "" || modulePath != "" || outputDir != "" ||
				apiType != "" || databaseType != "" || features != "" ||
				jwtSecret != "" || posthogAPIKey != "" || posthogHost != "" ||
				deploymentType != ""

			// If flags provided, use direct mode
			if flagsProvided {
				return generateDirect(projectName, modulePath, outputDir, apiType, databaseType, features, jwtSecret, posthogAPIKey, posthogHost, deploymentType)
			}

			// Otherwise, use TUI
			app := tui.NewApp()
			return app.Run()
		},
	}

	// Version command
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("create-go-service version %s\n", Version)
			fmt.Printf("  commit:     %s\n", Commit)
			fmt.Printf("  build time: %s\n", BuildTime)
		},
	}
	rootCmd.AddCommand(versionCmd)

	// Flags
	rootCmd.Flags().StringVar(&projectName, "project-name", "", "Project name")
	rootCmd.Flags().StringVar(&modulePath, "module-path", "", "Go module path (e.g., github.com/user/project)")
	rootCmd.Flags().StringVar(&outputDir, "output-dir", "", "Output directory (default: ./<project-name>)")
	rootCmd.Flags().StringVar(&apiType, "api", "", "API type: chi, grpc, or huma")
	rootCmd.Flags().StringVar(&databaseType, "database", "", "Database type: dynamodb or postgres")
	rootCmd.Flags().StringVar(&features, "features", "", "Comma-separated features: auth,posthog")
	rootCmd.Flags().StringVar(&jwtSecret, "jwt-secret", "", "JWT secret (required if auth feature is enabled)")
	rootCmd.Flags().StringVar(&posthogAPIKey, "posthog-api-key", "", "PostHog API key (required if posthog feature is enabled)")
	rootCmd.Flags().StringVar(&posthogHost, "posthog-host", "", "PostHog host (required if posthog feature is enabled)")
	rootCmd.Flags().StringVar(&deploymentType, "deployment", "", "Deployment type: fly")

	// Add version flag
	rootCmd.Flags().BoolP("version", "v", false, "Show version information")

	// Handle version flag before running
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if v, _ := cmd.Flags().GetBool("version"); v {
			fmt.Printf("create-go-service version %s\n", Version)
			fmt.Printf("  commit:     %s\n", Commit)
			fmt.Printf("  build time: %s\n", BuildTime)
			os.Exit(0)
		}
		return nil
	}

	return rootCmd.Execute()
}

func generateDirect(projectName, modulePath, outputDir, apiType, databaseType, features, jwtSecret, posthogAPIKey, posthogHost, deploymentType string) error {
	// Validate required fields
	if projectName == "" {
		return fmt.Errorf("--project-name is required")
	}
	if modulePath == "" {
		return fmt.Errorf("--module-path is required")
	}
	if apiType == "" {
		return fmt.Errorf("--api is required (chi, grpc, or huma)")
	}
	if databaseType == "" {
		return fmt.Errorf("--database is required (dynamodb or postgres)")
	}

	// Set default output directory
	if outputDir == "" {
		outputDir = "./" + projectName
	}

	// Parse API type
	var apiTypes []api.Type
	switch strings.ToLower(apiType) {
	case "chi":
		apiTypes = []api.Type{api.TypeChi}
	case "grpc":
		apiTypes = []api.Type{api.TypeGRPC}
	case "huma":
		apiTypes = []api.Type{api.TypeHuma}
	default:
		return fmt.Errorf("invalid API type: %s (must be chi, grpc, or huma)", apiType)
	}

	// Parse database type
	var dbType database.Type
	switch strings.ToLower(databaseType) {
	case "dynamodb":
		dbType = database.TypeDynamoDB
	case "postgres":
		dbType = database.TypePostgres
	default:
		return fmt.Errorf("invalid database type: %s (must be dynamodb or postgres)", databaseType)
	}

	// Parse features
	var featureList []config.Feature
	if features != "" {
		for _, f := range strings.Split(features, ",") {
			f = strings.TrimSpace(f)
			switch strings.ToLower(f) {
			case "auth":
				featureList = append(featureList, config.FeatureAuth)
			case "posthog":
				featureList = append(featureList, config.FeaturePostHog)
			default:
				return fmt.Errorf("invalid feature: %s (must be auth or posthog)", f)
			}
		}
	}

	// Validate feature requirements
	for _, f := range featureList {
		switch f {
		case config.FeatureAuth:
			if jwtSecret == "" {
				return fmt.Errorf("--jwt-secret is required when --features includes auth")
			}
		case config.FeaturePostHog:
			if posthogAPIKey == "" {
				return fmt.Errorf("--posthog-api-key is required when --features includes posthog")
			}
			if posthogHost == "" {
				return fmt.Errorf("--posthog-host is required when --features includes posthog")
			}
		}
	}

	// Parse deployment type (required)
	if deploymentType == "" {
		return fmt.Errorf("deployment type is required (use --deployment fly)")
	}
	var depType deployment.Type
	switch strings.ToLower(deploymentType) {
	case "fly":
		depType = deployment.TypeFly
	default:
		return fmt.Errorf("invalid deployment type: %s (must be fly)", deploymentType)
	}

	// Build config
	cfg := config.ProjectConfig{
		ProjectName: projectName,
		ModulePath:  modulePath,
		OutputDir:   outputDir,
		Features:    featureList,
		Auth: config.AuthConfig{
			JWTSecret: jwtSecret,
		},
		PostHog: config.PostHogConfig{
			APIKey: posthogAPIKey,
			Host:   posthogHost,
		},
		API: api.Config{
			Types: apiTypes,
		},
		Database: database.Config{
			Type: dbType,
		},
		Deployment: deployment.Config{
			Type: depType,
		},
	}

	// Generate project
	gen := generator.NewGenerator(cfg)
	if err := gen.Generate(); err != nil {
		return fmt.Errorf("failed to generate project: %w", err)
	}

	fmt.Printf("✓ Project generated successfully!\n")
	fmt.Printf("  Project: %s\n", projectName)
	fmt.Printf("  Module:  %s\n", modulePath)
	fmt.Printf("  Output:  %s\n", outputDir)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  cd %s\n", outputDir)
	fmt.Printf("  make deps\n")
	fmt.Printf("  make build\n")

	return nil
}
