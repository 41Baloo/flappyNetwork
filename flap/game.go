package game

import (
	"flappyNetwork/network"
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth  = 1000
	screenHeight = 500
	birdSize     = 30
	gravity      = 0.5
	jumpStrength = -8
	pipeWidth    = 60
	pipeGap      = 200
	pipeSpacing  = 300
	pipeSpeed    = 2
)

type Bird struct {
	Dead        bool
	Y           float64
	Velocity    float64
	JumpRate    float64
	TestMode    bool
	TestNetwork *network.NeuralNetwork
	Score       int
	Godmode     bool
	Color       color.Color
}

type Game struct {
	Birds       []*Bird
	pipes       []Pipe
	score       int
	IsGameOver  bool
	rand        rand.Rand
	gameSpeed   float64
	pipeGap     float64
	timesJumped int
	WaitTime    int
}

type Pipe struct {
	x      float64
	height float64
}

func (g *Game) Update() error {

	if ebiten.IsKeyPressed(ebiten.KeyR) {

		for _, bird := range g.Birds {
			bird.Y = screenHeight / 2
			bird.Velocity = 0
			bird.Score = 0
			bird.Dead = false
		}

		g.pipes = []Pipe{}
		g.score = 0
		g.gameSpeed = pipeSpeed
		g.pipeGap = pipeGap
		g.IsGameOver = false
	}

	time.Sleep(time.Duration(g.WaitTime) * time.Millisecond)

	if g.IsGameOver {
		return nil
	}

	for index, bird := range g.Birds {

		bird.Y += bird.Velocity
		bird.Velocity += gravity

		if bird.Godmode {

			bird.Dead = false

			if bird.Y > 475 {
				bird.Y = 475
				bird.Velocity = 0
			}

		}

		if !bird.Dead {

			if bird.TestMode {
				in1, in2, in3, in4, in5, in6 := g.GetInputs(bird)
				bird.JumpRate = bird.TestNetwork.Predict([]float64{in1, in2, in3, in4, in5, in6})
			}
			if bird.JumpRate > 0.5 || (ebiten.IsKeyPressed(ebiten.KeySpace) && index == 1) {
				bird.Velocity = jumpStrength
				g.timesJumped++
				bird.JumpRate = 0
			}

			if bird.Y < 0 || bird.Y > screenHeight {
				bird.Dead = true
				g.IsGameOver = g.CalcGameOver()
				if g.IsGameOver {
					return nil
				}
			}

		}
	}

	for i := 0; i < len(g.pipes); i++ {
		g.pipes[i].x -= g.gameSpeed
	}

	if len(g.pipes) == 0 || g.pipes[len(g.pipes)-1].x < screenWidth-pipeSpacing {
		height := g.rand.Float64()*(screenHeight-2*birdSize-g.pipeGap) + birdSize
		g.pipes = append(g.pipes, Pipe{x: screenWidth, height: height})
	}

	for i := 0; i < len(g.pipes); i++ {
		if g.pipes[i].x < -pipeWidth {
			g.pipes = append(g.pipes[:i], g.pipes[i+1:]...)
			i--
			g.score++
			if g.gameSpeed < 6 {
				g.gameSpeed += 0.1
			}
			if g.pipeGap > 180 {
				g.pipeGap -= 0.1
			}
			for _, bird := range g.Birds {
				if !bird.Dead {
					bird.Score++
				}
			}
			continue
		}

		upperPipeTop := 0.0
		upperPipeBottom := float64(g.pipes[i].height)
		upperPipeLeft := g.pipes[i].x
		upperPipeRight := g.pipes[i].x + pipeWidth

		lowerPipeTop := g.pipes[i].height + g.pipeGap
		lowerPipeBottom := float64(screenHeight)
		lowerPipeLeft := g.pipes[i].x
		lowerPipeRight := g.pipes[i].x + pipeWidth

		for _, bird := range g.Birds {
			birdLeft := float64(birdSize / 2)
			birdRight := float64(birdSize/2 + birdSize)
			birdTop := float64(bird.Y)
			birdBottom := bird.Y + birdSize

			collidesWithUpper := birdRight > upperPipeLeft && birdLeft < upperPipeRight && birdTop < upperPipeBottom && birdBottom > upperPipeTop
			collidesWithLower := birdRight > lowerPipeLeft && birdLeft < lowerPipeRight && birdTop < lowerPipeBottom && birdBottom > lowerPipeTop

			if collidesWithUpper || collidesWithLower {
				bird.Dead = true
				g.IsGameOver = g.CalcGameOver()
				if g.IsGameOver {
					return nil
				}
			}
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0x80, 0xa0, 0xc0, 0xff})

	for index, bird := range g.Birds {
		birdRect := ebiten.NewImage(birdSize, birdSize)
		if index == 1 {
			birdRect.Fill(bird.Color)
		} else {
			birdRect.Fill(color.White)
		}
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(birdSize/2, bird.Y)
		screen.DrawImage(birdRect, opts)
	}

	pipeColor := color.RGBA{0x00, 0x80, 0x00, 0xff}
	for _, pipe := range g.pipes {
		upperPipe := ebiten.NewImage(pipeWidth, int(pipe.height))
		lowerPipe := ebiten.NewImage(pipeWidth, screenHeight-int(pipe.height+g.pipeGap))

		upperPipe.Fill(pipeColor)
		lowerPipe.Fill(pipeColor)

		upperOpts := &ebiten.DrawImageOptions{}
		upperOpts.GeoM.Translate(pipe.x, 0)
		screen.DrawImage(upperPipe, upperOpts)

		lowerOpts := &ebiten.DrawImageOptions{}
		lowerOpts.GeoM.Translate(pipe.x, pipe.height+g.pipeGap)
		screen.DrawImage(lowerPipe, lowerOpts)

	}

	nPipe, distance := g.NearestPipe()
	xPipe, yPipe := g.GapBounds(nPipe)
	msg := `
	[ Inputs ]:
	gSpeed ` + fmt.Sprint(g.gameSpeed) + `;
	bHeight: ` + fmt.Sprint(g.Birds[0].Y) + `; 
	bVel: ` + fmt.Sprint(g.Birds[0].Velocity) + `;
	pDst: ` + fmt.Sprint(distance) + `; 
	pU: ` + fmt.Sprint(xPipe) + `;
	pL: ` + fmt.Sprint(yPipe) + `;
	
	[ Output ]:
	` + fmt.Sprint(g.Birds[0].JumpRate) + `;
	
	[ Info ]:
	FPS: ` + fmt.Sprint(ebiten.ActualFPS()) + `;
	TPS: ` + fmt.Sprint(ebiten.ActualTPS()) + `;
	Time: ` + fmt.Sprint(g.WaitTime) + `;`

	ebiten.SetWindowTitle("[ AI ]: " + fmt.Sprint(g.Birds[0].Score) + " [ Player ]: " + fmt.Sprint(g.Birds[1].Score))

	ebitenutil.DebugPrintAt(screen, msg, 0, 0)

	if g.IsGameOver {
		msg := "Game Over! Press 'R' to Restart\nAI: " + fmt.Sprint(g.Birds[0].Score) + " Player: " + fmt.Sprint(g.Birds[1].Score)
		textWidth := len(msg) * 8
		textX := (screenWidth - textWidth) / 2
		textY := screenHeight / 2
		ebitenutil.DebugPrintAt(screen, msg, textX, textY)
	}
}

