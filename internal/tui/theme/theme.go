package theme

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	ColorGreen  = lipgloss.Color("#00ff00")
	ColorYellow = lipgloss.Color("#ffff00")
	ColorRed    = lipgloss.Color("#ff0000")
	ColorBlue   = lipgloss.Color("#00aaff")
	ColorCyan   = lipgloss.Color("#00ffff")
	ColorWhite  = lipgloss.Color("#ffffff")
	ColorGray   = lipgloss.Color("#888888")
	ColorDimmed = lipgloss.Color("#555555")

	// Header styles
	LogoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorCyan)

	ClusterInfoStyle = lipgloss.NewStyle().
				Foreground(ColorGreen)

	ViewNameStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite).
			Background(lipgloss.Color("#333333")).
			Padding(0, 1)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(ColorCyan).
			Bold(true)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(ColorGray)

	SeparatorStyle = lipgloss.NewStyle().
			Foreground(ColorDimmed)

	// Table styles
	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorWhite).
				Padding(0, 1).
				BorderBottom(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(ColorDimmed)

	TableSelectedStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#3a3a8a")).
				Foreground(ColorWhite).
				Bold(true)

	// Status bar
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorWhite).
			Background(lipgloss.Color("#333333"))

	StatusBarSuccessStyle = lipgloss.NewStyle().
				Foreground(ColorGreen).
				Background(lipgloss.Color("#333333")).
				Bold(true)

	StatusBarErrorStyle = lipgloss.NewStyle().
				Foreground(ColorRed).
				Background(lipgloss.Color("#333333")).
				Bold(true)

	// Health indicator styles
	HealthGreenStyle  = lipgloss.NewStyle().Foreground(ColorGreen)
	HealthYellowStyle = lipgloss.NewStyle().Foreground(ColorYellow)
	HealthRedStyle    = lipgloss.NewStyle().Foreground(ColorRed)

	// Modal styles
	ModalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBlue).
			Padding(1, 2)

	ModalTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorCyan)

	// Error/status styles
	ErrorStyle   = lipgloss.NewStyle().Foreground(ColorRed)
	SuccessStyle = lipgloss.NewStyle().Foreground(ColorGreen)
)

func HealthStyle(health string) lipgloss.Style {
	switch health {
	case "green":
		return HealthGreenStyle
	case "yellow":
		return HealthYellowStyle
	case "red":
		return HealthRedStyle
	default:
		return lipgloss.NewStyle().Foreground(ColorGray)
	}
}
