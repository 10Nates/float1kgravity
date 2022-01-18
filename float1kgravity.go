package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/big"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	width     = 800
	height    = 800
	windowmul = 1
	fps       = 60
	tps       = 5000
	numBodies = 4
)

type body struct {
	mass   *big.Float
	posx   *big.Float
	posy   *big.Float
	radius int
	vx     *big.Float
	vy     *big.Float
	color  [3]uint8
}

type force struct {
	x *big.Float
	y *big.Float
}

var (
	G      = 100
	debug  = false
	bodies = [numBodies]body{}
	forces = [numBodies]force{}
	mspt   = 0.0
)

func newFloat1024(n float64) *big.Float {
	var x big.Float
	x.SetPrec(1024)
	x.SetFloat64(n)
	return &x
}

func initBodies() {
	//values from https://github.com/Nekodigi/Physics/blob/master/Three_body_problem/Three_body_problem.pde
	bodies[0] = body{
		mass:   newFloat1024(1000),
		posx:   newFloat1024(200),
		posy:   newFloat1024(400),
		radius: 10,
		vx:     newFloat1024(0),
		vy:     newFloat1024(-10),
		color:  [3]uint8{255, 0, 0},
	}
	bodies[1] = body{
		mass:   newFloat1024(1000),
		posx:   newFloat1024(600),
		posy:   newFloat1024(400),
		radius: 10,
		vx:     newFloat1024(0),
		vy:     newFloat1024(10),
		color:  [3]uint8{0, 255, 0},
	}
	bodies[2] = body{
		mass:   newFloat1024(1000),
		posx:   newFloat1024(400),
		posy:   newFloat1024(200),
		radius: 10,
		vx:     newFloat1024(10),
		vy:     newFloat1024(0),
		color:  [3]uint8{0, 0, 255},
	}
	bodies[3] = body{
		mass:   newFloat1024(1000),
		posx:   newFloat1024(400),
		posy:   newFloat1024(600),
		radius: 10,
		vx:     newFloat1024(-10),
		vy:     newFloat1024(0),
		color:  [3]uint8{255, 255, 0},
	}

	for i := 0; i < numBodies; i++ {
		forces[i] = force{
			x: newFloat1024(0),
			y: newFloat1024(0),
		}
	}

	mspt = 1 / tps
}

func gravityTick() {
	//inspired by https://towardsdatascience.com/implementing-2d-physics-in-javascript-860a7b152785 (simplest implementation of gravity I found)

	//calculate forces
	for i := 0; i < numBodies; i++ {
		for j := 0; j < numBodies; j++ {
			if i != j {
				//distance
				var dx big.Float
				dx.SetPrec(1024)
				dx.Sub(bodies[j].posx, bodies[i].posx)
				var dy big.Float
				dy.SetPrec(1024)
				dy.Sub(bodies[j].posy, bodies[i].posy)
				var r big.Float
				r.SetPrec(1024)
				var dxt big.Float // needed to keep dx
				dxt.SetPrec(1024)
				var dyt big.Float // needed to keep dy
				dyt.SetPrec(1024)
				r.Add(dxt.Mul(&dx, &dx), dyt.Mul(&dy, &dy))
				r.Sqrt(&r)

				//newtonian force
				var f big.Float
				f.SetPrec(1024)
				f.Mul(bodies[i].mass, bodies[j].mass).Mul(&f, newFloat1024(G)) // Gm1m2
				var r2 big.Float
				r2.SetPrec(1024)
				r2.Mul(&r, &r) // r^2
				f.Quo(&f, &r2) // Gm1m2/r^2

				//distribute force between x & y
				var fx big.Float
				fx.SetPrec(1024)
				fx.Mul(&f, &dx)
				fx.Quo(&fx, &r)
				var fy big.Float
				fy.SetPrec(1024)
				fy.Mul(&f, &dy)
				fy.Quo(&fy, &r)

				//distribute force between i & j
				forces[i].x.Add(forces[i].x, &fx)
				forces[i].y.Add(forces[i].y, &fy)
				forces[j].x.Sub(forces[j].x, &fx)
				forces[j].y.Sub(forces[j].y, &fy)
			}
		}
	}

	//apply forces
	var dt big.Float
	dt.Quo(newFloat1024(1), newFloat1024(tps))

	for i := 0; i < numBodies; i++ {

		var ax big.Float // acceleration x
		ax.SetPrec(1024)
		ax.Quo(forces[i].x, bodies[i].mass)

		var ay big.Float // acceleration y
		ay.SetPrec(1024)
		ay.Quo(forces[i].y, bodies[i].mass)

		ax.Mul(&ax, &dt)                    // maybe required?
		bodies[i].vx.Add(bodies[i].vx, &ax) // add acceleration to velocities
		ay.Mul(&ay, &dt)                    // maybe required?
		bodies[i].vy.Add(bodies[i].vy, &ay)

		var vx big.Float
		vx.SetPrec(1024)
		vx.Mul(bodies[i].vx, &dt)               // velocity * time
		bodies[i].posx.Add(bodies[i].posx, &vx) //change x position

		var vy big.Float
		vy.SetPrec(1024)
		vy.Mul(bodies[i].vy, &dt)               // velocity * time
		bodies[i].posy.Add(bodies[i].posy, &vy) //change y position
	}

	for i := 0; i < numBodies; i++ {
		forces[i] = force{
			x: newFloat1024(0),
			y: newFloat1024(0),
		}
	}
}

