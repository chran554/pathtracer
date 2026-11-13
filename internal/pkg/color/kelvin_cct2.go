package color

import (
	"bufio"
	"bytes"
	"embed"
	"math"
	"strconv"
	"strings"
	"sync"
)

// The color package cannot import color/cie because color/cie's RGB conversion
// helpers use color.Color. Embed the observer data here instead of introducing
// an import cycle. The data is the same CIE 1931 2 degree observer table used
// by the cie package.
//
//go:embed cie/resources/CIE_xyz_1931_2deg.csv
var kelvinCCT2Resources embed.FS

const (
	kelvinCCT2ObserverResource = "cie/resources/CIE_xyz_1931_2deg.csv"
	kelvinCCT2PlanckC2         = 1.4387768775039338e-2 // h*c/k, m*K
	kelvinCCT2LUTMin           = 1.0
	kelvinCCT2LUTMax           = 1.0e9
	kelvinCCT2LUTRatio         = 1.01
	kelvinCCT2DuvLimit         = 0.05
	kelvinCCT2GoldenIterations = 64
)

type kelvinCCT2ObserverSample struct {
	wavelength float64 // metres
	xBar       float64
	yBar       float64
	zBar       float64
}

type kelvinCCT2LocusPoint struct {
	logTemperature float64
	u              float64
	v              float64
}

var (
	kelvinCCT2ObserverOnce sync.Once
	kelvinCCT2Observer     []kelvinCCT2ObserverSample
	kelvinCCT2ObserverOK   bool

	kelvinCCT2LocusOnce sync.Once
	kelvinCCT2Locus     []kelvinCCT2LocusPoint
)

func kelvinTemperatureCCT2(linearRGB *Color) (cct, duv float64, valid bool) {
	u, v, ok := kelvinCCT2LinearRGBToUV(linearRGB)
	if !ok {
		return 0, 0, false
	}

	if !kelvinCCT2LoadObserver() {
		return 0, 0, false
	}

	locus := kelvinCCT2LoadLocus()
	if len(locus) < 2 {
		return 0, 0, false
	}

	target := kelvinCCT2UV{u: u, v: v}
	nearest := kelvinCCT2NearestLocusPoint(locus, target)

	best := kelvinCCT2Candidate{}
	for _, segment := range kelvinCCT2CandidateSegments(locus, nearest) {
		candidate, ok := kelvinCCT2MinimizeSegment(segment.low, segment.high, target)
		if ok && (!best.ok || candidate.distanceSquared < best.distanceSquared) {
			best = candidate
		}
	}

	// The LUT is deliberately an accelerator, not a supported-temperature
	// boundary. If the closest sampled point is at an edge, walk outward by
	// 1% steps until the distance stops improving or floating-point arithmetic
	// can no longer represent a useful Planckian chromaticity.
	if nearest == 0 {
		best = kelvinCCT2SearchLowerEdge(locus[0], target, best)
	}
	if nearest == len(locus)-1 {
		best = kelvinCCT2SearchUpperEdge(locus[len(locus)-1], target, best)
	}

	if !best.ok || !(best.temperature > 0) || !isFiniteFloat64(best.temperature) {
		return 0, 0, false
	}

	duv = kelvinCCT2Duv(best.logTemperature, target)
	if !isFiniteFloat64(duv) {
		return 0, 0, false
	}

	cct = best.temperature
	valid = math.Abs(duv) <= kelvinCCT2DuvLimit && !best.asymptotic
	return cct, duv, valid
}

type kelvinCCT2UV struct {
	u float64
	v float64
}

type kelvinCCT2Segment struct {
	low  float64
	high float64
}

type kelvinCCT2Candidate struct {
	logTemperature  float64
	temperature     float64
	distanceSquared float64
	asymptotic      bool
	ok              bool
}

