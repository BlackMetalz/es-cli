package tui

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/kienlt/es-cli/internal/es"
	"github.com/kienlt/es-cli/internal/tui/commands"
	"github.com/kienlt/es-cli/internal/tui/components/allocationmenu"
	"github.com/kienlt/es-cli/internal/tui/components/cmdpalette"
	"github.com/kienlt/es-cli/internal/tui/components/createilm"
	"github.com/kienlt/es-cli/internal/tui/components/createindex"
	"github.com/kienlt/es-cli/internal/tui/components/createtemplate"
	"github.com/kienlt/es-cli/internal/tui/header"
	"github.com/kienlt/es-cli/internal/tui/theme"
	"github.com/kienlt/es-cli/internal/tui/views"
	dashview "github.com/kienlt/es-cli/internal/tui/views/dashboard"
	detailview "github.com/kienlt/es-cli/internal/tui/views/detail"
	ilmview "github.com/kienlt/es-cli/internal/tui/views/ilm"
	indexview "github.com/kienlt/es-cli/internal/tui/views/index"
	"github.com/kienlt/es-cli/internal/tui/views/jsonview"
	nodeview "github.com/kienlt/es-cli/internal/tui/views/node"
	queryview "github.com/kienlt/es-cli/internal/tui/views/query"
	shardview "github.com/kienlt/es-cli/internal/tui/views/shard"
	templateview "github.com/kienlt/es-cli/internal/tui/views/template"

	tea "github.com/charmbracelet/bubbletea"
)

type overlayType int

const (
	overlayNone overlayType = iota
	overlayCreateIndex
	overlayConfirm
	overlayHelp
	overlayAllocation
	overlayCreateILM
	overlayCreateTemplate
	overlayError
)

const statusBarHeight = 1

type App struct {
	header    header.Model
	viewStack []views.View
	client    *es.Client
	router    *commands.Router
	width     int
	height    int

	overlay          overlayType
	createIndexFm    createindex.Model
	createILMFm      createilm.Model
	createTemplateFm createtemplate.Model
	allocMenu        allocationmenu.Model

	// Command palette
	cmdPalette   cmdpalette.Model
	cmdPaletteOn bool

	// Confirm overlay state
	confirmAction string
	confirmIndex  string

	// Status bar
	flashMsg          string
	errorPopupMsg     string
	flashStyle        lipgloss.Style
	allocationSetting string
}

type clusterInfoMsg struct {
	Name    string
	Version string
	Health  string
}

type clearFlashMsg struct{}

type flashErrorMsg struct {
	err error
}

type editILMMsg struct {
	Name        string
	DeleteAfter string
}

type openCreateTemplateMsg struct {
	ILMPolicies []string
	Existing    []createtemplate.ExistingTemplate
}

type editTemplateMsg struct {
	Name        string
	Patterns    string
	Shards      string
	Replicas    string
	ILMPolicy   string
	ILMPolicies []string
	Existing    []createtemplate.ExistingTemplate
}

type allocationSettingMsg struct {
	Current string
}

type showAllocationMenuMsg struct {
	Current string
}

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

func (a *App) switchView(v views.View) {
	v.SetSize(a.width, a.viewHeight())
	a.viewStack = []views.View{v}
	a.syncHeader()
}

func (a *App) syncHeader() {
	a.header.ViewName = a.currentView().Name()
	a.header.HelpGroups = a.currentView().HelpGroups()
}

func newRouter() *commands.Router {
	r := commands.NewRouter()
	r.Register(commands.Command{Name: "index", Aliases: []string{"indices"}, Description: "List indices"})
	r.Register(commands.Command{Name: "node", Aliases: []string{"nodes"}, Description: "List nodes"})
	r.Register(commands.Command{Name: "shard", Aliases: []string{"shards"}, Description: "List shards"})
	r.Register(commands.Command{Name: "dashboard", Aliases: []string{"dash"}, Description: "Cluster dashboard"})
	r.Register(commands.Command{Name: "ilm", Aliases: []string{"ilm-policy"}, Description: "ILM policies"})
	r.Register(commands.Command{Name: "template", Aliases: []string{"templates", "index-template"}, Description: "Index templates"})
	r.Register(commands.Command{Name: "discovery", Description: "Query logs"})
	return r
}

