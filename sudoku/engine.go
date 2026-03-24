// Package sudoku implements a Sudoku game engine.
//
// The engine exposes a Board type with the following public operations:
//
//   - PlaceNumberForce     – place a digit (1-9) even if it conflicts; shown in red.
//   - PlaceNumberForceUndo – like PlaceNumberForce, but records the move for undo.
//   - ClearCell / ClearCellUndo – remove a player-placed digit.
//   - Undo                 – reverse the most recent undoable action.
//   - GetBoard             – return the current state as a 9×9 matrix.
//   - SetBoard             – overwrite the board with a new puzzle matrix (locks givens).
//   - IsSolved             – true when all cells match the cached solution.
//   - Conflicts            – returns a 9×9 matrix flagging rule-violating cells.
//   - GetSolution          – returns the fully-solved grid cached at SetBoard time.
//
// Example:
//
//	b := sudoku.NewBoard()
//	_ = b.SetBoard(sudoku.GenerateBoard(sudoku.Easy))
//	_ = b.PlaceNumberForceUndo(0, 0, 5)
//	matrix := b.GetBoard()
package sudoku

// BoardSize is the number of rows/columns in a standard Sudoku grid.
const BoardSize = 9

// BoxSize is the side-length of each 3×3 sub-box.
const BoxSize = 3

// cell holds a single square on the board.
type cell struct {
	value  int  // 0 = empty, 1-9 = filled
	locked bool // true when the value was set as part of puzzle initialisation
}

// move records a single player action for undo purposes.
type move struct {
	row, col int
	oldVal   int // value before the action (0 = was empty)
	newVal   int // value after the action (0 = cell was cleared)
}

// Board represents the full 9×9 Sudoku grid and exposes the game engine API.
type Board struct {
	cells     [BoardSize][BoardSize]cell
	undoStack []move
	// solution is the fully-solved grid, cached by SetBoard/SetSolution so
	// GetHint can return a guaranteed-correct answer.
	solution [BoardSize][BoardSize]int
}

// NewBoard returns a new, empty Board ready for play.
func NewBoard() *Board {
	return &Board{}
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

// PlaceNumberForce places val at (row, col) even if it conflicts with other
// cells. Used by the UI to allow the player to enter a wrong answer (which is
// then shown in red) rather than silently blocking the input.
// Returns ErrOutOfBounds, ErrInvalidValue, or ErrCellLocked on hard errors;
// returns nil (and writes the value) even when a conflict exists.
func (b *Board) PlaceNumberForce(row, col, val int) error {
	if err := boundsCheck(row, col); err != nil {
		return err
	}
	if val < 1 || val > 9 {
		return ErrInvalidValue
	}
	if b.cells[row][col].locked {
		return ErrCellLocked
	}
	b.cells[row][col].value = val
	return nil
}

// ClearCell removes a player-placed value from (row, col), setting it to 0.
//
// Returns an error if the coordinates are out of bounds or the cell is locked.
func (b *Board) ClearCell(row, col int) error {
	if err := boundsCheck(row, col); err != nil {
		return err
	}
	if b.cells[row][col].locked {
		return ErrCellLocked
	}
	b.cells[row][col].value = 0
	return nil
}
func (b *Board) IsSolved() bool {
	if b.solution == ([BoardSize][BoardSize]int{}) {
		// No solution cached — fall back to the full conflict check.
		for r := 0; r < BoardSize; r++ {
			for c := 0; c < BoardSize; c++ {
				if b.cells[r][c].value == 0 {
					return false
				}
			}
		}
		conflicts := b.Conflicts()
		for r := 0; r < BoardSize; r++ {
			for c := 0; c < BoardSize; c++ {
				if conflicts[r][c] {
					return false
				}
			}
		}
		return true
	}
	// Fast path: compare directly against the cached solution — O(81), no allocs.
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			if b.cells[r][c].value != b.solution[r][c] {
				return false
			}
		}
	}
	return true
}

// GetBoard returns the current board state as a 9×9 integer matrix.
// Empty cells are represented by 0; filled cells by their digit (1-9).
func (b *Board) GetBoard() [BoardSize][BoardSize]int {
	var matrix [BoardSize][BoardSize]int
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			matrix[r][c] = b.cells[r][c].value
		}
	}
	return matrix
}

// GetLockedBoard returns a 9×9 matrix containing only the puzzle givens
// (locked cells). Player-placed values are returned as 0.
// Use this to persist the original puzzle so it can be reloaded cleanly.
func (b *Board) GetLockedBoard() [BoardSize][BoardSize]int {
	var matrix [BoardSize][BoardSize]int
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			if b.cells[r][c].locked {
				matrix[r][c] = b.cells[r][c].value
			}
		}
	}
	return matrix
}

