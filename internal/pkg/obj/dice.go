package obj

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"
)

// NewDice creates a new cornell box (open in the back) with the center of the floor in origin (0,0,0).
// Left wall is blue and right wall is red.
// The scale is the total (width, height, depth) of the cornell box.
func NewDice(scale float64) *scn.FacetStructure {
	cornellBox := dice(scale)
	return cornellBox
}

func dice(scale float64) (dice *scn.FacetStructure) {
	var diceFilename = "cube_dice.obj"
	var diceFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/obj/" + diceFilename

	dice = ReadOrPanic(diceFilenamePath)
	dice.Name = "dice"

	dice.CenterOn(&vec3.Zero)
	dice.Scale(&vec3.Zero, &vec3.T{1 / dice.Bounds.Xmax, 1 / dice.Bounds.Ymax, 1 / dice.Bounds.Zmax})
	dice.ScaleUniform(&vec3.Zero, scale/2)

	fmt.Printf("Dice bounds: %+v\n", dice.Bounds)

	diceMaterial := scn.NewMaterial().N("dice").
		C(color.NewColorGrey(0.9)).
		T(0.0, true, scn.RefractionIndex_AcrylicPlastic).
		M(0.075, 0.2)

	dice.Material = diceMaterial

	b := dice.Bounds
	diceMaterial1 := diceMaterial.Copy().N("1"). /*.C(color.NewColor(1, 0, 0))*/ PP("textures/dice/dice_1.png", &vec3.T{b.Xmax, b.Ymin, 0}, vec3.T{-scale, 0, 0}, vec3.T{0, scale, 0})
	diceMaterial2 := diceMaterial.Copy().N("2"). /*.C(color.NewColor(0, 1, 0))*/ PP("textures/dice/dice_2.png", &vec3.T{0, b.Ymin, b.Zmin}, vec3.T{0, 0, scale}, vec3.T{0, scale, 0})
	diceMaterial3 := diceMaterial.Copy().N("3"). /*.C(color.NewColor(0, 0, 1))*/ PP("textures/dice/dice_3.png", &vec3.T{b.Xmin, 0, b.Zmin}, vec3.T{scale, 0, 0}, vec3.T{0, 0, scale})
	diceMaterial4 := diceMaterial.Copy().N("4"). /*.C(color.NewColor(1, 0, 1))*/ PP("textures/dice/dice_4.png", &vec3.T{b.Xmin, 0, b.Zmax}, vec3.T{scale, 0, 0}, vec3.T{0, 0, -scale})
	diceMaterial5 := diceMaterial.Copy().N("5"). /*.C(color.NewColor(1, 1, 0))*/ PP("textures/dice/dice_5.png", &vec3.T{0, b.Ymin, b.Zmax}, vec3.T{0, 0, -scale}, vec3.T{0, scale, 0})
	diceMaterial6 := diceMaterial.Copy().N("6"). /*.C(color.NewColor(0, 1, 1))*/ PP("textures/dice/dice_6.png", &vec3.T{b.Xmin, b.Ymin, 0}, vec3.T{scale, 0, 0}, vec3.T{0, scale, 0})

	dice.ReplaceMaterial("1", diceMaterial1)
	dice.ReplaceMaterial("2", diceMaterial2)
	dice.ReplaceMaterial("3", diceMaterial3)
	dice.ReplaceMaterial("4", diceMaterial4)
	dice.ReplaceMaterial("5", diceMaterial5)
	dice.ReplaceMaterial("6", diceMaterial6)

	dice.UpdateBounds()

	return dice
}
