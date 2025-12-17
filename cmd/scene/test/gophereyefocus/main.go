package main

import (
	"fmt"
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"
	"pathtracer/internal/pkg/util"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "gophereyefocus"

var amountAnimationFrames = 96 // 72 * 2

var imageWidth = 400
var imageHeight = 400
var magnification = 0.5

var amountSamples = 512 * 4
var maxRecursionDepth = 4

var cameraAperture = 0.0
var cameraZoom = 1.0

var lightIntensityFactor = 1.65

func main() {
	textureGround, err := floatimage.EmptyPlaceholderImage("textures/floor/7451-diffuse 02 low contrast.png")
	if err != nil {
		panic(err)
	}

	// Cornell box
	roomScale := 2.0
	cornellBox := obj.NewWhiteCornellBox(&vec3.T{4 * roomScale, 2.5, 4 * roomScale}, true, lightIntensityFactor) // cm, as units. I.e. a 5x3x5m room
	floorMaterial := cornellBox.GetFirstMaterialByName("Floor")
	floorMaterial.M(0.01, 0.3).PP(textureGround, &vec3.Zero, vec3.T{1, 0, 0}, vec3.T{0, 0, 1})
	floorMaterial.FresnelMaxGlossiness = 0.15

	// Gopher
	gopherPupilMaterial := scene.NewMaterial().C(color.NewColorGrey(0.00)).M(0.01, 0.05)
	gopherPupilMaterial.FresnelMaxGlossiness = 0.05
	gopherNoseTipMaterial := scene.NewMaterial().C(color.NewColorGrey(0.01)).M(0.01, 0.30)

	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)

	for frameIndex := 0; frameIndex < amountAnimationFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountAnimationFrames)

		gopher := obj.NewGopher(0.4)
		gopher.ReplaceMaterial("nose_tip", gopherNoseTipMaterial)
		gopher.ReplaceMaterial("eye_pupil_left", gopherPupilMaterial)
		gopher.ReplaceMaterial("eye_pupil_right", gopherPupilMaterial)
		gopher.RotateY(&vec3.Zero, util.DegToRad(180-15))

		gopherNoseTipCenter := gopher.GetFirstObjectBySubstructureName("nose_tip").Bounds.Center()

		cameraPosition := &vec3.T{0, 0.6, -1.0}

		if amountAnimationFrames > 1 {
			easeInEaseOutHeight := 0.5*math.Sin(util.DegToRad(360*animationProgress-90)) + 0.5
			easeInEaseOut := 0.5*math.Sin(util.DegToRad(180*animationProgress-90)) + 0.5
			cameraHeight := 0.3 + 0.7*easeInEaseOutHeight
			cameraX := 1.0 * math.Cos(util.DegToRad(180*easeInEaseOut-180))
			cameraZ := 1.0 * math.Sin(util.DegToRad(180*easeInEaseOut-180))
			cameraPosition = &vec3.T{cameraX, cameraHeight, cameraZ}
		}

		cameraAim := gopher.Bounds.Center()
		focusPosition := gopherNoseTipCenter
		cameraFocusVector := focusPosition.Subed(cameraPosition)

		camera := scene.NewCamera(cameraPosition, cameraAim, amountSamples, magnification).
			D(maxRecursionDepth).
			A(cameraAperture, nil).
			F(cameraFocusVector.Length()).
			V(800 * cameraZoom)

		// Change gopher eyes
		changeEyeFocus(gopher, camera.Origin)

		scn := scene.NewSceneNode().FS(cornellBox, gopher)

		fi := -1
		if amountAnimationFrames > 1 {
			fi = frameIndex
		}

		frame := scene.NewFrame(animationName, fi, camera, scn)
		animation.AddFrame(frame)
	}

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err = renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}

func changeEyeFocus(gopher *scene.FacetStructure, pointOfFocus *vec3.T) {
	leftPupil := gopher.GetFirstObjectBySubstructureName("eye_pupil_left")
	leftEyeBall := gopher.GetFirstObjectBySubstructureName("eye_ball_left")
	rightPupil := gopher.GetFirstObjectBySubstructureName("eye_pupil_right")
	rightEyeBall := gopher.GetFirstObjectBySubstructureName("eye_ball_right")

	changeEyePupil(pointOfFocus, leftEyeBall, leftPupil)
	changeEyePupil(pointOfFocus, rightEyeBall, rightPupil)
}

func changeEyePupil(pointOfFocus *vec3.T, eyeBall *scene.FacetStructure, pupil *scene.FacetStructure) {
	eyeBallCenter := eyeBall.Bounds.Center()

	oldPupilViewDirection := pupil.Bounds.Center().Subed(eyeBallCenter)
	newPupilViewDirection := pointOfFocus.Subed(eyeBallCenter)

	ray := &scene.Ray{Origin: eyeBallCenter, Heading: &newPupilViewDirection}
	_, _, newPupilPosition, _, _, _ := scene.FacetStructureIntersection(ray, eyeBall, nil)

	axis, angle, ok := RotationFromTo(&oldPupilViewDirection, &newPupilViewDirection)
	if ok {
		// pupil to position in the direction of view
		pupil.RotateAxis(eyeBallCenter, axis, angle)

		// adjust the pupil position so the pupil is placed half-way into the eyeball
		pupilTranslation := newPupilPosition.Subed(pupil.Bounds.Center())
		pupil.Translate(&pupilTranslation)
	} else {
		// Simple translation of the pupil
		pupilTranslation := newPupilPosition.Subed(pupil.Bounds.Center())
		pupil.Translate(&pupilTranslation)
	}
}

func RotationFromTo(v1, v2 *vec3.T) (axis *vec3.T, angle float64, ok bool) {
	if v1.Length() == 0 || v2.Length() == 0 {
		return &vec3.Zero, 0, false
	}

	u := v1.Normalized()
	v := v2.Normalized()

	axis_ := vec3.Cross(&u, &v)
	axis = &axis_
	s := axis.Length()
	c := vec3.Dot(&u, &v)

	const eps = 1e-12
	if s < eps {
		if c > 0 {
			// already parallel
			return &vec3.T{0, 1, 0}, 0, true
		}
		// opposite: pick any perpendicular axis
		tmp := &vec3.T{1, 0, 0}
		if math.Abs(u[0]) > 0.9 {
			tmp = &vec3.T{0, 1, 0}
		}
		axis_ = vec3.Cross(&u, tmp)
		axis = &axis_
		axis.Normalize()
		return axis, math.Pi, true
	}

	axis.Scale(1.0 / s)      // normalize axis
	angle = math.Atan2(s, c) // stable angle
	return axis, angle, true
}
