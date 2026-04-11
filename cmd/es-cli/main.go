package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kienlt/es-cli/internal/auth"
	"github.com/kienlt/es-cli/internal/es"
	"github.com/kienlt/es-cli/internal/tui"
)

const defaultESURL = "http://localhost:9200"

func main() {
	authPath := auth.DefaultAuthPath()

	cred, err := auth.LoadAuth(authPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Auth error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Please create %s with format: {\"username\":\"password\"}\n", authPath)
		os.Exit(1)
	}

	esURL := defaultESURL
	if envURL := os.Getenv("ES_URL"); envURL != "" {
		esURL = envURL
	}

	client := es.NewClient(esURL, cred.Username, cred.Password)

	app := tui.NewApp(client, esURL)

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
