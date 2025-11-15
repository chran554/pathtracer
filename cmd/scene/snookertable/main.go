package main

import (
	"fmt"
	"math"
	"math/rand"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj"
	anm "pathtracer/internal/pkg/renderfile"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "snookertable"

var amountAnimationFrames = 1

var imageWidth = 1200
var imageHeight = 480
var magnification = 1.3

var amountSamples = 1024 // 1024 * 16

var apertureSize = 0.1

var ballDisplacementRadius = 0.3
var maxRotation = (math.Pi / 180) * 15 // Max angle rotation of ball

type SnookerBall string

const (
	SnookerBallWhite = "white"
	SnookerBallBlack = "black"
	SnookerBallBrown = "brown"
	SnookerBall00    = "00"
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

func main() {
	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, true, false)

	environmentSphere := scn.NewSphere(&vec3.T{0, 0, 0}, 3*100, scn.NewMaterial().
		E(color.White, 6, true).
		//C(color.NewColorGrey(0.2))).
		SP(floatimage.Load("textures/equirectangular/las-vegas-hotell-lobby.png"), &vec3.T{0, 0, 0}, vec3.T{1, 0, 0}, vec3.T{0, 1, 0})).N("sky dome")

	// Generally speaking, hanging billiard lights about 32"-36" above the bed of the table is about right.
	lamp2 := obj.NewBox(obj.BoxCentered)
	lamp2.ScaleUniform(&vec3.Zero, 0.5)
	lamp2.Scale(&vec3.Zero, &vec3.T{100, 2, 40})
	lamp2.Translate(&vec3.T{0, 33*2.54 + 5, 0}) // Raise lamp 33 inches above table cloth
	lamp2.Material = scn.NewMaterial().N("lamp").E(color.KelvinTemperatureColor2(5000), 15, true)
	// lamp1 := scn.NewSphere(&vec3.T{0, 150, -75}, 50, scn.NewMaterial().E(color.White, 18, true)).N("lamp")

	/*
		https://billiards.colostate.edu/faq/table/sizes/

		Standard size pool tables, along with the playing surface dimensions (measured between the noses of the cushions) are:

		12-ft (snooker):  140″ (356.9 cm) x 70″ (177.8 cm)
		10-ft (oversized):  112″ (284.5 cm) x 56″ (142.2 cm)
		9-ft (standard regulation size table):  100″ (254 cm) x 50″ (127 cm)
		8-ft+ (pro 8):  92″ (233.7 cm) x 46″ (116.8 cm)
		8-ft (typical home table): 88″ (223.5 cm) x 44″ (111.8 cm)
		7-ft+ (large “bar box”):  78-82″ (198.1-208.3 cm) x 39-41″ (99.1-104.1 cm)
		7-ft (“bar box”):  74-78″ (188-198.1 cm) x 37-39″ (94-99.1 cm)
		6-ft (“small bar box”):  70-74″ (177.8-188 cm) x 35-37″ (88.9-94 cm)

		The distance between the diamonds can be found by dividing the playing surface length by 8 or the width by 4.
	*/
	tableBoard := obj.NewBox(obj.BoxCentered)
	tableBoard.Translate(&vec3.T{0, -tableBoard.Bounds.Ymax, 0})
	tableBoard.Scale(&vec3.Zero, &vec3.T{356.9, 5, 177.8})
	tableBoard.Material = scn.NewMaterial().C(color.NewColorGrey(1.0)).M(0.05, 0.8).PP(floatimage.Load("textures/snooker/cloth02.png"), &vec3.T{0, 0, 0}, vec3.T{5, 0, 0}, vec3.T{0, 0, 5})

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
			snookerball.RotateZ(snookerball.Bounds().Center(), zRotAngle)
			snookerball.RotateX(snookerball.Bounds().Center(), xRotAngle)
			snookerball.RotateY(snookerball.Bounds().Center(), yRotAngle)

			x := 2.25 * snookerball.Radius * (float64(i) - float64(len(balls[j])-1)/2.0)
			z := ((float64(j) - 0.5) / 0.5) * snookerball.Radius * 2.5

			ballPerfectPosition := vec3.T{x, 0, z}
			displacementAngle := math.Pi * 2 * rand.Float64()
			ballPosition := ballPerfectPosition.Added(&vec3.T{ballDisplacementRadius * math.Cos(displacementAngle), 0, ballDisplacementRadius * math.Sin(displacementAngle)})

			snookerball.Translate(&ballPosition)
			snookerballs = append(snookerballs, snookerball)
		}
	}

	// steelSphere := scn.NewSphere(&vec3.T{0, 5.7, -10}, 5.7, scn.NewMaterial().M(0.9, 0.15))

	snookerBallsNode := scn.NewSceneNode().S(snookerballs...) // .S(steelSphere)

	scene := scn.NewSceneNode().
		S(environmentSphere).
		SN(snookerBallsNode).
		FS(tableBoard, lamp2)

	//animationStep := 1.0 / float64(amountAnimationFrames)
	for animationFrameIndex := 0; animationFrameIndex < amountAnimationFrames; animationFrameIndex++ {
		// animationProgress := float64(animationFrameIndex) * animationStep

		cameraOrigin := vec3.T{0, 25, -40}
		focusPoint := vec3.T{0, 2.5, -6}
		camera := scn.NewCamera(&cameraOrigin, &focusPoint, amountSamples, magnification).A(apertureSize, nil)

		frame := scn.NewFrame(animation.AnimationName, animationFrameIndex, camera, scene)
		animation.AddFrame(frame)
	}

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := anm.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}

func NewSnookerBall(ball SnookerBall) *scn.Sphere {
	// diameter := 5.25 // Snooker
	// diameter := 5.4  // Bumper Pool
	// diameter := 5.25 // Carom (Billiard) Balls
	diameter := 5.7 // Pool (Pocket Billiard] Balls

	radius := diameter / 2
	textureFilename := floatimage.Load(fmt.Sprintf("textures/snooker/wpi/%s_wpi.png", ball))
	ballMaterial := scn.NewMaterial().
		N(fmt.Sprintf("snooker ball %s", ball)).
		M(0.05, 0.1).
		T(0.0, true, scn.RefractionIndex_AcrylicPlastic).
		//SP(textureFilename, &vec3.T{0, radius, 0}, vec3.T{0, 0, diameter}, vec3.T{0, radius, 0})
		//PP(textureFilename, &vec3.T{-radius, 0, 0}, vec3.T{diameter, 0, 0}, vec3.T{0, diameter, 0})
		CP(textureFilename, &vec3.T{0, 0, 0}, vec3.UnitZ, vec3.T{0, diameter, 0}, false)

	return scn.NewSphere(&vec3.T{0, radius, 0}, radius, ballMaterial)
}
