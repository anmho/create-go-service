package generator

import (
	"embed"
	"path/filepath"
	"text/template"
)

//go:embed all:templates
var templatesFS embed.FS

// TemplateLoader abstracts template loading for testability
type TemplateLoader interface {
	LoadTemplate(path string) (*template.Template, error)
}

// EmbeddedTemplateLoader loads templates from embedded filesystem
type EmbeddedTemplateLoader struct{}

func NewEmbeddedTemplateLoader() *EmbeddedTemplateLoader {
	return &EmbeddedTemplateLoader{}
}

func (l *EmbeddedTemplateLoader) LoadTemplate(path string) (*template.Template, error) {
	// Read from embedded file system
	fsPath := filepath.Join("templates", path)
	content, err := templatesFS.ReadFile(fsPath)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New(filepath.Base(path)).Parse(string(content))
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

// MockTemplateLoader for testing
type MockTemplateLoader struct {
	Templates map[string]*template.Template
	LoadError error
}

func NewMockTemplateLoader() *MockTemplateLoader {
	return &MockTemplateLoader{
		Templates: make(map[string]*template.Template),
	}
}

func (m *MockTemplateLoader) LoadTemplate(path string) (*template.Template, error) {
	if m.LoadError != nil {
		return nil, m.LoadError
	}
	if tmpl, ok := m.Templates[path]; ok {
		return tmpl, nil
	}
	// Return a simple template that just outputs the path for testing
	tmpl := template.Must(template.New(path).Parse("{{.ProjectName}}"))
	m.Templates[path] = tmpl
	return tmpl, nil
}

