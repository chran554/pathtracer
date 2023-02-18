package main

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"math"
	"os"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"
)

var animationName = "primitive_display"

var amountAnimationFrames = 72 * 2

var imageWidth = 800
var imageHeight = 200
var magnification = 1.0

var amountSamples = 256 * 2

var cameraOrigin = vec3.T{0, 150, -800}
var cameraDistanceFactor = 0.8
var viewPlaneDistance = 600.0
var cameraAperture = 10.0

var lightIntensityFactor = 5.0

func main() {
	width := int(float64(imageWidth) * magnification)
	height := int(float64(imageHeight) * magnification)

	// Keep image proportions to an even amount of pixel for mp4 encoding
	if width%2 == 1 {
		width++
	}
	if height%2 == 1 {
		height++
	}

	// Cornell box

	cornellBox := GetCornellBox(&vec3.T{500, 300, 500}, lightIntensityFactor) // cm, as units. I.e. a 5x3x5m room

	// Gopher

	gopher := GetGopher(&vec3.T{1, 1, 1})
	gopher.Translate(&vec3.T{0, -gopher.Bounds.Ymin, 0})
	gopher.ScaleUniform(&vec3.Zero, 2.0)
	gopher.RotateY(&vec3.Zero, math.Pi*5.0/6.0)
	gopher.Translate(&vec3.T{800, 0, 800})

	gopherLightMaterial := scn.NewMaterial().E(color.Color{R: 6, G: 5.3, B: 4.5}, 1.0, true)
	gopherLight := scn.NewSphere(&vec3.T{850, 50, 600}, 70.0, gopherLightMaterial).N("Gopher light")

	// Diamond

	diamond := GetDiamond(&vec3.T{2, 2, 2})
	diamond.Translate(&vec3.T{0, -diamond.Bounds.Ymin, 0})
	diamond.RotateX(&vec3.Zero, 0.7173303226) // Pavilion angle, lay diamond on the side on the floor
	diamond.RotateY(&vec3.Zero, -math.Pi*3.0/6.0)
	diamond.Translate(&vec3.T{600, 0, 700})
	//diamond.Translate(&vec3.T{0, 0, -200})
	diamond.Material = &scn.Material{Color: &color.Color{R: 0.85, G: 0.85, B: 0.75}, Emission: &color.Color{R: 0.1, G: 0.08, B: 0.05}, Glossiness: 0.05, Roughness: 0.0, Transparency: 0.0}

	podiumHeight := 30.0
	podiumWidth := 200.0
	interPodiumDistance := 400.0

	triangleSize := 80.0 * 2.0
	sphereRadius := 80.0
	discRadius := 80.0

	sphereLocation := vec3.T{-interPodiumDistance, 0, 0}
	triangleLocation := vec3.T{interPodiumDistance, 0, 0}
	discLocation := vec3.T{0, 0, 0}

	podiumMaterial := scn.Material{Color: &color.Color{R: 0.9, G: 0.9, B: 0.9}, Roughness: 1.0}

	// Sphere

	spherePodium := obj.NewBox(obj.BoxCenteredYPositive)
	spherePodium.Material = &podiumMaterial
	spherePodium.Scale(&vec3.Zero, &vec3.T{podiumWidth, podiumHeight, podiumWidth})
	spherePodium.Translate(&sphereLocation)

	sphereMaterial := scn.NewMaterial().C(color.Color{R: 0.80, G: 1.00, B: 0.80}).M(0.3, 0.2)
	sphere := scn.NewSphere(&vec3.T{0, 0, 0}, sphereRadius, sphereMaterial).N("Sphere primitive")
	sphere.Translate(&vec3.T{0.0, podiumHeight + sphereRadius, 0.0})
	sphere.Translate(&sphereLocation)

	// Triangle

	trianglePodium := obj.NewBox(obj.BoxCenteredYPositive)
	trianglePodium.Material = &podiumMaterial
	trianglePodium.Scale(&vec3.Zero, &vec3.T{podiumWidth, podiumHeight, podiumWidth})
	trianglePodium.Translate(&triangleLocation)

	triangle := trianglePrimitive()
	triangle.ScaleUniform(&vec3.Zero, triangleSize)
	triangle.RotateX(&vec3.Zero, -math.Pi/8.0)
	triangle.RotateY(&vec3.Zero, -math.Pi/4.0)
	triangle.Translate(&vec3.T{0.0, podiumHeight, 0.0})
	triangle.Translate(&triangleLocation)

	// Disc

	discPodium := obj.NewBox(obj.BoxCenteredYPositive)
	discPodium.Material = &podiumMaterial
	discPodium.Scale(&vec3.Zero, &vec3.T{podiumWidth, podiumHeight, podiumWidth})
	discPodium.Translate(&discLocation) // Move podium to location

	discMaterial := scn.NewMaterial().C(color.Color{R: 1.00, G: 0.80, B: 0.80}).M(0.05, 0.1)
	disc := scn.NewDisc(&vec3.T{0, 0, 0}, &vec3.T{0, 0, -1}, discRadius, discMaterial).N("Disc primitive")
	disc.RotateX(&vec3.Zero, -math.Pi/8.0)
	disc.RotateY(&vec3.Zero, -math.Pi/8.0)
	disc.Translate(&vec3.T{0.0, podiumHeight + discRadius, 0.0})
	disc.Translate(&discLocation) // Move disc to location

	scene := scn.NewSceneNode().
		S(sphere, gopherLight).
		D(disc).
		FS(cornellBox, triangle, spherePodium, discPodium, trianglePodium, gopher, diamond)

	animationStartIndex := 0
	animationEndIndex := amountAnimationFrames - 1

	animation := scn.NewAnimation(animationName, width, height, magnification, false)

	for frameIndex := animationStartIndex; frameIndex <= animationEndIndex; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountAnimationFrames)

		camera := getCamera(animationProgress, sphereRadius+podiumHeight)

		frame := scn.NewFrame(animationName, frameIndex, camera, scene)
		animation.AddFrame(frame)
	}

	anm.WriteAnimationToFile(animation, false)
}

