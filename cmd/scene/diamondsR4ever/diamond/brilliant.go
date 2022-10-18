package diamond

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"math"
	"pathtracer/internal/pkg/color"
	scn "pathtracer/internal/pkg/scene"
)

// GetDiamondRoundBrilliantCut will return a facet structure of a round, brilliant cut, diamond.
// The diamond is a 57 facet, brilliant cut, diamond.
// It has a girdle. It has no culet (no 58:th facet), and is a "pointed" or "None" culet.
// The girdle itself consist of 32 facets. Two facets per each "upper"/"lower" half pair.
//
// The girdle is neither "dug out" nor "painted" but "normal".
// I.e. the distance in the girdle between crown and pavilion at hill positions
// where bezel facet meet pavilion main facet are equal to the positions where upper and lower half facet edges meet.
//
// Information: https://www.capediamonds.co.za/diamond-info/brilliant-cut/
// https://www.gia.edu/diamond-cut/diamond-cut-anatomy-round-brilliant
func GetDiamondRoundBrilliantCut(scale float64) *scn.FacetStructure {
	diamondMaterial := scn.Material{
		Color:           color.Color{R: 0.9, G: 0.6, B: 0.6},
		Glossiness:      0.0,
		Roughness:       1.0,
		RefractionIndex: 2.42,
		Transparency:    0.0,
	}

	// The girdle diameter of the diamond is the average widest diameter of the diamond when viewed from above.
	// The girdle diameter measurement is key as it determines the proportions of the brilliant cut diamond.
	// Most percentage proportions of the brilliant cut round diamond are calculated as percentages of the girdle diameter.

	girdleDiameter := 1.0                          // girdleDiameter The diameter of the round diamond as seen from above. All facet sizes of a diamond are given in percentage of the girdle (diamond) width. Thus, we set girdle size as 100%.
	crownAngleDegrees := 34.0                      // crownAngleDegrees The angle of the crown side, from the girdle to the table, as when you see the diamond from the side. (value is typically 33 degrees - 35 degrees). According to Marcel Tolkowsky, a 40.75-degree pavilion angle gets perfectly paired with a 34.5-degree crown angle. The recommended crown angle ranges from 34 to 35 degrees.
	tableFacetSizeRelativeGirdle := 0.56           // tableFacetSizeRelativeGirdle is the size of the tablet. It is expressed as a percentage of the Girdle diameter (value is typically 53%-58%, 56% might be a good start)
	starSizeRelativeCrownSide := 0.55              // starSizeRelativeCrownSide is the size of the star facets. It is expressed as percentage of the length of the side of the crown. <crown side width> = (<girdle width> - <table width>)/2.0 (value is typically 50%-60%)
	pavilionAngleDegrees := 40.9                   // According to Marcel Tolkowsky, the optimal pavilion angle degree is 40.75. However, unless you are looking for a super-ideal diamond cut with a pavilion angle of 40.9 degrees, opting for stone within a range of 40.6 – 41 degrees is safe, providing other parameters meet their recommended ranges.
	lowerHalfFacetSizeRelativeGirdleRadius := 0.77 //
	culetDiameter := 0.0                           // Do not change this from value 0.0. Culet facet is not included in 3D model. Only diamonds with "pointed" or "None" culet are created.

	// Create crown

	// Pre calculations
	girdleRadius := girdleDiameter / 2.0
	culetRadius := culetDiameter / 2.0
	tableRadius := girdleRadius * tableFacetSizeRelativeGirdle
	girdleToTableLength := girdleRadius - tableRadius
	crownSideLength := girdleToTableLength / math.Cos(deg2rad(crownAngleDegrees))
	crownHeight := crownSideLength * math.Sin(deg2rad(crownAngleDegrees))

	girdleToCuletLength := girdleRadius - culetRadius
	pavilionSideLength := girdleToCuletLength / math.Cos(deg2rad(pavilionAngleDegrees))
	pavilionHeight := pavilionSideLength * math.Sin(deg2rad(pavilionAngleDegrees))

	fmt.Printf("Table diameter:  %.2f\n", tableRadius*2.0)
	fmt.Printf("Girdle diameter: %.2f\n", girdleDiameter)
	fmt.Printf("Crown height:    %.2f\n", crownHeight)
	fmt.Printf("Pavilion height: %.2f\n", pavilionHeight)
	fmt.Printf("Total height:    %.2f\n", pavilionHeight+crownHeight)

	amountTableCorners := 8
	tableFacetPoints := calculateTableFacetPoints(amountTableCorners, tableRadius, crownHeight)
	girdleBezelPoints := calculateGirdleBezelPoints(amountTableCorners, girdleDiameter, crownHeight)
	girdleUpperHalfPoints := calculateGirdleUpperHalfPoints(amountTableCorners, girdleDiameter, crownHeight)
	starTipPoints := calculateStarTipPoints(tableFacetPoints, girdleBezelPoints, tableRadius, girdleToTableLength, starSizeRelativeCrownSide, crownHeight)

	// Crown "table"-facet (i.e. top flat facet, 8 equally sided facet)

	// The table of the diamond is the largest facet, the flat part on top of the diamond.
	// Having an advantageous proportion of table to girdle diameter creates the greatest aperture of sight
	// into the stone and the greatest diamond brilliance.
	// A good-sized table, measured as an average of the four widest points of the table,
	// allows us to see into the diamond, but having too large a table percentage can create a flat effect,
	// with little radiance and fire, known as the “fish eye” effect.
	// The table size, calculated as a percentage of the girdle diameter, will be indicated on the diamond’s GIA certificate.
	table := scn.Facet{Vertices: tableFacetPoints}

	// Crown "star"-facets (the facets that share a side with "table" facet.)
	var starFacets []*scn.Facet
	amountStarFacets := amountTableCorners
	for starFacetIndex := 0; starFacetIndex < amountStarFacets; starFacetIndex++ {
		starFacet := scn.Facet{
			Vertices: []*vec3.T{
				tableFacetPoints[starFacetIndex],
				starTipPoints[starFacetIndex],
				tableFacetPoints[(starFacetIndex+1)%amountStarFacets],
			},
		}
		starFacets = append(starFacets, &starFacet)
	}

	// Crown Bezel-facets (the kite like, 4 sided, facet of the crown)
	var bezelFacets []*scn.Facet
	amountBezelFacets := amountTableCorners
	for bezelFacetIndex := 0; bezelFacetIndex < amountBezelFacets; bezelFacetIndex++ {
		bezelFacet := scn.Facet{
			Vertices: []*vec3.T{
				tableFacetPoints[bezelFacetIndex],
				starTipPoints[(bezelFacetIndex+amountBezelFacets-1)%amountBezelFacets],
				girdleBezelPoints[bezelFacetIndex],
				starTipPoints[bezelFacetIndex],
			},
		}
		bezelFacets = append(bezelFacets, &bezelFacet)
	}

	// Upper half facet (the triangle facets on the crown closest to the girdle)
	var upperHalfPairFacets []*scn.Facet
	amountUpperHalfPairFacets := amountTableCorners
	for upperHalfPairFacetIndex := 0; upperHalfPairFacetIndex < amountUpperHalfPairFacets; upperHalfPairFacetIndex++ {
		pb1 := girdleBezelPoints[upperHalfPairFacetIndex]
		pb2 := girdleBezelPoints[(upperHalfPairFacetIndex+1)%amountUpperHalfPairFacets]
		ps := starTipPoints[upperHalfPairFacetIndex]
		p2 := girdleUpperHalfPoints[upperHalfPairFacetIndex]

		upperHalfPairBaseAngle := (float64(upperHalfPairFacetIndex) / float64(amountUpperHalfPairFacets)) * (2.0 * math.Pi)
		upperHalfPairSubAngleIncrement := (2.0 * math.Pi) / (float64(amountUpperHalfPairFacets) * 4.0) // 5 girdle corners for each upper half pair facets set

		plane1 := scn.NewPlane(ps, pb1, p2, "", nil)
		gp1Angle := upperHalfPairBaseAngle + 1.0*upperHalfPairSubAngleIncrement
		gp1 := &vec3.T{girdleRadius * math.Cos(gp1Angle), 0, girdleRadius * math.Sin(gp1Angle)}
		//gp1 := vec3.Interpolate(pb1, p2, 0.5)
		p1 := verticalLinePlaneIntersection(plane1, gp1)

		plane2 := scn.NewPlane(ps, p2, pb2, "", nil)
		gp2Angle := upperHalfPairBaseAngle + 3.0*upperHalfPairSubAngleIncrement
		gp2 := &vec3.T{girdleRadius * math.Cos(gp2Angle), 0, girdleRadius * math.Sin(gp2Angle)}
		//gp2 := vec3.Interpolate(p2, pb2, 0.5)
		p3 := verticalLinePlaneIntersection(plane2, gp2)

		upperHalfPairFacet1 := scn.Facet{
			//Vertices: []*vec3.T{ps, pb1, p2},
			Vertices: []*vec3.T{ps, pb1, p1, p2},
		}
		upperHalfPairFacet2 := scn.Facet{
			//Vertices: []*vec3.T{ps, p2, pb2},
			Vertices: []*vec3.T{ps, p2, p3, pb2},
		}

		upperHalfPairFacets = append(upperHalfPairFacets, &upperHalfPairFacet1, &upperHalfPairFacet2)
	}

	// Pavilion facets

	pavilionPoint := &vec3.T{0, -pavilionHeight, 0}

	lowerHalfTipPoints := calculateLowerHalfTipPoints(amountTableCorners, pavilionHeight, lowerHalfFacetSizeRelativeGirdleRadius, girdleDiameter, girdleBezelPoints, pavilionPoint)

	// Pavilion - Main facets (the kite like, 4 sided, facet of the pavilion)
	var pavilionMainFacets []*scn.Facet
	amountPavilionMainFacets := amountTableCorners
	for pavilionMainFacetIndex := 0; pavilionMainFacetIndex < amountPavilionMainFacets; pavilionMainFacetIndex++ {
		pavilionMainFacet := scn.Facet{
			Vertices: []*vec3.T{
				girdleBezelPoints[pavilionMainFacetIndex],
				lowerHalfTipPoints[pavilionMainFacetIndex],
				pavilionPoint,
				lowerHalfTipPoints[(pavilionMainFacetIndex+amountPavilionMainFacets-1)%amountPavilionMainFacets],
			},
		}
		pavilionMainFacets = append(pavilionMainFacets, &pavilionMainFacet)
	}

	// Pavilion - Lower half pair facets
	var lowerHalfPairFacets []*scn.Facet
	amountLowerHalfPairFacets := amountTableCorners
	for lowerHalfFacetPairIndex := 0; lowerHalfFacetPairIndex < amountLowerHalfPairFacets; lowerHalfFacetPairIndex++ {
		lowerHalfPairFacet1 := scn.Facet{
			Vertices: []*vec3.T{
				girdleUpperHalfPoints[lowerHalfFacetPairIndex],
				lowerHalfTipPoints[lowerHalfFacetPairIndex],
				girdleBezelPoints[lowerHalfFacetPairIndex],
			},
		}
		lowerHalfPairFacet2 := scn.Facet{
			Vertices: []*vec3.T{
				girdleBezelPoints[(lowerHalfFacetPairIndex+1)%amountLowerHalfPairFacets],
				lowerHalfTipPoints[lowerHalfFacetPairIndex],
				girdleUpperHalfPoints[lowerHalfFacetPairIndex],
			},
		}

		lowerHalfPairFacets = append(lowerHalfPairFacets, &lowerHalfPairFacet1, &lowerHalfPairFacet2)
	}

	crown := scn.FacetStructure{
		Name:     "Crown",
		Material: &diamondMaterial,
		Facets:   []*scn.Facet{},
	}

	crown.Facets = append(crown.Facets, &table)
	crown.Facets = append(crown.Facets, starFacets...)
	crown.Facets = append(crown.Facets, bezelFacets...)
	crown.Facets = append(crown.Facets, upperHalfPairFacets...)

	pavilion := scn.FacetStructure{
		Name:     "Pavilion",
		Material: &diamondMaterial,
		Facets:   []*scn.Facet{},
	}

	pavilion.Facets = append(pavilion.Facets, lowerHalfPairFacets...)
	pavilion.Facets = append(pavilion.Facets, pavilionMainFacets...)

	diamond := scn.FacetStructure{
		Name: "Diamond",
		//FacetStructures: []*scn.FacetStructure{&crown},
		// FacetStructures: []*scn.FacetStructure{&pavilion},
		FacetStructures: []*scn.FacetStructure{&crown, &pavilion},
	}

	// Girdle thickness:
	// The girdle should not be too thin or too thick on the edges.
	// The average girdle thickness percentage is calculated as a percentage of the girdle diameter.
	// Thin is <1 %, Medium is 1%- 3% and Thick is 4%<

	diamond.UpdateNormals()
	diamond.ScaleUniform(&vec3.Zero, scale)
	diamond.UpdateBounds()

	fmt.Printf("Crown bounds:    %+v\n", crown.Bounds)
	fmt.Printf("Pavilion bounds: %+v\n", pavilion.Bounds)
	fmt.Printf("Diamond bounds:  %+v\n", diamond.Bounds)

	return &diamond
}

