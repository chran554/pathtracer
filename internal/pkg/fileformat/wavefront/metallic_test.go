package wavefront

import (
	"bufio"
	"bytes"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/scene"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetallicRoughnessMTL(t *testing.T) {
	for _, m := range []float64{0, 0.4, 1} {
		material := scene.NewMaterial().N("pbr").C(color.NewColor(0.8, 0.3, 0.1)).M(m, 0.25)
		var buffer bytes.Buffer
		writer := bufio.NewWriter(&buffer)
		serializeMaterial(writer, material)
		require.NoError(t, writer.Flush())
		parsed, err := parseMaterialFile(strings.Split(buffer.String(), "\n"), "")
		require.NoError(t, err)
		restored := convertToSceneMaterial(parsed["pbr"])
		assert.Equal(t, material.Metallicness, restored.Metallicness)
		assert.Equal(t, material.Roughness, restored.Roughness)
		assert.Equal(t, material.Color, restored.Color)
	}
	parsed, err := parseMaterialFile([]string{"newmtl plain", "newmtl metal", "refl 0.8", "newmtl override", "refl 0.8", "Pm 0.2"}, "")
	require.NoError(t, err)
	assert.Zero(t, convertToSceneMaterial(parsed["plain"]).Metallicness)
	assert.Equal(t, 1.0, convertToSceneMaterial(parsed["plain"]).Roughness)
	assert.Equal(t, 0.8, convertToSceneMaterial(parsed["metal"]).Metallicness)
	assert.Equal(t, 0.2, convertToSceneMaterial(parsed["override"]).Metallicness)
}
