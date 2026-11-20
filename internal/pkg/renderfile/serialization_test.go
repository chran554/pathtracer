package renderfile

import (
	"archive/zip"
	"os"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/scene"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ungerik/go3d/float64/vec3"
	"github.com/vmihailenco/msgpack/v5"
)

func TestAbsorptionSerialization(t *testing.T) {
	for _, tc := range []struct {
		name       string
		absorption *color.Color
	}{
		{"legacy no absorption", nil},
		{"zero absorption", color.Black},
		{"colored absorption above one", color.NewColor(0.1, 2, 3)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			material := scene.NewMaterial().T(1, true, 1.5).Absorb(tc.absorption)
			node := scene.NewSceneNode()
			node.Spheres = []*scene.Sphere{scene.NewSphere(&vec3.T{}, 1, material)}
			camera := scene.NewCamera(&vec3.T{0, 0, 3}, &vec3.T{}, 1, 1)
			animation := scene.NewAnimation("absorption", 2, 2, 1, false, false).
				AddFrame(scene.NewFrame("absorption", 0, camera, node))
			filename := filepath.Join(t.TempDir(), "absorption.render.zip")
			require.NoError(t, WriteRenderFile(filename, animation))
			restored, err := ReadRenderFile(filename)
			require.NoError(t, err)
			assert.Equal(t, material, restored.Frames[0].SceneNode.Spheres[0].Material)
		})
	}
}

func TestAnimationSerialization(t *testing.T) {
	// Create a simple animation
	cam := scene.NewCamera(&vec3.T{0, 0, 0}, &vec3.T{0, 0, -1}, 1, 1.0)
	node := scene.NewSceneNode()

	// Create a material with a projection and texture
	mat := scene.NewMaterial().N("TestMaterial").C(color.NewColor(1, 0, 0))
	proj := scene.NewParallelImageProjection(nil, scene.TextureTypeAlbedo, &vec3.T{0, 0, 0}, vec3.T{1, 0, 0}, vec3.T{0, 1, 0})
	mat.P(&proj)

	sphere := &scene.Sphere{
		Name:     "TestSphere",
		Origin:   &vec3.T{0, 0, -5},
		Radius:   1.0,
		Material: mat,
	}
	node.Spheres = append(node.Spheres, sphere)

	frame := scene.NewFrame("test_frame", 0, cam, node)
	anim := scene.NewAnimation("TestAnimation", 100, 100, 1.0, false, false)
	anim.AddFrame(frame)

	filename := "test_animation.render.zip"
	defer os.Remove(filename)

	// Serialize
	err := WriteRenderFile(filename, anim)
	assert.NoError(t, err)

	// Deserialize
	deserializedAnim, err := ReadRenderFile(filename)
	assert.NoError(t, err)
	assert.NotNil(t, deserializedAnim)
	assert.Equal(t, anim.AnimationName, deserializedAnim.AnimationName)
	assert.Equal(t, len(anim.Frames), len(deserializedAnim.Frames))

	// Check material and projection
	if len(deserializedAnim.Frames) > 0 {
		df := deserializedAnim.Frames[0]
		assert.Equal(t, 1, len(df.SceneNode.Spheres))
		ds := df.SceneNode.Spheres[0]
		assert.NotNil(t, ds.Material)
		assert.Equal(t, "TestMaterial", ds.Material.Name)
		assert.NotNil(t, ds.Material.Projections)
		assert.Equal(t, string(scene.ProjectionTypeParallel), string(ds.Material.Projections[0].Type))
	}
}

func TestProjectionTextureSerialization(t *testing.T) {
	// This test specifically checks the new TextureIndex slice in Projection
	// and the Projection slice in Material.

	tempFile := "test_proj.zip"
	defer os.Remove(tempFile)

	zipFile, _ := os.Create(tempFile)
	zw := zip.NewWriter(zipFile)
	s := newSerializer(zw)

	// Create a scene texture
	tex := &scene.Texture{
		Type:     scene.TextureTypeAlbedo,
		Strength: 0.8,
	}

	proj := &scene.ImageProjection{
		Type:    scene.ProjectionTypeParallel,
		Texture: tex,
		Origin:  &vec3.T{1, 2, 3},
		U:       &vec3.T{1, 0, 0},
		V:       &vec3.T{0, 1, 0},
	}

	// 1. Test serializeProjection
	serializedProj, err := s.serializeProjection(proj)
	assert.NoError(t, err)
	assert.NotNil(t, serializedProj)
	assert.Equal(t, 1, len(serializedProj.Texture))

	// 2. Test deserializeProjection
	// We need to simulate the serializer state (s.st, s.sv)
	s.st = append(s.st, tex)
	s.sv = append(s.sv, proj.Origin)
	s.sv = append(s.sv, proj.U)
	s.sv = append(s.sv, proj.V)

	deserializedProj, err := s.deserializeProjection(serializedProj)
	assert.NoError(t, err)
	assert.NotNil(t, deserializedProj)
	assert.NotNil(t, deserializedProj.Texture)
	assert.Equal(t, tex.Strength, deserializedProj.Texture.Strength)
	assert.Equal(t, proj.Origin, deserializedProj.Origin)

	zw.Close()
	zipFile.Close()
}

func TestMetallicRoughnessSerialization(t *testing.T) {
	for _, metallic := range []float64{0, 0.4, 1} {
		material := scene.NewMaterial().N("metallic").C(color.NewColor(0.8, 0.4, 0.1)).M(metallic, 0.35).T(0.6, true, 1.6)
		node := scene.NewSceneNode().S(scene.NewSphere(&vec3.T{}, 1, material))
		camera := scene.NewCamera(&vec3.T{0, 0, 3}, &vec3.T{}, 1, 1)
		animation := scene.NewAnimation("metallic", 2, 2, 1, false, false).AddFrame(scene.NewFrame("metallic", 0, camera, node))
		filename := filepath.Join(t.TempDir(), "metallic.render.zip")
		require.NoError(t, WriteRenderFile(filename, animation))
		restored, err := ReadRenderFile(filename)
		require.NoError(t, err)
		assert.Equal(t, material, restored.Frames[0].SceneNode.Spheres[0].Material)
	}
	// Check wire names explicitly, including removal of the obsolete parameters.
	data, err := msgpack.Marshal(Material{Metallicness: 0.4, Roughness: 0.35})
	require.NoError(t, err)
	var fields map[string]interface{}
	require.NoError(t, msgpack.Unmarshal(data, &fields))
	assert.Equal(t, 0.4, fields["metallicness"])
	for _, key := range []string{"glossiness", "diffuse", "colorize-reflection", "fresnel-max-reflection"} {
		assert.NotContains(t, fields, key)
	}
}
