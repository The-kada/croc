package drawer

import (
	"fmt"
	"time"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

func Init() {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}

}

const (
	BEGIN = iota
	INTER
	END
)

type coordinate [2]int
type line struct {
	begin    *coordinate
	end      *coordinate
	lineType int
}

type drawStartegy struct {
	name    string
	command func(*DrawTool)
}

type DrawTool struct {
	render               *sdl.Renderer
	lineStack            []line
	completedLines       []line
	completeLineStrategy int
	strategies           []drawStartegy
	crusors              map[string]*sdl.Cursor
}

func (d *DrawTool) cleanup() {
	sdl.Quit()
	window, err := d.render.GetWindow()
	if err != nil {
		fmt.Printf("could not get window! :%v \n", err)
	}

	err = d.render.Destroy()
	if err != nil {
		fmt.Printf("could not destroy render! :%v \n", err)
	}

	err = window.Destroy()
	if err != nil {
		fmt.Printf("could not destroy window! :%v \n", err)
	}

}

func (d *DrawTool) getLastUnfinishedLine() *line {
	if len(d.lineStack) == 0 {
		return nil
	}

	lastline := &d.lineStack[len(d.lineStack)-1]
	if lastline.end != nil {
		return nil
	}
	return &d.lineStack[len(d.lineStack)-1]
}

func (d *DrawTool) completeLastLine() int {
	if len(d.lineStack) == 0 {
		panic("no lines in stack!")
	}
	d.completedLines = append(d.completedLines, d.lineStack[len(d.lineStack)-1])
	d.lineStack = d.lineStack[0 : len(d.lineStack)-1]
	return len(d.completedLines) - 1
}

func (d *DrawTool) eliminateInterLines(endLineIndex int) {
	beginLineIndex := -1
	lastCompletedLine := d.completedLines[endLineIndex]
	for lineIndex := endLineIndex; lineIndex >= 0; lineIndex-- {
		if d.completedLines[lineIndex].lineType == BEGIN {
			beginLineIndex = lineIndex
			break
		}
	}
	if beginLineIndex == -1 {
		panic("end must have begin line")
	}
	d.completedLines[beginLineIndex].end = lastCompletedLine.end
	toTheLeft := d.completedLines[0 : beginLineIndex+1]
	d.completedLines = append(toTheLeft, d.completedLines[endLineIndex:]...)
}

func (d *DrawTool) drawlines() {
	err := d.render.SetDrawColor(255, 255, 255, 0)
	if err != nil {
		panic(err)
	}

	d.render.Clear()
	d.render.SetDrawColor(0, 0, 0, 0)

	for _, lines := range d.completedLines {
		d.render.DrawLine(int32(lines.begin[0]), int32(lines.begin[1]), int32(lines.end[0]), int32(lines.end[1]))
	}
	d.render.Present()

}

func (d *DrawTool) handleMousevents() {
	startTime := time.Now()
	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				d.cleanup()
			case *sdl.KeyboardEvent:
				if t.Keysym.Sym == sdl.K_ESCAPE {
					d.cleanup()
				}
				if t.Keysym.Sym == sdl.K_SPACE && t.Type == sdl.KEYDOWN {
					nextStartegy := (d.completeLineStrategy + 1) % len(d.strategies)
					fmt.Printf("Changing line drawing strategy from %s to %s\n",
						d.strategies[d.completeLineStrategy].name, d.strategies[nextStartegy].name)
					d.completeLineStrategy = nextStartegy
				}
			case *sdl.MouseButtonEvent:
				if t.State == sdl.PRESSED {
					d.lineStack = append(d.lineStack, line{
						begin:    &coordinate{int(t.X), int(t.Y)},
						lineType: BEGIN,
					})
				}

				if t.State == sdl.RELEASED {
					lastLine := d.getLastUnfinishedLine()
					lastLine.end = &coordinate{int(t.X), int(t.Y)}
					d.strategies[d.completeLineStrategy].command(d)
				}

			}

		}

		// Illusion of flow.....
		// 1 frame == 10ms
		// xframe == 1000ms
		// 100 frames in 1000ms = 100FPS

		if time.Since(startTime).Milliseconds() > (10) {
			startTime = time.Now()
			d.makeInterLine()
		}
	}

}

func (d *DrawTool) makeInterLine() {
	lastUnfinishedLine := d.getLastUnfinishedLine()
	if lastUnfinishedLine == nil {
		return
	}
	x, y, _ := sdl.GetMouseState()
	lastUnfinishedLine.end = &coordinate{int(x), int(y)}
	d.completeLastLine()
	d.drawlines()
	d.lineStack = append(d.lineStack, line{
		begin:    &coordinate{int(x), int(y)},
		lineType: INTER,
	})
}

func (d *DrawTool) setCurrentDrawingIcon(icon string) {
	if cursor, ok := d.crusors[icon]; ok {
		sdl.SetCursor(cursor)
	} else {
		panic("impossible cursor must exist!")
	}
}

func Drawer() *DrawTool {

	window, err := sdl.CreateWindow("croc", sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, 800, 600, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}

	// 32x32 penicl image
	// todo: init drawing tools icons
	surface, err := img.Load("assets/pencilnew.png")
	if err != nil {
		panic(err)
	}
	pencilCursor := sdl.CreateColorCursor(surface, 0, 25)
	surface.Free()

	render, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
	if err != nil {
		panic(err)
	}

	drawingTool := &DrawTool{
		render:               render,
		completeLineStrategy: 0,
		strategies: []drawStartegy{
			{
				name: "straight-lines",
				command: func(dt *DrawTool) {
					dt.eliminateInterLines(dt.completeLastLine())
					dt.drawlines()
				},
			},
			{
				name: "wavy-lines",
				command: func(dt *DrawTool) {
					dt.completeLastLine()
					dt.drawlines()
				},
			},
		},
		crusors: map[string]*sdl.Cursor{
			"pencil": pencilCursor,
		},
	}

	drawingTool.setCurrentDrawingIcon("pencil")
	drawingTool.drawlines()
	drawingTool.handleMousevents()
	return drawingTool

}
