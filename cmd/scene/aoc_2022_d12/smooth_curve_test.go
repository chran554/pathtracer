package main

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"testing"
)

func Test_Subdivide(t *testing.T) {
	smoothCurve := SmoothCurve{
		points:         []*vec3.T{{0, 0, 0}, {100, 100, 0}, {200, 0, 0}},
		smoothStrength: 0.1,
		smoothLevel:    0,
		lockEndPoints:  true,
	}

	fmt.Printf("%+v\n", smoothCurve.points)

	smoothCurve.Subdivide(4)

	fmt.Printf("%+v\n", smoothCurve.points)
}

func Test_Smooth(t *testing.T) {
	//pathTextLines, _ := readLines("cmd/scene/aoc_2022_d12/resources/path.txt")
	pathTextLines, _ := readLines("resources/path.txt")
	pathPositions, _, _ := parsePath(pathTextLines)

	var curvePoints2 []*vec3.T
	for _, position := range pathPositions {
		curvePoint := &vec3.T{float64(position.x), float64(position.y), 0}
		curvePoints2 = append(curvePoints2, curvePoint)
	}

	// pathCurve00 := NewSmoothCurve(curvePoints2, 0.1, 0, true)
	// pathCurve05 := NewSmoothCurve(curvePoints2, 0.1, 5, true)
	// pathCurve10 := NewSmoothCurve(curvePoints2, 0.1, 10, true)
	pathCurve20 := NewSmoothCurve(curvePoints2, 0.25, 30, 2, true)

	// printPoints(pathCurve00.points, "0.1 00")
	// printPoints(pathCurve05.points, "0.1 05")
	// printPoints(pathCurve10.points, "0.1 10")
	printPoints(pathCurve20.points, "0.1 20")
}

func printPoints(points []*vec3.T, label string) {
	fmt.Println(label)
	for _, point := range points {
		fmt.Printf("(%f,%f)\n", point[0], point[1])
	}
}
