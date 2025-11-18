package main

import (
	"fmt"
	"math"
	dmd "pathtracer/cmd/obj/diamond/pkg/diamond"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	anm "pathtracer/internal/pkg/renderfile"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "diamondsR4ever"

var amountAnimationFrames = 60

var imageWidth = 800
var imageHeight = 800
var magnification = 1.0

var amountSamples = 128 * 8

var viewPlaneDistance = 1000.0
var cameraDistanceFactor = 1.0

func main() {
	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, true, false)

	roomScale := 100.0
	// cornellBox := getCornellBox(cornellBoxFilenamePath, 100.0)

	environmentSphere := scn.NewSphere(&vec3.T{0, 0, 0}, 1000*1000*1000, scn.NewMaterial().
		E(color.White, 1, true).
		SP(floatimage.Load("textures/planets/environmentmap/Stellarium3.jpeg"), &vec3.T{0, 0, 0}, vec3.T{1, 0, 0}, vec3.T{0, 1, 0})).N("sky dome")

	lamp1 := scn.NewSphere(&vec3.T{roomScale * 2.0, roomScale * 0.5, -roomScale * 1.0}, roomScale*0.8, scn.NewMaterial().E(color.White, 20, true)).N("Lamp1")

	lamp2 := scn.NewSphere(&vec3.T{-roomScale * 2.0, roomScale * 0.5, -roomScale * 1.0}, roomScale*0.8, scn.NewMaterial().E(color.White, 9, true)).N("Lamp2")

	diamondMaterial := scn.NewMaterial().
		N("diamond").
		C(color.NewColor(1.0, 0.95, 0.8)).
		T(1.0, true, scn.RefractionIndex_Diamond).
		M(0.3, 0.0)

	d := dmd.Diamond{
		GirdleDiameter:                         1.00, // 100.0%
		GirdleHeightRelativeGirdleDiameter:     0.03, //   3.0%
		CrownAngleDegrees:                      34.0, //  34.0°
		TableFacetSizeRelativeGirdle:           0.56, //  56.0%
		StarFacetSizeRelativeCrownSide:         0.55, //  55.0%
		PavilionAngleDegrees:                   41.1, //  41.1°
		LowerHalfFacetSizeRelativeGirdleRadius: 0.77, //  77.0%
	}

	diamond2 := dmd.NewDiamondRoundBrilliantCut(d, 75.0, *diamondMaterial)
	diamond2.UpdateBounds()
	floorLevel := diamond2.Bounds.Ymin

	floorMaterial := scn.NewMaterial().M(0.8, 0.1).PP(floatimage.Load("textures/marble/white_marble.png"), &vec3.T{0, 0, 0}, vec3.T{roomScale * 1.5 * 2, 0, 0}, vec3.T{0, 0, roomScale * 1.5 * 2})
	floor := scn.NewDisc(&vec3.T{0, floorLevel, 0}, &vec3.T{0, 1, 0}, roomScale*1.5, floorMaterial).N("Floor")

	animationStep := 1.0 / float64(amountAnimationFrames)
	for animationFrameIndex := 0; animationFrameIndex < amountAnimationFrames; animationFrameIndex++ {
		animationProgress := float64(animationFrameIndex) * animationStep

		diamond := dmd.NewDiamondRoundBrilliantCut(d, 75.0, *diamondMaterial)
		diamond.RotateY(&vec3.Zero, animationProgress*(math.Pi*2.0/8.0))
		diamond.RotateZ(&vec3.Zero, -15.0*animationStep*(math.Pi*2.0/8.0))
		diamond.RotateY(&vec3.Zero, 10.0*animationStep*(math.Pi*2.0/8.0))
		diamond.UpdateBounds()

		cameraOrigin := vec3.T{20, roomScale * 0.25, -100}
		cameraOrigin.Scale(cameraDistanceFactor)
		focusPoint := vec3.T{0, 0, 0}

		scene := scn.NewSceneNode().
			S(lamp1, lamp2, environmentSphere).
			D(floor).
			FS(diamond)

		camera := scn.NewCamera(&cameraOrigin, &focusPoint, amountSamples, magnification).V(viewPlaneDistance)

		frame := scn.NewFrame(animation.AnimationName, animationFrameIndex, camera, scene)
		animation.AddFrame(frame)
	}

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := anm.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}
