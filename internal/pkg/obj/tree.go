package obj

import (
	"path/filepath"
	"pathtracer/internal/pkg/fileformat/wavefront"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

func NewTree01(scale float64) *scene.FacetStructure {
	tree := loadTree01()
	tree.ScaleUniform(&vec3.Zero, scale)

	return tree
}

func loadTree01() *scene.FacetStructure {
	tree := wavefront.ReadFacetStructureOrPanic(filepath.Join(ObjFileDir, "tree/Tree1 - middle.obj"))

	ymin := tree.Bounds.Ymin
	ymax := tree.Bounds.Ymax
	tree.Translate(&vec3.T{0.0, -ymin, 0.0})       // feet touch the ground (xz-plane)
	tree.ScaleUniform(&vec3.Zero, 1.0/(ymax-ymin)) // resize to height == 1.0

	return tree
}

/*
func NewTree02(scale float64) *scene.FacetStructure {

	tree := &scene.FacetStructure{
		Name:            "Tree",
		FacetStructures: []*scene.FacetStructure{stem, branches},
		IgnoreBounds:    false,
	}

	tree.ScaleUniform(&vec3.Zero, scale)
	return tree
}
*/
