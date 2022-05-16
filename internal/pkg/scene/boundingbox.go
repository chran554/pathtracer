package scene

import (
	"math"
)

type Bounds struct {
	Xmin, Xmax float64
	Ymin, Ymax float64
	Zmin, Zmax float64
}

func (b *Bounds) AddSphereBounds(s *Sphere) {
	b.Xmin = math.Min(b.Xmin, s.Origin[0]-s.Radius)
	b.Xmax = math.Max(b.Xmax, s.Origin[0]+s.Radius)
	b.Ymin = math.Min(b.Ymin, s.Origin[1]-s.Radius)
	b.Ymax = math.Max(b.Ymax, s.Origin[1]+s.Radius)
	b.Zmin = math.Min(b.Zmin, s.Origin[2]-s.Radius)
	b.Zmax = math.Max(b.Zmax, s.Origin[2]+s.Radius)
}

func (b *Bounds) AddDiscBounds(d *Disc) {
	b.Xmin = math.Min(b.Xmin, d.Origin[0]-d.Radius)
	b.Xmax = math.Max(b.Xmax, d.Origin[0]+d.Radius)
	b.Ymin = math.Min(b.Ymin, d.Origin[1]-d.Radius)
	b.Ymax = math.Max(b.Ymax, d.Origin[1]+d.Radius)
	b.Zmin = math.Min(b.Zmin, d.Origin[2]-d.Radius)
	b.Zmax = math.Max(b.Zmax, d.Origin[2]+d.Radius)
}

func (b *Bounds) AddBounds(s *Bounds) {
	b.Xmin = math.Min(b.Xmin, s.Xmin)
	b.Xmax = math.Max(b.Xmax, s.Xmax)
	b.Ymin = math.Min(b.Ymin, s.Ymin)
	b.Ymax = math.Max(b.Ymax, s.Ymax)
	b.Zmin = math.Min(b.Zmin, s.Zmin)
	b.Zmax = math.Max(b.Zmax, s.Zmax)
}

func NewBounds() Bounds {
	return Bounds{
		Xmin: math.MaxFloat64,
		Xmax: -math.MaxFloat64,
		Ymin: math.MaxFloat64,
		Ymax: -math.MaxFloat64,
		Zmin: math.MaxFloat64,
		Zmax: -math.MaxFloat64,
	}
}

func BoundingBoxIntersection1(line *Ray, bounds *Bounds) bool {
	hit := false

	if bounds == nil {
		return false
	}

	if !hit && line.Heading[0] != 0.0 {

		txmin := (bounds.Xmin - line.Origin[0]) / line.Heading[0] // Intersection with bounding box yz-plane at min x
		pxmin := line.point(txmin)

		// Intersect bounding box on the "min x" side
		hit = bounds.Ymin < pxmin[1] && bounds.Ymax > pxmin[1] &&
			bounds.Zmin < pxmin[2] && bounds.Zmax > pxmin[2]

		if !hit {
			txmax := (bounds.Xmax - line.Origin[0]) / line.Heading[0] // Intersection with bounding box yz-plane at max x
			pxmax := line.point(txmax)

			// Intersect bounding box on the "max x" side
			hit = bounds.Ymin < pxmax[1] && bounds.Ymax > pxmax[1] &&
				bounds.Zmin < pxmax[2] && bounds.Zmax > pxmax[2]
		}
	}

	if !hit && line.Heading[1] != 0.0 {
		tymin := (bounds.Ymin - line.Origin[1]) / line.Heading[1] // Intersection with bounding box xz-plane at min y
		pymin := line.point(tymin)

		// Intersect bounding box on the "min y" side
		hit = bounds.Xmin < pymin[0] && bounds.Xmax > pymin[0] &&
			bounds.Zmin < pymin[2] && bounds.Zmax > pymin[2]

		if !hit {
			tymax := (bounds.Ymax - line.Origin[1]) / line.Heading[1] // Intersection with bounding box xz-plane at max y
			pymax := line.point(tymax)

			// Intersect bounding box on the "max y" side
			hit = bounds.Xmin < pymax[0] && bounds.Xmax > pymax[0] &&
				bounds.Zmin < pymax[2] && bounds.Zmax > pymax[2]
		}
	}

	if !hit && line.Heading[2] != 0.0 {
		tzmin := (bounds.Zmin - line.Origin[2]) / line.Heading[2] // Intersection with bounding box xy-plane at min z
		pzmin := line.point(tzmin)

		// Intersect bounding box on the "min z" side
		hit = bounds.Xmin < pzmin[0] && bounds.Xmax > pzmin[0] &&
			bounds.Ymin < pzmin[1] && bounds.Ymax > pzmin[1]

		if !hit {
			tzmax := (bounds.Zmax - line.Origin[2]) / line.Heading[2] // Intersection with bounding box xy-plane at max z
			pzmax := line.point(tzmax)

			// Intersect bounding box on the "max z" side
			hit = bounds.Xmin < pzmax[0] && bounds.Xmax > pzmax[0] &&
				bounds.Ymin < pzmax[1] && bounds.Ymax > pzmax[1]
		}
	}

	return hit
}

func BoundingBoxIntersection2(line *Ray, bounds *Bounds) bool {
	hit := true
	noHit := false

	txmin := (bounds.Xmin - line.Origin[0]) / line.Heading[0] // Intersection with bounding box yz-plane at min x
	txmax := (bounds.Xmax - line.Origin[0]) / line.Heading[0] // Intersection with bounding box yz-plane at max x
	if txmin > txmax {
		txmin, txmax = txmax, txmin
	}

	tymin := (bounds.Ymin - line.Origin[1]) / line.Heading[1] // Intersection with bounding box xz-plane at min y
	tymax := (bounds.Ymax - line.Origin[1]) / line.Heading[1] // Intersection with bounding box xz-plane at max y
	if tymin > tymax {
		tymin, tymax = tymax, tymin
	}

	if tymax < txmin {
		return noHit
	}

	if txmax < tymin {
		return noHit
	}

	tmin := math.Max(txmin, tymin)
	tmax := math.Min(txmax, tymax)

	tzmin := (bounds.Zmin - line.Origin[2]) / line.Heading[2] // Intersection with bounding box xy-plane at min z
	tzmax := (bounds.Zmax - line.Origin[2]) / line.Heading[2] // Intersection with bounding box xy-plane at max z
	if tzmin > tzmax {
		tzmin, tzmax = tzmax, tzmin
	}

	if tzmax < tmin {
		return noHit
	}

	if tmax < tzmin {
		return noHit
	}

	tmin = math.Max(tmin, tzmin)
	tmax = math.Min(tmax, tzmax)

	return hit
}
