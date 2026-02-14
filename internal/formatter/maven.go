package formatter

import "fmt"

type Maven struct{}

func (m *Maven) Name() string { return "Maven" }

func (m *Maven) Lexer() string { return "xml" }

func (m *Maven) Format(dep Dependency) string {
	scopeTag := ""
	if dep.Scope != "" {
		scopeTag = fmt.Sprintf("\n    <scope>%s</scope>", dep.Scope)
	}
	return fmt.Sprintf(`<dependency>
    <groupId>%s</groupId>
    <artifactId>%s</artifactId>
    <version>%s</version>%s
</dependency>`, dep.GroupID, dep.ArtifactID, dep.Version, scopeTag)
}