func controlsTick() {
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		if debug {
			debug = false
		} else {
			debug = true
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyI) {
		G += 1
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyK) {
		G -= 1
		if G < 1 {
			G = 1
		}
	}
}

//ebiten gameloop

type Game struct{}

func (g *Game) Update() error {
	tickTime := time.Now().Nanosecond()

	controlsTick()
	gravityTick()

	mspt = float64(time.Now().Nanosecond()-tickTime) / 1000000
	return nil
}

var path *ebiten.Image

func (g *Game) Draw(screen *ebiten.Image) {

	//screen.Fill(color.Black)
	doptions := ebiten.DrawImageOptions{}
	screen.DrawImage(path, &doptions)

	for i := 0; i < numBodies; i++ {
		clr := color.RGBA{
			R: bodies[i].color[0],
			G: bodies[i].color[1],
			B: bodies[i].color[2],
			A: 255,
		}
		posx, _ := bodies[i].posx.Float64()
		posy, _ := bodies[i].posy.Float64()

		//s := bodies[i].radius * 2
		img := ebiten.NewImage(3, 3)
		img.Fill(clr)
		//var ioptions ebiten.DrawImageOptions
		// ioptions.GeoM.Translate(posx-float64(bodies[i].radius), posy-float64(bodies[i].radius))
		// screen.DrawImage(img, &ioptions)

		//planets
		var planet vector.Path
		radius32 := float32(bodies[i].radius)
		planet.MoveTo(float32(posx), float32(posy))
		planet.Arc(float32(posx), float32(posy), radius32, 0, math.Pi*2, vector.Clockwise) // draw circle
		//circle colors
		vs, is := planet.AppendVerticesAndIndicesForFilling(nil, nil)
		for j := range vs {
			vs[j].ColorR = float32(clr.R)
			vs[j].ColorG = float32(clr.G)
			vs[j].ColorB = float32(clr.B)
		}
		//render circle
		var coptions ebiten.DrawTrianglesOptions
		coptions.FillRule = ebiten.EvenOdd
		screen.DrawTriangles(vs, is, img, &coptions)

		//path of bodies
		pimg := ebiten.NewImage(1, 1)
		clr.A = 127
		pimg.Fill(clr)
		var poptions ebiten.DrawImageOptions
		poptions.GeoM.Translate(posx, posy)
		path.DrawImage(pimg, &poptions)

		if debug {
			info := fmt.Sprintf("FPS:  %f / TPS:  %f / mspt: %f\nG: %d", ebiten.CurrentFPS(), ebiten.CurrentTPS(), mspt, G)
			ebitenutil.DebugPrint(screen, info)
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return width, height
}

func main() {
	// modify G & debug screen

	initBodies()

	//Ebiten
	game := &Game{}
	ebiten.SetWindowSize(width*windowmul, height*windowmul)
	ebiten.SetWindowTitle("Float1k Gravity")
	ebiten.SetMaxTPS(tps)
	path = ebiten.NewImage(width, height)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
