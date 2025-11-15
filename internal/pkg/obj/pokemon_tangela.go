package obj

import (
	"path/filepath"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj/wavefrontobj"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

func NewPokemonTangela(scale float64) *scn.FacetStructure {
	tangela := loadPokemonTangela()

	tangela.CenterOn(&vec3.Zero)
	tangela.Translate(&vec3.T{0, -tangela.Bounds.Ymin, 0})
	tangela.ScaleUniform(&vec3.Zero, scale/tangela.Bounds.Ymax)
	tangela.Scale(&vec3.Zero, &vec3.T{-1.0, 1.0, 1.0})

	body := tangela.GetFirstObjectBySubstructureName("body")
	body.ReplaceMaterial("body", scn.NewMaterial().N("body").SP(floatimage.Load("textures/pokemon/pokemon_tangela_texture.png"), body.Bounds.Center(), vec3.UnitZ.Scaled(-1), vec3.UnitY))

	return tangela
}

func loadPokemonTangela() *scn.FacetStructure {
	return wavefrontobj.ReadOrPanic(filepath.Join(ObjFileDir, "pokemon_tangela.obj"))
}
