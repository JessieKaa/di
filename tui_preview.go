package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type previewLoadedMsg struct {
	sock    string
	content string
}

func fetchHistoryPreview(sock string) tea.Cmd {
	return func() tea.Msg {
		conn, err := dialSession(sock)
		if err != nil {
			return previewLoadedMsg{sock, "unreachable"}
		}
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(2 * time.Second))
		if err := writeFrame(conn, frameHistoryReq, nil); err != nil {
			return previewLoadedMsg{sock, "send error"}
		}
		typ, payload, err := readFrame(conn)
		if err == nil && typ == frameHistoryResp {
			return previewLoadedMsg{sock, string(payload)}
		}
		return previewLoadedMsg{sock, ""}
	}
}

func renderPreview(meta sessionMeta, content string, width, height int) string {
	if width < 4 || height < 4 {
		return ""
	}
	innerW := width - 2
	innerH := height - 2

	name := meta.Name
	if name == "" {
		name = "?"
	}
	cmd := strings.Join(meta.Command, " ")
	if cmd == "" {
		cmd = name
	}
	age := formatAge(meta.StartedAt)

	header := stylePreviewTitle.Render(padOrTruncate(name, innerW))
	pwdLine := styleDim.Render(padOrTruncate(meta.PWD, innerW))
	cmdLine := styleDim.Render(padOrTruncate(cmd+"  "+age, innerW))
	sep := styleSeparator.Render(strings.Repeat("─", innerW))

	availableH := innerH - 4
	if availableH < 1 {
		availableH = 1
	}

	lines := getLastVisibleLines(content, availableH, innerW)

	var b strings.Builder
	b.WriteString(header)
	b.WriteByte('\n')
	b.WriteString(pwdLine)
	b.WriteByte('\n')
	b.WriteString(cmdLine)
	b.WriteByte('\n')
	b.WriteString(sep)
	if len(lines) > 0 {
		b.WriteByte('\n')
		b.WriteString(strings.Join(lines, "\n"))
	}
	return b.String()
}

// --- minimal virtual terminal ---

type vtScreen struct {
	cells      [][]rune
	curX, curY int
	w, h       int
}

func newVTScreen(w, h int) *vtScreen {
	s := &vtScreen{w: w, h: h}
	s.cells = make([][]rune, h)
	for i := range s.cells {
		s.cells[i] = make([]rune, w)
	}
	return s
}

func (s *vtScreen) process(raw string) {
	runes := []rune(raw)
	i := 0
	n := len(runes)
	for i < n {
		ch := runes[i]
		if ch == 0x1b && i+1 < n {
			i = s.parseEscape(runes, i, n)
		} else if ch == '\n' {
			s.lineFeed()
			i++
		} else if ch == '\r' {
			s.curX = 0
			i++
		} else if ch >= 0x20 && ch != 0x7f {
			s.putChar(ch)
			i++
		} else {
			i++
		}
	}
}

func (s *vtScreen) putChar(ch rune) {
	if s.curY >= 0 && s.curY < s.h && s.curX >= 0 && s.curX < s.w {
		s.cells[s.curY][s.curX] = ch
	}
	s.curX++
	if s.curX >= s.w {
		s.curX = 0
		s.lineFeed()
	}
}

func (s *vtScreen) lineFeed() {
	s.curY++
	if s.curY >= s.h {
		s.scrollUp()
		s.curY = s.h - 1
	}
}

func (s *vtScreen) scrollUp() {
	for r := 0; r < s.h-1; r++ {
		copy(s.cells[r], s.cells[r+1])
	}
	for c := range s.cells[s.h-1] {
		s.cells[s.h-1][c] = 0
	}
}

func (s *vtScreen) parseEscape(r []rune, pos, n int) int {
	if pos+1 >= n {
		return n
	}
	switch r[pos+1] {
	case '[':
		return s.parseCSI(r, pos+2, n)
	case ']':
		return s.skipOSC(r, pos+2, n)
	default:
		return pos + 2
	}
}

func (s *vtScreen) parseCSI(r []rune, pos, n int) int {
	i := pos
	if i < n && r[i] == '?' {
		i++
	}
	var args []int
	num := -1
	for i < n {
		ch := r[i]
		if ch >= '0' && ch <= '9' {
			d := int(ch - '0')
			if num < 0 {
				num = d
			} else {
				num = num*10 + d
			}
			i++
		} else if ch == ';' {
			if num >= 0 {
				args = append(args, num)
			} else {
				args = append(args, 0)
			}
			num = -1
			i++
		} else {
			break
		}
	}
	if num >= 0 {
		args = append(args, num)
	}
	if i >= n {
		return n
	}
	switch r[i] {
	case 'A':
		s.moveCursor(0, -s.arg(args, 0, 1))
	case 'B':
		s.moveCursor(0, s.arg(args, 0, 1))
	case 'C':
		s.moveCursor(s.arg(args, 0, 1), 0)
	case 'D':
		s.moveCursor(-s.arg(args, 0, 1), 0)
	case 'H', 'f':
		s.setCursor(s.arg(args, 1, 1)-1, s.arg(args, 0, 1)-1)
	case 'J':
		s.eraseDisplay(s.arg(args, 0, 0))
	case 'K':
		s.eraseLine(s.arg(args, 0, 0))
	case 'M':
		s.scrollUp()
	}
	return i + 1
}

