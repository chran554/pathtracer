package main

import (
	anm "pathtracer/internal/pkg/animation"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var ballRadius float64 = 30

// var cameraOrigin = vec3.T{0, 0, -200} // Looking straight on
var cameraOrigin = vec3.T{0, 200, -200} // Looking straight on, slightly from above
//var cameraOrigin = vec3.T{0, -200, -200} // Looking straight on, slightly from below

var amountSamples = 256
var lensRadius float64 = 0
var viewPlaneDistance = 1600.0
var magnification = 1.5

func main() {
	animation := scn.NewAnimation("spherical_projection", 1200, 600, magnification, false, false)

	sphere2Origin := vec3.T{0, 0, 0}
	sphere1Origin := sphere2Origin.Added(&vec3.T{-ballRadius * 2.2, 0, 0})
	sphere3Origin := sphere2Origin.Added(&vec3.T{ballRadius * 2.2, 0, 0})

	projection1Origin := sphere1Origin
	projection2Origin := sphere2Origin
	projection3Origin := sphere3Origin

	projectionU := vec3.T{0, 0, -ballRadius}
	projectionV := vec3.T{0, ballRadius, 0}

	projection1 := scn.NewSphericalImageProjection("textures/planets/earth_daymap.jpg", &projection1Origin, projectionU.Inverted(), projectionV)
	projection2 := scn.NewSphericalImageProjection("textures/checkered 360x180 with lines.png", &projection2Origin, projectionU, projectionV)
	projection3 := scn.NewSphericalImageProjection("textures/equirectangular/2560px-Plate_Carrée_with_Tissot's_Indicatrices_of_Distortion.svg.png", &projection3Origin, projectionU.Inverted(), projectionV)

	sphere1 := scn.NewSphere(&sphere1Origin, ballRadius, scn.NewMaterial().P(&projection1)).N("Textured sphere - Earth")
	sphere2 := scn.NewSphere(&sphere2Origin, ballRadius, scn.NewMaterial().P(&projection2)).N("Textured sphere - checkered")
	sphere3 := scn.NewSphere(&sphere3Origin, ballRadius, scn.NewMaterial().P(&projection3)).N("Textured sphere - Tissot's_Indicatrices_of_Distortion")

	scene := scn.NewSceneNode().S(sphere1, sphere2, sphere3)

	camera := scn.NewCamera(&cameraOrigin, &vec3.T{0, 0, 0}, amountSamples, magnification).V(viewPlaneDistance).A(lensRadius, "")
	frame := scn.NewFrame(animation.AnimationName, -1, camera, scene)
	animation.AddFrame(frame)

	anm.WriteAnimationToFile(animation, false)
}
