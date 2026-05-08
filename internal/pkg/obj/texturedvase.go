package obj

import (
	"fmt"
	"math"
	"path/filepath"
	"pathtracer/internal/pkg/fileformat/wavefront"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

// NewTexturedVase creates a new vase with the center of the vase bottom in origin (0,0,0) and a height of scale.
func NewTexturedVase(scale float64) *scene.FacetStructure {
	object := loadTexturedVase()
	object.ScaleUniform(&vec3.Zero, scale)

	return object
}

func loadTexturedVase() *scene.FacetStructure {
	objectFilename := filepath.Join(ObjEvaluationFileDir, "textured_vase", "textured_vase_obj", "Textured Vase.obj")
	object := wavefront.ReadFacetStructureOrPanic(objectFilename)

	object.RotateX(&vec3.Zero, math.Pi/2)
	object.FlipX(&vec3.Zero)
	object.RotateY(&vec3.Zero, math.Pi/2)
	object.CenterOn(&vec3.Zero)
	object.Translate(&vec3.T{0, -object.Bounds.Ymin, 0})
	object.ScaleUniform(&vec3.Zero, 1.0/object.Bounds.SizeY())

	object.UpdateNormals()

	fmt.Printf("Texture vase bounds: %+v\n", object.Bounds)

	return object
}
