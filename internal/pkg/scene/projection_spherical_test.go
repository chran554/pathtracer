package scene

import (
	"fmt"
	"math"
	"testing"

	"github.com/ungerik/go3d/float64/vec2"
	"github.com/ungerik/go3d/float64/vec3"
)

func Test_SphericalProjection(t *testing.T) {

	projectionSetups := []struct {
		name   string
		origin vec3.T
		u, v   vec3.T
	}{
		{origin: vec3.Zero, u: vec3.T{1, 0, 0}, v: vec3.T{0, 1, 0}, name: "projection at world origin, parallel to world"},
		{origin: vec3.T{100, 0, 0}, u: vec3.T{1, 0, 0}, v: vec3.T{0, 1, 0}, name: "projection with positive X offset to world origin, parallel to world"},
		{origin: vec3.T{0, 100, 0}, u: vec3.T{1, 0, 0}, v: vec3.T{0, 1, 0}, name: "projection with positive Y offset to world origin, parallel to world"},
		{origin: vec3.T{0, 0, 100}, u: vec3.T{1, 0, 0}, v: vec3.T{0, 1, 0}, name: "projection with positive Z offset to world origin, parallel to world"},
		{origin: vec3.T{-100, 0, 0}, u: vec3.T{1, 0, 0}, v: vec3.T{0, 1, 0}, name: "projection with negative X offset to world origin, parallel to world"},
		{origin: vec3.T{0, -100, 0}, u: vec3.T{1, 0, 0}, v: vec3.T{0, 1, 0}, name: "projection with negative Y offset to world origin, parallel to world"},
		{origin: vec3.T{0, 0, -100}, u: vec3.T{1, 0, 0}, v: vec3.T{0, 1, 0}, name: "projection with negative Z offset to world origin, parallel to world"},
	}

	pointScaleSetups := []float64{0.5, 1.0, 100.0, 200.0}

	for _, ps := range pointScaleSetups {

		for _, projectionSetup := range projectionSetups {
			projName := projectionSetup.name
			projOrigin := projectionSetup.origin

			testSetups := []struct {
				name       string
				p          vec3.T
				expectedXY vec2.T
			}{
				// Extreme above and below
				{p: projOrigin.Added(&vec3.T{0, ps, 0}), expectedXY: vec2.T{0.0, 0.0}, name: fmt.Sprintf(projName+", point placed exactly above projection sphere (at distance %0.2f)", ps)},
				{p: projOrigin.Added(&vec3.T{0, -ps, 0}), expectedXY: vec2.T{0.0, 0.0}, name: fmt.Sprintf(projName+", point placed exactly below projection sphere (at distance %0.2f)", ps)},

				// Along equatorial line (latitude 90 deg)
				{p: projOrigin.Added(&vec3.T{ps, 0, 0}), expectedXY: vec2.T{0.0, 0.5}, name: fmt.Sprintf(projName+", point placed at longitude 0 deg along equatorial line (at distance %0.2f)", ps)},
				{p: projOrigin.Added(&vec3.T{0, 0, ps}), expectedXY: vec2.T{0.25, 0.5}, name: fmt.Sprintf(projName+", point placed at longitude 90 deg along equatorial line (at distance %0.2f)", ps)},
				{p: projOrigin.Added(&vec3.T{-ps, 0, 0}), expectedXY: vec2.T{0.5, 0.5}, name: fmt.Sprintf(projName+", point placed at longitude 180 deg along equatorial line (at distance %0.2f)", ps)},
				{p: projOrigin.Added(&vec3.T{0, 0, -ps}), expectedXY: vec2.T{0.75, 0.5}, name: fmt.Sprintf(projName+", point placed at longitude 270 deg along equatorial line (at distance %0.2f)", ps)},

				// Along latitude 45 deg
				{p: projOrigin.Added(&vec3.T{ps, ps, 0}), expectedXY: vec2.T{0.0, 0.25}, name: fmt.Sprintf(projName+", point exactly placed at longitude 0 deg, latitude 45 deg (at distance %0.2f)", ps)},
				{p: projOrigin.Added(&vec3.T{0, ps, ps}), expectedXY: vec2.T{0.25, 0.25}, name: fmt.Sprintf(projName+", point placed at longitude 90 deg, latitude 45 deg (at distance %0.2f)", ps)},
				{p: projOrigin.Added(&vec3.T{-ps, ps, 0}), expectedXY: vec2.T{0.5, 0.25}, name: fmt.Sprintf(projName+", point placed at longitude 180 deg, latitude 45 deg (at distance %0.2f)", ps)},
				{p: projOrigin.Added(&vec3.T{0, ps, -ps}), expectedXY: vec2.T{0.75, 0.25}, name: fmt.Sprintf(projName+", point placed at longitude 270 deg, latitude 45 deg (at distance %0.2f)", ps)},

				// Along latitude 135 deg
				{p: projOrigin.Added(&vec3.T{ps, -ps, 0}), expectedXY: vec2.T{0.0, 0.75}, name: fmt.Sprintf(projName+", point placed at longitude 0 deg, latitude 135 deg (at distance %0.2f)", ps)},
				{p: projOrigin.Added(&vec3.T{0, -ps, ps}), expectedXY: vec2.T{0.25, 0.75}, name: fmt.Sprintf(projName+", point placed at longitude 90 deg, latitude 135 deg (at distance %0.2f)", ps)},
				{p: projOrigin.Added(&vec3.T{-ps, -ps, 0}), expectedXY: vec2.T{0.5, 0.75}, name: fmt.Sprintf(projName+", point placed at longitude 180 deg, latitude 135 deg (at distance %0.2f)", ps)},
				{p: projOrigin.Added(&vec3.T{0, -ps, -ps}), expectedXY: vec2.T{0.75, 0.75}, name: fmt.Sprintf(projName+", point placed at longitude 270 deg, latitude 135 deg (at distance %0.2f)", ps)},
			}

			for _, testSetup := range testSetups {
				projection := NewSphericalImageProjection("", projectionSetup.origin, projectionSetup.u, projectionSetup.v)
				projection.Initialize()

				fmt.Println(testSetup.name)

				calculatedXY := projection.getSphericalXY(&testSetup.p)

				practicallyEquals := calculatedXY.PracticallyEquals(&testSetup.expectedXY, 0.0000001)

				if !practicallyEquals {
					fmt.Println()
					fmt.Printf("Projection origin:  %+v\n", projection.Origin)
					fmt.Printf("Projection u and v: %+v  %+v\n", projection.U, projection.V)
					fmt.Printf("Point:              %+v\n", testSetup.p)
					fmt.Printf("Expected XY:        %+v\n", testSetup.expectedXY)
					fmt.Printf("Actual XY:          %+v\n", calculatedXY)

					t.Errorf("spherical xy expected to be %+v but was %+v for test \"%s\".", testSetup.expectedXY, calculatedXY, testSetup.name)
				}
			}
		}
	}

}

