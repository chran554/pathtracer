package obj

import (
	"path/filepath"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

func NewPixarBall(pixarBallOrigin *vec3.T, pixarBallRadius float64) *scene.Sphere {
	texturePixarBall := floatimage.LoadOrPanic(filepath.Join(TexturesDir, "pixar_ball_02.png"))

	textureOrigin := pixarBallOrigin.Added(&vec3.T{-pixarBallRadius, -pixarBallRadius, 0})
	material := scene.NewMaterial().N("pixar ball").PP(texturePixarBall, &textureOrigin, vec3.UnitX.Scaled(pixarBallRadius*2), vec3.UnitY.Scaled(pixarBallRadius*2))

	return scene.NewSphere(pixarBallOrigin, pixarBallRadius, material).N("pixar ball")
}
