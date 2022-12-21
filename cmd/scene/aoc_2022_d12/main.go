package main

import (
	"bufio"
	"fmt"
	"os"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"
	"regexp"
	"strconv"
	"strings"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "aoc_2022_d12"

var environmentEnvironMap = "textures/equirectangular/sunset horizon 2800x1400.jpg"
var environmentRadius = 500.0 * 1000.0
var environmentEmissionFactor = float32(1.5)

var amountFrames = 1

var imageWidth = 800
var imageHeight = 800
var magnification = 0.25

// var renderType = scn.Raycasting
var renderType = scn.Pathtracing
var amountSamples = 80 // 200 * 4 * 3 // 2000 * 2 * 4
var maxRecursion = 3

var viewPlaneDistance = 800.0
var apertureSize = 3.0 // 2.0

func main() {
	animation := getAnimation(int(float64(imageWidth)*magnification), int(float64(imageHeight)*magnification))
	animation.WriteRawImageFile = false

	environmentSphere := getEnvironmentMapping(environmentEnvironMap)

	moon := &scn.Sphere{Name: "moon", Origin: &vec3.T{-400, 500, -400}, Radius: 200}
	moonColor := color.Color{R: 0.9, G: 0.9, B: 1.0}
	moon.Material = scn.NewMaterial().
		C(moonColor, 1.0).
		E(moonColor, 8.0, true)

	sun := &scn.Sphere{Name: "sun", Origin: &vec3.T{3000, 600, 1600}, Radius: 400}
	sunColor := color.Color{R: 1.0, G: 1.0, B: 0.8}
	sun.Material = scn.NewMaterial().
		C(sunColor, 1.0).
		E(sunColor, 12.0, true)

	// Ground

	groundProjection := scn.NewParallelImageProjection("textures/ground/soil-cracked.png", vec3.Zero, vec3.UnitX.Scaled(150*3), vec3.UnitZ.Scaled(150*3))
	ground := scn.Disc{
		Name:     "Ground",
		Origin:   &vec3.Zero,
		Normal:   &vec3.UnitY,
		Radius:   environmentRadius,
		Material: &scn.Material{Name: "Ground material", Color: &color.White, Emission: &color.Black, Glossiness: 0.0, Roughness: 1.0, Projection: &groundProjection},
	}

	mapTextLines, _ := readLines("cmd/scene/aoc_2022_d12/resources/map.txt")
	karta := parseMap(mapTextLines)
	landscape := &scn.FacetStructure{Name: "karta"}
	boxUnit := 20.0
	pathTextLines, _ := readLines("cmd/scene/aoc_2022_d12/resources/path.txt")
	pathPositions, startPos, endPos := parsePath(pathTextLines)

	boxColor := color.Color{R: 0.9, G: 0.9, B: 0.9}
	pathColor := color.Color{R: 0.6, G: 0.5, B: 0.9}
	startColor := color.Color{R: 0.9, G: 0.6, B: 0.5}
	endColor := color.Color{R: 0.6, G: 0.9, B: 0.5}

	pathEmissionFactor := 0.75

	boxMaterial := scn.NewMaterial().
		C(boxColor, 1.0).
		E(color.White, 0.05, false).
		M(0.0, 1.0)

	lavaProjection := scn.NewImageProjection(scn.Spherical, "textures/planets/sun.jpg", vec3.T{1540 / 2, -1540 / 2, 820 / 2}, vec3.UnitX.Scaled(100), vec3.UnitZ.Scaled(100), true, true, true, true)
	lavaMaterial := scn.NewMaterial().
		C(color.White, 1.0).
		E(color.White, 1.5, false).
		M(0.2, 0.3).
		P(&lavaProjection)

	pathMaterial := scn.NewMaterial().
		C(pathColor, 1.0).
		E(pathColor, 2.0*pathEmissionFactor, false).
		M(0.3, 0.1)

	startMaterial := scn.NewMaterial().
		C(startColor, 1.0).
		E(startColor, 3.0*pathEmissionFactor, false).
		M(0.3, 0.1)

	endMaterial := scn.NewMaterial().
		C(endColor, 1.0).
		E(endColor, 3.0*pathEmissionFactor, false).
		M(0.3, 0.1)

	for yPos := 0; yPos < karta.GetHeight(); yPos++ {
		for xPos := 0; xPos < karta.GetWidth(); xPos++ {
			box := getBox(fmt.Sprintf("%d,%d", xPos, yPos))
			height := karta.GetPositionHeight(xPos, yPos)
			if height == 1 && xPos > 0 {
				box.Material = lavaMaterial
			} else {
				box.Material = boxMaterial
			}
			box.Scale(&vec3.Zero, &vec3.T{boxUnit, boxUnit * float64(height), boxUnit})
			box.Translate(&vec3.T{float64(xPos) * boxUnit, 0, float64(yPos) * boxUnit})

			landscape.FacetStructures = append(landscape.FacetStructures, box)
		}
	}

	for _, stepPos := range pathPositions {
		stepMarkerBox := getBox(fmt.Sprintf("%d,%d", stepPos.x, stepPos.y))
		stepMarkerBox.Scale(&vec3.Zero, &vec3.T{boxUnit * 0.5, boxUnit, boxUnit * 0.5})
		stepMarkerBox.Translate(&vec3.T{float64(stepPos.x)*boxUnit + boxUnit*0.25, float64(karta.GetPositionHeight(stepPos.x, stepPos.y)) * boxUnit, float64(stepPos.y)*boxUnit + boxUnit*0.25})

		if startPos == stepPos {
			stepMarkerBox.Material = startMaterial
		} else if endPos == stepPos {
			stepMarkerBox.Material = endMaterial
		} else {
			stepMarkerBox.Material = pathMaterial
		}

		landscape.FacetStructures = append(landscape.FacetStructures, stepMarkerBox)
	}
	landscape.UpdateBounds()

	//fmt.Printf("%+v\n", landscape.Bounds)

	scene := scn.SceneNode{
		Spheres: []*scn.Sphere{&environmentSphere, sun},
		//Spheres:         []*scn.Sphere{&environmentSphere, moon, sun},
		Discs:           []*scn.Disc{&ground},
		ChildNodes:      []*scn.SceneNode{},
		FacetStructures: []*scn.FacetStructure{landscape},
	}

	var curvePoints []*vec3.T
	for _, position := range pathPositions {
		curvePoint := &vec3.T{float64(position.x) * boxUnit, float64(karta.GetPositionHeight(position.x, position.y)) * boxUnit, float64(position.y) * boxUnit}
		curvePoints = append(curvePoints, curvePoint)
	}
	pathCurve := NewSmoothCurve(curvePoints, 0.1, 10, true)

	var curvePoints2 []*vec3.T
	for _, position := range pathPositions {
		curvePoint := &vec3.T{float64(position.x), float64(karta.GetPositionHeight(position.x, position.y)), float64(position.y)}
		curvePoints2 = append(curvePoints2, curvePoint)
	}
	pathCurve00 := NewSmoothCurve(curvePoints2, 0.1, 0, true)
	pathCurve05 := NewSmoothCurve(curvePoints2, 0.1, 5, true)
	pathCurve10 := NewSmoothCurve(curvePoints2, 0.1, 10, true)
	pathCurve20 := NewSmoothCurve(curvePoints2, 0.1, 20, true)

	printPoints(pathCurve00.points, "0.1 00")
	printPoints(pathCurve05.points, "0.1 05")
	printPoints(pathCurve10.points, "0.1 10")
	printPoints(pathCurve20.points, "0.1 20")

	endPosition := pathPositions[len(pathPositions)-1]
	pathEndPoint := &vec3.T{float64(endPosition.x) * boxUnit, float64(karta.GetPositionHeight(endPosition.x, endPosition.y)) * boxUnit, float64(endPosition.y) * boxUnit}

	// Actual animation
	amountFrames = len(pathPositions) * 2
	//headingSamples := 10
	//amountPositionSamples := len(pathPositions)
	for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountFrames)

		focusPoint, _ := pathCurve.GetPoint(animationProgress)
		headingVector, _ := pathCurve.GetSmoothInterpolatedTangent(animationProgress)

		cameraDistanceInHeading := 75.0
		scaledHeading := headingVector.Scaled(cameraDistanceInHeading)
		cameraOrigin := focusPoint.Subed(&scaledHeading)

		mountainCameraOffset := vec3.Sub(&cameraOrigin, pathEndPoint)
		mountainCameraOffset.Normalize()
		mountainCameraOffset.Scale(175.0)
		cameraOrigin.Add(&vec3.T{mountainCameraOffset[0], 0.0, mountainCameraOffset[2]}) // Move camera in xz-plane

		cameraOrigin.Add(&vec3.T{0.0, 100.0, 0.0}) // Raise the camera above the focus point

		cameraFocusPoint := focusPoint
		camera := getCamera(magnification, animationProgress, &cameraOrigin, cameraFocusPoint)

		frame := scn.Frame{
			Filename:   animation.AnimationName + "_" + fmt.Sprintf("%06d", frameIndex),
			FrameIndex: frameIndex,
			Camera:     &camera,
			SceneNode:  &scene,
		}

		animation.Frames = append(animation.Frames, frame)
	}

	anm.WriteAnimationToFile(animation, false)
}

