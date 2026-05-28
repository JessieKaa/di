package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	ansi "github.com/charmbracelet/x/ansi"
)

func padOrTruncate(s string, width int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\t", "    ")
	sw := ansi.StringWidth(s)
	if sw > width {
		return ansi.Truncate(s, width, "")
	}
	if sw < width {
		return s + strings.Repeat(" ", width-sw)
	}
	return s
}

func drawBorder(content string, width, height int, active bool) string {
	borderColor := colorBorder
	if active {
		borderColor = colorPrimary
	}
	innerW := width - 2
	innerH := height - 2
	if innerW < 1 || innerH < 1 {
		return ""
	}
	return lipgloss.
		NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Width(innerW).
		Height(innerH).
		Render(content)
}

func joinHorizontalFixed(left, right string, leftW, totalW int) string {
	leftLines := strings.Split(left, "\n")
	rightLines := strings.Split(right, "\n")
	maxH := len(leftLines)
	if len(rightLines) > maxH {
		maxH = len(rightLines)
	}
	rightW := totalW - leftW
	if rightW < 1 {
		rightW = 1
	}
	var b strings.Builder
	for i := 0; i < maxH; i++ {
		if i > 0 {
			b.WriteByte('\n')
		}
		ll := ""
		if i < len(leftLines) {
			ll = leftLines[i]
		}
		rl := ""
		if i < len(rightLines) {
			rl = rightLines[i]
		}
		llW := ansi.StringWidth(ll)
		pad := leftW - llW
		if pad < 0 {
			pad = 0
		}
		b.WriteString(ll)
		b.WriteString(strings.Repeat(" ", pad))
		b.WriteString(rl)
	}
	return b.String()
}
