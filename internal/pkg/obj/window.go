package obj

import (
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/fileformat/wavefront"
	"pathtracer/internal/pkg/scene"
	"pathtracer/internal/pkg/util"

	"github.com/ungerik/go3d/float64/vec3"
)

// NewWindow creates a window with proportions 2/3:1 (width:height) that is scaled with parameter.
func NewWindow(scale float64) *scene.FacetStructure {
	window := loadWindow()

	window.CenterOn(&vec3.Zero)
	window.RotateY(&vec3.Zero, util.DegToRad(180))
	window.Translate(&vec3.T{0, -window.Bounds.Ymin, 0})

	window.ScaleUniform(&vec3.Zero, scale/window.Bounds.Ymax)

	glassMaterial := scene.NewMaterialGlass("glass")
	glassMaterial.SolidObject = false

	paintMaterial := scene.NewMaterial().N("paint").
		C(color.NewColorGrey(0.9)).
		M(0.1, 0.6)

	brassMaterial := scene.NewMaterial().N("brass").
		C(color.NewColor(0.8, 0.7, 0.15)).
		M(0.8, 0.4)

	window.ReplaceMaterial("glass", glassMaterial)
	window.ReplaceMaterial("frame", paintMaterial)
	window.ReplaceMaterial("inner", paintMaterial)
	window.ReplaceMaterial("latch", brassMaterial)
	window.ReplaceMaterial("screw", brassMaterial)
	window.ReplaceMaterial("hook", brassMaterial)

	return window
}

func loadWindow() *scene.FacetStructure {
	return wavefront.ReadFacetStructureOrPanic(filepath.Join(ObjFileDir, "window.obj"))
}
