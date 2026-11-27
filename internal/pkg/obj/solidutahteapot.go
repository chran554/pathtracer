package obj

import (
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/fileformat/wavefront"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

func NewSolidUtahTeapot(scale float64, includeBody bool, includeLid bool) *scene.FacetStructure {
	utahTeaPot := wavefront.ReadFacetStructureOrPanic(filepath.Join(ObjFileDir, "utah_teapot_solid.obj"))

	if !includeBody {
		utahTeaPot.RemoveObjectsByName("teapot")
	}
	if !includeLid {
		utahTeaPot.RemoveObjectsByName("lid")
	}

	utahTeaPot.CenterOn(&vec3.Zero)
	utahTeaPot.Translate(&vec3.T{0, -utahTeaPot.Bounds.Ymin, 0})

	utahTeaPot.ScaleUniform(&vec3.Zero, scale/utahTeaPot.Bounds.Ymax)

	utahTeaPot.ClearMaterials()
	utahTeaPot.Material = porcelainMaterial()

	return utahTeaPot
}

func NewTeacup(scale float64, includeCup bool, includeSaucer bool, includeSpoon bool) *scene.FacetStructure {
	teacupSaucerSpoon := wavefront.ReadFacetStructureOrPanic(filepath.Join(ObjFileDir, "teacup.obj"))

	if includeCup {
		teacup := teacupSaucerSpoon.GetFirstObjectByName("teacup")
		teacup.ClearMaterials()
		teacup.Material = porcelainMaterial()
	} else {
		teacupSaucerSpoon.RemoveObjectsByName("teacup")
	}

	if includeSaucer {
		saucer := teacupSaucerSpoon.GetFirstObjectByName("saucer")
		saucer.ClearMaterials()
		saucer.Material = porcelainMaterial()
	} else {
		teacupSaucerSpoon.RemoveObjectsByName("saucer")
	}

	if includeSpoon {
		spoon := teacupSaucerSpoon.GetFirstObjectByName("spoon")
		spoon.ClearMaterials()
		spoon.Material = steelMaterial()
	} else {
		teacupSaucerSpoon.RemoveObjectsByName("spoon")
	}

	teacupSaucerSpoon.CenterOn(&vec3.Zero)
	teacupSaucerSpoon.Translate(&vec3.T{0, -teacupSaucerSpoon.Bounds.Ymin, 0})

	teacupSaucerSpoon.ScaleUniform(&vec3.Zero, scale/teacupSaucerSpoon.Bounds.Zmax)

	teacupSaucerSpoon.ClearMaterials()
	teacupSaucerSpoon.Material = porcelainMaterial()

	return teacupSaucerSpoon
}

func porcelainMaterial() *scene.Material {
	return scene.NewMaterial().
		N("Porcelain").
		C(color.NewColorGrey(0.90)).
		M(0, 0.02).
		T(0.0, true, scene.RefractionIndex_Porcelain)
}

func steelMaterial() *scene.Material {
	return scene.NewMaterial().
		N("steel").
		C(color.NewColorGrey(0.90)).
		M(1.0, 0.03).
		T(0.0, true, scene.RefractionIndex_Silver)
}
