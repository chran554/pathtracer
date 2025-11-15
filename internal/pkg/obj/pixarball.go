package obj

import (
	"path/filepath"
	"pathtracer/internal/pkg/floatimage"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

func NewPixarBall(pixarBallOrigin *vec3.T, pixarBallRadius float64) *scn.Sphere {
	textureOrigin := pixarBallOrigin.Added(&vec3.T{-pixarBallRadius, -pixarBallRadius, 0})
	material := scn.NewMaterial().N("pixar ball").
		PP(floatimage.Load(filepath.Join(TexturesDir, "pixar_ball_02.png")), &textureOrigin, vec3.UnitX.Scaled(pixarBallRadius*2), vec3.UnitY.Scaled(pixarBallRadius*2))

	return scn.NewSphere(pixarBallOrigin, pixarBallRadius, material).N("pixar ball")
}
