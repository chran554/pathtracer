package main

import (
	"fmt"
	"math"
	"math/rand"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"
	"pathtracer/internal/pkg/util"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "snookertable_die"

var amountAnimationFrames = 1

var imageWidth = 400
var imageHeight = 300
var magnification = 4.0

var amountSamples = 1024 * 24 // 1024 * 16

var apertureSize = 0.4

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
	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, true, true)

	textureEnvironment := floatimage.LoadOrPanic("textures/equirectangular/las-vegas-hotell-lobby.png")
	environmentSphere := scene.NewSphere(&vec3.T{0, 0, 0}, 3*100, scene.NewMaterial().
		E(color.White, 2, true).
		//C(color.NewColorGrey(0.2))).
		SP(textureEnvironment, &vec3.T{0, 0, 0}, vec3.T{1, 0, 0}, vec3.T{0, 1, 0})).N("sky dome")

	// Generally speaking, hanging billiard lights about 32"-36" above the bed of the table is about right.
	//
	// The lights over a pool, snooker or billiard table must be at least 520 lux,
	// and the minimum height of the fixture is no lower than 1.016m above the bed of the table.
	// https://www.cits.wa.gov.au/sport-and-recreation/sports-dimensions-guide/billiards-pool-and-snooker
	lamp2 := obj.NewSquareFacetStructure(obj.SquareTypeXZPlane, false, false)
	lamp2.CenterOn(&vec3.Zero)
	lamp2.Scale(&vec3.Zero, &vec3.T{100, 1, 40})
	lamp2.Translate(&vec3.T{0, 33 * 2.54, -10}) // Raise the lamp 33 inches above the table cloth (and a little bit in front of the balls)
	lamp2.Material = scene.NewMaterial().N("lamp").E(color.KelvinTemperatureColor2(4500), 8, true)
	// lamp1 := scene.NewSphere(&vec3.T{0, 150, -75}, 50, scene.NewMaterial().E(color.White, 18, true)).N("lamp")

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
	poolTable := obj.NewSquareFacetStructure(obj.SquareTypeXZPlane, false, false)
	poolTable.CenterOn(&vec3.Zero)
	poolTable.Scale(&vec3.Zero, &vec3.T{356.9, 1, 177.8})
	textureTableCloth := floatimage.LoadOrPanic("textures/snooker/cloth02.png")
	poolTable.Material = scene.NewMaterial().
		C(color.White).
		M(0.015, 0.9).
		PP(textureTableCloth, &vec3.T{2.5, 0, 2.5}, vec3.T{5, 0, 0}, vec3.T{0, 0, 5})

	var balls = [][]SnookerBall{
		{SnookerBall01, SnookerBall02, SnookerBall03, SnookerBall04, SnookerBall05, SnookerBall06, SnookerBall07, SnookerBall08},
		{SnookerBall09, SnookerBall10, SnookerBall11, SnookerBall12, SnookerBall13, SnookerBall14, SnookerBall15, SnookerBallWhite},
	}

	r := rand.New(rand.NewSource(99))
	var snookerballs []*scene.Sphere
	for j := 0; j < len(balls); j++ {
		for i := 0; i < len(balls[j]); i++ {
			snookerball := NewSnookerBall(balls[j][i])
			snookerball.RotateX(snookerball.Bounds().Center(), util.DegToRad(-10)) // Tilt the ball number slightly upwards

			xRotAngle := maxRotation * (r.Float64()*2 - 1)
			yRotAngle := maxRotation * (r.Float64()*2 - 1)
			zRotAngle := maxRotation * (r.Float64()*2 - 1)
			snookerball.RotateZ(snookerball.Bounds().Center(), zRotAngle)
			snookerball.RotateX(snookerball.Bounds().Center(), xRotAngle)
			snookerball.RotateY(snookerball.Bounds().Center(), yRotAngle)

			x := 2.25 * snookerball.Radius * (float64(i) - float64(len(balls[j])-1)/2.0)
			z := ((float64(j) - 0.5) / 0.5) * snookerball.Radius * 2.5

			ballPerfectPosition := vec3.T{x, 0, z}
			displacementAngle := math.Pi * 2 * r.Float64()
			ballPosition := ballPerfectPosition.Added(&vec3.T{ballDisplacementRadius * math.Cos(displacementAngle), 0, ballDisplacementRadius * math.Sin(displacementAngle)})

			snookerball.Translate(&ballPosition)
			snookerballs = append(snookerballs, snookerball)
		}
	}

	dice := obj.NewDice(3)
	dice.RotateY(&vec3.Zero, math.Pi)
	dice.RotateY(&vec3.Zero, math.Pi/4)
	dice.RotateX(&vec3.Zero, math.Pi/6)
	dice.RotateY(&vec3.Zero, math.Pi*3/8)
	dice.Translate(&vec3.T{0, -dice.Bounds.Ymin, -15})

	// steelSphere := scene.NewSphere(&vec3.T{0, 5.7, -10}, 5.7, scene.NewMaterial().M(0.9, 0.15))

	snookerBallsNode := scene.NewSceneNode().S(snookerballs...).FS(dice) // .S(steelSphere)

	scn := scene.NewSceneNode().
		S(environmentSphere).
		SN(snookerBallsNode).
		FS(poolTable, lamp2)

	//animationStep := 1.0 / float64(amountAnimationFrames)
	for animationFrameIndex := 0; animationFrameIndex < amountAnimationFrames; animationFrameIndex++ {
		// animationProgress := float64(animationFrameIndex) * animationStep

		cameraOrigin := &vec3.T{0, 9, -40}
		focusPoint := dice.Bounds.Center() // vec3.T{0, 2.5, -6}
		focusPoint[1] = dice.Bounds.Ymax * 0.75
		camera := scene.NewCamera(cameraOrigin, focusPoint, amountSamples, magnification).
			A(apertureSize, nil).
			D(6)

		fi := -1
		if fi > 1 {
			fi = animationFrameIndex
		}

		frame := scene.NewFrame(animation.AnimationName, fi, camera, scn)
		animation.AddFrame(frame)
	}

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}

func NewSnookerBall(ball SnookerBall) *scene.Sphere {
	// diameter := 5.25 // Snooker (cm)
	// diameter := 5.4  // Bumper Pool (cm)
	// diameter := 5.25 // Carom (Billiard) Balls (cm)
	diameter := 5.7 // Pool (Pocket Billiard) Balls (cm)

	radius := diameter / 2
	textureFilename, err := floatimage.EmptyPlaceholderImage(fmt.Sprintf("textures/snooker/wpi/%s_wpi.png", ball))
	if err != nil {
		panic(err)
	}

	ballMaterial := scene.NewMaterial().
		N(fmt.Sprintf("snooker ball %s", ball)).
		M(0.015, 0.1).
		T(0.0, true, scene.RefractionIndex_AcrylicPlastic).
		//SP(textureFilename, &vec3.T{0, radius, 0}, vec3.T{0, 0, diameter}, vec3.T{0, radius, 0})
		//PP(textureFilename, &vec3.T{-radius, 0, 0}, vec3.T{diameter, 0, 0}, vec3.T{0, diameter, 0})
		CP(textureFilename, &vec3.T{0, 0, 0}, vec3.UnitZ, vec3.T{0, diameter, 0}, false)

	return scene.NewSphere(&vec3.T{0, radius, 0}, radius, ballMaterial)
}
