package main

import (
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/scene"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ungerik/go3d/float64/vec3"
)

func TestCameraMediaFirstSegment(t *testing.T) {
	for _, absorbing := range []bool{false, true} {
		frame := &scene.Frame{
			Camera: &scene.Camera{RenderType: scene.Pathtracing},
			SceneNode: &scene.SceneNode{FacetStructures: []*scene.FacetStructure{
				orientationTestPlane(scene.NewMaterial().E(color.White, 1, true)),
			}},
		}
		want := color.White
		if absorbing {
			frame.WorldMedium = scene.NewMaterial().T(1, true, scene.RefractionIndex_Water).
				Absorb(color.NewColor(math.Ln2, 0, 2*math.Ln2))
			want = color.NewColor(0.25, 1, 0.0625)
		}
		initial, err := frame.InitialMediumStack()
		require.NoError(t, err)
		ray := &scene.Ray{Origin: &vec3.T{0, 0, 2}, Heading: &vec3.T{0, 0, -1}}
		assertTransmissionColor(t, traceCameraPath(ray, frame.Camera, frame.SceneNode, initial), want)
	}
}

// Analytic Snell angles aim at a small light. Absorption uses the known plane
// spacing, and the independent scalar Fresnel reference predicts mean radiance.
func TestCameraMediaRefraction(t *testing.T) {
	for _, name := range []string{"glass to air", "glass through water to air", "global water through glass"} {
		t.Run(name, func(t *testing.T) {
			glass := scene.NewMaterial().C(color.WhiteTransparent).M(0, 0).T(1, true, 1.5).Absorb(color.NewColor(0.1, 0.3, 0.7))
			water := scene.NewMaterial().C(color.WhiteTransparent).M(0, 0).T(1, true, scene.RefractionIndex_Water).Absorb(color.NewColor(0.08, 0.02, 0.01))
			air := scene.NewMaterial().T(1, true, scene.RefractionIndex_Air)
			camera := &scene.Camera{RenderType: scene.Pathtracing, InitialMedia: []*scene.Material{glass}}
			frame := &scene.Frame{Camera: camera, SceneNode: &scene.SceneNode{}}
			media := []*scene.Material{glass, air}
			boundaries := []*scene.Material{glass}
			exiting := []bool{true}
			if name == "glass through water to air" {
				camera.InitialMedia = []*scene.Material{water, glass}
				media = []*scene.Material{glass, water, air}
				boundaries, exiting = []*scene.Material{glass, water}, []bool{true, true}
			} else if name == "global water through glass" {
				frame.WorldMedium, camera.InitialMedia = water, nil
				media = []*scene.Material{water, glass, water}
				boundaries, exiting = []*scene.Material{glass, glass}, []bool{false, true}
			}
			camera.RecursionDepth = len(boundaries)
			initial, err := frame.InitialMediumStack()
			require.NoError(t, err)

			angle := 20.0
			heading, _ := dielectricTestDirections(angle, 0)
			position := vec3.T{0, 0, 1}
			opticalDepth := [3]float64{}
			transmissionProbability := 1.0
			for segment, medium := range media {
				direction, _ := dielectricTestDirections(angle, 0)
				length := 1 / -direction[2]
				step := direction.Scaled(length)
				position.Add(&step)
				if segment > 0 {
					length *= 1 - epsilonDistance // The tracing offset skips this z distance.
				}
				if medium.Absorption != nil {
					for channel, coefficient := range [3]float32{medium.Absorption.R, medium.Absorption.G, medium.Absorption.B} {
						opticalDepth[channel] += float64(coefficient) * length
					}
				}
				if segment == len(boundaries) {
					break
				}
				boundary := boundaries[segment]
				plane := orientationTestPlane(boundary)
				if exiting[segment] {
					plane.Facets[0].ChangeWindingOrder()
					plane.UpdateNormals()
				}
				plane.Translate(&vec3.T{0, 0, -float64(segment)})
				frame.SceneNode.FacetStructures = append(frame.SceneNode.FacetStructures, plane)
				refracted, fresnel := dielectricTestReference(angle, 0, medium.RefractionIndex, media[segment+1].RefractionIndex)
				angle = math.Atan2(refracted[0], -refracted[2]) * 180 / math.Pi
				transmissionProbability *= 1 - fresnel
			}
			frame.SceneNode.Discs = []*scene.Disc{scene.NewDisc(&position, &vec3.UnitZ, 0.01, scene.NewMaterial().E(color.White, 1, true))}
			ratio := media[0].RefractionIndex / media[len(media)-1].RefractionIndex
			want := color.NewColor(ratio*ratio*math.Exp(-opticalDepth[0]), ratio*ratio*math.Exp(-opticalDepth[1]), ratio*ratio*math.Exp(-opticalDepth[2]))
			ray := &scene.Ray{Origin: &vec3.T{0, 0, 1}, Heading: &heading}
			assertTransmissionColor(t, traceTransmissionPath(ray, camera, frame.SceneNode, 0, append([]*scene.Material(nil), initial...)), want)

			// With Fresnel enabled, only paths transmitting at every boundary
			// reach the light within this depth limit. Reflections contribute zero.
			const samples = 12000
			var total [3]float64
			for range samples {
				radiance := traceCameraPath(ray, camera, frame.SceneNode, initial)
				for channel, value := range [3]float32{radiance.R, radiance.G, radiance.B} {
					total[channel] += float64(value)
				}
			}
			for channel, value := range [3]float32{want.R, want.G, want.B} {
				tolerance := 8 * float64(value) * math.Sqrt(transmissionProbability*(1-transmissionProbability)/samples)
				assert.InDelta(t, float64(value)*transmissionProbability, total[channel]/samples, tolerance+1e-6)
			}
		})
	}
}

