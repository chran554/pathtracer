package scene

import (
	"fmt"
	"github.com/ungerik/go3d/float64/mat3"
	"github.com/ungerik/go3d/float64/vec3"
	"math"
	"pathtracer/internal/pkg/util"
	"strings"
)

type FacetStructure struct {
	Name             string            `json:"Name,omitempty"`
	SubstructureName string            `json:"SubstructureName,omitempty"`
	Material         *Material         `json:"Material,omitempty"`
	Facets           []*Facet          `json:"Facets,omitempty"`
	FacetStructures  []*FacetStructure `json:"FacetStructures,omitempty"`

	IgnoreBounds bool    `json:"IgnoreBounds,omitempty"`
	Bounds       *Bounds `json:"-"` // Calculated attribute. See UpdateBounds(). Derived from all vertices in all sub facets recursively.
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

// Purge removes empty substructures.
func (fs *FacetStructure) Purge() {
	for i := 0; i < len(fs.FacetStructures); {
		if fs.FacetStructures[i].GetAmountFacets() == 0 {
			fs.FacetStructures[i] = fs.FacetStructures[len(fs.FacetStructures)-1]
			fs.FacetStructures = fs.FacetStructures[:len(fs.FacetStructures)-1]
		} else {
			i++
		}
	}
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

	rotatedImageProjections := make(map[*ImageProjection]bool)

	rotationMatrix := mat3.T{}
	rotationMatrix.AssignXRotation(angle)

	fs.rotate(rotationOrigin, rotationMatrix, rotatedPoints, rotatedNormals, rotatedVertexNormals, rotatedImageProjections)
	fs.UpdateBounds()
}

func (fs *FacetStructure) RotateY(rotationOrigin *vec3.T, angle float64) {
	rotatedPoints := make(map[*vec3.T]bool)
	rotatedNormals := make(map[*vec3.T]bool)
	rotatedVertexNormals := make(map[*vec3.T]bool)

	rotatedImageProjections := make(map[*ImageProjection]bool)

	rotationMatrix := mat3.T{}
	rotationMatrix.AssignYRotation(angle)

	fs.rotate(rotationOrigin, rotationMatrix, rotatedPoints, rotatedNormals, rotatedVertexNormals, rotatedImageProjections)
	fs.UpdateBounds()
}

func (fs *FacetStructure) RotateZ(rotationOrigin *vec3.T, angle float64) {
	rotatedPoints := make(map[*vec3.T]bool)
	rotatedNormals := make(map[*vec3.T]bool)
	rotatedVertexNormals := make(map[*vec3.T]bool)

	rotatedImageProjections := make(map[*ImageProjection]bool)

	rotationMatrix := mat3.T{}
	rotationMatrix.AssignZRotation(angle)

	fs.rotate(rotationOrigin, rotationMatrix, rotatedPoints, rotatedNormals, rotatedVertexNormals, rotatedImageProjections)
	fs.UpdateBounds()
}

func (fs *FacetStructure) rotate(rotationOrigin *vec3.T, rotationMatrix mat3.T, rotatedPoints map[*vec3.T]bool, rotatedNormals map[*vec3.T]bool, rotatedVertexNormals map[*vec3.T]bool, rotatedImageProjections map[*ImageProjection]bool) {
	if fs.Material != nil && fs.Material.Projection != nil && !rotatedImageProjections[fs.Material.Projection] {
		projection := fs.Material.Projection

		projection.Origin.Sub(rotationOrigin)
		projection.Origin[2] *= -1
		rotatedVertex := rotationMatrix.MulVec3(projection.Origin)
		rotatedVertex[2] *= -1
		rotatedVertex.Add(rotationOrigin)

		projection.Origin[0] = rotatedVertex[0]
		projection.Origin[1] = rotatedVertex[1]
		projection.Origin[2] = rotatedVertex[2]

		projection.U[2] *= -1
		rotatedU := rotationMatrix.MulVec3(projection.U)
		rotatedU[2] *= -1

		projection.U[0] = rotatedU[0]
		projection.U[1] = rotatedU[1]
		projection.U[2] = rotatedU[2]

		projection.V[2] *= -1
		rotatedV := rotationMatrix.MulVec3(projection.V)
		rotatedV[2] *= -1

		projection.V[0] = rotatedV[0]
		projection.V[1] = rotatedV[1]
		projection.V[2] = rotatedV[2]

		rotatedImageProjections[fs.Material.Projection] = true
	}

	for _, facet := range fs.Facets {
		facet.rotate(rotationOrigin, rotationMatrix, rotatedPoints, rotatedNormals, rotatedVertexNormals)
	}

	if len(fs.FacetStructures) > 0 {
		for _, facetStructure := range fs.FacetStructures {
			facetStructure.rotate(rotationOrigin, rotationMatrix, rotatedPoints, rotatedNormals, rotatedVertexNormals, rotatedImageProjections)
		}
	}
}

func (fs *FacetStructure) Translate(translation *vec3.T) {
	translatedPoints := make(map[*vec3.T]bool)
	translatedImageProjections := make(map[*ImageProjection]bool)

	fs.translate(translation, translatedPoints, translatedImageProjections)
	fs.UpdateBounds()
}

func (fs *FacetStructure) translate(translation *vec3.T, translatedPoints map[*vec3.T]bool, translatedImageProjections map[*ImageProjection]bool) {
	// Translate image projection (i.e. image projection origin)
	if fs.Material != nil && fs.Material.Projection != nil {
		origin := fs.Material.Projection.Origin
		projectionOriginAlreadyTranslated := translatedPoints[origin]
		projectionAlreadyTranslated := translatedImageProjections[fs.Material.Projection]

		if !projectionOriginAlreadyTranslated && !projectionAlreadyTranslated {
			origin.Add(translation)
			translatedPoints[origin] = true
			translatedImageProjections[fs.Material.Projection] = true
		}
	}

	for _, facet := range fs.Facets {
		facet.translate(translation, translatedPoints)
	}

	if len(fs.FacetStructures) > 0 {
		for _, facetStructure := range fs.FacetStructures {
			facetStructure.translate(translation, translatedPoints, translatedImageProjections)
		}
	}
}

func (fs *FacetStructure) ScaleUniform(scaleOrigin *vec3.T, scale float64) {
	scale3d := &vec3.T{scale, scale, scale}
	fs.Scale(scaleOrigin, scale3d)
}

func (fs *FacetStructure) Scale(scaleOrigin *vec3.T, scale *vec3.T) {
	scaledPoints := make(map[*vec3.T]bool)
	scaledImageProjections := make(map[*ImageProjection]bool)
	fs.scale(scaleOrigin, scale, scaledPoints, scaledImageProjections)
	fs.UpdateBounds()
}

func (fs *FacetStructure) scale(scaleOrigin *vec3.T, scale *vec3.T, scaledPoints map[*vec3.T]bool, scaledImageProjections map[*ImageProjection]bool) {
	if fs.Material != nil && fs.Material.Projection != nil && !scaledImageProjections[fs.Material.Projection] {
		projection := fs.Material.Projection
		projection.Origin.Sub(scaleOrigin).Mul(scale).Add(scaleOrigin)
		projection.U.Mul(scale)
		projection.V.Mul(scale)

		scaledImageProjections[projection] = true
	}

	for _, facet := range fs.Facets {
		facet.scale(scaleOrigin, scale, scaledPoints)
	}

	if len(fs.FacetStructures) > 0 {
		for _, facetStructure := range fs.FacetStructures {
			facetStructure.scale(scaleOrigin, scale, scaledPoints, scaledImageProjections)
		}
	}
}

func (fs *FacetStructure) TwistY(origin *vec3.T, anglePerYUnit float64) {
	structureVertices := fs.getVertices()

	var verticesSet = make(map[*vec3.T]bool, 0)
	for _, vertex := range structureVertices {
		verticesSet[vertex] = true
	}

	for vertex := range verticesSet {
		x := vertex[0] - origin[0]
		y := vertex[1] - origin[1]
		z := vertex[2] - origin[2]

		r := math.Sqrt(x*x + z*z) // radius in xz-plane

		if r != 0.0 {
			angle := anglePerYUnit * y

			sa := z / r
			ca := x / r
			sb := math.Sin(angle)
			cb := math.Cos(angle)

			// sin(𝛼±𝛽)=sin𝛼cos𝛽±cos𝛼sin𝛽
			// cos(𝛼±𝛽)=cos𝛼cos𝛽∓sin𝛼sin𝛽
			x2 := r * (ca*cb - sa*sb)
			z2 := r * (sa*cb + ca*sb)

			vertex[0] = x2 + origin[0]
			vertex[2] = z2 + origin[2]
		}
	}

	fs.UpdateBounds()
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

func (fs *FacetStructure) GetObjectsByName(objectName string) []*FacetStructure {
	nameMatchFunction := func(structure *FacetStructure, name string) bool {
		return structure.Name == name
	}

	return getObjectsByName(fs, objectName, nameMatchFunction)
}

func (fs *FacetStructure) GetObjectsBySubstructureName(objectName string) []*FacetStructure {
	nameMatchFunction := func(structure *FacetStructure, name string) bool {
		return structure.SubstructureName == name
	}

	return getObjectsByName(fs, objectName, nameMatchFunction)
}

func (fs *FacetStructure) GetObjectsByMaterialName(materialName string) []*FacetStructure {
	nameMatchFunction := func(structure *FacetStructure, name string) bool {
		return (structure.Material != nil) && (structure.Material.Name == name)
	}

	return getObjectsByName(fs, materialName, nameMatchFunction)
}

func (fs *FacetStructure) ReplaceMaterial(materialName string, material *Material) {
	objects := fs.GetObjectsByMaterialName(materialName)
	for _, object := range objects {
		object.Material = material
	}
}

func (fs *FacetStructure) GetFirstMaterialByName(materialName string) *Material {
	objects := fs.GetObjectsByMaterialName(materialName)

	if len(objects) > 0 {
		return objects[0].Material
	} else {
		return nil
	}
}

func getObjectsByName(fs *FacetStructure, name string, matchFunction func(*FacetStructure, string) bool) []*FacetStructure {
	var objects []*FacetStructure

	if matchFunction(fs, name) {
		objects = append(objects, fs)
	}

	if len(fs.FacetStructures) > 0 {
		for _, facetStructure := range fs.FacetStructures {
			subObjects := getObjectsByName(facetStructure, name, matchFunction)

			if len(subObjects) > 0 {
				objects = append(objects, subObjects...)
			}
		}
	}

	return objects
}

func removeObjectsByName(fs *FacetStructure, name string, matchFunction func(*FacetStructure, string) bool) {
	if len(fs.FacetStructures) > 0 {
		for facetStructureIndex := 0; facetStructureIndex < len(fs.FacetStructures); {
			facetStructure := fs.FacetStructures[facetStructureIndex]

			if matchFunction(facetStructure, name) {
				fs.FacetStructures = append(fs.FacetStructures[:facetStructureIndex], fs.FacetStructures[facetStructureIndex+1:]...)
			} else {
				facetStructureIndex++
			}
		}
	}
}

func (fs *FacetStructure) RemoveObjectsByName(objectName string) {
	nameMatchFunction := func(structure *FacetStructure, name string) bool {
		return structure.Name == name
	}

	removeObjectsByName(fs, objectName, nameMatchFunction)
	fs.UpdateBounds()
}

func (fs *FacetStructure) RemoveObjectsBySubstructureName(objectName string) {
	nameMatchFunction := func(structure *FacetStructure, name string) bool {
		return structure.SubstructureName == name
	}

	removeObjectsByName(fs, objectName, nameMatchFunction)
	fs.UpdateBounds()
}

func (fs *FacetStructure) RemoveObjectsByMaterialName(materialName string) {
	nameMatchFunction := func(structure *FacetStructure, name string) bool {
		return (structure.Material != nil) && (structure.Material.Name == name)
	}

	removeObjectsByName(fs, materialName, nameMatchFunction)
	fs.UpdateBounds()
}

func (fs *FacetStructure) ClearMaterials() {
	fs.Material = nil

	if len(fs.FacetStructures) > 0 {
		for _, facetStructure := range fs.FacetStructures {
			facetStructure.ClearMaterials()
		}
	}
}

func (fs *FacetStructure) SubdivideFacetStructure(maxFacets int, level int) {
	if (fs.GetAmountFacets() > maxFacets) && (maxFacets > 2) {
		// fmt.Printf("Subdividing %s with %d facets.\n", facetStructure.Name, facetStructure.GetAmountFacets())

		if len(fs.Facets) > maxFacets {

			// Calculate dividing center of facets
			var facetStructureCenter *vec3.T
			if len(fs.Facets) > 0 {
				center := &vec3.T{}
				for _, facet := range fs.Facets {
					center.Add(facet.Center())
				}
				center.Scale(1.0 / float64(len(fs.Facets)))
				facetStructureCenter = center
			} else {
				fs.UpdateBounds()
				bounds := fs.Bounds
				facetStructureCenter = bounds.Center()
			}

			subFacetStructures := make([]*FacetStructure, 8)

			for _, facet := range fs.Facets {
				facetSubstructureIndex := 0

				facetCenter := facet.Center()
				facetRelativeStructurePosition := vec3.Sub(facetCenter, facetStructureCenter)

				if facetRelativeStructurePosition[0] >= 0 {
					facetSubstructureIndex = facetSubstructureIndex | 0b001
				}
				if facetRelativeStructurePosition[1] >= 0 {
					facetSubstructureIndex = facetSubstructureIndex | 0b010
				}
				if facetRelativeStructurePosition[2] >= 0 {
					facetSubstructureIndex = facetSubstructureIndex | 0b100
				}

				if subFacetStructures[facetSubstructureIndex] == nil {
					subFacetStructures[facetSubstructureIndex] = &FacetStructure{
						Name:     fmt.Sprintf("%s-%03b", fs.Name, facetSubstructureIndex),
						Material: fs.Material,
					}
				}

				subFacetStructures[facetSubstructureIndex].Facets = append(subFacetStructures[facetSubstructureIndex].Facets, facet)
			}

			// logSubdivision(subFacetStructures, level, fs)

			amountSubstructures := 0
			for _, subFacetStructure := range subFacetStructures {
				if subFacetStructure != nil {
					amountSubstructures++
				}
			}
			if amountSubstructures > 1 {
				// Update the content of the current facet structure
				fs.Facets = nil
				for _, subFacetStructure := range subFacetStructures {
					if subFacetStructure != nil {
						fs.FacetStructures = append(fs.FacetStructures, subFacetStructure)
					}
				}
			}
		}

		for _, facetStructure := range fs.FacetStructures {
			facetStructure.SubdivideFacetStructure(maxFacets, level+1)
		}
	}
}

func logSubdivision(subFacetStructures []*FacetStructure, level int, fs *FacetStructure) {
	// TODO Remove
	builder := strings.Builder{}
	for _, subFacetStructure := range subFacetStructures {
		if subFacetStructure == nil {
			continue
		}

		if builder.Len() > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(fmt.Sprintf("%d", len(subFacetStructure.Facets)))
	}
	fmt.Printf("level: %d, structures:%d, facets:%d --> %s\n", level, len(fs.FacetStructures), len(fs.Facets), builder.String())
}

// UpdateVertexNormals updates vertex normals with an average normal of all facets sharing that vertex.
func (fs *FacetStructure) UpdateVertexNormals(keepExistingVertexNormals bool) {
	fs.UpdateVertexNormalsWithThreshold(keepExistingVertexNormals, 180)
}

// UpdateVertexNormalsWithThreshold updates vertex normals for a facet with an average normal of all facets sharing that vertex
// and is facing the same direction within a threshold angle (in degrees [0..180], not radians).
// A threshold of 180 degrees will include all facets sharing the same vertex.
// A value of 0 degrees will only include all facets that share the same vertex and have the same normal.
// https://iquilezles.org/articles/normals/
// https://computergraphics.stackexchange.com/questions/4031/programmatically-generating-vertex-normals
func (fs *FacetStructure) UpdateVertexNormalsWithThreshold(keepExistingVertexNormals bool, facetAngleThreshold float64) {
	fs.UpdateNormals()

	angleCosineThreshold := math.Cos((math.Pi / 180) * facetAngleThreshold)

	vertexToFacetMap := fs.getVertexToFacetMap()
	facets := fs.getFacets()

	for _, facet := range facets {
		facetHasSharedVertex := false
		for _, vertex := range facet.Vertices {
			facetHasSharedVertex = facetHasSharedVertex || (len(vertexToFacetMap[vertex]) > 1)
		}

		if facetHasSharedVertex {
			originalAmountVertexNormals := len(facet.VertexNormals)
			createVertexNormals := originalAmountVertexNormals == 0
			if createVertexNormals {
				facet.VertexNormals = make([]*vec3.T, len(facet.Vertices))
			}

			if createVertexNormals || !keepExistingVertexNormals {
				for vertexIndex, vertex := range facet.Vertices {
					vertexFacets := vertexToFacetMap[vertex]

					var includedVertexFacets []*Facet
					for _, vertexFacet := range vertexFacets {
						// TODO calculate which facets to include for current facet (vertex)
						cosine := util.Cosine(facet.Normal, vertexFacet.Normal)
						if cosine >= angleCosineThreshold {
							includedVertexFacets = append(includedVertexFacets, vertexFacet)
						}
					}

					vertexNormal := calculateAverageNormal(includedVertexFacets)
					facet.VertexNormals[vertexIndex] = vertexNormal
				}
			}
		}
	}

}

func (fs *FacetStructure) RemoveVertexNormals() {
	facets := fs.getFacets()

	for _, facet := range facets {
		facet.VertexNormals = nil
	}
}

func calculateAverageNormal(facets []*Facet) *vec3.T {
	averageNormal := vec3.T{0.0, 0.0, 0.0}
	for _, facet := range facets {
		averageNormal.Add(facet.Normal) // Naive, non-weighted, average normal of all facets
	}
	averageNormal.Normalize()

	return &averageNormal
}

func (fs *FacetStructure) getVertexToFacetMap() map[*vec3.T][]*Facet {
	vertexToFacetMap := make(map[*vec3.T][]*Facet, 0)

	if len(fs.FacetStructures) > 0 {
		for _, facetStructure := range fs.FacetStructures {
			childVerticesToFacetMap := facetStructure.getVertexToFacetMap()

			for vertex, vertexFacets := range childVerticesToFacetMap {
				vertexToFacetMap[vertex] = append(vertexToFacetMap[vertex], vertexFacets...)
			}
		}
	}

	for _, facet := range fs.Facets {
		for _, facetVertex := range facet.Vertices {
			vertexToFacetMap[facetVertex] = append(vertexToFacetMap[facetVertex], facet)
		}
	}

	return vertexToFacetMap
}

func (fs *FacetStructure) getFacets() []*Facet {
	var facets []*Facet = nil

	for _, facetStructure := range fs.FacetStructures {
		facets = append(facets, facetStructure.getFacets()...)
	}

	facets = append(facets, fs.Facets...)

	return facets
}

func (fs *FacetStructure) CenterOn(newCenter *vec3.T) {
	boundsCenter := fs.Bounds.Center()
	boundsCenter.Invert()
	boundsCenter.Add(newCenter)

	fs.Translate(boundsCenter)
}

func (fs *FacetStructure) ChangeWindingOrder() {
	for _, facet := range fs.Facets {
		facet.ChangeWindingOrder()
	}

	if len(fs.FacetStructures) > 0 {
		for _, facetStructure := range fs.FacetStructures {
			facetStructure.ChangeWindingOrder()
		}
	}
}

// Tessellate subdivides any contained triangle (facet with three vertices) into four new triangle facets.
func (fs *FacetStructure) Tessellate() {
	if len(fs.Facets) > 0 {
		var newFacets []*Facet = nil
		for _, facet := range fs.Facets {
			tessellateFacets, err := facet.Tessellate()
			if err == nil {
				newFacets = append(newFacets, tessellateFacets...)
			} else {
				newFacets = append(newFacets, facet)
			}
		}

		fs.Facets = newFacets
	}

	if len(fs.FacetStructures) > 0 {
		for _, facetStructure := range fs.FacetStructures {
			facetStructure.Tessellate()
		}
	}
}

func (fs *FacetStructure) getVertices() []*vec3.T {
	var vertices []*vec3.T

	for _, facet := range fs.Facets {
		vertices = append(vertices, facet.Vertices...)
	}

	for _, facetStructure := range fs.FacetStructures {
		vertices = append(vertices, facetStructure.getVertices()...)
	}

	return vertices
}

//connectedFacetGroups groups a bunch of unordered facets into set of facets that are connected.
/*func connectedFacetGroups(facets []*Facet) [][]*Facet {
	var connectedFacets map[*Facet][]*Facet // connectedFacets is a mapping from each facet to a group of directly connected facets

	// find directly connected facets to each facet
	for _, keyFacet := range facets {
		connectedFacets[keyFacet] = make([]*Facet, 0)
		for _, testFacet := range facets {
			if (testFacet != keyFacet) && areTrianglesEdgeConnected(testFacet, keyFacet) {
				connectedFacets[keyFacet] = append(connectedFacets[keyFacet], testFacet)
			}
		}
	}

	return orderedFacets
}
*/
// areTrianglesEdgeConnected return if two facets are "connected" by a common edge (side).
func areTrianglesEdgeConnected(facet1, facet2 *Facet) bool {
	// A facet is connected to another facet if they share a common edge.
	// They share a common edge if they share two vertices.
	// The vertices must, in this implementation, be equal by reference not only by value
	// (as equal by value is still considered two different vertices).

	var facet2VertexSet map[*vec3.T]bool
	for _, vertex := range facet2.Vertices {
		facet2VertexSet[vertex] = true
	}

	amountCommonVertices := 0
	for _, vertex := range facet1.Vertices {
		if facet2VertexSet[vertex] {
			amountCommonVertices++
		}
	}

	return amountCommonVertices >= 2
}
