package index

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
)

func TestDefaultKeyMap(t *testing.T) {
	km := DefaultKeyMap()

	tests := []struct {
		name    string
		binding key.Binding
		key     string
		desc    string
	}{
		{"SortByName", km.SortByName, "Shift+I", "sort by index"},
		{"SortBySize", km.SortBySize, "Shift+S", "sort by size"},
		{"SortByDocs", km.SortByDocs, "Shift+C", "sort by docs"},
		{"ToggleOpenClose", km.ToggleOpenClose, "o", "open/close index"},
		{"DeleteIndex", km.DeleteIndex, "d", "delete index"},
		{"CreateIndex", km.CreateIndex, "n", "new index"},
		{"ToggleAll", km.ToggleAll, "a", "toggle hidden"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			help := tt.binding.Help()
			if help.Key != tt.key {
				t.Errorf("expected key %s, got %s", tt.key, help.Key)
			}
			if help.Desc != tt.desc {
				t.Errorf("expected desc %s, got %s", tt.desc, help.Desc)
			}
		})
	}
}
