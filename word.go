package main

import (
	"context"
	_ "embed"
	"encoding/json/v2"
	"fmt"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"
)

//go:embed words.txt
var wordsContent string
var wordsSet = loadWordSet(wordsContent)

//go:embed answers.txt
var answersContent string
var answers = loadWordList(answersContent)

const nytEndpoint = "https://www.nytimes.com/svc/wordle/v2/%s.json"

type nytResponse struct {
	ID              int    `json:"id"`
	Solution        string `json:"solution"`
	PrintDate       string `json:"print_date"`
	DaysSinceLaunch int    `json:"days_since_launch"`
	Editor          string `json:"editor"`
}

type WordSource int

const (
	SourceNYT WordSource = iota
	SourceOffline
)

func (s WordSource) String() string {
	if s == SourceNYT {
		return "nyt"
	}
	return "offline"
}

func loadWordList(contents string) []string {
	lines := strings.Fields(contents)
	words := make([]string, 0, len(lines))
	for _, line := range lines {
		word := strings.ToUpper(strings.TrimSpace(line))
		words = append(words, word)
	}

	return words
}

func loadWordSet(contents string) map[string]struct{} {
	words := strings.Fields(contents)
	set := make(map[string]struct{}, len(words))
	for _, word := range words {
		word = strings.ToUpper(word)
		set[word] = struct{}{}
	}

	return set
}

func randomAnswer() string {
	return answers[rand.IntN(len(answers))]
}

func fetchNYTWord(ctx context.Context, date string) (string, error) {
	url := fmt.Sprintf(nytEndpoint, date)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("requesting: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	var result nytResponse
	if err := json.UnmarshalRead(resp.Body, &result); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	return strings.ToUpper(result.Solution), nil
}

func GetDailyWord(ctx context.Context, date string) (string, WordSource, error) {
	fetchCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	word, err := fetchNYTWord(fetchCtx, date)
	if err == nil {
		return word, SourceNYT, nil
	}

	return randomAnswer(), SourceOffline, err
}
