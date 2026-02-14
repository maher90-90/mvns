package ui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/maher/mvns/internal/api"
	"github.com/maher/mvns/internal/formatter"
	"github.com/maher/mvns/internal/history"
	"github.com/maher/mvns/internal/i18n"
)

type screen int

const (
	screenSearch screen = iota
	screenVersions
	screenSnippets
)

type App struct {
	client     *api.Client
	locale     *i18n.Locale
	theme      *Theme
	formatters []formatter.Formatter
	history    *history.History

	screen    screen
	width     int
	height    int
	err       error
	statusMsg string
	searching bool
	spinner   spinner.Model

	// Search screen
	searchInput  textinput.Model
	suggestion   string
	results      []api.Doc
	resultCursor int
	totalResults int
	page         int
	perPage      int
	historyIdx   int
	prefetchIdx  int

	// Version screen
	selectedDoc    api.Doc
	versions       []api.Doc
	stableVersions []api.Doc
	preVersions    []api.Doc
	versionCursor  int

	// Snippet screen
	selectedVersion api.Doc
	formatIdx       int
	selectedScope   string
	snippetCache    map[string]string
}

type searchResultMsg struct {
	resp *api.SearchResponse
	err  error
}

type versionResultMsg struct {
	resp *api.SearchResponse
	err  error
}

func NewApp(client *api.Client, locale *i18n.Locale, theme *Theme, hist *history.History) *App {
	ti := textinput.New()
	ti.Placeholder = locale.T("search.placeholder")
	ti.Focus()
	ti.Prompt = ""
	ti.Width = 60

	s := spinner.New()
	s.Spinner = spinner.Pulse
	s.Style = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))

	return &App{
		client:     client,
		locale:     locale,
		theme:      theme,
		formatters: formatter.All(),
		history:    hist,
		screen:     screenSearch,
		searchInput: ti,
		spinner:    s,
		perPage:    20,
		historyIdx: -1,
		snippetCache: make(map[string]string),
	}
}

func (a *App) SetSearchValue(v string) {
	a.searchInput.SetValue(v)
}

func (a *App) Init() tea.Cmd {
	if a.searchInput.Value() != "" {
		return tea.Batch(textinput.Blink, a.spinner.Tick, a.doSearch())
	}
	return tea.Batch(textinput.Blink, a.spinner.Tick)
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.searchInput.Width = a.width - 10
		if a.searchInput.Width > 60 {
			a.searchInput.Width = 60
		}
		return a, nil
	case tea.KeyMsg:
		k := msg.String()
		if k == "ctrl+c" {
			return a, tea.Quit
		}
		// Global shortcut to jump to search
		if !a.searchInput.Focused() && (k == "/" || k == "s") {
			a.screen = screenSearch
			a.searchInput.Focus()
			return a, nil
		}
	case spinner.TickMsg:
		a.spinner, cmd = a.spinner.Update(msg)
		return a, cmd
	}

	switch a.screen {
	case screenSearch:
		return a.updateSearch(msg)
	case screenVersions:
		return a.updateVersions(msg)
	case screenSnippets:
		return a.updateSnippets(msg)
	}

	return a, nil
}

func (a *App) View() string {
	switch a.screen {
	case screenSearch:
		return a.viewSearch()
	case screenVersions:
		return a.viewVersions()
	case screenSnippets:
		return a.viewSnippets()
	}
	return ""
}
