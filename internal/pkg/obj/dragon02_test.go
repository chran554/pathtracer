package obj

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_LoadDragon02(t *testing.T) {
	t.Run("obj file: dragon 02 - load", func(t *testing.T) {
		setTestResourcesRoot()
		obj := loadDragon02()
		fmt.Printf("Facet structure to be tested: %+v\n", obj)
		require.NotNil(t, obj)
	})
}

func Test_Dragon02(t *testing.T) {
	t.Run("obj file: dragon 02", func(t *testing.T) {
		setTestResourcesRoot()
		obj := loadDragon02()
		fmt.Printf("Facet structure to be tested: %+v\n", obj)

		require.NotNil(t, obj)
		assertFacetStructure(t, obj, "dragon_02", "", "", 0, 2)

		dragon := getSubstructure(t, obj, "dragon", "dragon", "skin")
		pillar := getSubstructure(t, obj, "pillar", "pillar", "pillar")

		assertFacetStructure(t, dragon, "dragon", "dragon", "skin", -1, 0)
		assertFacetStructure(t, pillar, "pillar", "pillar", "pillar", -1, 0)
	})
}
