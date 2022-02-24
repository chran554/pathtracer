package main

import (
	"math"

	"github.com/ungerik/go3d/vec3"
)

func negative(t1 float32) bool {
	return math.Signbit(float64(t1))
}

func getLinePlaneIntersectionPoint(line ray, plane plane) (vec3.T, bool) {
	WarningNone := 0
	WarningNoIntersect := 1
	WarningIntersectBehind := 2

	warning := WarningNone

	m := plane.normal[0]*(plane.origin[0]-line.origin[0]) +
		plane.normal[1]*(plane.origin[1]-line.origin[1]) +
		plane.normal[2]*(plane.origin[2]-line.origin[2])

	n := plane.normal[0]*line.heading[0] + plane.normal[1]*line.heading[1] + plane.normal[2]*line.heading[2]

	t := float32(0)

	if n != 0.0 {
		t = m / n
	} else {
		warning = WarningNoIntersect
	}

	if t < 0.0 {
		warning = WarningIntersectBehind
	}

	if warning == WarningNone {
		intersectionPoint := vec3.T{
			line.origin[0] + t*line.heading[0],
			line.origin[1] + t*line.heading[1],
			line.origin[2] + t*line.heading[2],
		}

		return intersectionPoint, true
	}

	return vec3.T{0, 0, 0}, false
}

func discIntersection(line ray, disc Disc) (vec3.T, bool) {
	plane := plane{
		origin:   disc.Origin,
		normal:   disc.Normal,
		material: disc.Material,
	}

	intersectionPoint, intersection := getLinePlaneIntersectionPoint(line, plane)

	if intersection {
		intersectionDistance := vec3.Distance(&disc.Origin, &intersectionPoint)
		if intersectionDistance <= disc.Radius {
			return intersectionPoint, true
		}
	}

	return vec3.T{0, 0, 0}, false
}

func sphereIntersection(line ray, sphere Sphere) (vec3.T, bool) {
	WarningNone := 0
	WarningNoIntersection := 1
	WarningInside := 2
	WarningBehind := 3
	WarningOn := 4

	warning := WarningNone

	t3 := float32(0)

	m := line.heading[0]*(line.origin[0]-sphere.Origin[0]) +
		line.heading[1]*(line.origin[1]-sphere.Origin[1]) +
		line.heading[2]*(line.origin[2]-sphere.Origin[2])

	n := line.heading[0]*line.heading[0] +
		line.heading[1]*line.heading[1] +
		line.heading[2]*line.heading[2]

	o2 := sphere.Radius*sphere.Radius + float32(2)*((line.origin[0]*sphere.Origin[0])+(line.origin[1]*sphere.Origin[1])+(line.origin[2]*sphere.Origin[2]))

	o3 := (line.origin[0]*line.origin[0] + line.origin[1]*line.origin[1] + line.origin[2]*line.origin[2]) +
		(sphere.Origin[0]*sphere.Origin[0] + sphere.Origin[1]*sphere.Origin[1] + sphere.Origin[2]*sphere.Origin[2])

	o1 := o2 - o3
	p := m / n

	q := p*p + o1/n

	if q < 0.0 {
		// if (q < 0.0) there is no real root in calculation and therefore no intersection with Sphere
		warning = WarningNoIntersection

	} else {
		t1 := float32(0)
		t2 := float32(0)

		if q >= 0.0 {
			root := float32(math.Sqrt(float64(q)))
			t1 = -p + root
			t2 = -p - root

			t3 = t2
			if t1 < t2 {
				t3 = t1
			}
		}

		if (t1 == 0.0) || (t2 == 0.0) {
			warning = WarningOn

		} else if negative(t1) != negative(t2) {
			warning = WarningInside

		} else if negative(t1) && negative(t2) {
			warning = WarningBehind
		}
	}

	if (warning == WarningNone) || (warning == WarningOn) || (warning == WarningInside) {
		// Put in t3 into formula of line to get intersection point
		intersectionPoint := vec3.T{
			line.origin[0] + t3*line.heading[0],
			line.origin[1] + t3*line.heading[1],
			line.origin[2] + t3*line.heading[2],
		}

		return intersectionPoint, true
	}

	return vec3.T{0, 0, 0}, false
}
