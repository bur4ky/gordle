package main

import (
	"fmt"
	"strings"
	"unicode"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type phase int

const (
	phasePlaying phase = iota
	phaseWon
	phaseLost
)

var (
	styleCorrect = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#6AAA64")).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#6AAA64")).
			Padding(0, 1).
			Margin(0, 1, 0, 0)

	stylePresent = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#C9B458")).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#C9B458")).
			Padding(0, 1).
			Margin(0, 1, 0, 0)

	styleAbsent = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#787C7E")).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#787C7E")).
			Padding(0, 1).
			Margin(0, 1, 0, 0)

	styleEmpty = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F8F8F8")).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#565758")).
			Padding(0, 1).
			Margin(0, 1, 0, 0)

	styleActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F8F8F8")).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#F8F8F8")).
			Padding(0, 1).
			Margin(0, 1, 0, 0)

	styleStatus = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A0A0A0")).
			MarginTop(1)

	styleError = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D9534F")).
			MarginTop(1)

	styleStats = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8F8F8")).
			MarginTop(1)

	styleSourceOfficial = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6AAA64"))

	styleSourceUnofficial = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#A0A0A0"))
)

type Model struct {
	board    Board
	answer   string
	official bool

	session *Session
	stats   Stats

	phase      phase
	invalidMsg string
	quitting   bool
}

func NewModel(session *Session, stats Stats) Model {
	m := Model{
		session:  session,
		answer:   session.Word,
		official: session.Official,
		stats:    stats,
	}

	for _, g := range session.Guesses {
		for i, r := range []rune(g) {
			m.board.Rows[m.board.CurrentRow][i].Letter = r
		}

		m.board.CurrentCol = wordLength
		m.board.SubmitGuess(m.answer)
	}

	switch {
	case session.Won:
		m.phase = phaseWon
	case session.Lost:
		m.phase = phaseLost
	default:
		m.phase = phasePlaying
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "ctrl+n":
		return m.newRandomGame(), nil
	}

	switch m.phase {
	case phasePlaying:
		return m.handlePlayingKey(msg)
	default:
		if msg.String() == "esc" {
			m.quitting = true
			return m, tea.Quit
		}
		return m, nil
	}
}

func (m Model) newRandomGame() Model {
	m.board = Board{}
	m.answer = randomAnswer()
	m.official = false
	m.phase = phasePlaying
	m.invalidMsg = ""

	m.session = &Session{
		Date:     m.session.Date,
		Word:     m.answer,
		Official: false,
	}
	_ = SaveSession(m.session)

	return m
}

func (m Model) handlePlayingKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	m.invalidMsg = ""

	switch msg.String() {
	case "esc":
		m.quitting = true
		return m, tea.Quit

	case "backspace":
		m.board.RemoveLetter()
		return m, nil

	case "enter":
		if !m.board.RowFull() {
			return m, nil
		}

		guess := m.board.CurrentGuess()
		if _, ok := wordsSet[guess]; !ok {
			m.invalidMsg = "Invalid word"
			return m, nil
		}

		m.board.SubmitGuess(m.answer)
		m.session.Guesses = append(m.session.Guesses, guess)

		switch {
		case m.board.Won():
			m.phase = phaseWon
			m.session.Won = true
			m.stats.RecordResult(true, m.board.CurrentRow)
		case m.board.OutOfGuesses():
			m.phase = phaseLost
			m.session.Lost = true
			m.stats.RecordResult(false, 0)
		}

		_ = SaveSession(m.session)
		if m.phase == phaseWon || m.phase == phaseLost {
			_ = SaveStats(m.stats)
		}

		return m, nil
	}

	if r := singleUpperLetter(msg); r != 0 {
		m.board.AddLetter(r)
	}

	return m, nil
}

func singleUpperLetter(msg tea.KeyPressMsg) rune {
	runes := []rune(msg.Text)
	if len(runes) != 1 {
		return 0
	}

	r := unicode.ToUpper(runes[0])
	if r < 'A' || r > 'Z' {
		return 0
	}

	return r
}

func (m Model) View() tea.View {
	var b strings.Builder
	b.WriteString(m.renderSourceLabel())
	b.WriteString("\n")
	b.WriteString(m.renderBoard())
	b.WriteString("\n")

	switch m.phase {
	case phaseWon:
		b.WriteString(styleStatus.Render(fmt.Sprintf("You solved it in %d/%d! Press esc to quit, ctrl+n for a new word.", m.board.CurrentRow, maxGuesses)))
		b.WriteString("\n")
		b.WriteString(m.renderStats())
	case phaseLost:
		b.WriteString(styleStatus.Render(fmt.Sprintf("The word was %s. Press esc to quit, ctrl+n for a new word.", m.answer)))
		b.WriteString("\n")
		b.WriteString(m.renderStats())
	default:
		if m.invalidMsg != "" {
			b.WriteString(styleError.Render(m.invalidMsg))
		} else {
			b.WriteString(styleStatus.Render(fmt.Sprintf("Guess %d/%d  (ctrl+n: new word)", m.board.CurrentRow+1, maxGuesses)))
		}
	}

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}

func (m Model) renderSourceLabel() string {
	if m.official {
		return styleSourceOfficial.Render("Today's official Wordle")
	}
	return styleSourceUnofficial.Render("Today's unofficial Wordle")
}

func (m Model) renderBoard() string {
	var rows []string
	for r := range maxGuesses {
		var tiles []string
		for c := range wordLength {
			tile := m.board.Rows[r][c]
			letter := " "
			if tile.Letter != 0 {
				letter = string(tile.Letter)
			}

			var style lipgloss.Style
			switch {
			case tile.State == TileCorrect:
				style = styleCorrect
			case tile.State == TilePresent:
				style = stylePresent
			case tile.State == TileAbsent:
				style = styleAbsent
			case r == m.board.CurrentRow && m.phase == phasePlaying:
				style = styleActive
			default:
				style = styleEmpty
			}

			tiles = append(tiles, style.Render(letter))
		}

		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, tiles...))
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m Model) renderStats() string {
	s := m.stats
	return styleStats.Render(fmt.Sprintf(
		"Played: %d  Won: %d  Win%%: %.0f  Streak: %d  Best Streak: %d",
		s.Played, s.Won, s.WinPercentage(), s.CurrentStreak, s.BestStreak,
	))
}
