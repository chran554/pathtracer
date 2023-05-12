package obj

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_LoadLamppost(t *testing.T) {
	t.Run("obj file: lamppost - load", func(t *testing.T) {
		setTestResourcesRoot()
		obj := loadLamppost(1.0)
		fmt.Printf("Facet structure to be tested: %+v\n", obj)
		require.NotNil(t, obj)
	})
}

func Test_Lamppost(t *testing.T) {
	t.Run("obj file: lamppost", func(t *testing.T) {
		setTestResourcesRoot()
		obj := loadLamppost(1.0)
		fmt.Printf("Facet structure to be tested: %+v\n", obj)

		require.NotNil(t, obj)
		assertFacetStructure(t, obj, "lamppost", "", "", 0, 11)

		post := getSubstructure(t, obj, "", "post", "post")
		lamp_0 := getSubstructure(t, obj, "", "lamp_0", "lamp_0")
		lamp_1 := getSubstructure(t, obj, "", "lamp_1", "lamp_1")
		lamp_2 := getSubstructure(t, obj, "", "lamp_2", "lamp_2")
		lamp_3 := getSubstructure(t, obj, "", "lamp_3", "lamp_3")
		lamp_1_arm := getSubstructure(t, obj, "", "lamp_1_arm", "lamp_1_arm")
		lamp_1_attachment := getSubstructure(t, obj, "", "lamp_1_attachment", "lamp_1_attachment")
		lamp_2_arm := getSubstructure(t, obj, "", "lamp_2_arm", "lamp_2_arm")
		lamp_2_attachment := getSubstructure(t, obj, "", "lamp_2_attachment", "lamp_2_attachment")
		lamp_3_arm := getSubstructure(t, obj, "", "lamp_3_arm", "lamp_3_arm")
		lamp_3_attachment := getSubstructure(t, obj, "", "lamp_3_attachment", "lamp_3_attachment")

		assertFacetStructure(t, post, "", "post", "post", -1, 0)
		assertFacetStructure(t, lamp_0, "", "lamp_0", "lamp_0", -1, 0)
		assertFacetStructure(t, lamp_1, "", "lamp_1", "lamp_1", -1, 0)
		assertFacetStructure(t, lamp_2, "", "lamp_2", "lamp_2", -1, 0)
		assertFacetStructure(t, lamp_3, "", "lamp_3", "lamp_3", -1, 0)
		assertFacetStructure(t, lamp_1_arm, "", "lamp_1_arm", "lamp_1_arm", -1, 0)
		assertFacetStructure(t, lamp_1_attachment, "", "lamp_1_attachment", "lamp_1_attachment", -1, 0)
		assertFacetStructure(t, lamp_2_arm, "", "lamp_2_arm", "lamp_2_arm", -1, 0)
		assertFacetStructure(t, lamp_2_attachment, "", "lamp_2_attachment", "lamp_2_attachment", -1, 0)
		assertFacetStructure(t, lamp_3_arm, "", "lamp_3_arm", "lamp_3_arm", -1, 0)
		assertFacetStructure(t, lamp_3_attachment, "", "lamp_3_attachment", "lamp_3_attachment", -1, 0)
	})
}