func calculateLowerHalfTipPoints(amountTableCorners int, pavilionDepth float64, lowerHalfFacetSizeRelativeGirdleRadius float64, girdleDiameter float64, girdleBezelPoints []*vec3.T, pavilionPoint *vec3.T) []*vec3.T {
	var lowerHalfTipPoints []*vec3.T
	amountPavilionMainFacets := amountTableCorners
	for pavilionMainFacetIndex := 0; pavilionMainFacetIndex < amountPavilionMainFacets; pavilionMainFacetIndex++ {
		pavilionMainAngleProgress := float64(pavilionMainFacetIndex) / float64(amountPavilionMainFacets)
		pavilionMainAngle := pavilionMainAngleProgress * 2.0 * math.Pi
		pavilionMainSidePointAngle := pavilionMainAngle + math.Pi/8.0

		girdleEdgePoint := girdleBezelPoints[pavilionMainFacetIndex]
		pavilionMainFacetPlane := pavilionMainFacetPlane(pavilionPoint, girdleEdgePoint)

		girdleRadius := girdleDiameter / 2.0
		x2 := girdleRadius * (1.0 - lowerHalfFacetSizeRelativeGirdleRadius) * math.Cos(pavilionMainSidePointAngle)
		y2 := pavilionDepth - pavilionDepth
		z2 := girdleRadius * (1.0 - lowerHalfFacetSizeRelativeGirdleRadius) * math.Sin(pavilionMainSidePointAngle)
		lowerHalfFacetTip := &vec3.T{x2, y2, z2}

		pavilionMainFacetSidePoint := verticalLinePlaneIntersection(pavilionMainFacetPlane, lowerHalfFacetTip)
		lowerHalfTipPoints = append(lowerHalfTipPoints, pavilionMainFacetSidePoint)
	}
	return lowerHalfTipPoints
}

