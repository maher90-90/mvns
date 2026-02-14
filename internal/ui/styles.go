package ui

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Name        string
	Title       lipgloss.Style
	Subtitle    lipgloss.Style
	Selected    lipgloss.Style
	Normal      lipgloss.Style
	Dimmed      lipgloss.Style
	Help        lipgloss.Style
	Error       lipgloss.Style
	Success     lipgloss.Style
	TabActive   lipgloss.Style
	TabInactive lipgloss.Style
	Separator   lipgloss.Style
}

func NewTheme(name string) *Theme {
	switch name {
	case "light":
		return lightTheme()
	case "dark":
		return darkTheme()
	default:
		return darkTheme()
	}
}

func darkTheme() *Theme {
	return &Theme{
		Name:        "dark",
		Title:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")),
		Subtitle:    lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		Selected:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57")),
		Normal:      lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		Dimmed:      lipgloss.NewStyle().Foreground(lipgloss.Color("239")),
		Help:        lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
		Error:       lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
		Success:     lipgloss.NewStyle().Foreground(lipgloss.Color("82")),
		TabActive:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")).Underline(true),
		TabInactive: lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		Separator:   lipgloss.NewStyle().Foreground(lipgloss.Color("238")),
	}
}

func lightTheme() *Theme {
	return &Theme{
		Name:        "light",
		Title:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("55")),
		Subtitle:    lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		Selected:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255")).Background(lipgloss.Color("27")),
		Normal:      lipgloss.NewStyle().Foreground(lipgloss.Color("235")),
		Dimmed:      lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		Help:        lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		Error:       lipgloss.NewStyle().Foreground(lipgloss.Color("160")),
		Success:     lipgloss.NewStyle().Foreground(lipgloss.Color("28")),
		TabActive:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("55")).Underline(true),
		TabInactive: lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		Separator:   lipgloss.NewStyle().Foreground(lipgloss.Color("250")),
	}
}
