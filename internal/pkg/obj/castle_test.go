package obj

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_LoadCastle(t *testing.T) {
	t.Run("obj file: castle", func(t *testing.T) {
		setTestResourcesRoot()
		obj := loadCastle(1.0)
		fmt.Printf("Facet structure to be tested: %+v\n", obj)
		require.NotNil(t, obj)
	})
}

func Test_Castle(t *testing.T) {
	t.Run("obj file: castle", func(t *testing.T) {
		setTestResourcesRoot()
		obj := loadCastle(1.0)
		fmt.Printf("Facet structure to be tested: %+v\n", obj)

		require.NotNil(t, obj)
		assertFacetStructure(t, obj, "castle", "", "", 0, 13)

		wall := getSubstructure(t, obj, "", "wall", "")
		chapel := getSubstructure(t, obj, "", "chapel", "")
		house := getSubstructure(t, obj, "", "house", "")
		hall := getSubstructure(t, obj, "", "hall", "")
		house_tower := getSubstructure(t, obj, "", "house_tower", "")
		tower_middle_tall := getSubstructure(t, obj, "", "tower_middle_tall", "")
		tower_middle_short := getSubstructure(t, obj, "", "tower_middle_short", "")
		tower_entrance_left := getSubstructure(t, obj, "", "tower_entrance_left", "")
		tower_entrance_right := getSubstructure(t, obj, "", "tower_entrance_right", "")
		tower_front_left := getSubstructure(t, obj, "", "tower_front_left", "")
		tower_front_right := getSubstructure(t, obj, "", "tower_front_right", "")
		tower_back_left := getSubstructure(t, obj, "", "tower_back_left", "")
		tower_back_right := getSubstructure(t, obj, "", "tower_back_right", "")

		assertFacetStructure(t, wall, "", "wall", "", 0, 4)
		assertFacetStructure(t, chapel, "", "chapel", "", 0, 8)
		assertFacetStructure(t, house, "", "house", "", 0, 3)
		assertFacetStructure(t, hall, "", "hall", "", 0, 9)
		assertFacetStructure(t, house_tower, "", "house_tower", "", 0, 5)

		assertFacetStructure(t, tower_middle_tall, "", "tower_middle_tall", "", 0, 6)
		assertFacetStructure(t, tower_middle_short, "", "tower_middle_short", "", 0, 6)
		assertFacetStructure(t, tower_entrance_left, "", "tower_entrance_left", "", 0, 4)
		assertFacetStructure(t, tower_entrance_right, "", "tower_entrance_right", "", 0, 4)
		assertFacetStructure(t, tower_front_left, "", "tower_front_left", "", 0, 4)
		assertFacetStructure(t, tower_front_right, "", "tower_front_right", "", 0, 4)
		assertFacetStructure(t, tower_back_left, "", "tower_back_left", "", 0, 4)
		assertFacetStructure(t, tower_back_right, "", "tower_back_right", "", 0, 6)

		assertSubstructure(t, wall, "", "", "granite", -1, 0)
		assertSubstructure(t, wall, "", "", "erroded_cupper", -1, 0)
		assertSubstructure(t, wall, "", "", "sandstone", -1, 0)
		assertSubstructure(t, wall, "", "", "gold", -1, 0)

		assertSubstructure(t, chapel, "", "", "gold", -1, 0)
		assertSubstructure(t, chapel, "", "", "erroded_cupper", -1, 0)
		assertSubstructure(t, chapel, "", "", "colored_glass", -1, 0)
		assertSubstructure(t, chapel, "", "", "door", -1, 0)
		assertSubstructure(t, chapel, "", "", "bronze", -1, 0)
		assertSubstructure(t, chapel, "", "", "sandstone", -1, 0)
		assertSubstructure(t, chapel, "", "", "chapel_light", -1, 0)
		assertSubstructure(t, chapel, "", "", "chapel_tower_light", -1, 0)

		assertSubstructure(t, house, "", "", "gold", -1, 0)
		assertSubstructure(t, house, "", "", "granite", -1, 0)
		assertSubstructure(t, house, "", "", "erroded_cupper", -1, 0)

		assertSubstructure(t, hall, "", "", "glass", -1, 0)
		assertSubstructure(t, hall, "", "", "bronze", -1, 0)
		assertSubstructure(t, hall, "", "", "sandstone", -1, 0)
		assertSubstructure(t, hall, "", "", "hall_tower_left_light", -1, 0)
		assertSubstructure(t, hall, "", "", "granite", -1, 0)
		assertSubstructure(t, hall, "", "", "hall_light", -1, 0)
		assertSubstructure(t, hall, "", "", "erroded_cupper", -1, 0)
		assertSubstructure(t, hall, "", "", "door", -1, 0)
		assertSubstructure(t, hall, "", "", "hall_tower_right_light", -1, 0)

		assertSubstructure(t, house, "", "", "erroded_cupper", -1, 0)
		assertSubstructure(t, house, "", "", "gold", -1, 0)
		assertSubstructure(t, house, "", "", "granite", -1, 0)

		assertSubstructure(t, tower_middle_tall, "", "", "door", -1, 0)
		assertSubstructure(t, tower_middle_tall, "", "", "glass", -1, 0)
		assertSubstructure(t, tower_middle_tall, "", "", "gold", -1, 0)
		assertSubstructure(t, tower_middle_tall, "", "", "sandstone", -1, 0)
		assertSubstructure(t, tower_middle_tall, "", "", "erroded_cupper", -1, 0)
		assertSubstructure(t, tower_middle_tall, "", "", "tower_middle_tall_light", -1, 0)

		assertSubstructure(t, tower_middle_short, "", "", "tower_middle_short_light", -1, 0)
		assertSubstructure(t, tower_middle_short, "", "", "granite", -1, 0)
		assertSubstructure(t, tower_middle_short, "", "", "glass", -1, 0)
		assertSubstructure(t, tower_middle_short, "", "", "erroded_cupper", -1, 0)
		assertSubstructure(t, tower_middle_short, "", "", "sandstone", -1, 0)
		assertSubstructure(t, tower_middle_short, "", "", "gold", -1, 0)

		assertSubstructure(t, tower_entrance_left, "", "", "gold", -1, 0)
		assertSubstructure(t, tower_entrance_left, "", "", "granite", -1, 0)
		assertSubstructure(t, tower_entrance_left, "", "", "sandstone", -1, 0)
		assertSubstructure(t, tower_entrance_left, "", "", "erroded_cupper", -1, 0)

		assertSubstructure(t, tower_entrance_right, "", "", "erroded_cupper", -1, 0)
		assertSubstructure(t, tower_entrance_right, "", "", "granite", -1, 0)
		assertSubstructure(t, tower_entrance_right, "", "", "gold", -1, 0)
		assertSubstructure(t, tower_entrance_right, "", "", "sandstone", -1, 0)

		assertSubstructure(t, tower_front_left, "", "", "sandstone", -1, 0)
		assertSubstructure(t, tower_front_left, "", "", "erroded_cupper", -1, 0)
		assertSubstructure(t, tower_front_left, "", "", "gold", -1, 0)
		assertSubstructure(t, tower_front_left, "", "", "granite", -1, 0)

		assertSubstructure(t, tower_front_right, "", "", "gold", -1, 0)
		assertSubstructure(t, tower_front_right, "", "", "erroded_cupper", -1, 0)
		assertSubstructure(t, tower_front_right, "", "", "sandstone", -1, 0)
		assertSubstructure(t, tower_front_right, "", "", "granite", -1, 0)

		assertSubstructure(t, tower_back_left, "", "", "erroded_cupper", -1, 0)
		assertSubstructure(t, tower_back_left, "", "", "gold", -1, 0)
		assertSubstructure(t, tower_back_left, "", "", "granite", -1, 0)
		assertSubstructure(t, tower_back_left, "", "", "sandstone", -1, 0)

		assertSubstructure(t, tower_back_right, "", "", "sandstone", -1, 0)
		assertSubstructure(t, tower_back_right, "", "", "erroded_cupper", -1, 0)
		assertSubstructure(t, tower_back_right, "", "", "tower_back_right_light", -1, 0)
		assertSubstructure(t, tower_back_right, "", "", "glass", -1, 0)
		assertSubstructure(t, tower_back_right, "", "", "granite", -1, 0)
		assertSubstructure(t, tower_back_right, "", "", "gold", -1, 0)
	})
}