// The tip of the star facet is completely dependent of the bezel facet.
// Hence, a lot of bezel calculations.
func calculateStarTipPoints(tableFacetPoints []*vec3.T, girdleBezelPoints []*vec3.T, tableRadius float64, girdleToTableLength float64, starSizeRelativeCrownSide, crownHeight float64) []*vec3.T {
	var starTipPoints []*vec3.T
	amountBezelFacets := len(tableFacetPoints)
	for bezelFacetIndex := 0; bezelFacetIndex < amountBezelFacets; bezelFacetIndex++ {
		bezelAngleProgress := float64(bezelFacetIndex) / float64(amountBezelFacets)
		bezelAngle := bezelAngleProgress * 2.0 * math.Pi
		bezelSidePointAngle := bezelAngle + math.Pi/8.0

		tableFacetPoint := tableFacetPoints[bezelFacetIndex]
		girdleEdgePoint := girdleBezelPoints[bezelFacetIndex]
		bezelFacetPlane := bezelFacetPlane(tableFacetPoint, girdleEdgePoint)

		starInnerRadius := tableRadius
		starOuterRadius := starInnerRadius + starSizeRelativeCrownSide*girdleToTableLength
		x2 := starOuterRadius * math.Cos(bezelSidePointAngle)
		y2 := crownHeight - crownHeight
		z2 := starOuterRadius * math.Sin(bezelSidePointAngle)
		starFacetTip := &vec3.T{x2, y2, z2}

		bezelFacetSidePoint := verticalLinePlaneIntersection(bezelFacetPlane, starFacetTip)

		starTipPoints = append(starTipPoints, bezelFacetSidePoint)
	}
	return starTipPoints
}

