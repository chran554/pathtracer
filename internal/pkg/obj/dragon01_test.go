package obj

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_LoadDragon01(t *testing.T) {
	t.Run("obj file: dragon 01 - load", func(t *testing.T) {
		setTestResourcesRoot()
		obj := loadDragon01()
		fmt.Printf("Facet structure to be tested: %+v\n", obj)
		require.NotNil(t, obj)
	})
}

func Test_Dragon01(t *testing.T) {
	t.Run("obj file: dragon 01", func(t *testing.T) {
		setTestResourcesRoot()
		obj := loadDragon01()
		fmt.Printf("Facet structure to be tested: %+v\n", obj)

		require.NotNil(t, obj)
		assertFacetStructure(t, obj, "dragon", "", "", -1, 0)
	})
}