func printPoints(points []*vec3.T, label string) {
	fmt.Println(label)
	for _, point := range points {
		fmt.Printf("(%f,%f)\n", point[0], point[2])
	}
}

func getHeadingVector(progress float64, karta Map, positions []Pos, boxUnit float64) *vec3.T {
	maxIndex := len(positions) - 1
	positionProgress := progress * float64(len(positions))
	positionIndex := int(positionProgress)

	averageHeading := vec3.Zero

	amountContributions := 0
	for i := positionIndex; i < positionIndex+10; i++ {
		pos1 := positions[Min(i, maxIndex)]
		x1 := float64(pos1.x)
		y1 := float64(pos1.y)
		h1 := float64(karta.GetPositionHeight(pos1.x, pos1.y))

		pos2 := positions[Min(i+1, maxIndex)]
		x2 := float64(pos2.x)
		y2 := float64(pos2.y)
		h2 := float64(karta.GetPositionHeight(pos2.x, pos2.y))

		p1 := &vec3.T{x1*boxUnit + boxUnit*0.5, h1*boxUnit + boxUnit*0.5, y1*boxUnit + boxUnit*0.5}
		p2 := &vec3.T{x2*boxUnit + boxUnit*0.5, h2*boxUnit + boxUnit*0.5, y2*boxUnit + boxUnit*0.5}

		headingContribution := vec3.Sub(p2, p1)
		amountContributions++
		averageHeading.Add(&headingContribution)
	}

	averageHeading.Normalize()

	return &averageHeading
}

