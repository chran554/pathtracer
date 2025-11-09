package cie

import (
	"embed"
	"fmt"
	"math"
)

// CIEXYZ is the CIE 1931 XYZ coordinate.
//
// Note: The composants of the CIE 1931 coordinate are usually written in upper case
// (as opposed to CIE chromaticity coordinates written in lowercase).
// These struct attributes X, Y, and Z are written in their correct case.
type CIEXYZ struct {
	X float64
	Y float64
	Z float64
}

// CIEChromaticity is the CIE 1931 xyz chromaticity coordinate.
//
// Note: The composants of the chromaticity coordinate are usually written in lower case
// (as opposed to CIE XYZ coordinates written in uppercase).
// Goland requires however public attributes to be written in uppercase.
// These struct attributes X, Y, and Z of the CIE chromaticity coordinate are NOT written in their correct case.
type CIEChromaticity struct {
	X float64
	Y float64
	Z float64
}

type Illuminant struct {
	TableData
}

// Observer
//
// https://en.wikipedia.org/wiki/CIE_1931_color_space#CIE_standard_observer
type Observer struct {
	Name string
	xBar TableData
	yBar TableData
	zBar TableData
}

var IlluminantC Illuminant
var IlluminantD50 Illuminant
var IlluminantD55 Illuminant
var IlluminantD75 Illuminant
var IlluminantA Illuminant
var IlluminantD65 Illuminant

var Observer2Deg Observer
var Observer10Deg Observer

//go:embed resources/*.csv
var cieResourcesFS embed.FS // File system for embedded resource files with CIE tabular data

func init() {
	IlluminantC = parseIlluminantData(readEmbeddedTextFile(cieResourcesFS, "resources/CIE_illum_C.csv"), "C")
	IlluminantD55 = parseIlluminantData(readEmbeddedTextFile(cieResourcesFS, "resources/CIE_illum_D55.csv"), "D55")
	IlluminantD75 = parseIlluminantData(readEmbeddedTextFile(cieResourcesFS, "resources/CIE_illum_D75.csv"), "D75")
	IlluminantA = parseIlluminantData(readEmbeddedTextFile(cieResourcesFS, "resources/CIE_std_illum_A_1nm.csv"), "A")
	IlluminantD65 = parseIlluminantData(readEmbeddedTextFile(cieResourcesFS, "resources/CIE_std_illum_D65.csv"), "D65")

	Observer2Deg = parseObserverData(readEmbeddedTextFile(cieResourcesFS, "resources/CIE_xyz_1931_2deg.csv"), "2deg")
	Observer10Deg = parseObserverData(readEmbeddedTextFile(cieResourcesFS, "resources/CIE_xyz_1964_10deg.csv"), "10deg")
}

// CIE1931_XYZ calculates the CIE 1931 XYZ coordinate for this SPD (spectral power distribution).
//
// The Y composant is not normalized.
func (td *TableData) CIE1931_XYZ(o Observer) CIEXYZ {
	spdX := td.Copy()
	spdY := td.Copy()
	spdZ := td.Copy()

	spdX.MulInterpolated(o.xBar)
	spdY.MulInterpolated(o.yBar)
	spdZ.MulInterpolated(o.zBar)

	X := spdX.Integrate()
	Y := spdY.Integrate()
	Z := spdZ.Integrate()

	return CIEXYZ{X: X, Y: Y, Z: Z}
}

func (xyz CIEXYZ) NormalizeYLevel(yLevel float64) CIEXYZ {
	normalizeFactor := yLevel / xyz.Y
	return CIEXYZ{X: xyz.X * normalizeFactor, Y: yLevel, Z: xyz.Z * normalizeFactor}
}

// CIE1931_ChromaticityCoordinate gives the CIE 1931 chromaticity coordinates x, y, and z.
//
// These are the coordinates plotted in the "horse shoe" CIE 1931 color space chromaticity diagram.
// The diagram is used to plot monitor gamuts and color temperatures among other things.
func (xyz CIEXYZ) CIE1931_ChromaticityCoordinate() (c CIEChromaticity) {
	factor := 1.0 / (xyz.X + xyz.Y + xyz.Z)
	return CIEChromaticity{X: xyz.X * factor, Y: xyz.Y * factor, Z: xyz.Z * factor}
}

