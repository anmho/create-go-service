package generator

import (
	"bytes"
	"fmt"
	"path/filepath"
)

func (g *Generator) generateFile(outputPath, templatePath string, data interface{}) error {
	tmpl, err := g.templateLoader.LoadTemplate(templatePath)
	if err != nil {
		return fmt.Errorf("failed to load template %s: %w", templatePath, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template %s: %w", templatePath, err)
	}

	// Create full output path
	fullPath := filepath.Join(g.config.OutputDir, outputPath)

	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := g.fs.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := g.fs.WriteFile(fullPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
