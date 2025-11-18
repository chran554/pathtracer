package obj

import (
	"github.com/ungerik/go3d/float64/vec3"
	"path/filepath"
	"pathtracer/internal/pkg/obj/wavefrontobj"
	scn "pathtracer/internal/pkg/scene"
)

func NewCapsule(scale float64) *scn.FacetStructure {
	pill := loadPill()
	pill.ScaleUniform(&vec3.Zero, scale)

	return pill
}

func loadPill() *scn.FacetStructure {
	//capsule := wavefrontobj.ReadOrPanic(filepath.Join(ObjFileDir, "capsule/capsule.obj"))
	pill := wavefrontobj.ReadOrPanic(filepath.Join(ObjFileDir, "capsule/capsule.obj"))

	ymin := pill.Bounds.Ymin
	ymax := pill.Bounds.Ymax
	pill.Translate(&vec3.T{0.0, -ymin, 0.0})       // feet touch the ground (xz-plane)
	pill.ScaleUniform(&vec3.Zero, 1.0/(ymax-ymin)) // resize to height == 1.0

	return pill
}