func kelvinCCT2LinearRGBToUV(linearRGB *Color) (u, v float64, ok bool) {
	if linearRGB == nil {
		return 0, 0, false
	}

	r := float64(linearRGB.R)
	g := float64(linearRGB.G)
	b := float64(linearRGB.B)
	if !isFiniteFloat64(r) || !isFiniteFloat64(g) || !isFiniteFloat64(b) {
		return 0, 0, false
	}

	// IEC 61966-2-1 linear sRGB, D65 reference white.
	x := 0.4124564*r + 0.3575761*g + 0.1804375*b
	y := 0.2126729*r + 0.7151520*g + 0.0721750*b
	z := 0.0193339*r + 0.1191920*g + 0.9503041*b
	if !isFiniteFloat64(x) || !isFiniteFloat64(y) || !isFiniteFloat64(z) || x < 0 || y <= 0 || z < 0 {
		return 0, 0, false
	}

	xyzSum := x + y + z
	if !isFiniteFloat64(xyzSum) || xyzSum <= 0 {
		return 0, 0, false
	}

	// First calculate the CIE 1931 chromaticity coordinates, then convert
	// them to the CIE 1960 uniform-chromaticity coordinates used by CCT/Duv.
	xChromaticity := x / xyzSum
	yChromaticity := y / xyzSum
	uvDenominator := -2*xChromaticity + 12*yChromaticity + 3
	if !isFiniteFloat64(uvDenominator) || uvDenominator <= 0 {
		return 0, 0, false
	}

	u = 4 * xChromaticity / uvDenominator
	v = 6 * yChromaticity / uvDenominator
	return u, v, isFiniteFloat64(u) && isFiniteFloat64(v)
}

func kelvinCCT2LoadObserver() bool {
	kelvinCCT2ObserverOnce.Do(func() {
		data, err := kelvinCCT2Resources.ReadFile(kelvinCCT2ObserverResource)
		if err != nil {
			return
		}

		scanner := bufio.NewScanner(bytes.NewReader(data))
		for scanner.Scan() {
			record := strings.Split(scanner.Text(), ",")
			if len(record) != 4 {
				return
			}

			wavelength, err1 := strconv.ParseFloat(record[0], 64)
			xBar, err2 := strconv.ParseFloat(record[1], 64)
			yBar, err3 := strconv.ParseFloat(record[2], 64)
			zBar, err4 := strconv.ParseFloat(record[3], 64)
			if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
				return
			}

			kelvinCCT2Observer = append(kelvinCCT2Observer, kelvinCCT2ObserverSample{
				wavelength: wavelength * 1e-9,
				xBar:       xBar,
				yBar:       yBar,
				zBar:       zBar,
			})
		}
		if scanner.Err() != nil {
			return
		}

		kelvinCCT2ObserverOK = len(kelvinCCT2Observer) > 0
	})

	return kelvinCCT2ObserverOK
}

func kelvinCCT2LoadLocus() []kelvinCCT2LocusPoint {
	kelvinCCT2LocusOnce.Do(func() {
		for temperature := kelvinCCT2LUTMin; temperature < kelvinCCT2LUTMax; temperature *= kelvinCCT2LUTRatio {
			u, v, ok := kelvinCCT2PlanckianUV(temperature)
			if ok {
				kelvinCCT2Locus = append(kelvinCCT2Locus, kelvinCCT2LocusPoint{
					logTemperature: math.Log(temperature),
					u:              u,
					v:              v,
				})
			}
		}

		u, v, ok := kelvinCCT2PlanckianUV(kelvinCCT2LUTMax)
		if ok {
			kelvinCCT2Locus = append(kelvinCCT2Locus, kelvinCCT2LocusPoint{
				logTemperature: math.Log(kelvinCCT2LUTMax),
				u:              u,
				v:              v,
			})
		}
	})

	return kelvinCCT2Locus
}

func kelvinCCT2PlanckianUV(temperature float64) (u, v float64, ok bool) {
	if !(temperature > 0) || !isFiniteFloat64(temperature) || !kelvinCCT2LoadObserver() {
		return 0, 0, false
	}

	maxLogRadiance := math.Inf(-1)
	for _, sample := range kelvinCCT2Observer {
		logRadiance := kelvinCCT2LogPlanckRadiance(sample.wavelength, temperature)
		if logRadiance > maxLogRadiance {
			maxLogRadiance = logRadiance
		}
	}
	if math.IsInf(maxLogRadiance, -1) {
		return 0, 0, false
	}

	var x, y, z float64
	for _, sample := range kelvinCCT2Observer {
		logRadiance := kelvinCCT2LogPlanckRadiance(sample.wavelength, temperature)
		if math.IsInf(logRadiance, -1) {
			continue
		}

		weight := math.Exp(logRadiance - maxLogRadiance)
		x += weight * sample.xBar
		y += weight * sample.yBar
		z += weight * sample.zBar
	}

	denominator := x + 15*y + 3*z
	if !(x > 0) || !(y > 0) || !(z > 0) || !isFiniteFloat64(denominator) || denominator <= 0 {
		return 0, 0, false
	}

	return 4 * x / denominator, 6 * y / denominator, true
}

