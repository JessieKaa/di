package main

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	ansi "github.com/charmbracelet/x/ansi"
)

type tuiModel struct {
	sessions  []sessionInfo
	cursor    int
	preview   string
	width     int
	height    int
	attachSock string
	quitting  bool
}

type tickSessionsMsg struct{}
type tickPreviewMsg struct{}
type sessionsLoadedMsg struct {
	sessions []sessionInfo
}

func newTUIModel() tuiModel {
	return tuiModel{}
}

func (m tuiModel) Init() tea.Cmd {
	return tea.Batch(
		loadSessions(),
		tickSessions(),
		tickPreview(),
	)
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.cursor >= len(m.sessions) {
			m.cursor = len(m.sessions) - 1
		}
		if len(m.sessions) > 0 && m.cursor >= 0 {
			return m, fetchHistoryPreview(m.sessions[m.cursor].Sock)
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				m.preview = ""
				return m, fetchHistoryPreview(m.sessions[m.cursor].Sock)
			}
		case "down", "j":
			if m.cursor < len(m.sessions)-1 {
				m.cursor++
				m.preview = ""
				return m, fetchHistoryPreview(m.sessions[m.cursor].Sock)
			}
		case "g":
			m.cursor = 0
			m.preview = ""
			if len(m.sessions) > 0 {
				return m, fetchHistoryPreview(m.sessions[0].Sock)
			}
		case "G":
			if len(m.sessions) > 0 {
				m.cursor = len(m.sessions) - 1
				m.preview = ""
				return m, fetchHistoryPreview(m.sessions[m.cursor].Sock)
			}
		case "enter":
			if len(m.sessions) > 0 && m.cursor >= 0 && m.cursor < len(m.sessions) {
				m.attachSock = m.sessions[m.cursor].Sock
				return m, tea.Quit
			}
		}

	case tickSessionsMsg:
		return m, tea.Batch(loadSessions(), tickSessions())

	case sessionsLoadedMsg:
		m.sessions = msg.sessions
		if m.cursor >= len(m.sessions) {
			m.cursor = len(m.sessions) - 1
		}
		if m.cursor < 0 {
			m.cursor = 0
		}
		if len(m.sessions) > 0 && m.cursor >= 0 {
			return m, fetchHistoryPreview(m.sessions[m.cursor].Sock)
		}
		m.preview = ""
		return m, nil

	case tickPreviewMsg:
		if len(m.sessions) > 0 && m.cursor >= 0 && m.cursor < len(m.sessions) {
			return m, fetchHistoryPreview(m.sessions[m.cursor].Sock)
		}
		return m, tickPreview()

	case previewLoadedMsg:
		if m.cursor >= 0 && m.cursor < len(m.sessions) && m.sessions[m.cursor].Sock == msg.sock {
			m.preview = msg.content
		}
		return m, tickPreview()
	}

	return m, nil
}

func (m tuiModel) View() string {
	if m.quitting {
		return ""
	}
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	leftW := m.width * 40 / 100
	if leftW < 20 {
		leftW = 20
	}
	rightW := m.width - leftW
	if rightW < 20 {
		rightW = 20
		leftW = m.width - rightW
	}

	title := styleTitle.Render("di") + styleDim.Render(" sessions")
	help := styleHelp.Render("j/k ↑/↓ navigate  enter attach  q quit")
	panelH := m.height - 2
	if panelH < 4 {
		panelH = 4
	}

	leftPanel := renderSessionList(m.sessions, m.cursor, leftW, panelH, true)

	var rightContent string
	if len(m.sessions) > 0 && m.cursor >= 0 && m.cursor < len(m.sessions) {
		s := m.sessions[m.cursor]
		rightContent = renderPreview(s.Meta, m.preview, rightW, panelH)
	} else {
		rightContent = styleNoSessions.Render("Select a session to preview")
	}

	joined := joinHorizontalFixed(leftPanel, rightContent, leftW, m.width)

	var b strings.Builder
	b.WriteString(title)
	remaining := m.width - ansi.StringWidth("di sessions")
	if remaining > 0 {
		b.WriteString(strings.Repeat(" ", remaining))
	}
	b.WriteByte('\n')
	b.WriteString(joined)
	if m.height > 2 {
		b.WriteByte('\n')
		b.WriteString(help)
	}
	return b.String()
}

func loadSessions() tea.Cmd {
	return func() tea.Msg {
		sessions, err := allSessions()
		if err != nil {
			return sessionsLoadedMsg{sessions: nil}
		}
		return sessionsLoadedMsg{sessions: sessions}
	}
}

func tickSessions() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return tickSessionsMsg{}
	})
}

func tickPreview() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return tickPreviewMsg{}
	})
}