func (s *vtScreen) skipOSC(r []rune, pos, n int) int {
	for i := pos; i < n; i++ {
		if r[i] == 0x07 {
			return i + 1
		}
		if r[i] == 0x1b && i+1 < n && r[i+1] == '\\' {
			return i + 2
		}
	}
	return n
}

func (s *vtScreen) arg(args []int, idx, def int) int {
	if idx < len(args) && args[idx] > 0 {
		return args[idx]
	}
	return def
}

func (s *vtScreen) moveCursor(dx, dy int) {
	s.curX = clamp(s.curX+dx, 0, s.w-1)
	s.curY = clamp(s.curY+dy, 0, s.h-1)
}

func (s *vtScreen) setCursor(x, y int) {
	s.curX = clamp(x, 0, s.w-1)
	s.curY = clamp(y, 0, s.h-1)
}

func (s *vtScreen) eraseDisplay(mode int) {
	switch mode {
	case 0:
		for c := s.curX; c < s.w; c++ {
			s.cells[s.curY][c] = 0
		}
		for r := s.curY + 1; r < s.h; r++ {
			for c := range s.cells[r] {
				s.cells[r][c] = 0
			}
		}
	case 1:
		for r := 0; r < s.curY; r++ {
			for c := range s.cells[r] {
				s.cells[r][c] = 0
			}
		}
		for c := 0; c <= s.curX && c < s.w; c++ {
			s.cells[s.curY][c] = 0
		}
	case 2:
		for r := range s.cells {
			for c := range s.cells[r] {
				s.cells[r][c] = 0
			}
		}
	}
}

func (s *vtScreen) eraseLine(mode int) {
	if s.curY < 0 || s.curY >= s.h {
		return
	}
	switch mode {
	case 0:
		for c := s.curX; c < s.w; c++ {
			s.cells[s.curY][c] = 0
		}
	case 1:
		for c := 0; c <= s.curX && c < s.w; c++ {
			s.cells[s.curY][c] = 0
		}
	case 2:
		for c := range s.cells[s.curY] {
			s.cells[s.curY][c] = 0
		}
	}
}

func (s *vtScreen) strings() []string {
	lines := make([]string, s.h)
	for r := 0; r < s.h; r++ {
		var sb strings.Builder
		for c := 0; c < s.w; c++ {
			if s.cells[r][c] != 0 {
				sb.WriteRune(s.cells[r][c])
			} else {
				sb.WriteByte(' ')
			}
		}
		lines[r] = sb.String()
	}
	return lines
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// --- preview rendering ---

func getLastVisibleLines(raw string, maxLines, width int) []string {
	// Only process the tail of history to avoid replaying megabytes of
	// stale screen redraws. ~80KB is roughly 10 screenfuls at 80x24.
	const tailMax = 80 * 1024
	if len(raw) > tailMax {
		raw = raw[len(raw)-tailMax:]
	}

	const (
		screenW = 300
		screenH = 100
	)
	scr := newVTScreen(screenW, screenH)
	scr.process(raw)

	allRows := scr.strings()
	lastContent := 0
	for r, row := range allRows {
		if strings.TrimRight(row, " \t") != "" {
			lastContent = r
		}
	}

	start := lastContent - maxLines + 1
	if start < 0 {
		start = 0
	}
	if start > lastContent+1 {
		start = lastContent + 1
	}

	var visible []string
	for r := start; r <= lastContent && r < screenH; r++ {
		trimmed := strings.TrimRight(allRows[r], " \t")
		if trimmed == "" {
			visible = append(visible, strings.Repeat(" ", width))
		} else {
			visible = append(visible, padOrTruncate(trimmed, width))
		}
	}
	for len(visible) < maxLines {
		visible = append(visible, strings.Repeat(" ", width))
	}
	if len(visible) > maxLines {
		visible = visible[len(visible)-maxLines:]
	}
	return visible
}

func formatAge(startedAt string) string {
	t, err := time.Parse(time.RFC3339, startedAt)
	if err != nil {
		return ""
	}
	d := time.Since(t)
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}
