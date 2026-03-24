package ui

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"

	"sudoku_game/sudoku"
)

// Session preference keys.
const (
	sesKeyPuzzle     = "session.puzzle"     // 81 comma-separated ints (locked givens)
	sesKeyPlayer     = "session.player"     // 81 comma-separated ints (player moves)
	sesKeyDifficulty = "session.difficulty" // int
	sesKeyElapsed    = "session.elapsed"    // seconds (int)
	sesKeyHints      = "session.hints"      // hints remaining (int)
	sesKeyMistakes   = "session.mistakes"   // int
	sesKeyUndo       = "session.undo"       // semicolon-separated "r,c,old,new" quads
	sesKeyActive     = "session.active"     // "1" when a valid session exists
)

// sessionData holds everything needed to restore a game in progress.
type sessionData struct {
	puzzle     [sudoku.BoardSize][sudoku.BoardSize]int
	player     [sudoku.BoardSize][sudoku.BoardSize]int
	difficulty sudoku.Difficulty
	elapsed    int // seconds
	hints      int
	mistakes   int
	undoStack  [][4]int
}

// SaveSession persists the current board, timer, difficulty, hints, mistakes,
// and undo stack to fyne.Preferences so the game can be resumed after the app
// is closed.
func SaveSession(prefs fyne.Preferences, board *sudoku.Board, difficulty sudoku.Difficulty, elapsedSecs int, hintsLeft int, mistakes int) {
	puzzle := board.GetLockedBoard()
	player := board.GetPlayerBoard()

	prefs.SetString(sesKeyPuzzle, encodeMatrix(puzzle))
	prefs.SetString(sesKeyPlayer, encodeMatrix(player))
	prefs.SetInt(sesKeyDifficulty, int(difficulty))
	prefs.SetInt(sesKeyElapsed, elapsedSecs)
	prefs.SetInt(sesKeyHints, hintsLeft)
	prefs.SetInt(sesKeyMistakes, mistakes)
	prefs.SetString(sesKeyUndo, encodeUndoStack(board.GetUndoStack()))
	prefs.SetString(sesKeyActive, "1")
}

// ClearSession removes any saved session so the next launch starts fresh.
func ClearSession(prefs fyne.Preferences) {
	prefs.SetString(sesKeyActive, "")
}

// LoadSession returns the saved session and true, or (zero, false) if there is
// no valid saved session.
func LoadSession(prefs fyne.Preferences) (sessionData, bool) {
	if prefs.String(sesKeyActive) != "1" {
		return sessionData{}, false
	}

	puzzle, ok1 := decodeMatrix(prefs.String(sesKeyPuzzle))
	player, ok2 := decodeMatrix(prefs.String(sesKeyPlayer))
	if !ok1 || !ok2 {
		return sessionData{}, false
	}

	return sessionData{
		puzzle:     puzzle,
		player:     player,
		difficulty: sudoku.Difficulty(prefs.IntWithFallback(sesKeyDifficulty, int(sudoku.Easy))),
		elapsed:    prefs.IntWithFallback(sesKeyElapsed, 0),
		hints:      prefs.IntWithFallback(sesKeyHints, 3),
		mistakes:   prefs.IntWithFallback(sesKeyMistakes, 0),
		undoStack:  decodeUndoStack(prefs.String(sesKeyUndo)),
	}, true
}

// encodeMatrix flattens a 9×9 int matrix into a comma-separated string.
func encodeMatrix(m [sudoku.BoardSize][sudoku.BoardSize]int) string {
	parts := make([]string, 0, sudoku.BoardSize*sudoku.BoardSize)
	for r := 0; r < sudoku.BoardSize; r++ {
		for c := 0; c < sudoku.BoardSize; c++ {
			parts = append(parts, strconv.Itoa(m[r][c]))
		}
	}
	return strings.Join(parts, ",")
}

// decodeMatrix parses a comma-separated string back into a 9×9 int matrix.
func decodeMatrix(s string) ([sudoku.BoardSize][sudoku.BoardSize]int, bool) {
	parts := strings.Split(s, ",")
	if len(parts) != sudoku.BoardSize*sudoku.BoardSize {
		return [sudoku.BoardSize][sudoku.BoardSize]int{}, false
	}
	var m [sudoku.BoardSize][sudoku.BoardSize]int
	for i, p := range parts {
		v, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil || v < 0 || v > 9 {
			return [sudoku.BoardSize][sudoku.BoardSize]int{}, false
		}
		m[i/sudoku.BoardSize][i%sudoku.BoardSize] = v
	}
	return m, true
}

// encodeUndoStack serialises a [][4]int stack as "r,c,old,new;r,c,old,new;..."
func encodeUndoStack(stack [][4]int) string {
	if len(stack) == 0 {
		return ""
	}
	parts := make([]string, len(stack))
	for i, q := range stack {
		parts[i] = fmt.Sprintf("%d,%d,%d,%d", q[0], q[1], q[2], q[3])
	}
	return strings.Join(parts, ";")
}

// decodeUndoStack parses the output of encodeUndoStack back into [][4]int.
// Returns nil (empty stack) on any parse error.
func decodeUndoStack(s string) [][4]int {
	if s == "" {
		return nil
	}
	quads := strings.Split(s, ";")
	out := make([][4]int, 0, len(quads))
	for _, q := range quads {
		nums := strings.Split(q, ",")
		if len(nums) != 4 {
			return nil
		}
		var entry [4]int
		for j, n := range nums {
			v, err := strconv.Atoi(strings.TrimSpace(n))
			if err != nil {
				return nil
			}
			entry[j] = v
		}
		out = append(out, entry)
	}
	return out
}
