package obj

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/ungerik/go3d/float64/vec3"
	"testing"
)

func Test_LoadCornellBox(t *testing.T) {
	t.Run("obj file: cornell box", func(t *testing.T) {
		setTestResourcesRoot()
		obj := loadCornellBox(&vec3.UnitXYZ, false, 1.0)
		fmt.Printf("Facet structure to be tested: %+v\n", obj)
		require.NotNil(t, obj)
	})
}

func Test_CornellBox(t *testing.T) {
	t.Run("obj file: cornell box", func(t *testing.T) {
		setTestResourcesRoot()
		obj := loadCornellBox(&vec3.UnitXYZ, false, 1.0)
		fmt.Printf("Facet structure to be tested: %+v\n", obj)

		require.NotNil(t, obj)
		assertFacetStructure(t, obj, "CornellBox", "", "", 0, 9)

		left := getSubstructure(t, obj, "", "Left", "Left")
		right := getSubstructure(t, obj, "", "Right", "Right")
		ceiling := getSubstructure(t, obj, "", "Ceiling", "Ceiling")
		floor := getSubstructure(t, obj, "", "Floor", "Floor")
		back := getSubstructure(t, obj, "", "Back", "Back")

		lamp1 := getSubstructure(t, obj, "", "Lamp_1_-_left_away", "Lamp")
		lamp2 := getSubstructure(t, obj, "", "Lamp_2_-_left_close", "Lamp")
		lamp3 := getSubstructure(t, obj, "", "Lamp_3_-_right_away", "Lamp")
		lamp4 := getSubstructure(t, obj, "", "Lamp_4_-_right_close", "Lamp")

		assertFacetStructure(t, left, "", "Left", "Left", -1, 0)
		assertFacetStructure(t, right, "", "Right", "Right", -1, 0)
		assertFacetStructure(t, ceiling, "", "Ceiling", "Ceiling", -1, 0)
		assertFacetStructure(t, floor, "", "Floor", "Floor", -1, 0)
		assertFacetStructure(t, back, "", "Back", "Back", -1, 0)

		assertFacetStructure(t, lamp1, "", "Lamp_1_-_left_away", "Lamp", -1, 0)
		assertFacetStructure(t, lamp2, "", "Lamp_2_-_left_close", "Lamp", -1, 0)
		assertFacetStructure(t, lamp3, "", "Lamp_3_-_right_away", "Lamp", -1, 0)
		assertFacetStructure(t, lamp4, "", "Lamp_4_-_right_close", "Lamp", -1, 0)
	})
}
