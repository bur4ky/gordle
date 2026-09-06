package main

import (
	"context"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/log/v2"
)

func main() {
	logger := log.New(os.Stderr)

	if err := run(logger); err != nil {
		logger.Error("gordle exited with an error", "err", err)
		os.Exit(1)
	}
}

func run(logger *log.Logger) error {
	today := time.Now().Format("2006-01-02")

	session, err := LoadSession()
	if err != nil {
		logger.Error("failed to load existing session", "err", err)
	}

	if session == nil || session.Date != today {
		word, source, fetchErr := GetDailyWord(context.Background(), today)
		if fetchErr != nil {
			logger.Warn("failed to fetch daily word", "err", fetchErr)
		}

		session = &Session{
			Date:     today,
			Word:     word,
			Official: source == SourceNYT,
		}
		if err := SaveSession(session); err != nil {
			logger.Error("failed to save session", "err", err)
		}
	}

	stats, err := LoadStats()
	if err != nil {
		logger.Error("failed to load stats", "err", err)
	}

	p := tea.NewProgram(NewModel(session, stats))
	_, err = p.Run()
	return err
}
