package scene

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

func (sn *SceneNode) GetFacetStructures() []*FacetStructure {
	return sn.FacetStructures
}
