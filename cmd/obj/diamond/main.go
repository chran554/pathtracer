package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"pathtracer/cmd/obj/diamond/pkg/diamond"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"
	"strings"
)

func main() {
	scale := 100.0

	d := diamond.Diamond{
		GirdleDiameter:                         1.00, // 100.0%
		GirdleHeightRelativeGirdleDiameter:     0.03, //   3.0%
		CrownAngleDegrees:                      34.0, //  34.0°
		TableFacetSizeRelativeGirdle:           0.56, //  56.0%
		StarFacetSizeRelativeCrownSide:         0.55, //  55.0%
		PavilionAngleDegrees:                   41.1, //  41.1°
		LowerHalfFacetSizeRelativeGirdleRadius: 0.77, //  77.0%
	}

	m := scn.Material{
		Name:            "Diamond",
		Color:           color.Color{R: 1.00, G: 0.98, B: 0.95}, // Slightly yellowish color
		Glossiness:      0.98,
		Roughness:       0.0,
		RefractionIndex: 2.42, // Refraction index of diamond material
		Transparency:    1.0,
	}

	diamond := diamond.NewDiamondRoundBrilliantCut(d, scale, m)

	comments := fileComments(scale, d, m, diamond.Bounds)

	objFile := createFile("diamond.obj")
	defer objFile.Close()
	mtlFile := createFile("diamond.mtl")
	defer mtlFile.Close()

	obj.WriteObjFile(objFile, mtlFile, diamond, comments)

	fmt.Printf("\nCreated diamond obj-file: %s\n", objFile.Name())
}

func fileComments(scale float64, d diamond.Diamond, m scn.Material, b *scn.Bounds) []string {
	comments := []string{
		"Brilliant cut diamond 3D OBJ-file was created using algorithm from https://github.com/chran554/pathtracer/",
		"The following parameters were used creating these files:",
		"",
		fmt.Sprintf("Global scale of diamond: %.2f", scale),
		"",
	}
	comments = append(comments, "Diamond:")
	comments = append(comments, prettyPrintedStruct(d)...)
	comments = append(comments, "")
	comments = append(comments, "Diamond material")
	comments = append(comments, prettyPrintedStruct(m)...)
	comments = append(comments, "")
	comments = append(comments, "Diamond bounds")
	comments = append(comments, prettyPrintedStruct(b)...)
	return comments
}

func prettyPrintedStruct(anyStruct any) []string {
	empJSON, err := json.MarshalIndent(anyStruct, "", "  ")
	if err != nil {
		log.Fatalf(err.Error())
	}

	return strings.Split(string(empJSON), "\n")
}

func createFile(name string) *os.File {
	objFile, err := os.Create(name)
	if err != nil {
		fmt.Printf("could not create file: '%s'\n%s\n", objFile.Name(), err.Error())
		os.Exit(1)
	}
	return objFile
}
