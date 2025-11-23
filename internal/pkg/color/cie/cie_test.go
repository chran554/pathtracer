package cie

import (
	_ "embed"
	"github.com/stretchr/testify/assert"
	"strconv"
	"strings"
	"testing"
)

//go:embed resources/blackBodyLocus_1.txt
var blackBodyLocusText string

type blackBodyLocus struct {
	wavelength int
	x          float64
	y          float64
}

var blackBodyLocuses []blackBodyLocus

func init() {
	blackBodyLocusTextLines := strings.Split(blackBodyLocusText, "\n")
	for _, line := range blackBodyLocusTextLines {
		if !strings.HasPrefix(line, "#") && strings.TrimSpace(line) != "" {
			tokens := strings.Split(line, "\t")
			wavelength, _ := strconv.Atoi(tokens[0])
			x, _ := strconv.ParseFloat(tokens[1], 64)
			y, _ := strconv.ParseFloat(tokens[2], 64)
			locus := blackBodyLocus{wavelength: wavelength, x: x, y: y}

			blackBodyLocuses = append(blackBodyLocuses, locus)
		}
	}
}

func Test_TableDataValue(t *testing.T) {
	for i := IlluminantD65.Start; i <= IlluminantD65.End; i += IlluminantD65.Step {
		_, err2 := IlluminantD65.Value(i)
		assert.NoError(t, err2)
	}

	value300, err := IlluminantD65.Value(IlluminantD65.Start)
	assert.NoError(t, err)
	assert.Equal(t, 0.0341, value300)

	value830, err := IlluminantD65.Value(IlluminantD65.End)
	assert.NoError(t, err)
	assert.Equal(t, 60.3125, value830)

	value299, err := IlluminantD65.Value(IlluminantD65.Start - 1)
	assert.NoError(t, err)
	assert.Equal(t, 0.0, value299)

	value831, err := IlluminantD65.Value(IlluminantD65.End + 1)
	assert.NoError(t, err)
	assert.Equal(t, 0.0, value831)
}

func Test_TableDataInterpolatedValue(t *testing.T) {
	var x = float64(IlluminantD65.Start)
	for i := 1; x <= float64(IlluminantD65.End); i++ {
		_ = IlluminantD65.InterpolatedValue(x)
		x = float64(IlluminantD65.Start) + float64(i)*(float64(IlluminantD65.Step)/10.0)
	}

	value300 := IlluminantD65.InterpolatedValue(float64(IlluminantD65.Start))
	assert.Equal(t, 0.0341, value300)

	value830 := IlluminantD65.InterpolatedValue(float64(IlluminantD65.End))
	assert.Equal(t, 60.3125, value830)

	value299_5 := IlluminantD65.InterpolatedValue(float64(IlluminantD65.Start) - 0.5)
	assert.Equal(t, 0.0, value299_5)

	value830_5 := IlluminantD65.InterpolatedValue(float64(IlluminantD65.End) + 0.5)
	assert.Equal(t, 0.0, value830_5)
}

func Test_AddInterpolated(t *testing.T) {
	ill1 := IlluminantD65.Copy()
	ill2 := IlluminantD75.Copy()
	assert.Equal(t, IlluminantD65.Start, IlluminantD75.Start)

	err := ill1.AddInterpolated(ill2)
	assert.NoError(t, err)

	valueD65, err1 := IlluminantD65.Value(IlluminantD65.Start)
	valueD75, err2 := IlluminantD75.Value(IlluminantD75.Start)
	value, err3 := ill1.Value(IlluminantD65.Start)
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)
	assert.Equal(t, valueD65+valueD75, value)
}

func Test_SubInterpolated(t *testing.T) {
	ill1 := IlluminantD65.Copy()
	ill2 := IlluminantD75.Copy()
	assert.Equal(t, IlluminantD65.Start, IlluminantD75.Start)

	err := ill1.SubInterpolated(ill2)
	assert.NoError(t, err)

	valueD65, err1 := IlluminantD65.Value(IlluminantD65.Start)
	valueD75, err2 := IlluminantD75.Value(IlluminantD75.Start)
	value, err3 := ill1.Value(IlluminantD65.Start)
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)
	assert.Equal(t, valueD65-valueD75, value)
}

