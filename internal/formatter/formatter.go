package formatter

type Dependency struct {
	GroupID    string
	ArtifactID string
	Version    string
	Scope      string
}

type Formatter interface {
	Name() string
	Format(dep Dependency) string
	Lexer() string
}

func All() []Formatter {
	return []Formatter{
		&Maven{},
		&GradleGroovy{},
		&GradleKotlin{},
	}
}
