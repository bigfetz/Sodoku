package sudoku_test

import (
	"testing"

	"sudoku_game/sudoku"
)

// ---------------------------------------------------------------------------
// NewBoard
// ---------------------------------------------------------------------------

func TestNewBoard_EmptyBoard(t *testing.T) {
	b := sudoku.NewBoard()
	m := b.GetBoard()
	for r := 0; r < sudoku.BoardSize; r++ {
		for c := 0; c < sudoku.BoardSize; c++ {
			if m[r][c] != 0 {
				t.Errorf("expected empty board at [%d][%d], got %d", r, c, m[r][c])
			}
		}
	}
}

// ---------------------------------------------------------------------------
// SetBoard
// ---------------------------------------------------------------------------

func TestSetBoard_ValidMatrix(t *testing.T) {
	b := sudoku.NewBoard()

	var m [sudoku.BoardSize][sudoku.BoardSize]int
	m[8][8] = 9

	if err := b.SetBoard(m); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := b.GetBoard()
	if got[8][8] != 9 {
		t.Error("new value not applied by SetBoard")
	}
}

func TestSetBoard_ConflictingMatrix(t *testing.T) {
	b := sudoku.NewBoard()
	var m [sudoku.BoardSize][sudoku.BoardSize]int
	m[0][0] = 3
	m[0][5] = 3 // same row conflict

	if err := b.SetBoard(m); err != sudoku.ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
	// Board must remain unchanged.
	if b.GetBoard()[0][0] != 0 {
		t.Error("board was mutated despite conflict error")
	}
}

func TestSetBoard_InvalidValue(t *testing.T) {
	b := sudoku.NewBoard()
	var m [sudoku.BoardSize][sudoku.BoardSize]int
	m[0][0] = 11

	if err := b.SetBoard(m); err != sudoku.ErrInvalidValue {
		t.Fatalf("expected ErrInvalidValue, got %v", err)
	}
}

func TestSetBoard_LocksGivens(t *testing.T) {
	b := sudoku.NewBoard()
	var m [sudoku.BoardSize][sudoku.BoardSize]int
	m[0][0] = 7

	if err := b.SetBoard(m); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	locked := b.LockedCells()
	if !locked[0][0] {
		t.Error("cell [0][0] should be locked after SetBoard")
	}
	if locked[0][1] {
		t.Error("cell [0][1] should not be locked")
	}
}

// ---------------------------------------------------------------------------
// PlaceNumberForce
// ---------------------------------------------------------------------------

func TestPlaceNumberForce_ValidPlacement(t *testing.T) {
	b := sudoku.NewBoard()
	if err := b.PlaceNumberForce(0, 0, 5); err != nil {
		t.Fatalf("expected valid placement, got err=%v", err)
	}
	if b.GetBoard()[0][0] != 5 {
		t.Fatal("value not written to board")
	}
}

func TestPlaceNumberForce_OutOfBounds(t *testing.T) {
	b := sudoku.NewBoard()
	if err := b.PlaceNumberForce(9, 0, 5); err != sudoku.ErrOutOfBounds {
		t.Fatalf("expected ErrOutOfBounds, got %v", err)
	}
}

func TestPlaceNumberForce_InvalidValue(t *testing.T) {
	b := sudoku.NewBoard()
	if err := b.PlaceNumberForce(0, 0, 10); err != sudoku.ErrInvalidValue {
		t.Fatalf("expected ErrInvalidValue, got %v", err)
	}
}

func TestPlaceNumberForce_AllowsConflict(t *testing.T) {
	// PlaceNumberForce deliberately allows conflicts so the UI can show them in red.
	b := sudoku.NewBoard()
	_ = b.PlaceNumberForce(0, 0, 5)
	if err := b.PlaceNumberForce(0, 1, 5); err != nil {
		t.Fatalf("PlaceNumberForce should allow conflicting placement, got %v", err)
	}
	conflicts := b.Conflicts()
	if !conflicts[0][0] || !conflicts[0][1] {
		t.Error("expected both cells to be flagged as conflicts")
	}
}

