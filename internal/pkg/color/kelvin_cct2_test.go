package color

import (
	"math"
	"testing"
)

func TestKelvinTemperatureCCT2PlanckianTemperatures(t *testing.T) {
	for _, want := range []float64{500, 1000, 2700, 6500, 40000, 50000, 100000} {
		linearRGB := kelvinCCT2LinearRGBAtTemperature(want)
		got, duv, valid := KelvinTemperatureCCT2(linearRGB)

		if !valid {
			t.Fatalf("KelvinTemperatureCCT2(%g K) was invalid: cct=%g duv=%g", want, got, duv)
		}
		if math.Abs(got-want) > math.Max(0.5, want*1e-5) {
			t.Errorf("KelvinTemperatureCCT2(%g K) = %g K, want within tolerance", want, got)
		}
		if math.Abs(duv) > 5e-8 {
			t.Errorf("KelvinTemperatureCCT2(%g K) Duv = %g, want approximately zero", want, duv)
		}
	}
}

func TestKelvinTemperatureCCT2ScaleInvariant(t *testing.T) {
	linearRGB := kelvinCCT2LinearRGBAtTemperature(3200)
	baseCCT, baseDuv, baseValid := KelvinTemperatureCCT2(linearRGB)
	if !baseValid {
		t.Fatalf("base color was invalid: cct=%g duv=%g", baseCCT, baseDuv)
	}

	for _, scale := range []float32{0.01, 0.25, 2, 100} {
		scaled := linearRGB.Copy().Multiply(scale)
		cct, duv, valid := KelvinTemperatureCCT2(scaled)
		if !valid {
			t.Fatalf("scaled color with factor %g was invalid: cct=%g duv=%g", scale, cct, duv)
		}
		if math.Abs(cct-baseCCT) > 0.01 || math.Abs(duv-baseDuv) > 2e-8 {
			t.Errorf("scale %g changed result: got cct=%g duv=%g, want cct=%g duv=%g", scale, cct, duv, baseCCT, baseDuv)
		}
	}
}

func TestKelvinTemperatureCCT2DuvSign(t *testing.T) {
	base := kelvinCCT2LinearRGBAtTemperature(6500)
	u, v, ok := kelvinCCT2LinearRGBToUV(base)
	if !ok {
		t.Fatal("could not calculate base chromaticity")
	}

	step := 1e-5
	uLow, vLow, lowOK := kelvinCCT2PlanckianUV(math.Exp(math.Log(6500) - step))
	uHigh, vHigh, highOK := kelvinCCT2PlanckianUV(math.Exp(math.Log(6500) + step))
	if !lowOK || !highOK {
		t.Fatal("could not calculate local Planckian tangent")
	}
	tangentU := (uHigh - uLow) / (2 * step)
	tangentV := (vHigh - vLow) / (2 * step)

	positive := kelvinCCT2LinearRGBFromUV(u+5e-2*tangentV, v-5e-2*tangentU)
	negative := kelvinCCT2LinearRGBFromUV(u-5e-2*tangentV, v+5e-2*tangentU)

	_, positiveDuv, positiveValid := KelvinTemperatureCCT2(positive)
	_, negativeDuv, negativeValid := KelvinTemperatureCCT2(negative)
	if !positiveValid || !negativeValid {
		t.Fatalf("Duv perturbations should be valid: positive=%v negative=%v", positiveValid, negativeValid)
	}
	if !(positiveDuv > 0) || !(negativeDuv < 0) {
		t.Errorf("Duv signs were positive=%g negative=%g", positiveDuv, negativeDuv)
	}
}

func TestKelvinTemperatureCCT2InvalidInput(t *testing.T) {
	for name, input := range map[string]*Color{
		"nil":      nil,
		"black":    NewColor(0, 0, 0),
		"nan":      NewColor(math.NaN(), 1, 1),
		"infinity": NewColor(math.Inf(1), 1, 1),
	} {
		t.Run(name, func(t *testing.T) {
			cct, duv, valid := KelvinTemperatureCCT2(input)
			if cct != 0 || duv != 0 || valid {
				t.Errorf("got cct=%g duv=%g valid=%v, want zero, zero, false", cct, duv, valid)
			}
		})
	}
}

func TestKelvinTemperatureCCT2NonPlanckianColor(t *testing.T) {
	cct, duv, valid := KelvinTemperatureCCT2(NewColor(0, 1, 0))
	if valid {
		t.Fatalf("green should not be a valid Planckian color: cct=%g duv=%g", cct, duv)
	}
	if !(cct > 0) || !(math.Abs(duv) > kelvinCCT2DuvLimit) {
		t.Errorf("got diagnostic cct=%g duv=%g, want finite CCT and Duv outside validity limit", cct, duv)
	}
}

func TestKelvinTemperatureCCT2SearchesBeyondInitialLUT(t *testing.T) {
	linearRGB := kelvinCCT2LinearRGBAtTemperature(2e9)
	cct, duv, valid := KelvinTemperatureCCT2(linearRGB)
	if !valid || !(cct > kelvinCCT2LUTMax) || math.Abs(duv) > 1e-7 {
		t.Fatalf("got cct=%g duv=%g valid=%v, want a finite result beyond the LUT", cct, duv, valid)
	}
}

func kelvinCCT2LinearRGBAtTemperature(temperature float64) *Color {
	u, v, ok := kelvinCCT2PlanckianUV(temperature)
	if !ok {
		panic("could not calculate Planckian chromaticity")
	}
	return kelvinCCT2LinearRGBFromUV(u, v)
}

func kelvinCCT2LinearRGBFromUV(u, v float64) *Color {
	// Set Y=1 and invert the CIE 1960 u,v equations to XYZ.
	x := 1.5 * u / v
	y := 1.0
	z := (6/v - x - 15) / 3

	return NewColor(
		3.2404542*x-1.5371385*y-0.4985314*z,
		-0.9692660*x+1.8760108*y+0.0415560*z,
		0.0556434*x-0.2040259*y+1.0572252*z,
	)
}
