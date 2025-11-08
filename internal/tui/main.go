package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/anmho/create-go-service/internal/generator"
	"github.com/anmho/create-go-service/internal/generator/api"
	"github.com/anmho/create-go-service/internal/generator/config"
	"github.com/anmho/create-go-service/internal/generator/database"
	"github.com/anmho/create-go-service/internal/generator/deployment"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type App struct {
	model *Model
}

func NewApp() *App {
	return &App{
		model: NewModel(),
	}
}

func (a *App) Run() error {
	p := tea.NewProgram(a.model)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
	return nil
}

type Model struct {
	step                Step
	projectName         textInputModel
	modulePath          textInputModel
	outputDir           textInputModel
	apiSelect           singleSelectModel
	databaseSelect      singleSelectModel
	featuresSelect      multiSelectModel
	jwtSecret           textInputModel
	posthogAPIKey       textInputModel
	posthogHost         textInputModel
	deploymentSelect    singleSelectModel
	spinner             spinner.Model
	err                 error
	generating          bool
	generationSteps     []string
	currentStepIdx      int
}

type Step int

const (
	StepWelcome Step = iota
	StepProjectName
	StepModulePath
	StepOutputDir
	StepAPISelection
	StepDatabaseSelection
	StepFeaturesSelection
	StepJWTSecret
	StepPostHogAPIKey
	StepPostHogHost
	StepDeploymentSelection
	StepReview
	StepGenerating
	StepComplete
)

