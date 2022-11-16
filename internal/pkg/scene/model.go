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

// RenderType is the type used to define different render types
type RenderType string

const (
	// Pathtracing render type is used in camera settings to denote the path tracing algorithm to be used rendering the frame
	Pathtracing RenderType = "Pathtracing"
	// Raycasting render type is used in camera settings to denote the cheap and simple ray casting algorithm to be used rendering the frame
	Raycasting RenderType = "Raycasting"
)

type Ray struct {
	Origin          *vec3.T
	Heading         *vec3.T
	RefractionIndex float64
}

type Animation struct {
	AnimationName     string
	Frames            []Frame
	Width             int
	Height            int
	WriteRawImageFile bool
}

type SceneNode struct {
	Spheres         []*Sphere
	Discs           []*Disc
	ChildNodes      []*SceneNode
	FacetStructures []*FacetStructure
	Bounds          *Bounds `json:"-"`
}

type Material struct {
	Name            string           `json:"Name,omitempty"`
	Color           *color.Color     `json:"Color,omitempty"`
	Emission        *color.Color     `json:"Emission,omitempty"`
	Glossiness      float32          `json:"Glossiness,omitempty"` // Glossiness is the percent amount that will make out specular reflection. Values [0.0 .. 1.0] with default 0.0. Lower value the more diffuse color will appear and higher value the more mirror reflection will appear.
	Roughness       float32          `json:"Roughness,omitempty"`  // Roughness is the diffuse spread of the specular reflection. Values [0.0 .. 1.0] with default 0.0. Lower is like "brushed metal" or "foggy/hazy reflection" and higher value give a more mirror like reflection. A value of 0.0 is perfect mirror reflection and a value of 0.0 is a perfect diffuse material (no mirror at al).
	Projection      *ImageProjection `json:"Projection,omitempty"`
	RefractionIndex float64          `json:"RefractionIndex,omitempty"`
	Transparency    float64          `json:"Transparency,omitempty"`
	RayTerminator   bool             `json:"RayTerminator,omitempty"`
}

type ImageProjection struct {
	ProjectionType                  ProjectionType `json:"ProjectionType"`
	ImageFilename                   string         `json:"ImageFilename"`
	Origin                          *vec3.T        `json:"Origin"`
	U                               *vec3.T        `json:"U"`
	V                               *vec3.T        `json:"V"`
	RepeatU                         bool           `json:"RepeatU,omitempty"`
	RepeatV                         bool           `json:"RepeatV,omitempty"`
	FlipU                           bool           `json:"FlipU,omitempty"`
	FlipV                           bool           `json:"FlipV,omitempty"`
	Gamma                           float64        `json:"Gamma,omitempty"`
	_image                          *image.FloatImage
	_invertedCoordinateSystemMatrix *mat3.T
}

