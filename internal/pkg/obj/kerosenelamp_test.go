package obj

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_LoadKeroseneLamp(t *testing.T) {
	t.Run("obj file: kerosene lamp - load", func(t *testing.T) {
		setTestResourcesRoot()
		obj := loadKeroseneLamp(1.0)
		fmt.Printf("Facet structure to be tested: %+v\n", obj)
		require.NotNil(t, obj)
	})
}

func Test_KeroseneLamp(t *testing.T) {
	t.Run("obj file: kerosene lamp", func(t *testing.T) {
		setTestResourcesRoot()
		obj := loadKeroseneLamp(1.0)
		fmt.Printf("Facet structure to be tested: %+v\n", obj)

		filename := "kerosene_lamp"

		require.NotNil(t, obj)
		assertFacetStructure(t, obj, filename, "", "", 0, 6)

		flame := getSubstructure(t, obj, "", "flame", "flame")
		wickHolder := getSubstructure(t, obj, "", "wick_holder", "wick_holder")
		knob := getSubstructure(t, obj, "", "knob", "knob")
		handle := getSubstructure(t, obj, "", "handle", "handle")
		base := getSubstructure(t, obj, "", "base", "base")
		glass := getSubstructure(t, obj, "", "glass", "glass")

		assertFacetStructure(t, flame, "", "flame", "flame", -1, 0)
		assertFacetStructure(t, wickHolder, "", "wick_holder", "wick_holder", -1, 0)
		assertFacetStructure(t, knob, "", "knob", "knob", -1, 0)
		assertFacetStructure(t, handle, "", "handle", "handle", -1, 0)
		assertFacetStructure(t, base, "", "base", "base", -1, 0)
		assertFacetStructure(t, glass, "", "glass", "glass", -1, 0)
	})
}
