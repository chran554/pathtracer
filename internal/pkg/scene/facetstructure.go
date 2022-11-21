package scene

import (
	"fmt"
	"github.com/ungerik/go3d/float64/mat3"
	"github.com/ungerik/go3d/float64/vec3"
)

type FacetStructure struct {
	Name             string            `json:"Name,omitempty"`
	SubstructureName string            `json:"SubstructureName,omitempty"`
	Material         *Material         `json:"Material,omitempty"`
	Facets           []*Facet          `json:"Facets,omitempty"`
	FacetStructures  []*FacetStructure `json:"FacetStructures,omitempty"`

	Bounds *Bounds `json:"-"` // Calculated attribute. See UpdateBounds(). Derived from all vertices in all sub facets recursively.
}

func (fs *FacetStructure) Initialize() {
	fs.UpdateNormals()
	fs.UpdateBounds()
	fs.UpdateMaterials()
	fs.InitializeProjection()
}

func (fs *FacetStructure) InitializeProjection() {
	for _, facetStructure := range fs.FacetStructures {
		facetStructure.InitializeProjection()
	}

	if fs.Material != nil {
		projection := fs.Material.Projection
		if projection != nil {
			projection.Initialize()
		}
	}
}

// UpdateMaterials propagates parent materials down in facet structure hierarchy to sub structures without explicit own material
func (fs *FacetStructure) UpdateMaterials() {
	for _, facetStructure := range fs.FacetStructures {
		if facetStructure.Material == nil {
			facetStructure.Material = fs.Material
		}

		facetStructure.UpdateMaterials()
	}
}

func (fs *FacetStructure) UpdateNormals() {
	for _, facet := range fs.Facets {
		facet.UpdateNormal()
	}

	for _, facetStructure := range fs.FacetStructures {
		facetStructure.UpdateNormals()
	}
}

func (fs *FacetStructure) UpdateBounds() *Bounds {
	bounds := NewBounds()

	for _, facet := range fs.Facets {
		facetBounds := facet.UpdateBounds()
		if !facetBounds.IsZeroBounds() {
			bounds.AddBounds(facetBounds)
		}
	}

	for _, facetStructure := range fs.FacetStructures {
		faceStructureBounds := facetStructure.UpdateBounds()
		if !faceStructureBounds.IsZeroBounds() {
			bounds.AddBounds(faceStructureBounds)
		}
	}

	fs.Bounds = &bounds
	return fs.Bounds
}

func (fs *FacetStructure) GetAmountFacets() int {
	amount := len(fs.Facets)

	for _, facetStructure := range fs.FacetStructures {
		amount += facetStructure.GetAmountFacets()
	}

	return amount
}

func (fs *FacetStructure) SplitMultiPointFacets() {
	for i := 0; i < len(fs.Facets); {
		facet := fs.Facets[i]

		if facet.IsMultiPointFacet() {
			splitFacets := facet.SplitMultiPointFacet()

			allFacets := append(fs.Facets[:i], append(splitFacets, fs.Facets[i+1:]...)...)
			fs.Facets = allFacets

			i += len(splitFacets)
		} else {
			i++
		}
	}

	for _, facetStructure := range fs.FacetStructures {
		facetStructure.SplitMultiPointFacets()
	}
}

func (fs *FacetStructure) String() string {
	name := "<noname>"
	if fs.Name != "" {
		name = fs.Name
	}

	subStructures := ""
	if len(fs.FacetStructures) > 0 {
		subStructures = "{"
		for i, facetStructure := range fs.FacetStructures {
			if i > 0 {
				subStructures = subStructures + ", "
			}
			subStructures = subStructures + facetStructure.String()
		}
		subStructures = subStructures + "}"
	}

	return fmt.Sprintf("%s (%d facets)%s", name, len(fs.Facets), subStructures)
}

func (fs *FacetStructure) RotateX(rotationOrigin *vec3.T, angle float64) {
	rotatedPoints := make(map[*vec3.T]bool)
	rotatedNormals := make(map[*vec3.T]bool)
	rotatedVertexNormals := make(map[*vec3.T]bool)

	rotationMatrix := mat3.T{}
	rotationMatrix.AssignXRotation(angle)

	fs.rotate(rotationOrigin, rotationMatrix, rotatedPoints, rotatedNormals, rotatedVertexNormals)
}

