package sudoku_test

import (
	"testing"

	"sudoku_game/sudoku"
)

// The validator helpers (conflictsInRow/Col/Box) are internal, so we exercise
// them through the public Board.Conflicts() API which delegates to them.

func TestConflicts_EmptyBoard(t *testing.T) {
	b := sudoku.NewBoard()
	c := b.Conflicts()
	for r := 0; r < sudoku.BoardSize; r++ {
		for col := 0; col < sudoku.BoardSize; col++ {
			if c[r][col] {
				t.Errorf("empty board should have no conflicts, got one at [%d][%d]", r, col)
			}
		}
	}
}

func TestConflicts_RowDuplicate(t *testing.T) {
	b := sudoku.NewBoard()
	_ = b.PlaceNumberForce(0, 0, 7)
	_ = b.PlaceNumberForce(0, 8, 7) // same row, different col
	c := b.Conflicts()
	if !c[0][0] {
		t.Error("expected [0][0] to be flagged as conflict")
	}
	if !c[0][8] {
		t.Error("expected [0][8] to be flagged as conflict")
	}
	// No other cell should be flagged.
	for col := 1; col <= 7; col++ {
		if c[0][col] {
			t.Errorf("unexpected conflict at [0][%d]", col)
		}
	}
}

func TestConflicts_ColDuplicate(t *testing.T) {
	b := sudoku.NewBoard()
	_ = b.PlaceNumberForce(0, 3, 4)
	_ = b.PlaceNumberForce(8, 3, 4) // same col, different row
	c := b.Conflicts()
	if !c[0][3] || !c[8][3] {
		t.Error("column duplicates should both be flagged")
	}
}

func TestConflicts_BoxDuplicate(t *testing.T) {
	b := sudoku.NewBoard()
	// Top-left 3×3 box.
	_ = b.PlaceNumberForce(0, 0, 3)
	_ = b.PlaceNumberForce(2, 2, 3)
	c := b.Conflicts()
	if !c[0][0] || !c[2][2] {
		t.Error("box duplicates should both be flagged")
	}
	// Cell in a different box should not be affected.
	_ = b.PlaceNumberForce(3, 3, 3) // middle-centre box
	c = b.Conflicts()
	if c[3][3] {
		t.Error("[3][3] should not be flagged — it is in a different box")
	}
}

func TestConflicts_TripleDuplicate(t *testing.T) {
	b := sudoku.NewBoard()
	_ = b.PlaceNumberForce(0, 0, 9)
	_ = b.PlaceNumberForce(0, 1, 9)
	_ = b.PlaceNumberForce(0, 2, 9) // three 9s in row 0
	c := b.Conflicts()
	if !c[0][0] || !c[0][1] || !c[0][2] {
		t.Error("all three duplicate cells in a row should be flagged")
	}
}

func TestConflicts_SingleCellNoConflict(t *testing.T) {
	b := sudoku.NewBoard()
	_ = b.PlaceNumberForce(4, 4, 5)
	c := b.Conflicts()
	if c[4][4] {
		t.Error("single cell with no duplicates should not be a conflict")
	}
}
