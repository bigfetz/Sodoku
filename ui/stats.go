package ui

import (
	"fmt"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"sudoku_game/sudoku"
)

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
