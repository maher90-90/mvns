package history

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestAddAndList(t *testing.T) {
	dir := t.TempDir()
	h, err := New(filepath.Join(dir, "history.json"))
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	h.Add("guice")
	h.Add("junit")

	entries := h.List()
	if len(entries) != 2 {
		t.Fatalf("len = %d, want 2", len(entries))
	}
	if entries[0] != "junit" {
		t.Errorf("entries[0] = %q, want %q", entries[0], "junit")
	}
	if entries[1] != "guice" {
		t.Errorf("entries[1] = %q, want %q", entries[1], "guice")
	}
}

func TestNoDuplicates(t *testing.T) {
	dir := t.TempDir()
	h, err := New(filepath.Join(dir, "history.json"))
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	h.Add("guice")
	h.Add("junit")
	h.Add("guice")

	entries := h.List()
	if len(entries) != 2 {
		t.Fatalf("len = %d, want 2", len(entries))
	}
	if entries[0] != "guice" {
		t.Errorf("entries[0] = %q, want %q", entries[0], "guice")
	}
}

func TestMaxEntries(t *testing.T) {
	dir := t.TempDir()
	h, err := New(filepath.Join(dir, "history.json"))
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	for i := 0; i < 60; i++ {
		h.Add(fmt.Sprintf("query-%d", i))
	}

	if len(h.List()) != maxEntries {
		t.Errorf("len = %d, want %d", len(h.List()), maxEntries)
	}
}

func TestPersistence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")

	h1, _ := New(path)
	h1.Add("guice")
	h1.Save()

	h2, _ := New(path)
	entries := h2.List()
	if len(entries) != 1 || entries[0] != "guice" {
		t.Errorf("persistence failed: entries = %v", entries)
	}
}
