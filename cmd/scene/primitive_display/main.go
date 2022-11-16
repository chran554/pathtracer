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

var renderType = scn.Pathtracing

// var renderType = scn.Raycasting

var amountAnimationFrames = 72 // * 2

var imageWidth = 800
var imageHeight = 200
var magnification = 1.0

var maxRecursionDepth = 4
var amountSamples = 256 * 4 * 2

var cameraOrigin = vec3.T{0, 150, -800}
var cameraDistanceFactor = 0.8
var viewPlaneDistance = 600.0
var cameraAperture = 10.0

var lightIntensityFactor = 7.0 // 6.0

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

	cornellBox := GetCornellBox(&vec3.T{500, 300, 500}, float32(lightIntensityFactor)) // cm, as units. I.e. a 5x3x5m room

	// Gopher

	gopher := GetGopher(&vec3.T{1, 1, 1})
	gopher.Translate(&vec3.T{0, -gopher.Bounds.Ymin, 0})
	gopher.ScaleUniform(&vec3.Zero, 2.0)
	gopher.RotateY(&vec3.Zero, math.Pi*5.0/6.0)
	gopher.Translate(&vec3.T{800, 0, 800})

	gopherLightMaterial := scn.Material{Color: &color.White, Emission: &color.Color{R: 6, G: 5.3, B: 4.5}, Glossiness: 0.0, Roughness: 1.0, RayTerminator: true}
	gopherLight := scn.Sphere{Name: "Gopher light", Origin: &vec3.T{850, 50, 600}, Radius: 70.0, Material: &gopherLightMaterial}

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

	spherePodium := getBox()
	spherePodium.Material = &podiumMaterial
	spherePodium.Translate(&vec3.T{-0.5, 0, -0.5}) // Center podium around origin
	spherePodium.Scale(&vec3.Zero, &vec3.T{podiumWidth, podiumHeight, podiumWidth})
	spherePodium.Translate(&sphereLocation)

	sphere := spherePrimitive(sphereRadius)
	sphere.Translate(&vec3.T{0.0, podiumHeight + sphereRadius, 0.0})
	sphere.Translate(&sphereLocation)

	// Triangle

	trianglePodium := getBox()
	trianglePodium.Material = &podiumMaterial
	trianglePodium.Translate(&vec3.T{-0.5, 0, -0.5}) // Center podium around origin
	trianglePodium.Scale(&vec3.Zero, &vec3.T{podiumWidth, podiumHeight, podiumWidth})
	trianglePodium.Translate(&triangleLocation)

	triangle := trianglePrimitive()
	triangle.ScaleUniform(&vec3.Zero, triangleSize)
	triangle.RotateX(&vec3.Zero, -math.Pi/8.0)
	triangle.RotateY(&vec3.Zero, -math.Pi/4.0)
	triangle.Translate(&vec3.T{0.0, podiumHeight, 0.0})
	triangle.Translate(&triangleLocation)

	// Disc

	discPodium := getBox()
	discPodium.Material = &podiumMaterial
	discPodium.Translate(&vec3.T{-0.5, 0, -0.5}) // Center podium around origin
	discPodium.Scale(&vec3.Zero, &vec3.T{podiumWidth, podiumHeight, podiumWidth})
	discPodium.Translate(&discLocation) // Move podium to location

	disc := discPrimitive(sphereRadius)
	disc.RotateX(&vec3.Zero, -math.Pi/8.0)
	disc.RotateY(&vec3.Zero, -math.Pi/8.0)
	disc.Translate(&vec3.T{0.0, podiumHeight + discRadius, 0.0})
	disc.Translate(&discLocation) // Move disc to location

	scene := scn.SceneNode{
		Spheres:         []*scn.Sphere{&sphere, &gopherLight},
		Discs:           []*scn.Disc{&disc},
		FacetStructures: []*scn.FacetStructure{cornellBox, triangle, spherePodium, discPodium, trianglePodium, gopher, diamond},
	}

	animationStartIndex := 0
	animationEndIndex := amountAnimationFrames - 1

	var animationFrames []scn.Frame

	for frameIndex := animationStartIndex; frameIndex <= animationEndIndex; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountAnimationFrames)

		camera := getCamera(animationProgress, sphereRadius+podiumHeight)

		frame := scn.Frame{
			Filename:   animationName + "_" + fmt.Sprintf("%06d", frameIndex),
			FrameIndex: 0,
			Camera:     &camera,
			SceneNode:  &scene,
		}

		animationFrames = append(animationFrames, frame)
	}

	animation := scn.Animation{
		AnimationName:     animationName,
		Frames:            animationFrames,
		Width:             width,
		Height:            height,
		WriteRawImageFile: false,
	}

	anm.WriteAnimationToFile(animation, false)
}

