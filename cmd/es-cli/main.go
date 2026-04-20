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

func printUsage() {
	fmt.Println("es-cli - K9s-style terminal UI for Elasticsearch")
	fmt.Println()
	fmt.Println("Usage: es-cli [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --cluster <name>  Connect directly to a named cluster")
	fmt.Println("  --read-only       Disable all create/edit/delete operations")
	fmt.Println("  --version, -v     Print version and exit")
	fmt.Println("  --help, -h        Show this help message")
	fmt.Println()
	fmt.Println("Commands (press ':' in the TUI to open command palette):")
	fmt.Println("  :dashboard        Cluster overview (aliases: dash)")
	fmt.Println("  :index            List indices (aliases: indices)")
	fmt.Println("  :node             List nodes (aliases: nodes)")
	fmt.Println("  :shard            List shards (aliases: shards)")
	fmt.Println("  :ilm              ILM policies (aliases: ilm-policy)")
	fmt.Println("  :template         Index templates (aliases: templates, index-template)")
	fmt.Println("  :discovery        Log viewer / search")
}

func main() {
	flags := parseFlags()

	if flags.help {
		printUsage()
		return
	}
	if flags.version {
		fmt.Println(version)
		return
	}

	authPath := auth.DefaultAuthPath()

	configs, err := auth.LoadAuth(authPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Auth error: %v\n", err)
		os.Exit(1)
	}

	clusterName := flags.cluster

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
	app := tui.NewApp(client, cluster.URL, cluster.Name, flags.readOnly)

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

type flags struct {
	cluster  string
	readOnly bool
	version  bool
	help     bool
}

func parseFlags() flags {
	var f flags
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--version", "-v":
			f.version = true
		case "--help", "-h":
			f.help = true
		case "--read-only":
			f.readOnly = true
		case "--cluster":
			if i+1 < len(args) {
				i++
				f.cluster = args[i]
			} else {
				fmt.Fprintf(os.Stderr, "Error: --cluster requires a value\n")
				os.Exit(1)
			}
		default:
			fmt.Fprintf(os.Stderr, "Error: unsupported flag %q\n", args[i])
			fmt.Fprintf(os.Stderr, "Run 'es-cli --help' for usage.\n")
			os.Exit(1)
		}
	}
	return f
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
