package history

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const maxEntries = 50

type History struct {
	path    string
	entries []string
}

func New(path string) (*History, error) {
	h := &History{path: path}

	data, err := os.ReadFile(path)
	if err == nil {
		json.Unmarshal(data, &h.entries)
	}

	return h, nil
}

func (h *History) Add(query string) {
	for i, e := range h.entries {
		if e == query {
			h.entries = append(h.entries[:i], h.entries[i+1:]...)
			break
		}
	}
	h.entries = append([]string{query}, h.entries...)
	if len(h.entries) > maxEntries {
		h.entries = h.entries[:maxEntries]
	}
}

func (h *History) List() []string {
	return h.entries
}

func (h *History) Save() error {
	if err := os.MkdirAll(filepath.Dir(h.path), 0755); err != nil {
		return err
	}
	data, err := json.Marshal(h.entries)
	if err != nil {
		return err
	}
	return os.WriteFile(h.path, data, 0644)
}
