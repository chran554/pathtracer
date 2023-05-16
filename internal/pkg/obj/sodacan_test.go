package obj

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"pathtracer/internal/pkg/color"
	"testing"
)

func Test_LoadSodaCan(t *testing.T) {
	t.Run("obj file: soda can - load", func(t *testing.T) {
		setTestResourcesRoot()
		obj := loadSodaCan("test.jpg", color.White, 1.0)
		fmt.Printf("Facet structure to be tested: %+v\n", obj)
		require.NotNil(t, obj)
	})
}

func Test_SodaCan(t *testing.T) {
	t.Run("obj file: soda can", func(t *testing.T) {
		setTestResourcesRoot()
		obj := loadSodaCan("test.jpg", color.White, 1.0)
		fmt.Printf("Facet structure to be tested: %+v\n", obj)

		require.NotNil(t, obj)
		assertFacetStructure(t, obj, "soda_can", "", "", 0, 3)

		tab := getSubstructure(t, obj, "", "tab", "tab")
		lid := getSubstructure(t, obj, "", "lid", "lid")
		body := getSubstructure(t, obj, "", "body", "body")

		assertFacetStructure(t, tab, "", "tab", "tab", -1, 0)
		assertFacetStructure(t, lid, "", "lid", "lid", -1, 0)
		assertFacetStructure(t, body, "", "body", "body", -1, 0)
	})
}