// GetPlayerBoard returns a 9×9 matrix containing only player-placed values.
// Locked (given) cells are returned as 0.
// Use this alongside GetLockedBoard to persist and restore a session.
func (b *Board) GetPlayerBoard() [BoardSize][BoardSize]int {
	var matrix [BoardSize][BoardSize]int
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			if !b.cells[r][c].locked {
				matrix[r][c] = b.cells[r][c].value
			}
		}
	}
	return matrix
}

// RestorePlayerBoard writes player-placed values back onto the board without
// locking them. Locked cells are skipped. Used when resuming a saved session.
func (b *Board) RestorePlayerBoard(player [BoardSize][BoardSize]int) {
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			if !b.cells[r][c].locked && player[r][c] != 0 {
				b.cells[r][c].value = player[r][c]
			}
		}
	}
}

// SetBoard replaces the entire board with the provided 9×9 matrix.
// All non-zero values are treated as new "givens" (locked cells).
// The operation is atomic: if any value is invalid or causes a conflict the
// board is left unchanged and an error is returned.
//
// Possible errors: ErrBoardSize, ErrInvalidValue, ErrConflict.
func (b *Board) SetBoard(matrix [BoardSize][BoardSize]int) error {
	if err := b.loadMatrix(matrix, true); err != nil {
		return err
	}
	// Cache the solution so GetHint can answer correctly.
	sol := matrix
	count := 0
	SolveMatrix(&sol, &count)
	if count > 0 {
		b.solution = sol
	}
	b.undoStack = nil
	return nil
}

// PlaceNumberForceUndo is like PlaceNumberForce but records the action on
// the undo stack. Returns the same errors as PlaceNumberForce.
func (b *Board) PlaceNumberForceUndo(row, col, val int) error {
	old := b.cells[row][col].value
	if err := b.PlaceNumberForce(row, col, val); err != nil {
		return err
	}
	b.undoStack = append(b.undoStack, move{row: row, col: col, oldVal: old, newVal: val})
	return nil
}

// ClearCellUndo is like ClearCell but records the action on the undo stack.
func (b *Board) ClearCellUndo(row, col int) error {
	old := b.cells[row][col].value
	if err := b.ClearCell(row, col); err != nil {
		return err
	}
	b.undoStack = append(b.undoStack, move{row: row, col: col, oldVal: old, newVal: 0})
	return nil
}

// Undo reverses the most recent PlaceNumberForceUndo or ClearCellUndo.
// Returns (row, col, true) of the affected cell so the UI can re-select it,
// or (-1, -1, false) if the stack is empty.
func (b *Board) Undo() (int, int, bool) {
	if len(b.undoStack) == 0 {
		return -1, -1, false
	}
	m := b.undoStack[len(b.undoStack)-1]
	b.undoStack = b.undoStack[:len(b.undoStack)-1]
	b.cells[m.row][m.col].value = m.oldVal
	return m.row, m.col, true
}

// CanUndo reports whether there is anything to undo.
func (b *Board) CanUndo() bool {
	return len(b.undoStack) > 0
}

// GetUndoStack returns a snapshot of the undo stack as a flat slice of
// [row, col, oldVal, newVal] quads. Used for session persistence.
func (b *Board) GetUndoStack() [][4]int {
	out := make([][4]int, len(b.undoStack))
	for i, m := range b.undoStack {
		out[i] = [4]int{m.row, m.col, m.oldVal, m.newVal}
	}
	return out
}

// RestoreUndoStack replaces the undo stack from a previously saved snapshot.
// Call this after RestorePlayerBoard when resuming a session.
func (b *Board) RestoreUndoStack(stack [][4]int) {
	b.undoStack = make([]move, len(stack))
	for i, q := range stack {
		b.undoStack[i] = move{row: q[0], col: q[1], oldVal: q[2], newVal: q[3]}
	}
}

// GetHint returns (row, col, value, true) for one randomly chosen empty
// (or incorrectly filled) non-locked cell, using the cached solution.
// Returns (0, 0, 0, false) if no hint is available (puzzle solved, or no
// solution was cached).
func (b *Board) GetHint(rng interface{ Intn(int) int }) (int, int, int, bool) {
	if b.solution == ([BoardSize][BoardSize]int{}) {
		return 0, 0, 0, false
	}
	type candidate struct{ r, c, v int }
	var candidates []candidate
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			if b.cells[r][c].locked {
				continue
			}
			want := b.solution[r][c]
			if b.cells[r][c].value != want {
				candidates = append(candidates, candidate{r, c, want})
			}
		}
	}
	if len(candidates) == 0 {
		return 0, 0, 0, false
	}
	pick := candidates[rng.Intn(len(candidates))]
	return pick.r, pick.c, pick.v, true
}

