package obj

import (
	"github.com/ungerik/go3d/float64/vec3"
	"path/filepath"
	"pathtracer/internal/pkg/obj/wavefrontobj"
	scn "pathtracer/internal/pkg/scene"
)

func NewSolidUtahTeapot(scale float64, includeBody bool, includeLid bool) *scn.FacetStructure {
	utahTeaPot := wavefrontobj.ReadOrPanic(filepath.Join(ObjFileDir, "utah_teapot_solid.obj"))

	if !includeBody {
		utahTeaPot.RemoveObjectsByName("teapot")
	}
	if !includeLid {
		utahTeaPot.RemoveObjectsByName("lid")
	}

	utahTeaPot.CenterOn(&vec3.Zero)
	utahTeaPot.Translate(&vec3.T{0, -utahTeaPot.Bounds.Ymin, 0})

	utahTeaPot.ScaleUniform(&vec3.Zero, scale/utahTeaPot.Bounds.Ymax)

	// porcelainMaterial := scn.NewMaterial().
	//	N("Porcelain material").
	//	C(color.NewColorGrey(0.85)).
	//	M(0.1, 0.1).
	//	T(0.0, true, scn.RefractionIndex_Porcelain)

	// glassMaterial := scn.NewMaterial().
	// 	N("Glass material").
	// 	C(color.NewColor(0.95, 0.95, 0.97), 1.0).
	// 	M(0.2, 0.05).
	// 	T(1.0, true, scn.RefractionIndex_Glass)

	// utahTeaPot.ClearMaterials()
	// utahTeaPot.Material = porcelainMaterial

	return utahTeaPot
}

func NewTeacup(scale float64, includeCup bool, includeSaucer bool, includeSpoon bool) *scn.FacetStructure {
	teacup := wavefrontobj.ReadOrPanic(filepath.Join(ObjFileDir, "teacup.obj"))

	if !includeCup {
		teacup.RemoveObjectsByName("teacup")
	}
	if !includeSaucer {
		teacup.RemoveObjectsByName("saucer")
	}
	if !includeSpoon {
		teacup.RemoveObjectsByName("spoon")
	}

	teacup.CenterOn(&vec3.Zero)
	teacup.Translate(&vec3.T{0, -teacup.Bounds.Ymin, 0})

	teacup.ScaleUniform(&vec3.Zero, scale/teacup.Bounds.Zmax)

	return teacup
}
