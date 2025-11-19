package ply

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"pathtracer/internal/pkg/scene"
	"strings"

	"github.com/ungerik/go3d/float64/vec2"
	"github.com/ungerik/go3d/float64/vec3"
)

func ReadFacetStructureOrPanic(plyFilenamePath string) *scene.FacetStructure {
	objFile, err := os.Open(plyFilenamePath)
	if err != nil {
		message := fmt.Sprintf("Could not open ply file: '%s'\n%s\n", plyFilenamePath, err.Error())
		panic(message)
	}
	defer objFile.Close()

	ply, err := ReadFacetStructure(objFile)
	if err != nil {
		message := fmt.Sprintf("Could not read ply file: '%s'\n%s\n", objFile.Name(), err.Error())
		panic(message)
	}

	return ply
}

func ReadFacetStructure(file *os.File) (*scene.FacetStructure, error) {
	reader := bufio.NewReader(file)
	fmt.Printf("Reading ply file %s\n", file.Name())
	_, elementValues, err := Read(reader)
	if err != nil {
		return nil, fmt.Errorf("could read ply file '%s': %w", file.Name(), err)
	}

	fmt.Printf("Converting to face structure\n")
	var facetStructure *scene.FacetStructure
	if facetStructure, err = convertToFacetStructure(elementValues); err != nil {
		return nil, fmt.Errorf("could not convert ply file '%s' values to facet structure: %w", file.Name(), err)
	}

	fmt.Printf("Updating structure bounds\n")
	facetStructure.UpdateBounds()
	fmt.Printf("Updating structure normals\n")
	facetStructure.UpdateNormals()

	name := strings.TrimSuffix(strings.TrimSuffix(filepath.Base(file.Name()), ".ply"), ".PLY")
	facetStructure.Name = strings.ToLower(name)
	facetStructure.Material = scene.NewMaterial().N(strings.ToLower(name))

	return facetStructure, nil
}

func convertToFacetStructure(values []*Element) (*scene.FacetStructure, error) {
	var vertices []*vec3.T
	var vertexNormals []*vec3.T
	var textureCoordinates []*vec2.T
	var facets []*scene.Facet
	var err error

	if vertices, vertexNormals, textureCoordinates, err = extractVertices(values); err != nil {
		return nil, fmt.Errorf("could not extract verticies from ply values: %w", err)
	}

	if facets, err = extractFacets(values, vertices, vertexNormals, textureCoordinates); err != nil {
		return nil, fmt.Errorf("could not extract facets from ply values: %w", err)
	}

	return &scene.FacetStructure{Facets: facets}, nil
}

func extractFacets(elements []*Element, vertices []*vec3.T, vertexNormals []*vec3.T, textureCoordinates []*vec2.T) ([]*scene.Facet, error) {
	var facets []*scene.Facet

	for _, element := range elements {
		if (element.Name == "facet") || (element.Name == "face") {
			if element.ID != len(facets) {
				return nil, fmt.Errorf("illegal facet value index (%d), it do not match slice index (%d)", element.ID, len(vertices))
			}

			var facet scene.Facet
			for _, referenceID := range element.References {
				if element.ReferenceType == "vertex" {
					facet.Vertices = append(facet.Vertices, vertices[referenceID])
					if vertexNormals != nil {
						facet.VertexNormals = append(facet.VertexNormals, vertexNormals[referenceID])
					}
					if textureCoordinates != nil {
						facet.TextureCoordinates = append(facet.TextureCoordinates, textureCoordinates[referenceID])
					}
				} else {
					// If facet vertices is not a list of vertex indices in ply file, what is it then?
					return nil, fmt.Errorf("unknown ply format for face data (not a vertex index reference list): %+v", *element)
				}
			}
			facets = append(facets, &facet)
		}
	}

	return facets, nil
}

func extractVertices(elements []*Element) (vertices []*vec3.T, vertexNormals []*vec3.T, textureCoordinates []*vec2.T, err error) {
	anyVertex := false
	anyVertexNormal := false
	anyTextureCoordinate := false

	for _, e := range elements {
		if e.Name == "vertex" {
			if e.ID != len(vertices) {
				return nil, nil, nil, fmt.Errorf("illegal vertex value index (%d), it do not match slice index (%d)", e.ID, len(vertices))
			}

			vertex := vec3.T{}
			vertexNormal := vec3.T{}
			textureCoordinate := vec2.T{}
			for _, p := range e.Properties {
				switch p.Name {
				case "x":
					vertex[0] = getPropertyFloatValue(p)
					anyVertex = true
				case "y":
					vertex[1] = getPropertyFloatValue(p)
					anyVertex = true
				case "z":
					vertex[2] = getPropertyFloatValue(p)
					anyVertex = true

				case "nx":
					vertexNormal[0] = getPropertyFloatValue(p)
					anyVertexNormal = true
				case "ny":
					vertexNormal[1] = getPropertyFloatValue(p)
					anyVertexNormal = true
				case "nz":
					vertexNormal[2] = getPropertyFloatValue(p)
					anyVertexNormal = true

				case "u":
					textureCoordinate[0] = getPropertyFloatValue(p)
					anyTextureCoordinate = true
				case "v":
					textureCoordinate[1] = getPropertyFloatValue(p)
					anyTextureCoordinate = true
				}
			}

			vertices = append(vertices, &vertex)
			vertexNormals = append(vertexNormals, &vertexNormal)
			textureCoordinates = append(textureCoordinates, &textureCoordinate)
		}
	}

	if !anyVertex {
		vertices = nil
	}

	if !anyVertexNormal {
		vertexNormals = nil
	}

	if !anyTextureCoordinate {
		textureCoordinates = nil
	}

	return vertices, vertexNormals, textureCoordinates, nil
}

func getPropertyFloatValue(property *Property) float64 {
	var value float64

	if property.Type == PropertyTypeInt {
		value = float64(property.IntValue)
	} else if property.Type == PropertyTypeFloat {
		value = property.FloatValue
	}

	return value
}