func NewPlanckBlackBodyIlluminant(kelvinTemperature float64) Illuminant {
	start := 300
	end := 830
	step := 1
	spectralRadiantExcitance := make([]float64, end-start+1)
	for i := 0; i < len(spectralRadiantExcitance); i += step {
		wavelength := start + i
		spectralRadiantExcitance[i] = PlanckBlackBodySpectralRadiantExcitance(float64(wavelength)*1e-9, kelvinTemperature)
	}

	return Illuminant{TableData{
		Name:  fmt.Sprintf("Planck black body illuminant at %.2f Kelvin", kelvinTemperature),
		Data:  spectralRadiantExcitance,
		Start: start,
		End:   end,
		Step:  step,
	}}
}

// PlanckBlackBodySpectralRadiantExcitance gets the black body spectral radiant excitance
// at a given wavelength (in meters(!), NOT nano meters) and temperature (in Kelvin).
//
// https://en.wikipedia.org/wiki/Planckian_locus
// https://en.wikipedia.org/wiki/Planck%27s_law
func PlanckBlackBodySpectralRadiantExcitance(wavelength float64, kelvinTemperature float64) float64 {
	h := 6.62607015e-34 // Planck's constant (Wiki: https://en.wikipedia.org/wiki/Planck_constant, constant since 2019)
	c := 2.99792458e+08 // Speed of light in vacuum (Wiki: https://en.wikipedia.org/wiki/Speed_of_light, constant since 1983)
	k := 1.380649e-23   // Boltzmann's constant (Wiki: https://en.wikipedia.org/wiki/Boltzmann_constant, constant since 2019)

	c1 := h * c * c   // first radiation constant
	c2 := (h * c) / k // second radiation constant

	//fmt.Println("c1", c1)
	//fmt.Println("c2", c2)

	t1 := 2.0 * math.Pi * c1
	t2 := math.Pow(wavelength, 5)
	t3 := c2 / (wavelength * kelvinTemperature)
	sre := t1 / (t2 * (math.Exp(t3) - 1.0))

	return sre
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

// https://en.wikipedia.org/wiki/Cubic_Hermite_spline
// https://en.wikipedia.org/wiki/Cubic_Hermite_spline#Interpolation_on_the_unit_interval_with_matched_derivatives_at_endpoints
func cubicHermiteSplineInterpolateValue(pn_1, pn, pn1, pn2, u float64) float64 {
	return 0.5*(((-pn_1+3*pn-3*pn1+pn2)*u+(2*pn_1-5*pn+4*pn1-pn2))*u+(-pn_1+pn1))*u + pn
}

// srgbGammaCompression or "encoding gamma" transforms a linear value into a gamma compressed value.
func srgbGammaCompression(linearValue float64) (gammaValue float64) {
	if linearValue <= 0.0031308 {
		return 12.92 * linearValue
	} else {
		return 1.055*math.Pow(linearValue, 1.0/2.4) - 0.055
	}
}

// srgbGammaExpansion or "decoding gamma" transforms a gamma encoded value into an expanded linear value.
func srgbGammaExpansion(gammaValue float64) (linearValue float64) {
	if gammaValue <= 0.04045 {
		return gammaValue / 12.92
	} else {
		return math.Pow((gammaValue+0.055)/1.055, 2.4)
	}
}

func GammaCompression(linearValue float64, gamma float64) (gammaValue float64) {
	if gamma == 1.0 {
		return linearValue
	}
	return math.Pow(linearValue, 1.0/gamma)
}

func GammaExpansion(gammaValue float64, gamma float64) (linearValue float64) {
	if gamma == 1.0 {
		return gammaValue
	}
	return math.Pow(gammaValue, gamma)
}
