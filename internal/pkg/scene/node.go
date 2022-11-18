package scene

import "github.com/ungerik/go3d/float64/vec3"

type SceneNode struct {
	Spheres         []*Sphere
	Discs           []*Disc
	ChildNodes      []*SceneNode
	FacetStructures []*FacetStructure
	Bounds          *Bounds `json:"-"`
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

func (sn *SceneNode) Translate(translation *vec3.T) {
	for _, sphere := range sn.GetSpheres() {
		sphere.Translate(translation)
	}

	for _, disc := range sn.GetDiscs() {
		disc.Translate(translation)
	}

	for _, facetStructure := range sn.GetFacetStructures() {
		facetStructure.Translate(translation)
	}

	for _, childNode := range sn.GetChildNodes() {
		childNode.Translate(translation)
	}
}

func (sn *SceneNode) RotateX(rotationOrigin *vec3.T, angle float64) {
	for _, sphere := range sn.GetSpheres() {
		sphere.RotateX(rotationOrigin, angle)
	}

	for _, disc := range sn.GetDiscs() {
		disc.RotateX(rotationOrigin, angle)
	}

	for _, facetStructure := range sn.GetFacetStructures() {
		facetStructure.RotateX(rotationOrigin, angle)
	}

	for _, childNode := range sn.GetChildNodes() {
		childNode.RotateX(rotationOrigin, angle)
	}
}

func (sn *SceneNode) RotateY(rotationOrigin *vec3.T, angle float64) {
	for _, sphere := range sn.GetSpheres() {
		sphere.RotateY(rotationOrigin, angle)
	}

	for _, disc := range sn.GetDiscs() {
		disc.RotateY(rotationOrigin, angle)
	}

	for _, facetStructure := range sn.GetFacetStructures() {
		facetStructure.RotateY(rotationOrigin, angle)
	}

	for _, childNode := range sn.GetChildNodes() {
		childNode.RotateY(rotationOrigin, angle)
	}
}

func (sn *SceneNode) RotateZ(rotationOrigin *vec3.T, angle float64) {
	for _, sphere := range sn.GetSpheres() {
		sphere.RotateZ(rotationOrigin, angle)
	}

	for _, disc := range sn.GetDiscs() {
		disc.RotateZ(rotationOrigin, angle)
	}

	for _, facetStructure := range sn.GetFacetStructures() {
		facetStructure.RotateZ(rotationOrigin, angle)
	}

	for _, childNode := range sn.GetChildNodes() {
		childNode.RotateZ(rotationOrigin, angle)
	}
}