func kelvinCCT2LogPlanckRadiance(wavelength, temperature float64) float64 {
	// The common Planck-law constants and the wavelength unit conversion cancel
	// when chromaticity is normalized. Keeping logarithms avoids underflow at
	// low temperatures and overflow in exp(c2/(lambda*T)).
	x := kelvinCCT2PlanckC2 / wavelength / temperature
	if math.IsInf(x, 1) {
		return math.Inf(-1)
	}

	var logExpm1 float64
	switch {
	case x > 50:
		// log(exp(x)-1) = x + log1p(-exp(-x)); the correction is below
		// float64 precision at this point.
		logExpm1 = x
	case x < 1e-5:
		// expm1 is accurate here, while the explicit branch avoids losing
		// the leading term if this code is evaluated on another libm.
		logExpm1 = math.Log(x) + 0.5*x
	default:
		logExpm1 = math.Log(math.Expm1(x))
	}

	return -5*math.Log(wavelength) - logExpm1
}

func kelvinCCT2NearestLocusPoint(locus []kelvinCCT2LocusPoint, target kelvinCCT2UV) int {
	nearest := 0
	nearestDistance := math.Inf(1)
	for i, point := range locus {
		distance := kelvinCCT2DistanceSquared(kelvinCCT2UV{u: point.u, v: point.v}, target)
		if distance < nearestDistance {
			nearest = i
			nearestDistance = distance
		}
	}
	return nearest
}

func kelvinCCT2CandidateSegments(locus []kelvinCCT2LocusPoint, nearest int) []kelvinCCT2Segment {
	segments := make([]kelvinCCT2Segment, 0, 3)
	start := max(nearest-1, 0)
	end := nearest + 1
	if end >= len(locus)-1 {
		end = len(locus) - 2
	}

	for i := start; i <= end; i++ {
		segments = append(segments, kelvinCCT2Segment{
			low:  locus[i].logTemperature,
			high: locus[i+1].logTemperature,
		})
	}
	return segments
}

func kelvinCCT2MinimizeSegment(low, high float64, target kelvinCCT2UV) (kelvinCCT2Candidate, bool) {
	if !(low < high) {
		return kelvinCCT2Candidate{}, false
	}

	// Golden-section minimization is robust for the smooth, one-dimensional
	// distance to a local Planckian-locus interval.
	goldenRatio := (math.Sqrt(5) - 1) / 2
	a := low
	b := high
	c := b - goldenRatio*(b-a)
	d := a + goldenRatio*(b-a)
	fc := kelvinCCT2DistanceAtLogTemperature(c, target)
	fd := kelvinCCT2DistanceAtLogTemperature(d, target)

	for range kelvinCCT2GoldenIterations {
		if fc < fd {
			b = d
			d = c
			fd = fc
			c = b - goldenRatio*(b-a)
			fc = kelvinCCT2DistanceAtLogTemperature(c, target)
		} else {
			a = c
			c = d
			fc = fd
			d = a + goldenRatio*(b-a)
			fd = kelvinCCT2DistanceAtLogTemperature(d, target)
		}
	}

	logTemperature := (a + b) / 2
	distanceSquared := kelvinCCT2DistanceAtLogTemperature(logTemperature, target)
	if math.IsInf(distanceSquared, 1) {
		return kelvinCCT2Candidate{}, false
	}

	return kelvinCCT2Candidate{
		logTemperature:  logTemperature,
		temperature:     math.Exp(logTemperature),
		distanceSquared: distanceSquared,
		ok:              true,
	}, true
}

