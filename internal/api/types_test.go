package api

import (
	"encoding/json"
	"testing"
)

func TestParseSearchResponse(t *testing.T) {
	raw := `{
		"response": {
			"numFound": 100,
			"start": 0,
			"docs": [
				{
					"id": "com.google.inject:guice",
					"g": "com.google.inject",
					"a": "guice",
					"latestVersion": "7.0.0",
					"p": "jar",
					"timestamp": 1710460800000
				}
			]
		}
	}`

	var resp SearchResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if resp.Response.NumFound != 100 {
		t.Errorf("numFound = %d, want 100", resp.Response.NumFound)
	}
	if len(resp.Response.Docs) != 1 {
		t.Fatalf("docs len = %d, want 1", len(resp.Response.Docs))
	}

	doc := resp.Response.Docs[0]
	if doc.GroupID != "com.google.inject" {
		t.Errorf("groupId = %q, want %q", doc.GroupID, "com.google.inject")
	}
	if doc.ArtifactID != "guice" {
		t.Errorf("artifactId = %q, want %q", doc.ArtifactID, "guice")
	}
	if doc.LatestVersion != "7.0.0" {
		t.Errorf("latestVersion = %q, want %q", doc.LatestVersion, "7.0.0")
	}
	if doc.Packaging != "jar" {
		t.Errorf("packaging = %q, want %q", doc.Packaging, "jar")
	}
}

func TestParseVersionResponse(t *testing.T) {
	raw := `{
		"response": {
			"numFound": 3,
			"start": 0,
			"docs": [
				{"g": "com.google.inject", "a": "guice", "v": "7.0.0", "p": "jar", "timestamp": 1710460800000},
				{"g": "com.google.inject", "a": "guice", "v": "6.0.0", "p": "jar", "timestamp": 1683849600000},
				{"g": "com.google.inject", "a": "guice", "v": "7.0.0-rc1", "p": "jar", "timestamp": 1707782400000}
			]
		}
	}`

	var resp SearchResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if len(resp.Response.Docs) != 3 {
		t.Fatalf("docs len = %d, want 3", len(resp.Response.Docs))
	}
	if resp.Response.Docs[2].Version != "7.0.0-rc1" {
		t.Errorf("version = %q, want %q", resp.Response.Docs[2].Version, "7.0.0-rc1")
	}
}

func TestDocTimestamp(t *testing.T) {
	doc := Doc{Timestamp: 1710460800000}
	ts := doc.Time()
	if ts.Year() != 2024 || ts.Month() != 3 || ts.Day() != 15 {
		t.Errorf("time = %v, want 2024-03-15", ts)
	}
}

func TestDocIsPreRelease(t *testing.T) {
	tests := []struct {
		version string
		want    bool
	}{
		{"7.0.0", false},
		{"7.0.0-rc1", true},
		{"7.0.0-beta1", true},
		{"7.0.0-alpha1", true},
		{"7.0.0-SNAPSHOT", true},
		{"7.0.0-dev", true},
		{"7.0.0-M1", true},
		{"1.0.0.Final", false},
	}

	for _, tt := range tests {
		doc := Doc{Version: tt.version}
		if got := doc.IsPreRelease(); got != tt.want {
			t.Errorf("IsPreRelease(%q) = %v, want %v", tt.version, got, tt.want)
		}
	}
}
