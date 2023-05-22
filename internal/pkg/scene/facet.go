package scene

import (
	"fmt"
	"github.com/ungerik/go3d/float64/mat3"
	"github.com/ungerik/go3d/float64/vec3"
)

type Facet struct {
	Vertices        []*vec3.T `json:"Vertices"`
	TextureVertices []*vec3.T `json:"TextureVertices,omitempty"`
	VertexNormals   []*vec3.T `json:"VertexNormals,omitempty"`

	Normal *vec3.T `json:"-"` // Calculated attribute. See UpdateNormal(). Derived from the first three vertices of the triangle.
	Bounds *Bounds `json:"-"` // Calculated attribute. See UpdateBounds(). Derived from all vertices in the facet.
}

// SplitMultiPointFacet maps a multipoint (> 3 points) facet into a list of triangles.
// The facet to split must have more than 3 points and be a convex face.
func (f *Facet) SplitMultiPointFacet() []*Facet {
	var facets []*Facet

	if f.IsMultiPointFacet() {
		// Add consecutive triangles of facet
		amountVertices := len(f.Vertices)
		for i := 1; i < (amountVertices - 1); i++ {
			newVertices := []*vec3.T{f.Vertices[0], f.Vertices[i], f.Vertices[i+1]}

			var newTextureVertices []*vec3.T
			if len(f.TextureVertices) > 0 {
				newTextureVertices = []*vec3.T{f.TextureVertices[0], f.TextureVertices[i], f.TextureVertices[i+1]}
			}

			var newVertexNormals []*vec3.T
			if len(f.VertexNormals) > 0 {
				newVertexNormals = []*vec3.T{f.VertexNormals[0], f.VertexNormals[i], f.VertexNormals[i+1]}
			}

			newFace := Facet{
				Vertices:        newVertices,
				TextureVertices: newTextureVertices,
				VertexNormals:   newVertexNormals,
				Normal:          f.Normal,
			}
			facets = append(facets, &newFace)
		}
	} else {
		facets = append(facets, f)
	}

	return facets
}

func (f *Facet) UpdateBounds() *Bounds {
	if f.Bounds == nil {
		bounds := NewBounds()
		for _, vertex := range f.Vertices {
			bounds.IncludeVertex(vertex)
		}

		f.Bounds = &bounds
	}

	return f.Bounds
}

func (f *Facet) UpdateNormal() {
	sideVector1 := vec3.Sub(f.Vertices[1], f.Vertices[0])
	sideVector2 := vec3.Sub(f.Vertices[2], f.Vertices[0])
	normal := vec3.Cross(&sideVector1, &sideVector2)
	normal.Normalize()

	if f.Normal == nil {
		f.Normal = &normal
	} else {
		f.Normal[0] = normal[0]
		f.Normal[1] = normal[1]
		f.Normal[2] = normal[2]
	}
}

func (f *Facet) Center() *vec3.T {
	center := vec3.T{0, 0, 0}

	for _, vertex := range f.Vertices {
		center[0] += vertex[0]
		center[1] += vertex[1]
		center[2] += vertex[2]
	}

	amountVertices := float64(len(f.Vertices))

	center[0] /= amountVertices
	center[1] /= amountVertices
	center[2] /= amountVertices

	return &center
}

func (f *Facet) RotateX(rotationOrigin *vec3.T, angle float64) {
	rotatedPoints := make(map[*vec3.T]bool)
	rotatedNormals := make(map[*vec3.T]bool)
	rotatedVertexNormals := make(map[*vec3.T]bool)

	rotationMatrix := mat3.T{}
	rotationMatrix.AssignXRotation(angle)

	f.rotate(rotationOrigin, rotationMatrix, rotatedPoints, rotatedNormals, rotatedVertexNormals)
}

func (f *Facet) RotateY(rotationOrigin *vec3.T, angle float64) {
	rotatedPoints := make(map[*vec3.T]bool)
	rotatedNormals := make(map[*vec3.T]bool)
	rotatedVertexNormals := make(map[*vec3.T]bool)

	rotationMatrix := mat3.T{}
	rotationMatrix.AssignYRotation(angle)

	f.rotate(rotationOrigin, rotationMatrix, rotatedPoints, rotatedNormals, rotatedVertexNormals)
}

func (f *Facet) RotateZ(rotationOrigin *vec3.T, angle float64) {
	rotatedPoints := make(map[*vec3.T]bool)
	rotatedNormals := make(map[*vec3.T]bool)
	rotatedVertexNormals := make(map[*vec3.T]bool)

	rotationMatrix := mat3.T{}
	rotationMatrix.AssignZRotation(angle)

	f.rotate(rotationOrigin, rotationMatrix, rotatedPoints, rotatedNormals, rotatedVertexNormals)
}