func NewModel() *Model {
	apiOptions := []string{
		"REST with Chi",
		"REST with Huma (includes Swagger)",
		"gRPC with ConnectRPC",
	}

	databaseOptions := []string{
		"DynamoDB",
		"PostgreSQL",
	}

	featureOptions := []string{
		"JWT Auth (Supabase/Clerk)",
		"PostHog (Event Tracking)",
	}

	deploymentOptions := []string{
		"Fly.io",
		// Future: Add more deployment options here (e.g., "AWS ECS", "Kubernetes")
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	return &Model{
		step:                StepWelcome,
		projectName:         newTextInput("Project Name", "my-service"),
		modulePath:          newTextInput("Module Path", "github.com/user/service"),
		outputDir:           newTextInput("Output Directory", "./my-service"),
		apiSelect:           newSingleSelect("Select API Framework", apiOptions, 0),
		databaseSelect:      newSingleSelect("Select Database Type", databaseOptions, 0),
		featuresSelect:      newMultiSelect("Select Features (space to select, enter to continue)", featureOptions),
		jwtSecret:           newTextInput("JWT Secret", "your-jwt-secret"),
		posthogAPIKey:       newTextInput("PostHog API Key", "phc_..."),
		posthogHost:         newTextInput("PostHog Host", "https://app.posthog.com"),
		deploymentSelect:    newSingleSelect("Select Deployment Type", deploymentOptions, 0),
		spinner:             s,
		generationSteps: []string{
			"Creating directory structure...",
			"Generating base files...",
			"Setting up API framework...",
			"Configuring database...",
			"Adding features...",
			"Setting up deployment configs...",
			"Finalizing project...",
		},
		currentStepIdx: 0,
	}
}

func (m *Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case stepProgressMsg:
		m.currentStepIdx = int(msg)
		return m, nil
	case tea.KeyMsg:
		if m.generating {
			// Don't allow input while generating
			return m, nil
		}
		
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			if m.step > StepWelcome {
				m.step--
				return m, nil
			}
			return m, tea.Quit
		}

		switch m.step {
		case StepWelcome:
			if msg.String() == "enter" {
				m.step = StepProjectName
				return m, nil
			}
		case StepProjectName:
			m.projectName, cmd = m.projectName.Update(msg)
			if msg.String() == "enter" && m.projectName.value != "" {
				// Auto-set output directory based on project name
				m.outputDir.SetValue("./" + m.projectName.value)
				m.step = StepModulePath
			}
			return m, cmd
		case StepModulePath:
			m.modulePath, cmd = m.modulePath.Update(msg)
			if msg.String() == "enter" && m.modulePath.value != "" {
				m.step = StepOutputDir
			}
			return m, cmd
		case StepOutputDir:
			m.outputDir, cmd = m.outputDir.Update(msg)
			if msg.String() == "enter" && m.outputDir.value != "" {
				m.step = StepAPISelection
			}
			return m, cmd
		case StepAPISelection:
			m.apiSelect, cmd = m.apiSelect.Update(msg)
			if msg.String() == "enter" {
				m.step = StepDatabaseSelection
			}
			return m, cmd
		case StepDatabaseSelection:
			m.databaseSelect, cmd = m.databaseSelect.Update(msg)
			if msg.String() == "enter" {
				m.step = StepFeaturesSelection
			}
			return m, cmd
		case StepFeaturesSelection:
			m.featuresSelect, cmd = m.featuresSelect.Update(msg)
			if msg.String() == "enter" {
				// Check which features are selected
				selectedFeatures := m.featuresSelect.GetSelected()
				hasAuth := false
				hasPostHog := false
				for _, s := range selectedFeatures {
					if strings.Contains(s, "JWT Auth") {
						hasAuth = true
					}
					if strings.Contains(s, "PostHog") {
						hasPostHog = true
					}
				}
				// Prompt for JWT secret if auth is selected
				if hasAuth {
					m.step = StepJWTSecret
				} else if hasPostHog {
					m.step = StepPostHogAPIKey
				} else {
					m.step = StepDeploymentSelection
				}
			}
			return m, cmd
		case StepJWTSecret:
			m.jwtSecret, cmd = m.jwtSecret.Update(msg)
			if msg.String() == "enter" {
				// Check if PostHog is also selected
				selectedFeatures := m.featuresSelect.GetSelected()
				hasPostHog := false
				for _, s := range selectedFeatures {
					if strings.Contains(s, "PostHog") {
						hasPostHog = true
						break
					}
				}
				if hasPostHog {
					m.step = StepPostHogAPIKey
				} else {
					m.step = StepDeploymentSelection
				}
			}
			return m, cmd
		case StepPostHogAPIKey:
			m.posthogAPIKey, cmd = m.posthogAPIKey.Update(msg)
			if msg.String() == "enter" {
				m.step = StepPostHogHost
			}
			return m, cmd
		case StepPostHogHost:
			m.posthogHost, cmd = m.posthogHost.Update(msg)
			if msg.String() == "enter" {
				m.step = StepDeploymentSelection
			}
			return m, cmd
		case StepDeploymentSelection:
			m.deploymentSelect, cmd = m.deploymentSelect.Update(msg)
			if msg.String() == "enter" {
				m.step = StepReview
			}
			return m, cmd
		case StepReview:
			if msg.String() == "enter" {
				m.step = StepGenerating
				m.generating = true
				m.currentStepIdx = 0
				return m, tea.Batch(m.spinner.Tick, m.generate())
			}
		case StepComplete:
			if msg.String() == "enter" {
				return m, tea.Quit
			}
		}

	case GenerationCompleteMsg:
		m.step = StepComplete
		m.generating = false
		return m, nil
	case GenerationErrorMsg:
		m.err = msg.Err
		m.generating = false
		return m, nil
	}

	return m, nil
}

func (m *Model) generate() tea.Cmd {
	return func() tea.Msg {
		// Map API selection to type
		var apiType api.Type
		selected := m.apiSelect.GetSelected()
		if strings.Contains(selected, "Chi") {
			apiType = api.TypeChi
		} else if strings.Contains(selected, "Huma") {
			apiType = api.TypeHuma
		} else if strings.Contains(selected, "gRPC") {
			apiType = api.TypeGRPC
		}

		// Map database selection
		var dbType database.Type
		selectedDB := m.databaseSelect.GetSelected()
		if strings.Contains(selectedDB, "DynamoDB") {
			dbType = database.TypeDynamoDB
		} else if strings.Contains(selectedDB, "PostgreSQL") {
			dbType = database.TypePostgres
		}

		// Map features (metrics and hot reload are always enabled, not in Feature enum)
		var features []config.Feature
		selectedFeatures := m.featuresSelect.GetSelected()
		for _, s := range selectedFeatures {
			if strings.Contains(s, "JWT Auth") {
				features = append(features, config.FeatureAuth)
			}
			if strings.Contains(s, "PostHog") {
				features = append(features, config.FeaturePostHog)
			}
		}

		cfg := config.ProjectConfig{
			ProjectName: m.projectName.value,
			ModulePath:  m.modulePath.value,
			OutputDir:   m.outputDir.value,
			Features:    features,
			Auth: config.AuthConfig{
				JWTSecret: m.jwtSecret.value,
			},
			PostHog: config.PostHogConfig{
				APIKey: m.posthogAPIKey.value,
				Host:   m.posthogHost.value,
			},
			API: api.Config{
				Types: []api.Type{apiType},
			},
			Database: database.Config{
				Type: dbType,
			},
			Deployment: deployment.Config{
				Type: deployment.TypeFly,
			},
		}

		// Simulate progress steps
		time.Sleep(200 * time.Millisecond)

		gen := generator.NewGenerator(cfg)
		if err := gen.Generate(); err != nil {
			return GenerationErrorMsg{Err: err}
		}

		// Add a small delay to show completion
		time.Sleep(300 * time.Millisecond)

		return GenerationCompleteMsg{}
	}
}

