package main

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	configDirName = "gordle"
	sessionFile   = "session.json"
	statsFile     = "stats.json"
)

type Session struct {
	Date     string   `json:"date"`
	Word     string   `json:"word"`
	Official bool     `json:"official"`
	Guesses  []string `json:"guesses"`
	Won      bool     `json:"won"`
	Lost     bool     `json:"lost"`
}

type Stats struct {
	Played            int             `json:"played"`
	Won               int             `json:"won"`
	CurrentStreak     int             `json:"current_streak"`
	BestStreak        int             `json:"best_streak"`
	GuessDistribution [maxGuesses]int `json:"guess_distribution"`
}

func (s Stats) WinPercentage() float64 {
	if s.Played == 0 {
		return 0
	}

	return float64(s.Won) / float64(s.Played) * 100
}

func (s *Stats) RecordResult(won bool, guessCount int) {
	s.Played++

	if won {
		s.Won++
		s.CurrentStreak++
		if s.CurrentStreak > s.BestStreak {
			s.BestStreak = s.CurrentStreak
		}
		if guessCount >= 1 && guessCount <= len(s.GuessDistribution) {
			s.GuessDistribution[guessCount-1]++
		}
	} else {
		s.CurrentStreak = 0
	}
}

func configDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolving user config dir: %w", err)
	}

	return filepath.Join(base, configDirName), nil
}

func filePath(filename string) (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, filename), nil
}

func loadJSON[T any](path string) (T, error) {
	var value T

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return value, nil
	}
	if err != nil {
		return value, fmt.Errorf("reading %s: %w", filepath.Base(path), err)
	}

	if err := json.Unmarshal(data, &value); err != nil {
		return value, fmt.Errorf("parsing %s: %w", filepath.Base(path), err)
	}

	return value, nil
}

func saveJSON[T any](path string, value T) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encoding %s: %w", filepath.Base(path), err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", filepath.Base(path), err)
	}

	return nil
}

func LoadSession() (*Session, error) {
	path, err := filePath(sessionFile)
	if err != nil {
		return nil, err
	}

	return loadJSON[*Session](path)
}

func SaveSession(s *Session) error {
	path, err := filePath(sessionFile)
	if err != nil {
		return err
	}

	return saveJSON(path, s)
}

func LoadStats() (Stats, error) {
	path, err := filePath(statsFile)
	if err != nil {
		return Stats{}, err
	}

	return loadJSON[Stats](path)
}

func SaveStats(s Stats) error {
	path, err := filePath(statsFile)
	if err != nil {
		return err
	}

	return saveJSON(path, s)
}
