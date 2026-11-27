package obj

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SolidUtahTeapot_teapot(t *testing.T) {
	t.Run("obj file: solid utah teapot - teapot", func(t *testing.T) {
		setTestResourcesRoot()
		obj := NewSolidUtahTeapot(1.0, true, false)
		fmt.Printf("Facet structure to be tested: %+v\n", obj)

		filename := "utah_teapot_solid"

		require.NotNil(t, obj)
		assertFacetStructure(t, obj, filename, "", "Porcelain", 0, 1)

		teapot := getSubstructure(t, obj, "teapot", "", "")
		assertFacetStructure(t, teapot, "teapot", "", "", -1, 0)
	})
}

func Test_SolidUtahTeapot_lid(t *testing.T) {
	t.Run("obj file: solid utah teapot - lid", func(t *testing.T) {
		setTestResourcesRoot()
		obj := NewSolidUtahTeapot(1.0, false, true)
		fmt.Printf("Facet structure to be tested: %+v\n", obj)

		filename := "utah_teapot_solid"

		require.NotNil(t, obj)
		assertFacetStructure(t, obj, filename, "", "Porcelain", 0, 1)

		lid := getSubstructure(t, obj, "lid", "", "")
		assertFacetStructure(t, lid, "lid", "", "", -1, 0)
	})
}

func Test_SolidUtahTeapot_teapot_and_lid(t *testing.T) {
	t.Run("obj file: solid utah teapot - teapot, lid", func(t *testing.T) {
		setTestResourcesRoot()
		obj := NewSolidUtahTeapot(1.0, true, true)
		fmt.Printf("Facet structure to be tested: %+v\n", obj)

		filename := "utah_teapot_solid"

		require.NotNil(t, obj)
		assertFacetStructure(t, obj, filename, "", "Porcelain", 0, 2)

		teapot := getSubstructure(t, obj, "teapot", "", "")
		assertFacetStructure(t, teapot, "teapot", "", "", -1, 0)

		lid := getSubstructure(t, obj, "lid", "", "")
		assertFacetStructure(t, lid, "lid", "", "", -1, 0)
	})
}

func Test_SolidUtahTeapot_teacup(t *testing.T) {
	t.Run("obj file: solid utah teapot - cup", func(t *testing.T) {
		setTestResourcesRoot()
		obj := NewTeacup(1.0, true, false, false)
		fmt.Printf("Facet structure to be tested: %+v\n", obj)

		filename := "teacup"

		require.NotNil(t, obj)
		assertFacetStructure(t, obj, filename, "", "Porcelain", 0, 1)

		teacup := getSubstructure(t, obj, "teacup", "", "")
		assertFacetStructure(t, teacup, "teacup", "", "", -1, 0)
	})
}

func Test_SolidUtahTeapot_saucer(t *testing.T) {
	t.Run("obj file: solid utah teapot - saucer", func(t *testing.T) {
		setTestResourcesRoot()
		obj := NewTeacup(1.0, false, true, false)
		fmt.Printf("Facet structure to be tested: %+v\n", obj)

		filename := "teacup"

		require.NotNil(t, obj)
		assertFacetStructure(t, obj, filename, "", "Porcelain", 0, 1)

		saucer := getSubstructure(t, obj, "saucer", "", "")
		assertFacetStructure(t, saucer, "saucer", "", "", -1, 0)
	})
}

func Test_SolidUtahTeapot_spoon(t *testing.T) {
	t.Run("obj file: solid utah teapot - saucer", func(t *testing.T) {
		setTestResourcesRoot()
		obj := NewTeacup(1.0, false, false, true)
		fmt.Printf("Facet structure to be tested: %+v\n", obj)

		filename := "teacup"

		require.NotNil(t, obj)
		assertFacetStructure(t, obj, filename, "", "Porcelain", 0, 1)

		spoon := getSubstructure(t, obj, "spoon", "", "")
		assertFacetStructure(t, spoon, "spoon", "", "", -1, 0)
	})
}

func Test_SolidUtahTeapot_cup_and_saucer_and_spoon(t *testing.T) {
	t.Run("obj file: solid utah teapot - cup, saucer, lid", func(t *testing.T) {
		setTestResourcesRoot()
		obj := NewTeacup(1.0, true, true, true)
		fmt.Printf("Facet structure to be tested: %+v\n", obj)

		filename := "teacup"

		require.NotNil(t, obj)
		assertFacetStructure(t, obj, filename, "", "Porcelain", 0, 3)

		teacup := getSubstructure(t, obj, "teacup", "", "")
		assertFacetStructure(t, teacup, "teacup", "", "", -1, 0)

		saucer := getSubstructure(t, obj, "saucer", "", "")
		assertFacetStructure(t, saucer, "saucer", "", "", -1, 0)

		spoon := getSubstructure(t, obj, "spoon", "", "")
		assertFacetStructure(t, spoon, "spoon", "", "", -1, 0)
	})
}
