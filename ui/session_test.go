package ui

import (
	"testing"

	"fyne.io/fyne/v2/test"

	"sudoku_game/sudoku"
)

// ---------------------------------------------------------------------------
// SaveSession / LoadSession / ClearSession
// ---------------------------------------------------------------------------

func TestSession_RoundTrip_EmptyBoard(t *testing.T) {
	prefs := test.NewApp().Preferences()
	board := sudoku.NewBoard()

	SaveSession(prefs, board, sudoku.Easy, 0, 3, 0)

	sess, ok := LoadSession(prefs)
	if !ok {
		t.Fatal("LoadSession returned false after SaveSession")
	}
	if sess.difficulty != sudoku.Easy {
		t.Errorf("difficulty: want Easy, got %v", sess.difficulty)
	}
	if sess.elapsed != 0 {
		t.Errorf("elapsed: want 0, got %d", sess.elapsed)
	}
	if sess.hints != 3 {
		t.Errorf("hints: want 3, got %d", sess.hints)
	}
	if sess.mistakes != 0 {
		t.Errorf("mistakes: want 0, got %d", sess.mistakes)
	}
}

func TestSession_RoundTrip_WithPuzzleAndPlayer(t *testing.T) {
	prefs := test.NewApp().Preferences()

	board := sudoku.NewBoard()
	var m [sudoku.BoardSize][sudoku.BoardSize]int
	m[0][0] = 5
	if err := board.SetBoard(m); err != nil {
		t.Fatalf("SetBoard: %v", err)
	}
	_ = board.PlaceNumberForce(1, 1, 3)

	SaveSession(prefs, board, sudoku.Hard, 120, 1, 2)
	sess, ok := LoadSession(prefs)
	if !ok {
		t.Fatal("LoadSession returned false")
	}

	if sess.puzzle[0][0] != 5 {
		t.Errorf("puzzle[0][0]: want 5, got %d", sess.puzzle[0][0])
	}
	if sess.player[1][1] != 3 {
		t.Errorf("player[1][1]: want 3, got %d", sess.player[1][1])
	}
	if sess.difficulty != sudoku.Hard {
		t.Errorf("difficulty: want Hard, got %v", sess.difficulty)
	}
	if sess.elapsed != 120 {
		t.Errorf("elapsed: want 120, got %d", sess.elapsed)
	}
	if sess.hints != 1 {
		t.Errorf("hints: want 1, got %d", sess.hints)
	}
	if sess.mistakes != 2 {
		t.Errorf("mistakes: want 2, got %d", sess.mistakes)
	}
}

func TestSession_RoundTrip_Solution(t *testing.T) {
	prefs := test.NewApp().Preferences()

	board := sudoku.NewBoard()
	_ = board.SetBoard(sudoku.GenerateBoard(sudoku.Easy))
	wantSol := board.GetSolution()

	SaveSession(prefs, board, sudoku.Easy, 0, 3, 0)
	sess, ok := LoadSession(prefs)
	if !ok {
		t.Fatal("LoadSession returned false")
	}
	if sess.solution != wantSol {
		t.Error("solution did not survive SaveSession/LoadSession round-trip")
	}
}

func TestSession_LoadSolution_AbsentIsZero(t *testing.T) {
	// Simulate an old session that has no solution key stored.
	prefs := test.NewApp().Preferences()
	// Write a valid session but no solution key.
	board := sudoku.NewBoard()
	SaveSession(prefs, board, sudoku.Easy, 0, 3, 0)
	prefs.SetString(sesKeySolution, "") // blank out solution
	sess, ok := LoadSession(prefs)
	if !ok {
		t.Fatal("LoadSession should still succeed when solution key is absent")
	}
	var zero [sudoku.BoardSize][sudoku.BoardSize]int
	if sess.solution != zero {
		t.Error("absent solution key should decode to zero matrix")
	}
}

func TestSession_RoundTrip_UndoStack(t *testing.T) {
	prefs := test.NewApp().Preferences()

	board := sudoku.NewBoard()
	_ = board.PlaceNumberForceUndo(0, 0, 7)
	_ = board.PlaceNumberForceUndo(0, 1, 8)

	SaveSession(prefs, board, sudoku.Medium, 30, 2, 0)
	sess, ok := LoadSession(prefs)
	if !ok {
		t.Fatal("LoadSession returned false")
	}

	if len(sess.undoStack) != 2 {
		t.Fatalf("undo stack: want 2 entries, got %d", len(sess.undoStack))
	}
	if sess.undoStack[0] != [4]int{0, 0, 0, 7} {
		t.Errorf("undo[0]: want {0,0,0,7}, got %v", sess.undoStack[0])
	}
	if sess.undoStack[1] != [4]int{0, 1, 0, 8} {
		t.Errorf("undo[1]: want {0,1,0,8}, got %v", sess.undoStack[1])
	}
}

