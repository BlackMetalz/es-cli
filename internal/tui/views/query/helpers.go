package query

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kienlt/es-cli/internal/es"
)

func (m *Model) filterIndices() {
	filter := strings.ToLower(m.indexInput.Value())
	if filter == "" {
		m.filteredIdx = m.indices
		return
	}
	m.filteredIdx = nil
	for _, name := range m.indices {
		if strings.Contains(strings.ToLower(name), filter) {
			m.filteredIdx = append(m.filteredIdx, name)
		}
	}
	if m.indexCursor >= len(m.filteredIdx) {
		m.indexCursor = max(0, len(m.filteredIdx)-1)
	}
}

func (m *Model) rebuildTable() {
	if len(m.columns) == 0 || m.width == 0 {
		return
	}
	// Clear rows first to prevent panic when column count changes
	m.table.SetRows(nil)

	// # column (4 chars) + data columns
	numWidth := 5
	remaining := m.width - numWidth - 2
	colWidth := max(10, remaining/len(m.columns))

	cols := make([]table.Column, 0, len(m.columns)+1)
	cols = append(cols, table.Column{Title: "#", Width: numWidth})
	for _, c := range m.columns {
		cols = append(cols, table.Column{Title: c, Width: colWidth})
	}
	m.table.SetColumns(cols)
	m.table.SetWidth(m.width)
}

func (m *Model) updateTableRows() {
	rows := make([]table.Row, len(m.hits))
	for i, hit := range m.hits {
		row := make(table.Row, 0, len(m.columns)+1)
		row = append(row, fmt.Sprintf("%d", i+1))
		for _, col := range m.columns {
			row = append(row, getField(hit.Source, col))
		}
		rows[i] = row
	}
	m.table.SetRows(rows)
	if len(rows) > 0 {
		m.table.SetCursor(0)
	}
}

func getField(source map[string]interface{}, field string) string {
	parts := strings.Split(field, ".")
	var current interface{} = source
	for _, p := range parts {
		m, ok := current.(map[string]interface{})
		if !ok {
			return ""
		}
		current = m[p]
	}
	if current == nil {
		return ""
	}
	val := fmt.Sprintf("%v", current)

	// Convert @timestamp to local time
	if field == "@timestamp" {
		if t, err := time.Parse(time.RFC3339Nano, val); err == nil {
			return t.Local().Format("2006-01-02 15:04:05")
		}
		if t, err := time.Parse("2006-01-02T15:04:05.000Z", val); err == nil {
			return t.Local().Format("2006-01-02 15:04:05")
		}
	}

	return val
}

func defaultColumns(fields []es.FieldMapping) []string {
	cols := []string{}
	// Always include @timestamp if present
	for _, f := range fields {
		if f.Name == "@timestamp" {
			cols = append(cols, "@timestamp")
			break
		}
	}
	// Add first few keyword/text fields
	for _, f := range fields {
		if f.Name == "@timestamp" {
			continue
		}
		if f.Type == "keyword" || f.Type == "text" {
			cols = append(cols, f.Name)
			if len(cols) >= 5 {
				break
			}
		}
	}
	if len(cols) == 0 && len(fields) > 0 {
		for i := 0; i < len(fields) && i < 5; i++ {
			cols = append(cols, fields[i].Name)
		}
	}
	return cols
}

// parseQueryToFilters tries to parse a simple query string back into filter conditions.
// e.g., "level:error AND service:api" → [{level, error, AND}, {service, api, AND}]
func parseQueryToFilters(query string) []FilterCondition {
	if query == "" {
		return nil
	}
	var filters []FilterCondition
	parts := strings.Fields(query)
	i := 0
	for i < len(parts) {
		p := parts[i]
		if idx := strings.Index(p, ":"); idx > 0 {
			field := p[:idx]
			value := p[idx+1:]
			op := "AND"
			// Check if next token is an operator
			if i+1 < len(parts) {
				next := strings.ToUpper(parts[i+1])
				if next == "AND" || next == "OR" {
					op = next
					i++ // skip operator
				}
			}
			filters = append(filters, FilterCondition{Field: field, Value: value, Operator: op})
		}
		i++
	}
	return filters
}

// parseCustomDuration parses strings like "10d", "2h", "30m", "15s" into time.Duration.
func parseCustomDuration(s string) time.Duration {
	s = strings.TrimSpace(strings.ToLower(s))
	if len(s) < 2 {
		return 0
	}
	unit := s[len(s)-1]
	numStr := s[:len(s)-1]
	var num int
	for _, c := range numStr {
		if c < '0' || c > '9' {
			return 0
		}
		num = num*10 + int(c-'0')
	}
	if num <= 0 {
		return 0
	}
	switch unit {
	case 's':
		return time.Duration(num) * time.Second
	case 'm':
		return time.Duration(num) * time.Minute
	case 'h':
		return time.Duration(num) * time.Hour
	case 'd':
		return time.Duration(num) * 24 * time.Hour
	}
	return 0
}

// parseAbsoluteTime parses datetime strings in common formats.
// Supported: "2026-03-30 13:00:00", "2026-03-30T13:00:00", "2026-03-30 13:00", "2026-03-30".
func parseAbsoluteTime(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.ParseInLocation(f, s, time.Local); err == nil {
			return t
		}
	}
	return time.Time{}
}

// formatDocDetail formats a document source as a key-value table string.
func formatDocDetail(source map[string]interface{}, fields []es.FieldMapping) string {
	var b strings.Builder

	// Find max key width for alignment
	maxKey := 0
	var rows []struct{ k, v string }
	for _, f := range fields {
		val := getField(source, f.Name)
		if val == "" {
			continue
		}
		if len(f.Name) > maxKey {
			maxKey = len(f.Name)
		}
		rows = append(rows, struct{ k, v string }{f.Name, val})
	}

	// Also add any fields not in mapping
	for k, v := range source {
		found := false
		for _, r := range rows {
			if r.k == k {
				found = true
				break
			}
		}
		if !found {
			val := fmt.Sprintf("%v", v)
			if len(k) > maxKey {
				maxKey = len(k)
			}
			rows = append(rows, struct{ k, v string }{k, val})
		}
	}

	for _, r := range rows {
		pad := strings.Repeat(" ", maxKey-len(r.k))
		b.WriteString(fmt.Sprintf("  %s%s : %s\n", r.k, pad, r.v))
	}

	return b.String()
}

func followTickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return FollowTickMsg{}
	})
}
