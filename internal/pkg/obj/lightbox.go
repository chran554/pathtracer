package obj

import (
	"math"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj/wavefrontobj"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

func NewLightBox(scale *vec3.T, c color.Color, emission float64, formFilename string) *scn.FacetStructure {
	lightBox := loadLightBox()

	lightBox.CenterOn(&vec3.Zero)
	lightBox.Scale(&vec3.Zero, &vec3.T{0.5 / lightBox.Bounds.Xmax, 0.5 / lightBox.Bounds.Ymax, 0.5 / lightBox.Bounds.Zmax})
	lightBox.RotateY(&vec3.Zero, -math.Pi/2)

	lightBox.GetFirstMaterialByName("cube").C(color.Black).E(color.Black, 0.0, true)
	lightBox.GetFirstMaterialByName("lightpanel").E(c, emission, true)
	lightBox.GetFirstMaterialByName("front").PP(floatimage.Load(formFilename), &vec3.T{0.5, -0.5, 0}, vec3.UnitX.Scaled(-1), vec3.UnitY)

	lightBox.Scale(&vec3.Zero, scale)

	return lightBox
}

func loadLightBox() *scn.FacetStructure {
	return wavefrontobj.ReadOrPanic(filepath.Join(ObjFileDir, "lightbox_freeform.obj"))
}