func (g *Game) Layout(_, _ int) (int, int) {
	return screenWidth, screenHeight
}

func NewGame(seed int64, amount int) *Game {

	birds := make([]*Bird, amount)

	for i := range birds {
		birds[i] = &Bird{
			Y:        screenHeight / 2,
			Velocity: 0,
			Color:    color.Black,
		}
	}

	return &Game{
		Birds:     birds,
		rand:      *rand.New(rand.NewSource(seed)),
		gameSpeed: 2,
		pipeGap:   pipeGap,
	}
}

func (g *Game) NearestPipe() (*Pipe, float64) {
	for i := range g.pipes {
		if g.pipes[i].x+pipeWidth >= birdSize/2 {
			distance := g.pipes[i].x - birdSize/2
			return &g.pipes[i], distance
		}
	}
	return nil, 0
}

func (g *Game) GapBounds(pipe *Pipe) (float64, float64) {
	return pipe.height, pipe.height + g.pipeGap
}

func (g *Game) GetScore() float64 {
	_, distance := g.NearestPipe()
	return float64(g.score) + (1.0 / (math.Abs(distance) + 0.001))
}

func (g *Game) CalcGameOver() bool {
	allDead := true
	for _, b := range g.Birds {
		if !b.Dead {
			allDead = false
		}
	}
	return allDead
}

func (g *Game) GetInputs(bird *Bird) (pipeSpeed float64, birdC float64, birdV float64, distance float64, pipeU float64, pipeD float64) {
	pipe, distance := g.NearestPipe()
	pU := 0.0
	pD := 0.0
	if pipe != nil {
		pU, pD = g.GapBounds(pipe)
	}
	return g.gameSpeed, bird.Y, bird.Velocity, distance, pU, pD
}
