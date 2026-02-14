package formatter

import "fmt"

type GradleKotlin struct{}

func (g *GradleKotlin) Name() string { return "Gradle Kotlin DSL" }

func (g *GradleKotlin) Lexer() string { return "kotlin" }

func (g *GradleKotlin) Format(dep Dependency) string {
	config := "implementation"
	if dep.Scope == "test" {
		config = "testImplementation"
	} else if dep.Scope == "provided" {
		config = "compileOnly"
	} else if dep.Scope == "runtime" {
		config = "runtimeOnly"
	}
	return fmt.Sprintf(`%s("%s:%s:%s")`, config, dep.GroupID, dep.ArtifactID, dep.Version)
}
