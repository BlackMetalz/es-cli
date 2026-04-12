package query

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kienlt/es-cli/internal/tui/views"
)

func (m *Model) handleKey(msg tea.KeyMsg) (views.View, tea.Cmd) {
	if m.editing {
		return m.handleEditingKey(msg)
	}
	if m.viewingDoc {
		return m.handleDocViewKey(msg)
	}
	if m.timeRangeActive {
		return m.handleTimeRangeKey(msg)
	}
	if m.pickingTime {
		return m.handleTimePickerKey(msg)
	}
	if m.buildingQuery {
		return m.handleQueryBuilderKey(msg)
	}
	if m.indexSearching {
		return m.handleIndexSearchKey(msg)
	}
	if m.selectingIndex {
		return m.handleIndexSelectKey(msg)
	}
	if m.pickingColumns {
		return m.handleColumnPickerKey(msg)
	}
	return m.handleNormalKey(msg)
}

func (m *Model) handleEditingKey(msg tea.KeyMsg) (views.View, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		m.query = m.queryInput.Value()
		m.editing = false
		m.queryInput.Blur()
		m.loading = true
		m.page = 0
		m.pageSorts = nil
		return m, m.executeSearch()
	case tea.KeyEsc:
		m.editing = false
		m.queryInput.Blur()
		return m, nil
	}
	var cmd tea.Cmd
	m.queryInput, cmd = m.queryInput.Update(msg)
	return m, cmd
}

func (m *Model) handleIndexSearchKey(msg tea.KeyMsg) (views.View, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		m.indexSearching = false
		m.indexInput.Blur()
		m.indexCursor = 0
		return m, nil
	case tea.KeyEsc:
		m.indexSearching = false
		m.indexInput.SetValue("")
		m.indexInput.Blur()
		m.filteredIdx = m.indices
		m.indexCursor = 0
		return m, nil
	}
	var cmd tea.Cmd
	m.indexInput, cmd = m.indexInput.Update(msg)
	m.filterIndices()
	return m, cmd
}

