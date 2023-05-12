package obj

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_LoadDrWhoAngel(t *testing.T) {
	t.Run("obj file: dr who angel - load", func(t *testing.T) {
		setTestResourcesRoot()
		obj := loadDrWhoAngel()
		fmt.Printf("Facet structure to be tested: %+v\n", obj)
		require.NotNil(t, obj)
	})
}

func Test_DrWhoAngel(t *testing.T) {
	t.Run("obj file: dr who angel", func(t *testing.T) {
		setTestResourcesRoot()
		obj := loadDrWhoAngel()
		fmt.Printf("Facet structure to be tested: %+v\n", obj)

		filename := "drwhoangel"

		require.NotNil(t, obj)
		assertFacetStructure(t, obj, filename, "", "", 0, 2)

		pillar := getSubstructure(t, obj, "pillar", "", "pillar")
		angel := getSubstructure(t, obj, "angel", "", "angel")

		assertFacetStructure(t, pillar, "pillar", "", "pillar", -1, 0)
		assertFacetStructure(t, angel, "angel", "", "angel", -1, 0)
	})
}