func trianglePrimitive() *scn.FacetStructure {
	material := scn.Material{
		Color:      &color.Color{R: 0.80, G: 0.80, B: 1.00},
		Glossiness: 0.05,
		Roughness:  0.1,
	}
	triangleHeight := 1.0
	triangleWidth := triangleHeight / (2.0 * math.Cos(math.Pi/6.0))
	triangle := scn.Facet{
		Vertices: []*vec3.T{
			{0, 0, 0},              //  p3 *---* p2
			{triangleWidth, 1, 0},  //      \ /
			{-triangleWidth, 1, 0}, //       * p1
		},
	}
	facetStructure := scn.FacetStructure{
		SubstructureName: "Triangle primitive",
		Material:         &material,
		Facets:           []*scn.Facet{&triangle},
	}
	facetStructure.UpdateNormals()

	return &facetStructure
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
	cornellBox.UpdateBounds()
	fmt.Printf("Cornell box bounds: %+v\n", cornellBox.Bounds)

	cornellBox.Material = scn.NewMaterial().N("Cornell box default material").
		C(color.Color{R: 0.95, G: 0.95, B: 0.95})

	lampMaterial := scn.NewMaterial().N("Lamp").
		E(color.White, lightIntensityFactor, true)

	cornellBox.GetFirstObjectByName("Lamp_1").Material = lampMaterial
	cornellBox.GetFirstObjectByName("Lamp_2").Material = lampMaterial
	cornellBox.GetFirstObjectByName("Lamp_3").Material = lampMaterial
	cornellBox.GetFirstObjectByName("Lamp_4").Material = lampMaterial

	//backWallProjection := scn.NewParallelImageProjection("textures/wallpaper/anemone-rose-flower-eucalyptus-leaves-pampas-grass.png", vec3.Zero, vec3.UnitX.Scaled(scale[0]), vec3.UnitY.Scaled(scale[0]*0.66))
	//backWallMaterial := *cornellBox.Material
	//backWallMaterial.Projection = &backWallProjection
	//cornellBox.GetFirstObjectByName("Wall_back").Material = &backWallMaterial

	//sideWallProjection := scn.NewParallelImageProjection("textures/wallpaper/anemone-rose-flower-eucalyptus-leaves-pampas-grass.png", vec3.Zero, vec3.UnitZ.Scaled(scale[0]), vec3.UnitY.Scaled(scale[0]*0.66))
	//sideWallMaterial := *cornellBox.Material
	//sideWallMaterial.Projection = &sideWallProjection
	//cornellBox.GetFirstObjectByName("Wall_left").Material = &sideWallMaterial
	//cornellBox.GetFirstObjectByName("Wall_right").Material = &sideWallMaterial

	//floorProjection := scn.NewParallelImageProjection("textures/tilesf4.jpeg", vec3.Zero, vec3.UnitX.Scaled(scale[0]*0.5), vec3.UnitZ.Scaled(scale[0]*0.5))
	floorMaterial := *cornellBox.Material
	floorMaterial.Glossiness = 0.2
	floorMaterial.Roughness = 0.2
	//floorMaterial.Projection = &floorProjection
	cornellBox.ReplaceMaterial("Floor_2", &floorMaterial)

	return cornellBox
}