type GenerationCompleteMsg struct{}

type GenerationErrorMsg struct {
	Err error
}

type stepProgressMsg int

func simulateProgress() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(300 * time.Millisecond)
		return stepProgressMsg(0)
	}
}

func (m *Model) View() string {
	if m.generating {
		return m.renderGenerating()
	}

	if m.err != nil {
		return m.renderError()
	}

	switch m.step {
	case StepWelcome:
		return m.renderWelcome()
	case StepProjectName:
		return m.renderProjectName()
	case StepModulePath:
		return m.renderModulePath()
	case StepOutputDir:
		return m.renderOutputDir()
	case StepAPISelection:
		return m.renderAPISelection()
	case StepDatabaseSelection:
		return m.renderDatabaseSelection()
	case StepFeaturesSelection:
		return m.renderFeaturesSelection()
	case StepJWTSecret:
		return m.renderJWTSecret()
	case StepPostHogAPIKey:
		return m.renderPostHogAPIKey()
	case StepPostHogHost:
		return m.renderPostHogHost()
	case StepDeploymentSelection:
		return m.renderDeploymentSelection()
	case StepReview:
		return m.renderReview()
	case StepGenerating:
		return m.renderGenerating()
	case StepComplete:
		return m.renderComplete()
	default:
		return "Unknown step"
	}
}

func (m *Model) renderWelcome() string {
	logo := titleStyle.Render("🚀 create-go-service")
	subtitle := subtitleStyle.Render("Scaffold Go Microservice Boilerplate")

	description := lipgloss.NewStyle().
		Foreground(whiteColor).
		MarginTop(2).
		MarginBottom(1).
		Render(`Welcome! This tool will help you create a production-ready
Go microservice with:

  • REST API (Chi or Huma)
  • gRPC with ConnectRPC
  • Database integration (DynamoDB or PostgreSQL)
  • Metrics instrumentation
  • Deployment configs (Fly.io)
  • Hot reload with wgo`)

	help := helpStyle.Render("\nPress Enter to continue...")

	return lipgloss.JoinVertical(lipgloss.Left, logo, subtitle, description, help)
}

func (m *Model) renderProjectName() string {
	title := titleStyle.Render("📝 Project Configuration")
	form := m.projectName.View()
	help := helpStyle.Render("\n↑/↓: Navigate  Enter: Continue  Esc: Back  Ctrl+C: Quit")

	return lipgloss.JoinVertical(lipgloss.Left, title, "", form, help)
}

func (m *Model) renderModulePath() string {
	title := titleStyle.Render("📦 Module Path")
	form := m.modulePath.View()
	help := helpStyle.Render("\n↑/↓: Navigate  Enter: Continue  Esc: Back  Ctrl+C: Quit")

	return lipgloss.JoinVertical(lipgloss.Left, title, "", form, help)
}

func (m *Model) renderOutputDir() string {
	title := titleStyle.Render("📁 Output Directory")
	subtitle := lipgloss.NewStyle().
		Foreground(grayColor).
		Italic(true).
		Render("(auto-filled based on project name, you can edit)")
	form := m.outputDir.View()
	help := helpStyle.Render("\n↑/↓: Navigate  Enter: Continue  Esc: Back  Ctrl+C: Quit")

	return lipgloss.JoinVertical(lipgloss.Left, title, subtitle, "", form, help)
}

func (m *Model) renderAPISelection() string {
	title := titleStyle.Render("🔌 API Framework Selection")
	form := m.apiSelect.View()
	help := helpStyle.Render("\n↑/↓: Navigate  Space/Enter: Select  Esc: Back  Ctrl+C: Quit")

	return lipgloss.JoinVertical(lipgloss.Left, title, "", form, help)
}

