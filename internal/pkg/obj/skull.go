package obj

import (
	"fmt"
	"path/filepath"
	"pathtracer/internal/pkg/obj/ply"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

func NewSkull(scale float64) *scn.FacetStructure {
	skull := loadSkull(scale)
	skull.ClearMaterials()
	return skull
}

func loadSkull(scale float64) *scn.FacetStructure {
	skull := ply.ReadFacetStructureOrPanic(filepath.Join(PlyFileDir, "skull.ply"))

	xmin := skull.Bounds.Xmin
	xmax := skull.Bounds.Xmax
	xSize := xmax - xmin

	ymin := skull.Bounds.Ymin
	ymax := skull.Bounds.Ymax
	ySize := ymax - ymin

	zmin := skull.Bounds.Zmin
	zmax := skull.Bounds.Zmax
	zSize := zmax - zmin

	skull.Translate(&vec3.T{-xmin - (xSize / 2), -ymin - (ySize / 2), -zmin - (zSize / 2)}) // Place object in all positive octant
	normalizedSize := 1.0 / max(xSize, ySize, zSize)                                        // resize to size to at most 1.0 units in any axis direction
	skull.ScaleUniform(&vec3.Zero, normalizedSize)
	fmt.Printf("Skull bounds: %+v\n", skull.Bounds)
	skull.ScaleUniform(&vec3.Zero, scale)

	fmt.Printf("Skull bounds: %+v\n", skull.Bounds)

	return skull
}
