package obj

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_LoadGopher(t *testing.T) {
	t.Run("obj file: gogopher - load", func(t *testing.T) {
		setTestResourcesRoot()
		obj := loadGopher()
		fmt.Printf("Facet structure to be tested: %+v\n", obj)
		require.NotNil(t, obj)
	})
}

func Test_Gopher(t *testing.T) {
	t.Run("obj file: gogopher", func(t *testing.T) {
		setTestResourcesRoot()
		obj := loadGopher()
		fmt.Printf("Facet structure to be tested: %+v\n", obj)

		require.NotNil(t, obj)
		assertFacetStructure(t, obj, "go_gopher", "", "", 0, 10)

		assertSubstructure(t, obj, "", "", "eye_ball", -1, 0)
		assertSubstructure(t, obj, "", "", "nose_tip", -1, 0)
		assertSubstructure(t, obj, "", "", "nose", -1, 0)
		assertSubstructure(t, obj, "", "", "ear_inner", -1, 0)
		assertSubstructure(t, obj, "", "", "paw", -1, 0)
		assertSubstructure(t, obj, "", "", "foot", -1, 0)
		assertSubstructure(t, obj, "", "", "eye_pupil", -1, 0)
		assertSubstructure(t, obj, "", "", "tooth", -1, 0)
		assertSubstructure(t, obj, "", "", "tail", -1, 0)
		assertSubstructure(t, obj, "", "", "body", -1, 0)
	})
}
