package api

import "testing"

func TestBuildQuery(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"guice", "guice"},
		{"com.google.inject:guice", `g:"com.google.inject" AND a:"guice"`},
		{"com.google.inject", `g:"com.google.inject"`},
		{"junit", "junit"},
		{"org.apache.commons:commons-lang3", `g:"org.apache.commons" AND a:"commons-lang3"`},
		{"io.netty", `g:"io.netty"`},
	}

	for _, tt := range tests {
		got := BuildQuery(tt.input)
		if got != tt.want {
			t.Errorf("BuildQuery(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
