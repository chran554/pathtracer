package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"pathtracer/cmd/obj/diamond/pkg/diamond"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj/wavefrontobj"
	scn "pathtracer/internal/pkg/scene"
	"strings"
)

func main() {
	createPerfectBrilliantCutDiamondObjFile(100.0, "diamond")
}

func createPerfectBrilliantCutDiamondObjFile(scale float64, filename string) {
	d := diamond.Diamond{
		GirdleDiameter:                         1.00, // 100.0%
		GirdleHeightRelativeGirdleDiameter:     0.03, //   3.0%
		CrownAngleDegrees:                      34.0, //  34.0°
		TableFacetSizeRelativeGirdle:           0.56, //  56.0%
		StarFacetSizeRelativeCrownSide:         0.55, //  55.0%
		PavilionAngleDegrees:                   41.1, //  41.1°
		LowerHalfFacetSizeRelativeGirdleRadius: 0.77, //  77.0%
	}

	m := diamondMaterial()
	diamond := diamond.NewDiamondRoundBrilliantCut(d, scale, m)

	comments := fileComments(scale, d, m, diamond.Bounds)
	objFile := writeDiamondObjFile(filename, diamond, comments)
	fmt.Printf("\nCreated perfect brilliant cut diamond obj-file: %s\n", objFile.Name())
}

func writeDiamondObjFile(filename string, diamond *scn.FacetStructure, comments []string) *os.File {
	objFile := createFile(filename + ".obj")
	defer objFile.Close()
	mtlFile := createFile(filename + ".mtl")
	defer mtlFile.Close()

	wavefrontobj.WriteObjFile(objFile, mtlFile, diamond, comments)
	return objFile
}

func diamondMaterial() scn.Material {
	m := scn.Material{
		Name:            "Diamond",
		Color:           &color.Color{R: 1.00, G: 0.99, B: 0.97}, // Very slight yellowish color
		Glossiness:      0.01,
		Roughness:       0.01,
		RefractionIndex: 2.42, // Refraction index of diamond material
		Transparency:    0.99,
	}
	return m
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