func (m *Model) renderDatabaseSelection() string {
	title := titleStyle.Render("💾 Database Type Selection")
	form := m.databaseSelect.View()
	help := helpStyle.Render("\n↑/↓: Navigate  Space/Enter: Select  Esc: Back  Ctrl+C: Quit")

	return lipgloss.JoinVertical(lipgloss.Left, title, "", form, help)
}

func (m *Model) renderFeaturesSelection() string {
	title := titleStyle.Render("✨ Features Selection")
	form := m.featuresSelect.View()
	help := helpStyle.Render("\n↑/↓: Navigate  Space: Toggle  Enter: Continue  Esc: Back  Ctrl+C: Quit")

	return lipgloss.JoinVertical(lipgloss.Left, title, "", form, help)
}

func (m *Model) renderJWTSecret() string {
	title := titleStyle.Render("🔐 JWT Authentication")
	subtitle := lipgloss.NewStyle().
		Foreground(grayColor).
		Italic(true).
		Render("Enter your JWT secret for decoding JWTs (e.g., from Supabase Auth or Clerk)")
	form := m.jwtSecret.View()
	help := helpStyle.Render("\n↑/↓: Navigate  Enter: Continue  Esc: Back  Ctrl+C: Quit")

	return lipgloss.JoinVertical(lipgloss.Left, title, subtitle, "", form, help)
}

func (m *Model) renderPostHogAPIKey() string {
	title := titleStyle.Render("📊 PostHog Configuration")
	subtitle := lipgloss.NewStyle().
		Foreground(grayColor).
		Italic(true).
		Render("Enter your PostHog API key (get it from https://app.posthog.com/project/settings)")
	form := m.posthogAPIKey.View()
	help := helpStyle.Render("\n↑/↓: Navigate  Enter: Continue  Esc: Back  Ctrl+C: Quit")

	return lipgloss.JoinVertical(lipgloss.Left, title, subtitle, "", form, help)
}

func (m *Model) renderPostHogHost() string {
	title := titleStyle.Render("🌐 PostHog Host")
	subtitle := lipgloss.NewStyle().
		Foreground(grayColor).
		Italic(true).
		Render("Enter your PostHog host URL (default: https://app.posthog.com)")
	form := m.posthogHost.View()
	help := helpStyle.Render("\n↑/↓: Navigate  Enter: Continue  Esc: Back  Ctrl+C: Quit")

	return lipgloss.JoinVertical(lipgloss.Left, title, subtitle, "", form, help)
}

func (m *Model) renderDeploymentSelection() string {
	title := titleStyle.Render("🚀 Deployment Type Selection")
	form := m.deploymentSelect.View()
	help := helpStyle.Render("\n↑/↓: Navigate  Space/Enter: Select  Esc: Back  Ctrl+C: Quit")

	return lipgloss.JoinVertical(lipgloss.Left, title, "", form, help)
}

func (m *Model) renderReview() string {
	title := titleStyle.Render("📋 Review Configuration")

	var sections []string
	sections = append(sections, title, "")

	sections = append(sections, labelStyle.Render("Project Name:     ")+valueStyle.Render(m.projectName.value))
	sections = append(sections, labelStyle.Render("Module Path:      ")+valueStyle.Render(m.modulePath.value))
	sections = append(sections, labelStyle.Render("Output Dir:       ")+valueStyle.Render(m.outputDir.value))
	sections = append(sections, labelStyle.Render("API Framework:    ")+valueStyle.Render(m.apiSelect.GetSelected()))
	sections = append(sections, labelStyle.Render("Database:         ")+valueStyle.Render(m.databaseSelect.GetSelected()))

	// Build features list (always include metrics and hot reload)
	allFeatures := []string{"Metrics (Prometheus)", "Hot reload (wgo)"}
	selectedFeatures := m.featuresSelect.GetSelected()
	hasAuth := false
	hasPostHog := false
	for _, s := range selectedFeatures {
		if strings.Contains(s, "JWT Auth") {
			hasAuth = true
		}
		if strings.Contains(s, "PostHog") {
			hasPostHog = true
		}
		allFeatures = append(allFeatures, s)
	}
	sections = append(sections, labelStyle.Render("Features:         ")+valueStyle.Render(strings.Join(allFeatures, ", ")))

	// Show JWT secret if auth is selected
	if hasAuth {
		sections = append(sections, labelStyle.Render("JWT Secret:       ")+valueStyle.Render(m.jwtSecret.value))
	}

	// Show PostHog config if selected
	if hasPostHog {
		sections = append(sections, labelStyle.Render("PostHog API Key:  ")+valueStyle.Render(m.posthogAPIKey.value))
		sections = append(sections, labelStyle.Render("PostHog Host:     ")+valueStyle.Render(m.posthogHost.value))
	}

	sections = append(sections, labelStyle.Render("Deployment:       ")+valueStyle.Render(m.deploymentSelect.GetSelected()))

	help := helpStyle.Render("\nEnter: Generate  Esc: Back  Ctrl+C: Quit")

	return lipgloss.JoinVertical(lipgloss.Left, append(sections, "", help)...)
}

