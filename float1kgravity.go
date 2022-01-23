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
	radius int64
	vx     *big.Float
	vy     *big.Float
	color  [3]uint8
	dead   bool
}

type force struct {
	x *big.Float
	y *big.Float
}

var (
	G      = 100.0
	debug  = false
	lines  = true
	bodies = [numBodies]body{}
	mspt   = 0.0
	path   *ebiten.Image
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
		dead:   false,
	}
	bodies[1] = body{
		mass:   newFloat1024(1000),
		posx:   newFloat1024(600),
		posy:   newFloat1024(400),
		radius: 10,
		vx:     newFloat1024(0),
		vy:     newFloat1024(10),
		color:  [3]uint8{0, 255, 0},
		dead:   false,
	}
	bodies[2] = body{
		mass:   newFloat1024(1000),
		posx:   newFloat1024(400),
		posy:   newFloat1024(200),
		radius: 10,
		vx:     newFloat1024(10),
		vy:     newFloat1024(0),
		color:  [3]uint8{0, 0, 255},
		dead:   false,
	}
	bodies[3] = body{
		mass:   newFloat1024(1000),
		posx:   newFloat1024(400),
		posy:   newFloat1024(600),
		radius: 10,
		vx:     newFloat1024(-10),
		vy:     newFloat1024(0),
		color:  [3]uint8{255, 255, 0},
		dead:   false,
	}

	mspt = 1 / tps
}

func combineRadii(radius int64, radius2 int64) int64 {
	fpi := newFloat1024(math.Pi)
	//first body
	fr1 := newFloat1024(float64(radius))

	var area1 big.Float
	var r21 big.Float
	r21.SetPrec(1024)
	r21.Mul(fr1, fr1)
	area1.Mul(fpi, &r21) //Pi*r^2

	// second body
	fr2 := newFloat1024(float64(radius2))

	var area2 big.Float
	var r22 big.Float
	r22.SetPrec(1024)
	r22.Mul(fr2, fr2)
	area2.Mul(fpi, &r22) //Pi*r^2

	var areas big.Float
	areas.Add(&area1, &area2)
	areas.Quo(&areas, fpi)
	areas.Sqrt(&areas) // sqrt(A/Pi)

	intify, _ := areas.Int64()

	return intify
}

func collideBodies(body1 *body, body2 *body, dx *big.Float, dy *big.Float) {
	body1.mass.Add(body1.mass, body2.mass)                  // combine mass
	body1.posx.Add(body1.posx, dx.Quo(dx, newFloat1024(2))) // average x
	body1.posy.Add(body1.posy, dx.Quo(dy, newFloat1024(2))) // average y

	body1.radius = combineRadii(body1.radius, body2.radius) // combine radius

	body1.vx.Add(body1.vx, body2.vx) // combine velocities
	body1.vy.Add(body1.vy, body2.vy)

	body1.color[0] = (body1.color[0] + body2.color[0]) / 2 // average color
	body1.color[1] = (body1.color[1] + body2.color[1]) / 2
	body1.color[2] = (body1.color[2] + body2.color[2]) / 2

	body1.dead = false // enable
	body2.dead = true
}

type distance struct {
	r  big.Float
	dx big.Float
	dy big.Float
}

func findDistance(i int, j int) distance {
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
	r.Sqrt(&r) // Heaviest operation here

	//store distance
	newdist := distance{
		r:  r,
		dx: dx,
		dy: dy,
	}

	return newdist
}

func gravityTick() {
	//inspired by https://towardsdatascience.com/implementing-2d-physics-in-javascript-860a7b152785 (simplest implementation of gravity I found)

	//initialize variables
	distances := [numBodies][numBodies]distance{}
	// distanceschan := [numBodies][numBodies]chan *distance{}
	forces := [numBodies]force{}

	// // unsynchronize
	// for i := 0; i < numBodies; i++ {
	// 	for j := 0; j < numBodies; j++ {
	// 		distanceschan[i][j] = make(chan *distance)
	// 	}
	// }

	//create threads (threads removed because they broke everything)
	for i := 0; i < numBodies; i++ {
		for j := 0; j < numBodies; j++ {
			if i != j && !bodies[i].dead && !bodies[j].dead {
				distances[i][j] = findDistance(i, j)
				fmt.Println(distances[i][j])
			}
		}
	}

	// // synchronize
	// for i := 0; i < numBodies; i++ {
	// 	for j := 0; j < numBodies; j++ {
	// 		distances[i][j] = *<-distanceschan[i][j]
	// 	}
	// }

	//test collisions
	for i := 0; i < numBodies; i++ {
		for j := 0; j < numBodies; j++ {
			if i != j && !bodies[i].dead && !bodies[j].dead {
				r := &distances[i][j].r
				dx := &distances[i][j].dx
				dy := &distances[i][j].dy

				if r.Cmp(newFloat1024(float64(bodies[i].radius+bodies[j].radius))) == -1 {
					collideBodies(&bodies[i], &bodies[j], dx, dy)
				}
			}
		}
	}

	//calculate gravity forces
	for i := 0; i < numBodies; i++ {
		for j := 0; j < numBodies; j++ {
			if i != j && !bodies[i].dead && !bodies[j].dead {
				r := &distances[i][j].r
				dx := &distances[i][j].dx
				dy := &distances[i][j].dy

				//newtonian force
				var f big.Float
				f.SetPrec(1024)
				f.Mul(bodies[i].mass, bodies[j].mass).Mul(&f, newFloat1024(G)) // Gm1m2
				var r2 big.Float
				r2.SetPrec(1024)
				r2.Mul(r, r)   // r^2
				f.Quo(&f, &r2) // Gm1m2/r^2

				//distribute force between x & y
				var fx big.Float
				fx.SetPrec(1024)
				fx.Mul(&f, dx)
				fx.Quo(&fx, r)
				var fy big.Float
				fy.SetPrec(1024)
				fy.Mul(&f, dy)
				fy.Quo(&fy, r)

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
	if inpututil.IsKeyJustPressed(ebiten.KeyL) {
		path.Clear()
		if lines {
			lines = false
		} else {
			lines = true
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

func (g *Game) Draw(screen *ebiten.Image) {

	//screen.Fill(color.Black)
	if lines {
		doptions := ebiten.DrawImageOptions{}
		screen.DrawImage(path, &doptions)
	}

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
			info := fmt.Sprintf("FPS:  %f / TPS:  %f / mspt: %f\nG: %f", ebiten.CurrentFPS(), ebiten.CurrentTPS(), mspt, G)
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
