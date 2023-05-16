package obj

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_LoadStanfordBunny(t *testing.T) {
	t.Run("obj file: stanford bunny - load", func(t *testing.T) {
		setTestResourcesRoot()
		obj := loadStanfordBunny()
		fmt.Printf("Facet structure to be tested: %+v\n", obj)
		require.NotNil(t, obj)
	})
}

func Test_StanfordBunny(t *testing.T) {
	t.Run("obj file: stanford bunny", func(t *testing.T) {
		setTestResourcesRoot()
		obj := loadStanfordBunny()
		fmt.Printf("Facet structure to be tested: %+v\n", obj)

		require.NotNil(t, obj)
		assertFacetStructure(t, obj, "stanfordbunny", "", "stanfordbunny", -1, 0)
	})
}
