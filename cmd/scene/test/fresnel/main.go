package main

import (
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "fresnel"

var ballRadius float64 = 20

var maxRecursionDepth = 8
var amountSamples = 1024 * 36

var viewPlaneDistance = 1500.0
var cameraDistanceFactor = 1.0

var imageWidth = 800
var imageHeight = 500
var magnification = 1.0

var roofHeight = ballRadius * 3.0

func main() {
	openBox := getCornellBox()
	openBox.Scale(&vec3.Zero, &vec3.T{2 * ballRadius * 3, roofHeight, 3 * ballRadius * 3})
	openBox.Translate(&vec3.T{-ballRadius * 3, 0, -2 * ballRadius * 3})

	lampHeight := roofHeight - 0.001
	lampSize := ballRadius * 2

	lamp := &scn.FacetStructure{
		Material: scn.NewMaterial().N("lamp").E(color.White, 3.0, true),
		Facets: getFlatRectangleFacets(
			&vec3.T{-lampSize, lampHeight, +lampSize},
			&vec3.T{+lampSize, lampHeight, +lampSize},
			&vec3.T{+lampSize, lampHeight, -lampSize},
			&vec3.T{-lampSize, lampHeight, -lampSize},
		),
	}

	rightSphereMaterial := scn.NewMaterial().N("right_sphere").
		C(color.NewColor(0.9, 0.9, 0.7)).
		M(0.31, 0.3).
		T(0.0, true, scn.RefractionIndex_Air)

	leftSphereMaterial := scn.NewMaterial().N("left_sphere").
		C(color.NewColor(0.7, 0.9, 0.9)).
		M(0.0, 0.0).
		T(0.0, true, scn.RefractionIndex_AcrylicPlastic)

	sphereX := ballRadius + (ballRadius / 2)
	sphere1 := scn.NewSphere(&vec3.T{sphereX, ballRadius, 0}, ballRadius, rightSphereMaterial).N("Right sphere")
	sphere2 := scn.NewSphere(&vec3.T{-sphereX, ballRadius, 0}, ballRadius, leftSphereMaterial).N("Left sphere")

	scene := scn.NewSceneNode().S(sphere1, sphere2).FS(openBox, lamp)

	cameraOrigin := vec3.T{0, ballRadius * 2, -14 * ballRadius}
	cameraOrigin.Scale(cameraDistanceFactor)
	focusPoint := vec3.T{0, ballRadius, 0}
	camera := scn.NewCamera(&cameraOrigin, &focusPoint, amountSamples, magnification).V(viewPlaneDistance).D(maxRecursionDepth)

	frame := scn.NewFrame(animationName, -1, camera, scene)

	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, true)
	animation.AddFrame(frame)

	anm.WriteAnimationToFile(animation, false)
}

func getCornellBox() *scn.FacetStructure {
	//roofTexture := scn.NewParallelImageProjection("textures/uv.png", vec3.T{0, ballRadius * 6, 0}, vec3.T{ballRadius, 0, 0}, vec3.T{0, 0, ballRadius})
	//floorTexture := scn.NewParallelImageProjection("textures/uv.png", vec3.T{0, 0, 0}, vec3.T{ballRadius, 0, 0}, vec3.T{0, 0, ballRadius})

	boxP1 := vec3.T{1, 1, 0} // Top right close            3----------2
	boxP2 := vec3.T{1, 1, 1} // Top right away            /          /|
	boxP3 := vec3.T{0, 1, 1} // Top left away            /          / |
	boxP4 := vec3.T{0, 1, 0} // Top left close          4----------1  |
	boxP5 := vec3.T{1, 0, 0} // Bottom right close      | (7)      |  6
	boxP6 := vec3.T{1, 0, 1} // Bottom right away       |          | /
	boxP7 := vec3.T{0, 0, 1} // Bottom left away        |          |/
	boxP8 := vec3.T{0, 0, 0} // Bottom left close       8----------5

	boxMaterial := scn.NewMaterial().N("box").C(color.Color{R: 0.90, G: 0.90, B: 0.90})
	leftWallMaterial := scn.NewMaterial().N("left_wall").C(color.Color{R: 0.95, G: 0.05, B: 0.05})
	rightWallMaterial := scn.NewMaterial().N("right_wall").C(color.Color{R: 0.05, G: 0.05, B: 0.95})

	box := &scn.FacetStructure{
		Name: "Open box",
		FacetStructures: []*scn.FacetStructure{
			getRectangleFacetStructure("Roof", boxMaterial, getFlatRectangleFacets(&boxP1, &boxP2, &boxP3, &boxP4)),
			getRectangleFacetStructure("Floor", boxMaterial, getFlatRectangleFacets(&boxP8, &boxP7, &boxP6, &boxP5)),
			getRectangleFacetStructure("Back wall", boxMaterial, getFlatRectangleFacets(&boxP6, &boxP7, &boxP3, &boxP2)),
			getRectangleFacetStructure("Right side wall", rightWallMaterial, getFlatRectangleFacets(&boxP6, &boxP2, &boxP1, &boxP5)),
			getRectangleFacetStructure("Left side wall", leftWallMaterial, getFlatRectangleFacets(&boxP7, &boxP8, &boxP4, &boxP3)),
		},
	}

	return box
}

func getRectangleFacetStructure(name string, material *scn.Material, facets []*scn.Facet) *scn.FacetStructure {
	return &scn.FacetStructure{SubstructureName: name, Material: material, Facets: facets}
}

func getFlatRectangleFacets(p1, p2, p3, p4 *vec3.T) []*scn.Facet {
	v1 := p2.Subed(p1)
	v2 := p3.Subed(p1)
	normal := vec3.Cross(&v1, &v2)
	normal.Normalize()

	return []*scn.Facet{
		{Vertices: []*vec3.T{p1, p2, p4}, Normal: &normal},
		{Vertices: []*vec3.T{p4, p2, p3}, Normal: &normal},
	}
}