func calculateGirdleBezelPoints(amountTableCorners int, girdleDiameter float64, crownHeight float64) []*vec3.T {
	var girdleBezelPoints []*vec3.T
	amountBezelFacets := amountTableCorners
	for bezelFacetIndex := 0; bezelFacetIndex < amountBezelFacets; bezelFacetIndex++ {
		girdleEdgePoint := girdleEdgePoint(bezelFacetIndex, amountBezelFacets, girdleDiameter, crownHeight)
		girdleBezelPoints = append(girdleBezelPoints, &girdleEdgePoint)
	}
	return girdleBezelPoints
}

func calculateGirdleUpperHalfPoints(amountTableCorners int, girdleDiameter float64, crownHeight float64) []*vec3.T {
	var girdleUpperHalfPoints []*vec3.T
	amountBezelFacetSets := amountTableCorners
	for girdlePointIndex := 0; girdlePointIndex < amountBezelFacetSets; girdlePointIndex++ {
		bezelTipAngleProgress := float64(girdlePointIndex) / float64(amountBezelFacetSets)
		pointAngle := bezelTipAngleProgress*2.0*math.Pi + math.Pi/8.0
		x3 := (girdleDiameter / 2.0) * math.Cos(pointAngle)
		y3 := crownHeight - crownHeight
		z3 := (girdleDiameter / 2.0) * math.Sin(pointAngle)
		girdleEdgePoint := vec3.T{x3, y3, z3}

		girdleUpperHalfPoints = append(girdleUpperHalfPoints, &girdleEdgePoint)
	}
	return girdleUpperHalfPoints
}

