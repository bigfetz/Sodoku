package ui

import (
	"fmt"
	"image/color"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
)

var (
	colErrorLabel       = color.NRGBA{R: 0x88, G: 0x88, B: 0x88, A: 0xFF} // grey when zero
	colErrorLabelActive = color.NRGBA{R: 0xC0, G: 0x20, B: 0x20, A: 0xFF} // red when errors > 0
	colTitle            color.NRGBA                                       // set by setStatColours before first use
)

// setStatColours updates the stats-row colour variables for the current mode.
func setStatColours(dark bool) {
	if dark {
		colErrorLabel = color.NRGBA{R: 0x99, G: 0x99, B: 0x99, A: 0xFF}
		colErrorLabelActive = color.NRGBA{R: 0xFF, G: 0x60, B: 0x60, A: 0xFF}
		colTitle = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
	} else {
		colErrorLabel = color.NRGBA{R: 0x88, G: 0x88, B: 0x88, A: 0xFF}
		colErrorLabelActive = color.NRGBA{R: 0xC0, G: 0x20, B: 0x20, A: 0xFF}
		colTitle = color.NRGBA{R: 0x22, G: 0x22, B: 0x22, A: 0xFF}
	}
}

// variantTheme wraps DefaultTheme and forces a specific light/dark variant,
// avoiding the deprecated theme.LightTheme() / theme.DarkTheme() functions.
type variantTheme struct {
	fyne.Theme
	variant fyne.ThemeVariant
}

func (t variantTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return t.Theme.Color(name, t.variant)
}

func forcedTheme(variant fyne.ThemeVariant) fyne.Theme {
	return variantTheme{Theme: theme.DefaultTheme(), variant: variant}
}

// gameTimer drives a running mm:ss clock displayed in the stats row.
// It supports pause/resume so the clock freezes when the app backgrounds.
type gameTimer struct {
	label   *canvas.Text
	elapsed time.Duration // accumulated time before the current run
	startAt time.Time     // when the current run started (zero if paused/stopped)
	stopCh  chan struct{}
	paused  bool
	done    bool // true after stop() — never resumes
	mu      sync.Mutex
}

func (gt *gameTimer) start() {
	gt.startWithOffset(0)
}

// startWithOffset starts the timer with a pre-existing elapsed duration,
// used when restoring a saved session.
func (gt *gameTimer) startWithOffset(offset time.Duration) {
	gt.mu.Lock()
	gt.elapsed = offset
	gt.done = false
	gt.paused = false
	gt.startAt = time.Now()
	gt.stopCh = make(chan struct{})
	gt.mu.Unlock()
	gt.runTicker()
}

func (gt *gameTimer) runTicker() {
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				gt.mu.Lock()
				total := gt.elapsed + time.Since(gt.startAt)
				gt.mu.Unlock()
				m := int(total.Minutes())
				s := int(total.Seconds()) % 60
				text := fmt.Sprintf("%d:%02d", m, s)
				fyne.Do(func() {
					gt.label.Text = text
					gt.label.Refresh()
				})
			case <-gt.stopCh:
				return
			}
		}
	}()
}

func (gt *gameTimer) pause() {
	gt.mu.Lock()
	defer gt.mu.Unlock()
	if gt.paused || gt.done {
		return
	}
	gt.paused = true
	gt.elapsed += time.Since(gt.startAt)
	close(gt.stopCh)
}

func (gt *gameTimer) resume() {
	gt.mu.Lock()
	defer gt.mu.Unlock()
	if !gt.paused || gt.done {
		return
	}
	gt.paused = false
	gt.startAt = time.Now()
	gt.stopCh = make(chan struct{})
	gt.runTicker()
}

func (gt *gameTimer) reset() {
	gt.mu.Lock()
	if !gt.paused && !gt.done {
		close(gt.stopCh)
	}
	gt.done = false
	gt.paused = false
	gt.elapsed = 0
	gt.startAt = time.Now()
	gt.stopCh = make(chan struct{})
	gt.mu.Unlock()
	fyne.Do(func() {
		gt.label.Text = "0:00"
		gt.label.Refresh()
	})
	gt.runTicker()
}

func (gt *gameTimer) stop() {
	gt.mu.Lock()
	defer gt.mu.Unlock()
	if gt.done {
		return
	}
	gt.done = true
	if !gt.paused {
		gt.elapsed += time.Since(gt.startAt) // capture remaining run before stopping
		close(gt.stopCh)
	}
}
