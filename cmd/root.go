package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/maher/mvns/internal/api"
	"github.com/maher/mvns/internal/config"
	formatterPkg "github.com/maher/mvns/internal/formatter"
	"github.com/maher/mvns/internal/history"
	"github.com/maher/mvns/internal/i18n"
	"github.com/maher/mvns/internal/ui"
	"github.com/maher/mvns/locales"
)

var (
	flagLang       string
	flagTheme      string
	flagQuery      string
	flagFormat     string
	flagClearCache bool
	Version        = "dev"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "mvns",
		Short:   "Maven Central Search - find and copy dependency snippets",
		Version: Version,
		RunE:    run,
	}

	cmd.Flags().StringVar(&flagLang, "lang", "", "language (en, de)")
	cmd.Flags().StringVar(&flagTheme, "theme", "", "theme (dark, light)")
	cmd.Flags().StringVar(&flagQuery, "query", "", "non-interactive search query")
	cmd.Flags().StringVar(&flagFormat, "format", "", "output format for non-interactive mode (maven, gradle, gradle-kts)")
	cmd.Flags().BoolVar(&flagClearCache, "clear-cache", false, "clear the local results cache")

	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	var (
		cfg   *config.Config
		hist  *history.History
		cache *api.Cache
		wg    sync.WaitGroup
	)

	cachePath := filepath.Join(config.ConfigDir(), "cache.json")
	if flagClearCache {
		os.Remove(cachePath)
	}

	wg.Add(3)
	go func() {
		defer wg.Done()
		var err error
		cfg, err = config.Load(config.ConfigPath())
		if err != nil {
			def := config.Default()
			cfg = &def
		}
	}()

	go func() {
		defer wg.Done()
		histPath := filepath.Join(config.ConfigDir(), "history.json")
		var err error
		hist, err = history.New(histPath)
		if err != nil {
			hist, _ = history.New(filepath.Join(os.TempDir(), "mvns-history.json"))
		}
	}()

	go func() {
		defer wg.Done()
		cache = api.NewCache(cachePath)
	}()

	wg.Wait()

	lang := cfg.Lang
	if flagLang != "" {
		lang = flagLang
	}
	themeName := cfg.Theme
	if flagTheme != "" {
		themeName = flagTheme
	}

	locale, err := i18n.NewFromFS(locales.FS, ".", lang)
	if err != nil {
		locale, _ = i18n.NewFromFS(locales.FS, ".", "en")
	}

	client := api.NewClient(api.WithCache(cache))

	// If query is provided with a format, it's strictly non-interactive
	if flagQuery != "" && flagFormat != "" {
		return runNonInteractive(client, locale, flagQuery, flagFormat)
	}

	theme := ui.NewTheme(themeName)

	app := ui.NewApp(client, locale, theme, hist)

	// If query is provided without format, pre-fill and trigger search in TUI
	var p *tea.Program
	if flagQuery != "" {
		app.SetSearchValue(flagQuery)
		p = tea.NewProgram(app, tea.WithAltScreen())
	} else {
		p = tea.NewProgram(app, tea.WithAltScreen())
	}

	_, err = p.Run()
	return err
}

func runNonInteractive(client *api.Client, locale *i18n.Locale, query, format string) error {
	resp, err := client.SearchMultimodal(query, 10, 0, false)
	if err != nil {
		return fmt.Errorf("%s: %w", locale.T("error.network"), err)
	}

	if len(resp.Response.Docs) == 0 {
		fmt.Println(locale.T("error.noresults"))
		return nil
	}

	// Deduplicate and sort
	docs := sortMultimodalResults(resp.Response.Docs, query)

	if format != "" {
		doc := docs[0]
		
		version := doc.LatestVersion
		if version == "" {
			version = doc.Version
		}

		// If the latest version is a pre-release, try to find the latest stable one
		if doc.IsPreRelease() {
			vResp, err := client.Versions(doc.GroupID, doc.ArtifactID, 100, false)
			if err == nil && len(vResp.Response.Docs) > 0 {
				for _, vDoc := range vResp.Response.Docs {
					if !vDoc.IsPreRelease() {
						version = vDoc.Version
						break
					}
				}
			}
		}

		dep := formatterPkg.Dependency{
			GroupID:    doc.GroupID,
			ArtifactID: doc.ArtifactID,
			Version:    version,
			Scope:      doc.DetectScope(),
		}

		var f formatterPkg.Formatter
		switch format {
		case "maven":
			f = &formatterPkg.Maven{}
		case "gradle":
			f = &formatterPkg.GradleGroovy{}
		case "gradle-kts":
			f = &formatterPkg.GradleKotlin{}
		default:
			return fmt.Errorf("unknown format: %s (use maven, gradle, gradle-kts)", format)
		}

		fmt.Println(f.Format(dep))
		return nil
	}

	for i, doc := range docs {
		if i >= 10 {
			break
		}
		v := doc.LatestVersion
		if v == "" {
			v = doc.Version
		}
		fmt.Printf("%-50s v%-12s %s  %s\n",
			doc.GroupID+":"+doc.ArtifactID,
			v,
			doc.Time().Format("2006-01-02"),
			doc.Packaging,
		)
	}
	return nil
}

func sortMultimodalResults(results []api.Doc, query string) []api.Doc {
	// Deduplicate by ID
	seen := make(map[string]bool)
	unique := make([]api.Doc, 0, len(results))
	for _, doc := range results {
		if !seen[doc.ID] {
			seen[doc.ID] = true
			unique = append(unique, doc)
		}
	}
	
	// Sort
	lowerQuery := strings.ToLower(query)
	sort.Slice(unique, func(i, j int) bool {
		iExact := unique[i].ArtifactID == query
		jExact := unique[j].ArtifactID == query
		if iExact != jExact {
			return iExact
		}

		iNorm := strings.ReplaceAll(strings.ToLower(unique[i].GroupID), ".", "-")
		jNorm := strings.ReplaceAll(strings.ToLower(unique[j].GroupID), ".", "-")
		iGroup := strings.Contains(iNorm, lowerQuery)
		jGroup := strings.Contains(jNorm, lowerQuery)
		if iGroup != jGroup {
			return iGroup
		}

		if unique[i].VersionCount != unique[j].VersionCount {
			return unique[i].VersionCount > unique[j].VersionCount
		}
		return unique[i].Timestamp > unique[j].Timestamp
	})
	
	return unique
}
