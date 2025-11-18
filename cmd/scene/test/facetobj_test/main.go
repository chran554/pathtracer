package main

import (
	"fmt"
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj/wavefrontobj"
	anm "pathtracer/internal/pkg/renderfile"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/mat3"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "facetobj_test"

var amountImages = 180

var imageWidth = 1024 / 2
var imageHeight = 800 / 2
var magnification = 1.0 / 2

var amountSamples = 100

var objectFilename = "cube_smooth.obj"
var objectScale = 2.5
var objectStartAngle = 0.0

//var objectFilename = "human skull.obj"
//var objectScale = 10000.0 * 1.5
//var objectStartAngle = math.Pi // 180 degrees

//var environmentEnvironMap = "textures/equirectangular/canyon 3200x1600.jpeg"
//var environmentRadius = 100.0 * 10.0 // 10m (if 1 unit is 1 cm)
//var environmentEmissionFactor = float32(2.0)

var environmentRadius = 100.0 * 20.0 // 20m (if 1 unit is 1 cm)
var environmentEmissionFactor = 1.0

//var environmentEnvironMap = "textures/equirectangular/spruit_sunrise_2400x1200.jpeg"
//var environmentRadius = 100.0 * 400.0 // 200m (if 1 unit is 1 cm)
//var environmentEmissionFactor = float32(3.0)

var useLights = true
var lightColor = color.NewColor(1.0, 0.97, 0.95)
var lightOrigin = vec3.T{environmentRadius / 8, environmentRadius / 4, -environmentRadius / 1.5}
var lightRadius = environmentRadius / 14
var lightEmissionFactor = 2.0

var cameraDistance = 400.0
var cameraOrigin = vec3.T{cameraDistance / 1.0, cameraDistance * 4 / 5, -cameraDistance}
var viewPlaneDistance = 500.0

func main() {
	objectPath := "/Users/christian/projects/code/go/pathtracer/objects/"

	// objectFilename = "objects/lamppost.obj.3ds.obj"
	// objectFilename = "objects/Diamond.obj"
	// objectFilename = objectPath + "go_gopher_color.obj"
	objectFilename = objectPath + objectFilename
	// objectFilename = objectPath + "cube_smooth.obj"
	// objectFilename = objectPath + "cube.obj"
	// objectFilename = objectPath + "facet.obj"
	// objectFilename = objectPath + "triangle.obj"
	// objectFilename = "objects/go_gopher_high.obj"

	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, true, false)

	for imageIndex := 0; imageIndex < amountImages; imageIndex++ {
		fmt.Printf("\n\nCostructing frame %d\n", imageIndex)
		fmt.Printf("Reading file: %s\n", objectFilename)

		facetStructure := wavefrontobj.ReadOrPanic(objectFilename)

		facetStructure.UpdateBounds()
		fmt.Printf("Object in file \"%s\" has bounds %+v.\n", objectFilename, facetStructure.Bounds)

		facetStructure.ScaleUniform(&vec3.Zero, objectScale)
		facetStructure.RotateY(&vec3.Zero, objectStartAngle)

		scene := scn.NewSceneNode().FS(facetStructure)
		scene.UpdateBounds()
		scene.Bounds = nil

		if useLights {
			lampMaterial := scn.NewMaterial().C(lightColor).E(lightColor, lightEmissionFactor, true)
			lamp := scn.NewSphere(&lightOrigin, lightRadius, lampMaterial).N("Light")
			scene.S(lamp)
		}

		skyDomeOrigin := vec3.T{0, 0, 0}
		skyDomeMaterial := scn.NewMaterial().
			E(color.White, environmentEmissionFactor, true).
			SP(floatimage.Load("textures/equirectangular/white room 01 1836x918.png"), &skyDomeOrigin, vec3.UnitZ, vec3.UnitY)
		skyDome := scn.NewSphere(&skyDomeOrigin, environmentRadius, skyDomeMaterial).N("sky dome")
		scene.S(skyDome)

		animationProgress := float64(imageIndex) / float64(amountImages)
		heightFactor := math.Sin(2.0 * 2.0 * math.Pi * animationProgress)
		camera := getCamera(&cameraOrigin, facetStructure.Bounds.Center(), 2.0*math.Pi*animationProgress, heightFactor)

		frame := scn.NewFrame(animationName, imageIndex, camera, scene)

		animation.AddFrame(frame)
	}

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := anm.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}

func getCamera(cameraOrigin *vec3.T, focusPoint *vec3.T, yRotationAngle float64, heightFactor float64) *scn.Camera {
	rotationMatrix := mat3.T{}
	rotationMatrix.AssignYRotation(yRotationAngle)

	newCameraOrigin := vec3.T{cameraOrigin[0], cameraOrigin[1] * heightFactor, cameraOrigin[2]}
	newCameraOrigin = rotationMatrix.MulVec3(&newCameraOrigin)

	return scn.NewCamera(&newCameraOrigin, focusPoint, amountSamples, magnification).V(viewPlaneDistance)
}
