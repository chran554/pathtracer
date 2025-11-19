package obj

import (
	"path/filepath"
	"pathtracer/internal/pkg/obj/wavefrontobj"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

func NewGopher(scale float64) *scene.FacetStructure {
	gopher := loadGopher()
	gopher.ScaleUniform(&vec3.Zero, scale)

	return gopher
}

func loadGopher() *scene.FacetStructure {
	gopher := wavefrontobj.ReadOrPanic(filepath.Join(ObjFileDir, "go_gopher_color.obj"))

	ymin := gopher.Bounds.Ymin
	ymax := gopher.Bounds.Ymax
	gopher.Translate(&vec3.T{0.0, -ymin, 0.0})       // feet touch the ground (xz-plane)
	gopher.ScaleUniform(&vec3.Zero, 1.0/(ymax-ymin)) // resize to height == 1.0

	return gopher
}
