package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/kienlt/es-cli/internal/es"
	"github.com/kienlt/es-cli/internal/tui/components/createindex"
	"github.com/kienlt/es-cli/internal/tui/header"
	"github.com/kienlt/es-cli/internal/tui/theme"
	"github.com/kienlt/es-cli/internal/tui/views"
	detailview "github.com/kienlt/es-cli/internal/tui/views/detail"
	indexview "github.com/kienlt/es-cli/internal/tui/views/index"

	tea "github.com/charmbracelet/bubbletea"
)

type overlayType int

const (
	overlayNone overlayType = iota
	overlayCreateIndex
	overlayConfirm
	overlayHelp
)

const statusBarHeight = 1

type App struct {
	header    header.Model
	viewStack []views.View // stack of views for back navigation
	client    *es.Client
	width     int
	height    int

	overlay       overlayType
	createIndexFm createindex.Model

	// Confirm overlay state
	confirmAction string
	confirmIndex  string

	// Status bar flash message
	flashMsg   string
	flashStyle lipgloss.Style
}

type clusterInfoMsg struct {
	Name    string
	Version string
	Health  string
}

type clearFlashMsg struct{}

func (a *App) currentView() views.View {
	return a.viewStack[len(a.viewStack)-1]
}

func (a *App) pushView(v views.View) {
	a.viewStack = append(a.viewStack, v)
	a.syncHeader()
}

func (a *App) popView() {
	if len(a.viewStack) > 1 {
		a.viewStack = a.viewStack[:len(a.viewStack)-1]
		a.currentView().SetSize(a.width, a.viewHeight())
		a.syncHeader()
	}
}

func (a *App) syncHeader() {
	a.header.ViewName = a.currentView().Name()
	a.header.HelpGroups = a.currentView().HelpGroups()
}

func NewApp(client *es.Client, clusterURL string) *App {
	h := header.New(clusterURL)
	h.User = client.Username()
	idxView := indexview.New(client)

	h.ViewName = idxView.Name()
	h.HelpGroups = idxView.HelpGroups()

	return &App{
		header:    h,
		viewStack: []views.View{idxView},
		client:    client,
	}
}

func (a *App) fetchClusterInfo() tea.Cmd {
	return func() tea.Msg {
		info, err := a.client.GetClusterInfo()
		if err != nil {
			return clusterInfoMsg{}
		}
		return clusterInfoMsg{Name: info.Name, Version: info.Version, Health: info.Health}
	}
}

func (a *App) Init() tea.Cmd {
	return tea.Batch(a.currentView().Init(), a.fetchClusterInfo())
}

func (a *App) viewHeight() int {
	return a.height - a.header.Height() - statusBarHeight
}

func (a *App) setFlash(msg string, style lipgloss.Style) tea.Cmd {
	a.flashMsg = msg
	a.flashStyle = style
	return tea.Tick(4*time.Second, func(time.Time) tea.Msg {
		return clearFlashMsg{}
	})
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case clusterInfoMsg:
		a.header.ClusterName = msg.Name
		a.header.ESVersion = msg.Version
		a.header.ClusterHealth = msg.Health
		return a, nil

	case clearFlashMsg:
		a.flashMsg = ""
		return a, nil

	case detailview.GoBackMsg:
		a.popView()
		return a, nil

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.header.Width = msg.Width
		a.currentView().SetSize(msg.Width, a.viewHeight())
		return a, nil

	case createindex.SubmitMsg:
		a.overlay = overlayNone
		indexName := msg.Name
		return a, func() tea.Msg {
			err := a.client.CreateIndex(msg.Name, msg.Shards, msg.Replicas)
			if err != nil {
				return indexview.ErrorMsg{Err: err}
			}
			return indexview.ActionCompleteMsg{Action: "created", Index: indexName}
		}

	case createindex.CancelMsg:
		a.overlay = overlayNone
		return a, nil
	}

	// Route to overlays
	if a.overlay == overlayCreateIndex {
		var cmd tea.Cmd
		a.createIndexFm, cmd = a.createIndexFm.Update(msg)
		return a, cmd
	}

	if a.overlay == overlayHelp {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "esc" || keyMsg.String() == "?" {
				a.overlay = overlayNone
			}
		}
		return a, nil
	}

	if a.overlay == overlayConfirm {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			return a.handleConfirmKey(keyMsg)
		}
		return a, nil
	}

	// Handle flash messages from completed actions
	if msg, ok := msg.(indexview.ActionCompleteMsg); ok {
		flashText := fmt.Sprintf("Index '%s' %s successfully", msg.Index, msg.Action)
		cmd := a.setFlash(flashText, theme.StatusBarSuccessStyle)
		var viewCmd tea.Cmd
		a.viewStack[len(a.viewStack)-1], viewCmd = a.currentView().Update(msg)
		return a, tea.Batch(cmd, viewCmd)
	}

	// App-level keys (only when view is not in input mode)
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if !a.currentView().IsInputMode() {
			switch keyMsg.String() {
			case "n":
				a.overlay = overlayCreateIndex
				a.createIndexFm = createindex.New()
				return a, a.createIndexFm.Init()
			case "?":
				a.overlay = overlayHelp
				return a, nil
			}
		}
	}

	// Route to current view
	var cmd tea.Cmd
	a.viewStack[len(a.viewStack)-1], cmd = a.currentView().Update(msg)

	// Check for pending actions from view
	if pa := a.currentView().PopPendingAction(); pa != nil {
		return a.handlePendingAction(pa)
	}

	a.syncHeader()
	return a, cmd
}

