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
		E(color.White, skyDomeEmission, true).
		SP(floatimage.Load("textures/equirectangular/dimples.png"), &vec3.T{0, 0, 0}, vec3.T{1, 0, 0}, vec3.T{0, 1, 0})).N("sky dome")
	skyDome.RotateY(&vec3.Zero, util.DegToRad(-20))

	frontSideFacets := obj.NewSquare(obj.XYPlane, true)
	backSideFacets := obj.NewSquare(obj.XYPlane, true)

	leafMaterial1 := scene.NewMaterial().C(color.NewColorRGBA(0.85, 0.8, 0.9, 1.0)).TP(floatimage.Load("textures/test/test_alpha_transparency.png"))
	leafMaterial2 := scene.NewMaterial().C(color.NewColorRGBA(0.8, 0.65, 0.9, 1.0)).TP(floatimage.Load("textures/tree/Leaves0120_35_S_02.png")).T(0.05, false, 1.0)
	//leafMaterial := scene.NewMaterial().TP("Leaves0120_35_S.png")

	frontSide := &scene.FacetStructure{Facets: frontSideFacets, Material: leafMaterial1}
	backSide := &scene.FacetStructure{Facets: backSideFacets, Material: leafMaterial2}

	frontSide.RotateY(&vec3.T{0, 0, 0}, math.Pi) // rotate test image "[F]" along y-axis so it ends up to the left showing its backside
	backSide.ScaleUniform(&vec3.T{0, 0, 0}, 40)
	frontSide.ScaleUniform(&vec3.T{0, 0, 0}, 40)

	scn := scene.NewSceneNode().S(skyDome).FS(frontSide, backSide)

	cameraOrigin := &vec3.T{0, 20, -50}
	cameraOrigin.Scale(3)
	focusPoint := &vec3.T{0, 20, 0}

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
