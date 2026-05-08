package main

import (
	"fmt"
	"math"
	"math/rand"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/fileformat/wavefront"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "aperture_shape_test2"

var amountFrames = 120

var amountSamples = 200 // 200
var apertureRadius = 60.0

var viewPlaneDistance = 600.0

var imageWidth = 600
var imageHeight = 400
var magnification = 1.0

var lampEmissionFactor = 0.25
var sphereEmissionFactor = 3.0
var amountSpheres = 9
var sphereRadius = 8.0
var distanceFactor = 20.0

var sphereSpread = 80.0

func main() {
	cornellBox := GetCornellBox(&vec3.T{500, 300, 500}, lampEmissionFactor) // cm, as units. I.e. a 5x3x5m room

	quadrantSpheres := make([][]*scene.Sphere, 4)
	for i := range quadrantSpheres {
		quadrantSpheres[i] = make([]*scene.Sphere, 0)
	}

	for sz := 0; sz < amountSpheres; sz++ {
		for sx := 0; sx < amountSpheres; sx++ {
			positionOffsetX := sphereSpread * float64(sx+1)
			positionOffsetZ := sphereSpread * float64(sz)

			r := 0.4 + rand.Float64()*0.6
			g := 0.4 + rand.Float64()*0.6
			b := 0.4 + rand.Float64()*0.6

			sphereColor := color.NewColor(r, g, b)
			sphereMaterial := scene.NewMaterial().C(sphereColor).E(sphereColor, sphereEmissionFactor, true)

			for quad := 0; quad < 4; quad++ {
				sphere := scene.NewSphere(&vec3.T{positionOffsetX, sphereRadius * 3.0, positionOffsetZ}, sphereRadius, sphereMaterial)
				sphere.RotateY(&vec3.Zero, float64(quad)*math.Pi/2)

				quadrantSpheres[quad] = append(quadrantSpheres[quad], sphere)
			}
		}
	}

	q1 := scene.NewSceneNode().S(quadrantSpheres[0]...)
	q2 := scene.NewSceneNode().S(quadrantSpheres[1]...)
	q3 := scene.NewSceneNode().S(quadrantSpheres[2]...)
	q4 := scene.NewSceneNode().S(quadrantSpheres[3]...)

	scn := scene.NewSceneNode().FS(cornellBox).SN(q1, q2, q3, q4)

	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)

	for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountFrames)

		camera := getCamera(animationProgress, "textures/aperture/letter_F.png")
		frame := scene.NewFrame(animationName, frameIndex, camera, scn)
		animation.AddFrame(frame)
	}

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}

func GetCornellBox(scale *vec3.T, lightIntensityFactor float64) *scene.FacetStructure {
	var cornellBoxFilename = "cornellbox.obj"
	var cornellBoxFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/obj/" + cornellBoxFilename

	cornellBox := wavefront.ReadFacetStructureOrPanic(cornellBoxFilenamePath)

	cornellBox.Scale(&vec3.Zero, scale)
	cornellBox.ClearMaterials()

	cornellBox.Material = scene.NewMaterial().N("cornell box").C(color.NewColorGrey(0.95))

	lampMaterial := scene.NewMaterial().N("Lamp").E(color.White, lightIntensityFactor, true)

	cornellBox.GetFirstObjectBySubstructureName("Lamp_1_-_left_away").Material = lampMaterial
	cornellBox.GetFirstObjectBySubstructureName("Lamp_2_-_left_close").Material = lampMaterial
	cornellBox.GetFirstObjectBySubstructureName("Lamp_3_-_right_away").Material = lampMaterial
	cornellBox.GetFirstObjectBySubstructureName("Lamp_4_-_right_close").Material = lampMaterial

	projectionZoom := 0.33
	floorProjection := scene.NewParallelImageProjection(floatimage.LoadOrPanic("textures/floor/7451-diffuse 02 low contrast.png"), &vec3.T{-scale[0] * projectionZoom / 8, 0, -scale[0] * projectionZoom / 8}, vec3.UnitX.Scaled(scale[0]*projectionZoom), vec3.UnitZ.Scaled(scale[0]*projectionZoom))
	floorMaterial := *cornellBox.Material
	floorMaterial.M(0.0, 1.0)
	floorMaterial.P(&floorProjection)
	cornellBox.GetFirstObjectBySubstructureName("Floor").Material = &floorMaterial

	return cornellBox
}

func getCamera(animationProgress float64, apertureShape string) *scene.Camera {
	cameraDistance := 0.8
	cameraOrigin := vec3.T{0, distanceFactor * 25.0 * cameraDistance, -60 * distanceFactor * cameraDistance}
	viewPoint := vec3.T{0, sphereRadius, -distanceFactor * 18.0}

	focusDistance := vec3.Distance(&cameraOrigin, &vec3.T{0, sphereRadius, -distanceFactor * 12.0})

	// Animation
	angle := (math.Pi / 2.0) * animationProgress
	scene.RotateY(&cameraOrigin, &vec3.Zero, angle)
	scene.RotateY(&viewPoint, &vec3.Zero, angle)

	return scene.NewCamera(&cameraOrigin, &viewPoint, amountSamples, magnification).
		V(viewPlaneDistance).
		A(apertureRadius, floatimage.LoadOrPanic(apertureShape)).
		F(focusDistance)
}