func (f *Facet) rotate(rotationOrigin *vec3.T, rotationMatrix mat3.T, rotatedPoints map[*vec3.T]bool, rotatedNormals map[*vec3.T]bool, rotatedVertexNormals map[*vec3.T]bool) {
	for _, vertex := range f.Vertices {
		if rotatedPoints[vertex] {
			// fmt.Printf("Point already rotated: %+v\n", vertex)
		} else {
			newVertex := vertex.Subed(rotationOrigin)
			newVertex[2] *= -1 // Convert to right hand coordinate system before rotation matrix
			rotatedVertex := rotationMatrix.MulVec3(&newVertex)
			rotatedVertex[2] *= -1 // Convert back to left hand coordinate system after rotation matrix
			rotatedVertex.Add(rotationOrigin)

			vertex[0] = rotatedVertex[0]
			vertex[1] = rotatedVertex[1]
			vertex[2] = rotatedVertex[2]

			rotatedPoints[vertex] = true
		}
	}

	if rotatedNormals[f.Normal] {
		// fmt.Printf("Normal already rotated: %+v\n", f.Normal)
	} else {
		normal := *f.Normal
		normal[2] *= -1 // Convert to right hand coordinate system before rotation matrix
		rotatedNormal := rotationMatrix.MulVec3(&normal)
		rotatedNormal[2] *= -1 // Convert back to left hand coordinate system after rotation matrix
		f.Normal[0] = rotatedNormal[0]
		f.Normal[1] = rotatedNormal[1]
		f.Normal[2] = rotatedNormal[2]

		rotatedNormals[f.Normal] = true
	}

	for _, vertexNormal := range f.VertexNormals {
		if rotatedVertexNormals[vertexNormal] {
			// fmt.Printf("Vertex normal already rotated: %+v\n", vertexNormal)
		} else {
			normal := *vertexNormal
			normal[2] *= -1 // Convert to right hand coordinate system before rotation matrix
			rotatedNormal := rotationMatrix.MulVec3(&normal)
			rotatedNormal[2] *= -1 // Convert back to left hand coordinate system after rotation matrix
			vertexNormal[0] = rotatedNormal[0]
			vertexNormal[1] = rotatedNormal[1]
			vertexNormal[2] = rotatedNormal[2]

			rotatedVertexNormals[vertexNormal] = true
		}
	}

	f.Bounds = nil
}

func (f *Facet) translate(translation *vec3.T, translatedPoints map[*vec3.T]bool) {
	for _, vertex := range f.Vertices {
		vertexAlreadyTranslated := translatedPoints[vertex]

		if !vertexAlreadyTranslated {
			newVertex := vertex.Added(translation)

			vertex[0] = newVertex[0]
			vertex[1] = newVertex[1]
			vertex[2] = newVertex[2]

			translatedPoints[vertex] = true
		}
	}

	f.Bounds = nil
}

func (f *Facet) scale(scaleOrigin *vec3.T, scale *vec3.T, scaledPoints map[*vec3.T]bool, scaledNormals map[*vec3.T]bool) {

	for _, vertex := range f.Vertices {
		if !scaledPoints[vertex] {
			newVertex := vertex.Subed(scaleOrigin)
			newVertex.Mul(scale)
			newVertex.Add(scaleOrigin)

			vertex[0] = newVertex[0]
			vertex[1] = newVertex[1]
			vertex[2] = newVertex[2]

			scaledPoints[vertex] = true
		}
	}

	if f.Normal != nil {
		f.UpdateNormal()
		scaledNormals[f.Normal] = true
	}

	// TODO Update vertex normals?! How?

	f.Bounds = nil
}

func (f *Facet) ChangeWindingOrder() {
	amountVertices := len(f.Vertices)
	if amountVertices == 3 {
		// Flip second and third vertex in triangle (facet) to change winding order of facet vertices
		f.Vertices[0], f.Vertices[1], f.Vertices[2] = f.Vertices[2], f.Vertices[1], f.Vertices[0]
	} else if amountVertices > 3 {
		for i := 0; i < amountVertices/2; i++ {
			f.Vertices[i], f.Vertices[amountVertices-i-1] = f.Vertices[amountVertices-i-1], f.Vertices[i]
		}
	}
}

// Tessellate will divide this triangle facet (facet of 3 vertices) into four triangles.
// Returns the subdividing four facets or nil with an error if the facet is not a triangle.
func (f *Facet) Tessellate() ([]*Facet, error) {
	if len(f.Vertices) != 3 {
		return nil, fmt.Errorf(fmt.Sprintf("facet is not a triangle but has %d vertices", len(f.Vertices)))
	}

	var newFacets []*Facet = nil

	p1 := f.Vertices[0]
	p2 := f.Vertices[1]
	p3 := f.Vertices[2]

	p12 := vec3.Interpolate(p1, p2, 0.5)
	p23 := vec3.Interpolate(p2, p3, 0.5)
	p31 := vec3.Interpolate(p3, p1, 0.5)

	newFacets = append(newFacets, &Facet{Vertices: []*vec3.T{p1, &p12, &p31}})
	newFacets = append(newFacets, &Facet{Vertices: []*vec3.T{&p12, p2, &p23}})
	newFacets = append(newFacets, &Facet{Vertices: []*vec3.T{&p12, &p23, &p31}})
	newFacets = append(newFacets, &Facet{Vertices: []*vec3.T{&p31, &p23, p3}})

	return newFacets, nil
}

func (f *Facet) IsMultiPointFacet() bool {
	return len(f.Vertices) > 3
}