func getFocusPoint(progress float64, karta Map, positions []Pos, boxUnit float64) *vec3.T {
	maxIndex := len(positions) - 1
	positionProgress := progress * float64(len(positions))
	positionIndex := int(positionProgress) + 10

	pos := positions[Min(positionIndex, maxIndex)]
	x := float64(pos.x)
	y := float64(pos.y)

	height := float64(karta.GetPositionHeight(pos.x, pos.y))
	aim := &vec3.T{x*boxUnit + boxUnit*0.5, height*boxUnit + boxUnit*0.5, y*boxUnit + boxUnit*0.5}

	return aim
}

func Max(i int, i2 int) int {
	if i > i2 {
		return i
	} else {
		return i2
	}
}

func Min(i int, i2 int) int {
	if i < i2 {
		return i
	} else {
		return i2
	}
}

func parsePath(lines []string) ([]Pos, Pos, Pos) {
	var pathPositions []Pos

	var startPos Pos
	var endPos Pos

	coordR, _ := regexp.Compile("(\\d+),(\\d+)")
	for _, line := range lines {
		if coordR.MatchString(line) {
			tokens := coordR.FindStringSubmatch(line)
			x, _ := strconv.Atoi(tokens[1])
			y, _ := strconv.Atoi(tokens[2])
			pos := Pos{x: x, y: y}

			pathPositions = append([]Pos{pos}, pathPositions...)

			if len(pathPositions) == 1 {
				endPos = pos
			}

			startPos = pos
		}
	}

	return pathPositions, startPos, endPos
}

func GetLampPost(lampPostScale *vec3.T) *scn.FacetStructure {
	lampPostMaterial := scn.Material{
		Name:          "lamp post",
		Color:         &color.Color{R: 0.9, G: 0.4, B: 0.3},
		Emission:      &color.Black,
		Glossiness:    0.2,
		Roughness:     0.3,
		RayTerminator: false,
	}

	lampMaterial := scn.Material{
		Name:          "lamp",
		Color:         &color.Color{R: 1.0, G: 1.0, B: 1.0},
		Emission:      (&color.Color{R: 10.0, G: 10.0, B: 9.0}).Multiply(3.0),
		Glossiness:    0.0,
		Roughness:     1.0,
		RayTerminator: true,
	}

	lampPost := LoadLampPost(lampPostScale)
	lampPost.ClearMaterials()
	lampPost.Material = &lampPostMaterial

	lampPost.GetFirstObjectByName("lamp_0").Material = &lampMaterial
	lampPost.GetFirstObjectByName("lamp_1").Material = &lampMaterial
	lampPost.GetFirstObjectByName("lamp_2").Material = &lampMaterial
	lampPost.GetFirstObjectByName("lamp_3").Material = &lampMaterial
	return lampPost
}

