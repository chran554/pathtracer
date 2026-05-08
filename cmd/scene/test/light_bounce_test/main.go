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

var animationName = "light_bounce_test"

var environmentRadius = 50000.0
var environmentEmissionFactor = 1.0 / 255.0

var amountFrames = 180

var imageWidth = 600
var imageHeight = 800
var magnification = 2.0

var amountSamples = 1024 * 8
var maxRecursion = 60

var lampScale = 15.0

func main() {
	var textureEnvironment = floatimage.LoadOrPanic("textures/equirectangular/sunset horizon 2800x1400.jpg")
	//var textureSoilCracked = floatimage.LoadOrPanic("textures/ground/soil-cracked.png")
	// var textureLightBox = floatimage.LoadOrPanic("textures/lights/lightboxtexture_2.0.png")

	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)

	// Sky dome
	skyDomeOrigin := vec3.T{0, 0, 0}
	skyDomeMaterial := scene.NewMaterial().
		E(color.White, environmentEmissionFactor, true).
		SP(textureEnvironment, &skyDomeOrigin, vec3.T{1, 0, 0}, vec3.T{0, 1, 0})
	skyDome := scene.NewSphere(&skyDomeOrigin, environmentRadius, skyDomeMaterial).N("sky dome")
	skyDome.RotateY(&vec3.Zero, util.DegToRad(0))

	// Floor
	floor := obj.NewSquareFacetStructure(obj.SquareTypeXZPlane, nil, true)
	floor.Scale(&vec3.Zero, &vec3.T{1000, 1, 1000})
	floor.Material = scene.NewMaterial().N("floor").C(color.NewColorGrey(0.5))

	wallDistance := 40.0
	wallLength := wallDistance * 10
	wall1 := obj.NewSquareFacetStructure(obj.SquareTypeXYPlane, nil, false)
	wall1.Scale(&vec3.Zero, &vec3.T{wallLength, wallDistance, 1})
	wall1.Translate(&vec3.T{-(wallLength + wallDistance/2), 0, +wallDistance / 2})

	wall2 := obj.NewSquareFacetStructure(obj.SquareTypeXYPlane, nil, false)
	wall2.Scale(&vec3.Zero, &vec3.T{wallLength, wallDistance, 1})
	wall2.Translate(&vec3.T{-wallLength + wallDistance, 0, -wallDistance / 2})

	wallMaterial := scene.NewMaterial().N("wall").C(color.NewColorGrey(0.9))
	wall1.Material = wallMaterial
	wall2.Material = wallMaterial

	var decorativeLamps []*scene.Sphere
	decorativeLampRadius := 5.0
	decorativeLampMaterial := scene.NewMaterial().N("lamp").E(color.KelvinTemperatureColor2(3000), 1, false)

	var wallDivisions []*scene.FacetStructure
	wallDivisionSpacing := wallDistance
	for wallDivisionIndex := 0; wallDivisionIndex < int(wallLength/wallDivisionSpacing); wallDivisionIndex++ {
		wallDivision := obj.NewSquareFacetStructure(obj.SquareTypeYZPlane, nil, false)
		wallDivision.Material = wallMaterial
		wallDivision.Scale(&vec3.Zero, &vec3.T{1, wallDistance, wallDistance / 2})

		decorativeLamp := scene.NewSphere(&vec3.T{0, decorativeLampRadius, 0}, decorativeLampRadius, decorativeLampMaterial).N("lamp")

		if wallDivisionIndex%2 == 0 {
			wallDivision.Translate(&vec3.T{0, 0, -wallDivisionSpacing / 2})
			decorativeLamp.Translate(&vec3.T{0, 0, -(wallDivisionSpacing + 2*decorativeLampRadius)})
		} else {
			wallDivision.Translate(&vec3.T{0, 0, 0})
			decorativeLamp.Translate(&vec3.T{0, 0, wallDivisionSpacing + 2*decorativeLampRadius})
		}

		wallDivision.Translate(&vec3.T{-1 * float64(wallDivisionIndex) * wallDivisionSpacing, 0, 0})
		decorativeLamp.Translate(&vec3.T{-1 * float64(wallDivisionIndex) * wallDivisionSpacing, 0, 0})

		if wallDivisionIndex == 0 {
			// Some extra distance to the first division wall to position it end at the end of the long wall
			wallDivision.Translate(&vec3.T{wallDivisionSpacing, 0, 0})
			wallDivision.Scale(&vec3.T{wallDivision.Bounds.Xmin, wallDivision.Bounds.Ymin, wallDivision.Bounds.Zmin}, &vec3.T{1, 1, 2})
		}

		wallDivisions = append(wallDivisions, wallDivision)
		decorativeLamps = append(decorativeLamps, decorativeLamp)
	}

	// Lamp
	lampRadius := 20.0
	lampMaterial := scene.NewMaterial().N("lamp").E(color.White, lampScale, true)
	lamp := scene.NewSphere(&vec3.T{0, lampRadius, lampRadius + wallDistance/2}, lampRadius, lampMaterial).N("lamp")

	for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountFrames)

		scn := scene.NewSceneNode().FS(floor).S(lamp).FS(wall1, wall2).FS(wallDivisions...).S(decorativeLamps...)

		rotationOrigin := &vec3.T{-65, 0, 0}

		// Camera
		cameraFocusPoint := &vec3.T{-100, 0, 0}
		scene.RotateY(cameraFocusPoint, rotationOrigin, animationProgress*util.DegToRad(360*animationProgress))
		cameraFocusPoint[2] = 0.0

		cameraHeight := 100.0 + 100.0*math.Sin(util.DegToRad(360*(animationProgress*2)-90)) + 100
		cameraOrigin := vec3.Zero.Added(&vec3.T{-250, cameraHeight, 0})
		scene.RotateY(&cameraOrigin, rotationOrigin, util.DegToRad(360*animationProgress))
		camera := scene.NewCamera(&cameraOrigin, cameraFocusPoint, amountSamples, magnification).D(maxRecursion)

		fi := -1
		if amountFrames > 1 {
			fi = frameIndex
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
