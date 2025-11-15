package main

import (
	"fmt"
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj"
	anm "pathtracer/internal/pkg/renderfile"
	scn "pathtracer/internal/pkg/scene"
	"pathtracer/internal/pkg/util"

	"github.com/ungerik/go3d/float64/vec3"
)

var amountAnimationFrames = 300
var amountSamples = 256

var apertureSize = 1.5

var width = 800
var height = 600
var magnification = 0.25

func main() {
	size := 800.0 // 100 cm

	animation := scn.NewAnimation("heightmap_test", width, height, magnification, false, true)

	landscapeMaterial := scn.NewMaterial().N("Ground material").
		T(0.0, false, scn.RefractionIndex_Quartz).
		PP(floatimage.Load("textures/ground/soil-cracked-bright.png"), &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(75), vec3.UnitZ.Scaled(75))
	landscape := obj.NewHeightMap("textures/height maps/test_heightmap.png", vec3.T{size, size / 7, size})
	landscape.UpdateVertexNormalsWithThreshold(false, 0)
	//landscape.UpdateVertexNormalsWithThreshold(false, 75)
	landscape.Material = landscapeMaterial

	groundDisc := scn.NewDisc(&vec3.Zero, &vec3.UnitY, 10*1000, landscapeMaterial)

	skyDomeOrigin := &vec3.T{0, 0, 0}
	skyDomeMaterial := scn.NewMaterial().
		E(color.White, 3, true).
		SP(floatimage.Load("textures/equirectangular/sunset horizon 2800x1400.jpg"), skyDomeOrigin, vec3.T{-0.2, 0, -1}, vec3.T{0, 1, 0})
	skyDome := scn.NewSphere(skyDomeOrigin, 10*1000, skyDomeMaterial).N("Sky dome")

	animationStartIndex := 0
	animationEndIndex := amountAnimationFrames - 1

	cameraStartAngle := util.DegToRad(-90)
	lightStartAngle := util.DegToRad(-45)

	for frameIndex := animationStartIndex; frameIndex <= animationEndIndex; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountAnimationFrames)
		angle := animationProgress * util.DegToRad(360*0.75)

		lamp := scn.NewSphere(&vec3.T{0.75 * size * math.Cos(lightStartAngle+angle), 400, 0.75 * size * math.Sin(lightStartAngle+angle)}, 200.0, scn.NewMaterial().E(color.White, 12, true))

		cameraOrigin := vec3.T{size * math.Cos(cameraStartAngle), size * 0.75, size * math.Sin(cameraStartAngle)}
		focusDistance := (&vec3.T{0, landscape.Bounds.SizeY() / 2, 0}).Sub(&cameraOrigin).Length()
		viewPoint := vec3.T{0, landscape.Bounds.SizeY() / 2, 0}
		camera := scn.NewCamera(&cameraOrigin, &viewPoint, amountSamples, magnification).A(apertureSize, nil).F(focusDistance)

		scene := scn.NewSceneNode().FS(landscape).S(lamp, skyDome).D(groundDisc)

		frame := scn.NewFrame(animation.AnimationName, frameIndex, camera, scene)

		animation.AddFrame(frame)
	}

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := anm.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}
