package obj

import (
	"path/filepath"
)

var (
	// pathtracerRoot = "/Users/christian/Projects/go/pathtracer" // Old mac
	pathtracerRoot = "/Users/christian/projects/code/go/pathtracer" // New mac
	ObjFileDir     = filepath.Join(pathtracerRoot, "objects/obj")
	PlyFileDir     = filepath.Join(pathtracerRoot, "objects/ply")
)
