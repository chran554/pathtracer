package obj

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"math"
	"path/filepath"
	"pathtracer/internal/pkg/obj/wavefrontobj"
	scn "pathtracer/internal/pkg/scene"
)

type CylinderType int

const (
	CylinderCentered  CylinderType = iota // CylinderCentered is a "standing" cylinder centered on origin (0,0,0).
	CylinderYPositive                     // CylinderPositive is a standing cylinder with the base centered on origin (0,0,0).
)

// NewCylinder return a box which sides all have the unit length 1.
// It is placed with one corner ni origin (0, 0, 0) and opposite corner in (1, 1, 1).
// Side normals point outwards from each side.
func NewCylinder(cylinderType CylinderType, radius float64, height float64) *scn.FacetStructure {
	cylinder := loadCylinder()

	switch cylinderType {
	case CylinderCentered: // Nothing by intention
	case CylinderYPositive:
		cylinder.Translate(&vec3.T{0, -cylinder.Bounds.Ymin, 0})
	}

	cylinder.Scale(&vec3.Zero, &vec3.T{2 * radius, height, 2 * radius})
	fmt.Printf("Cylinder bounds: %+v\n", cylinder.Bounds)

	return cylinder
}

func loadCylinder() *scn.FacetStructure {
	cylinder := wavefrontobj.ReadOrPanic(filepath.Join(ObjFileDir, "cylinder_no_caps.obj"))

	cylinder.CenterOn(&vec3.Zero)
	cylinder.RotateX(&vec3.Zero, math.Pi/2)

	xmin := cylinder.Bounds.Xmin
	xmax := cylinder.Bounds.Xmax
	ymin := cylinder.Bounds.Ymin
	ymax := cylinder.Bounds.Ymax
	zmin := cylinder.Bounds.Zmin
	zmax := cylinder.Bounds.Zmax

	cylinder.Scale(&vec3.Zero, &vec3.T{1.0 / (xmax - xmin), 1.0 / (ymax - ymin), 1.0 / (zmax - zmin)}) // resize/scale to height and radius == 1.0 units

	return cylinder
}