func (m *Model) handleIndexSelectKey(msg tea.KeyMsg) (views.View, tea.Cmd) {
	switch {
	case msg.String() == "j" || msg.Type == tea.KeyDown:
		if m.indexCursor < len(m.filteredIdx)-1 {
			m.indexCursor++
		}
		return m, nil
	case msg.String() == "k" || msg.Type == tea.KeyUp:
		if m.indexCursor > 0 {
			m.indexCursor--
		}
		return m, nil
	case msg.Type == tea.KeyEnter:
		if len(m.filteredIdx) > 0 {
			m.index = m.filteredIdx[m.indexCursor]
			m.selectingIndex = false
			m.loading = true
			return m, m.fetchMapping()
		}
		return m, nil
	case msg.String() == "/":
		m.indexSearching = true
		m.indexInput.Focus()
		return m, m.indexInput.Cursor.BlinkCmd()
	case key.Matches(msg, m.keys.Refresh):
		m.loading = true
		return m, m.fetchIndices()
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) handleColumnPickerKey(msg tea.KeyMsg) (views.View, tea.Cmd) {
	switch {
	case msg.String() == "j" || msg.Type == tea.KeyDown:
		if m.columnCursor < len(m.allFields)-1 {
			m.columnCursor++
		}
		return m, nil
	case msg.String() == "k" || msg.Type == tea.KeyUp:
		if m.columnCursor > 0 {
			m.columnCursor--
		}
		return m, nil
	case msg.String() == " ":
		if m.columnCursor < len(m.allFields) {
			name := m.allFields[m.columnCursor].Name
			m.columnSelected[name] = !m.columnSelected[name]
		}
		return m, nil
	case msg.Type == tea.KeyEnter || msg.Type == tea.KeyEsc:
		m.pickingColumns = false
		m.columns = nil
		for _, f := range m.allFields {
			if m.columnSelected[f.Name] {
				m.columns = append(m.columns, f.Name)
			}
		}
		if len(m.columns) == 0 {
			m.columns = defaultColumns(m.allFields)
		}
		m.rebuildTable()
		m.updateTableRows()
		return m, nil
	}
	return m, nil
}

func (m *Model) handleNormalKey(msg tea.KeyMsg) (views.View, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Search):
		// Open query builder
		m.buildingQuery = true
		m.qbSt = qbList
		// Pre-populate from existing filters if any, otherwise start empty
		if len(m.qbFilters) == 0 && m.query != "" {
			// Try to parse existing query back into filters (best effort)
			m.qbFilters = parseQueryToFilters(m.query)
		}
		m.qbCursor = 0
		return m, nil
	case key.Matches(msg, m.keys.RawQuery):
		// Open raw query input for advanced users
		m.editing = true
		m.queryInput.SetValue(m.query)
		m.queryInput.Focus()
		return m, m.queryInput.Cursor.BlinkCmd()
	case key.Matches(msg, m.keys.TimeRange):
		m.pickingTime = true
		m.timeCursor = m.timeRangeIdx
		m.timeCustomActive = false
		return m, nil
	case key.Matches(msg, m.keys.Follow):
		m.followIdx = (m.followIdx + 1) % len(followIntervals)
		m.followInterval = followIntervals[m.followIdx]
		if m.followInterval > 0 {
			m.page = 0
			m.pageSorts = nil
			return m, followTickCmd(m.followInterval)
		}
		return m, nil
	case key.Matches(msg, m.keys.NextPage):
		pageJump := m.height - 8
		if pageJump < 5 {
			pageJump = 5
		}
		cursor := m.table.Cursor() + pageJump
		if cursor >= len(m.hits) {
			cursor = len(m.hits) - 1
		}
		if cursor < 0 {
			cursor = 0
		}
		m.table.SetCursor(cursor)
		return m, nil
	case key.Matches(msg, m.keys.PrevPage):
		pageJump := m.height - 8
		if pageJump < 5 {
			pageJump = 5
		}
		cursor := m.table.Cursor() - pageJump
		if cursor < 0 {
			cursor = 0
		}
		m.table.SetCursor(cursor)
		return m, nil
	case key.Matches(msg, m.keys.Expand):
		cursor := m.table.Cursor()
		if cursor >= 0 && cursor < len(m.hits) {
			m.viewingDoc = true
			m.docScroll = 0
			m.docContent = formatDocDetail(m.hits[cursor].Source, m.allFields)
		}
		return m, nil
	case key.Matches(msg, m.keys.Columns):
		m.pickingColumns = true
		m.columnCursor = 0
		return m, nil
	case key.Matches(msg, m.keys.Refresh):
		if m.index != "" {
			m.page = 0
			m.pageSorts = nil
			m.loading = true
			return m, m.executeSearch()
		}
		return m, nil
	case key.Matches(msg, m.keys.Back):
		m.selectingIndex = true
		m.index = ""
		m.hits = nil
		m.total = 0
		m.followInterval = 0
		m.followIdx = 0
		m.loading = true
		return m, m.fetchIndices()
	case key.Matches(msg, m.keys.Help):
		return m, nil
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	}
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *Model) handleTimePickerKey(msg tea.KeyMsg) (views.View, tea.Cmd) {
	if m.timeCustomActive {
		switch msg.Type {
		case tea.KeyEnter:
			val := strings.TrimSpace(m.timeCustomInput.Value())
			d := parseCustomDuration(val)
			if d > 0 {
				// Replace the Relative entry with the parsed value
				relIdx := len(timeRanges) - 2
				m.timeRangeIdx = relIdx
				timeRanges[relIdx] = timeRange{Label: "Relative: " + val, Duration: d}
				m.pickingTime = false
				m.timeCustomActive = false
				m.timeCustomInput.Blur()
				if m.index != "" {
					m.loading = true
					return m, m.executeSearch()
				}
			}
			return m, nil
		case tea.KeyEsc:
			m.timeCustomActive = false
			m.timeCustomInput.Blur()
			return m, nil
		}
		var cmd tea.Cmd
		m.timeCustomInput, cmd = m.timeCustomInput.Update(msg)
		return m, cmd
	}

	switch {
	case msg.String() == "j" || msg.Type == tea.KeyDown:
		if m.timeCursor < len(timeRanges)-1 {
			m.timeCursor++
		}
	case msg.String() == "k" || msg.Type == tea.KeyUp:
		if m.timeCursor > 0 {
			m.timeCursor--
		}
	case msg.Type == tea.KeyEnter:
		tr := timeRanges[m.timeCursor]
		if tr.Duration == -1 {
			// Custom: open relative duration input
			m.timeCustomActive = true
			m.timeCustomInput.SetValue("")
			m.timeCustomInput.Focus()
			return m, m.timeCustomInput.Cursor.BlinkCmd()
		}
		if tr.Duration == -2 {
			// Custom Range: open from/to input
			m.pickingTime = false
			m.timeRangeActive = true
			m.timeRangeField = 0
			now := time.Now().Local()
			m.timeFromInput.SetValue(now.Add(-3 * time.Hour).Format("2006-01-02 15:04:05"))
			m.timeToInput.SetValue(now.Format("2006-01-02 15:04:05"))
			m.timeFromInput.Focus()
			m.timeToInput.Blur()
			return m, m.timeFromInput.Cursor.BlinkCmd()
		}
		m.timeRangeIdx = m.timeCursor
		m.pickingTime = false
		m.page = 0
		m.pageSorts = nil
		if m.index != "" {
			m.loading = true
			return m, m.executeSearch()
		}
	case msg.Type == tea.KeyEsc:
		m.pickingTime = false
	}
	return m, nil
}

