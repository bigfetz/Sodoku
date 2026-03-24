package ui

import (
	"math"
	"testing"

	"fyne.io/fyne/v2/test"

	"sudoku_game/sudoku"
)

// ---------------------------------------------------------------------------
// prefKey
// ---------------------------------------------------------------------------

func TestPrefKey_Format(t *testing.T) {
	key := prefKey(sudoku.Easy, "wins")
	expected := "stats.36.wins"
	if key != expected {
		t.Errorf("prefKey: want %q, got %q", expected, key)
	}
}

// ---------------------------------------------------------------------------
// RecordWin
// ---------------------------------------------------------------------------

func TestRecordWin_IncrementsWinCount(t *testing.T) {
	prefs := test.NewApp().Preferences()

	RecordWin(prefs, sudoku.Easy, 60)
	RecordWin(prefs, sudoku.Easy, 90)

	wins := prefs.IntWithFallback(prefKey(sudoku.Easy, "wins"), 0)
	if wins != 2 {
		t.Errorf("want 2 wins, got %d", wins)
	}
}

func TestRecordWin_StoresBestTime(t *testing.T) {
	prefs := test.NewApp().Preferences()

	RecordWin(prefs, sudoku.Medium, 120)
	RecordWin(prefs, sudoku.Medium, 60) // faster
	RecordWin(prefs, sudoku.Medium, 90) // slower than best

	best := prefs.IntWithFallback(prefKey(sudoku.Medium, "best"), math.MaxInt32)
	if best != 60 {
		t.Errorf("want best=60, got %d", best)
	}
}

func TestRecordWin_FirstWinSetsBest(t *testing.T) {
	prefs := test.NewApp().Preferences()

	RecordWin(prefs, sudoku.Hard, 200)

	best := prefs.IntWithFallback(prefKey(sudoku.Hard, "best"), math.MaxInt32)
	if best != 200 {
		t.Errorf("want best=200, got %d", best)
	}
}

func TestRecordWin_SeparatesByDifficulty(t *testing.T) {
	prefs := test.NewApp().Preferences()

	RecordWin(prefs, sudoku.Easy, 50)
	RecordWin(prefs, sudoku.Hard, 300)

	easyWins := prefs.IntWithFallback(prefKey(sudoku.Easy, "wins"), 0)
	hardWins := prefs.IntWithFallback(prefKey(sudoku.Hard, "wins"), 0)

	if easyWins != 1 {
		t.Errorf("easy wins: want 1, got %d", easyWins)
	}
	if hardWins != 1 {
		t.Errorf("hard wins: want 1, got %d", hardWins)
	}

	easyBest := prefs.IntWithFallback(prefKey(sudoku.Easy, "best"), math.MaxInt32)
	hardBest := prefs.IntWithFallback(prefKey(sudoku.Hard, "best"), math.MaxInt32)

	if easyBest != 50 {
		t.Errorf("easy best: want 50, got %d", easyBest)
	}
	if hardBest != 300 {
		t.Errorf("hard best: want 300, got %d", hardBest)
	}
}

// ---------------------------------------------------------------------------
// ResetStats
// ---------------------------------------------------------------------------

func TestResetStats_ClearsAllDifficulties(t *testing.T) {
	prefs := test.NewApp().Preferences()

	for _, d := range difficultyOrder {
		RecordWin(prefs, d, 100)
	}

	ResetStats(prefs)

	for _, d := range difficultyOrder {
		wins := prefs.IntWithFallback(prefKey(d, "wins"), -1)
		if wins != 0 {
			t.Errorf("difficulty %v: wins should be 0 after reset, got %d", d, wins)
		}
		best := prefs.IntWithFallback(prefKey(d, "best"), -1)
		if best != math.MaxInt32 {
			t.Errorf("difficulty %v: best should be MaxInt32 after reset, got %d", d, best)
		}
	}
}

// ---------------------------------------------------------------------------
// formatBest
// ---------------------------------------------------------------------------

func TestFormatBest_Unset(t *testing.T) {
	got := formatBest(math.MaxInt32)
	if got != "–" {
		t.Errorf("unset best: want –, got %q", got)
	}
}

func TestFormatBest_Zero(t *testing.T) {
	got := formatBest(0)
	if got != "–" {
		t.Errorf("zero best: want –, got %q", got)
	}
}

func TestFormatBest_Minutes(t *testing.T) {
	// 90 seconds = 1:30
	got := formatBest(90)
	if got != "1:30" {
		t.Errorf("90s: want 1:30, got %q", got)
	}
}

func TestFormatBest_ExactMinutes(t *testing.T) {
	// 120 seconds = 2:00
	got := formatBest(120)
	if got != "2:00" {
		t.Errorf("120s: want 2:00, got %q", got)
	}
}

func TestFormatBest_SubMinute(t *testing.T) {
	// 45 seconds = 0:45
	got := formatBest(45)
	if got != "0:45" {
		t.Errorf("45s: want 0:45, got %q", got)
	}
}
