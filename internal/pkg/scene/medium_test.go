package scene

import (
	"math"
	"pathtracer/internal/pkg/color"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitialMediumStack(t *testing.T) {
	frame := &Frame{Camera: &Camera{}}
	stack, err := frame.InitialMediumStack()
	require.NoError(t, err)
	require.Len(t, stack, 1)
	assert.Equal(t, RefractionIndex_Air, stack[0].RefractionIndex)
	assert.Nil(t, stack[0].Absorption)
	assert.Nil(t, frame.WorldMedium)
	assert.Nil(t, frame.Camera.InitialMedia)

	world := NewMaterial().Absorb(color.NewColor(0.01, 0.02, 0.03))
	water := NewMaterial().T(1, true, RefractionIndex_Water)
	glass := NewMaterial().T(1, true, RefractionIndex_Glass)
	frame.WorldMedium = world
	frame.Camera.InitialMedia = []*Material{water, glass}
	stack, err = frame.InitialMediumStack()
	require.NoError(t, err)
	require.Len(t, stack, 3)
	assert.Same(t, world, stack[0])
	assert.Same(t, water, stack[1])
	assert.Same(t, glass, stack[2])
	stack[1] = glass
	assert.Same(t, water, frame.Camera.InitialMedia[0], "building a stack must not alias camera slice storage")

	frame.Camera.InitialMedia = nil
	stack, err = frame.InitialMediumStack()
	require.NoError(t, err)
	require.Len(t, stack, 1)
	assert.Same(t, world, stack[0])
}

func TestInitialMediumStackValidation(t *testing.T) {
	_, err := (&Frame{}).InitialMediumStack()
	require.ErrorContains(t, err, "camera")
	_, err = (&Frame{Camera: &Camera{InitialMedia: []*Material{nil}}}).InitialMediumStack()
	require.ErrorContains(t, err, "camera initial medium 0")

	for _, value := range []float64{-1, 0, math.NaN(), math.Inf(1), math.Inf(-1)} {
		medium := NewMaterial()
		medium.RefractionIndex = value
		for _, world := range []bool{false, true} {
			frame := &Frame{Camera: &Camera{}}
			if world {
				frame.WorldMedium = medium
			} else {
				frame.Camera.InitialMedia = []*Material{medium}
			}
			_, err := frame.InitialMediumStack()
			require.ErrorContains(t, err, "refraction index")
		}
	}
	for channel := range 3 {
		for _, value := range []float32{-1, float32(math.NaN()), float32(math.Inf(1))} {
			coefficients := [3]float32{}
			coefficients[channel] = value
			medium := NewMaterial()
			// Direct assignment and archive decoding also need validation.
			medium.Absorption = &color.Color{R: coefficients[0], G: coefficients[1], B: coefficients[2]}
			_, err := (&Frame{WorldMedium: medium, Camera: &Camera{}}).InitialMediumStack()
			require.ErrorContains(t, err, "absorption channel")
		}
	}
}
