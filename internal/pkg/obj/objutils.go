package obj

import (
	"path/filepath"
)

var (
	resourcesRoot        = "."
	ObjFileDir           = filepath.Join(resourcesRoot, "objects/obj")
	ObjEvaluationFileDir = filepath.Join(resourcesRoot, "objects")
	PlyFileDir           = filepath.Join(resourcesRoot, "objects/ply")
	TexturesDir          = filepath.Join(resourcesRoot, "textures")
)

func SetResourceRoot(resourceRoot string) {
	resourcesRoot = resourceRoot
	ObjFileDir = filepath.Join(resourcesRoot, "objects/obj")
	PlyFileDir = filepath.Join(resourcesRoot, "objects/ply")
	ObjEvaluationFileDir = filepath.Join(resourcesRoot, "objects")
}
