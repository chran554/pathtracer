package obj

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_LoadDiamond(t *testing.T) {
	t.Run("obj file: diamond", func(t *testing.T) {
		setTestResourcesRoot()
		obj := loadDiamond()
		fmt.Printf("Facet structure to be tested: %+v\n", obj)
		require.NotNil(t, obj)
	})
}

func Test_Diamond(t *testing.T) {
	t.Run("obj file: diamond", func(t *testing.T) {
		setTestResourcesRoot()
		obj := loadDiamond()
		fmt.Printf("Facet structure to be tested: %+v\n", obj)

		require.NotNil(t, obj)
		assertFacetStructure(t, obj, "diamond", "", "", 0, 3)

		crown := getSubstructure(t, obj, "", "crown", "diamond")
		girdle := getSubstructure(t, obj, "", "girdle", "diamond")
		pavilion := getSubstructure(t, obj, "", "pavilion", "diamond")

		assertFacetStructure(t, crown, "", "crown", "diamond", -1, 0)
		assertFacetStructure(t, girdle, "", "girdle", "diamond", -1, 0)
		assertFacetStructure(t, pavilion, "", "pavilion", "diamond", -1, 0)
	})
}
