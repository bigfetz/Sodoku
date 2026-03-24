package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"image/color"

	"sudoku_game/sudoku"
)

// Run creates the Fyne application, builds the main window, and blocks until
// the window is closed.
func Run() {
	a := app.NewWithID("Fetzco.com-matthewfetzer-sudoku")

	w := a.NewWindow("Sudoku")

	// ---- Session restore or fresh start ------------------------------------
	board := sudoku.NewBoard()
	startElapsed := 0
	startHints := 3
	startDifficulty := sudoku.Easy
	startMistakes := 0

	if sess, ok := LoadSession(a.Preferences()); ok {
		// Restore the puzzle givens, then layer player moves on top.
		_ = board.SetBoard(sess.puzzle)
		board.RestorePlayerBoard(sess.player)
		board.RestoreUndoStack(sess.undoStack)
		startElapsed = sess.elapsed
		startHints = sess.hints
		startDifficulty = sess.difficulty
		startMistakes = sess.mistakes
		// Restore the solution so culprit colouring works immediately on resume.
		if sess.solution != ([sudoku.BoardSize][sudoku.BoardSize]int{}) {
			board.SetSolutionCache(sess.solution)
		}
	} else {
		_ = board.SetBoard(sudoku.GenerateBoard(sudoku.Easy))
	}

	bw := NewBoardWidget(board)
	bw.conflict = board.Conflicts()
	bw.errorCount = startMistakes // restore mistake count before any callbacks wire up

	// currentDifficulty tracks the active puzzle for stats recording.
	currentDifficulty := startDifficulty

	// saveCurrent is declared here so closures in the toolbar can reference it
	// before it is fully assigned after the timer is created.
	var saveCurrent func()

	// ---- Toolbar -----------------------------------------------------------

	newBtn := widget.NewButton("New Game", func() {
		showNewGameDialog(w, bw, &currentDifficulty)
	})
	newBtn.Importance = widget.HighImportance

	// Undo button — disabled until there is something to undo.
	undoBtn := widget.NewButtonWithIcon("", theme.ContentUndoIcon(), func() {
		bw.UndoLast()
	})
	undoBtn.Disable()

	// Hint button — restore remaining hints from session if resuming.
	hintsLeft := startHints
	var hintBtn *widget.Button
	hintBtnLabel := func() string {
		if hintsLeft > 0 {
			return fmt.Sprintf("Hint (%d)", hintsLeft)
		}
		return "No Hints"
	}
	hintBtn = widget.NewButton(hintBtnLabel(), func() {})
	if hintsLeft == 0 {
		hintBtn.Disable()
	}
	hintBtn.OnTapped = func() {
		if hintsLeft <= 0 {
			return
		}
		if bw.ApplyHint() {
			hintsLeft--
			hintBtn.SetText(hintBtnLabel())
			if hintsLeft == 0 {
				hintBtn.Disable()
			}
			// Save immediately after hint so hintsLeft is correct.
			saveCurrent()
		}
	}

	// Dark-mode toggle — initialised from the OS/user preference.
	// onThemeChange is assigned below once the stats labels exist.
	darkMode := a.Settings().ThemeVariant() == theme.VariantDark
	var onThemeChange func()

	// ⋯ overflow menu — Stats, Theme toggle, Solve.
	var moreBtn *widget.Button
	moreBtn = widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		themeLbl := "☀️  Light Mode"
		if !darkMode {
			themeLbl = "🌙  Dark Mode"
		}
		items := []*fyne.MenuItem{
			fyne.NewMenuItem("📊  Stats", func() {
				ShowStatsDialog(w, a.Preferences())
			}),
			fyne.NewMenuItem(themeLbl, func() {
				if onThemeChange != nil {
					onThemeChange()
				}
			}),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("🔍  Solve Puzzle", func() {
				showSolveConfirm(w, bw)
			}),
		}
		pop := widget.NewPopUpMenu(fyne.NewMenu("", items...), w.Canvas())
		btnPos := fyne.CurrentApp().Driver().AbsolutePositionForObject(moreBtn)
		btnSize := moreBtn.Size()
		pop.ShowAtPosition(fyne.NewPos(btnPos.X+btnSize.Width-pop.MinSize().Width, btnPos.Y+btnSize.Height))
	})

	// tallToolbarSlot wraps a button in a fixed-height container so all toolbar
	// buttons have the same taller touch target on iOS.
	const toolbarBtnHeight = float32(52)
	tallToolbarSlot := func(obj fyne.CanvasObject) fyne.CanvasObject {
		return container.NewStack(
			newMinSizeBox(fyne.NewSize(0, toolbarBtnHeight)),
			obj,
		)
	}
	wideToolbarSlot := func(obj fyne.CanvasObject) fyne.CanvasObject {
		return container.NewStack(
			newMinSizeBox(fyne.NewSize(58, toolbarBtnHeight)),
			obj,
		)
	}

	// Set colTitle before creating the label so the initial colour is correct.
	if darkMode {
		colTitle = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
	} else {
		colTitle = color.NRGBA{R: 0x22, G: 0x22, B: 0x22, A: 0xFF}
	}

	titleLabel := canvas.NewText("Sudoku", colTitle)
	titleLabel.TextSize = 22
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	titleLabel.Alignment = fyne.TextAlignCenter

	// Right-side button group pinned to the right; title fills the rest and centers.
	rightBtns := container.NewHBox(
		wideToolbarSlot(undoBtn),
		tallToolbarSlot(hintBtn),
		tallToolbarSlot(newBtn),
		wideToolbarSlot(moreBtn),
	)
	toolbar := container.NewBorder(nil, nil, nil, rightBtns, titleLabel)

	// ---- On-screen number pad (essential for touch/iOS) --------------------
	numpad, refreshNumpad := buildNumpad(bw)
	bw.OnDigitCountsChanged = refreshNumpad

	// ---- Error counter -----------------------------------------------------
	// Initialise text from restored session (may be non-zero on resume).
	errorLabelText := fmt.Sprintf("Mistakes: %d", startMistakes)
	errorLabelCol := colErrorLabel
	if startMistakes > 0 {
		errorLabelCol = colErrorLabelActive
	}
	errorLabel := canvas.NewText(errorLabelText, errorLabelCol)
	errorLabel.TextSize = 16
	errorLabel.TextStyle = fyne.TextStyle{Bold: true}
	errorLabel.Alignment = fyne.TextAlignCenter
	bw.OnErrorCountChanged = func(n int) {
		if n == 0 {
			errorLabel.Text = "Mistakes: 0"
			errorLabel.Color = colErrorLabel
		} else {
			errorLabel.Text = fmt.Sprintf("Mistakes: %d", n)
			errorLabel.Color = colErrorLabelActive
		}
		errorLabel.Refresh()
	}

	// Restore undo button state if the stack was loaded from session.
	if board.CanUndo() {
		undoBtn.Enable()
	}

	// ---- Timer -------------------------------------------------------------
	// Initialise the label to the restored elapsed time immediately so there
	// is no 0:00 flash before the first ticker tick fires.
	startElapsedDur := time.Duration(startElapsed) * time.Second
	initialTimerText := fmt.Sprintf("%d:%02d", int(startElapsedDur.Minutes()), int(startElapsedDur.Seconds())%60)
	timerLabel := canvas.NewText(initialTimerText, colErrorLabel)
	timerLabel.TextSize = 16
	timerLabel.TextStyle = fyne.TextStyle{Bold: true}
	timerLabel.Alignment = fyne.TextAlignCenter

	// Now that both stat labels exist, wire up the theme toggle callback.
	onThemeChange = func() {
		darkMode = !darkMode
		if darkMode {
			a.Settings().SetTheme(forcedTheme(theme.VariantDark))
		} else {
			a.Settings().SetTheme(forcedTheme(theme.VariantLight))
		}
		SetDarkMode(darkMode)
		setStatColours(darkMode)
		titleLabel.Color = colTitle
		titleLabel.Refresh()
		errorLabel.Color = colErrorLabel
		errorLabel.Refresh()
		timerLabel.Color = colErrorLabel
		timerLabel.Refresh()
		bw.Refresh()
	}

	gt := &gameTimer{label: timerLabel}
	if startElapsed > 0 {
		gt.startWithOffset(startElapsedDur)
	} else {
		gt.start()
	}

	// Sync colours/widget state to the detected startup theme.
	SetDarkMode(darkMode)
	setStatColours(darkMode)

	// saveCurrent persists the ongoing game state so it survives app restarts.
	// The var is declared earlier so toolbar closures (hint button) can call it.
	saveCurrent = func() {
		gt.mu.Lock()
		var elapsed time.Duration
		if gt.paused || gt.done {
			elapsed = gt.elapsed
		} else {
			elapsed = gt.elapsed + time.Since(gt.startAt)
		}
		gt.mu.Unlock()
		SaveSession(a.Preferences(), board, currentDifficulty, int(elapsed.Seconds()), hintsLeft, bw.errorCount)
	}

	// Save on every player move so iOS SIGKILL can't lose progress.
	bw.OnBoardChanged = saveCurrent

	// Pause/resume the timer when the app backgrounds/foregrounds.
	a.Lifecycle().SetOnExitedForeground(func() {
		gt.pause()
		if !bw.solved {
			saveCurrent()
		}
	})
	a.Lifecycle().SetOnEnteredForeground(func() {
		if !bw.solved {
			gt.resume()
		}
	})

	// Wire undo button enable/disable to board undo stack state.
	bw.OnUndoStateChanged = func(canUndo bool) {
		if canUndo {
			undoBtn.Enable()
		} else {
			undoBtn.Disable()
		}
	}

	// Reset timer, mistakes, hints, and undo stack when a new game starts.
	// Also clear any saved session — the new game will be saved on next background.
	bw.OnNewGame = func() {
		gt.reset()
		hintsLeft = 3
		hintBtn.SetText(hintBtnLabel())
		hintBtn.Enable()
		undoBtn.Disable()
		ClearSession(a.Preferences())
	}

	// Stop timer and record win when puzzle is solved by the player.
	bw.OnSolved = func() {
		gt.stop()
		elapsed := int(gt.elapsed.Seconds())
		RecordWin(a.Preferences(), currentDifficulty, elapsed)
		ClearSession(a.Preferences())
		bw.WaveAllCells()
	}

	// Auto-solve: stop the timer but do not record in stats.
	bw.OnAutoSolved = func() {
		gt.stop()
		ClearSession(a.Preferences())
	}

	// ---- Full layout -------------------------------------------------------
	// Border layout: toolbar top, numpad bottom, board fills the rest.
	statsRow := container.NewHBox(
		layout.NewSpacer(),
		errorLabel,
		widget.NewLabel("  |  "),
		timerLabel,
		layout.NewSpacer(),
	)
	content := container.NewBorder(
		container.NewVBox(toolbar, widget.NewSeparator(), statsRow),
		container.NewVBox(widget.NewSeparator(), numpad),
		nil, nil,
		bw, // board fills centre — Fyne stretches it to all available space
	)

	w.SetContent(content)
	w.Canvas().Focus(bw)

	// Stop the timer and save session when the window is closed (desktop/simulator).
	// On real iOS/Android the app is killed without this intercept, so the
	// background-lifecycle save (SetOnExitedForeground) covers those devices.
	w.SetCloseIntercept(func() {
		if !bw.solved {
			saveCurrent()
		}
		gt.stop()
		w.Close()
	})

	w.ShowAndRun()
}
