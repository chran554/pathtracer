package obj

import (
	"fmt"
	"math"
	"path/filepath"
	"pathtracer/internal/pkg/obj/ply"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

func NewLucy(scale float64) *scn.FacetStructure {
	lucy := loadLucy(scale)
	lucy.ClearMaterials()
	return lucy
}

func loadLucy(scale float64) *scn.FacetStructure {
	lucy := ply.ReadFacetStructureOrPanic(filepath.Join(PlyFileDir, "lucy.ply"))

	xmin := lucy.Bounds.Xmin
	xmax := lucy.Bounds.Xmax
	xSize := xmax - xmin
	ymin := lucy.Bounds.Ymin
	ymax := lucy.Bounds.Ymax
	ySize := ymax - ymin
	zmin := lucy.Bounds.Zmin
	zmax := lucy.Bounds.Zmax

	lucy.Translate(&vec3.T{-xmin - xSize/2, -ymin - ySize/2, -zmin}) // Place lucy statue base bottom at origin
	lucy.RotateX(&vec3.Zero, math.Pi/2)                              // Raise the statue to align with y-axis (wing span along x-axis, facing negative z-axis)

	normalizedHeight := 1.0 / (zmax - zmin) // resize to height == 1.0
	lucy.ScaleUniform(&vec3.Zero, scale*normalizedHeight)

	fmt.Printf("Lucy bounds: %+v\n", lucy.Bounds)

	return lucy
}
