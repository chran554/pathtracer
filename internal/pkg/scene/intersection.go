package scene

import (
	"math"

	"github.com/ungerik/go3d/float64/vec3"
)

func negative(t1 float64) bool {
	return math.Signbit(t1)
}

/*
func GetLinePlaneIntersectionPoint(line *Ray, plane *Plane) (*vec3.T, bool) {
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

		return &intersectionPoint, true
	}

	return &vec3.T{0, 0, 0}, false
}
*/

// GetLinePlaneIntersectionPoint2 gets the intersection point of a line with a plane
// http://paulbourke.net/geometry/pointlineplane/   ("Intersection of a plane and a line")
func GetLinePlaneIntersectionPoint2(line *Ray, plane *Plane) (*vec3.T, bool) {
	WarningNone := 0
	WarningNoIntersect := 1
	WarningIntersectBehind := 2

	warning := WarningNone

	l := vec3.Sub(plane.Origin, line.Origin)
	m := vec3.Dot(plane.Normal, &l)
	n := vec3.Dot(plane.Normal, line.Heading)

	t := 0.0

	if n != 0.0 {
		t = m / n
	} else {
		warning = WarningNoIntersect
	}

	if t < 0.0 {
		warning = WarningIntersectBehind
	}

	if warning == WarningNone {
		//lineOffset := line.Heading.Scaled(t)
		//intersectionPoint := vec3.Add(line.Origin, &lineOffset)
		intersectionPoint := line.point(t)
		return intersectionPoint, true
	}

	return &vec3.T{0, 0, 0}, false
}

func BoundsIntersection(line *Ray, bounds *Bounds) (intersection bool) {
	return BoundingBoxIntersection2(line, bounds)
}

/*
func FacetIntersection(line *Ray, facet *Facet) (intersection bool, intersectionPoint *vec3.T, intersectionNormal *vec3.T) {
	if BoundsIntersection(line, facet.Bounds) {
		// https://www.scratchapixel.com/lessons/3d-basic-rendering/ray-tracing-rendering-a-triangle/moller-trumbore-ray-triangle-intersection

		v0v1 := facet.Vertices[1].Subed(facet.Vertices[0]) // v1 - v0
		v0v2 := facet.Vertices[2].Subed(facet.Vertices[0]) // v2 - v0
		pvec := vec3.Cross(line.Heading, &v0v2)            // dir.crossProduct(v0v2)
		det := vec3.Dot(&v0v1, &pvec)                      // v0v1.dotProduct(pvec)

		if math.Abs(det) < 0.000001 {
			return false, nil, nil
		}

		invDet := 1.0 / det

		tvec := line.Origin.Subed(facet.Vertices[0]) // orig - v0;
		u := vec3.Dot(&tvec, &pvec) * invDet         // tvec.dotProduct(pvec) * invDet
		if (u < 0.0) || (u > 1.0) {
			return false, nil, nil
		}

		qvec := vec3.Cross(&tvec, &v0v1)            // tvec.crossProduct(v0v1)
		v := vec3.Dot(line.Heading, &qvec) * invDet // dir.dotProduct(qvec) * invDet;
		if (v < 0) || (u+v > 1) {
			return false, nil, nil
		}

		t := vec3.Dot(&v0v2, &qvec) * invDet // v0v2.dotProduct(qvec) * invDet

		return true, line.point(t), facet.VertexNormals[0]
	}

	return false, nil, nil
}
*/

/*
bool rayTriangleIntersect(
    const Vec3f &orig, const Vec3f &dir,
    const Vec3f &v0, const Vec3f &v1, const Vec3f &v2,
    float &t, float &u, float &v)
{
#ifdef MOLLER_TRUMBORE
    Vec3f v0v1 = v1 - v0;
    Vec3f v0v2 = v2 - v0;
    Vec3f pvec = dir.crossProduct(v0v2);
    float det = v0v1.dotProduct(pvec);
#ifdef CULLING
    // if the determinant is negative the triangle is backfacing
    // if the determinant is close to 0, the ray misses the triangle
    if (det < kEpsilon) return false;
#else
    // ray and triangle are parallel if det is close to 0
    if (fabs(det) < kEpsilon) return false;
#endif
    float invDet = 1 / det;

    Vec3f tvec = orig - v0;
    u = tvec.dotProduct(pvec) * invDet;
    if (u < 0 || u > 1) return false;

    Vec3f qvec = tvec.crossProduct(v0v1);
    v = dir.dotProduct(qvec) * invDet;
    if (v < 0 || u + v > 1) return false;

    t = v0v2.dotProduct(qvec) * invDet;

    return true;
#else
    ...
#endif
}
*/

func FacetIntersection2(line *Ray, facet *Facet) (intersection bool, intersectionPoint *vec3.T, intersectionNormal *vec3.T) {
	plane := Plane{
		Origin: facet.Vertices[0],
		Normal: facet.Normal,
	}

	intersectionPoint, intersection = GetLinePlaneIntersectionPoint2(line, &plane)

	if intersection {
		withinTriangle, vertexWeights := isPointWithinTriangleFacet(intersectionPoint, facet)
		if withinTriangle {
			normal := interpolateTriangleFacetNormal(facet, vertexWeights)
			return true, intersectionPoint, &normal
		}
	}

	return false, nil, nil
}

// interpolateTriangleFacetNormal interpolates intersection point normal from facet (triangle) vertex normals
func interpolateTriangleFacetNormal(facet *Facet, vertexWeights *vec3.T) vec3.T {
	normal := vec3.T{}
	if len(facet.VertexNormals) == 3 {
		for i := 0; i < 3; i++ {
			weightedVertexNormal := facet.VertexNormals[i].Scaled(vertexWeights[i])
			normal.Add(&weightedVertexNormal)
		}
	} else {
		normal = *facet.Normal
	}
	normal.Normalize()
	return normal
}

