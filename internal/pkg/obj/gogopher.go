package obj

import (
	"fmt"
	"path/filepath"
	"pathtracer/internal/pkg/fileformat/wavefront"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

func NewGopher(scale float64) *scene.FacetStructure {
	gopher := loadGopher()
	gopher.ScaleUniform(&vec3.Zero, scale)

	return gopher
}

func loadGopher() *scene.FacetStructure {
	gopher := wavefront.ReadFacetStructureOrPanic(filepath.Join(ObjFileDir, "go_gopher_color.obj"))
	gopher.FlipX(&vec3.Zero)

	ymin := gopher.Bounds.Ymin
	ymax := gopher.Bounds.Ymax
	gopher.Translate(&vec3.T{0.0, -ymin, 0.0})       // feet touch the ground (xz-plane)
	gopher.ScaleUniform(&vec3.Zero, 1.0/(ymax-ymin)) // resize to height == 1.0

	return gopher
}

func GopherFacing(gopher *scene.FacetStructure, facingPoint *vec3.T) {
	fmt.Print(gopher)
}

func GopherFocusEye(gopher *scene.FacetStructure, eyeFocusPoint *vec3.T) {
	fmt.Print(gopher)
}
