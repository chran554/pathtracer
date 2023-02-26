package main

import (
	"fmt"
	"math"
	"math/rand"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "snookertable"

var amountAnimationFrames = 1

var imageWidth = 1200
var imageHeight = 480
var magnification = 1.3

var amountSamples = 1024 * 10

var apertureSize = 0.1

var ballDisplacementRadius = 0.3
var maxRotation = (math.Pi / 180) * 5 // Max angle rotation of ball

func main() {
	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, true)

	environmentSphere := scn.NewSphere(&vec3.T{0, 0, 0}, 3*100, scn.NewMaterial().
		E(color.White, 1, true).
		//C(color.NewColorGrey(0.2))).
		SP("textures/equirectangular/las-vegas-hotell-lobby.png", &vec3.T{0, 0, 0}, vec3.T{1, 0, 0}, vec3.T{0, 1, 0})).N("sky dome")

	lamp1 := scn.NewSphere(&vec3.T{0, 150, -75}, 50, scn.NewMaterial().E(color.White, 18, true)).N("lamp")

	tableBoard := obj.NewBox(obj.BoxCentered)
	tableBoard.Translate(&vec3.T{0, -tableBoard.Bounds.Ymax, 0})
	tableBoard.Scale(&vec3.Zero, &vec3.T{356.9, 5, 177.8})
	tableBoard.Material = scn.NewMaterial().C(color.NewColorGrey(1.0)).M(0.05, 0.8).PP("textures/snooker/cloth02.png", &vec3.T{0, 0, 0}, vec3.T{5, 0, 0}, vec3.T{0, 0, 5})

	var balls = [][]SnookerBall{
		{SnookerBall01, SnookerBall02, SnookerBall03, SnookerBall04, SnookerBall05, SnookerBall06, SnookerBall07, SnookerBall08},
		{SnookerBall09, SnookerBall10, SnookerBall11, SnookerBall12, SnookerBall13, SnookerBall14, SnookerBall15, SnookerBallWhite},
	}

	var snookerballs []*scn.Sphere
	for j := 0; j < len(balls); j++ {
		for i := 0; i < len(balls[j]); i++ {
			snookerball := NewSnookerBall(balls[j][i])

			xRotAngle := maxRotation * (rand.Float64()*2 - 1)
			yRotAngle := maxRotation * (rand.Float64()*2 - 1)
			zRotAngle := maxRotation * (rand.Float64()*2 - 1)
			snookerball.RotateX(snookerball.Bounds().Center(), xRotAngle)
			snookerball.RotateY(snookerball.Bounds().Center(), yRotAngle)
			snookerball.RotateZ(snookerball.Bounds().Center(), zRotAngle)

			ballPerfectPosition := vec3.T{2.25 * snookerball.Radius * (float64(i) - float64(len(balls[j])-1)/2.0), 0, ((float64(j) - 0.5) / 0.5) * snookerball.Radius * 2.5}
			displacementAngle := math.Pi * 2 * rand.Float64()
			ballPosition := ballPerfectPosition.Added(&vec3.T{ballDisplacementRadius * math.Cos(displacementAngle), 0, ballDisplacementRadius * math.Sin(displacementAngle)})

			snookerball.Translate(&ballPosition)
			snookerballs = append(snookerballs, snookerball)
		}
	}

	// steelSphere := scn.NewSphere(&vec3.T{0, 5.7, -10}, 5.7, scn.NewMaterial().M(0.9, 0.15))

	snookerBallsNode := scn.NewSceneNode().S(snookerballs...) // .S(steelSphere)

	scene := scn.NewSceneNode().
		S(lamp1, environmentSphere).
		SN(snookerBallsNode).
		FS(tableBoard)

	//animationStep := 1.0 / float64(amountAnimationFrames)
	for animationFrameIndex := 0; animationFrameIndex < amountAnimationFrames; animationFrameIndex++ {
		// animationProgress := float64(animationFrameIndex) * animationStep

		cameraOrigin := vec3.T{0, 25, -40}
		focusPoint := vec3.T{0, 2.5, -6}
		camera := scn.NewCamera(&cameraOrigin, &focusPoint, amountSamples, magnification).A(apertureSize, "")

		frame := scn.NewFrame(animation.AnimationName, animationFrameIndex, camera, scene)
		animation.AddFrame(frame)
	}

	anm.WriteAnimationToFile(animation, false)
}

type SnookerBall string

const (
	SnookerBallWhite = "white"
	SnookerBall01    = "01"
	SnookerBall02    = "02"
	SnookerBall03    = "03"
	SnookerBall04    = "04"
	SnookerBall05    = "05"
	SnookerBall06    = "06"
	SnookerBall07    = "07"
	SnookerBall08    = "08"
	SnookerBall09    = "09"
	SnookerBall10    = "10"
	SnookerBall11    = "11"
	SnookerBall12    = "12"
	SnookerBall13    = "13"
	SnookerBall14    = "14"
	SnookerBall15    = "15"
)

func NewSnookerBall(ball SnookerBall) *scn.Sphere {
	//diameter := 5.25 // Snooker
	//diameter := 5.4  // Bumper Pool
	//diameter := 5.25 // Carom (Billiard) Balls
	diameter := 5.7 // Pool (Pocket Billiard] Balls

	radius := diameter / 2
	textureFilename := fmt.Sprintf("textures/snooker/%s.png", ball)
	ballMaterial := scn.NewMaterial().
		N(fmt.Sprintf("snooker ball %s", ball)).
		M(0.05, 0.05).
		PP(textureFilename, &vec3.T{-radius, 0, 0}, vec3.T{diameter, 0, 0}, vec3.T{0, diameter, 0})
	//CP(textureFilename, &vec3.T{0, 0, 0}, vec3.UnitZ, vec3.T{0, diameter, 0}, false)

	return scn.NewSphere(&vec3.T{0, radius, 0}, radius, ballMaterial)
}
