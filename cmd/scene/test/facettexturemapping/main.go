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
var magnification = 2.0 // 0.5

var amountSamples = 128

var skyDomeEmission = 1.0

var maxRayDepth = 3

func main() {
	skyDome := scene.NewSphere(&vec3.T{0, 0, 0}, 200*100, scene.NewMaterial().
		E(color.NewColorGrey(0.25), skyDomeEmission, true).
		SP(floatimage.Load("textures/equirectangular/dimples.png"), &vec3.T{0, 0, 0}, vec3.T{1, 0, 0}, vec3.T{0, 1, 0})).N("environment dome")
	skyDome.RotateY(&vec3.Zero, util.DegToRad(-20))

	testSquareFacets := obj.NewSquare(obj.XYPlane, true)
	leafSquareFacets := obj.NewSquare(obj.XYPlane, true)
	flameSquareFacets := obj.NewSquare(obj.XYPlane, true)

	testTextureMaterial := scene.NewMaterial().C(color.NewColorRGBA(0.85, 0.8, 0.9, 1.0)).TP(floatimage.Load("textures/test/test_alpha_transparency.png"))
	leafTextureMaterial := scene.NewMaterial().C(color.NewColorRGBA(0.8, 0.65, 0.9, 1.0)).TP(floatimage.Load("textures/tree/Leaves0120_35_S_02.png")).T(0.05, false, 1.0)
	flameTextureMaterial := scene.NewMaterial().C(color.NewColorRGBA(1.0, 1.0, 1.0, 1.0)).TP(floatimage.Load("textures/misc/kerosenelamp/kerosenelamp_flame_wave.png")).T(0.0, false, 1.0).E(color.White, 1.0, false)
	//leafMaterial := scene.NewMaterial().TP("Leaves0120_35_S.png")

	testSquare := &scene.FacetStructure{Facets: testSquareFacets, Material: testTextureMaterial}
	leafSquare := &scene.FacetStructure{Facets: leafSquareFacets, Material: leafTextureMaterial}
	flameSquare := &scene.FacetStructure{Facets: flameSquareFacets, Material: flameTextureMaterial}

	testSquare.RotateY(&vec3.T{0, 0, 0}, math.Pi)  // rotate test image "[F]" along the y-axis so it ends up to the left showing its backside
	flameSquare.RotateX(&vec3.T{0, 0, 0}, math.Pi) // rotate test image "flame" along the x-axis so it ends up to the right bottom, showing its backside

	leafSquare.ScaleUniform(&vec3.T{0, 0, 0}, 30)
	testSquare.ScaleUniform(&vec3.T{0, 0, 0}, 30)
	flameSquare.ScaleUniform(&vec3.T{0, 0, 0}, 30)

	scn := scene.NewSceneNode().S(skyDome).FS(testSquare, leafSquare, flameSquare)

	cameraOrigin := &vec3.T{0, 0, -50}
	cameraOrigin.Scale(3)
	focusPoint := &vec3.T{0, 0, 0}

	viewVector := focusPoint.Subed(cameraOrigin)
	focusDistance := viewVector.Length()

	camera := scene.NewCamera(cameraOrigin, focusPoint, amountSamples, magnification).
		F(focusDistance).
		D(maxRayDepth)

	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, true, true)
	frame := scene.NewFrame(animation.AnimationName, -1, camera, scn)
	animation.AddFrame(frame)

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}
