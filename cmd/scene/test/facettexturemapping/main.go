package main

import (
	"fmt"
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"
	"pathtracer/internal/pkg/util"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "facettexturemapping"

var imageWidth = 600
var imageHeight = 400
var magnification = 1.0 // 0.5

var amountSamples = 128 * 4

var skyDomeEmission = 1.0

var maxRayDepth = 3

func main() {
	//environmentTexture := floatimage.LoadOrPanic("textures/equirectangular/dimples.png")
	environmentTexture := floatimage.LoadOrPanic("textures/equirectangular/wirebox 6192x3098.png")

	environmentDome := scene.NewSphere(&vec3.T{0, 0, 0}, 200*100, scene.NewMaterial().N("environment dome").
		E(color.NewColorGrey(1.0), skyDomeEmission, true).
		SP(environmentTexture, &vec3.T{0, 0, 0}, vec3.T{1, 0, 0}, vec3.T{0, 1, 0})).N("environment dome")
	environmentDome.RotateY(&vec3.Zero, util.DegToRad(-20))

	// transparencyImage := floatimage.LoadOrPanic("textures/test/test_alpha_transparency.png")
	// leafImage := floatimage.LoadOrPanic("textures/tree/Leaves0120_35_S_02.png")
	// keroseneFlameImage := floatimage.LoadOrPanic("textures/misc/kerosenelamp/kerosenelamp_flame_wave.png")

	// testSquareFacets := obj.NewSquare(obj.SquareTypeXYPlane, transparencyImage)
	// leafSquareFacets := obj.NewSquare(obj.SquareTypeXYPlane, leafImage)
	// flameSquareFacets := obj.NewSquare(obj.SquareTypeXYPlane, keroseneFlameImage)

	testImage1 := floatimage.LoadOrPanic("textures/test/uv.png")
	testImage2 := floatimage.LoadOrPanic("textures/test/christian.png")
	testImage3 := floatimage.LoadOrPanic("textures/test/interpolation_nearest.png")

	square1Facets := obj.NewSquare(obj.SquareTypeXYPlane, testImage1)
	square2Facets := obj.NewSquare(obj.SquareTypeXYPlane, testImage2)
	square3Facets := obj.NewSquare(obj.SquareTypeXYPlane, testImage3)

	square1TextureMaterial := scene.NewMaterial().N("square 1 material")
	square2TextureMaterial := scene.NewMaterial().N("square 2 material").T(0.0, false, 1.0)
	square3TextureMaterial := scene.NewMaterial().N("square 3 material").T(0.0, false, 1.0).E(color.White, 1.0, false)

	square1 := &scene.FacetStructure{Facets: square1Facets, Material: square1TextureMaterial}
	square2 := &scene.FacetStructure{Facets: square2Facets, Material: square2TextureMaterial}
	square3 := &scene.FacetStructure{Facets: square3Facets, Material: square3TextureMaterial}

	square2.RotateY(&vec3.T{0, 0, 0}, math.Pi) // rotate test image "[F]" along the y-axis so it ends up to the left showing its backside
	square3.RotateX(&vec3.T{0, 0, 0}, math.Pi) // rotate test image "flame" along the x-axis so it ends up to the right bottom, showing its backside

	square1.ScaleUniform(&vec3.T{0, 0, 0}, 30)
	square2.ScaleUniform(&vec3.T{0, 0, 0}, 30)
	square3.ScaleUniform(&vec3.T{0, 0, 0}, 30)

	scn := scene.NewSceneNode().S(environmentDome).FS(square1, square2, square3)

	cameraOrigin := &vec3.T{0, 0, -50}
	cameraOrigin.Scale(3)
	focusPoint := &vec3.T{0, 0, 0}

	viewVector := focusPoint.Subed(cameraOrigin)
	focusDistance := viewVector.Length()

	camera := scene.NewCamera(cameraOrigin, focusPoint, amountSamples, magnification).
		F(focusDistance).
		D(maxRayDepth)

	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, true)
	frame := scene.NewFrame(animation.AnimationName, -1, camera, scn)
	animation.AddFrame(frame)

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}
