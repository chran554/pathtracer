package scene

import (
	"math"
	"pathtracer/internal/pkg/color"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaterialAttenuate(t *testing.T) {
	material := NewMaterial().Attenuate(color.NewColorRGBA(0.5, 0.25, 1, 0), 2)
	assert.InDelta(t, math.Ln2/2, material.Absorption.R, 1e-7)
	assert.InDelta(t, math.Ln2, material.Absorption.G, 1e-7)
	assert.Zero(t, material.Absorption.B)
	assert.Equal(t, color.White, material.Color)
	assert.Zero(t, material.Transparency)
	assert.Nil(t, material.Absorb(nil).Absorption)
}

func TestMaterialAbsorptionInvalidInputs(t *testing.T) {
	for _, coefficient := range []float64{-1, math.NaN(), math.Inf(1)} {
		assert.Panics(t, func() { NewMaterial().Absorb(color.NewColorGrey(coefficient)) })
	}
	for _, distance := range []float64{0, -1, math.NaN(), math.Inf(1)} {
		assert.Panics(t, func() { NewMaterial().Attenuate(color.White, distance) })
	}
	for _, transmission := range []float64{-1, 0, 1.1, math.NaN(), math.Inf(1)} {
		assert.Panics(t, func() { NewMaterial().Attenuate(color.NewColorGrey(transmission), 1) })
	}
	assert.NotPanics(t, func() { NewMaterial().Attenuate(nil, 1) })
}

func TestMetallicMaterialParameters(t *testing.T) {
	material := NewMaterial()
	assert.Zero(t, material.Metallicness)
	assert.Equal(t, 1.0, material.Roughness)
	assert.Equal(t, 1.5, material.RefractionIndex)
	assert.Zero(t, material.Transparency)
	assert.Equal(t, NewMaterial().M(0.4, 0.2).T(0.7, true, 1.3), NewMaterial().T(0.7, true, 1.3).M(0.4, 0.2))

	glass := NewMaterialSolidGlass("glass")
	assert.Zero(t, glass.Metallicness)
	assert.Equal(t, 1.0, glass.Transparency)
	assert.True(t, glass.SolidObject)

	for _, value := range []float64{-1, 1.1, math.NaN(), math.Inf(1)} {
		assert.Panics(t, func() { NewMaterial().M(value, 0) })
		assert.Panics(t, func() { NewMaterial().M(0, value) })
		assert.Panics(t, func() { NewMaterial().T(value, false, 1.5) })
	}
}
