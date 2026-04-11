package clusterselect

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kienlt/es-cli/internal/auth"
	"github.com/kienlt/es-cli/internal/tui/theme"
)

type Model struct {
	clusters []auth.ClusterConfig
	cursor   int
	width    int
	height   int
	selected *auth.ClusterConfig
	quitting bool
}

func New(clusters []auth.ClusterConfig) Model {
	return Model{clusters: clusters}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			m.cursor = (m.cursor + 1) % len(m.clusters)
		case "k", "up":
			m.cursor = (m.cursor - 1 + len(m.clusters)) % len(m.clusters)
		case "enter":
			m.selected = &m.clusters[m.cursor]
			return m, tea.Quit
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.ColorCyan)
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.ColorWhite)
	urlStyle := lipgloss.NewStyle().Foreground(theme.ColorGray)
	dimStyle := lipgloss.NewStyle().Foreground(theme.ColorDimmed)

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString("  " + titleStyle.Render("es-cli") + " — Select Cluster\n\n")

	// Find max name width for alignment
	maxName := 0
	for _, c := range m.clusters {
		if len(c.Name) > maxName {
			maxName = len(c.Name)
		}
	}

	for i, c := range m.clusters {
		cursor := "  "
		name := dimStyle.Render(c.Name)
		url := urlStyle.Render(c.URL)
		if i == m.cursor {
			cursor = theme.HelpKeyStyle.Render("> ")
			name = selectedStyle.Render(c.Name)
		}
		pad := strings.Repeat(" ", maxName-len(c.Name)+2)
		b.WriteString(fmt.Sprintf("  %s%s%s%s\n", cursor, name, pad, url))
	}

	b.WriteString("\n")
	b.WriteString("  " + theme.HelpDescStyle.Render("enter: connect • q: quit") + "\n")

	return b.String()
}

// Selected returns the selected cluster config, or nil if user quit.
func (m Model) Selected() *auth.ClusterConfig {
	return m.selected
}

// Quitting returns true if user chose to quit.
func (m Model) Quitting() bool {
	return m.quitting
}
