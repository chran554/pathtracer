package scene

import (
	"math"
)

type Bounds struct {
	xmin, xmax float64
	ymin, ymax float64
	zmin, zmax float64
}

type BoundingBox interface {
	GetBounds() *Bounds
}

func BoundingBoxIntersection1(line *Ray, box *BoundingBox) bool {
	hit := false

	bounds := (*box).GetBounds()

	if !hit && line.Heading[0] != 0.0 {

		txmin := (bounds.xmin - line.Origin[0]) / line.Heading[0] // Intersection with bounding box yz-plane at min x
		pxmin := line.point(txmin)

		// Intersect bounding box on the "min x" side
		hit = bounds.ymin < pxmin[1] && bounds.ymax > pxmin[1] &&
			bounds.zmin < pxmin[2] && bounds.zmax > pxmin[2]

		if !hit {
			txmax := (bounds.xmax - line.Origin[0]) / line.Heading[0] // Intersection with bounding box yz-plane at max x
			pxmax := line.point(txmax)

			// Intersect bounding box on the "max x" side
			hit = bounds.ymin < pxmax[1] && bounds.ymax > pxmax[1] &&
				bounds.zmin < pxmax[2] && bounds.zmax > pxmax[2]
		}
	}

	if !hit && line.Heading[1] != 0.0 {
		tymin := (bounds.ymin - line.Origin[1]) / line.Heading[1] // Intersection with bounding box xz-plane at min y
		pymin := line.point(tymin)

		// Intersect bounding box on the "min y" side
		hit = bounds.xmin < pymin[0] && bounds.xmax > pymin[0] &&
			bounds.zmin < pymin[2] && bounds.zmax > pymin[2]

		if !hit {
			tymax := (bounds.ymax - line.Origin[1]) / line.Heading[1] // Intersection with bounding box xz-plane at max y
			pymax := line.point(tymax)

			// Intersect bounding box on the "max y" side
			hit = bounds.xmin < pymax[0] && bounds.xmax > pymax[0] &&
				bounds.zmin < pymax[2] && bounds.zmax > pymax[2]
		}
	}

	if !hit && line.Heading[1] != 0.0 {
		tzmin := (bounds.zmin - line.Origin[2]) / line.Heading[2] // Intersection with bounding box xy-plane at min z
		pzmin := line.point(tzmin)

		// Intersect bounding box on the "min z" side
		hit = bounds.xmin < pzmin[0] && bounds.xmax > pzmin[0] &&
			bounds.ymin < pzmin[1] && bounds.ymax > pzmin[1]

		if !hit {
			tzmax := (bounds.zmax - line.Origin[2]) / line.Heading[2] // Intersection with bounding box xy-plane at max z
			pzmax := line.point(tzmax)

			// Intersect bounding box on the "max z" side
			hit = bounds.xmin < pzmax[0] && bounds.xmax > pzmax[0] &&
				bounds.ymin < pzmax[1] && bounds.ymax > pzmax[1]
		}
	}

	return hit
}

func BoundingBoxIntersection2(line *Ray, box *BoundingBox) bool {
	hit := true
	noHit := false

	txmin := ((*box).GetBounds().xmin - line.Origin[0]) / line.Heading[0] // Intersection with bounding box yz-plane at min x
	txmax := ((*box).GetBounds().xmax - line.Origin[0]) / line.Heading[0] // Intersection with bounding box yz-plane at max x
	if txmin > txmax {
		txmin, txmax = txmax, txmin
	}

	tymin := ((*box).GetBounds().ymin - line.Origin[1]) / line.Heading[1] // Intersection with bounding box xz-plane at min y
	tymax := ((*box).GetBounds().ymax - line.Origin[1]) / line.Heading[1] // Intersection with bounding box xz-plane at max y
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

	tzmin := ((*box).GetBounds().zmin - line.Origin[2]) / line.Heading[2] // Intersection with bounding box xy-plane at min z
	tzmax := ((*box).GetBounds().zmax - line.Origin[2]) / line.Heading[2] // Intersection with bounding box xy-plane at max z
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
