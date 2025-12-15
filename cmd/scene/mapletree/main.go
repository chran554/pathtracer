package main

import (
	"fmt"
	"math"
	"math/rand/v2"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"
	"pathtracer/internal/pkg/util"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "mapletree"

var imageWidth = 300
var imageHeight = 600
var magnification = 2.0 // 0.5

var amountSamples = 512 // 512

var skyDomeEmission = 1.0

var maxRayDepth = 8

var leafCount = 512 * 12
var leafScale = 1.75

var rnd *rand.Rand

func init() {
	rnd = rand.New(rand.NewPCG(1, 2))
}

func main() {
	skyDomeRadius := 150.0

	textureGround, err := floatimage.EmptyPlaceholderImage("textures/equirectangular/336_PDM_BG7.jpg")
	if err != nil {
		panic(err)
	}
	textureEnv, err := floatimage.EmptyPlaceholderImage("textures/equirectangular/336_PDM_BG7.jpg")
	if err != nil {
		panic(err)
	}
	textureLeaf, err := floatimage.EmptyPlaceholderImage("textures/tree/Leaves0120_35_S_02.png")
	if err != nil {
		panic(err)
	}

	skydomeMaterial := scene.NewMaterial().
		E(color.White, skyDomeEmission, true).
		SP(textureEnv, &vec3.T{0, 0, 0}, vec3.T{1, 0, 0}, vec3.T{0, 1, 0})

	skyDome := scene.NewSphere(&vec3.T{0, 0, 0}, skyDomeRadius, skydomeMaterial).N("sky dome")
	skyDome.RotateY(&vec3.Zero, util.DegToRad(-20))
	//skyDome.Translate(&vec3.T{0, 2, 0})

	ground := &scene.FacetStructure{Facets: obj.NewSquare(obj.SquareTypeXZPlane, false)}
	ground.Translate(&vec3.T{-0.5, 0, -0.5})
	ground.ScaleUniform(&vec3.T{0, 0, 0}, skyDomeRadius*2)
	ground.Translate(&vec3.T{0, -2, 0})
	groundMaterial := scene.NewMaterial().E(color.White, skyDomeEmission, true).SP(textureGround, &vec3.T{0, skyDomeRadius, 0}, vec3.T{1, 0, 0}, vec3.T{0, 1, 0})

	ground.Material = groundMaterial

	// Add leafs
	var leaves []*scene.FacetStructure

	leafMaterial := scene.NewMaterial().TP(textureLeaf).
		C(color.NewColorRGBA(1.0, 1.0, 1.0, 1.0)).
		T(0.05, false, 1.44). // Refraction index green leaves 1.40-1.48
		M(0.15, 0.85)

	for leafIndex := 0; leafIndex < leafCount; leafIndex++ {
		leaf := obj.NewSquareFacetStructure(obj.SquareTypeXYPlane, true, true)
		leaf.Material = leafMaterial

		// Scale leaf
		maxLeafWidth := 0.14 * leafScale // 14 cm
		minLeafWidth := 0.10 * leafScale // 10 cm
		leaf.Scale(&vec3.T{0, 0, 0}, &vec3.T{random(minLeafWidth, maxLeafWidth), random(minLeafWidth, maxLeafWidth), random(minLeafWidth, maxLeafWidth)})

		// Distort leaf facets by vertex distortion
		maxDistortion := 0.03 * leafScale // 2cm max distortion on leaf vertices
		for _, leafFacet := range leaf.Facets {
			for _, vertex := range leafFacet.Vertices {
				vertex.Add(&vec3.T{random(-maxDistortion, maxDistortion), random(-maxDistortion, maxDistortion), random(-maxDistortion, maxDistortion)})
			}
			leafFacet.GetBounds()
			leafFacet.UpdateNormal()
		}

		// Rotate leaf
		leaf.RotateY(&vec3.T{0, 0, 0}, random(0, math.Pi*2))
		leaf.RotateX(&vec3.T{0, 0, 0}, random(-math.Pi/2, math.Pi/2))
		leaf.RotateZ(&vec3.T{0, 0, 0}, random(-math.Pi/2, math.Pi/2))

		// Move leaf to position
		leafCloudRadius := 2.5 // 2.5m radius. The radius of the tree crown
		leafTowardsSphereShellFactor := 4.0
		leafTranslationRadius := leafCloudRadius * math.Pow(random(0, 1), 1.0/leafTowardsSphereShellFactor)
		leafTranslation := UniformOnSphereGaussian()
		leafTranslation.Normalize().Scale(leafTranslationRadius)
		leaf.Translate(leafTranslation)

		leaf.Translate(&vec3.T{0, 2 + leafCloudRadius, 0}) // Move tree crown 2m above ground

		leaves = append(leaves, leaf)
	}

	treeCrown := &scene.FacetStructure{Name: "tree crown", FacetStructures: leaves}

	scn := scene.NewSceneNode().S(skyDome).FS(ground).FS(treeCrown)

	cameraOrigin := &vec3.T{0, 2, -15}
	focusPoint := &vec3.T{0, 3, 0}

	viewVector := focusPoint.Subed(cameraOrigin)
	focusDistance := viewVector.Length()

	camera := scene.NewCamera(cameraOrigin, focusPoint, amountSamples, magnification).
		F(focusDistance).
		D(maxRayDepth)

	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, true, true)
	frame := scene.NewFrame(animation.AnimationName, -1, camera, scn)
	animation.AddFrame(frame)

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err = renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}

func random(min, max float64) float64 {
	return rnd.Float64()*(max-min) + min
}

func UniformOnSphereGaussian() *vec3.T {
	for {
		x := rnd.NormFloat64()
		y := rnd.NormFloat64()
		z := rnd.NormFloat64()
		n2 := x*x + y*y + z*z
		if n2 > 0 {
			inv := 1.0 / math.Sqrt(n2)
			return &vec3.T{x * inv, y * inv, z * inv}
		}
	}
}
