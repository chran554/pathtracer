package obj

import (
	"path/filepath"
)

var (
	pathtracerRoot       = "."
	ObjFileDir           = filepath.Join(pathtracerRoot, "objects/obj")
	ObjEvaluationFileDir = filepath.Join(pathtracerRoot, "objects")
	PlyFileDir           = filepath.Join(pathtracerRoot, "objects/ply")
	TexturesDir          = filepath.Join(pathtracerRoot, "textures")
)
