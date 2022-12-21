package main

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
)

// https://github.com/gberrante/fastcurve_3d/blob/master/src/lib.rs

type Curve struct {
	points      []*vec3.T
	smoothLevel int
}

func NewCurve(controlPoints []*vec3.T, smoothLevel int) Curve {
	var curve Curve

	// Copy control points
	curve.points = make([]*vec3.T, len(controlPoints))
	for pointIndex, point := range controlPoints {
		curve.points[pointIndex] = &(*point)
	}

	curve.Smooth(smoothLevel)
	return curve
}

func (c *Curve) GetPoint(t float64) (*vec3.T, error) {
	if t < 0 || t > 1.0 {
		return nil, fmt.Errorf("can not get curve point at location %f, valid vaues [0..1]", t)
	}

	if t == 0.0 {
		return c.points[0], nil
	} else if t == 1.0 {
		return c.points[len(c.points)], nil
	}

	exactControlPointIndex := t * float64(c.AmountControlPoints())
	controlPointIndex := int(exactControlPointIndex)
	interControlPointDistance := exactControlPointIndex - float64(controlPointIndex)

	interpolatedPoint := vec3.Interpolate(c.points[controlPointIndex], c.points[controlPointIndex+1], interControlPointDistance)
	return &interpolatedPoint, nil
}

func (c *Curve) AmountControlPoints() int {
	return len(c.points)
}

func (c *Curve) GetControlPoint(pointIndex int) (*vec3.T, error) {
	if pointIndex < 0 || pointIndex >= len(c.points) {
		return nil, fmt.Errorf("point index %d out of bounds (%d points)", pointIndex, len(c.points))
	}

	return c.points[pointIndex], nil
}

func (c *Curve) Smooth(n int) {
	for _i := 0; _i < n; _i++ {
		c.points = fast3DStep(c.points)
	}
	c.smoothLevel += n
}

func (c *Curve) GetInterpolatedTangent(t float64) (*vec3.T, error) {
	if t == 0 {
		return c.tangentAtControlPoint(0)
	} else if t == 1 {
		return c.tangentAtControlPoint(len(c.points) - 1)
	}

	exactControlPointIndex := t * float64(c.AmountControlPoints())
	controlPointIndex := int(exactControlPointIndex)
	interControlPointDistance := exactControlPointIndex - float64(controlPointIndex)

	if interControlPointDistance == 0.0 {
		return c.tangentAtControlPoint(controlPointIndex)
	} else {
		tangent1, _ := c.tangentAtControlPoint(controlPointIndex)
		tangent2, _ := c.tangentAtControlPoint(controlPointIndex + 1)
		interpolatedTangent := vec3.Interpolate(tangent1, tangent2, interControlPointDistance)
		interpolatedTangent.Normalize()
		return &interpolatedTangent, nil
	}
}

func (c *Curve) tangentAtControlPoint(controlPointIndex int) (*vec3.T, error) {
	if controlPointIndex < 0 || controlPointIndex > len(c.points) {
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

func fast3DStep(points []*vec3.T) []*vec3.T {
	var newPoints []*vec3.T

	// Add first point
	newPoints = append(newPoints, points[0])

	for i := 1; i < len(points)-1; i++ {
		a1X := 0.25*points[i-1][0] + 0.75*points[i][0]
		a1Y := 0.25*points[i-1][1] + 0.75*points[i][1]
		a1Z := 0.25*points[i-1][2] + 0.75*points[i][2]

		c1X := 0.75*points[i][0] + 0.25*points[i+1][0]
		c1Y := 0.75*points[i][1] + 0.25*points[i+1][1]
		c1Z := 0.75*points[i][2] + 0.25*points[i+1][2]

		newPoints = append(newPoints, &vec3.T{a1X, a1Y, a1Z})
		newPoints = append(newPoints, &vec3.T{c1X, c1Y, c1Z})
	}

	// Add last point
	newPoints = append(newPoints, points[len(points)-1])

	return newPoints
}
