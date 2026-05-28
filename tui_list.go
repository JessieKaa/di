package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var reNameSuffix = regexp.MustCompile(`-\d+-\d+$`)

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
	name = reNameSuffix.ReplaceAllString(name, "")

	pwd := s.Meta.PWD
	if pwd == "" {
		pwd = "-"
	} else {
		pwd = filepath.Base(pwd)
		if pwd == "" || pwd == "/" {
			pwd = "/"
		}
	}

	age := formatAge(s.Meta.StartedAt)

	cur := "  "
	if selected {
		cur = styleCursor.Render("▶")
	}

	namePart := name
	pwdPart := styleDim.Render(pwd)
	agePart := styleDim.Render(age)

	row := cur + " " + namePart + "  " + pwdPart + "  " + agePart

	if selected {
		row = styleSelected.Render(padOrTruncate(row, width))
	} else {
		row = padOrTruncate(row, width)
	}
	return row
}

func shortenPath(pwd string) string {
	home := homeDir()
	if home != "" && strings.HasPrefix(pwd, home) {
		pwd = "~" + strings.TrimPrefix(pwd, home)
	}
	parts := strings.Split(strings.TrimPrefix(pwd, "~"), "/")
	if len(parts) <= 2 {
		return pwd
	}
	// Show ".../parent/current"
	return "~/" + "..." + "/" + parts[len(parts)-2] + "/" + parts[len(parts)-1]
}

func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}
