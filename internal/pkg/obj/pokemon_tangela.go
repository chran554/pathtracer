package obj

import (
	"path/filepath"
	"pathtracer/internal/pkg/fileformat/wavefront"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

func NewPokemonTangela(scale float64) *scene.FacetStructure {
	tangela := loadPokemonTangela()

	tangela.CenterOn(&vec3.Zero)
	tangela.Translate(&vec3.T{0, -tangela.Bounds.Ymin, 0})
	tangela.ScaleUniform(&vec3.Zero, scale/tangela.Bounds.Ymax)
	tangela.Scale(&vec3.Zero, &vec3.T{-1.0, 1.0, 1.0})

	body := tangela.GetFirstObjectBySubstructureName("body")
	body.ReplaceMaterial("body", scene.NewMaterial().N("body").SP(floatimage.LoadOrPanic("textures/pokemon/pokemon_tangela_texture.png"), body.Bounds.Center(), vec3.UnitZ.Scaled(-1), vec3.UnitY))

	return tangela
}

func loadPokemonTangela() *scene.FacetStructure {
	return wavefront.ReadFacetStructureOrPanic(filepath.Join(ObjFileDir, "pokemon_tangela.obj"))
}
