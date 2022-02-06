package main

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/gonutz/sudoku"
	"github.com/gonutz/w32/v2"
	"github.com/gonutz/wui/v2"
)

var (
	tileSize         = 90
	thinBorderSize   = 3
	thickBorderSize  = 3 * thinBorderSize
	boardSize        = 4*thickBorderSize + 6*thinBorderSize + 9*tileSize
	mediumFontHeight = tileSize / 2

	backColor      = wui.RGB(64, 64, 64)
	hotColor       = wui.RGB(64, 64, 192)
	borderColor    = wui.RGB(192, 192, 192)
	textColor      = wui.RGB(255, 255, 255)
	fixedColor     = wui.RGB(192, 192, 255)
	highlightColor = wui.RGB(255, 192, 192)
	gameMode       = false
)

func main() {
	largeFont, _ := wui.NewFont(wui.FontDesc{
		Name:   "Tahoma",
		Height: tileSize - tileSize/10,
	})

	mediumFont, _ := wui.NewFont(wui.FontDesc{
		Name:   "Tahoma",
		Height: mediumFontHeight,
	})

	smallFont, _ := wui.NewFont(wui.FontDesc{
		Name:   "Tahoma",
		Height: tileSize / 4,
	})

	icon, _ := wui.NewIconFromExeResource(10)

	window := wui.NewWindow()
	window.SetTitle("Soduko")
	window.SetIcon(icon)
	window.SetInnerSize(boardSize, boardSize)
	window.SetResizable(false)
	window.SetHasMaxButton(false)

	var b board

	wantHighlight := func(n int) bool {
		ok := false
		for y := 0; y < 9; y++ {
			for x := 0; x < 9; x++ {
				if b[x][y].hot {
					ok = true
					if b[x][y].number != n {
						return false
					}
				}
			}
		}
		return ok
	}

	board := wui.NewPaintBox()
	window.Add(board)
	board.SetBounds(0, 0, window.InnerWidth(), window.InnerHeight())
	board.SetOnPaint(func(canvas *wui.Canvas) {
		if gameMode {
			canvas.FillRect(0, 0, boardSize, boardSize, borderColor)
			for row := 0; row < 9; row++ {
				for col := 0; col < 9; col++ {
					f := b[col][row]
					x, y := tileTopLeft(col, row)

					color := backColor
					if f.hot {
						color = hotColor
					}
					canvas.FillRect(x, y, tileSize, tileSize, color)

					if f.number > 0 {
						text := strconv.Itoa(f.number)
						canvas.SetFont(largeFont)
						w, h := canvas.TextExtent(text)
						color := textColor
						if f.fixed {
							color = fixedColor
						}
						if wantHighlight(f.number) {
							color = highlightColor
						}
						canvas.TextOut(x+(tileSize-w)/2, y+(tileSize-h)/2, text, color)
					} else {
						// Draw pencil marks.
						// Draw the corner marks.
						canvas.SetFont(smallFont)
						for i := 0; i < 9; i++ {
							if f.corner[i] {
								text := strconv.Itoa(i + 1)
								w, h := canvas.TextExtent(text)
								bx, by, bw, bh := cornerPencilMarkBounds(i)
								canvas.TextOut(x+bx+(bw-w)/2, y+by+(bh-h)/2, text, textColor)
							}
						}
						// Draw the center marks.
						var centerText string
						for i := 0; i < 9; i++ {
							if f.center[i] {
								centerText += strconv.Itoa(i + 1)
							}
						}
						half := len(centerText) / 2
						if half >= 3 {
							centerText = centerText[:half] + "\n" + centerText[half:]
						}
						canvas.TextRectFormat(x, y, tileSize, tileSize, centerText, wui.FormatCenter, textColor)
					}
				}
			}
		} else {
			canvas.FillRect(0, 0, boardSize, boardSize, backColor)
			canvas.SetFont(mediumFont)
			canvas.TextRectFormat(0, 0, boardSize, boardSize, `
F1 - Help On/Off
F2 - New Game
Ctrl +/- - Zoom In/Out
Enter - Check Solution
Number - Enter Number
Shift+Number - Pencil Mark Corner
Control+Number - Pencil Mark Center
Backspace - Clear Number/Pencil Marks
Mouse/Arrow Keys - Select Cells
Escape - Clear Selection
`, wui.FormatCenter, textColor)
		}
	})
	board.SetAnchors(wui.AnchorMinAndMax, wui.AnchorMinAndMax)

	putNumber := func(n int) func() {
		return func() {
			if !gameMode {
				return
			}
			for y := 0; y < 9; y++ {
				for x := 0; x < 9; x++ {
					if b[x][y].hot && !b[x][y].fixed {
						b[x][y].number = n
					}
				}
			}
			board.Paint()
		}
	}

	putCenterPencilMark := func(n int) func() {
		return func() {
			if !gameMode {
				return
			}

			var setMark bool

			for y := 0; y < 9; y++ {
				for x := 0; x < 9; x++ {
					if b[x][y].hot && b[x][y].number == 0 && !b[x][y].center[n-1] {
						setMark = true
					}
				}
			}

			for y := 0; y < 9; y++ {
				for x := 0; x < 9; x++ {
					if b[x][y].hot && b[x][y].number == 0 {
						b[x][y].center[n-1] = setMark
					}
				}
			}

			board.Paint()
		}
	}

	putCornerPencilMark := func(n int) func() {
		return func() {
			if !gameMode {
				return
			}

			var setMark bool

			for y := 0; y < 9; y++ {
				for x := 0; x < 9; x++ {
					if b[x][y].hot && b[x][y].number == 0 && !b[x][y].corner[n-1] {
						setMark = true
					}
				}
			}

			for y := 0; y < 9; y++ {
				for x := 0; x < 9; x++ {
					if b[x][y].hot && b[x][y].number == 0 {
						b[x][y].corner[n-1] = setMark
					}
				}
			}

			board.Paint()
		}
	}

	clearFields := func() {
		if !gameMode {
			return
		}

		var hasNumber, hasCenter bool
		for y := 0; y < 9; y++ {
			for x := 0; x < 9; x++ {
				if b[x][y].hot && !b[x][y].fixed {
					hasNumber = hasNumber || b[x][y].number != 0
					for i := range b[x][y].center {
						hasCenter = hasCenter || b[x][y].center[i]
					}
				}
			}
		}

		if hasNumber {
			// Delete number.
			for y := 0; y < 9; y++ {
				for x := 0; x < 9; x++ {
					if b[x][y].hot && !b[x][y].fixed {
						b[x][y].number = 0
					}
				}
			}
		} else if hasCenter {
			// Delete center pencil mark.
			for y := 0; y < 9; y++ {
				for x := 0; x < 9; x++ {
					if b[x][y].hot && !b[x][y].fixed {
						for i := range b[x][y].center {
							b[x][y].center[i] = false
						}
					}
				}
			}
		} else {
			// Delete corner pencil mark.
			for y := 0; y < 9; y++ {
				for x := 0; x < 9; x++ {
					if b[x][y].hot && !b[x][y].fixed {
						for i := range b[x][y].corner {
							b[x][y].corner[i] = false
						}
					}
				}
			}
		}

		board.Paint()
	}

	clearCorners := func() {
		if !gameMode {
			return
		}

		for y := 0; y < 9; y++ {
			for x := 0; x < 9; x++ {
				if b[x][y].hot && !b[x][y].fixed && b[x][y].number == 0 {
					for i := range b[x][y].corner {
						b[x][y].corner[i] = false
					}
				}
			}
		}
		board.Paint()
	}

	clearCenter := func() {
		if !gameMode {
			return
		}

		for y := 0; y < 9; y++ {
			for x := 0; x < 9; x++ {
				if b[x][y].hot && !b[x][y].fixed && b[x][y].number == 0 {
					for i := range b[x][y].center {
						b[x][y].center[i] = false
					}
				}
			}
		}
		board.Paint()
	}

	var lastSelection [2]int
	moveSelection := func(dx, dy int) {
		if !gameMode {
			return
		}

		s := lastSelection
		if s[0] != -1 {
			for y := 0; y < 9; y++ {
				for x := 0; x < 9; x++ {
					b[x][y].hot = false
				}
			}
			x := ((s[0] + dx) + 9) % 9
			y := ((s[1] + dy) + 9) % 9
			b[x][y].hot = true
			lastSelection = [2]int{x, y}
		}
		board.Paint()
	}

	expandSelection := func(dx, dy int) func() {
		return func() {
			if !gameMode {
				return
			}

			s := lastSelection
			if s[0] != -1 {
				x := ((s[0] + dx) + 9) % 9
				y := ((s[1] + dy) + 9) % 9
				b[x][y].hot = true
				lastSelection = [2]int{x, y}
			}
			board.Paint()
		}
	}

	selectAll := func() {
		if !gameMode {
			return
		}

		for y := 0; y < 9; y++ {
			for x := 0; x < 9; x++ {
				b[x][y].hot = true
			}
		}
		board.Paint()
	}

	unselectAll := func() {
		if !gameMode {
			return
		}

		for y := 0; y < 9; y++ {
			for x := 0; x < 9; x++ {
				b[x][y].hot = false
			}
		}
		board.Paint()
	}

	solution := sudoku.Game{-1}
	givenDigits := 30
	newGame := func() {
		dlg := wui.NewWindow()
		dlg.SetFont(mediumFont)
		dlg.SetInnerSize(9*tileSize, 3*mediumFontHeight)
		dlg.SetHasBorder(false)
		dlg.SetResizable(false)
		dlg.SetPosition(
			window.X()+(window.Width()-dlg.Width())/2,
			window.Y()+(window.Height()-dlg.Height())/2,
		)

		left := wui.NewLabel()
		dlg.Add(left)
		left.SetBounds(0, mediumFontHeight, 4*tileSize, mediumFontHeight)
		left.SetAlignment(wui.AlignRight)
		left.SetText("Give me at least ")

		digits := wui.NewIntUpDown()
		dlg.Add(digits)
		digits.SetBounds(4*tileSize, mediumFontHeight, 2*tileSize, mediumFontHeight+mediumFontHeight/8)
		digits.SetMinMax(17, 80)
		digits.SetValue(givenDigits)

		right := wui.NewLabel()
		dlg.Add(right)
		right.SetBounds(6*tileSize, mediumFontHeight, 3*tileSize, mediumFontHeight)
		right.SetText(" numbers.")

		dlg.SetOnShow(func() {
			digits.Focus()
			digits.SelectAll()
		})

		var wantNewGame bool
		ok := func() {
			givenDigits = digits.Value()
			dlg.Close()
			wantNewGame = true
		}
		cancel := func() {
			dlg.Close()
		}
		dlg.SetShortcut(ok, wui.KeyReturn)
		dlg.SetShortcut(cancel, wui.KeyEscape)

		dlg.ShowModal()
		dlg.Destroy()

		if !wantNewGame {
			return
		}

		lastSelection = [2]int{}
		var start sudoku.Game
		solution, start = generateNewGame(givenDigits)
		for y := 0; y < 9; y++ {
			for x := 0; x < 9; x++ {
				b[x][y].number = start[x+9*y]
				for i := range b[x][y].center {
					b[x][y].center[i] = false
				}
				for i := range b[x][y].corner {
					b[x][y].corner[i] = false
				}
				b[x][y].hot = false
				b[x][y].fixed = b[x][y].number != 0
			}
		}
		gameMode = true
		board.Paint()
	}

	checkGame := func() {
		if !gameMode {
			return
		}
		var have sudoku.Game
		for y := 0; y < 9; y++ {
			for x := 0; x < 9; x++ {
				have[x+y*9] = b[x][y].number
			}
		}
		if have == solution {
			wui.MessageBoxInfo("You Win!", "This is correct.")
		} else {
			wui.MessageBoxError("Not Yet!", "Your answer is wrong.")
		}
	}

	toggleHelp := func() {
		gameMode = !gameMode
		board.Paint()
	}

	zoom := func(delta int) {
		tileSize += delta
		if tileSize < 30 {
			tileSize = 30
		}

		thinBorderSize = 3
		thickBorderSize = 3 * thinBorderSize
		boardSize = 4*thickBorderSize + 6*thinBorderSize + 9*tileSize
		mediumFontHeight = tileSize / 2
		window.SetInnerSize(boardSize, boardSize)

		largeFont, _ = wui.NewFont(wui.FontDesc{
			Name:   "Tahoma",
			Height: tileSize - tileSize/10,
		})

		mediumFont, _ = wui.NewFont(wui.FontDesc{
			Name:   "Tahoma",
			Height: mediumFontHeight,
		})

		smallFont, _ = wui.NewFont(wui.FontDesc{
			Name:   "Tahoma",
			Height: tileSize / 4,
		})
	}
	zoomIn := func() { zoom(1) }
	zoomOut := func() { zoom(-1) }

	window.SetShortcut(putNumber(1), wui.Key1)
	window.SetShortcut(putNumber(2), wui.Key2)
	window.SetShortcut(putNumber(3), wui.Key3)
	window.SetShortcut(putNumber(4), wui.Key4)
	window.SetShortcut(putNumber(5), wui.Key5)
	window.SetShortcut(putNumber(6), wui.Key6)
	window.SetShortcut(putNumber(7), wui.Key7)
	window.SetShortcut(putNumber(8), wui.Key8)
	window.SetShortcut(putNumber(9), wui.Key9)
	window.SetShortcut(putNumber(1), wui.KeyNum1)
	window.SetShortcut(putNumber(2), wui.KeyNum2)
	window.SetShortcut(putNumber(3), wui.KeyNum3)
	window.SetShortcut(putNumber(4), wui.KeyNum4)
	window.SetShortcut(putNumber(5), wui.KeyNum5)
	window.SetShortcut(putNumber(6), wui.KeyNum6)
	window.SetShortcut(putNumber(7), wui.KeyNum7)
	window.SetShortcut(putNumber(8), wui.KeyNum8)
	window.SetShortcut(putNumber(9), wui.KeyNum9)
	window.SetShortcut(putCenterPencilMark(1), wui.KeyControl, wui.Key1)
	window.SetShortcut(putCenterPencilMark(2), wui.KeyControl, wui.Key2)
	window.SetShortcut(putCenterPencilMark(3), wui.KeyControl, wui.Key3)
	window.SetShortcut(putCenterPencilMark(4), wui.KeyControl, wui.Key4)
	window.SetShortcut(putCenterPencilMark(5), wui.KeyControl, wui.Key5)
	window.SetShortcut(putCenterPencilMark(6), wui.KeyControl, wui.Key6)
	window.SetShortcut(putCenterPencilMark(7), wui.KeyControl, wui.Key7)
	window.SetShortcut(putCenterPencilMark(8), wui.KeyControl, wui.Key8)
	window.SetShortcut(putCenterPencilMark(9), wui.KeyControl, wui.Key9)
	window.SetShortcut(putCenterPencilMark(1), wui.KeyControl, wui.KeyNum1)
	window.SetShortcut(putCenterPencilMark(2), wui.KeyControl, wui.KeyNum2)
	window.SetShortcut(putCenterPencilMark(3), wui.KeyControl, wui.KeyNum3)
	window.SetShortcut(putCenterPencilMark(4), wui.KeyControl, wui.KeyNum4)
	window.SetShortcut(putCenterPencilMark(5), wui.KeyControl, wui.KeyNum5)
	window.SetShortcut(putCenterPencilMark(6), wui.KeyControl, wui.KeyNum6)
	window.SetShortcut(putCenterPencilMark(7), wui.KeyControl, wui.KeyNum7)
	window.SetShortcut(putCenterPencilMark(8), wui.KeyControl, wui.KeyNum8)
	window.SetShortcut(putCenterPencilMark(9), wui.KeyControl, wui.KeyNum9)
	window.SetShortcut(putCornerPencilMark(1), wui.KeyShift, wui.Key1)
	window.SetShortcut(putCornerPencilMark(2), wui.KeyShift, wui.Key2)
	window.SetShortcut(putCornerPencilMark(3), wui.KeyShift, wui.Key3)
	window.SetShortcut(putCornerPencilMark(4), wui.KeyShift, wui.Key4)
	window.SetShortcut(putCornerPencilMark(5), wui.KeyShift, wui.Key5)
	window.SetShortcut(putCornerPencilMark(6), wui.KeyShift, wui.Key6)
	window.SetShortcut(putCornerPencilMark(7), wui.KeyShift, wui.Key7)
	window.SetShortcut(putCornerPencilMark(8), wui.KeyShift, wui.Key8)
	window.SetShortcut(putCornerPencilMark(9), wui.KeyShift, wui.Key9)
	window.SetShortcut(clearFields, wui.KeyBack)
	window.SetShortcut(clearFields, wui.Key0)
	window.SetShortcut(clearFields, wui.KeyNum0)
	window.SetShortcut(clearCorners, wui.KeyBack, wui.KeyShift)
	window.SetShortcut(clearCenter, wui.KeyBack, wui.KeyControl)
	window.SetShortcut(expandSelection(1, 0), wui.KeyRight, wui.KeyShift)
	window.SetShortcut(expandSelection(-1, 0), wui.KeyLeft, wui.KeyShift)
	window.SetShortcut(expandSelection(0, 1), wui.KeyDown, wui.KeyShift)
	window.SetShortcut(expandSelection(0, -1), wui.KeyUp, wui.KeyShift)
	window.SetShortcut(expandSelection(1, 0), wui.KeyRight, wui.KeyControl)
	window.SetShortcut(expandSelection(-1, 0), wui.KeyLeft, wui.KeyControl)
	window.SetShortcut(expandSelection(0, 1), wui.KeyDown, wui.KeyControl)
	window.SetShortcut(expandSelection(0, -1), wui.KeyUp, wui.KeyControl)
	window.SetShortcut(selectAll, wui.KeyA, wui.KeyControl)
	window.SetShortcut(unselectAll, wui.KeyEscape)
	window.SetShortcut(newGame, wui.KeyF2)
	window.SetShortcut(toggleHelp, wui.KeyF1)
	window.SetShortcut(checkGame, wui.KeyReturn)
	window.SetShortcut(zoomIn, wui.KeyControl, wui.KeyAdd)
	window.SetShortcut(zoomIn, wui.KeyControl, wui.KeyOEMPlus)
	window.SetShortcut(zoomOut, wui.KeyControl, wui.KeySubtract)
	window.SetShortcut(zoomOut, wui.KeyControl, wui.KeyOEMMinus)

	var (
		selecting    bool
		setSelection bool
	)
	window.SetOnMouseDown(func(button wui.MouseButton, x, y int) {
		if button == wui.MouseButtonLeft {
			shift := w32.GetKeyState(w32.VK_SHIFT)&0x80 != 0
			control := w32.GetKeyState(w32.VK_CONTROL)&0x80 != 0
			toggle := shift || control
			col, row := screenToBoard(x, y)
			lastSelection = [2]int{col, row}
			if toggle {
				b[col][row].hot = !b[col][row].hot
				setSelection = b[col][row].hot
			} else {
				for y := 0; y < 9; y++ {
					for x := 0; x < 9; x++ {
						b[x][y].hot = false
					}
				}
				b[col][row].hot = true
				setSelection = true
			}
			selecting = true
			board.Paint()
		}
	})
	window.SetOnMouseUp(func(button wui.MouseButton, x, y int) {
		if button == wui.MouseButtonLeft {
			selecting = false
		}
	})
	window.SetOnMouseMove(func(x, y int) {
		if selecting {
			col, row := screenToBoard(x, y)
			lastSelection = [2]int{col, row}
			if b[col][row].hot != setSelection {
				b[col][row].hot = setSelection
				board.Paint()
			}
		}
	})

	// Handle Shift+Numpad keys.
	window.SetOnMessage(func(window uintptr, msg uint32, w, l uintptr) (handled bool, result uintptr) {
		if msg == w32.WM_KEYDOWN {
			extended := l&(1<<24) != 0
			if extended {
				// This might have been a regular arrow key.
				switch w {
				case w32.VK_LEFT:
					moveSelection(-1, 0)
				case w32.VK_RIGHT:
					moveSelection(1, 0)
				case w32.VK_DOWN:
					moveSelection(0, 1)
				case w32.VK_UP:
					moveSelection(0, -1)
				}
			} else {
				n := 0
				switch w {
				case w32.VK_LEFT:
					n = 4
				case w32.VK_RIGHT:
					n = 6
				case w32.VK_UP:
					n = 8
				case w32.VK_DOWN:
					n = 2
				case w32.VK_HOME:
					n = 7
				case w32.VK_END:
					n = 1
				case w32.VK_PRIOR:
					n = 9
				case w32.VK_NEXT:
					n = 3
				case w32.VK_CLEAR:
					n = 5
				}
				if n != 0 {
					handled = true
					putCornerPencilMark(n)()
					board.Paint()
				}
			}
		}
		return
	})

	window.Show()
}

