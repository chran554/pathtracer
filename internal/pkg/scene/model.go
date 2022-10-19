package scene

import (
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/image"

	"github.com/ungerik/go3d/float64/mat3"
	"github.com/ungerik/go3d/float64/vec3"
)

type ProjectionType string

const (
	Parallel    ProjectionType = "Parallel"
	Cylindrical ProjectionType = "Cylindrical"
	Spherical   ProjectionType = "Spherical"
)

type RenderType string

const (
	Pathtracing RenderType = "Pathtracing"
	Raycasting  RenderType = "Raycasting"
)

type Animation struct {
	AnimationName     string
	Frames            []Frame
	Width             int
	Height            int
	WriteRawImageFile bool
}

type FacetStructure struct {
	Name            string            `json:"Name,omitempty"`
	Material        *Material         `json:"Material,omitempty"`
	Facets          []*Facet          `json:"Facets,omitempty"`
	FacetStructures []*FacetStructure `json:"FacetStructures,omitempty"`

	Bounds *Bounds `json:"-"` // Calculated attribute. See UpdateBounds(). Derived from all vertices in all sub facets recursively.
}

type Facet struct {
	Vertices        []*vec3.T `json:"Vertices"`
	TextureVertices []*vec3.T `json:"TextureVertices,omitempty"`
	VertexNormals   []*vec3.T `json:"VertexNormals,omitempty"`

	Normal *vec3.T `json:"-"` // Calculated attribute. See UpdateNormal(). Derived from the first three vertices of the triangle.
	Bounds *Bounds `json:"-"` // Calculated attribute. See UpdateBounds(). Derived from all vertices in the facet.
}

type Camera struct {
	Origin            *vec3.T
	Heading           *vec3.T
	ViewUp            *vec3.T
	ViewPlaneDistance float64
	_coordinateSystem *mat3.T
	LensRadius        float64
	FocalDistance     float64
	Samples           int
	AntiAlias         bool
	Magnification     float64
	RenderType        RenderType
	RecursionDepth    int
}

type Frame struct {
	Filename   string
	FrameIndex int
	Camera     *Camera
	SceneNode  *SceneNode
}

type SceneNode struct {
	Spheres         []*Sphere
	Discs           []*Disc
	ChildNodes      []*SceneNode
	FacetStructures []*FacetStructure
	Bounds          *Bounds `json:"-"`
}

type Sphere struct {
	Name     string
	Origin   vec3.T
	Radius   float64
	Material *Material `json:"material,omitempty"`
}

