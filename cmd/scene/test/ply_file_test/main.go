package main

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"math"
	"os"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/ply"
	scn "pathtracer/internal/pkg/scene"
)

var animationName = "ply_file_test"

var renderType = scn.Pathtracing

//var renderType = scn.Raycasting

var maxRecursionDepth = 3
var amountSamples = 1024 // 2500
var lensRadius = 5.0

var viewPlaneDistance = 800.0
var cameraDistanceFactor = 1.0

var imageWidth = 450
var imageHeight = 450
var magnification = 2.0 // 2.0

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

	cornellBox := GetCornellBox(&vec3.T{500, 300, 500}, 9.0) // cm, as units. I.e. a 5x3x5m room

	boxHeight := 40.0
	boxWidth := 120.0
	box := getBox()
	boxProjection := scn.NewParallelImageProjection("textures/concrete/Polished-Concrete-Architextures.jpg", vec3.Zero, (&vec3.UnitX).Scaled(boxWidth), (&vec3.UnitZ).Add(&vec3.T{0, 0.5, 0}).Scaled(boxWidth))
	box.Material = &scn.Material{
		Color:      (&color.Color{R: 1, G: 1, B: 1}).Multiply(0.9),
		Glossiness: 0.4,
		Roughness:  0.1,
		Projection: &boxProjection,
	}
	box.Translate(&vec3.T{-0.5, 0, -0.5})
	box.Scale(&vec3.Zero, &vec3.T{boxWidth * 2, boxHeight, boxWidth * 2})

	var plyFilename = "beethoven.ply"
	var plyFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/ply/" + plyFilename

	plyFile, err := os.Open(plyFilenamePath)
	if err != nil {
		fmt.Printf("ouupps, something went wrong loading file: '%s'\n%s\n", plyFilenamePath, err.Error())
		return
	}
	defer plyFile.Close()

	plyFacetStructure, err := ply.ReadPlyFile(plyFile)
	if err != nil {
		fmt.Printf("could not read ply-file '%s': %s", plyFile.Name(), err.Error())
		return
	}
	plyFacetStructure.Translate(&vec3.T{0, -plyFacetStructure.Bounds.Ymin, 0})
	plyFacetStructure.ScaleUniform(&vec3.Zero, 25.0)
	plyFacetStructure.RotateY(&vec3.Zero, math.Pi-math.Pi/12.0)
	plyFacetStructure.Translate(&vec3.T{0, boxHeight, 0})
	plyFacetStructure.UpdateBounds()
	plyFacetStructure.UpdateNormals()
	//plyFacetStructure.UpdateVertexNormals()

	fmt.Printf("ply object bounds: %+v\n", plyFacetStructure.Bounds)

	plyProjection := scn.NewParallelImageProjection("textures/marble/white_marble.png", vec3.Zero, vec3.UnitX.Scaled(200), vec3.UnitY.Scaled(200))
	plyMaterial := &scn.Material{
		Color:      &color.White,
		Glossiness: 0.1,
		Roughness:  0.6,
		Projection: &plyProjection,
	}
	plyFacetStructure.Material = plyMaterial

	scene := scn.SceneNode{
		FacetStructures: []*scn.FacetStructure{cornellBox, box, plyFacetStructure},
	}

	camera := getCamera()

	animation := scn.Animation{
		AnimationName:     animationName,
		Frames:            []scn.Frame{},
		Width:             width,
		Height:            height,
		WriteRawImageFile: false,
	}

	frame := scn.Frame{
		Filename:   animation.AnimationName,
		FrameIndex: 0,
		Camera:     &camera,
		SceneNode:  &scene,
	}

	animation.Frames = append(animation.Frames, frame)

	anm.WriteAnimationToFile(animation, false)
}

func GetCornellBox(scale *vec3.T, lightIntensityFactor float32) *scn.FacetStructure {
	var cornellBoxFilename = "cornellbox.obj"
	var cornellBoxFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/obj/" + cornellBoxFilename

	cornellBoxFile, err := os.Open(cornellBoxFilenamePath)
	if err != nil {
		fmt.Printf("ouupps, something went wrong loading file: '%s'\n%s\n", cornellBoxFilenamePath, err.Error())
	}
	defer cornellBoxFile.Close()

	cornellBox, err := obj.Read(cornellBoxFile)
	cornellBox.Scale(&vec3.Zero, scale)
	cornellBox.ClearMaterials()

	cornellBox.Material = &scn.Material{
		Name:       "Cornell box default material",
		Color:      &color.Color{R: 0.95, G: 0.95, B: 0.95},
		Glossiness: 0.0,
		Roughness:  1.0,
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

	backWallProjection := scn.NewParallelImageProjection("textures/wallpaper/geometric-yellow-wallpaper.jpg", vec3.Zero, vec3.UnitX.Scaled(scale[0]), vec3.UnitY.Scaled(scale[0]*0.66))
	backWallMaterial := *cornellBox.Material
	backWallMaterial.Projection = &backWallProjection
	cornellBox.GetFirstObjectByName("Wall_back").Material = &backWallMaterial

	sideWallProjection := scn.NewParallelImageProjection("textures/wallpaper/geometric-yellow-wallpaper.jpg", vec3.Zero, vec3.UnitZ.Scaled(scale[0]), vec3.UnitY.Scaled(scale[0]*0.66))
	sideWallMaterial := *cornellBox.Material
	sideWallMaterial.Projection = &sideWallProjection
	cornellBox.GetFirstObjectByName("Wall_left").Material = &sideWallMaterial
	cornellBox.GetFirstObjectByName("Wall_right").Material = &sideWallMaterial

	floorProjection := scn.NewParallelImageProjection("textures/floor/Calacatta-Vena-French-Pattern-Architextures.jpg", vec3.Zero, vec3.UnitX.Scaled(scale[0]/2), vec3.UnitZ.Scaled(scale[0]/2))
	floorMaterial := *cornellBox.Material
	floorMaterial.Glossiness = 0.6
	floorMaterial.Roughness = 0.1
	floorMaterial.Projection = &floorProjection
	cornellBox.GetFirstObjectByName("Floor_2").Material = &floorMaterial

	return cornellBox
}

func getCamera() scn.Camera {
	origin := vec3.T{0, 60 * 3, -800}
	origin.Scale(cameraDistanceFactor)

	focusPoint := vec3.T{0, 60 * 3, -30 * 3}

	heading := focusPoint.Subed(&origin)
	focusDistance := heading.Length()

	return scn.Camera{
		Origin:            &origin,
		Heading:           &heading,
		ViewUp:            &vec3.T{0, 1, 0},
		ViewPlaneDistance: viewPlaneDistance,
		ApertureSize:      lensRadius,
		FocusDistance:     focusDistance,
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
