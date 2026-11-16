package renderfile

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/scene"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ungerik/go3d/float64/vec3"
	"github.com/vmihailenco/msgpack/v5"
)

func mediumArchiveTestFrame() *scene.Frame {
	world := scene.NewMaterial().N("world").T(1, true, 1.1).Absorb(color.NewColor(0.01, 0.02, 0.03))
	water := scene.NewMaterial().N("water").T(1, true, scene.RefractionIndex_Water).Absorb(color.NewColor(0.1, 0.2, 0.3))
	glass := scene.NewMaterial().N("glass").T(1, true, scene.RefractionIndex_Glass).Absorb(color.NewColor(0.7, 0.2, 0))
	camera := scene.NewCamera(&vec3.T{}, &vec3.T{0, 0, 1}, 1, 1)
	camera.InitialMedia = []*scene.Material{water, glass}
	frame := scene.NewFrame("media", 0, camera, &scene.SceneNode{
		Spheres: []*scene.Sphere{scene.NewSphere(&vec3.T{}, 1, glass)},
	})
	frame.WorldMedium = world
	return frame
}

func writeMediumTestArchive(t *testing.T, frames ...*scene.Frame) string {
	t.Helper()
	animation := scene.NewAnimation("media", 2, 2, 1, false, false)
	animation.Frames = frames
	filename := filepath.Join(t.TempDir(), "media.render.zip")
	require.NoError(t, WriteRenderFile(filename, animation))
	return filename
}

func TestInitialMediaSerialization(t *testing.T) {
	for _, shared := range []bool{false, true} {
		frame := mediumArchiveTestFrame()
		if shared {
			// Otherwise world and water are referenced only by the new settings.
			frame.SceneNode.Spheres = append(frame.SceneNode.Spheres,
				scene.NewSphere(&vec3.T{}, 2, frame.Camera.InitialMedia[0]),
				scene.NewSphere(&vec3.T{}, 3, frame.WorldMedium))
		}
		defaults := scene.NewFrame("default", 1,
			scene.NewCamera(&vec3.T{}, &vec3.T{0, 0, 1}, 1, 1), scene.NewSceneNode())
		filename := writeMediumTestArchive(t, frame, defaults)
		restored, err := ReadRenderFile(filename)
		require.NoError(t, err)
		require.Len(t, restored.Frames, 2)
		result := restored.Frames[0]
		assert.Equal(t, frame.WorldMedium, result.WorldMedium)
		require.Equal(t, frame.Camera.InitialMedia, result.Camera.InitialMedia)
		assert.Same(t, result.SceneNode.Spheres[0].Material, result.Camera.InitialMedia[1])
		if shared {
			assert.Same(t, result.SceneNode.Spheres[1].Material, result.Camera.InitialMedia[0])
			assert.Same(t, result.SceneNode.Spheres[2].Material, result.WorldMedium)
		}
		stack, err := result.InitialMediumStack()
		require.NoError(t, err)
		require.Len(t, stack, 3)
		assert.Same(t, result.WorldMedium, stack[0])
		assert.Same(t, result.Camera.InitialMedia[0], stack[1])
		assert.Same(t, result.Camera.InitialMedia[1], stack[2])

		assert.Nil(t, restored.Frames[1].WorldMedium)
		assert.Nil(t, restored.Frames[1].Camera.InitialMedia)
		stack, err = restored.Frames[1].InitialMediumStack()
		require.NoError(t, err)
		require.Len(t, stack, 1)
		assert.Equal(t, scene.RefractionIndex_Air, stack[0].RefractionIndex)
		assert.Nil(t, stack[0].Absorption)
	}
}

func TestLegacyArchiveWithoutInitialMedia(t *testing.T) {
	frame := mediumArchiveTestFrame()
	filename := writeMediumTestArchive(t, frame)
	legacy := rewriteMediumTestArchive(t, filename, ".frame.msgpack", func(data []byte) []byte {
		var fields, cameraFields map[string]msgpack.RawMessage
		require.NoError(t, msgpack.Unmarshal(data, &fields))
		require.Contains(t, fields, "world-medium")
		require.NoError(t, msgpack.Unmarshal(fields["camera"], &cameraFields))
		require.Contains(t, cameraFields, "initial-media")
		// Existing camera tags are misspelled, so legacy files use Go names.
		// Adding the new settings must not change these established wire keys.
		require.Contains(t, cameraFields, "Origin")
		require.Contains(t, cameraFields, "RecursionDepth")
		delete(fields, "world-medium")
		delete(cameraFields, "initial-media")
		var err error
		fields["camera"], err = msgpack.Marshal(cameraFields)
		require.NoError(t, err)
		data, err = msgpack.Marshal(fields)
		require.NoError(t, err)
		return data
	})
	restored, err := ReadRenderFile(legacy)
	require.NoError(t, err)
	result := restored.Frames[0]
	assert.Nil(t, result.WorldMedium)
	assert.Nil(t, result.Camera.InitialMedia)
	assert.Equal(t, frame.Camera.Origin, result.Camera.Origin)
	assert.Equal(t, frame.Camera.Heading, result.Camera.Heading)
	stack, err := result.InitialMediumStack()
	require.NoError(t, err)
	require.Len(t, stack, 1)
	assert.Equal(t, scene.RefractionIndex_Air, stack[0].RefractionIndex)
	assert.Nil(t, stack[0].Absorption)
}