func GetGopher(scale *vec3.T) *scn.FacetStructure {
	var objFilename = "go_gopher_color.obj"
	var objFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/obj/" + objFilename

	objFile, err := os.Open(objFilenamePath)
	if err != nil {
		message := fmt.Sprintf("ouupps, something went wrong loading file: '%s'\n%s\n", objFilenamePath, err.Error())
		panic(message)
	}
	defer objFile.Close()

	obj, err := obj.Read(objFile)
	obj.Scale(&vec3.Zero, scale)
	// obj.ClearMaterials()
	obj.UpdateBounds()
	fmt.Printf("Gopher bounds: %+v\n", obj.Bounds)

	return obj
}

func GetDiamond(scale *vec3.T) *scn.FacetStructure {
	var objFilename = "diamond.obj"
	var objFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/obj/" + objFilename

	objFile, err := os.Open(objFilenamePath)
	if err != nil {
		fmt.Printf("ouupps, something went wrong loading file: '%s'\n%s\n", objFilenamePath, err.Error())
	}
	defer objFile.Close()

	obj, err := obj.Read(objFile)
	obj.Scale(&vec3.Zero, scale)
	// obj.ClearMaterials()
	obj.UpdateBounds()
	fmt.Printf("Diamond bounds: %+v\n", obj.Bounds)

	return obj
}

func getCamera(animationProgress float64, focusHeight float64) *scn.Camera {
	var cameraOrigin = cameraOrigin // vec3.T{0, 150, -800}
	var cameraFocus = vec3.T{0, focusHeight, 0}

	startAngle := -math.Pi / 2.0

	cameraOriginRadius := 600.0
	cameraOriginAnimationTranslation := vec3.T{
		math.Cos(startAngle-animationProgress*2.0*math.Pi) * cameraOriginRadius,
		0,
		math.Sin(startAngle-animationProgress*2.0*math.Pi) * cameraOriginRadius / 2.0,
	}
	cameraOrigin.Add(&cameraOriginAnimationTranslation)

	cameraFocusRadius := 200.0 // Same as inter podium distance
	cameraFocusAnimationTranslation := vec3.T{
		math.Cos(startAngle-animationProgress*2.0*math.Pi) * cameraFocusRadius,
		0,
		0,
	}
	cameraFocus.Add(&cameraFocusAnimationTranslation)

	origin := cameraOrigin.Scaled(cameraDistanceFactor)

	return scn.NewCamera(&origin, &cameraFocus, amountSamples, magnification).V(viewPlaneDistance).A(cameraAperture, "")
}
