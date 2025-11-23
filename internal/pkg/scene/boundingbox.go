package scene

import (
	"math"

	"github.com/ungerik/go3d/float64/vec3"
)

type Bounds struct {
	Xmin, Xmax float64
	Ymin, Ymax float64
	Zmin, Zmax float64
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

func (b *Bounds) IsZeroBounds() bool {
	return b.Xmin == math.MaxFloat64 &&
		b.Xmax == -math.MaxFloat64 &&
		b.Ymin == math.MaxFloat64 &&
		b.Ymax == -math.MaxFloat64 &&
		b.Zmin == math.MaxFloat64 &&
		b.Zmax == -math.MaxFloat64

}

func (b *Bounds) Center() *vec3.T {
	return &vec3.T{
		(b.Xmax-b.Xmin)/2.0 + b.Xmin,
		(b.Ymax-b.Ymin)/2.0 + b.Ymin,
		(b.Zmax-b.Zmin)/2.0 + b.Zmin,
	}
}

func (b *Bounds) SizeX() float64 {
	return b.Xmax - b.Xmin
}

func (b *Bounds) SizeY() float64 {
	return b.Ymax - b.Ymin
}

func (b *Bounds) SizeZ() float64 {
	return b.Zmax - b.Zmin
}

func (b *Bounds) Max() float64 {
	return max(b.Xmax, max(b.Ymax, b.Zmax))
}

func (b *Bounds) Min() float64 {
	return min(b.Xmax, min(b.Ymax, b.Zmax))
}

func (b *Bounds) AddDiscBounds(d *Disc) {
	b.Xmin = min(b.Xmin, d.Origin[0]-d.Radius)
	b.Xmax = max(b.Xmax, d.Origin[0]+d.Radius)
	b.Ymin = min(b.Ymin, d.Origin[1]-d.Radius)
	b.Ymax = max(b.Ymax, d.Origin[1]+d.Radius)
	b.Zmin = min(b.Zmin, d.Origin[2]-d.Radius)
	b.Zmax = max(b.Zmax, d.Origin[2]+d.Radius)
}

func (b *Bounds) AddBounds(bounds *Bounds) {
	b.Xmin = min(b.Xmin, bounds.Xmin)
	b.Xmax = max(b.Xmax, bounds.Xmax)
	b.Ymin = min(b.Ymin, bounds.Ymin)
	b.Ymax = max(b.Ymax, bounds.Ymax)
	b.Zmin = min(b.Zmin, bounds.Zmin)
	b.Zmax = max(b.Zmax, bounds.Zmax)
}

func (b *Bounds) IncludeVertex(vertex *vec3.T) {
	b.Xmin = min(b.Xmin, vertex[0])
	b.Xmax = max(b.Xmax, vertex[0])
	b.Ymin = min(b.Ymin, vertex[1])
	b.Ymax = max(b.Ymax, vertex[1])
	b.Zmin = min(b.Zmin, vertex[2])
	b.Zmax = max(b.Zmax, vertex[2])
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

	invHeadingX := 1.0 / line.Heading[0]
	txmin := (bounds.Xmin - line.Origin[0]) * invHeadingX // Intersection with bounding box yz-plane at min x
	txmax := (bounds.Xmax - line.Origin[0]) * invHeadingX // Intersection with bounding box yz-plane at max x
	if txmin > txmax {
		txmin, txmax = txmax, txmin
	}

	invHeadingY := 1.0 / line.Heading[1]
	tymin := (bounds.Ymin - line.Origin[1]) * invHeadingY // Intersection with bounding box xz-plane at min y
	tymax := (bounds.Ymax - line.Origin[1]) * invHeadingY // Intersection with bounding box xz-plane at max y
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

	invHeadingZ := 1.0 / line.Heading[2]
	tzmin := (bounds.Zmin - line.Origin[2]) * invHeadingZ // Intersection with bounding box xy-plane at min z
	tzmax := (bounds.Zmax - line.Origin[2]) * invHeadingZ // Intersection with bounding box xy-plane at max z
	if tzmin > tzmax {
		tzmin, tzmax = tzmax, tzmin
	}

	if tzmax < tmin {
		return noHit
	}

	if tmax < tzmin {
		return noHit
	}

	tmin = max(tmin, tzmin)
	tmax = min(tmax, tzmax)

	return hit
}