func Test_SphericalProjection2(t *testing.T) {
	type testSetup struct {
		name       string
		p          vec3.T
		expectedXY vec2.T
	}

	projectionSetups := []struct {
		name   string
		origin vec3.T
		u, v   vec3.T
	}{
		{origin: vec3.Zero, u: vec3.T{1, 0, 0}, v: vec3.T{0, 1, 0}, name: "projection at world origin, parallel to world"},
	}

	pointScaleSetups := []float64{100.0}

	for _, ps := range pointScaleSetups {

		for _, projectionSetup := range projectionSetups {
			// projName := projectionSetup.name
			// projOrigin := projectionSetup.origin

			var testSetups []testSetup

			radPerDeg := math.Pi / 180.0
			for degree := 0; degree < 180; degree += 5 {
				p := vec3.T{}
				p[0] = ps * math.Sin(float64(degree)*radPerDeg)
				p[1] = ps * math.Cos(float64(degree)*radPerDeg)
				p[2] = 0.0

				testSetups = append(testSetups, testSetup{
					name:       fmt.Sprintf("Degree %d", degree),
					p:          p,
					expectedXY: vec2.T{0.0, float64(degree) / 180.0},
				})
			}

			testSetups = append(testSetups, testSetup{
				name:       fmt.Sprintf("Degree %d", 180),
				p:          vec3.T{0.0, ps, 0.0},
				expectedXY: vec2.T{0.0, 0.0}, // Extreme cases at top and bottom are treated as [x:0, y:0]
			})

			for _, testSetup := range testSetups {
				projection := NewSphericalImageProjection("", projectionSetup.origin, projectionSetup.u, projectionSetup.v)
				projection.Initialize()

				fmt.Println(testSetup.name)

				calculatedXY := projection.getSphericalXY(&testSetup.p)

				practicallyEquals := calculatedXY.PracticallyEquals(&testSetup.expectedXY, 0.0000001)

				if !practicallyEquals {
					fmt.Println()
					fmt.Printf("Projection origin:  %+v\n", projection.Origin)
					fmt.Printf("Projection u and v: %+v  %+v\n", projection.U, projection.V)
					fmt.Printf("Point:              %+v\n", testSetup.p)
					fmt.Printf("Expected XY:        %+v\n", testSetup.expectedXY)
					fmt.Printf("Actual XY:          %+v\n", calculatedXY)

					t.Errorf("spherical xy expected to be %+v but was %+v for test \"%s\".", testSetup.expectedXY, calculatedXY, testSetup.name)
				}
			}
		}
	}

}
