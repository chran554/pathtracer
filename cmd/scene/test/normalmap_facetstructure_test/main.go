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

var animationName = "normalmap_facetstructure_test"

var amountAnimationFrames = 120

var imageWidth = 640
var imageHeight = 640
var magnification = 0.5

var amountSamples = 512 // * 2 * 4 // * 2 * 15 // * 8 // * 15
var maxRecursionDepth = 2

var cameraAperture = 0.0
var cameraZoom = 1.0

var lampIntensity = 18.0

func main() {
	normalMapTexture, err := floatimage.EmptyPlaceholderImage("textures/normal maps/normal_map_example.png")
	if err != nil {
		panic(err)
	}

	lamp := scene.NewSphere(&vec3.T{}, 1.0, scene.NewMaterial().
		N("lamp").
		E(color.KelvinTemperatureColor2(5500), lampIntensity, true))

	lampOffsetFromBack := 4.0
	// lamp.Origin.Add(&vec3.T{-(wall.Bounds.SizeX()/2 + lamp.Radius), 0, -lampOffsetFromBack})
	wallLength := 6.0
	lampOffset := wallLength/2 + lamp.Radius
	lamp.Origin.Add(&vec3.T{-lampOffset, lampOffset, -lampOffsetFromBack})
	// lamp.RotateZ(&vec3.Zero, util.DegToRad(animationProgress*360))
	//lamp.Origin.Mul(&vec3.T{1.0, 0.5, 1.0})

	wall := createWall()

	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)

	cameraPosition := &vec3.T{0, 0, -8}

	cameraAim := wall.Bounds.Center()
	focusPosition := cameraAim
	cameraFocusVector := focusPosition.Subed(cameraPosition)

	camera := scene.NewCamera(cameraPosition, cameraAim, amountSamples, magnification).
		D(maxRecursionDepth).
		A(cameraAperture, nil).
		F(cameraFocusVector.Length()).
		V(550 * cameraZoom)

	for frameIndex := range amountAnimationFrames {
		animationProgress := float64(frameIndex) / float64(amountAnimationFrames)

		wall = createWall()
		wb := wall.Bounds
		wall.Material.PP(normalMapTexture, scene.TextureTypeNormalMap, &vec3.T{wb.Xmin, wb.Ymin, 0}, vec3.T{wb.SizeX(), 0, 0}, vec3.T{0, wb.SizeY(), 0})
		wall.RotateZ(&vec3.Zero, util.DegToRad(135))
		wall.RotateX(&vec3.Zero, util.DegToRad(35))
		wall.RotateY(&vec3.Zero, util.DegToRad(360*animationProgress))

		scn := scene.NewSceneNode().FS(wall).S(lamp)

		fi := -1
		if amountAnimationFrames > 1 {
			fi = frameIndex
		}

		frame := scene.NewFrame(animationName, fi, camera, scn)
		animation.AddFrame(frame)
	}

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err = renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}

func createWall() *scene.FacetStructure {
	colorWhite := color.NewColorGrey(0.95)

	wallMaterial := scene.NewMaterial().N("wall")
	wallMaterial.C(colorWhite)
	wallMaterial.M(0.4, 0.5)
	wallMaterial.FresnelMaxGlossiness = 0.15

	wall := obj.NewSquareFacetStructure(obj.SquareTypeXYPlane, nil, true)
	wall.Name = "wall"
	wall.Material = wallMaterial
	wall.RotateY(&vec3.Zero, math.Pi)
	wall.ScaleUniform(&vec3.Zero, 6)
	return wall
}
