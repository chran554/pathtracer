package obj

import (
	"github.com/ungerik/go3d/float64/vec3"
	"path/filepath"
	"pathtracer/internal/pkg/obj/wavefrontobj2"
	scn "pathtracer/internal/pkg/scene"
)

func NewGopher(scale float64) *scn.FacetStructure {
	gopher := loadGopher()
	gopher.ScaleUniform(&vec3.Zero, scale)

	return gopher
}

func loadGopher() *scn.FacetStructure {
	gopher := wavefrontobj2.ReadOrPanic(filepath.Join(ObjFileDir, "go_gopher_color.obj"))

	ymin := gopher.Bounds.Ymin
	ymax := gopher.Bounds.Ymax
	gopher.Translate(&vec3.T{0.0, -ymin, 0.0})       // feet touch the ground (xz-plane)
	gopher.ScaleUniform(&vec3.Zero, 1.0/(ymax-ymin)) // resize to height == 1.0

	return gopher
}