func tileTopLeft(col, row int) (x, y int) {
	x = (1+col/3)*(thickBorderSize-thinBorderSize) + col*(thinBorderSize+tileSize)
	y = (1+row/3)*(thickBorderSize-thinBorderSize) + row*(thinBorderSize+tileSize)
	return
}

func cornerPencilMarkBounds(i int) (x, y, w, h int) {
	switch i {
	case 0, 1, 2, 3, 4, 5, 6, 8:
		w = tileSize / 4
	case 7:
		w = tileSize / 3
	}

	switch i {
	case 0, 1, 2, 3:
		x = i * w
	case 4, 6:
		x = 0
	case 5, 8:
		x = 3 * w
	case 7:
		x = w
	}

	h = tileSize / 3
	switch i {
	case 0, 1, 2, 3:
		y = 0
	case 4, 5:
		y = h
	case 6, 7, 8:
		y = 2 * h
	}

	return
}

func screenToBoard(x, y int) (col, row int) {
	{
		best := 9999999
		for c := 0; c < 9; c++ {
			cx, _ := tileTopLeft(c, 0)
			cx += tileSize / 2
			if abs(x-cx) < best {
				col = c
				best = abs(x - cx)
			}
		}
	}

	{
		best := 9999999
		for r := 0; r < 9; r++ {
			_, cy := tileTopLeft(0, r)
			cy += tileSize / 2
			if abs(y-cy) < best {
				row = r
				best = abs(y - cy)
			}
		}
	}

	return
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

type board [9][9]field

type field struct {
	number int
	corner [9]bool
	center [9]bool
	hot    bool
	fixed  bool
}

func generateNewGame(givenDigits int) (solution, start sudoku.Game) {
	rand.Seed(time.Now().UnixNano())

	n := [9]int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	for i := range n {
		j := i + rand.Intn(9-i)
		n[i], n[j] = n[j], n[i]
	}

	game := sudoku.Game{
		n[0], n[1], n[2], n[3], n[4], n[5], n[6], n[7], n[8],
		n[6], n[7], n[8], n[0], n[1], n[2], n[3], n[4], n[5],
		n[3], n[4], n[5], n[6], n[7], n[8], n[0], n[1], n[2],
		n[8], n[0], n[1], n[2], n[3], n[4], n[5], n[6], n[7],
		n[5], n[6], n[7], n[8], n[0], n[1], n[2], n[3], n[4],
		n[2], n[3], n[4], n[5], n[6], n[7], n[8], n[0], n[1],
		n[7], n[8], n[0], n[1], n[2], n[3], n[4], n[5], n[6],
		n[4], n[5], n[6], n[7], n[8], n[0], n[1], n[2], n[3],
		n[1], n[2], n[3], n[4], n[5], n[6], n[7], n[8], n[0],
	}

	for i := 0; i < 1000; i++ {
		a := rand.Intn(3) * 3
		b := rand.Intn(3) + a
		if rand.Intn(2) == 0 {
			swapLines(&game, a, b)
		} else {
			swapCols(&game, a, b)
		}
	}

	solution = game

	have := 81
	want := givenDigits

	rest := make([]int, 81)
	for i := range rest {
		rest[i] = i
	}

	for len(rest) > 0 && have != want {
		n := rand.Intn(len(rest))
		i := rest[n]
		was := game[i]
		game[i] = 0
		if sudoku.HasUniqueSolution(game) {
			have--
		} else {
			game[i] = was
		}
		rest[0], rest[n] = rest[n], rest[0]
		rest = rest[1:]
	}

	start = game

	return
}

func swapLines(g *sudoku.Game, a, b int) {
	if a != b {
		aa := a * 9
		bb := b * 9
		for i := 0; i < 9; i++ {
			g[aa+i], g[bb+i] = g[bb+i], g[aa+i]
		}
	}
}

func swapCols(g *sudoku.Game, a, b int) {
	if a != b {
		for i := 0; i < 9; i++ {
			g[a+i*9], g[b+i*9] = g[b+i*9], g[a+i*9]
		}
	}
}