func isPointWithinTriangleFacet(point *vec3.T, facet *Facet) (isWithin bool, vertexWeights *vec3.T) {
	// Point within triangle: http://softsurfer.com/Archive/algorithm_0105/algorithm_0105.htm
	// http://mrl.nyu.edu/~dzorin/rendering/lectures/lecture2/lecture2-6pp.pdf (Not used here)
	// https://gamedev.stackexchange.com/questions/23743/whats-the-most-efficient-way-to-find-barycentric-coordinates

	u := facet.Vertices[1].Subed(facet.Vertices[0])
	v := facet.Vertices[2].Subed(facet.Vertices[0])
	w := point.Subed(facet.Vertices[0])

	t1 := vec3.Dot(&u, &v)
	t2 := vec3.Dot(&u, &u) // length of u vector squared
	t3 := vec3.Dot(&v, &v) // length of v vector squared

	t4 := vec3.Dot(&w, &v)
	t5 := vec3.Dot(&w, &u)

	t6 := 1.0 / (t1*t1 - t2*t3)
	a := (t1*t4 - t3*t5) * t6
	b := (t1*t5 - t2*t4) * t6
	c := 1.0 - (a + b)

	return (a >= 0) && (b >= 0) && (c >= 0), &vec3.T{c, a, b}
}

/*
func isPointWithinTriangle2(point *vec3.T, facet *Facet) (isWithin bool, vertexWeights *vec3.T) {
	// https://math.stackexchange.com/questions/4322/check-whether-a-point-is-within-a-3d-triangle
	// https://answers.unity.com/questions/383804/calculate-uv-coordinates-of-3d-point-on-plane-of-m.html

	// Check whether a point P is within triangle made up of points A,B, and C

	AB := vec3.Sub(facet.Vertices[1], facet.Vertices[0])
	AC := vec3.Sub(facet.Vertices[2], facet.Vertices[0])

	PA := vec3.Sub(facet.Vertices[0], point)
	PB := vec3.Sub(facet.Vertices[1], point)
	PC := vec3.Sub(facet.Vertices[2], point)

	triangleABCNormal := vec3.Cross(&AB, &AC)
	doubleTriangleAreaInv := 1.0 / triangleABCNormal.Length()

	pointNormalPBC := vec3.Cross(&PB, &PC)
	pointNormalPCA := vec3.Cross(&PC, &PA)
	pointNormalPAB := vec3.Cross(&PA, &PB)

	alpha := pointNormalPBC.Length() * doubleTriangleAreaInv
	beta := pointNormalPCA.Length() * doubleTriangleAreaInv
	gamma := pointNormalPAB.Length() * doubleTriangleAreaInv
	// gamma := 1.0 - alpha - beta

	delta := alpha + beta + gamma

	return (0 <= alpha) && (alpha <= 1) &&
		(0 <= beta) && (beta <= 1) &&
		(0 <= gamma) && (gamma <= 1) &&
		(delta-1.0+0.0000001) <= 2*0.0000001, &vec3.T{alpha, beta, gamma}
}
*/

func FacetStructureIntersection(line *Ray, facetStructure *FacetStructure) (intersection bool, intersectionPoint *vec3.T, intersectionNormal *vec3.T, intersectionMaterial *Material) {
	if BoundsIntersection(line, facetStructure.Bounds) {
		var closestIntersectionDistance = math.MaxFloat64

		for _, facet := range facetStructure.Facets {
			facetIntersection, facetIntersectionPoint, facetIntersectionNormal := FacetIntersection2(line, facet)

			if facetIntersection {
				tempIntersectionDistance := vec3.Distance(line.Origin, facetIntersectionPoint)

				if tempIntersectionDistance < closestIntersectionDistance {
					intersection = facetIntersection
					intersectionPoint = facetIntersectionPoint
					intersectionNormal = facetIntersectionNormal
					intersectionMaterial = facetStructure.Material

					closestIntersectionDistance = tempIntersectionDistance
				}
			}

		}

		for _, facetSubStructure := range facetStructure.FacetStructures {
			subStructureIntersection, subStructureIntersectionPoint, subStructureIntersectionNormal, subStructureIntersectionMaterial := FacetStructureIntersection(line, facetSubStructure)

			if subStructureIntersection {
				tempIntersectionDistance := vec3.Distance(line.Origin, subStructureIntersectionPoint)

				if tempIntersectionDistance < closestIntersectionDistance {
					intersection = subStructureIntersection
					intersectionPoint = subStructureIntersectionPoint
					intersectionNormal = subStructureIntersectionNormal
					intersectionMaterial = subStructureIntersectionMaterial

					closestIntersectionDistance = tempIntersectionDistance
				}
			}
		}
	}

	return intersection, intersectionPoint, intersectionNormal, intersectionMaterial
}

func DiscIntersection(line *Ray, disc *Disc) (*vec3.T, bool) {
	plane := Plane{Origin: disc.Origin, Normal: disc.Normal}
	intersectionPoint, intersection := GetLinePlaneIntersectionPoint2(line, &plane)

	if intersection {
		intersectionDistance := vec3.Distance(disc.Origin, intersectionPoint)
		if intersectionDistance <= disc.Radius {
			return intersectionPoint, true
		}
	}

	return nil, false
}

func SphereIntersection(line *Ray, sphere *Sphere) (*vec3.T, bool) {
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

		return &intersectionPoint, true
	}

	return nil, false
}
