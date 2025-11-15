package main

import (
	"fmt"
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	anm "pathtracer/internal/pkg/renderfile"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "projection_parallel"

var discRadius float64 = 30

var magnification = 2.0 * 3

var amountSamples = 1024 * 20

func main() {
	skyDome := scn.NewSphere(&vec3.T{0, 0, 0}, 1000, scn.NewMaterial().
		E(color.White, 1.0, true).
		//C(color.NewColorGrey(0.2))).
		SP(floatimage.Load("textures/equirectangular/open_grassfield_sunny_day.jpg"), &vec3.T{0, 0, 0}, vec3.T{1, 0, 0}, vec3.T{0, 1, 0})).N("sky dome")

	discOrigin := vec3.T{0, 0, 0}
	projectionOrigin := discOrigin

	projectionU := vec3.T{discRadius / 2.0, 0, 0}
	projectionV := vec3.T{0, 0, discRadius / 2.0}
	projection := scn.NewParallelImageProjection(floatimage.Load("textures/test/uv.png"), &projectionOrigin, projectionU, projectionV)

	texturedDisc := scn.NewDisc(&discOrigin, &vec3.T{0, 1, 0}, discRadius, scn.NewMaterial().P(&projection))

	sphereRadius := 5.0

	sphere1Origin := vec3.From(&projectionU)
	sphere1Origin.Add(&vec3.T{+sphereRadius / 3, sphereRadius, sphereRadius / 2})
	sphere1 := scn.NewSphere(&sphere1Origin, sphereRadius, scn.NewMaterial().P(&projection))

	sphere2Origin := vec3.From(&projectionV)
	sphere2Origin.Add(&vec3.T{0, sphereRadius, 0})
	checkeredMaterial := scn.NewMaterial().PP(floatimage.Load("textures/floor/7451-diffuse 02.png"), &sphere2Origin, *(&vec3.T{1, 0, -1}).Scale(2.0), *(&vec3.T{1, 1, 0}).Scale(2.0))
	sphere2 := scn.NewSphere(&sphere2Origin, sphereRadius, checkeredMaterial).N("checkered sphere")

	sphere3Origin := vec3.T{-discRadius / 2, sphereRadius, -discRadius / 2}
	treeSphereProjectionOrigin := vec3.From(&sphere3Origin)
	treeSphereMaterial := scn.NewMaterial().T(0.0, true, scn.RefractionIndex_WoodSap).M(0.0, 0.0).PP(floatimage.Load("textures/tree/tree_rings_03.jpg"), (&treeSphereProjectionOrigin).Add(&vec3.T{-sphereRadius, 0, -sphereRadius}), *(&vec3.T{1, 0, 0}).Scale(sphereRadius * 2), *(&vec3.T{0, 0, 1}).Scale(sphereRadius * 2))
	sphere3 := scn.NewSphere(&sphere3Origin, sphereRadius, treeSphereMaterial).N("tree sphere")
	sphere3.RotateX(sphere3.Bounds().Center(), math.Pi/180*22.5)
	sphere3.RotateY(sphere3.Bounds().Center(), math.Pi/180*45)

	lamp := scn.NewSphere(&vec3.T{0, 250, 0}, 150.0, scn.NewMaterial().E(color.White, 4.0, true))

	origin := &vec3.T{0, 150, -200}
	focusPoint := &vec3.T{0, 0, 0}
	camera := scn.NewCamera(origin, focusPoint, amountSamples, magnification).A(1.0, nil)

	scene := scn.NewSceneNode().
		D(texturedDisc).
		S(lamp, sphere1, sphere2, sphere3, skyDome)

	frame := scn.NewFrame(animationName, -1, camera, scene)

	animation := scn.NewAnimation(animationName, 200, 150, magnification, false, false)
	animation.AddFrame(frame)

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := anm.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}
