package main

import (
	"strings"
)

const (
	wordLength = 5
	maxGuesses = 6
)

type TileState int

const (
	TileEmpty TileState = iota
	TileCorrect
	TilePresent
	TileAbsent
)

type Tile struct {
	Letter rune
	State  TileState
}

type Board struct {
	Rows       [maxGuesses][wordLength]Tile
	CurrentRow int
	CurrentCol int
}

func (b *Board) CurrentGuess() string {
	var sb strings.Builder
	for i := 0; i < b.CurrentCol; i++ {
		sb.WriteRune(b.Rows[b.CurrentRow][i].Letter)
	}

	return sb.String()
}

func (b *Board) AddLetter(r rune) {
	if b.CurrentCol >= wordLength {
		return
	}

	b.Rows[b.CurrentRow][b.CurrentCol].Letter = r
	b.CurrentCol++
}

func (b *Board) RemoveLetter() {
	if b.CurrentCol == 0 {
		return
	}

	b.CurrentCol--
	b.Rows[b.CurrentRow][b.CurrentCol] = Tile{}
}

func (b *Board) RowFull() bool {
	return b.CurrentCol == wordLength
}

func (b *Board) SubmitGuess(answer string) {
	guess := b.CurrentGuess()
	states := HandleGuess(guess, answer)

	for i, s := range states {
		b.Rows[b.CurrentRow][i].State = s
	}

	b.CurrentRow++
	b.CurrentCol = 0
}

func HandleGuess(guess, answer string) [wordLength]TileState {
	g := []rune(strings.ToUpper(guess))
	a := []rune(strings.ToUpper(answer))

	var states [wordLength]TileState
	remaining := make(map[rune]int, wordLength)

	for i := range wordLength {
		if i < len(g) && i < len(a) && g[i] == a[i] {
			states[i] = TileCorrect
		} else if i < len(a) {
			remaining[a[i]]++
		}
	}

	for i := range wordLength {
		if states[i] == TileCorrect {
			continue
		}

		if i >= len(g) {
			states[i] = TileAbsent
			continue
		}

		letter := g[i]
		if remaining[letter] > 0 {
			states[i] = TilePresent
			remaining[letter]--
		} else {
			states[i] = TileAbsent
		}
	}

	return states
}

func (b *Board) Won() bool {
	if b.CurrentRow == 0 {
		return false
	}

	last := b.Rows[b.CurrentRow-1]
	for _, t := range last {
		if t.State != TileCorrect {
			return false
		}
	}

	return true
}

func (b *Board) OutOfGuesses() bool {
	return b.CurrentRow >= maxGuesses
}