func LoadGopher(scale *vec3.T) *scn.FacetStructure {
	var objFilename = "go_gopher_color.obj"
	var objFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/obj/" + objFilename

	objFile, err := os.Open(objFilenamePath)
	if err != nil {
		fmt.Printf("ouupps, something went wrong loading file: '%s'\n%s\n", objFilenamePath, err.Error())
	}
	defer objFile.Close()

	obj, err := obj.Read(objFile)
	ymin := obj.Bounds.Ymin
	ymax := obj.Bounds.Ymax
	obj.Translate(&vec3.T{0.0, -ymin, 0.0})       // feet touch the ground (xz-plane)
	obj.ScaleUniform(&vec3.Zero, 1.0/(ymax-ymin)) // resize to height == 1.0

	obj.Scale(&vec3.Zero, scale)

	obj.UpdateBounds()
	fmt.Printf("Gopher bounds: %+v\n", obj.Bounds)

	return obj
}

func LoadLampPost(scale *vec3.T) *scn.FacetStructure {
	var objFilename = "lamp_post.obj"
	var objFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/obj/" + objFilename

	objFile, err := os.Open(objFilenamePath)
	if err != nil {
		fmt.Printf("ouupps, something went wrong loading file: '%s'\n%s\n", objFilenamePath, err.Error())
	}
	defer objFile.Close()

	obj, err := obj.Read(objFile)
	ymin := obj.Bounds.Ymin
	ymax := obj.Bounds.Ymax
	obj.Translate(&vec3.T{0.0, -ymin, 0.0})       // lamp post base touch the ground (xz-plane)
	obj.ScaleUniform(&vec3.Zero, 1.0/(ymax-ymin)) // resize to height == 1.0

	obj.Scale(&vec3.Zero, scale)

	obj.UpdateBounds()
	fmt.Printf("Lamp post bounds: %+v\n", obj.Bounds)

	return obj
}

func LoadKerosineLamp(scale *vec3.T) *scn.FacetStructure {
	var objFilename = "kerosine_lamp.obj"
	var objFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/obj/" + objFilename

	objFile, err := os.Open(objFilenamePath)
	if err != nil {
		fmt.Printf("ouupps, something went wrong loading file: '%s'\n%s\n", objFilenamePath, err.Error())
	}
	defer objFile.Close()

	obj, err := obj.Read(objFile)
	ymin := obj.Bounds.Ymin
	ymax := obj.Bounds.Ymax
	obj.Translate(&vec3.T{0.0, -ymin, 0.0})       // lamp post base touch the ground (xz-plane)
	obj.ScaleUniform(&vec3.Zero, 1.0/(ymax-ymin)) // resize to height == 1.0

	obj.Scale(&vec3.Zero, scale)

	obj.UpdateBounds()
	fmt.Printf("Kerosine lamp bounds: %+v\n", obj.Bounds)

	return obj
}

func getEnvironmentMapping(filename string) scn.Sphere {
	origin := vec3.T{0, 0, 0}

	projection := scn.ImageProjection{
		ProjectionType: scn.Spherical,
		ImageFilename:  filename,
		Gamma:          1.5,
		Origin:         &origin,
		U:              &vec3.T{-0.55, 0, -0.45},
		V:              &vec3.T{0, 1, 0},
		RepeatU:        true,
		RepeatV:        true,
		FlipU:          false,
		FlipV:          false,
	}

	material := scn.Material{
		Color:         &color.Color{R: 1.0, G: 1.0, B: 1.0},
		Emission:      (&color.Color{R: 1.0, G: 1.0, B: 1.0}).Multiply(environmentEmissionFactor),
		RayTerminator: true,
		Projection:    &projection,
	}

	sphere := scn.Sphere{
		Name:     "Environment mapping",
		Origin:   &origin,
		Radius:   environmentRadius,
		Material: &material,
	}

	return sphere
}

func getAnimation(width int, height int) scn.Animation {
	animation := scn.Animation{
		AnimationName:     animationName,
		Frames:            []scn.Frame{},
		Width:             width,
		Height:            height,
		WriteRawImageFile: false,
	}
	return animation
}

func getCamera(magnification float64, progress float64, cameraOrigin *vec3.T, cameraFocusPoint *vec3.T) scn.Camera {

	// Point heading towards center of sphere ring (heading vector starts in camera origin)
	heading := cameraFocusPoint.Subed(cameraOrigin)

	focusDistance := heading.Length()

	return scn.Camera{
		Origin:            cameraOrigin,
		Heading:           &heading,
		ViewUp:            &vec3.T{0, 1, 0},
		ViewPlaneDistance: viewPlaneDistance,
		ApertureSize:      apertureSize,
		FocusDistance:     focusDistance,
		Samples:           amountSamples,
		AntiAlias:         true,
		Magnification:     magnification,
		RenderType:        renderType,
		RecursionDepth:    maxRecursion,
	}
}

