package main

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
)

type SmoothCurve struct {
	points         []*vec3.T
	smoothStrength float64
	smoothLevel    int
	lockEndPoints  bool
}

func NewSmoothCurve(controlPoints []*vec3.T, smoothStrength float64, smoothLevel int, subdivision int, lockEndPoints bool) SmoothCurve {
	var curve SmoothCurve
	curve.smoothStrength = smoothStrength
	curve.lockEndPoints = lockEndPoints

	// Copy control points
	curve.points = make([]*vec3.T, len(controlPoints))
	for pointIndex, point := range controlPoints {
		curve.points[pointIndex] = &(*point)
	}

	curve.Subdivide(subdivision)

	curve.Smooth(smoothLevel)
	return curve
}

func (c *SmoothCurve) Smooth(n int) {
	for _i := 0; _i < n; _i++ {
		c.points = c.smooth(c.points)
	}
	c.smoothLevel += n
}

func (c *SmoothCurve) Subdivide(n int) {
	if n <= 0 {
		return
	}

	amountNewPoints := len(c.points)*(n+1) - n
	newPoints := make([]*vec3.T, amountNewPoints)

	// Copy last point
	newPoints[len(newPoints)-1] = c.points[len(c.points)-1]

	m := n + 1
	t := 1.0 / float64(m)
	for pointIndex := 0; pointIndex <= len(c.points)-2; pointIndex++ {
		newPoints[pointIndex*m] = c.points[pointIndex]
		newPoints[(pointIndex+1)*m] = c.points[pointIndex+1]
		a := newPoints[pointIndex*m]
		b := newPoints[(pointIndex+1)*m]

		for subdivision := 1; subdivision <= n; subdivision++ {
			interpolatedPoint := vec3.Interpolate(a, b, t*float64(subdivision))
			newPoints[pointIndex*m+subdivision] = &interpolatedPoint
		}
	}

	c.points = newPoints
}

func (c *SmoothCurve) smooth(points []*vec3.T) []*vec3.T {
	var newPoints []*vec3.T

	// Add first point
	if c.lockEndPoints {
		newPoint := *points[0]
		newPoints = append(newPoints, &newPoint)
	} else {
		smoothedPoint := vec3.Interpolate(points[0], points[1], c.smoothStrength)
		newPoints = append(newPoints, &smoothedPoint)
	}

	for pointIndex := 1; pointIndex < len(points)-1; pointIndex++ {
		// Move current point delta towards previous point
		v1 := points[pointIndex-1].Subed(points[pointIndex]) // <currentPoint> ----[v1]----> <previousPoint>
		d1 := v1.Scaled(c.smoothStrength)

		// Move current point delta towards next point
		v2 := points[pointIndex+1].Subed(points[pointIndex]) // <currentPoint> ----[v2]----> <nextPoint>
		d2 := v2.Scaled(c.smoothStrength)

		smoothedPoint := *points[pointIndex]
		smoothedPoint.Add(&d1).Add(&d2)

		newPoints = append(newPoints, &smoothedPoint)
	}

	// Add last point
	if c.lockEndPoints {
		newPoint := *points[len(points)-1]
		newPoints = append(newPoints, &newPoint)
	} else {
		smoothedPoint := vec3.Interpolate(points[len(points)-1], points[len(points)-2], c.smoothStrength)
		newPoints = append(newPoints, &smoothedPoint)
	}

	return newPoints
}

func (c *SmoothCurve) GetPoint(t float64) (*vec3.T, error) {
	if t < 0 || t > 1.0 {
		return nil, fmt.Errorf("can not get curve point at location %f, valid vaues [0..1]", t)
	}

	if t == 0.0 {
		return c.points[0], nil
	} else if t == 1.0 {
		return c.points[len(c.points)], nil
	}

	exactControlPointIndex := t * float64(c.AmountControlPoints()-1)
	controlPointIndex := int(exactControlPointIndex)
	interControlPointDistance := exactControlPointIndex - float64(controlPointIndex)

	interpolatedPoint := vec3.Interpolate(c.points[controlPointIndex], c.points[controlPointIndex+1], interControlPointDistance)
	return &interpolatedPoint, nil
}

