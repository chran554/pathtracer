package obj

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_LoadDice(t *testing.T) {
	t.Run("obj file: box", func(t *testing.T) {
		setTestResourcesRoot()
		obj := loadDice()
		fmt.Printf("Facet structure to be tested: %+v\n", obj)
		require.NotNil(t, obj)
	})
}

func Test_Dice(t *testing.T) {
	t.Run("obj file: box", func(t *testing.T) {
		setTestResourcesRoot()
		obj := loadDice()
		fmt.Printf("Facet structure to be tested: %+v\n", obj)

		require.NotNil(t, obj)
		assertFacetStructure(t, obj, "Cube", "", "", 0, 7)

		dice := getSubstructure(t, obj, "", "", "dice")
		side1 := getSubstructure(t, obj, "", "", "1")
		side2 := getSubstructure(t, obj, "", "", "2")
		side3 := getSubstructure(t, obj, "", "", "3")
		side4 := getSubstructure(t, obj, "", "", "4")
		side5 := getSubstructure(t, obj, "", "", "5")
		side6 := getSubstructure(t, obj, "", "", "6")

		assertFacetStructure(t, dice, "", "", "dice", -1, 0)
		assertFacetStructure(t, side1, "", "", "1", -1, 0)
		assertFacetStructure(t, side2, "", "", "2", -1, 0)
		assertFacetStructure(t, side3, "", "", "3", -1, 0)
		assertFacetStructure(t, side4, "", "", "4", -1, 0)
		assertFacetStructure(t, side5, "", "", "5", -1, 0)
		assertFacetStructure(t, side6, "", "", "6", -1, 0)
	})
}
