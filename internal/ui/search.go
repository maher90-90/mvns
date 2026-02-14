package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/maher/mvns/internal/api"
)

func (a *App) doSearch() tea.Cmd {
	query := strings.TrimSpace(a.searchInput.Value())
	if query == "" {
		return nil
	}

	a.history.Add(query)
	a.history.Save()
	a.historyIdx = -1
	a.searching = true
	a.results = nil
	a.totalResults = 0

	start := a.page * a.perPage

	return func() tea.Msg {
		// Use Multimodal search for concurrent execution of specific queries
		resp, err := a.client.SearchMultimodal(query, a.perPage, start, false)
		return searchResultMsg{resp: resp, err: err}
	}
}

func (a *App) doRefresh() tea.Cmd {
	query := strings.TrimSpace(a.searchInput.Value())
	if query == "" {
		return nil
	}

	a.searching = true
	a.results = nil
	a.totalResults = 0
	
	start := a.page * a.perPage

	return func() tea.Msg {
		resp, err := a.client.SearchMultimodal(query, a.perPage, start, true)
		return searchResultMsg{resp: resp, err: err}
	}
}

func (a *App) prefetchNextPages(query string) tea.Cmd {
	return func() tea.Msg {
		// Fetch next 2 pages in the background
		for p := 1; p <= 2; p++ {
			targetPage := a.page + p
			if targetPage*a.perPage >= a.totalResults {
				break
			}
			start := targetPage * a.perPage
			// Just fire and forget, the client cache will handle the storage
			go a.client.SearchMultimodal(query, a.perPage, start, false)
		}
		return nil
	}
}

type prefetchMsg struct {
	cursor int
}

func (a *App) prefetch() tea.Cmd {
	if len(a.results) == 0 || a.resultCursor < 0 || a.resultCursor >= len(a.results) {
		return nil
	}

	cursor := a.resultCursor
	return tea.Tick(300*time.Millisecond, func(t time.Time) tea.Msg {
		return prefetchMsg{cursor: cursor}
	})
}

func (a *App) findSuggestion() {
	input := a.searchInput.Value()
	if input == "" {
		a.suggestion = ""
		return
	}

	entries := a.history.List()
	for _, entry := range entries {
		if strings.HasPrefix(strings.ToLower(entry), strings.ToLower(input)) && len(entry) > len(input) {
			a.suggestion = entry[len(input):]
			return
		}
	}
	a.suggestion = ""
}

