package ply

import (
	"fmt"
	"os"
	"testing"
)

func Test_ReadPly(t *testing.T) {
	plyFilenamePath := "/Users/christian/projects/code/go/pathtracer/objects/ply/dart.ply"
	//plyFilenamePath := "/Users/christian/projects/code/go/pathtracer/objects/ply/icosahedron.ply"

	plyFile, err := os.Open(plyFilenamePath)
	if err != nil {
		fmt.Printf("ouupps, something went wrong loading file: '%s'\n%s\n", plyFilenamePath, err.Error())
	}
	defer plyFile.Close()

	facetStructure, err := Read(plyFile)
	if err != nil {
		fmt.Printf("could not test read ply file '%s': %s\n", plyFilenamePath, err.Error())
	}

	facetStructure.UpdateNormals()
	facetStructure.UpdateVertexNormals(false)

	expectedFacetStructureName := "dart"
	if facetStructure.Name != expectedFacetStructureName {
		t.Logf("the facet structure name '%s' was not based on the file name '%s' correctly, expected '%s'", facetStructure.Name, plyFilenamePath, expectedFacetStructureName)
		t.Fail()
	}

	expectedAmountFacets := 6
	if len(facetStructure.Facets) != expectedAmountFacets {
		t.Logf("the actual amount facets %d, do not match expected %d", len(facetStructure.Facets), expectedAmountFacets)
		t.Fail()
	}
}

func Test_GetElementReferenceType(t *testing.T) {
	s0 := ""
	s1 := "text"
	s2 := "text_text"
	s3 := "index"
	s4 := "text_index"
	s5 := "text_text_text_index"
	s6 := "text_text_text_indices"
	s7 := "text_text_text_inDicEs"

	if getElementReferenceType(s0) != s0 {
		t.Fail()
	}
	if getElementReferenceType(s1) != s1 {
		t.Fail()
	}
	if getElementReferenceType(s2) != s2 {
		t.Fail()
	}
	if getElementReferenceType(s3) != s3 {
		t.Fail()
	}
	if getElementReferenceType(s4) != "text" {
		t.Fail()
	}
	if getElementReferenceType(s5) != "text_text_text" {
		t.Fail()
	}
	if getElementReferenceType(s6) != "text_text_text" {
		t.Fail()
	}
	if getElementReferenceType(s7) != "text_text_text" {
		t.Fail()
	}
}
