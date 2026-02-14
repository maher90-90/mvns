package ui

import "testing"

func TestNewThemeDark(t *testing.T) {
	theme := NewTheme("dark")
	if theme.Name != "dark" {
		t.Errorf("Name = %q, want %q", theme.Name, "dark")
	}
}

func TestNewThemeLight(t *testing.T) {
	theme := NewTheme("light")
	if theme.Name != "light" {
		t.Errorf("Name = %q, want %q", theme.Name, "light")
	}
}

func TestNewThemeDefaultsToDark(t *testing.T) {
	theme := NewTheme("invalid")
	if theme.Name != "dark" {
		t.Errorf("Name = %q, want %q", theme.Name, "dark")
	}
}
