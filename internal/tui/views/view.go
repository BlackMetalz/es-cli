package views

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// HelpGroup represents a group of keybindings displayed in the help menu.
type HelpGroup struct {
	Title    string
	Bindings []key.Binding
}

// PendingAction represents an action requested by a view that the app should handle.
type PendingAction struct {
	Type  string // "close", "open", "delete", "view_detail", etc.
	Index string // target index name
}

// View is the interface that all TUI views must implement.
type View interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (View, tea.Cmd)
	View() string
	Name() string
	HelpGroups() []HelpGroup
	SetSize(width, height int)

	// IsInputMode returns true when the view is capturing text input (search, etc.)
	// so the app knows not to intercept keys.
	IsInputMode() bool

	// PopPendingAction returns and clears any pending action the view wants the app to handle.
	PopPendingAction() *PendingAction

	// StatusInfo returns context text for the status bar.
	StatusInfo() string
}
