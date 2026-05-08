package stl

import (
	"fmt"
	"os"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/scene"
	"strings"

	"github.com/ungerik/go3d/float64/vec3"
)

func ReadFacetStructureOrPanic(stlFilenamePath string) *scene.FacetStructure {
	stlFile, err := os.Open(stlFilenamePath)
	if err != nil {
		currentPath, err2 := filepath.Abs(".")
		if err2 != nil {
			currentPath = "[unknown]"
		}
		message := fmt.Sprintf("Could not open stl file '%s' at current path '%s': %s\n", stlFilenamePath, currentPath, err.Error())
		panic(message)
	}
	defer stlFile.Close()

	facetStructure, err := ReadFacetStructure(stlFile)
	if err != nil {
		message := fmt.Sprintf("Could not read stl file '%s': %s\n", stlFile.Name(), err.Error())
		panic(message)
	}

	return facetStructure
}

// ReadFacetStructure reads a stl file and returns a facet structure.
func ReadFacetStructure(stlFile *os.File) (*scene.FacetStructure, error) {
	stl, err := Read(stlFile)
	if err != nil {
		return nil, fmt.Errorf("could not read stl file '%s': %v", stlFile.Name(), err)
	}

	facetStructure := convertToFacetStructure(stl)

	facetStructure.UpdateBounds()
	facetStructure.UpdateNormals()

	// Set name of facet structure
	name := strings.TrimSuffix(strings.TrimSuffix(filepath.Base(stlFile.Name()), ".stl"), ".STL")
	if !stl.IsBinary && stl.Header != "" {
		name = stl.Header
	}
	facetStructure.Name = strings.ToLower(name)

	// Set material of facet structure with colors from stl header
	material := scene.NewMaterial().N(facetStructure.Name)
	if stl.Color != nil {
		material.Color = convertColor(stl.Color)
	}
	if stl.Material != nil {
		material.Color = convertColor(&stl.Material.Diffuse)
	}
	facetStructure.Material = material

	return facetStructure, nil
}

func convertToFacetStructure(stl *Stl) *scene.FacetStructure {
	facets := make([]*scene.Facet, 0, len(stl.Facets))
	vertexMap := make(map[*Vertex]*vec3.T)

	for _, stlFacet := range stl.Facets {
		sceneFacet := &scene.Facet{}
		// STL facets have exactly 3 vertices
		for i := 0; i < 3; i++ {
			v := stlFacet.Vertices[i]
			if _, exist := vertexMap[v]; !exist {
				vertexMap[v] = &vec3.T{float64(v.X), float64(v.Y), float64(v.Z)}
			}
			sceneFacet.Vertices = append(sceneFacet.Vertices, vertexMap[v])
		}
		// STL facets also have a normal
		sceneFacet.Normal = &vec3.T{float64(stlFacet.Normal.X), float64(stlFacet.Normal.Y), float64(stlFacet.Normal.Z)}

		facets = append(facets, sceneFacet)
	}

	return &scene.FacetStructure{Facets: facets}
}

func convertColor(c *Color) *color.Color {
	return color.NewColorRGBA(
		float64(c.R)/255.0,
		float64(c.G)/255.0,
		float64(c.B)/255.0,
		float64(c.A)/255.0,
	)
}
