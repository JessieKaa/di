package main

import (
	"os"
	"path/filepath"
	"strings"
)

func renderSessionList(sessions []sessionInfo, cursor int, width, height int, focused bool) string {
	if width < 4 || height < 4 {
		return ""
	}
	innerW := width - 2
	innerH := height - 2

	if len(sessions) == 0 {
		empty := styleNoSessions.Render("No active sessions")
		return drawBorder(empty, width, height, focused)
	}

	scrollOff := 0
	if cursor >= innerH {
		scrollOff = cursor - innerH + 1
	}

	var b strings.Builder
	for i := 0; i < innerH; i++ {
		if i > 0 {
			b.WriteByte('\n')
		}
		idx := i + scrollOff
		if idx >= len(sessions) {
			b.WriteString(strings.Repeat(" ", innerW))
			continue
		}
		s := sessions[idx]
		b.WriteString(renderSessionRow(s, idx == cursor, innerW))
	}

	return drawBorder(b.String(), width, height, focused)
}

func renderSessionRow(s sessionInfo, selected bool, width int) string {
	name := s.Meta.Name
	if name == "" {
		name = strings.TrimSuffix(filepath.Base(s.Sock), ".sock")
	}
	pwd := s.Meta.PWD
	if pwd == "" {
		pwd = "-"
	}
	home := homeDir()
	if home != "" && strings.HasPrefix(pwd, home) {
		pwd = "~" + strings.TrimPrefix(pwd, home)
	}
	cmd := strings.Join(s.Meta.Command, " ")
	if cmd == "" {
		cmd = name
	}
	age := formatAge(s.Meta.StartedAt)

	cursor := "  "
	if selected {
		cursor = styleCursor.Render("▶")
	}

	namePart := name
	pwdPart := styleDim.Render(pwd)
	cmdPart := styleDim.Render(cmd)
	agePart := styleDim.Render(age)

	row := cursor + " " + namePart + "  " + pwdPart + "  " + cmdPart + "  " + agePart

	if selected {
		row = styleSelected.Render(padOrTruncate(row, width))
	} else {
		row = padOrTruncate(row, width)
	}
	return row
}

func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}