func TestCameraMediaTotalInternalReflection(t *testing.T) {
	glass := scene.NewMaterial().C(color.WhiteTransparent).M(0, 0).T(1, true, 1.5).Absorb(color.NewColorGrey(0.2))
	frame := &scene.Frame{
		Camera: &scene.Camera{RenderType: scene.Pathtracing, RecursionDepth: 1, InitialMedia: []*scene.Material{glass}},
		SceneNode: &scene.SceneNode{
			FacetStructures: []*scene.FacetStructure{orientationTestPlane(glass)},
			Discs: []*scene.Disc{scene.NewDisc(&vec3.T{2 * math.Sqrt(3), 0, -1}, &vec3.UnitZ, 0.01,
				scene.NewMaterial().E(color.White, 1, true))},
		},
	}
	initial, err := frame.InitialMediumStack()
	require.NoError(t, err)
	ray := &scene.Ray{Origin: &vec3.T{0, 0, -1}, Heading: &vec3.T{math.Sqrt(0.75), 0, 0.5}}
	want := color.NewColorGrey(math.Exp(-float64(glass.Absorption.R) * (4 - 2*epsilonDistance)))
	assertTransmissionColor(t, traceCameraPath(ray, frame.Camera, frame.SceneNode, initial), want)
	assert.Same(t, glass, initial[1])
}

func TestCameraMediaStackIsolation(t *testing.T) {
	water := scene.NewMaterial().C(color.WhiteTransparent).T(1, true, 1).Absorb(color.NewColor(math.Ln2, 0, 0))
	glass := scene.NewMaterial().C(color.WhiteTransparent).T(1, true, 1).Absorb(color.NewColor(0, math.Ln2, 0))
	frame := &scene.Frame{
		WorldMedium: scene.NewMaterial().T(1, true, 1),
		Camera:      &scene.Camera{RenderType: scene.Pathtracing, RecursionDepth: 3, InitialMedia: []*scene.Material{water}},
		SceneNode: &scene.SceneNode{
			Spheres: []*scene.Sphere{scene.NewSphere(&vec3.T{}, 1, water), scene.NewSphere(&vec3.T{0, 0, 4}, 1, glass)},
			Discs:   []*scene.Disc{scene.NewDisc(&vec3.T{0, 0, 6}, &vec3.UnitZ, 1, scene.NewMaterial().E(color.White, 1, true))},
		},
	}
	initial, err := frame.InitialMediumStack()
	require.NoError(t, err)
	want := color.NewColor(0.5, math.Exp(-float64(glass.Absorption.G)*(2-epsilonDistance)), 1)
	trace := func() *color.Color {
		// Exit water, enter and exit glass: the push reuses a slot formerly
		// occupied by water. Sharing the initial slice would overwrite it.
		return traceCameraPath(&scene.Ray{Origin: &vec3.T{}, Heading: &vec3.T{0, 0, 1}}, frame.Camera, frame.SceneNode, initial)
	}
	for range 2 {
		assertTransmissionColor(t, trace(), want)
		require.Same(t, water, initial[1])
	}
	const workers, samples = 16, 50
	results := make([]*color.Color, workers*samples)
	var wg sync.WaitGroup
	for worker := range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for sample := range samples {
				results[worker*samples+sample] = trace()
			}
		}()
	}
	wg.Wait()
	for _, result := range results {
		assertTransmissionColor(t, result, want)
	}
	assert.Same(t, water, initial[1])
	assert.Same(t, water, frame.Camera.InitialMedia[0])
}

func TestRenderRejectsInvalidMediaBeforeWorkers(t *testing.T) {
	frame := &scene.Frame{Filename: "bad media", Camera: &scene.Camera{InitialMedia: []*scene.Material{nil}}}
	// Nil output/monitor are safe because invalid settings fail before rendering.
	require.ErrorContains(t, render(frame, 2, 2, nil, nil), "camera initial medium 0")
}
