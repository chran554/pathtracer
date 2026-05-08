package obj

import (
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/fileformat/wavefront"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

func NewLens(scale float64) *scene.FacetStructure {
	lens := loadLens()

	lens.CenterOn(&vec3.Zero)

	lens.ScaleUniform(&vec3.Zero, scale/lens.Bounds.Ymax)

	lens.ClearMaterials()
	lens.Material = scene.NewMaterial().
		N("glass").
		C(color.NewColor(0.90, 0.92, 0.95)).
		M(0.270, 0.030).
		T(0.700, true, scene.RefractionIndex_Glass)

	return lens
}

func loadLens() *scene.FacetStructure {
	glass := wavefront.ReadFacetStructureOrPanic(filepath.Join(ObjFileDir, "lens.obj"))
	return glass
}
