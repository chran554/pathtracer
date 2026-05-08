package obj

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
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
		assertFacetStructure(t, obj, "go_gopher", "", "", 0, 12)

		assertSubstructure(t, obj, "", "eye_ball_left", "eye_ball_left", -1, 0)
		assertSubstructure(t, obj, "", "eye_ball_right", "eye_ball_right", -1, 0)
		assertSubstructure(t, obj, "", "nose_tip", "nose_tip", -1, 0)
		assertSubstructure(t, obj, "", "nose", "nose", -1, 0)
		assertSubstructure(t, obj, "", "ear_inner", "ear_inner", -1, 0)
		assertSubstructure(t, obj, "", "paw", "paw", -1, 0)
		assertSubstructure(t, obj, "", "foot", "foot", -1, 0)
		assertSubstructure(t, obj, "", "eye_pupil_left", "eye_pupil_left", -1, 0)
		assertSubstructure(t, obj, "", "eye_pupil_right", "eye_pupil_right", -1, 0)
		assertSubstructure(t, obj, "", "tooth", "tooth", -1, 0)
		assertSubstructure(t, obj, "", "tail", "tail", -1, 0)
		assertSubstructure(t, obj, "", "body", "body", -1, 0)
	})
}