func (a *App) updateSearch(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case prefetchMsg:
		if !a.searchInput.Focused() && msg.cursor == a.resultCursor {
			doc := a.results[a.resultCursor]
			return a, func() tea.Msg {
				_, _ = a.client.Versions(doc.GroupID, doc.ArtifactID, 200, false)
				return nil
			}
		}
		return a, nil

	case searchResultMsg:
		a.searching = false
		if msg.err != nil {
			a.err = msg.err
			a.statusMsg = a.locale.T("error.network")
			return a, nil
		}
		a.results = msg.resp.Response.Docs
		// Deduplicate results by ID
		uniqueResults := make([]api.Doc, 0, len(a.results))
		seen := make(map[string]bool)
		for _, doc := range a.results {
			if !seen[doc.ID] {
				seen[doc.ID] = true
				uniqueResults = append(uniqueResults, doc)
			}
		}
		a.results = uniqueResults

		// Sort results: 
		// 1. Exact ArtifactID match
		// 2. GroupID contains query (official package proxy)
		// 3. VersionCount (popularity)
		// 4. Timestamp (recency)
		query := strings.TrimSpace(a.searchInput.Value())
		sort.Slice(a.results, func(i, j int) bool {
			// Tier 1: Exact artifactId match
			iExact := a.results[i].ArtifactID == query
			jExact := a.results[j].ArtifactID == query
			if iExact != jExact {
				return iExact
			}

			// Tier 2: GroupID contains query (e.g. org.junit.jupiter contains junit-jupiter)
			// We normalize dots to hyphens to catch matches like org.junit.jupiter vs junit-jupiter
			iNormalizedGroup := strings.ReplaceAll(strings.ToLower(a.results[i].GroupID), ".", "-")
			jNormalizedGroup := strings.ReplaceAll(strings.ToLower(a.results[j].GroupID), ".", "-")
			lowerQuery := strings.ToLower(query)

			iGroupMatch := strings.Contains(iNormalizedGroup, lowerQuery)
			jGroupMatch := strings.Contains(jNormalizedGroup, lowerQuery)
			if iGroupMatch != jGroupMatch {
				return iGroupMatch
			}

			// Tier 3: Version count as popularity proxy
			if a.results[i].VersionCount != a.results[j].VersionCount {
				return a.results[i].VersionCount > a.results[j].VersionCount
			}

			// Tier 4: Timestamp as recency
			return a.results[i].Timestamp > a.results[j].Timestamp
		})

		// Limit to perPage for display
		if len(a.results) > a.perPage {
			a.results = a.results[:a.perPage]
		}

		a.totalResults = msg.resp.Response.NumFound
		a.resultCursor = 0
		a.err = nil
		a.statusMsg = ""
		if len(a.results) == 0 {
			a.statusMsg = a.locale.T("error.noresults")
		}

		// Trigger prefetch for current item versions
		cmds := []tea.Cmd{a.prefetch()}

		// Trigger background prefetch for next 2 pages
		if a.totalResults > (a.page+1)*a.perPage {
			cmds = append(cmds, a.prefetchNextPages(query))
		}

		return a, tea.Batch(cmds...)

	case tea.KeyMsg:
		k := msg.String()

		// 1. History navigation: only when focused and (empty OR already in history)
		if a.searchInput.Focused() && (a.searchInput.Value() == "" || a.historyIdx != -1) {
			entries := a.history.List()
			if k == "up" || k == "down" {
				switch k {
				case "up":
					if len(entries) > 0 && a.historyIdx < len(entries)-1 {
						a.historyIdx++
						a.searchInput.SetValue(entries[a.historyIdx])
						a.searchInput.SetCursor(len(entries[a.historyIdx]))
						a.suggestion = ""
						return a, nil
					}
				case "down":
					if a.historyIdx > 0 {
						a.historyIdx--
						a.searchInput.SetValue(entries[a.historyIdx])
						a.searchInput.SetCursor(len(entries[a.historyIdx]))
						a.suggestion = ""
						return a, nil
					} else if a.historyIdx == 0 {
						a.historyIdx = -1
						a.searchInput.SetValue("")
						a.suggestion = ""
						return a, nil
					}
				}
			}
		}

		// 2. Global / Navigation keys
		switch k {
		case "tab", "right":
			if a.searchInput.Focused() && a.suggestion != "" {
				a.searchInput.SetValue(a.searchInput.Value() + a.suggestion)
				a.searchInput.SetCursor(len(a.searchInput.Value()))
				a.suggestion = ""
				return a, nil
			}
		case "ctrl+r":
			return a, a.doRefresh()
		case "enter":
			if a.searchInput.Focused() {
				a.searchInput.Blur()
				a.page = 0
				a.resultCursor = 0
				a.suggestion = ""
				return a, a.doSearch()
			}
			if len(a.results) > 0 {
				a.selectedDoc = a.results[a.resultCursor]
				a.screen = screenVersions
				a.versionCursor = 0
				return a, a.fetchVersions()
			}
		case "esc":
			if !a.searchInput.Focused() && len(a.results) > 0 {
				a.searchInput.Focus()
				return a, nil
			}
			return a, tea.Quit
		case "up":
			if !a.searchInput.Focused() {
				if a.resultCursor > 0 {
					a.resultCursor--
					return a, a.prefetch()
				} else {
					a.searchInput.Focus()
				}
				return a, nil
			}
		case "down":
			if a.searchInput.Focused() {
				if len(a.results) > 0 {
					a.searchInput.Blur()
					a.resultCursor = 0
					a.suggestion = ""
					return a, a.prefetch()
				}
				return a, nil
			}
			if a.resultCursor < len(a.results)-1 {
				a.resultCursor++
				return a, a.prefetch()
			}
			return a, nil
		}

		// 3. TUI-only keys (vim-style, paging) - only when input NOT focused
		if !a.searchInput.Focused() {
			switch k {
			case "k":
				if a.resultCursor > 0 {
					a.resultCursor--
					return a, a.prefetch()
				} else {
					a.searchInput.Focus()
				}
			case "j":
				if a.resultCursor < len(a.results)-1 {
					a.resultCursor++
					return a, a.prefetch()
				}
			case "/":
				a.searchInput.Focus()
			case "n":
				maxPage := (a.totalResults - 1) / a.perPage
				if a.page < maxPage {
					a.page++
					return a, a.doSearch()
				}
			case "p":
				if a.page > 0 {
					a.page--
					return a, a.doSearch()
				}
			}
			return a, nil
		}
	}

	var cmd tea.Cmd
	oldVal := a.searchInput.Value()
	a.searchInput, cmd = a.searchInput.Update(msg)
	if a.searchInput.Value() != oldVal {
		a.findSuggestion()
	}
	return a, cmd
}

