package i18n

import (
	"embed"
	"testing"
)

//go:embed testdata/*.json
var testLocales embed.FS

func TestLoadAndTranslate(t *testing.T) {
	loc, err := NewFromFS(testLocales, "testdata", "en")
	if err != nil {
		t.Fatalf("NewFromFS failed: %v", err)
	}

	if got := loc.T("search.placeholder"); got != "Search packages..." {
		t.Errorf("T(search.placeholder) = %q, want %q", got, "Search packages...")
	}
}

func TestGermanTranslation(t *testing.T) {
	loc, err := NewFromFS(testLocales, "testdata", "de")
	if err != nil {
		t.Fatalf("NewFromFS failed: %v", err)
	}

	if got := loc.T("search.placeholder"); got != "Pakete suchen..." {
		t.Errorf("T(search.placeholder) = %q, want %q", got, "Pakete suchen...")
	}
}

func TestFallbackToKey(t *testing.T) {
	loc, err := NewFromFS(testLocales, "testdata", "en")
	if err != nil {
		t.Fatalf("NewFromFS failed: %v", err)
	}

	if got := loc.T("nonexistent.key"); got != "nonexistent.key" {
		t.Errorf("T(nonexistent.key) = %q, want key back", got)
	}
}