func calculateTableFacetPoints(amountTableCorners int, tableRadius float64, crownHeight float64) []*vec3.T {
	var tableFacetPoints []*vec3.T
	for tableCornerIndex := 0; tableCornerIndex < amountTableCorners; tableCornerIndex++ {
		tableCornerAngleProgress := float64(tableCornerIndex) / float64(amountTableCorners)
		x := tableRadius * math.Cos(tableCornerAngleProgress*2.0*math.Pi)
		z := tableRadius * math.Sin(tableCornerAngleProgress*2.0*math.Pi)
		tableFacetPoints = append(tableFacetPoints, &vec3.T{x, crownHeight, z})
	}
	return tableFacetPoints
}

func girdleEdgePoint(girdlePointIndex int, amountGirdlePoints int, girdleDiameter float64, crownHeight float64) vec3.T {
	bezelTipAngleProgress := float64(girdlePointIndex) / float64(amountGirdlePoints)
	x3 := (girdleDiameter / 2.0) * math.Cos(bezelTipAngleProgress*2.0*math.Pi)
	y3 := crownHeight - crownHeight
	z3 := (girdleDiameter / 2.0) * math.Sin(bezelTipAngleProgress*2.0*math.Pi)
	girdleEdgePoint := vec3.T{x3, y3, z3}
	return girdleEdgePoint
}

func verticalLinePlaneIntersection(plane *scn.Plane, linePoint *vec3.T) *vec3.T {
	lp := *linePoint

	lp[1] = -1000 //* 1000 * 1000 //* 1000 // Really, really large negative number, oh boy...
	line := scn.Ray{Origin: &lp, Heading: &vec3.UnitY}
	bfp, intersection := scn.GetLinePlaneIntersectionPoint2(&line, plane)
	if !intersection {
		fmt.Printf("no plane intersection\n")
	}

	return bfp
}

func bezelFacetPlane(tableFacetPoint, girdleEdgePoint *vec3.T) *scn.Plane {
	v1 := girdleEdgePoint.Subed(tableFacetPoint)
	n := bezelFacetPlaneNormal(&v1)

	return &scn.Plane{Origin: tableFacetPoint, Normal: n}
}

// Vector v1 going straight through bezel facet, from table facet edge to girdle edge
func bezelFacetPlaneNormal(v1 *vec3.T) *vec3.T {
	v1n := v1.Normalized()
	v1 = &v1n
	v2 := vec3.T{0, 1, 0}
	v3 := vec3.Cross(v1, &v2)
	n := vec3.Cross(&v3, v1)
	n.Normalize()

	return &n
}

func pavilionMainFacetPlane(pavilionPoint, girdleEdgePoint *vec3.T) *scn.Plane {
	v1 := girdleEdgePoint.Subed(pavilionPoint)
	n := pavilionMainFacetPlaneNormal(&v1)

	return &scn.Plane{Origin: girdleEdgePoint, Normal: n}
}

// Vector v1 going straight through pavilion main facet, from pavilion point to girdle edge
func pavilionMainFacetPlaneNormal(v1 *vec3.T) *vec3.T {
	v1n := v1.Normalized()
	v1 = &v1n
	v2 := &vec3.T{0, -1, 0}
	v3 := vec3.Cross(v1, v2)
	n := vec3.Cross(&v3, v1)
	n.Normalize()

	return &n
}

func deg2rad(degrees float64) float64 {
	return degrees * (math.Pi / 180.0)
}