func TestSession_LoadReturnsFalseWhenNoSession(t *testing.T) {
	prefs := test.NewApp().Preferences()
	_, ok := LoadSession(prefs)
	if ok {
		t.Fatal("LoadSession should return false when no session is saved")
	}
}

func TestSession_ClearSession(t *testing.T) {
	prefs := test.NewApp().Preferences()
	board := sudoku.NewBoard()
	SaveSession(prefs, board, sudoku.Easy, 0, 3, 0)

	ClearSession(prefs)

	_, ok := LoadSession(prefs)
	if ok {
		t.Fatal("LoadSession should return false after ClearSession")
	}
}

func TestSession_RoundTrip_AllDifficulties(t *testing.T) {
	difficulties := []sudoku.Difficulty{
		sudoku.VeryEasy, sudoku.Easy, sudoku.Medium, sudoku.Hard, sudoku.VeryHard,
	}
	names := []string{"VeryEasy", "Easy", "Medium", "Hard", "VeryHard"}
	for i, d := range difficulties {
		d, name := d, names[i]
		t.Run(name, func(t *testing.T) {
			prefs := test.NewApp().Preferences()
			board := sudoku.NewBoard()
			SaveSession(prefs, board, d, 0, 3, 0)
			sess, ok := LoadSession(prefs)
			if !ok {
				t.Fatal("LoadSession returned false")
			}
			if sess.difficulty != d {
				t.Errorf("want %v, got %v", d, sess.difficulty)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// encodeMatrix / decodeMatrix
// ---------------------------------------------------------------------------

func TestEncodeDecodeMatrix_RoundTrip(t *testing.T) {
	var m [sudoku.BoardSize][sudoku.BoardSize]int
	for r := 0; r < sudoku.BoardSize; r++ {
		for c := 0; c < sudoku.BoardSize; c++ {
			m[r][c] = (r*sudoku.BoardSize+c)%9 + 1
		}
	}
	encoded := encodeMatrix(m)
	decoded, ok := decodeMatrix(encoded)
	if !ok {
		t.Fatal("decodeMatrix returned false")
	}
	if decoded != m {
		t.Error("decoded matrix does not match original")
	}
}

func TestDecodeMatrix_InvalidLength(t *testing.T) {
	_, ok := decodeMatrix("1,2,3")
	if ok {
		t.Fatal("should return false for wrong number of values")
	}
}

func TestDecodeMatrix_InvalidValue(t *testing.T) {
	// Build an 81-element string with one bad value.
	parts := make([]string, 81)
	for i := range parts {
		parts[i] = "1"
	}
	parts[40] = "x" // non-numeric
	s := ""
	for i, p := range parts {
		if i > 0 {
			s += ","
		}
		s += p
	}
	_, ok := decodeMatrix(s)
	if ok {
		t.Fatal("should return false for non-numeric value")
	}
}

// ---------------------------------------------------------------------------
// encodeUndoStack / decodeUndoStack
// ---------------------------------------------------------------------------

func TestEncodeDecodeUndoStack_Empty(t *testing.T) {
	encoded := encodeUndoStack(nil)
	if encoded != "" {
		t.Errorf("empty stack should encode to empty string, got %q", encoded)
	}
	decoded := decodeUndoStack("")
	if decoded != nil {
		t.Errorf("empty string should decode to nil, got %v", decoded)
	}
}

func TestEncodeDecodeUndoStack_RoundTrip(t *testing.T) {
	stack := [][4]int{
		{0, 0, 0, 5},
		{3, 4, 0, 9},
		{8, 8, 7, 0},
	}
	encoded := encodeUndoStack(stack)
	decoded := decodeUndoStack(encoded)
	if len(decoded) != len(stack) {
		t.Fatalf("length mismatch: want %d, got %d", len(stack), len(decoded))
	}
	for i := range stack {
		if decoded[i] != stack[i] {
			t.Errorf("[%d]: want %v, got %v", i, stack[i], decoded[i])
		}
	}
}

func TestDecodeUndoStack_CorruptData(t *testing.T) {
	result := decodeUndoStack("0,0,0,5;bad_data;3,4,0,9")
	if result != nil {
		t.Fatal("corrupt undo stack should decode to nil")
	}
}
