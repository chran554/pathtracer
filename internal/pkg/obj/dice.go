package obj

import (
	"fmt"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/fileformat/wavefront"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

type Dice struct {
	*scene.FacetStructure
}

func AsDice(fs *scene.FacetStructure) *Dice {
	return &(Dice{FacetStructure: fs})
}

func (dice *Dice) BorderColor(c *color.Color) {
	dice.FacetStructure.Material.Color = c
}

// NewDice creates a new dice centered around origin.
// The dice is scaled so that the longest side will be as long as the parameter scale.
func NewDice(scale float64, hires bool) *scene.FacetStructure {
	dice := loadDice(hires)

	dice.CenterOn(&vec3.Zero)
	dice.ScaleUniform(&vec3.Zero, 1/max(dice.Bounds.SizeX(), dice.Bounds.SizeY(), dice.Bounds.SizeZ()))
	dice.ScaleUniform(&vec3.Zero, scale)

	fmt.Printf("Dice bounds: %+v\n", dice.Bounds)

	//diceMaterial := scene.NewMaterial().N("dice").
	//	C(color.NewColorGrey(1.0)). // TODO Set to same color as texture background, should be 0.9 or something. Color should affect color diffuse textures (by operation multiplication)?
	//	C(color.NewColor(0.9, 0.6, 0.7)).
	//	M(0.045, 0.1).
	//	T(0.0, true, scene.RefractionIndex_AcrylicPlastic)
	//dice.Material = diceMaterial

	return dice
}

func loadDice(hires bool) *scene.FacetStructure {
	filename := "dice_lores.obj"
	if hires {
		filename = "dice_hires.obj"
	}
	return wavefront.ReadFacetStructureOrPanic(filepath.Join(ObjFileDir, "dice", filename))
}
