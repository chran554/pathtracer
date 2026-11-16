package scene

import (
	"fmt"
	"math"
)

// InitialMediumStack builds and validates the media at the camera's origin.
// The world medium is the bottom entry and is never popped during transport.
// Material pointers are shared; callers must keep materials read-only during
// rendering and copy the returned slice for each independent camera path.
func (f *Frame) InitialMediumStack() ([]*Material, error) {
	if f.Camera == nil {
		return nil, fmt.Errorf("initial media require a camera")
	}

	world := f.WorldMedium
	if world == nil {
		world = NewMaterial().N("default world medium").T(1, true, RefractionIndex_Air)
	}

	if err := validateMedium(world); err != nil {
		return nil, fmt.Errorf("world medium: %w", err)
	}

	stack := make([]*Material, 1, 1+len(f.Camera.InitialMedia))
	stack[0] = world
	for i, medium := range f.Camera.InitialMedia {
		if err := validateMedium(medium); err != nil {
			return nil, fmt.Errorf("camera initial medium %d: %w", i, err)
		}
		stack = append(stack, medium)
	}

	return stack, nil
}

func validateMedium(medium *Material) error {
	if medium == nil {
		return fmt.Errorf("material must not be nil")
	}

	if !(medium.RefractionIndex > 0) || math.IsInf(medium.RefractionIndex, 0) {
		return fmt.Errorf("refraction index must be finite and positive")
	}

	if medium.Absorption != nil {
		for i, coefficient := range []float32{medium.Absorption.R, medium.Absorption.G, medium.Absorption.B} {
			if !(coefficient >= 0) || math.IsInf(float64(coefficient), 0) {
				return fmt.Errorf("absorption channel %d must be finite and nonnegative", i)
			}
		}
	}

	return nil
}
