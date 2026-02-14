package ui

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/maher/mvns/internal/formatter"
)

type clipboardMsg struct {
	err error
}

func (a *App) highlight(code, language string) string {
	lexer := lexers.Get(language)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	styleName := "dracula"
	if a.theme.Name == "light" {
		styleName = "friendly"
	}
	style := styles.Get(styleName)
	if style == nil {
		style = styles.Fallback
	}
	formatter := formatters.Get("terminal16m")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return code
	}

	var b bytes.Buffer
	err = formatter.Format(&b, style, iterator)
	if err != nil {
		return code
	}

	return b.String()
}

func (a *App) updateSnippets(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case clipboardMsg:
		if msg.err != nil {
			a.statusMsg = msg.err.Error()
		} else {
			a.statusMsg = a.locale.T("snippets.copied")
		}
		return a, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			a.screen = screenVersions
			a.statusMsg = ""
			return a, nil
		case "tab", "right", "l":
			a.formatIdx = (a.formatIdx + 1) % len(a.formatters)
			a.statusMsg = ""
		case "shift+tab", "left", "h":
			a.formatIdx = (a.formatIdx - 1 + len(a.formatters)) % len(a.formatters)
			a.statusMsg = ""
		case "c":
			// Toggle scopes: "" (compile) -> "test" -> "provided" -> "runtime"
			scopes := []string{"", "test", "provided", "runtime"}
			for i, s := range scopes {
				if s == a.selectedScope {
					a.selectedScope = scopes[(i+1)%len(scopes)]
					break
				}
			}
			a.statusMsg = ""
		case "enter":
			snippet := a.currentSnippet()
			return a, func() tea.Msg {
				err := clipboard.WriteAll(snippet)
				return clipboardMsg{err: err}
			}
		}
	}

	return a, nil
}

func (a *App) currentSnippet() string {
	dep := formatter.Dependency{
		GroupID:    a.selectedVersion.GroupID,
		ArtifactID: a.selectedVersion.ArtifactID,
		Version:    a.selectedVersion.Version,
		Scope:      a.selectedScope,
	}
	return a.formatters[a.formatIdx].Format(dep)
}

func (a *App) viewSnippets() string {
	var b strings.Builder

	f := a.formatters[a.formatIdx]
	name := fmt.Sprintf("%s:%s:%s", a.selectedVersion.GroupID, a.selectedVersion.ArtifactID, a.selectedVersion.Version)
	b.WriteString("  " + a.theme.Title.Render(name) + "\n\n")

	var tabs []string
	for i, fmtObj := range a.formatters {
		if i == a.formatIdx {
			tabs = append(tabs, a.theme.TabActive.Render("["+fmtObj.Name()+"]"))
		} else {
			tabs = append(tabs, a.theme.TabInactive.Render(" "+fmtObj.Name()+" "))
		}
	}
	b.WriteString("  " + strings.Join(tabs, "  ") + "\n\n")

	// Display scope
	scopeName := a.selectedScope
	if scopeName == "" {
		scopeName = "compile"
	}
	b.WriteString("  " + a.theme.Dimmed.Render("Scope: ") + a.theme.Normal.Render(scopeName) + " (press 'c' to change)\n\n")

	// Cache lookup - include scope in key
	cacheKey := fmt.Sprintf("%s:%s:%s:%s:%s", a.selectedVersion.ID, a.selectedVersion.Version, f.Name(), a.theme.Name, a.selectedScope)
	highlighted, ok := a.snippetCache[cacheKey]
	if !ok {
		snippet := a.currentSnippet()
		highlighted = a.highlight(snippet, f.Lexer())
		a.snippetCache[cacheKey] = highlighted
	}

	for _, line := range strings.Split(highlighted, "\n") {
		b.WriteString("  " + line + "\n")
	}

	if a.statusMsg != "" {
		b.WriteString("\n  " + a.theme.Success.Render(a.statusMsg))
	}

	b.WriteString("\n\n  " + a.theme.Help.Render(a.locale.T("snippets.help")))

	return b.String()
}
