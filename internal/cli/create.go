package cli

import (
	"github.com/anmho/create-go-service/internal/tui"
)

// Execute runs the CLI application
func Execute() error {
	app := tui.NewApp()
	return app.Run()
}

