package obj

import (
	scn "pathtracer/internal/pkg/scene"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ungerik/go3d/float64/vec3"
)

func setTestResourcesRoot() {
	SetResourceRoot("../../..")
}

func assertSubstructure(t *testing.T, structure *scn.FacetStructure, objectName string, groupName string, materialName string, amountFacets int, amountSubstructures int) {
	substructure := getSubstructure(t, structure, objectName, groupName, materialName)
	assert.NotNilf(t, substructure, "Substructure \"%s\" of structure \"%s\" could not be found.", (&scn.FacetStructure{Name: objectName, SubstructureName: groupName, Material: &scn.Material{Name: materialName}}).StructureNames(), structure.StructureNames())

	assertFacetStructure(t, substructure, objectName, groupName, materialName, amountFacets, amountSubstructures)

}

func assertFacetStructure(t *testing.T, f *scn.FacetStructure, name, substructureName, materialName string, amountFacets, amountSubStructures int) {
	require.NotNil(t, f)

	assert.Equalf(t, name, f.Name, "Facet structure \"%s\" do not have expected name.", f.StructureNames())
	assert.Equalf(t, substructureName, f.SubstructureName, "Facet structure \"%s\" do not have expected sub structure name.", f.StructureNames())
	if f.Material == nil {
		assert.Equal(t, materialName, "")
	} else {
		assert.Equal(t, materialName, f.Material.Name)
	}

	if amountFacets == -1 {
		assert.Truef(t, len(f.Facets) > 0, "The amount of immediate facets on structure \"%s\" is expected to be strict larger than 0.", f.StructureNames())
	} else if amountFacets == 0 {
		assert.Equalf(t, 0, len(f.Facets), "The amount of immediate facets on structure \"%s\" is not 0 as expected.", f.StructureNames())
		assert.Nilf(t, f.Facets, "Facet slice is expected to be nil (not slice of size 0) when there are no facets on structure \"%s\".", f.StructureNames())
	} else {
		assert.Equalf(t, amountFacets, len(f.Facets), "The amount of immediate facets on structure \"%s\" do not match expected amount.", f.StructureNames())
	}

	if amountSubStructures == 0 {
		assert.Nilf(t, f.FacetStructures, "Amount substructures are %d and not expected 0 for facet structure \"%s\".", len(f.FacetStructures), f.StructureNames())
	} else {
		assert.NotNil(t, f.FacetStructures)
		assert.Equalf(t, amountSubStructures, len(f.FacetStructures), "Amount of substructures of facet structure \"%s\" differ from expected.", f.StructureNames())
	}

	if materialName == "" {
		assert.Nil(t, f.Material)
	} else {
		assert.NotNil(t, f.Material)
		assert.Equal(t, materialName, f.Material.Name)
	}

	// Get the recursive amount of unique vertices for the facet structure f
	vertices := make(map[*vec3.T]bool)
	facetStructures := []*scn.FacetStructure{f}
	for len(facetStructures) > 0 {
		facetStructure := facetStructures[0]
		facetStructures = facetStructures[1:]

		for _, facet := range facetStructure.Facets {
			for _, vertex := range facet.Vertices {
				vertices[vertex] = true
			}
		}

		facetStructures = append(facetStructures, facetStructure.FacetStructures...)
	}
}

func getSubstructure(t *testing.T, structure *scn.FacetStructure, objectName, groupName, materialName string) *scn.FacetStructure {
	var substructure *scn.FacetStructure

	for _, facetSubStructure := range structure.FacetStructures {
		if (objectName == facetSubStructure.Name) &&
			(groupName == facetSubStructure.SubstructureName) &&
			((facetSubStructure.Material == nil && materialName == "") || (facetSubStructure.Material != nil && facetSubStructure.Material.Name == materialName)) {
			substructure = facetSubStructure
		}
	}

	require.NotNil(t, substructure, "Could not find expected substructure \"%s::%s::%s\" of facet structure \"%s\".", objectName, groupName, materialName, structure.StructureNames())

	return substructure
}
