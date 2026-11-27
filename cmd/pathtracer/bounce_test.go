package main

import (
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

// bounceTestScene returns a new scene with a triangle at z=0 using material,
// facing +Z, and surrounded by a spherical emitter with unit white radiance.
func bounceTestScene(material *scene.Material) *scene.SceneNode {
	return &scene.SceneNode{
		FacetStructures: []*scene.FacetStructure{orientationTestPlane(material)},
		Spheres: []*scene.Sphere{scene.NewSphere(&vec3.T{}, 100,
			scene.NewMaterial().E(color.White, 1, true))},
	}
}

// bounceTestTrace returns the radiance and toOpacity of one path through sceneNode,
// starting at (0,0,1) toward -Z with refractive index one and recursion depth one.
func bounceTestTrace(sceneNode *scene.SceneNode) *color.Color {
	ray := &scene.Ray{Origin: &vec3.T{0, 0, 1}, Heading: &vec3.T{0, 0, -1}}
	camera := &scene.Camera{RenderType: scene.Pathtracing, RecursionDepth: 1}
	return tracePath(ray, camera, sceneNode, 0, []*scene.Material{scene.NewMaterial().T(1, true, 1)})
}
