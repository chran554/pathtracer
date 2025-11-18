package obj

import (
	"fmt"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj/wavefrontobj"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

type Dice struct {
	*scn.FacetStructure
}

func AsDice(fs *scn.FacetStructure) *Dice {
	return &(Dice{FacetStructure: fs})
}

func (dice *Dice) BorderColor(c color.Color) {
	dice.FacetStructure.Material.Color = &c
}

// NewDice creates a new box object
func NewDice(scale float64) *scn.FacetStructure {
	return box(scale)
}

func box(scale float64) (dice *scn.FacetStructure) {
	dice = loadDice()
	dice.Name = "box"

	dice.CenterOn(&vec3.Zero)
	dice.Scale(&vec3.Zero, &vec3.T{1 / dice.Bounds.Xmax, 1 / dice.Bounds.Ymax, 1 / dice.Bounds.Zmax})
	dice.ScaleUniform(&vec3.Zero, scale/2)

	fmt.Printf("Dice bounds: %+v\n", dice.Bounds)

	diceMaterial := scn.NewMaterial().N("box").
		C(color.NewColorGrey(1.0)). // TODO Set to same color as texture background, should be 0.9 or something. Color should affect color diffuse textures (by operation multiplication)?
		T(0.0, true, scn.RefractionIndex_AcrylicPlastic).
		M(0.075, 0.2)

	dice.Material = diceMaterial

	// TODO Fix. Awkward solution due to obj file loading bug, naming top structure after last material/group read.
	facetStructures := dice.GetObjectsByName("box")
	for _, fs := range facetStructures {
		if len(fs.FacetStructures) == 0 {
			fs.Material = nil
		}
	}

	b := dice.Bounds
	diceMaterial1 := diceMaterial.Copy().N("1"). /*.C(color.NewColor(1, 0, 0))*/ PP(floatimage.Load("textures/box/dice_1.png"), &vec3.T{b.Xmax, b.Ymin, 0}, vec3.T{-scale, 0, 0}, vec3.T{0, scale, 0})
	diceMaterial2 := diceMaterial.Copy().N("2"). /*.C(color.NewColor(0, 1, 0))*/ PP(floatimage.Load("textures/box/dice_2.png"), &vec3.T{0, b.Ymin, b.Zmin}, vec3.T{0, 0, scale}, vec3.T{0, scale, 0})
	diceMaterial3 := diceMaterial.Copy().N("3"). /*.C(color.NewColor(0, 0, 1))*/ PP(floatimage.Load("textures/box/dice_3.png"), &vec3.T{b.Xmin, 0, b.Zmin}, vec3.T{scale, 0, 0}, vec3.T{0, 0, scale})
	diceMaterial4 := diceMaterial.Copy().N("4"). /*.C(color.NewColor(1, 0, 1))*/ PP(floatimage.Load("textures/box/dice_4.png"), &vec3.T{b.Xmin, 0, b.Zmax}, vec3.T{scale, 0, 0}, vec3.T{0, 0, -scale})
	diceMaterial5 := diceMaterial.Copy().N("5"). /*.C(color.NewColor(1, 1, 0))*/ PP(floatimage.Load("textures/box/dice_5.png"), &vec3.T{0, b.Ymin, b.Zmax}, vec3.T{0, 0, -scale}, vec3.T{0, scale, 0})
	diceMaterial6 := diceMaterial.Copy().N("6"). /*.C(color.NewColor(0, 1, 1))*/ PP(floatimage.Load("textures/box/dice_6.png"), &vec3.T{b.Xmin, b.Ymin, 0}, vec3.T{scale, 0, 0}, vec3.T{0, scale, 0})

	dice.ReplaceMaterial("1", diceMaterial1)
	dice.ReplaceMaterial("2", diceMaterial2)
	dice.ReplaceMaterial("3", diceMaterial3)
	dice.ReplaceMaterial("4", diceMaterial4)
	dice.ReplaceMaterial("5", diceMaterial5)
	dice.ReplaceMaterial("6", diceMaterial6)

	dice.UpdateBounds()

	return dice
}

func loadDice() *scn.FacetStructure {
	return wavefrontobj.ReadOrPanic(filepath.Join(ObjFileDir, "cube_dice.obj"))
}