func (fs *FacetStructure) RotateY(rotationOrigin *vec3.T, angle float64) {
	rotatedPoints := make(map[*vec3.T]bool)
	rotatedNormals := make(map[*vec3.T]bool)
	rotatedVertexNormals := make(map[*vec3.T]bool)

	rotationMatrix := mat3.T{}
	rotationMatrix.AssignYRotation(angle)

	fs.rotate(rotationOrigin, rotationMatrix, rotatedPoints, rotatedNormals, rotatedVertexNormals)
}

func (fs *FacetStructure) RotateZ(rotationOrigin *vec3.T, angle float64) {
	rotatedPoints := make(map[*vec3.T]bool)
	rotatedNormals := make(map[*vec3.T]bool)
	rotatedVertexNormals := make(map[*vec3.T]bool)

	rotationMatrix := mat3.T{}
	rotationMatrix.AssignZRotation(angle)

	fs.rotate(rotationOrigin, rotationMatrix, rotatedPoints, rotatedNormals, rotatedVertexNormals)
}

func (fs *FacetStructure) rotate(rotationOrigin *vec3.T, rotationMatrix mat3.T, rotatedPoints map[*vec3.T]bool, rotatedNormals map[*vec3.T]bool, rotatedVertexNormals map[*vec3.T]bool) {
	for _, facet := range fs.Facets {
		facet.rotate(rotationOrigin, rotationMatrix, rotatedPoints, rotatedNormals, rotatedVertexNormals)
	}

	if len(fs.FacetStructures) > 0 {
		for _, facetStructure := range fs.FacetStructures {
			facetStructure.rotate(rotationOrigin, rotationMatrix, rotatedPoints, rotatedNormals, rotatedVertexNormals)
		}
	}
}

func (fs *FacetStructure) Translate(translation *vec3.T) {
	translatedPoints := make(map[*vec3.T]bool)
	fs.translate(translation, translatedPoints)
}

func (fs *FacetStructure) translate(translation *vec3.T, translatedPoints map[*vec3.T]bool) {
	for _, facet := range fs.Facets {
		facet.translate(translation, translatedPoints)
	}

	if len(fs.FacetStructures) > 0 {
		for _, facetStructure := range fs.FacetStructures {
			facetStructure.translate(translation, translatedPoints)
		}
	}
}

func (fs *FacetStructure) ScaleUniform(scaleOrigin *vec3.T, scale float64) {
	scale3d := &vec3.T{scale, scale, scale}
	fs.Scale(scaleOrigin, scale3d)
}

func (fs *FacetStructure) Scale(scaleOrigin *vec3.T, scale *vec3.T) {
	scaledPoints := make(map[*vec3.T]bool)
	fs.scale(scaleOrigin, scale, scaledPoints)
}

func (fs *FacetStructure) scale(scaleOrigin *vec3.T, scale *vec3.T, scaledPoints map[*vec3.T]bool) {
	for _, facet := range fs.Facets {
		facet.scale(scaleOrigin, scale, scaledPoints)
	}

	if len(fs.FacetStructures) > 0 {
		for _, facetStructure := range fs.FacetStructures {
			facetStructure.scale(scaleOrigin, scale, scaledPoints)
		}
	}
}

func (fs *FacetStructure) GetFirstObjectByName(objectName string) *FacetStructure {
	if fs.Name == objectName {
		return fs
	}

	if len(fs.FacetStructures) > 0 {
		for _, facetStructure := range fs.FacetStructures {
			object := facetStructure.GetFirstObjectByName(objectName)

			if object != nil {
				return object
			}
		}
	}

	return nil
}

func (fs *FacetStructure) GetFirstObjectBySubstructureName(objectName string) *FacetStructure {
	if fs.SubstructureName == objectName {
		return fs
	}

	if len(fs.FacetStructures) > 0 {
		for _, facetStructure := range fs.FacetStructures {
			object := facetStructure.GetFirstObjectBySubstructureName(objectName)

			if object != nil {
				return object
			}
		}
	}

	return nil
}

func (fs *FacetStructure) ClearMaterials() {
	fs.Material = nil

	if len(fs.FacetStructures) > 0 {
		for _, facetStructure := range fs.FacetStructures {
			facetStructure.ClearMaterials()
		}
	}
}