func kelvinCCT2SearchLowerEdge(edge kelvinCCT2LocusPoint, target kelvinCCT2UV, best kelvinCCT2Candidate) kelvinCCT2Candidate {
	current := edge
	currentDistance := kelvinCCT2DistanceSquared(kelvinCCT2UV{u: current.u, v: current.v}, target)
	for {
		nextLogTemperature := current.logTemperature - math.Log(kelvinCCT2LUTRatio)
		if nextLogTemperature < math.Log(math.SmallestNonzeroFloat64) {
			best.asymptotic = true
			return best
		}

		nextTemperature := math.Exp(nextLogTemperature)
		nextU, nextV, ok := kelvinCCT2PlanckianUV(nextTemperature)
		if !ok {
			best.asymptotic = true
			return best
		}
		nextDistance := kelvinCCT2DistanceSquared(kelvinCCT2UV{u: nextU, v: nextV}, target)
		if !(nextDistance < currentDistance) {
			candidate, candidateOK := kelvinCCT2MinimizeSegment(nextLogTemperature, current.logTemperature, target)
			if candidateOK && (!best.ok || candidate.distanceSquared < best.distanceSquared) {
				best = candidate
			}
			return best
		}

		current = kelvinCCT2LocusPoint{logTemperature: nextLogTemperature, u: nextU, v: nextV}
		currentDistance = nextDistance
		if !best.ok || currentDistance < best.distanceSquared {
			best = kelvinCCT2Candidate{
				logTemperature:  current.logTemperature,
				temperature:     nextTemperature,
				distanceSquared: currentDistance,
				asymptotic:      true,
				ok:              true,
			}
		}
	}
}

func kelvinCCT2SearchUpperEdge(edge kelvinCCT2LocusPoint, target kelvinCCT2UV, best kelvinCCT2Candidate) kelvinCCT2Candidate {
	current := edge
	currentDistance := kelvinCCT2DistanceSquared(kelvinCCT2UV{u: current.u, v: current.v}, target)
	for {
		nextLogTemperature := current.logTemperature + math.Log(kelvinCCT2LUTRatio)
		if nextLogTemperature > math.Log(math.MaxFloat64) {
			best.asymptotic = true
			return best
		}

		nextTemperature := math.Exp(nextLogTemperature)
		if math.IsInf(nextTemperature, 0) || nextTemperature == math.MaxFloat64 {
			best.asymptotic = true
			return best
		}
		nextU, nextV, ok := kelvinCCT2PlanckianUV(nextTemperature)
		if !ok {
			best.asymptotic = true
			return best
		}
		nextDistance := kelvinCCT2DistanceSquared(kelvinCCT2UV{u: nextU, v: nextV}, target)
		if !(nextDistance < currentDistance) {
			candidate, candidateOK := kelvinCCT2MinimizeSegment(current.logTemperature, nextLogTemperature, target)
			if candidateOK && (!best.ok || candidate.distanceSquared < best.distanceSquared) {
				best = candidate
			}
			return best
		}

		current = kelvinCCT2LocusPoint{logTemperature: nextLogTemperature, u: nextU, v: nextV}
		currentDistance = nextDistance
		if !best.ok || currentDistance < best.distanceSquared {
			best = kelvinCCT2Candidate{
				logTemperature:  current.logTemperature,
				temperature:     nextTemperature,
				distanceSquared: currentDistance,
				asymptotic:      true,
				ok:              true,
			}
		}
	}
}

func kelvinCCT2DistanceAtLogTemperature(logTemperature float64, target kelvinCCT2UV) float64 {
	u, v, ok := kelvinCCT2PlanckianUV(math.Exp(logTemperature))
	if !ok {
		return math.Inf(1)
	}
	return kelvinCCT2DistanceSquared(kelvinCCT2UV{u: u, v: v}, target)
}

func kelvinCCT2DistanceSquared(a, b kelvinCCT2UV) float64 {
	du := a.u - b.u
	dv := a.v - b.v
	return du*du + dv*dv
}

func kelvinCCT2Duv(logTemperature float64, target kelvinCCT2UV) float64 {
	u, v, ok := kelvinCCT2PlanckianUV(math.Exp(logTemperature))
	if !ok {
		return math.NaN()
	}

	step := 1e-5
	uLow, vLow, lowOK := kelvinCCT2PlanckianUV(math.Exp(logTemperature - step))
	uHigh, vHigh, highOK := kelvinCCT2PlanckianUV(math.Exp(logTemperature + step))
	if !lowOK || !highOK {
		return math.NaN()
	}

	tangentU := (uHigh - uLow) / (2 * step)
	tangentV := (vHigh - vLow) / (2 * step)
	tangentLength := math.Hypot(tangentU, tangentV)
	if !(tangentLength > 0) || !isFiniteFloat64(tangentLength) {
		return math.NaN()
	}

	// This orientation makes the side with larger v, relative to the local
	// locus tangent, positive Duv.
	return ((target.u-u)*tangentV - (target.v-v)*tangentU) / tangentLength
}

func isFiniteFloat64(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
