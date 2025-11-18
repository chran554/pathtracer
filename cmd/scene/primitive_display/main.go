package main

import (
	"fmt"
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	anm "pathtracer/internal/pkg/renderfile"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "primitive_display"

var amountAnimationFrames = 1 // 72 * 2

var imageWidth = 800
var imageHeight = 200
var magnification = 1.0

var amountSamples = 256 * 2

var cameraOrigin = vec3.T{0, 150, -800}
var cameraDistanceFactor = 0.8
var viewPlaneDistance = 600.0
var cameraAperture = 15.0

var lightIntensityFactor = 7.0

func main() {
	// Cornell box
	cornellBox := obj.NewWhiteCornellBox(&vec3.T{500 * 4, 300 * 2, 500 * 4}, false, lightIntensityFactor) // cm, as units. I.e. a 5x3x5m room
	cornellBox.GetFirstMaterialByName("floor").M(0.01, 0.3)

	// Gopher
	gopherPupilMaterial := scn.NewMaterial().C(color.NewColorGrey(0.0)).M(0.05, 0.05)

	gopherBlue := obj.NewGopher(180.0)
	gopherBlue.ReplaceMaterial("eye_pupil", gopherPupilMaterial)
	gopherBlue.Translate(&vec3.T{0, -gopherBlue.Bounds.Ymin, 0})
	gopherBlue.ScaleUniform(&vec3.Zero, 2.0)
	gopherBlue.RotateY(&vec3.Zero, math.Pi*5/6)
	gopherBlue.Translate(&vec3.T{800, 0, 800})

	gopherPurple := obj.NewGopher(180.0)
	gopherPurple.ReplaceMaterial("body", scn.NewMaterial().C(color.NewColor(0.72, 0.55, 0.90)))
	gopherPurple.ReplaceMaterial("eye_pupil", gopherPupilMaterial)
	gopherPurple.Translate(&vec3.T{0, -gopherPurple.Bounds.Ymin, 0})
	gopherPurple.ScaleUniform(&vec3.Zero, 2.0)
	gopherPurple.RotateY(&vec3.Zero, -math.Pi*4/6)
	gopherPurple.Translate(&vec3.T{-800, 0, 800})

	gopherYellow := obj.NewGopher(180.0)
	gopherYellow.ReplaceMaterial("body", scn.NewMaterial().C(color.NewColor(0.90, 0.86, 0.55)))
	gopherYellow.ReplaceMaterial("eye_pupil", gopherPupilMaterial)
	gopherYellow.Translate(&vec3.T{0, -gopherYellow.Bounds.Ymin, 0})
	gopherYellow.ScaleUniform(&vec3.Zero, 2.0)
	gopherYellow.RotateY(&vec3.Zero, -math.Pi*7/8)
	gopherYellow.Translate(&vec3.T{-300, 0, 800})

	// Diamond
	diamond := obj.NewDiamond(200)
	diamond.RotateX(&vec3.Zero, 0.7173303226) // Pavilion angle, lay diamond on the side on the floor
	diamond.RotateY(&vec3.Zero, -math.Pi*3/6)
	diamond.Translate(&vec3.T{600, 0, 700})
	//diamond.Translate(&vec3.T{0, 0, -200})
	diamond.Material = scn.NewMaterial().
		C(color.NewColor(0.85, 0.85, 0.75)).
		E(color.NewColor(0.1, 0.08, 0.05), 1.0, false).
		M(0.05, 0.0)

	podiumHeight := 30.0
	podiumWidth := 200.0
	interPodiumDistance := 300.0

	triangleSize := 80.0 * 2.0
	sphereRadius := 80.0
	discRadius := 80.0

	sphereLocation := vec3.T{-interPodiumDistance, 0, 0}
	triangleLocation := vec3.T{0, 0, 0}
	discLocation := vec3.T{interPodiumDistance, 0, 0}

	podiumMaterial := scn.NewMaterial().C(color.NewColorGrey(0.9))

	// Sphere primitive
	spherePodium := obj.NewBox(obj.BoxCenteredYPositive)
	spherePodium.Material = podiumMaterial
	spherePodium.Scale(&vec3.Zero, &vec3.T{podiumWidth, podiumHeight, podiumWidth})
	spherePodium.Translate(&sphereLocation)

	sphereMaterial := scn.NewMaterial().C(color.NewColor(0.70, 1.00, 0.70)).M(0.1, 0.2)
	sphere := scn.NewSphere(&vec3.T{0, 0, 0}, sphereRadius, sphereMaterial).N("Sphere primitive")
	sphere.Translate(&vec3.T{0.0, podiumHeight + sphereRadius, 0.0})
	sphere.Translate(&sphereLocation)

	// Triangle primitive
	trianglePodium := obj.NewBox(obj.BoxCenteredYPositive)
	trianglePodium.Material = podiumMaterial
	trianglePodium.Scale(&vec3.Zero, &vec3.T{podiumWidth, podiumHeight, podiumWidth})
	trianglePodium.Translate(&triangleLocation)

	triangle := trianglePrimitive()
	triangle.ScaleUniform(&vec3.Zero, triangleSize)
	triangle.RotateX(&vec3.Zero, -math.Pi/8)
	triangle.RotateY(&vec3.Zero, -math.Pi/4)
	triangle.Translate(&vec3.T{0.0, podiumHeight, 0.0})
	triangle.Translate(&triangleLocation)

	// Disc primitive
	discPodium := obj.NewBox(obj.BoxCenteredYPositive)
	discPodium.Material = podiumMaterial
	discPodium.Scale(&vec3.Zero, &vec3.T{podiumWidth, podiumHeight, podiumWidth})
	discPodium.Translate(&discLocation) // Move podium to location

	discMaterial := scn.NewMaterial().C(color.NewColor(1.00, 0.70, 0.70)).M(0.1, 0.2)
	disc := scn.NewDisc(&vec3.T{0, 0, 0}, &vec3.T{0, 0, -1}, discRadius, discMaterial).N("Disc primitive")
	disc.RotateX(&vec3.Zero, -math.Pi/8)
	disc.RotateY(&vec3.Zero, -math.Pi*2/8)
	disc.Translate(&vec3.T{0.0, podiumHeight + discRadius, 0.0})
	disc.Translate(&discLocation) // Move disc to location

	// Scene
	scene := scn.NewSceneNode().
		S(sphere).
		D(disc).
		FS(cornellBox, triangle, spherePodium, discPodium, trianglePodium, gopherBlue, gopherPurple, gopherYellow)

	animationStartIndex := 0
	animationEndIndex := amountAnimationFrames - 1

	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)

	for frameIndex := animationStartIndex; frameIndex <= animationEndIndex; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountAnimationFrames)

		camera := getCamera(animationProgress, sphereRadius+podiumHeight)

		frame := scn.NewFrame(animationName, frameIndex, camera, scene)
		animation.AddFrame(frame)
	}

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := anm.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}

