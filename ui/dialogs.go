package ui

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"sudoku_game/sudoku"
)

// ---- Stats ----------------------------------------------------------------

// difficultyName maps Difficulty to a display string.
var difficultyName = map[sudoku.Difficulty]string{
	sudoku.VeryEasy: "Very Easy",
	sudoku.Easy:     "Easy",
	sudoku.Medium:   "Medium",
	sudoku.Hard:     "Hard",
	sudoku.VeryHard: "Very Hard",
}

// difficultyOrder is the display order for the stats screen.
var difficultyOrder = []sudoku.Difficulty{
	sudoku.VeryEasy,
	sudoku.Easy,
	sudoku.Medium,
	sudoku.Hard,
	sudoku.VeryHard,
}

// prefKey builds a Preferences key for a given difficulty and stat name.
func prefKey(d sudoku.Difficulty, stat string) string {
	return fmt.Sprintf("stats.%d.%s", int(d), stat)
}

// RecordWin saves a win for the given difficulty and time (seconds).
// Updates win count and best time.
func RecordWin(prefs fyne.Preferences, d sudoku.Difficulty, elapsed int) {
	wins := prefs.IntWithFallback(prefKey(d, "wins"), 0)
	prefs.SetInt(prefKey(d, "wins"), wins+1)

	best := prefs.IntWithFallback(prefKey(d, "best"), math.MaxInt32)
	if elapsed < best {
		prefs.SetInt(prefKey(d, "best"), elapsed)
	}
}

// ResetStats wipes all stored stats.
func ResetStats(prefs fyne.Preferences) {
	for _, d := range difficultyOrder {
		prefs.SetInt(prefKey(d, "wins"), 0)
		prefs.SetInt(prefKey(d, "best"), math.MaxInt32)
	}
}

// formatBest formats a best-time int (seconds) as mm:ss, or "–" if unset.
func formatBest(secs int) string {
	if secs == math.MaxInt32 || secs == 0 {
		return "–"
	}
	return fmt.Sprintf("%d:%02d", secs/60, secs%60)
}

// ShowStatsDialog opens the stats screen as a Fyne dialog.
func ShowStatsDialog(w fyne.Window, prefs fyne.Preferences) {
	// Build a grid: header row + one row per difficulty.
	header := []fyne.CanvasObject{
		widget.NewLabelWithStyle("Difficulty", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Wins", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Best Time", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
	}

	rows := make([]fyne.CanvasObject, 0, len(difficultyOrder)*3)
	for _, d := range difficultyOrder {
		wins := prefs.IntWithFallback(prefKey(d, "wins"), 0)
		best := prefs.IntWithFallback(prefKey(d, "best"), math.MaxInt32)

		rows = append(rows,
			widget.NewLabel(difficultyName[d]),
			widget.NewLabelWithStyle(fmt.Sprintf("%d", wins), fyne.TextAlignCenter, fyne.TextStyle{}),
			widget.NewLabelWithStyle(formatBest(best), fyne.TextAlignCenter, fyne.TextStyle{}),
		)
	}

	allCells := append(header, rows...)
	grid := container.NewGridWithColumns(3, allCells...)

	resetBtn := widget.NewButton("Reset Stats", func() {
		dialog.ShowConfirm(
			"Reset Stats",
			"Clear all wins and best times?",
			func(ok bool) {
				if ok {
					ResetStats(prefs)
				}
			},
			w,
		)
	})
	resetBtn.Importance = widget.DangerImportance

	content := container.NewVBox(
		grid,
		widget.NewSeparator(),
		resetBtn,
	)

	dlg := dialog.NewCustom("Statistics", "Close", content, w)
	dlg.Show()
}

// ---- New-game dialog ------------------------------------------------------

// showNewGameDialog presents Easy / Medium / Hard choices and generates a
// fresh random puzzle for the selected difficulty.
func showNewGameDialog(w fyne.Window, bw *BoardWidget, difficulty *sudoku.Difficulty) {
	type option struct {
		label string
		diff  sudoku.Difficulty
	}
	options := []option{
		{"Very Easy", sudoku.VeryEasy},
		{"Easy", sudoku.Easy},
		{"Medium", sudoku.Medium},
		{"Hard", sudoku.Hard},
		{"Very Hard", sudoku.VeryHard},
	}

	var dlg dialog.Dialog
	buttons := make([]fyne.CanvasObject, len(options))
	for i, o := range options {
		o := o // capture
		btn := widget.NewButton(o.label, func() {
			dlg.Hide()
			*difficulty = o.diff
			puzzle := sudoku.GenerateBoard(o.diff)
			bw.UpdateBoard(puzzle) //nolint:errcheck
		})
		buttons[i] = btn
	}

	content := container.NewVBox(
		widget.NewLabel("Choose difficulty:"),
		container.NewGridWithColumns(3, buttons[:3]...),
		container.NewGridWithColumns(2, buttons[3:]...),
	)
	dlg = dialog.NewCustom("New Game", "Cancel", content, w)
	dlg.Show()
}

// ---- Solve dialog ---------------------------------------------------------

// showSolveConfirm asks the user to confirm before auto-solving.
// It solves the puzzle in memory first, then places each answer cell one by
// one with a short delay so the player can watch the board fill in.
func showSolveConfirm(w fyne.Window, bw *BoardWidget) {
	dialog.ShowConfirm(
		"Auto-Solve",
		"Solve the puzzle automatically?",
		func(ok bool) {
			if !ok {
				return
			}

			// Snapshot the current board and solve a copy in memory.
			snapshot := bw.board.GetBoard()
			solved := snapshot // will be overwritten by solver
			count := 0
			sudoku.SolveMatrix(&solved, &count)

			if count == 0 {
				fyne.Do(func() {
					dialog.ShowInformation("No Solution", "This puzzle has no valid solution.", w)
				})
				return
			}

			// Collect only the cells that were empty (need to be filled).
			type cell struct{ r, c, v int }
			var toPlace []cell
			for r := 0; r < sudoku.BoardSize; r++ {
				for c := 0; c < sudoku.BoardSize; c++ {
					if snapshot[r][c] == 0 {
						toPlace = append(toPlace, cell{r, c, solved[r][c]})
					}
				}
			}

			// Shuffle so numbers appear in a random order.
			rand.Shuffle(len(toPlace), func(i, j int) {
				toPlace[i], toPlace[j] = toPlace[j], toPlace[i]
			})

			// Place each answer on the main thread with a short delay between.
			go func() {
				fyne.Do(func() {
					bw.solving = true
					if bw.OnSelectionChanged != nil {
						bw.OnSelectionChanged()
					}
				})
				for _, cell := range toPlace {
					cell := cell
					fyne.Do(func() {
						bw.selRow = cell.r
						bw.selCol = cell.c
						bw.PlaceDigit(cell.v)
					})
					time.Sleep(80 * time.Millisecond)
				}
				// Final cleanup — deselect, clear solving flag, refresh numpad, wave.
				fyne.Do(func() {
					bw.solved = true // prevent any trailing save from resurrecting the session
					bw.solving = false
					bw.selRow = -1
					bw.selCol = -1
					bw.Refresh()
					bw.notifyDigitCounts()
					if bw.OnSelectionChanged != nil {
						bw.OnSelectionChanged()
					}
					bw.WaveAllCells()
					// Puzzle is now complete via auto-solve — stop the timer only,
					// do NOT record in stats.
					if bw.OnAutoSolved != nil {
						bw.OnAutoSolved()
					}
				})
			}()
		},
		w,
	)
}
