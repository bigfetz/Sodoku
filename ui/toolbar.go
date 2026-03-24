package ui

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"sudoku_game/sudoku"
)

// tallSlot wraps a button in a fixed-height container so the numpad row is
// taller than the default button height, and hidden buttons still occupy space.
func tallSlot(btn fyne.CanvasObject) fyne.CanvasObject {
	const btnHeight = float32(72)
	return container.NewStack(
		// Invisible spacer that enforces the minimum height.
		newMinSizeBox(fyne.NewSize(0, btnHeight)),
		widget.NewLabel(""), // keeps slot visible when button hidden
		btn,
	)
}

// newMinSizeBox returns a canvas object whose MinSize is the given size.
func newMinSizeBox(size fyne.Size) fyne.CanvasObject {
	r := &minSizeRect{size: size}
	r.ExtendBaseWidget(r)
	return r
}

type minSizeRect struct {
	widget.BaseWidget
	size fyne.Size
}

func (m *minSizeRect) CreateRenderer() fyne.WidgetRenderer {
	r := canvas.NewRectangle(color.Transparent)
	return widget.NewSimpleRenderer(r)
}

func (m *minSizeRect) MinSize() fyne.Size { return m.size }

// buildNumpad returns the numpad widget and a refresh function that hides
// buttons for digits that are fully used up (all 9 placed on the board).
// Exhausted digit slots stay blank — the grid never collapses or reflows.
func buildNumpad(bw *BoardWidget) (fyne.CanvasObject, func()) {
	// Keep references to digit buttons so we can show/hide them.
	digitBtns := make([]*widget.Button, 10) // index 1-9 used; 0 unused

	cells := make([]fyne.CanvasObject, 0, 10)
	for d := 1; d <= 9; d++ {
		d := d // capture
		btn := widget.NewButton(fmt.Sprintf("%d", d), func() {
			bw.PlaceDigit(d)
		})
		digitBtns[d] = btn
		cells = append(cells, tallSlot(btn))
	}
	clearBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		bw.ClearSelected()
	})
	cells = append(cells, tallSlot(clearBtn))

	grid := container.New(layout.NewGridLayoutWithColumns(5), cells...)

	// refresh updates digit buttons (count-exhausted) and the clear button
	// (only visible when a cell is selected).
	refresh := func() {
		counts := bw.DigitCounts()
		for d := 1; d <= 9; d++ {
			if counts[d] >= sudoku.BoardSize {
				digitBtns[d].Hide()
			} else {
				digitBtns[d].Show()
			}
		}
		if bw.selRow >= 0 && !bw.solving {
			clearBtn.Show()
		} else {
			clearBtn.Hide()
		}
	}
	// Apply initial state.
	refresh()

	// Also refresh when the selection changes (tap or new game).
	bw.OnSelectionChanged = refresh

	return grid, refresh
}
