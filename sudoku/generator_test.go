package sudoku_test

import (
	"testing"

	"sudoku_game/sudoku"
)

// ---------------------------------------------------------------------------
// GenerateBoard
// ---------------------------------------------------------------------------

func TestGenerateBoard_HasExpectedGivens(t *testing.T) {
	tests := []struct {
		diff      sudoku.Difficulty
		wantMin   int
		wantLabel string
	}{
		{sudoku.VeryEasy, 40, "VeryEasy"},
		{sudoku.Easy, 30, "Easy"},
		{sudoku.Medium, 22, "Medium"},
		{sudoku.Hard, 17, "Hard"},
		{sudoku.VeryHard, 17, "VeryHard"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.wantLabel, func(t *testing.T) {
			puzzle := sudoku.GenerateBoard(tt.diff)
			count := 0
			for r := 0; r < sudoku.BoardSize; r++ {
				for c := 0; c < sudoku.BoardSize; c++ {
					if puzzle[r][c] != 0 {
						count++
					}
				}
			}
			if count < tt.wantMin {
				t.Errorf("%s: expected at least %d givens, got %d", tt.wantLabel, tt.wantMin, count)
			}
		})
	}
}

func TestGenerateBoard_IsValidPuzzle(t *testing.T) {
	puzzle := sudoku.GenerateBoard(sudoku.Easy)

	// Load puzzle into a Board — SetBoard validates conflicts.
	b := sudoku.NewBoard()
	if err := b.SetBoard(puzzle); err != nil {
		t.Fatalf("generated puzzle fails SetBoard validation: %v", err)
	}

	// No given should be in conflict.
	conflicts := b.Conflicts()
	for r := 0; r < sudoku.BoardSize; r++ {
		for c := 0; c < sudoku.BoardSize; c++ {
			if conflicts[r][c] {
				t.Errorf("conflict at [%d][%d] in generated puzzle", r, c)
			}
		}
	}
}

func TestGenerateBoard_HasUniqueSolution(t *testing.T) {
	// SolveMatrix is limited to 2 solutions max; unique means count==1.
	puzzle := sudoku.GenerateBoard(sudoku.Easy)
	count := 0
	sudoku.SolveMatrix(&puzzle, &count)
	if count != 1 {
		t.Errorf("expected exactly 1 solution, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// SolveMatrix
// ---------------------------------------------------------------------------

func TestSolveMatrix_SolvesKnownPuzzle(t *testing.T) {
	// A well-known minimal Sudoku puzzle with a unique solution.
	var m [sudoku.BoardSize][sudoku.BoardSize]int
	givens := [][3]int{
		{0, 0, 5}, {0, 1, 3}, {0, 4, 7},
		{1, 0, 6}, {1, 3, 1}, {1, 4, 9}, {1, 5, 5},
		{2, 1, 9}, {2, 2, 8}, {2, 7, 6},
		{3, 0, 8}, {3, 4, 6}, {3, 8, 3},
		{4, 0, 4}, {4, 3, 8}, {4, 5, 3}, {4, 8, 1},
		{5, 0, 7}, {5, 4, 2}, {5, 8, 6},
		{6, 1, 6}, {6, 6, 2}, {6, 7, 8},
		{7, 3, 4}, {7, 4, 1}, {7, 5, 9}, {7, 8, 5},
		{8, 4, 8}, {8, 7, 7}, {8, 8, 9},
	}
	for _, g := range givens {
		m[g[0]][g[1]] = g[2]
	}
	count := 0
	sudoku.SolveMatrix(&m, &count)
	if count != 1 {
		t.Fatalf("expected exactly 1 solution, got %d", count)
	}
	// Verify the board is now fully filled.
	for r := 0; r < sudoku.BoardSize; r++ {
		for c := 0; c < sudoku.BoardSize; c++ {
			if m[r][c] == 0 {
				t.Errorf("[%d][%d] is still empty after solve", r, c)
			}
		}
	}
}
