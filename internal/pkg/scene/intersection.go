package scene

import (
	"math"

	"github.com/ungerik/go3d/float64/vec3"
)

func negative(t1 float64) bool {
	return math.Signbit(t1)
}

func GetLinePlaneIntersectionPoint(line *Ray, plane *Plane) (vec3.T, bool) {
	WarningNone := 0
	WarningNoIntersect := 1
	WarningIntersectBehind := 2

	warning := WarningNone

	m := plane.Normal[0]*(plane.Origin[0]-line.Origin[0]) +
		plane.Normal[1]*(plane.Origin[1]-line.Origin[1]) +
		plane.Normal[2]*(plane.Origin[2]-line.Origin[2])

	n := plane.Normal[0]*line.Heading[0] + plane.Normal[1]*line.Heading[1] + plane.Normal[2]*line.Heading[2]

	t := float64(0)

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
			line.Origin[0] + t*line.Heading[0],
			line.Origin[1] + t*line.Heading[1],
			line.Origin[2] + t*line.Heading[2],
		}

		return intersectionPoint, true
	}

	return vec3.T{0, 0, 0}, false
}

func DiscIntersection(line *Ray, disc *Disc) (vec3.T, bool) {
	plane := Plane{
		Origin:   disc.Origin,
		Normal:   disc.Normal,
		Material: disc.Material,
	}

	intersectionPoint, intersection := GetLinePlaneIntersectionPoint(line, &plane)

	if intersection {
		intersectionDistance := vec3.Distance(&disc.Origin, &intersectionPoint)
		if intersectionDistance <= disc.Radius {
			return intersectionPoint, true
		}
	}

	return vec3.T{0, 0, 0}, false
}

func SphereIntersection(line *Ray, sphere *Sphere) (vec3.T, bool) {
	WarningNone := 0
	WarningNoIntersection := 1
	WarningInside := 2
	WarningBehind := 3
	WarningOn := 4

	warning := WarningNone

	t3 := float64(0)

	m := line.Heading[0]*(line.Origin[0]-sphere.Origin[0]) +
		line.Heading[1]*(line.Origin[1]-sphere.Origin[1]) +
		line.Heading[2]*(line.Origin[2]-sphere.Origin[2])

	n := line.Heading[0]*line.Heading[0] +
		line.Heading[1]*line.Heading[1] +
		line.Heading[2]*line.Heading[2]

	o2 := sphere.Radius*sphere.Radius + float64(2)*((line.Origin[0]*sphere.Origin[0])+(line.Origin[1]*sphere.Origin[1])+(line.Origin[2]*sphere.Origin[2]))

	o3 := (line.Origin[0]*line.Origin[0] + line.Origin[1]*line.Origin[1] + line.Origin[2]*line.Origin[2]) +
		(sphere.Origin[0]*sphere.Origin[0] + sphere.Origin[1]*sphere.Origin[1] + sphere.Origin[2]*sphere.Origin[2])

	o1 := o2 - o3
	p := m / n

	q := p*p + o1/n

	if q < 0.0 {
		// if (q < 0.0) there is no real root in calculation and therefore no intersection with Sphere
		warning = WarningNoIntersection

	} else {
		t1 := float64(0)
		t2 := float64(0)

		if q >= 0.0 {
			root := math.Sqrt(q)
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

			if negative(t1) {
				t3 = t2
			} else {
				t3 = t1
			}

		} else if negative(t1) && negative(t2) {
			warning = WarningBehind
		}
	}

	if (warning == WarningNone) || (warning == WarningOn) || (warning == WarningInside) {
		// Put in t3 into formula of line to get intersection point
		intersectionPoint := vec3.T{
			line.Origin[0] + t3*line.Heading[0],
			line.Origin[1] + t3*line.Heading[1],
			line.Origin[2] + t3*line.Heading[2],
		}

		return intersectionPoint, true
	}

	return vec3.T{0, 0, 0}, false
}