// GetSolution returns the fully-solved grid that was cached when SetBoard was
// called. Returns an all-zero matrix if no solution has been cached yet.
func (b *Board) GetSolution() [BoardSize][BoardSize]int {
	return b.solution
}

// SetSolutionCache explicitly stores the fully-solved grid. Use this when
// restoring a saved session that persisted the solution separately.
func (b *Board) SetSolutionCache(sol [BoardSize][BoardSize]int) {
	b.solution = sol
}

// LockedCells returns a 9×9 boolean matrix where true means the cell is a
// puzzle "given" (locked) and cannot be modified by the player.
func (b *Board) LockedCells() [BoardSize][BoardSize]bool {
	var m [BoardSize][BoardSize]bool
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			m[r][c] = b.cells[r][c].locked
		}
	}
	return m
}

// Conflicts returns a 9×9 boolean matrix where true means the cell's current
// value violates Sudoku rules (duplicate in its row, column, or 3×3 box).
// Empty cells (value 0) are never flagged.
//
// The implementation uses a single-pass bitmask strategy: each row, column,
// and box gets a uint16 where bit n is set when digit n has already been seen.
// A cell is a conflict when its digit bit is already set in any of its three
// groups (row, col, box). Both occurrences are marked, so the second pass
// flags each cell whose value bit is set in the pre-built duplicate mask.
func (b *Board) Conflicts() [BoardSize][BoardSize]bool {
	// First pass: count occurrences per digit per group using bitmask OR.
	// duplicateRows[r] has bit n set if digit n appears more than once in row r.
	var seenRows, seenCols, seenBoxes [BoardSize]uint16
	var dupRows, dupCols, dupBoxes [BoardSize]uint16

	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			v := b.cells[r][c].value
			if v == 0 {
				continue
			}
			bit := uint16(1) << v
			box := (r/BoxSize)*BoxSize + c/BoxSize

			if seenRows[r]&bit != 0 {
				dupRows[r] |= bit
			}
			seenRows[r] |= bit

			if seenCols[c]&bit != 0 {
				dupCols[c] |= bit
			}
			seenCols[c] |= bit

			if seenBoxes[box]&bit != 0 {
				dupBoxes[box] |= bit
			}
			seenBoxes[box] |= bit
		}
	}

	// Second pass: flag each cell whose digit is in any duplicate mask.
	var m [BoardSize][BoardSize]bool
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			v := b.cells[r][c].value
			if v == 0 {
				continue
			}
			bit := uint16(1) << v
			box := (r/BoxSize)*BoxSize + c/BoxSize
			if dupRows[r]&bit != 0 || dupCols[c]&bit != 0 || dupBoxes[box]&bit != 0 {
				m[r][c] = true
			}
		}
	}
	return m
}

// CountDigits returns the count of each placed digit (1-9) on the board.
// Index 0 is always 0; indices 1-9 hold the count of that digit.
// This avoids the matrix allocation of GetBoard() for digit-count queries.
func (b *Board) CountDigits() [10]int {
	var counts [10]int
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			if v := b.cells[r][c].value; v >= 1 && v <= 9 {
				counts[v]++
			}
		}
	}
	return counts
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// loadMatrix validates then applies a 9×9 matrix to the board.
// When lock is true, every non-zero cell is marked as a puzzle given.
func (b *Board) loadMatrix(matrix [BoardSize][BoardSize]int, lock bool) error {
	// Validate all values before touching the board (atomic update).
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			v := matrix[r][c]
			if v < 0 || v > 9 {
				return ErrInvalidValue
			}
		}
	}

	// Build a fresh grid so we can validate conflicts before committing.
	var draft [BoardSize][BoardSize]cell
	for r := 0; r < BoardSize; r++ {
		for c := 0; c < BoardSize; c++ {
			v := matrix[r][c]
			if v != 0 {
				if !isValidPlacement(draft, r, c, v) {
					return ErrConflict
				}
				draft[r][c] = cell{value: v, locked: lock}
			}
		}
	}

	b.cells = draft
	return nil
}

// boundsCheck returns ErrOutOfBounds if row or col is outside [0, BoardSize).
func boundsCheck(row, col int) error {
	if row < 0 || row >= BoardSize || col < 0 || col >= BoardSize {
		return ErrOutOfBounds
	}
	return nil
}