func (m *Model) handleTimeRangeKey(msg tea.KeyMsg) (views.View, tea.Cmd) {
	switch msg.Type {
	case tea.KeyTab:
		// Toggle between from and to fields
		if m.timeRangeField == 0 {
			m.timeRangeField = 1
			m.timeFromInput.Blur()
			m.timeToInput.Focus()
			return m, m.timeToInput.Cursor.BlinkCmd()
		}
		m.timeRangeField = 0
		m.timeToInput.Blur()
		m.timeFromInput.Focus()
		return m, m.timeFromInput.Cursor.BlinkCmd()
	case tea.KeyEnter:
		fromStr := strings.TrimSpace(m.timeFromInput.Value())
		toStr := strings.TrimSpace(m.timeToInput.Value())
		fromTime := parseAbsoluteTime(fromStr)
		toTime := parseAbsoluteTime(toStr)
		if fromTime.IsZero() {
			return m, nil // invalid from, ignore
		}
		m.timeFrom = fromTime
		m.timeTo = toTime // can be zero (means "now")
		label := fromStr + " → "
		if toStr != "" {
			label += toStr
		} else {
			label += "now"
		}
		m.timeRangeIdx = len(timeRanges) - 1 // Custom Range index
		timeRanges[len(timeRanges)-1] = timeRange{Label: label, Duration: -2}
		m.timeRangeActive = false
		m.timeFromInput.Blur()
		m.timeToInput.Blur()
		m.page = 0
		m.pageSorts = nil
		if m.index != "" {
			m.loading = true
			return m, m.executeSearch()
		}
		return m, nil
	case tea.KeyEsc:
		m.timeRangeActive = false
		m.timeFromInput.Blur()
		m.timeToInput.Blur()
		return m, nil
	}
	// Route to active input
	var cmd tea.Cmd
	if m.timeRangeField == 0 {
		m.timeFromInput, cmd = m.timeFromInput.Update(msg)
	} else {
		m.timeToInput, cmd = m.timeToInput.Update(msg)
	}
	return m, cmd
}

func (m *Model) handleDocViewKey(msg tea.KeyMsg) (views.View, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.viewingDoc = false
		return m, nil
	case "]", "j", "down":
		// Next document
		cursor := m.table.Cursor()
		if cursor < len(m.hits)-1 {
			m.table.SetCursor(cursor + 1)
			m.docContent = formatDocDetail(m.hits[cursor+1].Source, m.allFields)
		}
		return m, nil
	case "[", "k", "up":
		// Previous document
		cursor := m.table.Cursor()
		if cursor > 0 {
			m.table.SetCursor(cursor - 1)
			m.docContent = formatDocDetail(m.hits[cursor-1].Source, m.allFields)
		}
		return m, nil
	}
	return m, nil
}