func TestPlaceNumberForce_LockedCell(t *testing.T) {
	var m [sudoku.BoardSize][sudoku.BoardSize]int
	m[0][0] = 7
	b := sudoku.NewBoard()
	_ = b.SetBoard(m)

	if err := b.PlaceNumberForce(0, 0, 3); err != sudoku.ErrCellLocked {
		t.Fatalf("expected ErrCellLocked, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// PlaceNumberForceUndo
// ---------------------------------------------------------------------------

func TestPlaceNumberForceUndo_PushesStack(t *testing.T) {
	b := sudoku.NewBoard()
	if err := b.PlaceNumberForceUndo(0, 0, 5); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !b.CanUndo() {
		t.Fatal("undo stack should have one entry")
	}
}

// ---------------------------------------------------------------------------
// ClearCell
// ---------------------------------------------------------------------------

func TestClearCell_Success(t *testing.T) {
	b := sudoku.NewBoard()
	_ = b.PlaceNumberForce(4, 4, 9)
	if err := b.ClearCell(4, 4); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.GetBoard()[4][4] != 0 {
		t.Fatal("cell was not cleared")
	}
}

func TestClearCell_LockedCell(t *testing.T) {
	var m [sudoku.BoardSize][sudoku.BoardSize]int
	m[2][3] = 4
	b := sudoku.NewBoard()
	_ = b.SetBoard(m)

	if err := b.ClearCell(2, 3); err != sudoku.ErrCellLocked {
		t.Fatalf("expected ErrCellLocked, got %v", err)
	}
}

func TestClearCell_OutOfBounds(t *testing.T) {
	b := sudoku.NewBoard()
	if err := b.ClearCell(-1, 0); err != sudoku.ErrOutOfBounds {
		t.Fatalf("expected ErrOutOfBounds, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Undo
// ---------------------------------------------------------------------------

func TestUndo_RevertsPlacement(t *testing.T) {
	b := sudoku.NewBoard()
	_ = b.PlaceNumberForceUndo(0, 0, 5)
	r, c, ok := b.Undo()
	if !ok {
		t.Fatal("expected undo to succeed")
	}
	if r != 0 || c != 0 {
		t.Errorf("expected undo to return (0,0), got (%d,%d)", r, c)
	}
	if b.GetBoard()[0][0] != 0 {
		t.Fatal("undo did not restore previous value")
	}
}

func TestUndo_EmptyStackReturnsFalse(t *testing.T) {
	b := sudoku.NewBoard()
	_, _, ok := b.Undo()
	if ok {
		t.Fatal("undo on empty stack should return false")
	}
}

// ---------------------------------------------------------------------------
// GetBoard / GetLockedBoard / GetPlayerBoard
// ---------------------------------------------------------------------------

func TestGetBoard_ReturnsSnapshot(t *testing.T) {
	b := sudoku.NewBoard()
	_ = b.PlaceNumberForce(3, 3, 6)
	m := b.GetBoard()
	// Mutating the returned matrix must not affect the board.
	m[3][3] = 0
	if b.GetBoard()[3][3] != 6 {
		t.Fatal("GetBoard returned a reference instead of a copy")
	}
}

func TestGetLockedBoard_OnlyGivens(t *testing.T) {
	b := sudoku.NewBoard()
	var m [sudoku.BoardSize][sudoku.BoardSize]int
	m[0][0] = 5
	_ = b.SetBoard(m)
	_ = b.PlaceNumberForce(1, 0, 3) // player move

	locked := b.GetLockedBoard()
	if locked[0][0] != 5 {
		t.Error("expected given at [0][0]")
	}
	if locked[1][0] != 0 {
		t.Error("player-placed value should not appear in GetLockedBoard")
	}
}

func TestGetPlayerBoard_OnlyPlayerMoves(t *testing.T) {
	b := sudoku.NewBoard()
	var m [sudoku.BoardSize][sudoku.BoardSize]int
	m[0][0] = 5
	_ = b.SetBoard(m)
	_ = b.PlaceNumberForce(1, 0, 3)

	player := b.GetPlayerBoard()
	if player[0][0] != 0 {
		t.Error("given should not appear in GetPlayerBoard")
	}
	if player[1][0] != 3 {
		t.Error("player-placed value should appear in GetPlayerBoard")
	}
}

// ---------------------------------------------------------------------------
// Conflicts
// ---------------------------------------------------------------------------

func TestConflicts_FlagsRowDuplicate(t *testing.T) {
	b := sudoku.NewBoard()
	_ = b.PlaceNumberForce(0, 0, 5)
	_ = b.PlaceNumberForce(0, 1, 5)
	c := b.Conflicts()
	if !c[0][0] || !c[0][1] {
		t.Error("row duplicates should both be flagged as conflicts")
	}
}

func TestConflicts_FlagsColDuplicate(t *testing.T) {
	b := sudoku.NewBoard()
	_ = b.PlaceNumberForce(0, 0, 5)
	_ = b.PlaceNumberForce(1, 0, 5)
	c := b.Conflicts()
	if !c[0][0] || !c[1][0] {
		t.Error("column duplicates should both be flagged")
	}
}

func TestConflicts_FlagsBoxDuplicate(t *testing.T) {
	b := sudoku.NewBoard()
	_ = b.PlaceNumberForce(0, 0, 5)
	_ = b.PlaceNumberForce(1, 1, 5) // same box
	c := b.Conflicts()
	if !c[0][0] || !c[1][1] {
		t.Error("box duplicates should both be flagged")
	}
}

func TestConflicts_NoConflicts(t *testing.T) {
	b := sudoku.NewBoard()
	_ = b.PlaceNumberForce(0, 0, 1)
	_ = b.PlaceNumberForce(0, 1, 2)
	_ = b.PlaceNumberForce(1, 0, 3)
	c := b.Conflicts()
	for r := 0; r < sudoku.BoardSize; r++ {
		for col := 0; col < sudoku.BoardSize; col++ {
			if c[r][col] {
				t.Errorf("unexpected conflict at [%d][%d]", r, col)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// IsSolved
// ---------------------------------------------------------------------------

func TestIsSolved_EmptyBoardNotSolved(t *testing.T) {
	b := sudoku.NewBoard()
	if b.IsSolved() {
		t.Fatal("empty board should not be solved")
	}
}

func TestIsSolved_AfterGenerate(t *testing.T) {
	// A generated puzzle should not immediately be solved (has empty cells).
	puzzle := sudoku.GenerateBoard(sudoku.Easy)
	b := sudoku.NewBoard()
	_ = b.SetBoard(puzzle)
	if b.IsSolved() {
		t.Fatal("freshly generated puzzle should not be solved")
	}
}

func TestIsSolved_WhenBoardMatchesSolution(t *testing.T) {
	// Fill the board with the cached solution — IsSolved must return true.
	b := sudoku.NewBoard()
	_ = b.SetBoard(sudoku.GenerateBoard(sudoku.Easy))
	sol := b.GetSolution()
	// Overwrite every non-locked cell with its solution value.
	player := b.GetLockedBoard()
	for r := 0; r < sudoku.BoardSize; r++ {
		for c := 0; c < sudoku.BoardSize; c++ {
			if player[r][c] == 0 {
				_ = b.PlaceNumberForce(r, c, sol[r][c])
			}
		}
	}
	if !b.IsSolved() {
		t.Fatal("board filled with solution values should be solved")
	}
}

func TestIsSolved_WithConflictNotSolved(t *testing.T) {
	// A full board with a conflict must not be considered solved.
	b := sudoku.NewBoard()
	_ = b.SetBoard(sudoku.GenerateBoard(sudoku.Easy))
	sol := b.GetSolution()
	// Fill all player cells correctly first.
	for r := 0; r < sudoku.BoardSize; r++ {
		for c := 0; c < sudoku.BoardSize; c++ {
			if b.GetLockedBoard()[r][c] == 0 {
				_ = b.PlaceNumberForce(r, c, sol[r][c])
			}
		}
	}
	// Now corrupt one player cell by writing a conflicting value.
	// Find a non-locked cell and set it to a wrong value.
	for r := 0; r < sudoku.BoardSize; r++ {
		for c := 0; c < sudoku.BoardSize; c++ {
			if b.LockedCells()[r][c] {
				continue
			}
			wrong := (sol[r][c] % 9) + 1 // guaranteed != sol[r][c]
			_ = b.PlaceNumberForce(r, c, wrong)
			if b.IsSolved() {
				t.Fatal("board with wrong value should not be solved")
			}
			return
		}
	}
}

// ---------------------------------------------------------------------------
// GetSolution / SetSolutionCache
// ---------------------------------------------------------------------------

func TestGetSolution_EmptyBoardReturnsZero(t *testing.T) {
	b := sudoku.NewBoard()
	sol := b.GetSolution()
	for r := 0; r < sudoku.BoardSize; r++ {
		for c := 0; c < sudoku.BoardSize; c++ {
			if sol[r][c] != 0 {
				t.Fatalf("expected zero solution on empty board, got %d at [%d][%d]", sol[r][c], r, c)
			}
		}
	}
}

func TestGetSolution_SetBySetBoard(t *testing.T) {
	b := sudoku.NewBoard()
	_ = b.SetBoard(sudoku.GenerateBoard(sudoku.Easy))
	sol := b.GetSolution()
	// Verify the solution is a valid complete board (all 1–9, no zeros).
	for r := 0; r < sudoku.BoardSize; r++ {
		for c := 0; c < sudoku.BoardSize; c++ {
			if sol[r][c] < 1 || sol[r][c] > 9 {
				t.Fatalf("solution cell [%d][%d] = %d, expected 1–9", r, c, sol[r][c])
			}
		}
	}
}

func TestSetSolutionCache_OverridesSolution(t *testing.T) {
	b := sudoku.NewBoard()
	_ = b.SetBoard(sudoku.GenerateBoard(sudoku.Easy))
	original := b.GetSolution()

	// Build a known synthetic solution (all 1s — not valid Sudoku, but
	// sufficient to verify the cache is stored and retrieved correctly).
	var fake [sudoku.BoardSize][sudoku.BoardSize]int
	for r := 0; r < sudoku.BoardSize; r++ {
		for c := 0; c < sudoku.BoardSize; c++ {
			fake[r][c] = 1
		}
	}
	b.SetSolutionCache(fake)

	got := b.GetSolution()
	if got == original {
		t.Fatal("SetSolutionCache did not replace the solution")
	}
	if got != fake {
		t.Fatal("GetSolution did not return the value set by SetSolutionCache")
	}
}
