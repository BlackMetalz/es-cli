package query

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kienlt/es-cli/internal/tui/theme"
	"github.com/kienlt/es-cli/internal/tui/views"
)

type FilterCondition struct {
	Field    string
	Value    string
	Operator string // "AND" or "OR" (between this and next)
}

type qbState int

const (
	qbList        qbState = iota // viewing/navigating filter list
	qbSelectField                // picking a field
	qbEnterValue                 // typing value
)

// qb fields are stored directly on Model:
// buildingQuery bool
// qbSt          qbState
// qbFilters     []FilterCondition
// qbCursor      int
// qbFieldCursor int
// qbFilteredFields []es.FieldMapping
// qbFieldInput  textinput.Model
// qbValueInput  textinput.Model
// qbPickedField string

func (m *Model) handleQueryBuilderKey(msg tea.KeyMsg) (views.View, tea.Cmd) {
	switch m.qbSt {
	case qbList:
		return m.handleQBListKey(msg)
	case qbSelectField:
		return m.handleQBFieldKey(msg)
	case qbEnterValue:
		return m.handleQBValueKey(msg)
	}
	return m, nil
}

func (m *Model) handleQBListKey(msg tea.KeyMsg) (views.View, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.qbCursor < len(m.qbFilters)-1 {
			m.qbCursor++
		}
	case "k", "up":
		if m.qbCursor > 0 {
			m.qbCursor--
		}
	case "n", "+":
		// Add new filter — open field selector
		m.qbSt = qbSelectField
		m.qbFieldCursor = 0
		m.qbFieldInput.SetValue("")
		m.qbFieldInput.Focus()
		m.qbFilteredFields = m.allFields
		return m, m.qbFieldInput.Cursor.BlinkCmd()
	case "d":
		if len(m.qbFilters) > 0 && m.qbCursor < len(m.qbFilters) {
			m.qbFilters = append(m.qbFilters[:m.qbCursor], m.qbFilters[m.qbCursor+1:]...)
			if m.qbCursor >= len(m.qbFilters) && m.qbCursor > 0 {
				m.qbCursor--
			}
		}
	case "a":
		// Toggle operator between the selected filter and the next one
		// Only works when there are 2+ filters and cursor is not on the last one
		if len(m.qbFilters) >= 2 && m.qbCursor < len(m.qbFilters)-1 {
			f := &m.qbFilters[m.qbCursor]
			if f.Operator == "AND" {
				f.Operator = "OR"
			} else {
				f.Operator = "AND"
			}
		}
	case "enter":
		// Apply filters
		m.buildingQuery = false
		m.query = buildQueryString(m.qbFilters)
		m.loading = true
		m.page = 0
		m.pageSorts = nil
		return m, m.executeSearch()
	case "esc":
		m.buildingQuery = false
	}
	return m, nil
}

func (m *Model) handleQBFieldKey(msg tea.KeyMsg) (views.View, tea.Cmd) {
	switch msg.Type {
	case tea.KeyTab:
		// Auto-complete field name from current selection
		if len(m.qbFilteredFields) > 0 && m.qbFieldCursor < len(m.qbFilteredFields) {
			m.qbFieldInput.SetValue(m.qbFilteredFields[m.qbFieldCursor].Name)
			m.qbFieldInput.CursorEnd()
			m.qbFilteredFields = m.allFields
			m.qbFieldCursor = 0
			// Re-filter to show exact match
			for i, f := range m.allFields {
				if f.Name == m.qbFieldInput.Value() {
					m.qbFieldCursor = i
					break
				}
			}
		}
		return m, nil
	case tea.KeyEnter:
		if len(m.qbFilteredFields) > 0 && m.qbFieldCursor < len(m.qbFilteredFields) {
			m.qbPickedField = m.qbFilteredFields[m.qbFieldCursor].Name
			m.qbSt = qbEnterValue
			m.qbFieldInput.Blur()
			m.qbValueInput.SetValue("")
			m.qbValueInput.Focus()
			m.qbValueInput.Placeholder = fmt.Sprintf("value for %s...", m.qbPickedField)
			return m, m.qbValueInput.Cursor.BlinkCmd()
		}
	case tea.KeyEsc:
		m.qbSt = qbList
		m.qbFieldInput.Blur()
		return m, nil
	case tea.KeyUp:
		if m.qbFieldCursor > 0 {
			m.qbFieldCursor--
		}
		return m, nil
	case tea.KeyDown:
		if m.qbFieldCursor < len(m.qbFilteredFields)-1 {
			m.qbFieldCursor++
		}
		return m, nil
	}

	// Update field filter input
	var cmd tea.Cmd
	m.qbFieldInput, cmd = m.qbFieldInput.Update(msg)

	// Filter fields by typed text
	filter := strings.ToLower(m.qbFieldInput.Value())
	if filter == "" {
		m.qbFilteredFields = m.allFields
	} else {
		m.qbFilteredFields = nil
		for _, f := range m.allFields {
			if strings.Contains(strings.ToLower(f.Name), filter) {
				m.qbFilteredFields = append(m.qbFilteredFields, f)
			}
		}
	}
	if m.qbFieldCursor >= len(m.qbFilteredFields) {
		m.qbFieldCursor = max(0, len(m.qbFilteredFields)-1)
	}
	return m, cmd
}

