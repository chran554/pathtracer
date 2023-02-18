package main

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"math"
	"math/rand"
	"os"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"
)

var animationName = "aperture_shape_test2"

var amountFrames = 90

var amountSamples = 200
var apertureRadius = 50.0

var viewPlaneDistance = 600.0

var imageWidth = 1000
var imageHeight = 800
var magnification = 0.4

var lampEmissionFactor = 1.5
var sphereEmissionFactor = 6.0
var amountSpheres = 6
var sphereRadius = 8.0
var distanceFactor = 14.0

func main() {
	cornellBox := GetCornellBox(&vec3.T{500, 300, 500}, lampEmissionFactor) // cm, as units. I.e. a 5x3x5m room

	sphereSpread := distanceFactor * 6.0

	quadrantSpheres := make([][]*scn.Sphere, 4)
	for i := range quadrantSpheres {
		quadrantSpheres[i] = make([]*scn.Sphere, 0)
	}

	for sz := 0; sz < amountSpheres; sz++ {
		for sx := 0; sx < amountSpheres; sx++ {
			positionOffsetX := sphereSpread * float64(sx+1)
			positionOffsetZ := sphereSpread * float64(sz)

			r := 0.5 + rand.Float64()*0.5
			g := 0.5 + rand.Float64()*0.5
			b := 0.5 + rand.Float64()*0.5

			sphereColor := color.NewColor(r, g, b)
			sphereMaterial := scn.NewMaterial().C(sphereColor).E(sphereColor, sphereEmissionFactor, true)

			for quad := 0; quad < 4; quad++ {
				sphere := scn.NewSphere(&vec3.T{positionOffsetX, sphereRadius * 3.0, positionOffsetZ}, sphereRadius, sphereMaterial)
				sphere.RotateY(&vec3.Zero, float64(quad)*math.Pi/2)

				quadrantSpheres[quad] = append(quadrantSpheres[quad], sphere)
			}
		}
	}

	scene := scn.NewSceneNode().
		FS(cornellBox).
		S(quadrantSpheres[0]...).
		S(quadrantSpheres[1]...).
		S(quadrantSpheres[2]...).
		S(quadrantSpheres[3]...)

	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, false)

	for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountFrames)

		camera := getCamera(animationProgress)
		frame := scn.NewFrame(animationName, frameIndex, camera, scene)
		animation.AddFrame(frame)
	}

	anm.WriteAnimationToFile(animation, false)
}

func GetCornellBox(scale *vec3.T, lightIntensityFactor float64) *scn.FacetStructure {
	var cornellBoxFilename = "cornellbox.obj"
	var cornellBoxFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/obj/" + cornellBoxFilename

	cornellBoxFile, err := os.Open(cornellBoxFilenamePath)
	if err != nil {
		message := fmt.Sprintf("ouupps, something went wrong loading file: '%s'\n%s\n", cornellBoxFilenamePath, err.Error())
		panic(message)
	}
	defer cornellBoxFile.Close()

	cornellBox, err := obj.Read(cornellBoxFile)
	cornellBox.Scale(&vec3.Zero, scale)
	cornellBox.ClearMaterials()

	cornellBox.Material = scn.NewMaterial().N("cornell box").C(color.NewColorGrey(0.95))

	lampMaterial := scn.NewMaterial().N("Lamp").E(color.White, lightIntensityFactor, true)

	cornellBox.GetFirstObjectByName("Lamp_1").Material = lampMaterial
	cornellBox.GetFirstObjectByName("Lamp_2").Material = lampMaterial
	cornellBox.GetFirstObjectByName("Lamp_3").Material = lampMaterial
	cornellBox.GetFirstObjectByName("Lamp_4").Material = lampMaterial

	projectionZoom := 0.33
	floorProjection := scn.NewParallelImageProjection("textures/tilesf4.jpeg", &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(scale[0]*projectionZoom), vec3.UnitZ.Scaled(scale[0]*projectionZoom))
	floorMaterial := *cornellBox.Material
	floorMaterial.Projection = &floorProjection
	cornellBox.GetFirstObjectByName("Floor_2").Material = &floorMaterial

	return cornellBox
}

func getCamera(animationProgress float64) *scn.Camera {
	cameraOrigin := vec3.T{0, distanceFactor * 15.0, -500}
	focusPoint := vec3.T{0, distanceFactor * 2.0, -distanceFactor * 22.0}

	// Animation
	angle := (math.Pi / 2.0) * animationProgress
	scn.RotateY(&cameraOrigin, &vec3.Zero, angle)
	scn.RotateY(&focusPoint, &vec3.Zero, angle)

	return scn.NewCamera(&cameraOrigin, &focusPoint, amountSamples, magnification).
		V(viewPlaneDistance).
		A(apertureRadius, "")
}