func (fs *FacetStructure) GetAmountFacets() int {
	amount := len(fs.Facets)

	for _, facetStructure := range fs.FacetStructures {
		amount += facetStructure.GetAmountFacets()
	}

	return amount
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

func (fs *FacetStructure) UpdateNormals() {
	for _, facet := range fs.Facets {
		facet.UpdateNormal()
	}

	for _, facetStructure := range fs.FacetStructures {
		facetStructure.UpdateNormals()
	}
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

func (fs *FacetStructure) Initialize() {
	fs.UpdateNormals()
	fs.UpdateBounds()
	fs.InitializeProjection()
}

func (fs *FacetStructure) RotateX(rotationOrigin *vec3.T, angle float64) {
	for _, facet := range fs.Facets {
		facet.RotateX(rotationOrigin, angle)
	}

	if len(fs.FacetStructures) > 0 {
		for _, facetStructure := range fs.FacetStructures {
			facetStructure.RotateX(rotationOrigin, angle)
		}
	}
}

func (fs *FacetStructure) RotateY(rotationOrigin *vec3.T, angle float64) {
	rotatedPoints := make(map[*vec3.T]bool)
	rotatedNormals := make(map[*vec3.T]bool)
	rotatedVertexNormals := make(map[*vec3.T]bool)

	rotationMatrix := mat3.T{}
	rotationMatrix.AssignYRotation(angle)

	fs.rotateY(rotationOrigin, rotationMatrix, rotatedPoints, rotatedNormals, rotatedVertexNormals)
}

func (fs *FacetStructure) rotateY(rotationOrigin *vec3.T, rotationMatrix mat3.T, rotatedPoints map[*vec3.T]bool, rotatedNormals map[*vec3.T]bool, rotatedVertexNormals map[*vec3.T]bool) {
	for _, facet := range fs.Facets {
		facet.rotateY(rotationOrigin, rotationMatrix, rotatedPoints, rotatedNormals, rotatedVertexNormals)
	}

	if len(fs.FacetStructures) > 0 {
		for _, facetStructure := range fs.FacetStructures {
			facetStructure.rotateY(rotationOrigin, rotationMatrix, rotatedPoints, rotatedNormals, rotatedVertexNormals)
		}
	}
}

func (fs *FacetStructure) RotateZ(rotationOrigin *vec3.T, angle float64) {
	for _, facet := range fs.Facets {
		facet.RotateZ(rotationOrigin, angle)
	}

	if len(fs.FacetStructures) > 0 {
		for _, facetStructure := range fs.FacetStructures {
			facetStructure.RotateZ(rotationOrigin, angle)
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

// SplitMultiPointFacet maps a multipoint (> 3 points) face into a list of triangles.
// The supplied face must have at least 3 points and be a convex face.
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
	bounds := NewBounds()
	for _, vertex := range f.Vertices {
		bounds.IncludeVertex(vertex)
	}

	f.Bounds = &bounds
	return &bounds
}

func (f *Facet) UpdateNormal() {
	if f.Normal == nil {
		sideVector1 := vec3.Sub(f.Vertices[1], f.Vertices[0])
		sideVector2 := vec3.Sub(f.Vertices[2], f.Vertices[0])
		normal := vec3.Cross(&sideVector1, &sideVector2)
		normal.Normalize()
		f.Normal = &normal
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
	rotationMatrix := mat3.T{}
	rotationMatrix.AssignXRotation(angle)

	for i := range f.Vertices {
		newVertex := f.Vertices[i].Subed(rotationOrigin)
		rotatedVertex := rotationMatrix.MulVec3(&newVertex)
		rotatedVertex.Add(rotationOrigin)

		f.Vertices[i] = &rotatedVertex
	}

	for i := range f.VertexNormals {
		vertexNormal := f.VertexNormals[i]
		rotatedVertex := rotationMatrix.MulVec3(vertexNormal)
		rotatedVertex.Add(rotationOrigin)

		f.Vertices[i] = &rotatedVertex
	}
}

func (f *Facet) RotateY(rotationOrigin *vec3.T, angle float64) {
	rotatedPoints := make(map[*vec3.T]bool)
	rotatedNormals := make(map[*vec3.T]bool)
	rotatedVertexNormals := make(map[*vec3.T]bool)

	rotationMatrix := mat3.T{}
	rotationMatrix.AssignYRotation(angle)

	f.rotateY(rotationOrigin, rotationMatrix, rotatedPoints, rotatedNormals, rotatedVertexNormals)
}

func (f *Facet) rotateY(rotationOrigin *vec3.T, rotationMatrix mat3.T, rotatedPoints map[*vec3.T]bool, rotatedNormals map[*vec3.T]bool, rotatedVertexNormals map[*vec3.T]bool) {
	for _, vertex := range f.Vertices {
		if rotatedPoints[vertex] {
			// fmt.Printf("Point already rotated: %+v\n", vertex)
		} else {
			newVertex := vertex.Subed(rotationOrigin)
			rotatedVertex := rotationMatrix.MulVec3(&newVertex)
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
		rotatedNormal := rotationMatrix.MulVec3(f.Normal)
		f.Normal[0] = rotatedNormal[0]
		f.Normal[1] = rotatedNormal[1]
		f.Normal[2] = rotatedNormal[2]

		rotatedNormals[f.Normal] = true
	}

	for _, vertexNormal := range f.VertexNormals {
		if rotatedVertexNormals[vertexNormal] {
			// fmt.Printf("Vertex normal already rotated: %+v\n", vertexNormal)
		} else {
			rotatedNormal := rotationMatrix.MulVec3(vertexNormal)
			vertexNormal[0] = rotatedNormal[0]
			vertexNormal[1] = rotatedNormal[1]
			vertexNormal[2] = rotatedNormal[2]

			rotatedVertexNormals[vertexNormal] = true
		}
	}
}

func (f *Facet) RotateZ(rotationOrigin *vec3.T, angle float64) {
	rotationMatrix := mat3.T{}
	rotationMatrix.AssignZRotation(angle)

	for i := range f.Vertices {
		newVertex := f.Vertices[i].Subed(rotationOrigin)
		rotatedVertex := rotationMatrix.MulVec3(&newVertex)
		rotatedVertex.Add(rotationOrigin)

		f.Vertices[i] = &rotatedVertex
	}

	for i := range f.VertexNormals {
		vertexNormal := f.VertexNormals[i]
		rotatedVertex := rotationMatrix.MulVec3(vertexNormal)
		rotatedVertex.Add(rotationOrigin)

		f.Vertices[i] = &rotatedVertex
	}
}

func (f *Facet) translate(translation *vec3.T, translatedPoints map[*vec3.T]bool) {
	for _, vertex := range f.Vertices {
		if translatedPoints[vertex] {
			// fmt.Printf("Point already translated: %+v\n", vertex)
		} else {
			newVertex := vertex.Added(translation)

			vertex[0] = newVertex[0]
			vertex[1] = newVertex[1]
			vertex[2] = newVertex[2]

			translatedPoints[vertex] = true
		}
	}
}

func (f *Facet) scale(scaleOrigin *vec3.T, scale *vec3.T, scaledPoints map[*vec3.T]bool) {
	for _, vertex := range f.Vertices {
		if scaledPoints[vertex] {
			// fmt.Printf("Point already scaled: %+v\n", vertex)
		} else {
			newVertex := vertex.Subed(scaleOrigin)
			newVertex.Mul(scale)
			newVertex.Add(scaleOrigin)

			vertex[0] = newVertex[0]
			vertex[1] = newVertex[1]
			vertex[2] = newVertex[2]

			scaledPoints[vertex] = true
		}
	}
}

func (f *Facet) IsMultiPointFacet() bool {
	return len(f.Vertices) > 3
}

func (sn *SceneNode) GetAmountFacets() int {
	amount := 0
	for _, facetStructure := range sn.GetFacetStructures() {
		amount += facetStructure.GetAmountFacets()
	}
	return amount
}

func (sn *SceneNode) GetSpheres() []*Sphere {
	return sn.Spheres
}

func (sn *SceneNode) GetAmountSpheres() int {
	amountSpheres := len(sn.Spheres)
	for _, node := range sn.GetChildNodes() {
		amountSpheres += node.GetAmountSpheres()
	}
	return amountSpheres
}

func (sn *SceneNode) GetDiscs() []*Disc {
	return sn.Discs
}

func (sn *SceneNode) GetAmountDiscs() int {
	amountDiscs := len(sn.Discs)
	for _, node := range sn.GetChildNodes() {
		amountDiscs += node.GetAmountDiscs()
	}
	return amountDiscs
}

func (sn *SceneNode) Clear() {
	sn.Spheres = nil
	sn.Discs = nil

	for _, node := range sn.GetChildNodes() {
		node.Clear()
	}
}

func (sn *SceneNode) GetChildNodes() []*SceneNode {
	return sn.ChildNodes
}

func (sn *SceneNode) HasChildNodes() bool {
	return len(sn.ChildNodes) > 0
}

func (sn *SceneNode) GetBounds() *Bounds {
	return sn.Bounds
}

func (sn *SceneNode) GetFacetStructures() []*FacetStructure {
	return sn.FacetStructures
}

type Material struct {
	Color           color.Color      `json:"Color,omitempty"`
	Emission        *color.Color     `json:"Emission,omitempty"`
	Glossiness      float32          `json:"Glossiness,omitempty"` // Glossiness is the percent amount that will make out specular reflection. Values [0.0 .. 1.0] with default 0.0. Lower value the more diffuse color will appear and higher value the more mirror reflection will appear.
	Roughness       float32          `json:"Roughness,omitempty"`  // Roughness is the diffuse spread of the specular reflection. Values [0.0 .. 1.0] with default 0.0. Lower is like "brushed metal" or "foggy/hazy reflection" and higher value give a more mirror like reflection. A value of 0.0 is perfect mirror reflection and a value of 0.0 is a perfect diffuse material (no mirror at al).
	Projection      *ImageProjection `json:"Projection,omitempty"`
	RefractionIndex float64
	Transparency    float64
	RayTerminator   bool
}

type ImageProjection struct {
	ProjectionType                  ProjectionType `json:"ProjectionType"`
	ImageFilename                   string         `json:"ImageFilename"`
	_image                          *image.FloatImage
	_invertedCoordinateSystemMatrix mat3.T
	Origin                          vec3.T  `json:"Origin"`
	U                               vec3.T  `json:"U"`
	V                               vec3.T  `json:"V"`
	RepeatU                         bool    `json:"RepeatU,omitempty"`
	RepeatV                         bool    `json:"RepeatV,omitempty"`
	FlipU                           bool    `json:"FlipU,omitempty"`
	FlipV                           bool    `json:"FlipV,omitempty"`
	Gamma                           float64 `json:"Gamma,omitempty"`
}

type Ray struct {
	Origin          *vec3.T
	Heading         *vec3.T
	RefractionIndex float64
}

func (r Ray) point(t float64) *vec3.T {
	return &vec3.T{
		r.Origin[0] + r.Heading[0]*t,
		r.Origin[1] + r.Heading[1]*t,
		r.Origin[2] + r.Heading[2]*t,
	}
}

type Plane struct {
	Name     string
	Origin   *vec3.T
	Normal   *vec3.T
	Material *Material `json:"Material,omitempty"`
}

func NewPlane(v1, v2, v3 *vec3.T, name string, material *Material) *Plane {
	a := v2.Subed(v1)
	b := v3.Subed(v1)
	n := vec3.Cross(&a, &b)
	n.Normalize()

	return &Plane{
		Name:     name,
		Origin:   v1,
		Normal:   &n,
		Material: material,
	}
}

type Disc struct {
	Name     string
	Origin   *vec3.T
	Normal   *vec3.T
	Radius   float64
	Material *Material `json:"Material,omitempty"`
}

func (sphere Sphere) Initialize() {
	projection := sphere.Material.Projection
	if projection != nil {
		projection.Initialize()
	}
}

func (disc *Disc) Initialize() {
	disc.Normal.Normalize()

	projection := disc.Material.Projection
	if projection != nil {
		projection.Initialize()
	}
}
