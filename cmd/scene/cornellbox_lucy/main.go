package main

import (
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "cornellbox_lucy"

var cornellBoxUnit float64 = 60

var amountSamples = 1024 * 1 // 1024 * 8 // 1024 * 32
var maxRayDepth = 5          // 4      // 6 // max ray recursion depth
var apertureSize = 1.25

var imageWidth = 400
var imageHeight = 400
var magnification = 0.5

var viewPlaneDistance = 1500.0

var lampIntensity = 5.0 * 3

func main() {
	cornellBox := obj.NewCornellBox(&vec3.T{cornellBoxUnit, cornellBoxUnit, 3 * cornellBoxUnit}, true, lampIntensity)
	cornellBox.GetFirstObjectByName("Lamp").Scale(&vec3.Zero, &vec3.T{0.35, 1.0, 1.0})

	floor := cornellBox.GetFirstObjectByMaterialName("Floor")
	floor.Material = floor.Material.Copy()
	floor.Material.PP("textures/floor/granite_tiles.jpg", &vec3.Zero, vec3.T{cornellBoxUnit, 0, 0}, vec3.T{0, 0, cornellBoxUnit})
	floor.Material.M(0.0, 0.75)
	floor.Material.Color = floor.Material.Color.Copy()
	floor.Material.Color.Multiply(0.85)

	backWall := cornellBox.GetFirstObjectByMaterialName("Back")
	backWall.Material = backWall.Material.Copy()
	backWall.Material.Color = backWall.Material.Color.Copy()
	//backWall.Material.PP("textures/tapeter/ArchiveLandscape_Image_Flatshot_Item_9477w.jpg", &vec3.T{-cornellBoxUnit / 2, 0, 0}, vec3.T{cornellBoxUnit * 1.5, 0, 0}, vec3.T{0, cornellBoxUnit * 2 * 1.5, 0})
	backWall.Material.PP("textures/concrete/rough_plaster_bright.png", &vec3.T{-cornellBoxUnit / 2, 0, 0}, vec3.T{cornellBoxUnit / 2, 0, 0}, vec3.T{0, cornellBoxUnit / 2, 0})

	leftWall := cornellBox.GetFirstObjectByMaterialName("Left")
	leftWall.Material = leftWall.Material.Copy()
	//leftWall.Material.PP("textures/concrete/Polished-Concrete-Architextures.jpg", &vec3.T{0, 0, -cornellBoxUnit * 3 / 2}, vec3.T{0, 0, cornellBoxUnit * 3 * 1.3}, vec3.T{0, cornellBoxUnit, 0})
	leftWall.Material.PP("textures/concrete/rough_plaster_bright.png", &vec3.T{0, 0, -cornellBoxUnit * 3 / 2}, vec3.T{0, 0, -cornellBoxUnit * 3 / 2}, vec3.T{0, cornellBoxUnit / 2, 0})

	rightWall := cornellBox.GetFirstObjectByMaterialName("Right")
	rightWall.Material = rightWall.Material.Copy()
	//rightWall.Material.PP("textures/concrete/Polished-Concrete-Architextures.jpg", &vec3.T{0, 0, -cornellBoxUnit * 3 / 2}, vec3.T{0, 0, cornellBoxUnit * 3 * 1.3}, vec3.T{0, cornellBoxUnit, 0})
	rightWall.Material.PP("textures/concrete/rough_plaster_bright.png", &vec3.T{0, 0, -cornellBoxUnit * 3 / 2}, vec3.T{0, 0, cornellBoxUnit * 3 / 2}, vec3.T{0, cornellBoxUnit / 2, 0})

	lucy := obj.NewLucy(cornellBoxUnit * 0.8)

	//lucy := obj.NewTessellatedSphere(3, true)
	//lucy.Translate(&vec3.T{0.0, 1.0, 0.0})
	//lucy.Scale(&vec3.Zero, &vec3.T{0.35 * cornellBoxUnit / 2, 0.8 * cornellBoxUnit / 2, 0.35 * cornellBoxUnit / 2})

	lucy.Translate(&vec3.T{0, -cornellBoxUnit * 0.005, 0})
	v := vec3.T{0, cornellBoxUnit / 4, 0}
	u := vec3.T{1, 0, 0}
	lucy.Material = scn.NewMaterial().N("lucy").
		C(color.NewColorGrey(0.90)).
		CP("textures/marble/white_marble_double_width.png", &vec3.Zero, u, v, true)

	scene := scn.NewSceneNode().FS(lucy).FS(cornellBox)

	cameraOrigin := cornellBox.Bounds.Center().Add(&vec3.T{0, 0, -15 * (cornellBoxUnit / 3)})
	focusPoint := cornellBox.Bounds.Center()

	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)

	camera := scn.NewCamera(cameraOrigin, focusPoint, amountSamples, magnification).V(viewPlaneDistance).D(maxRayDepth).A(apertureSize, "")
	frame := scn.NewFrame(animationName, -1, camera, scene)
	animation.AddFrame(frame)

	anm.WriteAnimationToFile(animation, false)
}