func (m *Model) handleQBValueKey(msg tea.KeyMsg) (views.View, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		val := strings.TrimSpace(m.qbValueInput.Value())
		if val != "" {
			m.qbFilters = append(m.qbFilters, FilterCondition{
				Field:    m.qbPickedField,
				Value:    val,
				Operator: "AND",
			})
			m.qbCursor = len(m.qbFilters) - 1
		}
		m.qbSt = qbList
		m.qbValueInput.Blur()
		return m, nil
	case tea.KeyEsc:
		m.qbSt = qbList
		m.qbValueInput.Blur()
		return m, nil
	}

	var cmd tea.Cmd
	m.qbValueInput, cmd = m.qbValueInput.Update(msg)
	return m, cmd
}

// viewQueryBuilder renders the query builder popup
func (m *Model) viewQueryBuilder() string {
	var b strings.Builder
	b.WriteString(theme.ModalTitleStyle.Render("Query Builder") + "\n\n")

	switch m.qbSt {
	case qbSelectField:
		b.WriteString("  " + theme.HelpKeyStyle.Render("Select field:") + " " + m.qbFieldInput.View() + "\n\n")
		maxShow := 10
		for i, f := range m.qbFilteredFields {
			if i >= maxShow {
				b.WriteString(theme.HelpDescStyle.Render(fmt.Sprintf("  ... %d more\n", len(m.qbFilteredFields)-maxShow)))
				break
			}
			cursor := "  "
			style := theme.HelpDescStyle
			if i == m.qbFieldCursor {
				cursor = "> "
				style = theme.HelpKeyStyle
			}
			b.WriteString("  " + cursor + style.Render(f.Name) + theme.HelpDescStyle.Render(" ("+f.Type+")") + "\n")
		}
		if len(m.qbFilteredFields) == 0 {
			b.WriteString("  " + theme.HelpDescStyle.Render("(no matching fields)") + "\n")
		}

	case qbEnterValue:
		b.WriteString("  " + theme.HelpKeyStyle.Render(m.qbPickedField) + " " + m.qbValueInput.View() + "\n")

	case qbList:
		if len(m.qbFilters) == 0 {
			b.WriteString("  " + theme.HelpDescStyle.Render("No filters. Press n to add.") + "\n")
		}
		for i, f := range m.qbFilters {
			cursor := "  "
			if i == m.qbCursor {
				cursor = theme.HelpKeyStyle.Render("> ")
			}
			entry := theme.HelpKeyStyle.Render(f.Field) + theme.HelpDescStyle.Render(" = ") + theme.HelpDescStyle.Render(f.Value)
			b.WriteString(fmt.Sprintf("  %s%d. %s\n", cursor, i+1, entry))
			if i < len(m.qbFilters)-1 {
				opLabel := f.Operator
				if i == m.qbCursor {
					// Highlight that 'a' can toggle this operator
					opLabel = theme.HealthYellowStyle.Render("[ " + f.Operator + " ]")
				} else {
					opLabel = theme.HelpDescStyle.Render("  " + f.Operator)
				}
				b.WriteString("     " + opLabel + "\n")
			}
		}
	}

	b.WriteString("\n")
	switch m.qbSt {
	case qbList:
		help := "  n: add • d: delete • enter: apply • esc: cancel"
		if len(m.qbFilters) >= 2 {
			help = "  n: add • d: delete • a: toggle AND/OR • enter: apply • esc: cancel"
		}
		b.WriteString(theme.HelpDescStyle.Render(help))
	case qbSelectField:
		b.WriteString(theme.HelpDescStyle.Render("  type to filter • tab: complete • enter: pick • esc: cancel"))
	case qbEnterValue:
		b.WriteString(theme.HelpDescStyle.Render("  enter: confirm • esc: cancel"))
	}

	return theme.ModalStyle.Render(b.String())
}

func buildQueryString(filters []FilterCondition) string {
	if len(filters) == 0 {
		return ""
	}
	if len(filters) == 1 {
		return filters[0].Field + ":" + filters[0].Value
	}

	// Group consecutive filters by operator to build proper parenthesized query.
	// e.g., level:INFO AND (service:worker OR service:api)
	// Strategy: group OR'd filters together with parens.
	var parts []string
	i := 0
	for i < len(filters) {
		// Collect a group of OR'd filters starting from i
		group := []string{filters[i].Field + ":" + filters[i].Value}
		for i < len(filters)-1 && filters[i].Operator == "OR" {
			i++
			group = append(group, filters[i].Field+":"+filters[i].Value)
		}

		if len(group) > 1 {
			parts = append(parts, "("+strings.Join(group, " OR ")+")")
		} else {
			parts = append(parts, group[0])
		}

		if i < len(filters)-1 {
			parts = append(parts, filters[i].Operator)
		}
		i++
	}

	return strings.Join(parts, " ")
}
