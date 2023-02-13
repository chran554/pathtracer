package obj

import (
	"github.com/ungerik/go3d/float64/vec3"
	scn "pathtracer/internal/pkg/scene"
)

func NewPixarBall(pixarBallOrigin *vec3.T, pixarBallRadius float64) *scn.Sphere {
	pixarBall := &scn.Sphere{
		Name:   "pixar ball",
		Origin: pixarBallOrigin,
		Radius: pixarBallRadius,
		Material: scn.NewMaterial().N("pixar ball").PP("textures/pixar_ball_02.png",
			pixarBallOrigin.Added(&vec3.T{-pixarBallRadius, -pixarBallRadius, 0}),
			vec3.UnitX.Scaled(pixarBallRadius*2),
			vec3.UnitY.Scaled(pixarBallRadius*2)),
	}
	return pixarBall
}