func NewApp(client *es.Client, clusterURL, clusterName string) *App {
	h := header.New(clusterURL)
	h.User = client.Username()
	h.ClusterName = clusterName
	dashView := dashview.New(client)

	h.ViewName = dashView.Name()
	h.HelpGroups = dashView.HelpGroups()

	return &App{
		header:    h,
		viewStack: []views.View{dashView},
		client:    client,
		router:    newRouter(),
	}
}

func (a *App) fetchExistingTemplates() []createtemplate.ExistingTemplate {
	templates, err := a.client.ListIndexTemplates()
	if err != nil {
		return nil
	}
	var existing []createtemplate.ExistingTemplate
	for _, t := range templates {
		patterns := strings.Split(t.IndexPatterns, ", ")
		existing = append(existing, createtemplate.ExistingTemplate{Name: t.Name, Patterns: patterns})
	}
	return existing
}

func (a *App) fetchUserILMPolicies() []string {
	policies, err := a.client.ListILMPolicies()
	if err != nil {
		return nil
	}
	var names []string
	for _, p := range policies {
		if !strings.HasPrefix(p.Name, ".") && !p.Managed {
			names = append(names, p.Name)
		}
	}
	return names
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
	// Global messages
	switch msg := msg.(type) {
	case clusterInfoMsg:
		// Keep cluster name from config, only update version and health
		a.header.ESVersion = msg.Version
		a.header.ClusterHealth = msg.Health
		return a, nil

	case clearFlashMsg:
		a.flashMsg = ""
		return a, nil

	case flashErrorMsg:
		a.errorPopupMsg = msg.err.Error()
		a.overlay = overlayError
		return a, nil

	case detailview.GoBackMsg:
		a.popView()
		return a, nil

	case jsonview.GoBackMsg:
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
				return flashErrorMsg{err: err}
			}
			return indexview.ActionCompleteMsg{Action: "created", Index: indexName}
		}

	case createindex.CancelMsg:
		a.overlay = overlayNone
		return a, nil

	case createilm.SubmitMsg:
		a.overlay = overlayNone
		name := msg.Name
		body := msg.Body
		action := "created"
		if msg.Editing {
			action = "updated"
		}
		return a, func() tea.Msg {
			err := a.client.CreateILMPolicy(name, body)
			if err != nil {
				return flashErrorMsg{err: err}
			}
			return ilmview.ActionCompleteMsg{Action: action, Policy: name}
		}

	case createilm.CancelMsg:
		a.overlay = overlayNone
		return a, nil

	case createtemplate.SubmitMsg:
		a.overlay = overlayNone
		name := msg.Name
		body := msg.Body
		action := "created"
		if msg.Editing {
			action = "updated"
		}
		return a, func() tea.Msg {
			err := a.client.CreateIndexTemplate(name, body)
			if err != nil {
				return flashErrorMsg{err: err}
			}
			return templateview.ActionCompleteMsg{Action: action, Template: name}
		}

	case createtemplate.CancelMsg:
		a.overlay = overlayNone
		return a, nil

	case cmdpalette.SubmitMsg:
		a.cmdPaletteOn = false
		return a.handleCommand(msg.Command)

	case cmdpalette.CancelMsg:
		a.cmdPaletteOn = false
		return a, nil

	case allocationmenu.SubmitMsg:
		a.overlay = overlayNone
		value := msg.Value
		return a, func() tea.Msg {
			err := a.client.SetAllocationSetting(value)
			if err != nil {
				return flashErrorMsg{err: err}
			}
			return allocationSettingMsg{Current: value}
		}

	case allocationmenu.CancelMsg:
		a.overlay = overlayNone
		return a, nil

	case allocationSettingMsg:
		a.allocationSetting = msg.Current
		label := "reset to default"
		if msg.Current != "" {
			label = msg.Current
		}
		cmd := a.setFlash(fmt.Sprintf("Allocation set to %s", label), theme.StatusBarSuccessStyle)
		return a, cmd

	case openCreateTemplateMsg:
		a.overlay = overlayCreateTemplate
		a.createTemplateFm = createtemplate.New(msg.ILMPolicies, msg.Existing)
		return a, a.createTemplateFm.Init()

	case editILMMsg:
		a.overlay = overlayCreateILM
		a.createILMFm = createilm.NewEdit(msg.Name, msg.DeleteAfter)
		return a, a.createILMFm.Init()

	case editTemplateMsg:
		a.overlay = overlayCreateTemplate
		a.createTemplateFm = createtemplate.NewEdit(msg.Name, msg.Patterns, msg.Shards, msg.Replicas, msg.ILMPolicy, msg.ILMPolicies, msg.Existing)
		return a, a.createTemplateFm.Init()

	case showAllocationMenuMsg:
		a.allocationSetting = msg.Current
		a.overlay = overlayAllocation
		a.allocMenu = allocationmenu.New(msg.Current)
		return a, nil
	}

	// Route to command palette
	if a.cmdPaletteOn {
		var cmd tea.Cmd
		a.cmdPalette, cmd = a.cmdPalette.Update(msg)
		return a, cmd
	}

	// Route to overlays
	if a.overlay == overlayCreateIndex {
		var cmd tea.Cmd
		a.createIndexFm, cmd = a.createIndexFm.Update(msg)
		return a, cmd
	}
	if a.overlay == overlayError {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "enter" || keyMsg.String() == "esc" {
				a.overlay = overlayNone
				a.errorPopupMsg = ""
			}
		}
		return a, nil
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
	if a.overlay == overlayAllocation {
		var cmd tea.Cmd
		a.allocMenu, cmd = a.allocMenu.Update(msg)
		return a, cmd
	}
	if a.overlay == overlayCreateILM {
		var cmd tea.Cmd
		a.createILMFm, cmd = a.createILMFm.Update(msg)
		return a, cmd
	}
	if a.overlay == overlayCreateTemplate {
		var cmd tea.Cmd
		a.createTemplateFm, cmd = a.createTemplateFm.Update(msg)
		return a, cmd
	}

	// Handle flash messages from completed actions
	if msg, ok := msg.(indexview.ActionCompleteMsg); ok {
		flashText := fmt.Sprintf("Index '%s' %s successfully", msg.Index, msg.Action)
		cmd := a.setFlash(flashText, theme.StatusBarSuccessStyle)
		var viewCmd tea.Cmd
		a.viewStack[len(a.viewStack)-1], viewCmd = a.currentView().Update(msg)
		return a, tea.Batch(cmd, viewCmd)
	}
	if msg, ok := msg.(ilmview.ActionCompleteMsg); ok {
		flashText := fmt.Sprintf("Policy '%s' %s successfully", msg.Policy, msg.Action)
		cmd := a.setFlash(flashText, theme.StatusBarSuccessStyle)
		var viewCmd tea.Cmd
		a.viewStack[len(a.viewStack)-1], viewCmd = a.currentView().Update(msg)
		return a, tea.Batch(cmd, viewCmd)
	}
	if msg, ok := msg.(templateview.ActionCompleteMsg); ok {
		flashText := fmt.Sprintf("Template '%s' %s successfully", msg.Template, msg.Action)
		cmd := a.setFlash(flashText, theme.StatusBarSuccessStyle)
		var viewCmd tea.Cmd
		a.viewStack[len(a.viewStack)-1], viewCmd = a.currentView().Update(msg)
		return a, tea.Batch(cmd, viewCmd)
	}

	// App-level keys (only when view is not in input mode)
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if !a.currentView().IsInputMode() {
			switch keyMsg.String() {
			case ":":
				a.cmdPaletteOn = true
				a.cmdPalette = cmdpalette.New(a.router, a.width)
				return a, a.cmdPalette.Init()
			case "n":
				switch a.currentView().Name() {
				case "Indices":
					a.overlay = overlayCreateIndex
					a.createIndexFm = createindex.New()
					return a, a.createIndexFm.Init()
				case "ILM Policies":
					a.overlay = overlayCreateILM
					a.createILMFm = createilm.New()
					return a, a.createILMFm.Init()
				case "Index Templates":
					return a, func() tea.Msg {
						return openCreateTemplateMsg{
							ILMPolicies: a.fetchUserILMPolicies(),
							Existing:    a.fetchExistingTemplates(),
						}
					}
				}
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

func (a *App) handleCommand(name string) (tea.Model, tea.Cmd) {
	var v views.View
	switch name {
	case "index":
		v = indexview.New(a.client)
	case "node":
		v = nodeview.New(a.client)
	case "shard":
		v = shardview.New(a.client)
	case "dashboard":
		v = dashview.New(a.client)
	case "ilm":
		v = ilmview.New(a.client)
	case "template":
		v = templateview.New(a.client)
	case "discovery":
		v = queryview.New(a.client)
	default:
		return a, nil
	}
	a.switchView(v)
	return a, v.Init()
}

func (a *App) handlePendingAction(pa *views.PendingAction) (tea.Model, tea.Cmd) {
	switch pa.Type {
	case "view_detail":
		dv := detailview.New(a.client, pa.Index)
		dv.SetSize(a.width, a.viewHeight())
		a.pushView(dv)
		return a, dv.Init()

	case "view_ilm_detail":
		name := pa.Index
		client := a.client
		jv := jsonview.New("ILM: "+name, func() tea.Msg {
			data, err := client.GetILMPolicy(name)
			if err != nil {
				return jsonview.ErrorMsg{Err: err}
			}
			return jsonview.DataLoadedMsg{Data: data}
		})
		jv.SetSize(a.width, a.viewHeight())
		a.pushView(jv)
		return a, jv.Init()

	case "view_template_detail":
		name := pa.Index
		client := a.client
		jv := jsonview.New("Template: "+name, func() tea.Msg {
			data, err := client.GetIndexTemplate(name)
			if err != nil {
				return jsonview.ErrorMsg{Err: err}
			}
			return jsonview.DataLoadedMsg{Data: data}
		})
		jv.SetSize(a.width, a.viewHeight())
		a.pushView(jv)
		return a, jv.Init()

	case "explain_shard":
		// pa.Index format: "index_name|shard_num|primary_bool"
		parts := strings.SplitN(pa.Index, "|", 3)
		if len(parts) == 3 {
			idx, shardN, primary := parts[0], parts[1], parts[2] == "true"
			client := a.client
			jv := jsonview.New("Shard Explain: "+idx+"["+shardN+"]", func() tea.Msg {
				data, err := client.AllocationExplain(idx, shardN, primary)
				if err != nil {
					return jsonview.ErrorMsg{Err: err}
				}
				return jsonview.DataLoadedMsg{Data: data}
			})
			jv.SetSize(a.width, a.viewHeight())
			a.pushView(jv)
			return a, jv.Init()
		}
		return a, nil

	case "edit_ilm":
		name := pa.Index
		client := a.client
		return a, func() tea.Msg {
			data, err := client.GetILMPolicy(name)
			if err != nil {
				return flashErrorMsg{err: err}
			}
			msg := editILMMsg{Name: name}
			var raw map[string]interface{}
			if json.Unmarshal(data, &raw) == nil {
				if pd, ok := raw[name].(map[string]interface{}); ok {
					if p, ok := pd["policy"].(map[string]interface{}); ok {
						if phases, ok := p["phases"].(map[string]interface{}); ok {
							if del, ok := phases["delete"].(map[string]interface{}); ok {
								if v, ok := del["min_age"].(string); ok {
									msg.DeleteAfter = v
								}
							}
						}
					}
				}
			}
			return msg
		}

	case "edit_template":
		name := pa.Index
		client := a.client
		return a, func() tea.Msg {
			data, err := client.GetIndexTemplate(name)
			if err != nil {
				return flashErrorMsg{err: err}
			}
			msg := editTemplateMsg{Name: name}
			var raw map[string]interface{}
			if json.Unmarshal(data, &raw) == nil {
				if templates, ok := raw["index_templates"].([]interface{}); ok && len(templates) > 0 {
					if entry, ok := templates[0].(map[string]interface{}); ok {
						if it, ok := entry["index_template"].(map[string]interface{}); ok {
							if patterns, ok := it["index_patterns"].([]interface{}); ok {
								var pats []string
								for _, p := range patterns {
									pats = append(pats, es.JsonStr(p))
								}
								msg.Patterns = strings.Join(pats, ",")
							}
							if tmpl, ok := it["template"].(map[string]interface{}); ok {
								if settings, ok := tmpl["settings"].(map[string]interface{}); ok {
									if idx, ok := settings["index"].(map[string]interface{}); ok {
										msg.Shards = es.JsonStr(idx["number_of_shards"])
										msg.Replicas = es.JsonStr(idx["number_of_replicas"])
										if lc, ok := idx["lifecycle"].(map[string]interface{}); ok {
											msg.ILMPolicy = es.JsonStr(lc["name"])
										}
									}
								}
							}
						}
					}
				}
			}
			msg.ILMPolicies = a.fetchUserILMPolicies()
			msg.Existing = a.fetchExistingTemplates()
			return msg
		}

	case "close", "open", "delete", "delete_ilm", "delete_template":
		a.overlay = overlayConfirm
		a.confirmAction = pa.Type
		a.confirmIndex = pa.Index
		return a, nil

	case "set_allocation":
		return a, func() tea.Msg {
			current, _ := a.client.GetAllocationSetting()
			return showAllocationMenuMsg{Current: current}
		}
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
			case "delete_ilm":
				err = a.client.DeleteILMPolicy(indexName)
				if err != nil {
					return flashErrorMsg{err: err}
				}
				return ilmview.ActionCompleteMsg{Action: "deleted", Policy: indexName}
			case "delete_template":
				err = a.client.DeleteIndexTemplate(indexName)
				if err != nil {
					return flashErrorMsg{err: err}
				}
				return templateview.ActionCompleteMsg{Action: "deleted", Template: indexName}
			}
			if err != nil {
				return flashErrorMsg{err: err}
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
	action := a.confirmAction
	resource := "index"
	switch action {
	case "delete_ilm":
		action = "delete"
		resource = "ILM policy"
	case "delete_template":
		action = "delete"
		resource = "index template"
	}

	return theme.ModalStyle.Render(
		theme.ModalTitleStyle.Render("Confirm") + "\n\n" +
			"Are you sure you want to " + action + " " + resource + " '" + a.confirmIndex + "'?\n\n" +
			theme.HelpKeyStyle.Render("y") + theme.HelpDescStyle.Render("/") + theme.HelpKeyStyle.Render("N"),
	)
}

func (a *App) statusBarView() string {
	if a.cmdPaletteOn {
		return a.cmdPalette.View()
	}

	viewName := theme.ViewNameStyle.Render(a.currentView().Name())
	contextInfo := a.currentView().StatusInfo()

	left := " " + viewName
	if contextInfo != "" {
		left += "  " + theme.HelpDescStyle.Render(contextInfo)
	}

	right := ""
	if a.allocationSetting != "" {
		right += theme.HealthYellowStyle.Render("alloc: "+a.allocationSetting) + "  "
	}
	if a.flashMsg != "" {
		right += a.flashStyle.Render(a.flashMsg) + " "
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
	case overlayAllocation:
		overlay = a.allocMenu.View()
	case overlayCreateILM:
		overlay = a.createILMFm.View()
	case overlayCreateTemplate:
		overlay = a.createTemplateFm.View()
	case overlayError:
		// Wrap long error text
		maxW := a.width - 20
		if maxW < 40 {
			maxW = 40
		}
		errText := lipgloss.NewStyle().Width(maxW).Foreground(theme.ColorRed).Render(a.errorPopupMsg)
		overlay = theme.ModalStyle.Render(
			theme.ModalTitleStyle.Render("Error") + "\n\n" +
				errText + "\n\n" +
				theme.HelpDescStyle.Render("Press ") + theme.HelpKeyStyle.Render("Enter") + theme.HelpDescStyle.Render(" or ") + theme.HelpKeyStyle.Render("Esc") + theme.HelpDescStyle.Render(" to close"),
		)
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
