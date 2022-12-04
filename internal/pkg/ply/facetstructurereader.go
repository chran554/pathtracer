package ply

import (
	"bufio"
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"os"
	"path/filepath"
	"pathtracer/internal/pkg/scene"
	"strings"
)

func ReadPlyFile(file *os.File) (*scene.FacetStructure, error) {
	reader := bufio.NewReader(file)
	_, elementValues, err := ReadPly(reader)

	var facetStructure *scene.FacetStructure
	if facetStructure, err = convertToFacetStructure(elementValues); err != nil {
		return nil, fmt.Errorf("could not convert ply file '%s' values to facet structure: %w", file.Name(), err)
	}

	facetStructure.UpdateBounds()
	facetStructure.UpdateNormals()

	facetStructure.Name = strings.TrimSuffix(strings.TrimSuffix(filepath.Base(file.Name()), ".ply"), ".PLY")

	return facetStructure, nil
}

func convertToFacetStructure(values []*ElementValue) (*scene.FacetStructure, error) {
	var vertices []*vec3.T
	var facets []*scene.Facet
	var err error

	if vertices, err = extractVertices(values); err != nil {
		return nil, fmt.Errorf("could not extract verticies from ply values: %w", err)
	}

	if facets, err = extractFacets(values, vertices); err != nil {
		return nil, fmt.Errorf("could not extract facets from ply values: %w", err)
	}

	return &scene.FacetStructure{Facets: facets}, nil
}

func extractFacets(values []*ElementValue, vertices []*vec3.T) ([]*scene.Facet, error) {
	var facets []*scene.Facet

	for _, value := range values {
		if (value.name == "facet") || (value.name == "face") {
			if value.index != len(facets) {
				return nil, fmt.Errorf("illegal facet value index (%d), it do not match slice index (%d)", value.index, len(vertices))
			}

			var facet scene.Facet
			for _, propertyValue := range value.values {
				if (propertyValue.valueType == IndexReference) && (propertyValue.name == "vertex") {
					facet.Vertices = append(facet.Vertices, vertices[propertyValue.intValue])
				} else {
					// If facet vertices is not a list of vertex indices in ply file, what is it then?
					return nil, fmt.Errorf("unknown ply format for face data (not a vertex index reference list): %+v", *value)
				}
			}
			facets = append(facets, &facet)
		}
	}

	return facets, nil
}

func extractVertices(values []*ElementValue) ([]*vec3.T, error) {
	var vertices []*vec3.T

	for _, value := range values {
		if value.name == "vertex" {
			if value.index != len(vertices) {
				return nil, fmt.Errorf("illegal vertex value index (%d), it do not match slice index (%d)", value.index, len(vertices))
			}

			vertex := vec3.T{}
			for _, propertyValue := range value.values {
				switch propertyValue.name {
				case "x":
					vertex[0] = getPropertyValue(propertyValue)
				case "y":
					vertex[1] = getPropertyValue(propertyValue)
				case "z":
					vertex[2] = getPropertyValue(propertyValue)
				}
			}
			vertices = append(vertices, &vertex)
		}
	}

	return vertices, nil
}

func getPropertyValue(propertyValue *PropertyValue) float64 {
	var value float64

	if propertyValue.valueType == Int {
		value = float64(propertyValue.intValue)
	} else if propertyValue.valueType == Float {
		value = propertyValue.floatValue
	}

	return value
}
