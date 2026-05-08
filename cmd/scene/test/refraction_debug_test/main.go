package main

import (
	"fmt"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"
	"pathtracer/internal/pkg/util"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "refraction_debug_test"
var animationFrameCount = 1
var startFrameIndex = 0
var refractionIndexStart = scene.RefractionIndex_Glass
var refractionIndexEnd = scene.RefractionIndex_Diamond

var ballRadius float64 = 35

var maxRecursionDepth = 8
var amountSamples = 512 * 1 // * 2 // * 5 // * 4
var lensRadius = 0.0        // 2.0        // 0.25

var imageWidth = 500
var imageHeight = 240
var magnification = 0.75

func main() {
	//textureBackplane := floatimage.LoadOrPanic("textures/floor/7451-diffuse 02.png")
	textureBackplane := floatimage.LoadOrPanic("textures/ground/seamless-cobblestone-texture.jpeg")

	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)

	backplane := &scene.FacetStructure{
		Name: "backplane",
		Material: scene.NewMaterial().N("backplane").
			C(color.NewColor(0.95, 0.95, 0.95)).
			E(color.White, 0.05, false),
		Facets: obj.NewSquare(obj.SquareTypeXYPlane, textureBackplane),
	}
	backplane.Scale(&vec3.Zero, &vec3.T{200, 200, 1})
	backplane.Translate(&vec3.T{-backplane.Bounds.SizeX() / 2, -backplane.Bounds.SizeY() / 2, 0})

	//lamp1 := scene.NewSphere(&vec3.T{-100, 100, -100}, 40, scene.NewMaterial().E(color.White, 4, true)).N("lamp1")

	for frameIndex := startFrameIndex; frameIndex < animationFrameCount; frameIndex++ {
		progress := 1.0
		if animationFrameCount > 1 {
			progress = float64(frameIndex) / float64(animationFrameCount-1)
		}

		// refractionIndex := (refractionIndexStart * (1.0 - progress)) + (refractionIndexEnd * progress)
		refractionIndex := scene.RefractionIndex_Glass

		sphereMaterial := scene.NewMaterial().
			N("glass").
			C(color.NewColor(0.90, 0.92, 0.95)).
			M(0.270, 0.030).
			T(0.700, true, refractionIndex)

		sphereOffset := 0.02

		origin1 := &vec3.T{-(ballRadius + sphereOffset), 0, -(ballRadius + sphereOffset)}
		sphere1 := scene.NewSphere(origin1, ballRadius, sphereMaterial).N("sphere1")

		origin2 := &vec3.T{+(ballRadius + sphereOffset), 0, -(ballRadius + sphereOffset)}
		sphere2 := scene.NewSphere(origin2, ballRadius, sphereMaterial).N("sphere2")

		drinkingGlass := obj.NewGlassIkeaSkoja(20, false)
		drinkingGlass.Translate(origin2)

		lampRadius := 40.0
		lampOrigin := &vec3.T{-200, 60, 0}
		lampIntensity := 5.0
		lamp2 := scene.NewSphere(lampOrigin, lampRadius, scene.NewMaterial().E(color.White, lampIntensity, true)).N("lamp2")

		// Animate lamp position
		lampAnimationOrigin := origin1.Added(origin2)
		lampAnimationOrigin.Scale(0.5)
		lamp2.RotateY(&lampAnimationOrigin, util.DegToRad(180.0*progress))

		//scn := scene.NewSceneNode().S(sphere1).FS(backplane).S(lamp2).FS(pokal)
		scn := scene.NewSceneNode().FS(backplane).S(lamp2).FS(drinkingGlass)

		b1 := sphere1.Bounds()
		b2 := sphere2.Bounds()
		cameraOrigin := &vec3.T{(b1.Center()[0] + b2.Center()[0]) / 2, b1.Center()[1], -300}
		cameraFocusPoint := &vec3.T{(b1.Center()[0] + b2.Center()[0]) / 2, b1.Center()[1], -(ballRadius * 2 / 3)}
		camera := scene.NewCamera(cameraOrigin, cameraFocusPoint, amountSamples, magnification).D(maxRecursionDepth).A(lensRadius, nil)

		fi := frameIndex
		if animationFrameCount == 1 {
			fi = -1
		}

		frame := scene.NewFrame(animation.AnimationName, fi, camera, scn)
		animation.AddFrame(frame)
	}

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}