func (a *App) handlePendingAction(pa *views.PendingAction) (tea.Model, tea.Cmd) {
	switch pa.Type {
	case "view_detail":
		dv := detailview.New(a.client, pa.Index)
		dv.SetSize(a.width, a.viewHeight())
		a.pushView(dv)
		return a, dv.Init()

	case "close", "open", "delete":
		a.overlay = overlayConfirm
		a.confirmAction = pa.Type
		a.confirmIndex = pa.Index
		return a, nil
	}
	return a, nil
}

func (a *App) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		action := a.confirmAction
		indexName := a.confirmIndex
		a.overlay = overlayNone
		a.confirmAction = ""
		a.confirmIndex = ""
		return a, func() tea.Msg {
			var err error
			actionStr := action + "d"
			switch action {
			case "close":
				err = a.client.CloseIndex(indexName)
				actionStr = "closed"
			case "open":
				err = a.client.OpenIndex(indexName)
				actionStr = "opened"
			case "delete":
				err = a.client.DeleteIndex(indexName)
				actionStr = "deleted"
			}
			if err != nil {
				return indexview.ErrorMsg{Err: err}
			}
			return indexview.ActionCompleteMsg{Action: actionStr, Index: indexName}
		}
	case "n", "N", "esc":
		a.overlay = overlayNone
		a.confirmAction = ""
		a.confirmIndex = ""
		return a, nil
	}
	return a, nil
}

func (a *App) confirmOverlayView() string {
	return theme.ModalStyle.Render(
		theme.ModalTitleStyle.Render("Confirm") + "\n\n" +
			"Are you sure you want to " + a.confirmAction + " index '" + a.confirmIndex + "'?\n\n" +
			theme.HelpKeyStyle.Render("y") + theme.HelpDescStyle.Render("/") + theme.HelpKeyStyle.Render("N"),
	)
}

func (a *App) statusBarView() string {
	viewName := theme.ViewNameStyle.Render(a.currentView().Name())
	contextInfo := a.currentView().StatusInfo()

	left := " " + viewName
	if contextInfo != "" {
		left += "  " + theme.HelpDescStyle.Render(contextInfo)
	}

	right := ""
	if a.flashMsg != "" {
		right = a.flashStyle.Render(a.flashMsg) + " "
	}

	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	gap := a.width - leftWidth - rightWidth
	if gap < 0 {
		gap = 0
	}

	bar := left + strings.Repeat(" ", gap) + right
	return theme.StatusBarStyle.Width(a.width).Render(bar)
}

func (a *App) View() string {
	if a.overlay == overlayHelp {
		return renderHelpFullScreen(a.currentView().HelpGroups(), a.width, a.height)
	}

	headerView := a.header.View()
	contentView := a.currentView().View()
	statusBar := a.statusBarView()

	mainView := lipgloss.JoinVertical(lipgloss.Left, headerView, contentView, statusBar)

	var overlay string
	switch a.overlay {
	case overlayCreateIndex:
		overlay = a.createIndexFm.View()
	case overlayConfirm:
		overlay = a.confirmOverlayView()
	}

	if overlay != "" {
		overlayWidth := lipgloss.Width(overlay)
		overlayHeight := lipgloss.Height(overlay)
		x := (a.width - overlayWidth) / 2
		y := (a.height - overlayHeight) / 2
		if x < 0 {
			x = 0
		}
		if y < 0 {
			y = 0
		}
		mainView = placeOverlay(x, y, overlay, mainView)
	}

	return mainView
}
