package obj

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_LoadWindow(t *testing.T) {
	t.Run("obj file: window - load", func(t *testing.T) {
		setTestResourcesRoot()
		obj := loadWindow()
		fmt.Printf("Facet structure to be tested: %+v\n", obj)
		require.NotNil(t, obj)
	})
}

func Test_Window(t *testing.T) {
	t.Run("obj file: window", func(t *testing.T) {
		setTestResourcesRoot()
		obj := loadWindow()
		fmt.Printf("Facet structure to be tested: %+v\n", obj)

		require.NotNil(t, obj)
		assertFacetStructure(t, obj, "window", "", "", 0, 6)

		frame := getSubstructure(t, obj, "", "frame", "frame")
		glass := getSubstructure(t, obj, "", "glass", "glass")
		hook := getSubstructure(t, obj, "", "hook", "hook")
		screw := getSubstructure(t, obj, "", "screw", "screw")
		latch := getSubstructure(t, obj, "", "latch", "latch")
		inner := getSubstructure(t, obj, "", "inner", "inner")

		assertFacetStructure(t, frame, "", "frame", "frame", -1, 0)
		assertFacetStructure(t, glass, "", "glass", "glass", -1, 0)
		assertFacetStructure(t, hook, "", "hook", "hook", -1, 0)
		assertFacetStructure(t, screw, "", "screw", "screw", -1, 0)
		assertFacetStructure(t, latch, "", "latch", "latch", -1, 0)
		assertFacetStructure(t, inner, "", "inner", "inner", -1, 0)
	})
}