func trianglePrimitive() *scn.FacetStructure {
	material := scn.NewMaterial().C(color.NewColor(0.70, 0.70, 1.00)).M(0.1, 0.2)
	triangleHeight := 1.0
	triangleWidth := triangleHeight / (2.0 * math.Cos(math.Pi/6.0))
	triangle := scn.Facet{
		Vertices: []*vec3.T{
			{0, 0, 0},              //  p3 *---* p2
			{triangleWidth, 1, 0},  //      \ /
			{-triangleWidth, 1, 0}, //       * p1
		},
	}
	facetStructure := scn.FacetStructure{
		SubstructureName: "Triangle primitive",
		Material:         material,
		Facets:           []*scn.Facet{&triangle},
	}
	facetStructure.UpdateNormals()

	return &facetStructure
}

func getCamera(animationProgress float64, focusHeight float64) *scn.Camera {
	var cameraOrigin = cameraOrigin // vec3.T{0, 150, -800}
	var cameraFocus = vec3.T{0, focusHeight, 0}

	startAngle := -math.Pi / 2.0

	cameraOriginRadius := 600.0
	cameraOriginAnimationTranslation := vec3.T{
		math.Cos(startAngle-animationProgress*2.0*math.Pi) * cameraOriginRadius,
		0,
		math.Sin(startAngle-animationProgress*2.0*math.Pi) * cameraOriginRadius / 2.0,
	}
	cameraOrigin.Add(&cameraOriginAnimationTranslation)

	cameraFocusRadius := 200.0 // Same as inter podium distance
	cameraFocusAnimationTranslation := vec3.T{
		math.Cos(startAngle-animationProgress*2.0*math.Pi) * cameraFocusRadius,
		0,
		0,
	}
	cameraFocus.Add(&cameraFocusAnimationTranslation)

	origin := cameraOrigin.Scaled(cameraDistanceFactor)

	// Still image camera settings
	// origin = cameraOrigin.Added(&vec3.T{0, 0, 0})
	// origin.Scale(cameraDistanceFactor)
	// cameraFocus = vec3.T{0, focusHeight * 1.5, 0}

	return scn.NewCamera(&origin, &cameraFocus, amountSamples, magnification).V(viewPlaneDistance).A(cameraAperture, nil)
}