func (m *Model) renderGenerating() string {
	title := titleStyle.Render("⚙️  Generating project...")
	
	var steps []string
	for i, step := range m.generationSteps {
		if i < m.currentStepIdx {
			// Completed step
			steps = append(steps, successStyle.Render("✓ ")+lipgloss.NewStyle().Foreground(grayColor).Render(step))
		} else if i == m.currentStepIdx {
			// Current step with spinner
			steps = append(steps, m.spinner.View()+" "+lipgloss.NewStyle().Foreground(whiteColor).Render(step))
		} else {
			// Pending step
			steps = append(steps, lipgloss.NewStyle().Foreground(grayColor).Render("  "+step))
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, steps...)
	
	return lipgloss.JoinVertical(lipgloss.Left, title, "", content)
}

func (m *Model) renderComplete() string {
	title := successStyle.Render("✅ Project generated successfully!")
	
	var completedSteps []string
	for _, step := range m.generationSteps {
		completedSteps = append(completedSteps, successStyle.Render("✓ ")+lipgloss.NewStyle().Foreground(grayColor).Render(step))
	}
	
	stepsView := lipgloss.JoinVertical(lipgloss.Left, completedSteps...)
	
	location := lipgloss.NewStyle().
		Foreground(whiteColor).
		Bold(true).
		MarginTop(2).
		Render(fmt.Sprintf("📁 Location: %s", m.outputDir.value))
	
	// Build next steps based on selections
	var nextStepsList []string
	nextStepsList = append(nextStepsList, fmt.Sprintf("cd %s", m.outputDir.value))
	nextStepsList = append(nextStepsList, "make deps    # Install dependencies")
	nextStepsList = append(nextStepsList, "make build   # Build the project")
	nextStepsList = append(nextStepsList, "make run     # Run locally")
	nextStepsList = append(nextStepsList, "make test    # Run tests")
	
	// Add database-specific steps
	databaseSelected := m.databaseSelect.GetSelected()
	if strings.Contains(databaseSelected, "PostgreSQL") {
		nextStepsList = append(nextStepsList, "make atlas-init  # Initialize migrations")
		nextStepsList = append(nextStepsList, "make migrate      # Run migrations")
	} else if strings.Contains(databaseSelected, "DynamoDB") {
		nextStepsList = append(nextStepsList, "make terraform    # Provision infrastructure")
	}
	
	// Add deployment steps
	deploymentSelected := m.deploymentSelect.GetSelected()
	if strings.Contains(deploymentSelected, "Fly.io") {
		nextStepsList = append(nextStepsList, "make deploy       # Deploy to Fly.io")
		nextStepsList = append(nextStepsList, "make deploy-local # Deploy with local build")
	}
	
	nextStepsText := "Next steps:\n"
	for _, step := range nextStepsList {
		nextStepsText += "  " + step + "\n"
	}
	
	nextSteps := lipgloss.NewStyle().
		Foreground(secondaryColor).
		MarginTop(1).
		Render(nextStepsText)
	
	help := helpStyle.Render("\nPress Enter to exit...")

	return lipgloss.JoinVertical(lipgloss.Left, title, "", stepsView, location, nextSteps, help)
}

func (m *Model) renderError() string {
	title := errorStyle.Render("❌ Error generating project")
	errorMsg := errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	help := helpStyle.Render("\nPress Ctrl+C to exit...")

	return lipgloss.JoinVertical(lipgloss.Left, title, "", errorMsg, help)
}
