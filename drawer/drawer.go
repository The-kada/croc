package drawer

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

func Init() {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}

}

type coordinate [2]int
type line struct {
	begin coordinate
	end   coordinate
}

type DrawTool struct {
	render         *sdl.Renderer
	lineStack      []line
	completedLines []line
	blockChan      chan struct{}
}

func (d *DrawTool) WaitForCanvasClose() {
	<-d.blockChan
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

	close(d.blockChan)

}

func (d *DrawTool) getLastLine() *line {
	if len(d.lineStack) == 0 {
		panic("no lines in stack!")
	}
	return &d.lineStack[len(d.lineStack)-1]
}

func (d *DrawTool) CompleteLastLine() {
	if len(d.lineStack) == 0 {
		panic("no lines in stack!")
	}
	d.completedLines = append(d.completedLines, d.lineStack[len(d.lineStack)-1])
	d.lineStack = d.lineStack[0 : len(d.lineStack)-1]
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
	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				d.cleanup()
			case *sdl.KeyboardEvent:
				if t.Keysym.Sym == sdl.K_ESCAPE {
					d.cleanup()
				}
			case *sdl.MouseButtonEvent:
				if t.State == sdl.PRESSED {
					d.lineStack = append(d.lineStack, line{
						begin: coordinate{int(t.X), int(t.Y)},
					})
				}

				if t.State == sdl.RELEASED {
					lastLine := d.getLastLine()
					lastLine.end = coordinate{int(t.X), int(t.Y)}
					d.CompleteLastLine()
					d.drawlines()
				}

			}

		}
	}

}

func Drawer() *DrawTool {

	window, err := sdl.CreateWindow("croc", sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, 800, 600, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}

	render, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
	if err != nil {
		panic(err)
	}

	drawingTool := &DrawTool{render: render, blockChan: make(chan struct{})}
	drawingTool.drawlines()
	// handle user clicks in background
	drawingTool.handleMousevents()
	return drawingTool

}