func spherePrimitive(sphereRadius float64) scn.Sphere {
	material := scn.Material{
		Color:      &color.Color{R: 0.80, G: 1.00, B: 0.80},
		Glossiness: 0.3,
		Roughness:  0.2,
	}
	sphere := scn.Sphere{
		Name:     "Sphere primitive",
		Origin:   &vec3.T{0, 0, 0},
		Radius:   sphereRadius,
		Material: &material,
	}
	return sphere
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

func discPrimitive(discRadius float64) scn.Disc {
	material := scn.Material{
		Color:      &color.Color{R: 1.00, G: 0.80, B: 0.80},
		Glossiness: 0.05,
		Roughness:  0.1,
	}
	disc := scn.Disc{
		Name:     "Disc primitive",
		Normal:   &vec3.T{0, 0, -1},
		Origin:   &vec3.T{0, 0, 0},
		Radius:   discRadius,
		Material: &material,
	}
	return disc
}

func GetCornellBox(scale *vec3.T, lightIntensityFactor float32) *scn.FacetStructure {
	var cornellBoxFilename = "cornellbox.obj"
	var cornellBoxFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/" + cornellBoxFilename

	cornellBoxFile, err := os.Open(cornellBoxFilenamePath)
	if err != nil {
		fmt.Printf("ouupps, something went wrong loading file: '%s'\n%s\n", cornellBoxFilenamePath, err.Error())
	}
	defer cornellBoxFile.Close()

	cornellBox, err := obj.Read(cornellBoxFile)
	cornellBox.Scale(&vec3.Zero, scale)
	cornellBox.ClearMaterials()
	cornellBox.UpdateBounds()
	fmt.Printf("Cornell box bounds: %+v", cornellBox.Bounds)

	cornellBox.Material = &scn.Material{
		Name:  "Cornell box default material",
		Color: &color.Color{R: 0.95, G: 0.95, B: 0.95},
		//Roughness: 0.0,
	}

	lampMaterial := scn.Material{
		Name:          "Lamp",
		Color:         &color.Color{R: 1.0, G: 1.0, B: 1.0},
		Emission:      (&color.Color{R: 1.0, G: 1.0, B: 1.0}).Multiply(lightIntensityFactor),
		RayTerminator: true,
	}
	cornellBox.GetFirstObjectByName("Lamp_1").Material = &lampMaterial
	cornellBox.GetFirstObjectByName("Lamp_2").Material = &lampMaterial
	cornellBox.GetFirstObjectByName("Lamp_3").Material = &lampMaterial
	cornellBox.GetFirstObjectByName("Lamp_4").Material = &lampMaterial

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
	cornellBox.GetFirstObjectByName("Floor_2").Material = &floorMaterial

	return cornellBox
}

func GetGopher(scale *vec3.T) *scn.FacetStructure {
	var objFilename = "go_gopher_color.obj"
	var objFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/" + objFilename

	objFile, err := os.Open(objFilenamePath)
	if err != nil {
		fmt.Printf("ouupps, something went wrong loading file: '%s'\n%s\n", objFilenamePath, err.Error())
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
	var objFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/" + objFilename

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

func getCamera(animationProgress float64, focusHeight float64) scn.Camera {
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
	heading := cameraFocus.Subed(&origin)
	focalDistance := heading.Length()

	return scn.Camera{
		Origin:            &origin,
		Heading:           &heading,
		ViewUp:            &vec3.T{0, 1, 0},
		ViewPlaneDistance: viewPlaneDistance,
		ApertureSize:      cameraAperture,
		FocusDistance:     focalDistance,
		Samples:           amountSamples,
		AntiAlias:         true,
		Magnification:     magnification,
		RenderType:        renderType,
		RecursionDepth:    maxRecursionDepth,
	}
}

// getBox return a box which sides all have the unit length 1.
// It is placed with one corner ni origin (0, 0, 0) and opposite corner in (1, 1, 1).
// Side normals point outwards from each side.
func getBox() *scn.FacetStructure {
	boxP1 := vec3.T{1, 1, 0} // Top right close            3----------2
	boxP2 := vec3.T{1, 1, 1} // Top right away            /          /|
	boxP3 := vec3.T{0, 1, 1} // Top left away            /          / |
	boxP4 := vec3.T{0, 1, 0} // Top left close          4----------1  |
	boxP5 := vec3.T{1, 0, 0} // Bottom right close      | (7)      |  6
	boxP6 := vec3.T{1, 0, 1} // Bottom right away       |          | /
	boxP7 := vec3.T{0, 0, 1} // Bottom left away        |          |/
	boxP8 := vec3.T{0, 0, 0} // Bottom left close       8----------5

	box := scn.FacetStructure{
		FacetStructures: []*scn.FacetStructure{
			{Facets: getRectangleFacets(&boxP1, &boxP2, &boxP3, &boxP4), Name: "ymax"},
			{Facets: getRectangleFacets(&boxP8, &boxP7, &boxP6, &boxP5), Name: "ymin"},
			{Facets: getRectangleFacets(&boxP4, &boxP8, &boxP5, &boxP1), Name: "zmin"},
			{Facets: getRectangleFacets(&boxP6, &boxP7, &boxP3, &boxP2), Name: "zmax"},
			{Facets: getRectangleFacets(&boxP6, &boxP2, &boxP1, &boxP5), Name: "xmax"},
			{Facets: getRectangleFacets(&boxP7, &boxP8, &boxP4, &boxP3), Name: "xmin"},
		},
	}

	return &box
}

// Creates a "four corner facet" using four points (p1,p2,p3,p4).
// The result is two triangles side by side (p1,p2,p4) and (p4,p2,p3).
// Normal direction is calculated as pointing towards observer if the points are listed in counter-clockwise order.
// No test nor calculation is made that the points are exactly in the same plane.
func getRectangleFacets(p1, p2, p3, p4 *vec3.T) []*scn.Facet {
	//       p1
	//       *
	//      / \
	//     /   \
	// p2 *-----* p4
	//     \   /
	//      \ /
	//       *
	//      p3
	//
	// (Normal calculated for each sub-triangle and s aimed towards observer.)

	n1v1 := p2.Subed(p1)
	n1v2 := p4.Subed(p1)
	normal1 := vec3.Cross(&n1v1, &n1v2)
	normal1.Normalize()

	n2v1 := p4.Subed(p3)
	n2v2 := p2.Subed(p3)
	normal2 := vec3.Cross(&n2v1, &n2v2)
	normal2.Normalize()

	return []*scn.Facet{
		{Vertices: []*vec3.T{p1, p2, p4}, Normal: &normal1},
		{Vertices: []*vec3.T{p4, p2, p3}, Normal: &normal2},
	}
}
