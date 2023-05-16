package obj

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_LoadCylinder(t *testing.T) {
	t.Run("obj file: cylinder", func(t *testing.T) {
		setTestResourcesRoot()
		obj := loadCylinder()
		fmt.Printf("Facet structure to be tested: %+v\n", obj)
		require.NotNil(t, obj)
	})
}

func Test_Cylinder(t *testing.T) {
	t.Run("obj file: cylinder", func(t *testing.T) {
		setTestResourcesRoot()
		obj := loadCylinder()
		fmt.Printf("Facet structure to be tested: %+v\n", obj)

		require.NotNil(t, obj)
		assertFacetStructure(t, obj, "cylinder", "", "cylinder", -1, 0)
	})
}
