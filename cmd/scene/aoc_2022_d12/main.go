package main

import (
	"bufio"
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"os"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"
	"regexp"
	"strconv"
)

var animationName = "aoc_2022_d12"

var environmentEnvironMap = "textures/equirectangular/sunset horizon 2800x1400.jpg"
var environmentRadius = 500.0 * 1000.0
var environmentEmissionFactor = 0.8

var amountFrames = 1

var imageWidth = 1000
var imageHeight = 1000
var magnification = 0.25

// var renderType = scn.Raycasting
var amountSamples = 200 // 200 * 4 * 3 // 2000 * 2 * 4

var apertureSize = 2.0

func main() {
	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, false)

	skyDomeOrigin := &vec3.T{0, 0, 0}
	skyDomeMaterial := scn.NewMaterial().
		E(color.White, environmentEmissionFactor, true).
		SP(environmentEnvironMap, skyDomeOrigin, vec3.T{-0.55, 0, -0.45}, vec3.T{0, 1, 0})
	skyDome := scn.NewSphere(skyDomeOrigin, environmentRadius, skyDomeMaterial).N("sky dome")

	//moonColor := color.Color{R: 0.9, G: 0.9, B: 1.0}
	//moonMaterial := scn.NewMaterial().C(moonColor).E(moonColor, 8.0, true)
	//moon := scn.NewSphere(&vec3.T{-400, 500, -400}, 200, moonMaterial).N("moon")

	//sunColor := color.Color{R: 1.0, G: 0.97, B: 0.8}
	//sunMaterial := scn.NewMaterial().C(sunColor).E(sunColor, 10.0, true)
	//sun := scn.NewSphere(&vec3.T{3000, 600, 1600}, 400, sunMaterial).N("sun")

	// Ground
	groundMaterial := scn.NewMaterial().N("Ground material").PP("textures/ground/soil-cracked.png", &vec3.Zero, vec3.UnitX.Scaled(150*3), vec3.UnitZ.Scaled(150*3))
	ground := scn.NewDisc(&vec3.T{0, 0, 0}, &vec3.UnitY, environmentRadius, groundMaterial).N("Ground")

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
		C(boxColor).
		E(color.White, 0.02, false)

	lavaMaterial := scn.NewMaterial().
		C(color.White).
		E(color.White, 1.5, false).
		M(0.2, 0.3).
		SP("textures/planets/sun.jpg", &vec3.T{1540 / 2, -1540 / 2, 820 / 2}, vec3.UnitX.Scaled(-100), vec3.UnitZ.Scaled(100))

	pathMaterial := scn.NewMaterial().
		C(pathColor).
		E(pathColor, 2.0*pathEmissionFactor, false).
		M(0.3, 0.1)

	startMaterial := scn.NewMaterial().
		C(startColor).
		E(startColor, 1.5*pathEmissionFactor, false).
		M(0.3, 0.1)

	endMaterial := scn.NewMaterial().
		C(endColor).
		E(endColor, 1.5*pathEmissionFactor, false).
		M(0.3, 0.1)

	for yPos := 0; yPos < karta.GetHeight(); yPos++ {
		for xPos := 0; xPos < karta.GetWidth(); xPos++ {
			box := obj.NewBox(obj.BoxPositive)
			box.Name = fmt.Sprintf("ground (%d,%d)", xPos, yPos)
			height := karta.GetPositionHeight(xPos, yPos)
			box.Scale(&vec3.Zero, &vec3.T{boxUnit, boxUnit * float64(height), boxUnit})
			box.Translate(&vec3.T{float64(xPos) * boxUnit, 0, float64(yPos) * boxUnit})

			if height == 1 && xPos > 0 {
				box.Material = lavaMaterial
			} else {
				box.Material = boxMaterial
			}

			landscape.FacetStructures = append(landscape.FacetStructures, box)
		}
	}

	for stepIndex, stepPos := range pathPositions {
		stepMarkerBox := obj.NewBox(obj.BoxPositive)
		stepMarkerBox.Name = fmt.Sprintf("step %d: (%d,%d)", stepIndex, stepPos.x, stepPos.y)
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

	scene := scn.NewSceneNode().S(skyDome /*, sun*/ /*, moon*/).D(ground).FS(landscape)

	var curvePoints []*vec3.T
	for _, position := range pathPositions {
		curvePoint := &vec3.T{float64(position.x) * boxUnit, float64(karta.GetPositionHeight(position.x, position.y)) * boxUnit, float64(position.y) * boxUnit}
		curvePoints = append(curvePoints, curvePoint)
	}
	pathCurve := NewSmoothCurve(curvePoints, 0.1, 10, 0, true)
	pathCurve.Subdivide(2)
	pathCurve.Smooth(10)

	endPosition := pathPositions[len(pathPositions)-1]
	pathEndPoint := &vec3.T{float64(endPosition.x) * boxUnit, float64(karta.GetPositionHeight(endPosition.x, endPosition.y)) * boxUnit, float64(endPosition.y) * boxUnit}

	// Actual animation
	amountFrames = len(pathPositions) * 4
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
		camera := scn.NewCamera(&cameraOrigin, cameraFocusPoint, amountSamples, magnification).A(apertureSize, "")

		// Still image camera settings
		// cameraOrigin = vec3.T{-300, 350, 150}
		// focusPoint = &vec3.T{850, 10, 400}
		// camera = scn.NewCamera(&cameraOrigin, focusPoint, amountSamples, magnification).A(apertureSize, "")

		frame := scn.NewFrame(animationName, frameIndex, camera, scene)
		animation.AddFrame(frame)
	}

	anm.WriteAnimationToFile(animation, false)
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
