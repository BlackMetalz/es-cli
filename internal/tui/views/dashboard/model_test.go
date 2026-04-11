package dashboard

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kienlt/es-cli/internal/es"
)

func testClient() *es.Client {
	return es.NewClient("http://localhost:9200", "elastic", "elastic")
}

func testData() *es.DashboardData {
	return &es.DashboardData{
		Health:            "green",
		HealthDescription: "All shards assigned",
		Version:           "8.17.0",
		Uptime:            3*time.Hour + 15*time.Minute,
		License:           "basic",
		NodeCount:         3,
		DiskAvailBytes:    53687091200,
		DiskTotalBytes:    107374182400,
		HeapUsedBytes:     536870912,
		HeapMaxBytes:      1073741824,
		IndexCount:        45,
		DocCount:          123456,
		DiskUsage:         "5.0gb",
		PrimaryShards:     45,
		ReplicaShards:     45,
	}
}

func TestNew(t *testing.T) {
	m := New(testClient())
	if m.Name() != "Dashboard" {
		t.Fatalf("expected Dashboard, got %s", m.Name())
	}
	if !m.loading {
		t.Fatal("expected loading=true")
	}
}

func TestUpdate_DashboardLoaded(t *testing.T) {
	m := New(testClient())
	v, _ := m.Update(DashboardLoadedMsg{Data: testData()})
	updated := v.(*Model)

	if updated.loading {
		t.Fatal("expected loading=false")
	}
	if updated.data == nil {
		t.Fatal("expected data set")
	}
	if updated.data.Health != "green" {
		t.Fatalf("expected green, got %s", updated.data.Health)
	}
}

func TestUpdate_ErrorMsg(t *testing.T) {
	m := New(testClient())
	v, _ := m.Update(ErrorMsg{Err: nil})
	updated := v.(*Model)

	if updated.loading {
		t.Fatal("expected loading=false")
	}
}

func TestUpdate_Refresh(t *testing.T) {
	m := New(testClient())
	m.loading = false

	v, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	updated := v.(*Model)

	if !updated.loading {
		t.Fatal("expected loading=true after refresh")
	}
	if cmd == nil {
		t.Fatal("expected command")
	}
}

func TestView_Loading(t *testing.T) {
	m := New(testClient())
	view := m.View()
	if !strings.Contains(view, "Loading") {
		t.Fatal("expected loading text")
	}
}

func TestView_Rendered(t *testing.T) {
	m := New(testClient())
	m.loading = false
	m.data = testData()
	m.width = 120
	m.height = 40

	view := m.View()
	if !strings.Contains(view, "Overview") {
		t.Fatal("expected Overview section")
	}
	if !strings.Contains(view, "Nodes") {
		t.Fatal("expected Nodes section")
	}
	if !strings.Contains(view, "Indices") {
		t.Fatal("expected Indices section")
	}
	if !strings.Contains(view, "8.17.0") {
		t.Fatal("expected version")
	}
	if !strings.Contains(view, "123,456") {
		t.Fatal("expected formatted doc count")
	}
}

func TestHelpGroups(t *testing.T) {
	m := New(testClient())
	groups := m.HelpGroups()
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if groups[0].Title != "Dashboard" {
		t.Fatalf("expected Dashboard, got %s", groups[0].Title)
	}
}

func TestIsInputMode(t *testing.T) {
	m := New(testClient())
	if m.IsInputMode() {
		t.Fatal("expected false")
	}
}

func TestStatusInfo(t *testing.T) {
	m := New(testClient())
	m.data = testData()
	info := m.StatusInfo()
	if !strings.Contains(info, "green") {
		t.Fatalf("expected green in status, got %s", info)
	}
}

func TestSetSize(t *testing.T) {
	m := New(testClient())
	m.SetSize(120, 40)
	if m.width != 120 {
		t.Fatalf("expected 120, got %d", m.width)
	}
}

func TestFormatUptime(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{0, "n/a"},
		{30 * time.Minute, "30m"},
		{2*time.Hour + 15*time.Minute, "2h 15m"},
		{26*time.Hour + 30*time.Minute, "1d 2h 30m"},
	}
	for _, tt := range tests {
		got := formatUptime(tt.d)
		if got != tt.want {
			t.Errorf("formatUptime(%v) = %s, want %s", tt.d, got, tt.want)
		}
	}
}

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		n    int64
		want string
	}{
		{0, "0"},
		{999, "999"},
		{1000, "1,000"},
		{123456, "123,456"},
		{1234567, "1,234,567"},
	}
	for _, tt := range tests {
		got := formatNumber(tt.n)
		if got != tt.want {
			t.Errorf("formatNumber(%d) = %s, want %s", tt.n, got, tt.want)
		}
	}
}
