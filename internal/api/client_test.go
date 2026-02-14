package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("q") != "guice" {
			t.Errorf("query = %q, want %q", r.URL.Query().Get("q"), "guice")
		}
		if r.URL.Query().Get("rows") != "20" {
			t.Errorf("rows = %q, want %q", r.URL.Query().Get("rows"), "20")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"response": {
				"numFound": 1,
				"start": 0,
				"docs": [{"id":"com.google.inject:guice","g":"com.google.inject","a":"guice","latestVersion":"7.0.0","p":"jar","timestamp":1710460800000}]
			}
		}`))
	}))
	defer server.Close()

	c := NewClient(WithBaseURL(server.URL))
	resp, err := c.Search("guice", 20, 0)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if resp.Response.NumFound != 1 {
		t.Errorf("numFound = %d, want 1", resp.Response.NumFound)
	}
	if resp.Response.Docs[0].ArtifactID != "guice" {
		t.Errorf("artifactId = %q, want %q", resp.Response.Docs[0].ArtifactID, "guice")
	}
}

func TestClientVersions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("core") != "gav" {
			t.Errorf("core = %q, want %q", r.URL.Query().Get("core"), "gav")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"response": {
				"numFound": 2,
				"start": 0,
				"docs": [
					{"g":"com.google.inject","a":"guice","v":"7.0.0","p":"jar","timestamp":1710460800000},
					{"g":"com.google.inject","a":"guice","v":"6.0.0","p":"jar","timestamp":1683849600000}
				]
			}
		}`))
	}))
	defer server.Close()

	c := NewClient(WithBaseURL(server.URL))
	resp, err := c.Versions("com.google.inject", "guice", 20)
	if err != nil {
		t.Fatalf("Versions failed: %v", err)
	}
	if len(resp.Response.Docs) != 2 {
		t.Fatalf("docs len = %d, want 2", len(resp.Response.Docs))
	}
}
