package main

import (
	"fmt"
	"os"
	"pathtracer/cmd/obj/diamond/pkg/diamond"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"
)

func main() {
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

	diamond := diamond.NewDiamondRoundBrilliantCut(d, 100.0, m)

	objFile := createFile("diamond.obj")
	defer objFile.Close()
	mtlFile := createFile("diamond.mtl")
	defer mtlFile.Close()

	obj.WriteObjFile(objFile, mtlFile, diamond)

	fmt.Printf("Created diamond obj-file: %s\n", objFile.Name())
}

func createFile(name string) *os.File {
	objFile, err := os.Create(name)
	if err != nil {
		fmt.Printf("could not create file: '%s'\n%s\n", objFile.Name(), err.Error())
		os.Exit(1)
	}
	return objFile
}
