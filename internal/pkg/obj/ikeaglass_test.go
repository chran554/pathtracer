package obj

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_IkeaGlassSkoja_glass(t *testing.T) {
	t.Run("obj file: ikea glass skoja - glass", func(t *testing.T) {
		setTestResourcesRoot()
		obj := NewGlassIkeaSkoja(1.0, false)
		fmt.Printf("Facet structure to be tested: %+v\n", obj)

		require.NotNil(t, obj)
		assertFacetStructure(t, obj, "ikea_skoja", "", "", 0, 1)

		glass := getSubstructure(t, obj, "", "glass", "glass")
		assertFacetStructure(t, glass, "", "glass", "glass", -1, 0)
	})
}

func Test_IkeaGlassSkoja_glass_and_liquid(t *testing.T) {
	t.Run("obj file: ikea glass skoja - glass, liquid", func(t *testing.T) {
		setTestResourcesRoot()
		obj := NewGlassIkeaSkoja(1.0, true)
		fmt.Printf("Facet structure to be tested: %+v\n", obj)

		require.NotNil(t, obj)
		assertFacetStructure(t, obj, "ikea_skoja", "", "", 0, 2)

		glass := getSubstructure(t, obj, "", "glass", "glass")
		assertFacetStructure(t, glass, "", "glass", "glass", -1, 0)

		liquid := getSubstructure(t, obj, "", "liquid", "red juice")
		assertFacetStructure(t, liquid, "", "liquid", "red juice", -1, 0)
	})
}

func Test_IkeaGlassPokal(t *testing.T) {
	t.Run("obj file: ikea glass pokal - glass", func(t *testing.T) {
		setTestResourcesRoot()
		obj := NewGlassIkeaPokal(1.0)
		fmt.Printf("Facet structure to be tested: %+v\n", obj)

		require.NotNil(t, obj)
		assertFacetStructure(t, obj, "ikea_pokal", "glass", "glass", -1, 0)
	})
}
