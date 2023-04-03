package obj

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj/wavefrontobj"
	scn "pathtracer/internal/pkg/scene"
)

func NewLampPost(lampPostScale *vec3.T) *scn.FacetStructure {
	lampPostMaterial := scn.NewMaterial().N("lamp post").C(color.NewColor(0.9, 0.4, 0.3)).M(0.2, 0.3)

	lampMaterial := scn.NewMaterial().N("lamp").C(color.White).E(color.White, 2.0, true)

	lampPost := loadLampPost(lampPostScale)
	lampPost.ClearMaterials()
	lampPost.Material = lampPostMaterial

	lampPost.GetFirstObjectByName("lamp_0").Material = lampMaterial
	lampPost.GetFirstObjectByName("lamp_1").Material = lampMaterial
	lampPost.GetFirstObjectByName("lamp_2").Material = lampMaterial
	lampPost.GetFirstObjectByName("lamp_3").Material = lampMaterial
	return lampPost
}

func loadLampPost(scale *vec3.T) *scn.FacetStructure {
	lampPost := wavefrontobj.ReadOrPanic(filepath.Join(ObjFileDir, "lamp_post.obj"))

	ymin := lampPost.Bounds.Ymin
	ymax := lampPost.Bounds.Ymax
	lampPost.Translate(&vec3.T{0.0, -ymin, 0.0})       // lamp post base touch the ground (xz-plane)
	lampPost.ScaleUniform(&vec3.Zero, 1.0/(ymax-ymin)) // resize to height == 1.0

	lampPost.Scale(&vec3.Zero, scale)

	fmt.Printf("Lamp post bounds: %+v\n", lampPost.Bounds)

	return lampPost
}