func (a *App) viewSearch() string {
	var b strings.Builder

	title := a.theme.Title.Render(a.locale.T("search.title"))
	if a.searching {
		title = fmt.Sprintf("%s %s %s", title, a.spinner.View(), a.theme.Dimmed.Render(a.locale.T("search.searching")))
	}
	b.WriteString(title + "\n\n")

	// Render label + Search input + suggestion
	label := a.theme.Normal.Render(a.locale.T("search.label"))
	
	// Temporarily shrink width to match content for seamless ghost text
	originalWidth := a.searchInput.Width
	contentLen := len(a.searchInput.Value())
	if contentLen == 0 {
		a.searchInput.Width = 1 
	} else {
		a.searchInput.Width = contentLen + 1
	}
	inputView := a.searchInput.View()
	a.searchInput.Width = originalWidth

	line := "  " + label + inputView
	if a.searchInput.Focused() && a.suggestion != "" {
		line += a.theme.Dimmed.Render(a.suggestion)
	}
	b.WriteString(line + "\n\n")

	if a.statusMsg != "" {
		b.WriteString("  " + a.theme.Error.Render(a.statusMsg) + "\n\n")
	}

	// Calculate available height for results
	// Header: title(2) + input(3) + status(maybe 2) = ~5-7 lines
	// Footer: page(1) + help(2) = 3 lines
	headerHeight := 5
	if a.statusMsg != "" {
		headerHeight += 2
	}
	footerHeight := 4
	availableHeight := a.height - headerHeight - footerHeight
	if availableHeight < 0 {
		availableHeight = 0
	}

	// Each result takes 3 lines
	maxVisibleResults := availableHeight / 3
	if maxVisibleResults < 1 {
		maxVisibleResults = 1
	}

	// Scrolling logic
	startIdx := 0
	if len(a.results) > maxVisibleResults {
		startIdx = a.resultCursor - (maxVisibleResults / 2)
		if startIdx < 0 {
			startIdx = 0
		}
		if startIdx+maxVisibleResults > len(a.results) {
			startIdx = len(a.results) - maxVisibleResults
		}
	}
	endIdx := startIdx + maxVisibleResults
	if endIdx > len(a.results) {
		endIdx = len(a.results)
	}

	for i := startIdx; i < endIdx; i++ {
		doc := a.results[i]
		name := fmt.Sprintf("%s:%s", doc.GroupID, doc.ArtifactID)
		version := doc.LatestVersion
		if version == "" {
			version = doc.Version
		}

		versionCountStr := fmt.Sprintf(a.locale.T("results.versionCount"), doc.VersionCount)

		var line1, line2 string
		if i == a.resultCursor && !a.searchInput.Focused() {
			selectedStyle := a.theme.Selected.Copy()
			if a.width > 2 {
				selectedStyle = selectedStyle.Width(a.width - 2)
			}
			line1 = selectedStyle.Render(fmt.Sprintf("> %-55s %s", name, "v"+version))
			line2 = selectedStyle.Render(fmt.Sprintf("  %s | %s | %s", doc.Time().Format("2006-01-02"), doc.Packaging, versionCountStr))
		} else {
			line1 = "  " + a.theme.Normal.Render(fmt.Sprintf("%-55s %s", name, "v"+version))
			line2 = "  " + a.theme.Dimmed.Render(fmt.Sprintf("%s | %s | %s", doc.Time().Format("2006-01-02"), doc.Packaging, versionCountStr))
		}

		b.WriteString(line1 + "\n")
		b.WriteString(line2 + "\n\n")
	}

	// Fill remaining space to keep footer at the bottom if desired, 
	// or just let it float. We'll let it float for now but ensure it's not off-screen.

	if a.totalResults > 0 {
		totalPages := (a.totalResults-1)/a.perPage + 1
		pageInfo := fmt.Sprintf(a.locale.T("results.page"), a.page+1, totalPages)
		itemsInfo := fmt.Sprintf(a.locale.T("results.itemsCount"), len(a.results))
		
		// Calculate spacing for right alignment
		padding := a.width - lipgloss.Width(pageInfo) - lipgloss.Width(itemsInfo) - 4
		if padding < 1 {
			padding = 1
		}
		
		footer := "  " + pageInfo + strings.Repeat(" ", padding) + itemsInfo
		b.WriteString(a.theme.Dimmed.Render(footer) + "\n")
	}

	if a.searchInput.Focused() {
		b.WriteString("\n  " + a.theme.Help.Render(a.locale.T("search.help")))
	} else {
		b.WriteString("\n  " + a.theme.Help.Render(a.locale.T("results.help")))
	}

	return b.String()
}
