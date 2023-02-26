package scene

import (
	"github.com/ungerik/go3d/float64/mat3"
	"github.com/ungerik/go3d/float64/vec3"
)

type SceneNode struct {
	Spheres         []*Sphere         `json:"Spheres,omitempty"`
	Discs           []*Disc           `json:"Discs,omitempty"`
	ChildNodes      []*SceneNode      `json:"ChildNodes,omitempty"`
	FacetStructures []*FacetStructure `json:"FacetStructures,omitempty"`
	Bounds          *Bounds           `json:"-"`
}

func NewSceneNode() *SceneNode {
	return &SceneNode{}
}

func (sn *SceneNode) S(spheres ...*Sphere) *SceneNode {
	sn.Spheres = append(sn.Spheres, spheres...)
	sn.UpdateBounds()
	return sn
}

func (sn *SceneNode) D(discs ...*Disc) *SceneNode {
	sn.Discs = append(sn.Discs, discs...)
	sn.UpdateBounds()
	return sn
}

func (sn *SceneNode) FS(facetStructures ...*FacetStructure) *SceneNode {
	sn.FacetStructures = append(sn.FacetStructures, facetStructures...)
	sn.UpdateBounds()
	return sn
}

func (sn *SceneNode) SN(sceneChildNodes ...*SceneNode) *SceneNode {
	sn.ChildNodes = append(sn.ChildNodes, sceneChildNodes...)
	sn.UpdateBounds()
	return sn
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

func (sn *SceneNode) UpdateBounds() *Bounds {
	bounds := NewBounds()
	sn.Bounds = &bounds

	for _, sphere := range sn.GetSpheres() {
		bounds.AddBounds(sphere.Bounds())
	}

	for _, disc := range sn.GetDiscs() {
		bounds.AddBounds(disc.Bounds())
	}

	for _, facetStructure := range sn.GetFacetStructures() {
		bounds.AddBounds(facetStructure.UpdateBounds())
	}

	for _, childNode := range sn.GetChildNodes() {
		bounds.AddBounds(childNode.UpdateBounds())
	}

	return sn.Bounds
}

func (sn *SceneNode) GetFacetStructures() []*FacetStructure {
	return sn.FacetStructures
}

func (sn *SceneNode) Scale(scaleOrigin *vec3.T, scale *vec3.T) {
	scaledPoints := make(map[*vec3.T]bool)
	scaledImageProjections := make(map[*ImageProjection]bool)

	sn.scale(scaleOrigin, scale, scaledPoints, scaledImageProjections)
}

func (sn *SceneNode) scale(scaleOrigin *vec3.T, scale *vec3.T, scaledPoints map[*vec3.T]bool, scaledImageProjections map[*ImageProjection]bool) {
	for _, sphere := range sn.GetSpheres() {
		sphere.scale(scaleOrigin, scale, scaledPoints, scaledImageProjections)
	}

	for _, disc := range sn.GetDiscs() {
		disc.scale(scaleOrigin, scale, scaledPoints, scaledImageProjections)
	}

	for _, facetStructure := range sn.GetFacetStructures() {
		facetStructure.scale(scaleOrigin, scale, scaledPoints, scaledImageProjections)
	}

	for _, childNode := range sn.GetChildNodes() {
		childNode.scale(scaleOrigin, scale, scaledPoints, scaledImageProjections)
	}
}

func (sn *SceneNode) Translate(translation *vec3.T) {
	translatedPoints := make(map[*vec3.T]bool)
	translatedImageProjections := make(map[*ImageProjection]bool)

	sn.translate(translation, translatedPoints, translatedImageProjections)
}

func (sn *SceneNode) translate(translation *vec3.T, translatedPoints map[*vec3.T]bool, translatedImageProjections map[*ImageProjection]bool) {
	for _, sphere := range sn.GetSpheres() {
		sphere.translate(translation, translatedPoints, translatedImageProjections)
	}

	for _, disc := range sn.GetDiscs() {
		disc.translate(translation, translatedPoints, translatedImageProjections)
	}

	for _, facetStructure := range sn.GetFacetStructures() {
		facetStructure.translate(translation, translatedPoints, translatedImageProjections)
	}

	for _, childNode := range sn.GetChildNodes() {
		childNode.translate(translation, translatedPoints, translatedImageProjections)
	}
}

func (sn *SceneNode) RotateX(rotationOrigin *vec3.T, angle float64) {
	rotatedPoints := make(map[*vec3.T]bool)
	rotatedNormals := make(map[*vec3.T]bool)
	rotatedVertexNormals := make(map[*vec3.T]bool)
	rotatedImageProjections := make(map[*ImageProjection]bool)

	rotationMatrix := mat3.T{}
	rotationMatrix.AssignXRotation(angle)

	sn.rotate(rotationOrigin, rotationMatrix, rotatedPoints, rotatedNormals, rotatedVertexNormals, rotatedImageProjections)
}

func (sn *SceneNode) RotateY(rotationOrigin *vec3.T, angle float64) {
	rotatedPoints := make(map[*vec3.T]bool)
	rotatedNormals := make(map[*vec3.T]bool)
	rotatedVertexNormals := make(map[*vec3.T]bool)
	rotatedImageProjections := make(map[*ImageProjection]bool)

	rotationMatrix := mat3.T{}
	rotationMatrix.AssignYRotation(angle)

	sn.rotate(rotationOrigin, rotationMatrix, rotatedPoints, rotatedNormals, rotatedVertexNormals, rotatedImageProjections)
}

func (sn *SceneNode) RotateZ(rotationOrigin *vec3.T, angle float64) {
	rotatedPoints := make(map[*vec3.T]bool)
	rotatedNormals := make(map[*vec3.T]bool)
	rotatedVertexNormals := make(map[*vec3.T]bool)
	rotatedImageProjections := make(map[*ImageProjection]bool)

	rotationMatrix := mat3.T{}
	rotationMatrix.AssignZRotation(angle)

	sn.rotate(rotationOrigin, rotationMatrix, rotatedPoints, rotatedNormals, rotatedVertexNormals, rotatedImageProjections)
}

func (sn *SceneNode) rotate(rotationOrigin *vec3.T, rotationMatrix mat3.T, rotatedPoints map[*vec3.T]bool, rotatedNormals map[*vec3.T]bool, rotatedVertexNormals map[*vec3.T]bool, rotatedImageProjections map[*ImageProjection]bool) {
	for _, sphere := range sn.GetSpheres() {
		sphere.rotate(rotationOrigin, rotationMatrix, rotatedPoints, rotatedImageProjections)
	}

	for _, disc := range sn.GetDiscs() {
		disc.rotate(rotationOrigin, rotationMatrix)
	}

	for _, facetStructure := range sn.GetFacetStructures() {
		facetStructure.rotate(rotationOrigin, rotationMatrix, rotatedPoints, rotatedNormals, rotatedVertexNormals, rotatedImageProjections)
	}

	for _, childNode := range sn.GetChildNodes() {
		childNode.rotate(rotationOrigin, rotationMatrix, rotatedPoints, rotatedNormals, rotatedVertexNormals, rotatedImageProjections)
	}
}
