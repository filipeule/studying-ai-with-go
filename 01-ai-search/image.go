package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"

	"github.com/StephaneBunel/bresenham"
	"github.com/kmicki/apng"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// constant
const cellSize = 60

// variables for colour
var (
	green     = color.RGBA{G: 255, A: 255}
	darkGreen = color.RGBA{R: 1, G: 100, B: 32, A: 255}
	red       = color.RGBA{R: 255, A: 255}
	yellow    = color.RGBA{R: 255, G: 255, B: 101, A: 255}
	gray      = color.RGBA{R: 125, G: 125, B: 125, A: 255}
	orange    = color.RGBA{R: 255, G: 140, B: 25, A: 255}
	blue      = color.RGBA{R: 14, G: 180, B: 173, A: 255}
)

// OutputImage draw the maze as a png file
func (g *Maze) OutputImage(filename ...string) {
	fmt.Printf("generating image %s...\n", filename)

	width := cellSize * (g.Width - 1)
	height := cellSize * g.Height

	var outfile = "image.png"
	if len(filename) > 0 {
		outfile = filename[0]
	}

	upLeft := image.Point{}
	lowRight := image.Point{X: width, Y: height}

	img := image.NewRGBA(image.Rectangle{Min: upLeft, Max: lowRight})
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.Black}, image.Point{}, draw.Src)

	// draw squares on the image
	for i, row := range g.Walls {
		for j, col := range row {
			p := Point{Row: i, Col: j}
			if col.wall {
				// draw black square for wall
				g.drawSquare(col, p, img, color.Black, cellSize, j*cellSize, i*cellSize)
			} else if col.State.Row == g.Start.Row && col.State.Col == g.Start.Col {
				// starting point, dark green square
				g.drawSquare(col, p, img, darkGreen, cellSize, j*cellSize, i*cellSize)
			} else if col.State.Row == g.Goal.Row && col.State.Col == g.Goal.Col {
				// ending point, red square
				g.drawSquare(col, p, img, red, cellSize, j*cellSize, i*cellSize)
			} else if g.inSolution(p) {
				// part of solution, so draw green square
				g.drawSquare(col, p, img, green, cellSize, j*cellSize, i*cellSize)
			} else if col.State == g.CurrentNode.State {
				// current location, draw in orange
				g.drawSquare(col, p, img, orange, cellSize, j*cellSize, i*cellSize)
			} else if col.State.Water {
				// flooded point, blue square
				g.drawSquare(col, p, img, blue, cellSize, j*cellSize, i*cellSize)
			} else if inExplored(Point{i, j, false}, g.Explored) {
				// an explored cell, draw in yellow
				g.drawSquare(col, p, img, yellow, cellSize, j*cellSize, i*cellSize)
			} else {
				// empty unexplored, draw in white
				g.drawSquare(col, p, img, color.White, cellSize, j*cellSize, i*cellSize)
			}
		}
	}

	// draw a grid
	for i := range g.Walls {
		bresenham.DrawLine(img, 0, i*cellSize, g.Width*cellSize, i*cellSize, gray)
	}

	for i := 0; i <= g.Width; i++ {
		bresenham.DrawLine(img, i*cellSize, 0, i*cellSize, g.Height*cellSize, gray)
	}

	f, _ := os.Create(outfile)
	defer f.Close()

	_ = png.Encode(f, img)
}

// drawSquare
func (g *Maze) drawSquare(col Wall, p Point, img *image.RGBA, c color.Color, size, x, y int) {
	patch := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.Draw(patch, patch.Bounds(), &image.Uniform{C: c}, image.Point{}, draw.Src)

	if !col.wall {
		switch g.SearchType {
		case DIJKSTRA, GBFS:
			g.printManhattanCost(p, color.Black, patch)
		case ASTAR:
			g.printTotalCost(p, color.Black, patch)
		default:
			// do nothing
		}

		// check to see if this cell is flooded
		if col.State.Water {
			g.printWater(blue, patch)
		}

		// print the x y coordinates of this cell
		g.printLocation(p, color.Black, patch)
	}

	draw.Draw(img, image.Rect(x, y, x+size, y+size), patch, image.Point{}, draw.Src)
}

func (g *Maze) printManhattanCost(p Point, c color.Color, patch *image.RGBA) {
	point := fixed.Point26_6{X: fixed.I(6), Y: fixed.I(17)}
	d := &font.Drawer{
		Dst:  patch,
		Src:  image.NewUniform(c),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	n := Node{
		State: p,
	}

	switch g.SearchType {
	case DIJKSTRA:
		d.DrawString(fmt.Sprintf("%d", n.ManhattanDistance(g.Start)))
	case GBFS:
		d.DrawString(fmt.Sprintf("%d", n.ManhattanDistance(g.Goal)))
	default:
		// do nothing
	}
}

func (g *Maze) printTotalCost(p Point, c color.Color, patch *image.RGBA) {
	point := fixed.Point26_6{X: fixed.I(6), Y: fixed.I(17)}
	d := &font.Drawer{
		Dst:  patch,
		Src:  image.NewUniform(c),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	n := Node{
		State: p,
	}

	fromStart := n.ManhattanDistance(g.Start)
	toGoal := euclideanDist(p, g.Goal)
	d.DrawString(fmt.Sprintf("%.2f", float64(fromStart)+toGoal))
}

// printLocation
func (g *Maze) printLocation(p Point, c color.Color, patch *image.RGBA) {
	point := fixed.Point26_6{X: fixed.I(6), Y: fixed.I(40)}
	d := &font.Drawer{
		Dst:  patch,
		Src:  image.NewUniform(c),
		Face: basicfont.Face7x13,
		Dot:  point,
	}

	d.DrawString(fmt.Sprintf("[%d %d]", p.Row, p.Col))
}

func (g *Maze) printWater(c color.Color, patch *image.RGBA) {
	point := fixed.Point26_6{X: fixed.I(50), Y: fixed.I(18)}
	d := &font.Drawer{
		Dst:  patch,
		Src:  image.NewUniform(c),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString("W")
}

func (g *Maze) OutputAnimatedImage() {
	output := "./animation.png"

	files, _ := os.ReadDir("./tmp")

	var images []string
	var delays []int

	for _, file := range files {
		images = append(images, fmt.Sprintf("./tmp/%s", file.Name()))
		delays = append(delays, 15)
	}

	images = append(images, "./image.png")
	delays = append(delays, 200)

	a := apng.APNG{
		Frames: make([]apng.Frame, len(images)),
	}

	out, _ := os.Create(output)
	defer out.Close()

	for i, s := range images {
		in, err := os.Open(s)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer in.Close()

		m, err := png.Decode(in)
		if err != nil {
			continue
		}

		a.Frames[i].Image = m
		a.Frames[i].DelayNumerator = uint16(delays[i])
	}

	err := apng.Encode(out, a)
	if err != nil {
		fmt.Println(err)
	}
}
