package obj

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_LoadDice(t *testing.T) {
	setTestResourcesRoot()
	obj := loadDice(true)
	fmt.Printf("Facet structure to be tested: %+v\n", obj)
	require.NotNil(t, obj)
}

func Test_Dice(t *testing.T) {
	setTestResourcesRoot()
	obj := loadDice(true)
	fmt.Printf("Facet structure to be tested: %+v\n", obj)

	require.NotNil(t, obj)
	assertFacetStructure(t, obj, "Dice", "", "", 0, 7)

	dice := getSubstructure(t, obj, "", "", "dice")
	side1 := getSubstructure(t, obj, "", "", "dice_face1")
	side2 := getSubstructure(t, obj, "", "", "dice_face2")
	side3 := getSubstructure(t, obj, "", "", "dice_face3")
	side4 := getSubstructure(t, obj, "", "", "dice_face4")
	side5 := getSubstructure(t, obj, "", "", "dice_face5")
	side6 := getSubstructure(t, obj, "", "", "dice_face6")

	assertFacetStructure(t, dice, "", "", "dice", -1, 0)
	assertFacetStructure(t, side1, "", "", "dice_face1", -1, 0)
	assertFacetStructure(t, side2, "", "", "dice_face2", -1, 0)
	assertFacetStructure(t, side3, "", "", "dice_face3", -1, 0)
	assertFacetStructure(t, side4, "", "", "dice_face4", -1, 0)
	assertFacetStructure(t, side5, "", "", "dice_face5", -1, 0)
	assertFacetStructure(t, side6, "", "", "dice_face6", -1, 0)
}