func TestInitialMediaInvalidArchiveReferences(t *testing.T) {
	filename := writeMediumTestArchive(t, mediumArchiveTestFrame())
	for _, tc := range []struct {
		name string
		edit func(*Frame)
		want string
	}{
		{"world index", func(f *Frame) { f.WorldMedium = 1000 }, "world medium: medium material index"},
		{"camera index", func(f *Frame) { f.Camera.InitialMedia = []MaterialIndex{1000} }, "camera initial medium 0: medium material index"},
		{"null entry", func(f *Frame) { f.Camera.InitialMedia = []MaterialIndex{0} }, "camera initial medium 0: material index must not be zero"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			broken := rewriteMediumTestArchive(t, filename, ".frame.msgpack", func(data []byte) []byte {
				var frame Frame
				require.NoError(t, msgpack.Unmarshal(data, &frame))
				tc.edit(&frame)
				data, err := msgpack.Marshal(frame)
				require.NoError(t, err)
				return data
			})
			result, err := ReadRenderFile(broken)
			require.ErrorContains(t, err, tc.want)
			assert.Nil(t, result)
		})
	}
	broken := rewriteMediumTestArchive(t, filename, ".material.msgpack", func(data []byte) []byte {
		var materials []Material
		require.NoError(t, msgpack.Unmarshal(data, &materials))
		materials[0].RefractionIndex = 0
		data, err := msgpack.Marshal(materials)
		require.NoError(t, err)
		return data
	})
	_, err := ReadRenderFile(broken)
	require.ErrorContains(t, err, "refraction index")
}

func TestInitialMediaSerializationErrors(t *testing.T) {
	frame := mediumArchiveTestFrame()
	frame.Camera.InitialMedia = []*scene.Material{nil}
	animation := scene.NewAnimation("invalid", 2, 2, 1, false, false).AddFrame(frame)
	require.ErrorContains(t, WriteRenderFile(filepath.Join(t.TempDir(), "invalid.zip"), animation), "camera initial medium 0")

	for _, world := range []bool{false, true} {
		frame := mediumArchiveTestFrame()
		image := floatimage.NewFloatImage(filepath.Join(t.TempDir(), "missing.png"), 1, 1)
		projection := scene.NewParallelImageProjection(image, scene.TextureTypeAlbedo, &vec3.T{}, vec3.UnitX, vec3.UnitY)
		medium := scene.NewMaterial().P(&projection)
		context := "camera initial medium 0"
		if world {
			frame.WorldMedium = medium
			context = "world medium"
		} else {
			frame.Camera.InitialMedia = []*scene.Material{medium}
		}
		animation := scene.NewAnimation("invalid", 2, 2, 1, false, false).AddFrame(frame)
		err := WriteRenderFile(filepath.Join(t.TempDir(), "invalid.zip"), animation)
		require.ErrorContains(t, err, context)
		require.ErrorContains(t, err, "could not read resource file")
	}
}

// Rewrite just the selected entries to exercise the public reader with legacy
// or malformed data, leaving all other archive tables and resources intact.
func rewriteMediumTestArchive(t *testing.T, filename, suffix string, edit func([]byte) []byte) string {
	t.Helper()
	reader, err := zip.OpenReader(filename)
	require.NoError(t, err)
	defer reader.Close()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for _, file := range reader.File {
		data, err := readZipFileEntry(file)
		require.NoError(t, err)
		if strings.HasSuffix(file.Name, suffix) {
			data = edit(data)
		}
		entry, err := writer.Create(file.Name)
		require.NoError(t, err)
		_, err = entry.Write(data)
		require.NoError(t, err)
	}
	require.NoError(t, writer.Close())
	result := filepath.Join(t.TempDir(), "rewritten.zip")
	require.NoError(t, os.WriteFile(result, buffer.Bytes(), 0600))
	return result
}
