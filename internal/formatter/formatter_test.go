package formatter

import "testing"

func TestMavenFormat(t *testing.T) {
	f := &Maven{}
	got := f.Format(Dependency{
		GroupID:    "com.google.inject",
		ArtifactID: "guice",
		Version:    "7.0.0",
	})

	want := `<dependency>
    <groupId>com.google.inject</groupId>
    <artifactId>guice</artifactId>
    <version>7.0.0</version>
</dependency>`

	if got != want {
		t.Errorf("Maven.Format():\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestMavenName(t *testing.T) {
	f := &Maven{}
	if f.Name() != "Maven" {
		t.Errorf("Name() = %q, want %q", f.Name(), "Maven")
	}
}

func TestGradleGroovyFormat(t *testing.T) {
	f := &GradleGroovy{}
	got := f.Format(Dependency{
		GroupID:    "com.google.inject",
		ArtifactID: "guice",
		Version:    "7.0.0",
	})
	want := "implementation 'com.google.inject:guice:7.0.0'"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestGradleKotlinFormat(t *testing.T) {
	f := &GradleKotlin{}
	got := f.Format(Dependency{
		GroupID:    "com.google.inject",
		ArtifactID: "guice",
		Version:    "7.0.0",
	})
	want := `implementation("com.google.inject:guice:7.0.0")`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestAllFormatters(t *testing.T) {
	formatters := All()
	if len(formatters) != 3 {
		t.Errorf("All() len = %d, want 3", len(formatters))
	}
}