type Camera struct {
	Origin            *vec3.T
	Heading           *vec3.T
	ViewUp            *vec3.T
	ViewPlaneDistance float64
	_coordinateSystem *mat3.T
	ApertureSize      float64 // ApertureSize is the size of the aperture opening. The wider the aperture the less focus depth. Value 0.0 is infinite focus depth.
	ApertureShape     string  // ApertureShape file path to a black and white image where white define the aperture shape. Aperture size determine the size of the longest side (width or height) of the image. If nil then a default round aperture shape is used.
	FocusDistance     float64
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

type FacetStructure struct {
	Name             string            `json:"Name,omitempty"`
	SubstructureName string            `json:"SubstructureName,omitempty"`
	Material         *Material         `json:"Material,omitempty"`
	Facets           []*Facet          `json:"Facets,omitempty"`
	FacetStructures  []*FacetStructure `json:"FacetStructures,omitempty"`

	Bounds *Bounds `json:"-"` // Calculated attribute. See UpdateBounds(). Derived from all vertices in all sub facets recursively.
}

type Facet struct {
	Vertices        []*vec3.T `json:"Vertices"`
	TextureVertices []*vec3.T `json:"TextureVertices,omitempty"`
	VertexNormals   []*vec3.T `json:"VertexNormals,omitempty"`

	Normal *vec3.T `json:"-"` // Calculated attribute. See UpdateNormal(). Derived from the first three vertices of the triangle.
	Bounds *Bounds `json:"-"` // Calculated attribute. See UpdateBounds(). Derived from all vertices in the facet.
}

type Sphere struct {
	Name     string
	Origin   *vec3.T
	Radius   float64
	Material *Material `json:"material,omitempty"`
}

type Plane struct {
	Name     string
	Origin   *vec3.T
	Normal   *vec3.T
	Material *Material `json:"Material,omitempty"`
}

type Disc struct {
	Name     string
	Origin   *vec3.T
	Normal   *vec3.T
	Radius   float64
	Material *Material `json:"Material,omitempty"`
}

func RotateY(point *vec3.T, rotationOrigin *vec3.T, angle float64) {
	rotationMatrix := mat3.T{}
	rotationMatrix.AssignYRotation(angle)

	origin := *point
	origin.Sub(rotationOrigin)
	origin[2] *= -1 // Change to right hand coordinate system from left hand coordinate system
	rotatedOrigin := rotationMatrix.MulVec3(&origin)
	rotatedOrigin[2] *= -1 // Change back from right hand coordinate system to left hand coordinate system
	rotatedOrigin.Add(rotationOrigin)

	point[0] = rotatedOrigin[0]
	point[1] = rotatedOrigin[1]
	point[2] = rotatedOrigin[2]
}

func (sphere *Sphere) RotateY(rotationOrigin *vec3.T, angle float64) {
	rotationMatrix := mat3.T{}
	rotationMatrix.AssignYRotation(angle)

	sphere.rotate(rotationOrigin, rotationMatrix)
}

func (sphere *Sphere) rotate(rotationOrigin *vec3.T, rotationMatrix mat3.T) {
	origin := *sphere.Origin
	origin.Sub(rotationOrigin)
	origin[2] *= -1 // Change to right hand coordinate system from left hand coordinate system
	rotatedOrigin := rotationMatrix.MulVec3(&origin)
	rotatedOrigin[2] *= -1 // Change back from right hand coordinate system to left hand coordinate system
	rotatedOrigin.Add(rotationOrigin)

	sphere.Origin[0] = rotatedOrigin[0]
	sphere.Origin[1] = rotatedOrigin[1]
	sphere.Origin[2] = rotatedOrigin[2]
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

// UpdateMaterials propagates parent materials down in facet structure hierarchy to sub structures without explicit own material
func (fs *FacetStructure) UpdateMaterials() {
	for _, facetStructure := range fs.FacetStructures {
		if facetStructure.Material == nil {
			facetStructure.Material = fs.Material
		}

		facetStructure.UpdateMaterials()
	}
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
	fs.UpdateMaterials()
	fs.InitializeProjection()
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

func (fs *FacetStructure) ClearMaterials() {
	fs.Material = nil

	if len(fs.FacetStructures) > 0 {
		for _, facetStructure := range fs.FacetStructures {
			facetStructure.ClearMaterials()
		}
	}
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

func (sn *SceneNode) Initialize() {
	// Empty by intention
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

func (r Ray) point(t float64) *vec3.T {
	return &vec3.T{
		r.Origin[0] + r.Heading[0]*t,
		r.Origin[1] + r.Heading[1]*t,
		r.Origin[2] + r.Heading[2]*t,
	}
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

func (sphere *Sphere) Initialize() {
	projection := sphere.Material.Projection
	if projection != nil {
		projection.Initialize()
	}
}

func (sphere *Sphere) Translate(translation *vec3.T) {
	sphere.Origin.Add(translation)
}

func (sphere *Sphere) Bounds() *Bounds {
	return &Bounds{
		Xmin: sphere.Origin[0] - sphere.Radius,
		Xmax: sphere.Origin[0] + sphere.Radius,
		Ymin: sphere.Origin[1] - sphere.Radius,
		Ymax: sphere.Origin[1] + sphere.Radius,
		Zmin: sphere.Origin[2] - sphere.Radius,
		Zmax: sphere.Origin[2] + sphere.Radius,
	}
}

func (disc *Disc) Initialize() {
	disc.Normal.Normalize()

	projection := disc.Material.Projection
	if projection != nil {
		projection.Initialize()
	}
}

func (disc *Disc) Translate(translation *vec3.T) {
	disc.Origin.Add(translation)
}

func (disc *Disc) rotate(rotationOrigin *vec3.T, rotationMatrix mat3.T) {
	normal := *disc.Normal
	normal[2] *= -1 // Change to right hand coordinate system from left hand coordinate system
	rotatedNormal := rotationMatrix.MulVec3(&normal)
	rotatedNormal[2] *= -1 // Change back from right hand coordinate system to left hand coordinate system
	disc.Normal[0] = rotatedNormal[0]
	disc.Normal[1] = rotatedNormal[1]
	disc.Normal[2] = rotatedNormal[2]

	origin := *disc.Origin
	origin.Sub(rotationOrigin)
	origin[2] *= -1 // Change to right hand coordinate system from left hand coordinate system
	rotatedOrigin := rotationMatrix.MulVec3(&origin)
	rotatedOrigin[2] *= -1 // Change back from right hand coordinate system to left hand coordinate system
	rotatedOrigin.Add(rotationOrigin)

	disc.Origin[0] = rotatedOrigin[0]
	disc.Origin[1] = rotatedOrigin[1]
	disc.Origin[2] = rotatedOrigin[2]
}

func (disc *Disc) RotateY(rotationOrigin *vec3.T, angle float64) {
	// A matrix m of type mat3.T is addressed as: m[columnIndex][rowIndex]
	rotationMatrix := mat3.T{}
	rotationMatrix.AssignYRotation(angle)

	disc.rotate(rotationOrigin, rotationMatrix)
}

func (disc *Disc) RotateX(rotationOrigin *vec3.T, angle float64) {
	// A matrix m of type mat3.T is addressed as: m[columnIndex][rowIndex]
	rotationMatrix := mat3.T{}
	rotationMatrix.AssignXRotation(angle)

	disc.rotate(rotationOrigin, rotationMatrix)
}

func (disc *Disc) RotateZ(rotationOrigin *vec3.T, angle float64) {
	// A matrix m of type mat3.T is addressed as: m[columnIndex][rowIndex]
	rotationMatrix := mat3.T{}
	rotationMatrix.AssignZRotation(angle)

	disc.rotate(rotationOrigin, rotationMatrix)
}