func (c *SmoothCurve) AmountControlPoints() int {
	return len(c.points)
}

func (c *SmoothCurve) GetControlPoint(pointIndex int) (*vec3.T, error) {
	if pointIndex < 0 || pointIndex >= len(c.points) {
		return nil, fmt.Errorf("point index %d out of bounds (%d points)", pointIndex, len(c.points))
	}

	return c.points[pointIndex], nil
}

func (c *SmoothCurve) GetSmoothInterpolatedTangent(t float64) (*vec3.T, error) {
	if t == 0 {
		return c.tangentAtControlPoint(0)
	} else if t == 1 {
		return c.tangentAtControlPoint(len(c.points) - 1)
	}

	exactControlPointIndex := t * float64(c.AmountControlPoints()-1)
	controlPointIndex := int(exactControlPointIndex)
	interControlPointDistance := exactControlPointIndex - float64(controlPointIndex)

	if interControlPointDistance == 0.0 {
		// At exact control point
		return c.tangentAtControlPoint(controlPointIndex)

	} else if interControlPointDistance <= 0.5 {
		// Between start control point and middle of edge
		tangent1, _ := c.tangentAtControlPoint(controlPointIndex)
		tangent2, _ := c.tangent(t)
		interpolatedTangent := vec3.Interpolate(tangent1, tangent2, interControlPointDistance)
		interpolatedTangent.Normalize()
		return &interpolatedTangent, nil

	} else /* if interControlPointDistance > 0.5 */ {
		// Between middle of edge and end control point
		tangent1, _ := c.tangent(t)
		tangent2, _ := c.tangentAtControlPoint(controlPointIndex + 1)
		interpolatedTangent := vec3.Interpolate(tangent1, tangent2, interControlPointDistance)
		interpolatedTangent.Normalize()
		return &interpolatedTangent, nil
	}
}

func (c *SmoothCurve) tangentAtControlPoint(controlPointIndex int) (*vec3.T, error) {
	if controlPointIndex < 0 || controlPointIndex >= len(c.points) {
		return nil, fmt.Errorf("point index %d out of bounds (%d points)", controlPointIndex, len(c.points))
	}

	if controlPointIndex == 0 {
		tangent := vec3.Sub(c.points[1], c.points[0])
		return (&tangent).Normalize(), nil
	} else if controlPointIndex == len(c.points)-1 {
		tangent := vec3.Sub(c.points[len(c.points)-1], c.points[len(c.points)-2])
		return (&tangent).Normalize(), nil
	}

	tangent1 := vec3.Sub(c.points[controlPointIndex], c.points[controlPointIndex-1])
	tangent1.Normalize()

	tangent2 := vec3.Sub(c.points[controlPointIndex+1], c.points[controlPointIndex])
	tangent2.Normalize()

	tangent1.Add(&tangent2)
	tangent1.Normalize()

	return &tangent1, nil
}

func (c *SmoothCurve) tangent(t float64) (*vec3.T, error) {
	if t < 0.0 || t > 1.0 {
		return nil, fmt.Errorf("illegal curve point reference %f, allowed values [0.0 .. 1.0]", t)
	}

	if t == 1.0 {
		// tangent in t == 1.0 is the same as the last edge (before the last control point)
		tangentHeading := vec3.Sub(c.points[len(c.points)-1], c.points[len(c.points)-2])
		return tangentHeading.Normalize(), nil
	} else if t == 0.0 {
		// tangent in t == 0.0 is the same as the first edge (after start control point)
		tangentHeading := vec3.Sub(c.points[1], c.points[0])
		return tangentHeading.Normalize(), nil
	}

	exactControlPointIndex := t * float64(c.AmountControlPoints()-1)
	controlPointIndex := int(exactControlPointIndex)
	// interControlPointDistance := exactControlPointIndex - float64(controlPointIndex)

	tangentHeading := vec3.Sub(c.points[controlPointIndex+1], c.points[controlPointIndex])
	return tangentHeading.Normalize(), nil
}
