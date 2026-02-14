package formatter

import "fmt"

type GradleGroovy struct{}

func (g *GradleGroovy) Name() string { return "Gradle Groovy" }

func (g *GradleGroovy) Lexer() string { return "groovy" }

func (g *GradleGroovy) Format(dep Dependency) string {
	config := "implementation"
	if dep.Scope == "test" {
		config = "testImplementation"
	} else if dep.Scope == "provided" {
		config = "compileOnly"
	} else if dep.Scope == "runtime" {
		config = "runtimeOnly"
	}
	return fmt.Sprintf("%s '%s:%s:%s'", config, dep.GroupID, dep.ArtifactID, dep.Version)
}
