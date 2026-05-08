package stl

import (
	"os"
	"testing"
)

func TestReadFacetStructure(t *testing.T) {
	asciiData := `solid Test
  facet normal 0.0 0.0 1.0
    outer loop
      vertex 0.0 0.0 0.0
      vertex 1.0 0.0 0.0
      vertex 0.0 1.0 0.0
    endloop
  endfacet
endsolid Test`

	filename := "test_facet_structure.stl"
	err := os.WriteFile(filename, []byte(asciiData), 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(filename)

	fs := ReadFacetStructureOrPanic(filename)
	if fs == nil {
		t.Fatal("ReadFacetStructureOrPanic returned nil")
	}

	if len(fs.Facets) != 1 {
		t.Errorf("expected 1 facet, got %d", len(fs.Facets))
	}

	if fs.Name != "test" {
		t.Errorf("expected name 'test', got '%s'", fs.Name)
	}

	facet := fs.Facets[0]
	if len(facet.Vertices) != 3 {
		t.Errorf("expected 3 vertices, got %d", len(facet.Vertices))
	}

	if facet.Normal[2] != 1.0 {
		t.Errorf("expected normal Z=1.0, got %f", facet.Normal[2])
	}
}
