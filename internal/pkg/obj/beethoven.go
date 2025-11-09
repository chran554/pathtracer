package obj

import (
	"fmt"
	"math"
	"path/filepath"
	"pathtracer/internal/pkg/obj/ply"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

// NewBeethoven creates a new beethoven statue with the center of the statue bottom in origin (0,0,0).
func NewBeethoven(scale float64) *scn.FacetStructure {
	statue := loadGopher()
	statue.ScaleUniform(&vec3.Zero, scale)

	return statue
}

func loadBeethoven() *scn.FacetStructure {
	statue := ply.ReadFacetStructureOrPanic(filepath.Join(PlyFileDir, "beethoven.ply"))

	statue.CenterOn(&vec3.Zero)
	statue.RotateY(&vec3.Zero, math.Pi-math.Pi/12.0)
	statue.Translate(&vec3.T{0, -statue.Bounds.Ymin, 0})
	statue.ScaleUniform(&vec3.Zero, 1.0/statue.Bounds.SizeY())

	statue.UpdateNormals()

	fmt.Printf("Beethoven bounds: %+v\n", statue.Bounds)

	return statue
}
