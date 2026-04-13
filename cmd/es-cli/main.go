package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kienlt/es-cli/internal/auth"
	"github.com/kienlt/es-cli/internal/es"
	"github.com/kienlt/es-cli/internal/tui"
	"github.com/kienlt/es-cli/internal/tui/clusterselect"
)

var version = "dev"

func main() {
	for _, arg := range os.Args[1:] {
		if arg == "--version" || arg == "-v" {
			fmt.Println(version)
			return
		}
	}

	authPath := auth.DefaultAuthPath()

	configs, err := auth.LoadAuth(authPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Auth error: %v\n", err)
		os.Exit(1)
	}

	// Parse --cluster flag
	clusterName := parseClusterFlag()

	var cluster auth.ClusterConfig

	switch {
	case clusterName != "":
		// Direct connect via --cluster flag
		c := findCluster(configs, clusterName)
		if c == nil {
			fmt.Fprintf(os.Stderr, "Cluster %q not found in %s\n", clusterName, authPath)
			fmt.Fprintf(os.Stderr, "Available clusters: ")
			for i, cfg := range configs {
				if i > 0 {
					fmt.Fprintf(os.Stderr, ", ")
				}
				fmt.Fprintf(os.Stderr, "%s", cfg.Name)
			}
			fmt.Fprintln(os.Stderr)
			os.Exit(1)
		}
		cluster = *c

	case len(configs) == 1:
		// Single cluster, connect directly
		cluster = configs[0]

	default:
		// Show cluster selection UI
		selected := runClusterSelect(configs)
		if selected == nil {
			os.Exit(0)
		}
		cluster = *selected
	}

	client := es.NewClient(cluster.URL, cluster.Username, cluster.Password)
	app := tui.NewApp(client, cluster.URL, cluster.Name)

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func parseClusterFlag() string {
	for i, arg := range os.Args[1:] {
		if arg == "--cluster" && i+1 < len(os.Args[1:]) {
			return os.Args[i+2]
		}
	}
	return ""
}

func findCluster(configs []auth.ClusterConfig, name string) *auth.ClusterConfig {
	for i := range configs {
		if configs[i].Name == name {
			return &configs[i]
		}
	}
	return nil
}

func runClusterSelect(configs []auth.ClusterConfig) *auth.ClusterConfig {
	m := clusterselect.New(configs)
	p := tea.NewProgram(m, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return nil
	}
	final := result.(clusterselect.Model)
	if final.Quitting() {
		return nil
	}
	return final.Selected()
}
