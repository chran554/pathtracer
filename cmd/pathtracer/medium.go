package main

import (
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/scene"
	"slices"
)

// traceCameraPath gives each primary ray private stack storage. After a pop,
// a later push can overwrite an initial entry even without growing the slice.
// initialMedia is a validated template built by Frame.InitialMediumStack.
func traceCameraPath(ray *scene.Ray, camera *scene.Camera, scn *scene.SceneNode, initialMedia []*scene.Material) *color.Color {
	return tracePath(ray, camera, scn, 0, slices.Clone(initialMedia))
}