func Test_MulInterpolated(t *testing.T) {
	ill1 := IlluminantD65.Copy()
	ill2 := IlluminantD75.Copy()
	assert.Equal(t, IlluminantD65.Start, IlluminantD75.Start)

	err := ill1.MulInterpolated(ill2)
	assert.NoError(t, err)

	valueD65, err1 := IlluminantD65.Value(IlluminantD65.Start)
	valueD75, err2 := IlluminantD75.Value(IlluminantD75.Start)
	value, err3 := ill1.Value(IlluminantD65.Start)
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)
	assert.Equal(t, valueD65*valueD75, value)
}

func Test_D65(t *testing.T) {
	ill1 := IlluminantD65.Copy()
	ill2 := IlluminantD75.Copy()
	assert.Equal(t, IlluminantD65.Start, IlluminantD75.Start)

	err := ill1.MulInterpolated(ill2)
	assert.NoError(t, err)

	valueD65, err1 := IlluminantD65.Value(IlluminantD65.Start)
	valueD75, err2 := IlluminantD75.Value(IlluminantD75.Start)
	value, err3 := ill1.Value(IlluminantD65.Start)
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)
	assert.Equal(t, valueD65*valueD75, value)
}

func Test_D65_2deg_XYZ(t *testing.T) {
	D65 := IlluminantD65.Copy()
	D65xyz := D65.CIE1931_XYZ(Observer2Deg)
	D65xyzNormalized := D65xyz.NormalizeYLevel(100.0)

	// expected values according to
	// https://en.wikipedia.org/wiki/Illuminant_D65
	assert.InDeltaf(t, 95.04, D65xyzNormalized.X, 0.01, "")
	assert.InDeltaf(t, 100.0, D65xyzNormalized.Y, 0.01, "")
	assert.InDeltaf(t, 108.88, D65xyzNormalized.Z, 0.01, "")
}

func Test_D65_10deg_XYZ(t *testing.T) {
	D65 := IlluminantD65.Copy()
	D65xyz := D65.CIE1931_XYZ(Observer10Deg)
	D65xyzNormalized := D65xyz.NormalizeYLevel(100.0)

	// expected values according to
	// https://en.wikipedia.org/wiki/Illuminant_D65
	assert.InDeltaf(t, 94.811, D65xyzNormalized.X, 0.01, "")
	assert.InDeltaf(t, 100.0, D65xyzNormalized.Y, 0.01, "")
	assert.InDeltaf(t, 107.304, D65xyzNormalized.Z, 0.01, "")
}

// Test planck black body chromaticity calculated values against other pre-calculated values table.
//
// Source: https://www.waveformlighting.com/tech/black-body-and-reconstituted-daylight-locus-coordinates-by-cct-csv-excel-format
// Source: https://www.waveformlighting.com/files/blackBodyLocus_1.txt
func Test_PlanckBlackBodyChromaticity(t *testing.T) {
	for temperature := 1000; temperature <= 20000; temperature += 1 {
		spd := NewPlanckBlackBodyIlluminant(float64(temperature))
		cie1931_xyz := spd.CIE1931_XYZ(Observer2Deg)
		cc := cie1931_xyz.CIE1931_ChromaticityCoordinate()

		assert.InDelta(t, blackBodyLocuses[temperature-1000].x, cc.X, 0.00001)
		assert.InDelta(t, blackBodyLocuses[temperature-1000].y, cc.Y, 0.00001)
	}
}

func Test_XYZtoSRGB(t *testing.T) {
	D65 := IlluminantD65.Copy()
	xyz := D65.CIE1931_XYZ(Observer2Deg)
	srgb := xyz.D65_SRGB()

	assert.InDelta(t, 1.0, srgb.R, 0.001)
	assert.InDelta(t, 1.0, srgb.G, 0.001)
	assert.InDelta(t, 1.0, srgb.B, 0.001)
	assert.Equal(t, float32(1.0), srgb.A)
}

/*
func Test_PrintPlanckBlackBody4000(t *testing.T) {
	for wavelength := 50; wavelength <= 5001; wavelength += 50 {
		blackBodySpectralRadiantExitance := PlanckBlackBodySpectralRadiantExcitance(float64(wavelength)*1e-9, 4000.0)
		text := fmt.Sprintf("=%0.f", blackBodySpectralRadiantExitance)
		fmt.Println(strings.ReplaceAll(strings.ReplaceAll(text, ".", ","), "e", "E"))
	}
}
*/
