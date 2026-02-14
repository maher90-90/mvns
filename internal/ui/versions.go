package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/maher/mvns/internal/api"
)

func (a *App) fetchVersions() tea.Cmd {
	g := a.selectedDoc.GroupID
	ar := a.selectedDoc.ArtifactID

	return func() tea.Msg {
		resp, err := a.client.Versions(g, ar, 200, false)
		return versionResultMsg{resp: resp, err: err}
	}
}

func (a *App) updateVersions(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case versionResultMsg:
		if msg.err != nil {
			a.err = msg.err
			a.statusMsg = a.locale.T("error.network")
			return a, nil
		}
		a.versions = msg.resp.Response.Docs
		a.stableVersions = nil
		a.preVersions = nil
		for _, v := range a.versions {
			if v.IsPreRelease() {
				a.preVersions = append(a.preVersions, v)
			} else {
				a.stableVersions = append(a.stableVersions, v)
			}
		}
		a.versionCursor = 0
		a.statusMsg = ""
		return a, nil

	case tea.KeyMsg:
		allVersions := a.allVersionsSorted()
		switch msg.String() {
		case "esc":
			a.screen = screenSearch
			return a, nil
		case "enter":
			if len(allVersions) > 0 {
				a.selectedVersion = allVersions[a.versionCursor]
				a.screen = screenSnippets
				a.formatIdx = 0
				a.selectedScope = a.selectedVersion.DetectScope()
				return a, nil
			}
		case "up", "k":
			if a.versionCursor > 0 {
				a.versionCursor--
			}
		case "down", "j":
			if a.versionCursor < len(allVersions)-1 {
				a.versionCursor++
			}
		}
	}

	return a, nil
}

func (a *App) allVersionsSorted() []api.Doc {
	var all []api.Doc
	all = append(all, a.stableVersions...)
	all = append(all, a.preVersions...)
	return all
}

func (a *App) viewVersions() string {
	var b strings.Builder

	name := fmt.Sprintf("%s:%s", a.selectedDoc.GroupID, a.selectedDoc.ArtifactID)
	b.WriteString("  " + a.theme.Title.Render(name) + "\n\n")

	if a.statusMsg != "" {
		b.WriteString("  " + a.theme.Error.Render(a.statusMsg) + "\n\n")
		return b.String()
	}

	allVersions := a.allVersionsSorted()
	if len(allVersions) == 0 {
		return b.String()
	}

	// Calculate available height
	// Header: Title(3)
	// Footer: Help(2)
	// Sections: Headers(2 lines each)
	reservedHeight := 6
	availableHeight := a.height - reservedHeight
	if availableHeight < 1 {
		availableHeight = 1
	}

	// Simple scrolling: just show a window of versions around the cursor
	startIdx := 0
	if len(allVersions) > availableHeight {
		startIdx = a.versionCursor - (availableHeight / 2)
		if startIdx < 0 {
			startIdx = 0
		}
		if startIdx+availableHeight > len(allVersions) {
			startIdx = len(allVersions) - availableHeight
		}
	}
	endIdx := startIdx + availableHeight
	if endIdx > len(allVersions) {
		endIdx = len(allVersions)
	}

	// Map version to their original section for rendering
	for i := startIdx; i < endIdx; i++ {
		v := allVersions[i]

		// Show section header if it's the first item of its type in the viewport
		// OR if it's the first item in the entire list
		isFirstInStable := len(a.stableVersions) > 0 && i == 0
		isFirstInPre := len(a.preVersions) > 0 && i == len(a.stableVersions)

		if isFirstInStable {
			b.WriteString("  " + a.theme.Subtitle.Render(a.locale.T("versions.stable")) + "\n")
			b.WriteString("  " + a.theme.Separator.Render(strings.Repeat("─", 40)) + "\n")
		} else if isFirstInPre {
			b.WriteString("  " + a.theme.Subtitle.Render(a.locale.T("versions.prerelease")) + "\n")
			b.WriteString("  " + a.theme.Separator.Render(strings.Repeat("─", 40)) + "\n")
		}

		line := fmt.Sprintf("  %-20s %s", v.Version, v.Time().Format("2006-01-02"))
		if i == a.versionCursor {
			b.WriteString(a.theme.Selected.Render(line) + "\n")
		} else {
			b.WriteString(a.theme.Normal.Render(line) + "\n")
		}
	}

	b.WriteString("\n  " + a.theme.Help.Render(a.locale.T("versions.help")))

	return b.String()
}