func parseMap(lines []string) Map {
	var start Pos
	var end Pos
	heightMap := make([][]int, 0)
	for y, line := range lines {
		heightMap = append(heightMap, make([]int, 0))
		for x, heightText := range line {
			if heightText == 'S' {
				heightText = 'a'
				start = Pos{x: x, y: y}
			} else if heightText == 'E' {
				heightText = 'z'
				end = Pos{x: x, y: y}
			}
			height := heightText - 'a' + 1
			heightMap[y] = append(heightMap[y], int(height))
		}
	}

	return Map{heights: heightMap, start: start, end: end}
}

type Pos struct {
	x, y int
}

type Map struct {
	heights [][]int
	start   Pos
	end     Pos
}

func (m Map) GetHeight() int {
	return len(m.heights)
}

func (m Map) GetWidth() int {
	return len(m.heights[0])
}

func (m Map) GetPositionHeight(x int, y int) int {
	if x < 0 || x >= m.GetWidth() || y < 0 || y >= m.GetHeight() {
		return 1000000
	}
	return m.heights[y][x]
}

func readTextFileLines(inputFilepath string, purgeEmpty bool) []string {
	lines, err := readLines(inputFilepath)
	if err != nil {
		fmt.Printf("could not read file: %s\n", err.Error())
		os.Exit(1)
	}

	for lineIndex := 0; lineIndex < len(lines); {
		line := lines[lineIndex]

		if strings.TrimSpace(line) == "" {
			lines = append(lines[:lineIndex], lines[lineIndex+1:]...)
			continue
		}

		lineIndex++
	}

	return lines
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

func getBox(name string) *scn.FacetStructure {
	//roofTexture := scn.NewParallelImageProjection("textures/uv.png", vec3.T{0, ballRadius * 6, 0}, vec3.T{ballRadius, 0, 0}, vec3.T{0, 0, ballRadius})
	//floorTexture := scn.NewParallelImageProjection("textures/uv.png", vec3.T{0, 0, 0}, vec3.T{ballRadius, 0, 0}, vec3.T{0, 0, ballRadius})

	boxP1 := vec3.T{1, 1, 0} // Top right close            3----------2
	boxP2 := vec3.T{1, 1, 1} // Top right away            /          /|
	boxP3 := vec3.T{0, 1, 1} // Top left away            /          / |
	boxP4 := vec3.T{0, 1, 0} // Top left close          4----------1  |
	boxP5 := vec3.T{1, 0, 0} // Bottom right close      | (7)      |  6
	boxP6 := vec3.T{1, 0, 1} // Bottom right away       |          | /
	boxP7 := vec3.T{0, 0, 1} // Bottom left away        |          |/
	boxP8 := vec3.T{0, 0, 0} // Bottom left close       8----------5

	box := scn.FacetStructure{Name: name}

	roof := getFlatRectangleFacets(&boxP1, &boxP2, &boxP3, &boxP4)
	floor := getFlatRectangleFacets(&boxP8, &boxP7, &boxP6, &boxP5)
	back := getFlatRectangleFacets(&boxP6, &boxP7, &boxP3, &boxP2)
	front := getFlatRectangleFacets(&boxP5, &boxP1, &boxP4, &boxP8)
	right := getFlatRectangleFacets(&boxP6, &boxP2, &boxP1, &boxP5)
	left := getFlatRectangleFacets(&boxP7, &boxP8, &boxP4, &boxP3)

	box.Facets = append(box.Facets, front...)
	box.Facets = append(box.Facets, back...)
	box.Facets = append(box.Facets, roof...)
	box.Facets = append(box.Facets, floor...)
	box.Facets = append(box.Facets, right...)
	box.Facets = append(box.Facets, left...)

	return &box
}

func getFlatRectangleFacets(p1, p2, p3, p4 *vec3.T) []*scn.Facet {
	v1 := p2.Subed(p1)
	v2 := p3.Subed(p1)
	normal := vec3.Cross(&v1, &v2)
	normal.Normalize()

	return []*scn.Facet{
		{Vertices: []*vec3.T{p1, p2, p4}, Normal: &normal},
		{Vertices: []*vec3.T{p4, p2, p3}, Normal: &normal},
	}
}
