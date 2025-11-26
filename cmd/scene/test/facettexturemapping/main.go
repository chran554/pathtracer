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

var amountSamples = 128

var skyDomeEmission = 1.0

var maxRayDepth = 3

func main() {
	//environmentTexture := floatimage.Load("textures/equirectangular/dimples.png")
	environmentTexture := floatimage.Load("textures/equirectangular/wirebox 6192x3098.png")

	environmentDome := scene.NewSphere(&vec3.T{0, 0, 0}, 200*100, scene.NewMaterial().N("environment dome").
		E(color.NewColorGrey(1.0), skyDomeEmission, true).
		SP(environmentTexture, &vec3.T{0, 0, 0}, vec3.T{1, 0, 0}, vec3.T{0, 1, 0})).N("environment dome")
	environmentDome.RotateY(&vec3.Zero, util.DegToRad(-20))

	testSquareFacets := obj.NewSquare(obj.XYPlane, true)
	leafSquareFacets := obj.NewSquare(obj.XYPlane, true)
	flameSquareFacets := obj.NewSquare(obj.XYPlane, true)

	transparencyTestImage := floatimage.Load("textures/test/test_alpha_transparency.png")
	leafImage := floatimage.Load("textures/tree/Leaves0120_35_S_02.png")
	keroseneFlameTextureImage := floatimage.Load("textures/misc/kerosenelamp/kerosenelamp_flame_wave.png")

	testTextureMaterial := scene.NewMaterial().N("test material").C(color.NewColorRGBA(1.0, 1.0, 1.0, 1.0)).TP(transparencyTestImage)

	leafTextureMaterial := scene.NewMaterial().N("leaf material").C(color.NewColorRGBA(1.0, 1.0, 1.0, 1.0)).TP(leafImage).
		T(0.0, false, 1.0)

	flameTextureMaterial := scene.NewMaterial().N("flame material").C(color.NewColorRGBA(1.0, 1.0, 1.0, 1.0)).TP(keroseneFlameTextureImage).
		T(0.0, false, 1.0).E(color.White, 1.0, false)

	testSquare := &scene.FacetStructure{Facets: testSquareFacets, Material: testTextureMaterial}
	leafSquare := &scene.FacetStructure{Facets: leafSquareFacets, Material: leafTextureMaterial}
	flameSquare := &scene.FacetStructure{Facets: flameSquareFacets, Material: flameTextureMaterial}

	testSquare.RotateY(&vec3.T{0, 0, 0}, math.Pi)  // rotate test image "[F]" along the y-axis so it ends up to the left showing its backside
	flameSquare.RotateX(&vec3.T{0, 0, 0}, math.Pi) // rotate test image "flame" along the x-axis so it ends up to the right bottom, showing its backside

	leafSquare.ScaleUniform(&vec3.T{0, 0, 0}, 30)
	testSquare.ScaleUniform(&vec3.T{0, 0, 0}, 30)
	flameSquare.ScaleUniform(&vec3.T{0, 0, 0}, 30)

	scn := scene.NewSceneNode().S(environmentDome).FS(testSquare, leafSquare, flameSquare)

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
