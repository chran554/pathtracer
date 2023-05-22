package obj

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_LoadPokemonTangela(t *testing.T) {
	t.Run("obj file: pokemon tangela - load", func(t *testing.T) {
		setTestResourcesRoot()
		obj := loadPokemonTangela()
		fmt.Printf("Facet structure to be tested: %+v\n", obj)
		require.NotNil(t, obj)
	})
}

func Test_PokemonTangela(t *testing.T) {
	t.Run("obj file: pokemon tangela", func(t *testing.T) {
		setTestResourcesRoot()
		obj := loadPokemonTangela()
		fmt.Printf("Facet structure to be tested: %+v\n", obj)

		require.NotNil(t, obj)
		assertFacetStructure(t, obj, "tangela", "", "", 0, 7)

		assertSubstructure(t, obj, "", "foot_left", "foot_left", -1, 0)
		assertSubstructure(t, obj, "", "foot_right", "foot_right", -1, 0)
		assertSubstructure(t, obj, "", "body", "body", -1, 0)
		assertSubstructure(t, obj, "", "hair_top_left", "hair", -1, 0)
		assertSubstructure(t, obj, "", "hair_top_right", "hair", -1, 0)
		assertSubstructure(t, obj, "", "hair_bottom_left", "hair", -1, 0)
		assertSubstructure(t, obj, "", "hair_bottom_right", "hair", -1, 0)
	})
}
