package main

import "github.com/charmbracelet/lipgloss"

var (
	colorPrimary   = "#7C3AED"
	colorAccent    = "#22D3EE"
	colorDimText   = "#6B7280"
	colorBorder    = "#4B5563"
	colorSelected  = "#1E1B4B"
	colorHighlight = "#A78BFA"
	colorTitle     = "#C4B5FD"
	colorError     = "#EF4444"
)

var (
	styleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorTitle))

	styleHelp = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorDimText))

	styleCursor = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorAccent)).
			Bold(true)

	styleSelected = lipgloss.NewStyle().
			Background(lipgloss.AdaptiveColor{Light: "#E0E7FF", Dark: colorSelected}).
			Foreground(lipgloss.AdaptiveColor{Light: "#312E81", Dark: colorHighlight})

	styleDim = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorDimText))

	styleFolder = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: colorHighlight})

	styleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colorPrimary))

	styleSeparator = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorBorder))

	stylePreviewTitle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(colorAccent))

	styleNoSessions = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorDimText)).
			Italic(true)
)
