package api

import (
	"strings"
	"time"
)

type SearchResponse struct {
	Response ResponseBody `json:"response"`
}

type ResponseBody struct {
	NumFound int   `json:"numFound"`
	Start    int   `json:"start"`
	Docs     []Doc `json:"docs"`
}

type Doc struct {
	ID            string `json:"id"`
	GroupID       string `json:"g"`
	ArtifactID    string `json:"a"`
	LatestVersion string `json:"latestVersion"`
	Version       string `json:"v"`
	Packaging     string `json:"p"`
	Timestamp     int64  `json:"timestamp"`
	VersionCount  int    `json:"versionCount"`
}

func (d Doc) Time() time.Time {
	return time.UnixMilli(d.Timestamp)
}

func (d Doc) IsPreRelease() bool {
	v := strings.ToLower(d.Version)
	if v == "" {
		v = strings.ToLower(d.LatestVersion)
	}
	prefixes := []string{"-rc", "-beta", "-alpha", "-snapshot", "-dev", "-m"}
	for _, p := range prefixes {
		if strings.Contains(v, p) {
			return true
		}
	}
	return false
}

func (d Doc) DetectScope() string {
	id := strings.ToLower(d.GroupID + ":" + d.ArtifactID)
	testKeywords := []string{"junit", "mockito", "testcontainers", "assertj", "hamcrest", "test"}
	for _, kw := range testKeywords {
		if strings.Contains(id, kw) {
			return "test"
		}
	}
	return ""
}
