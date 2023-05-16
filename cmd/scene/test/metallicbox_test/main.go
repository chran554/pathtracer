package main

import (
	"github.com/ungerik/go3d/float64/mat3"
	"math"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj/wavefrontobj"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "metallicbox_test"

var cornellBoxFilename = "cornellbox.obj"
var cornellBoxFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/obj/" + cornellBoxFilename

var amountAnimationFrames = 90

var ballRadius float64 = 60

var amountSamples = 128 * 4 * 2 * 4

var viewPlaneDistance = 1000.0
var cameraDistanceFactor = 1.0

var imageWidth = 800
var imageHeight = 500
var magnification = 1.5

func main() {
	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, true, false)

	scale := 100.0
	cornellBox := NewCornellBox(100.0)

	footballMaterial := scn.NewMaterial().C(color.White).M(0.1, 0.6).SP("textures/equirectangular/football.png", &vec3.T{ballRadius + (ballRadius / 2), ballRadius, 0}, vec3.T{1, 0, 0}, vec3.T{0, 1, 0})
	football := scn.NewSphere(&vec3.T{ballRadius + (ballRadius / 2), ballRadius, 0}, ballRadius, footballMaterial).N("football")

	sphere2Material := scn.NewMaterial().C(color.NewColorGrey(0.9))
	sphere2 := scn.NewSphere(&vec3.T{-(ballRadius + (ballRadius / 2)), ballRadius, 0}, ballRadius, sphere2Material).N("Left sphere")

	mirrorSphereMaterial := scn.NewMaterial().C(color.NewColor(0.8, 0.85, 0.9)).M(0.95, 0.4)
	mirrorSphere1 := scn.NewSphere(&vec3.T{0, ballRadius / 1.5, ballRadius * 1.5}, ballRadius/1.5, mirrorSphereMaterial).N("Mirror sphere on floor")

	mirrorSphereRadius := scale * 0.75

	wallMirrorMaterial := scn.NewMaterial().C(color.NewColor(0.85, 0.85, 0.9)).M(0.95, 0.05)
	mirrorSphere2 := scn.NewSphere(&vec3.T{-(scale*2 + mirrorSphereRadius*0.25), scale + mirrorSphereRadius, scale*2 + mirrorSphereRadius*0.25}, mirrorSphereRadius, wallMirrorMaterial).N("Mirror sphere on wall left")
	mirrorSphere3 := scn.NewSphere(&vec3.T{scale*2 + mirrorSphereRadius*0.25, scale + mirrorSphereRadius, scale*2 + mirrorSphereRadius*0.25}, mirrorSphereRadius, wallMirrorMaterial).N("Mirror sphere on wall right")

	scene := scn.NewSceneNode().S(football, sphere2, mirrorSphere1, mirrorSphere2, mirrorSphere3).FS(cornellBox)

	for animationFrameIndex := 0; animationFrameIndex < amountAnimationFrames; animationFrameIndex++ {
		animationProgress := float64(animationFrameIndex) / float64(amountAnimationFrames)

		cameraOrigin := vec3.T{0, scale * 0.5, -600}
		cameraOrigin.Scale(cameraDistanceFactor)
		focusPoint := vec3.T{0, scale, -scale / 2}

		maxHeightAngle := math.Pi / 14
		heightAngle := maxHeightAngle * (1.0 + math.Cos(animationProgress*2*math.Pi-math.Pi)) / 2.0
		rotationXMatrix := mat3.T{}
		rotationXMatrix.AssignXRotation(heightAngle)
		animatedCameraOrigin := rotationXMatrix.MulVec3(&cameraOrigin)

		maxSideAngle := math.Pi / 8
		sideAngle := maxSideAngle * -1 * math.Sin(animationProgress*2*math.Pi)
		rotationYMatrix := mat3.T{}
		rotationYMatrix.AssignYRotation(sideAngle)
		animatedCameraOrigin = rotationYMatrix.MulVec3(&animatedCameraOrigin)

		camera := scn.NewCamera(&animatedCameraOrigin, &focusPoint, amountSamples, magnification).V(viewPlaneDistance)

		frame := scn.NewFrame(animation.AnimationName, animationFrameIndex, camera, scene)
		animation.AddFrame(frame)
	}

	anm.WriteAnimationToFile(animation, false)
}

func NewCornellBox(scale float64) *scn.FacetStructure {
	cornellBox := wavefrontobj.ReadOrPanic(cornellBoxFilenamePath)
	cornellBox.ScaleUniform(&vec3.Zero, scale)

	cornellBox.ReplaceMaterial("Right", scn.NewMaterial().N("Right").C(color.NewColor(0.9, 0.1, 0.1)).M(0.1, 0.2))
	cornellBox.ReplaceMaterial("Left", scn.NewMaterial().N("Left").C(color.NewColor(0.1, 0.1, 0.9)).M(0.1, 0.2))
	cornellBox.ReplaceMaterial("Back", scn.NewMaterial().N("Back").C(color.NewColorGrey(0.7)).M(0.1, 0.2))
	cornellBox.ReplaceMaterial("Ceiling", scn.NewMaterial().N("Ceiling").C(color.NewColorGrey(0.9)))
	cornellBox.ReplaceMaterial("Floor", scn.NewMaterial().N("Floor").C(color.NewColorGrey(0.8)).M(0.3, 0.05))

	cornellBox.GetFirstObjectBySubstructureName("Floor").Material.PP("textures/marble/white_marble.png", &vec3.T{0, 0, 0}, vec3.T{200 * 2, 0, 0}, vec3.T{0, 0, 200 * 2})

	lampMaterial := scn.NewMaterial().N("Lamp").E(color.White, 6, true)
	cornellBox.ReplaceMaterial("Lamp_1", lampMaterial)
	cornellBox.ReplaceMaterial("Lamp_2", lampMaterial)
	cornellBox.ReplaceMaterial("Lamp_3", lampMaterial)
	cornellBox.ReplaceMaterial("Lamp_4", lampMaterial)

	return cornellBox
}
