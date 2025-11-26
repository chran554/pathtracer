package obj

import (
	"path/filepath"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

func NewPixarBall(pixarBallOrigin *vec3.T, pixarBallRadius float64) *scene.Sphere {
	textureOrigin := pixarBallOrigin.Added(&vec3.T{-pixarBallRadius, -pixarBallRadius, 0})
	material := scene.NewMaterial().N("pixar ball").
		PP(floatimage.Load(filepath.Join(TexturesDir, "pixar_ball_02.png")), &textureOrigin, vec3.UnitX.Scaled(pixarBallRadius*2), vec3.UnitY.Scaled(pixarBallRadius*2))

	return scene.NewSphere(pixarBallOrigin, pixarBallRadius, material).N("pixar ball")
}
