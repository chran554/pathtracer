package diamond

import (
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"math"
	"os"
	scn "pathtracer/internal/pkg/scene"
)

type Diamond struct {
	GirdleDiameter                     float64 // GirdleDiameter is the diameter of the round diamond as seen from above. All facet sizes of a diamond are given in percentage of the girdle (diamond) width. Thus, we set girdle size as 100%.
	GirdleHeightRelativeGirdleDiameter float64 // GirdleHeightRelativeGirdleDiameter is girdle height in percent of girdle radius. The height is measured at the girdle highest point (i.e. at the point where bezel facets and pavilion main facets point to each other). At least nn percent for "razor sharp" edges. nn% = "THIN", nn% = "MEDUIM" ..

	CrownAngleDegrees              float64 // CrownAngleDegrees is the angle of the crown side, from the girdle to the table, as when you see the diamond from the side. (value is typically 33 degrees - 35 degrees). According to Marcel Tolkowsky, a 40.75-degree pavilion angle gets perfectly paired with a 34.5-degree crown angle. The recommended crown angle ranges from 34 to 35 degrees.
	TableFacetSizeRelativeGirdle   float64 // TableFacetSizeRelativeGirdle is the size of the tablet. It is expressed as a percentage of the Girdle diameter (value is typically 53%-58%, 56% might be a good start)
	StarFacetSizeRelativeCrownSide float64 // StarFacetSizeRelativeCrownSide is the size of the star facets. It is expressed as percentage of the length of the side of the crown. <crown side width> = (<girdle width> - <table width>)/2.0 (value is typically 50%-60%)

	PavilionAngleDegrees                   float64 // PavilionAngleDegrees According to Marcel Tolkowsky, the optimal pavilion angle degree is 40.75. However, unless you are looking for a super-ideal diamond cut with a pavilion angle of 40.9 degrees, opting for stone within a range of 40.6 – 41 degrees is safe, providing other parameters meet their recommended ranges.
	LowerHalfFacetSizeRelativeGirdleRadius float64 // // LowerHalfFacetSizeRelativeGirdleRadius is the size of the lower half facets. It is expressed as percentage of the length of the side of the pavilion.
}

// GetDiamondRoundBrilliantCut will return a facet structure of a round, brilliant cut, diamond.
// The diamond is a 57 facet, brilliant cut, diamond.
//
// It has a faceted girdle. The facets of the girdle is not included in the 57 facet count.
// The girdle is neither "dug out" nor "painted" but "normal".
// (The girdle itself consist of 32 facets. Two girdle facets per each "upper"/"lower" half pair facet.)
//
// It has no culet (no 58:th facet), at the bottom of the diamond, but is a "pointed" or "None" culet cut.
func NewDiamondRoundBrilliantCut(d Diamond, scale float64, material scn.Material) *scn.FacetStructure {
	// The girdle diameter of the diamond is the average widest diameter of the diamond when viewed from above.
	// The girdle diameter measurement is key as it determines the proportions of the brilliant cut diamond.
	// Most percentage proportions of the brilliant cut round diamond are calculated as percentages of the girdle diameter.

	// girdleDiameter := 1.0                          // girdleDiameter is the diameter of the round diamond as seen from above. All facet sizes of a diamond are given in percentage of the girdle (diamond) width. Thus, we set girdle size as 100%.
	// crownAngleDegrees := 34.0                      // crownAngleDegrees is the angle of the crown side, from the girdle to the table, as when you see the diamond from the side. (value is typically 33 degrees - 35 degrees). According to Marcel Tolkowsky, a 40.75-degree pavilion angle gets perfectly paired with a 34.5-degree crown angle. The recommended crown angle ranges from 34 to 35 degrees.
	// tableFacetSizeRelativeGirdle := 0.56           // tableFacetSizeRelativeGirdle is the size of the tablet. It is expressed as a percentage of the Girdle diameter (value is typically 53%-58%, 56% might be a good start)
	// starFacetSizeRelativeCrownSide := 0.55         // starFacetSizeRelativeCrownSide is the size of the star facets. It is expressed as percentage of the length of the side of the crown. <crown side width> = (<girdle width> - <table width>)/2.0 (value is typically 50%-60%)
	// pavilionAngleDegrees := 40.9                   // According to Marcel Tolkowsky, the optimal pavilion angle degree is 40.75. However, unless you are looking for a super-ideal diamond cut with a pavilion angle of 40.9 degrees, opting for stone within a range of 40.6 – 41 degrees is safe, providing other parameters meet their recommended ranges.
	// lowerHalfFacetSizeRelativeGirdleRadius := 0.77 //
	// girdleHeightRelativeGirdleRadius := 0.03       // Girdle height in percent of girdle radius. At least nn percent for "razor sharp" edges. Thin is less than 1%=0.01, Medium is between 1%-3% and Thick is more 4%<
	culetDiameter := 0.0 // Do not change this from value 0.0. Culet facet is not included in 3D model. Only diamonds with "pointed" or "None" culet are created.

	// Create crown

	// Pre calculations
	girdleRadius := d.GirdleDiameter / 2.0
	culetRadius := culetDiameter / 2.0
	tableRadius := girdleRadius * d.TableFacetSizeRelativeGirdle
	girdleToTableLength := girdleRadius - tableRadius
	crownSideLength := girdleToTableLength / math.Cos(deg2rad(d.CrownAngleDegrees))
	crownHeight := crownSideLength * math.Sin(deg2rad(d.CrownAngleDegrees))

	girdleToCuletLength := girdleRadius - culetRadius
	pavilionSideLength := girdleToCuletLength / math.Cos(deg2rad(d.PavilionAngleDegrees))
	pavilionHeight := pavilionSideLength * math.Sin(deg2rad(d.PavilionAngleDegrees))

	fmt.Printf("Girdle diameter: %.2f\n", d.GirdleDiameter)
	fmt.Printf("Table diameter:  %.2f%%\n", tableRadius*2.0*100)
	fmt.Printf("Crown height:    %.2f\n", crownHeight)
	fmt.Printf("Pavilion height: %.2f\n", pavilionHeight)
	fmt.Printf("Total height:    %.2f\n", pavilionHeight+crownHeight+(d.GirdleHeightRelativeGirdleDiameter*d.GirdleDiameter))

	amountTableCorners := 8
	tableFacetPoints := calculateTableFacetPoints(amountTableCorners, tableRadius, crownHeight)
	girdleBezelPoints := calculateGirdlePoints(amountTableCorners, d.GirdleDiameter, 0.0)
	upperHalfFacetGirdlePoints := calculateGirdleHalfPoints(amountTableCorners, d.GirdleDiameter, 0.0)
	lowerHalfFacetGirdlePoints := calculateGirdleHalfPoints(amountTableCorners, d.GirdleDiameter, -d.GirdleDiameter*d.GirdleHeightRelativeGirdleDiameter)
	girdlePavilionMainPoints := calculateGirdlePoints(amountTableCorners, d.GirdleDiameter, -d.GirdleDiameter*d.GirdleHeightRelativeGirdleDiameter)
	starTipPoints := calculateStarTipPoints(tableFacetPoints, girdleBezelPoints, tableRadius, girdleToTableLength, d.StarFacetSizeRelativeCrownSide, crownHeight)
	girdleUpperPoints := make([]*vec3.T, amountTableCorners*4)
	girdleLowerPoints := make([]*vec3.T, amountTableCorners*4)

	// Update girdle upper points
	for i, point := range girdleBezelPoints {
		girdleUpperPoints[i*4] = point
	}
	for i, point := range upperHalfFacetGirdlePoints {
		girdleUpperPoints[i*4+2] = point
	}

	// Update girdle lower points
	for i, point := range girdlePavilionMainPoints {
		girdleLowerPoints[i*4] = point
	}
	for i, point := range lowerHalfFacetGirdlePoints {
		girdleLowerPoints[i*4+2] = point
	}

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
	var upperHalfFacets []*scn.Facet
	amountUpperHalfFacetPairs := amountTableCorners
	for upperHalfFacetPairIndex := 0; upperHalfFacetPairIndex < amountUpperHalfFacetPairs; upperHalfFacetPairIndex++ {
		pb1 := girdleBezelPoints[upperHalfFacetPairIndex]
		pb2 := girdleBezelPoints[(upperHalfFacetPairIndex+1)%amountUpperHalfFacetPairs]
		ps := starTipPoints[upperHalfFacetPairIndex]
		p2 := upperHalfFacetGirdlePoints[upperHalfFacetPairIndex]

		upperHalfPairBaseAngle := (float64(upperHalfFacetPairIndex) / float64(amountUpperHalfFacetPairs)) * (2.0 * math.Pi)
		upperHalfPairSubAngleIncrement := (2.0 * math.Pi) / (float64(amountUpperHalfFacetPairs) * 4.0) // 5 girdle corners for each upper half pair facets set

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

		girdleUpperPoints[upperHalfFacetPairIndex*4+1] = p1
		girdleUpperPoints[upperHalfFacetPairIndex*4+3] = p3
		upperHalfFacets = append(upperHalfFacets, &upperHalfPairFacet1, &upperHalfPairFacet2)
	}

	// Pavilion facets

	pavilionTipPoint := &vec3.T{0, -pavilionHeight, 0}

	lowerHalfTipPoints := calculateLowerHalfTipPoints(amountTableCorners, pavilionHeight, d.LowerHalfFacetSizeRelativeGirdleRadius, d.GirdleDiameter, girdlePavilionMainPoints, pavilionTipPoint)

	// Pavilion - Main facets (the kite like, 4 sided, facet of the pavilion)
	var pavilionMainFacets []*scn.Facet
	amountPavilionMainFacets := amountTableCorners
	for pavilionMainFacetIndex := 0; pavilionMainFacetIndex < amountPavilionMainFacets; pavilionMainFacetIndex++ {
		pavilionMainFacet := scn.Facet{
			Vertices: []*vec3.T{
				girdlePavilionMainPoints[pavilionMainFacetIndex],
				lowerHalfTipPoints[(pavilionMainFacetIndex+amountPavilionMainFacets-1)%amountPavilionMainFacets],
				pavilionTipPoint,
				lowerHalfTipPoints[pavilionMainFacetIndex],
			},
		}
		pavilionMainFacets = append(pavilionMainFacets, &pavilionMainFacet)
	}

	// Pavilion - Lower half pair facets
	var lowerHalfFacets []*scn.Facet
	amountLowerHalfPairFacets := amountTableCorners
	for lowerHalfFacetPairIndex := 0; lowerHalfFacetPairIndex < amountLowerHalfPairFacets; lowerHalfFacetPairIndex++ {
		pb1 := girdlePavilionMainPoints[lowerHalfFacetPairIndex]
		pb2 := girdlePavilionMainPoints[(lowerHalfFacetPairIndex+1)%amountLowerHalfPairFacets]
		ps := lowerHalfTipPoints[lowerHalfFacetPairIndex]
		p2 := lowerHalfFacetGirdlePoints[lowerHalfFacetPairIndex]

		lowerHalfPairBaseAngle := (float64(lowerHalfFacetPairIndex) / float64(amountLowerHalfPairFacets)) * (2.0 * math.Pi)
		lowerHalfPairSubAngleIncrement := (2.0 * math.Pi) / (float64(amountLowerHalfPairFacets) * 4.0) // 5 girdle corners for each upper half pair facets set

		plane1 := scn.NewPlane(ps, pb1, p2, "", nil)
		gp1Angle := lowerHalfPairBaseAngle + 1.0*lowerHalfPairSubAngleIncrement
		gp1 := &vec3.T{girdleRadius * math.Cos(gp1Angle), 0, girdleRadius * math.Sin(gp1Angle)}
		p1 := verticalLinePlaneIntersection(plane1, gp1)

		plane2 := scn.NewPlane(ps, p2, pb2, "", nil)
		gp2Angle := lowerHalfPairBaseAngle + 3.0*lowerHalfPairSubAngleIncrement
		gp2 := &vec3.T{girdleRadius * math.Cos(gp2Angle), 0, girdleRadius * math.Sin(gp2Angle)}
		p3 := verticalLinePlaneIntersection(plane2, gp2)

		lowerHalfPairFacet1 := scn.Facet{Vertices: []*vec3.T{ps, p2, p1, pb1}}
		lowerHalfPairFacet2 := scn.Facet{Vertices: []*vec3.T{ps, pb2, p3, p2}}

		girdleLowerPoints[lowerHalfFacetPairIndex*4+1] = p1
		girdleLowerPoints[lowerHalfFacetPairIndex*4+3] = p3
		lowerHalfFacets = append(lowerHalfFacets, &lowerHalfPairFacet1, &lowerHalfPairFacet2)
	}

	// Girdle
	var girdleFacets []*scn.Facet
	amountGirdleFacets := len(girdleUpperPoints)
	for i := 0; i < amountGirdleFacets; i++ {
		girdleFacet := scn.Facet{Vertices: []*vec3.T{girdleUpperPoints[i], girdleLowerPoints[i], girdleLowerPoints[(i+1)%amountGirdleFacets], girdleUpperPoints[(i+1)%amountGirdleFacets]}}
		girdleFacets = append(girdleFacets, &girdleFacet)
	}

	// Check that valley position heights on girdle is still positive lengths.
	// Report to std out the valley position heights (and hill heights) in percent of girdle size.
	girdleHeightAtHillPosition := girdleFacets[0].Vertices[0][1] - girdleFacets[0].Vertices[1][1]
	girdleHeightAtValleyPosition := girdleFacets[0].Vertices[3][1] - girdleFacets[0].Vertices[2][1]
	girdleRelativeHeightAtHillPosition := girdleHeightAtHillPosition / d.GirdleDiameter
	girdleRelativeHeightAtValleyPosition := girdleHeightAtValleyPosition / d.GirdleDiameter
	minimGirdleHeight := girdleHeightAtHillPosition - girdleHeightAtValleyPosition
	minimGirdleRelativeHeight := minimGirdleHeight / d.GirdleDiameter
	fmt.Println()
	fmt.Printf("Girdle height at hill position %.2f%% (and height at valley positions %.2f%%).\n", girdleRelativeHeightAtHillPosition*100.0, girdleRelativeHeightAtValleyPosition*100.0)
	fmt.Printf("Minimum girdle height for a diamond with these propprtions is %.3f%% (at hill positions).\n", minimGirdleRelativeHeight*100.0)

	if girdleHeightAtValleyPosition < 0.0 {
		fmt.Printf("Resulting girdle height at valley positions %.3f%% is negative, which physically impossible. Increase girdle height to at least %.3f%%.\n", girdleRelativeHeightAtValleyPosition*100.0, minimGirdleRelativeHeight*100.0)
		os.Exit(1)
	}

	// Diamond assembly

	crown := scn.FacetStructure{
		SubstructureName: "Crown",
		Facets:           []*scn.Facet{},
	}

	crown.Facets = append(crown.Facets, &table)
	crown.Facets = append(crown.Facets, starFacets...)
	crown.Facets = append(crown.Facets, bezelFacets...)
	crown.Facets = append(crown.Facets, upperHalfFacets...)

	girdle := scn.FacetStructure{
		SubstructureName: "Girdle",
		Facets:           []*scn.Facet{},
	}

	girdle.Facets = append(girdle.Facets, girdleFacets...)

	pavilion := scn.FacetStructure{
		SubstructureName: "Pavilion",
		Facets:           []*scn.Facet{},
	}

	pavilion.Facets = append(pavilion.Facets, lowerHalfFacets...)
	pavilion.Facets = append(pavilion.Facets, pavilionMainFacets...)

	diamond := scn.FacetStructure{
		Name:            "Diamond",
		Material:        &material,
		FacetStructures: []*scn.FacetStructure{&crown, &pavilion, &girdle},
	}

	// Girdle thickness:
	// The girdle should not be too thin or too thick on the edges.
	// The average girdle thickness percentage is calculated as a percentage of the girdle diameter.
	// Thin is <1 %, Medium is 1%- 3% and Thick is 4%<

	diamond.UpdateNormals()
	diamond.ScaleUniform(&vec3.Zero, scale)
	diamond.UpdateBounds()

	// fmt.Printf("Crown bounds:    %+v\n", crown.Bounds)
	// fmt.Printf("Girdle bounds:   %+v\n", girdle.Bounds)
	// fmt.Printf("Pavilion bounds: %+v\n", pavilion.Bounds)
	// fmt.Printf("Diamond bounds:  %+v\n", diamond.Bounds)

	return &diamond
}

func calculateLowerHalfTipPoints(amountTableCorners int, pavilionDepth float64, lowerHalfFacetSizeRelativeGirdleRadius float64, girdleDiameter float64, girdlePavilionMainPoints []*vec3.T, pavilionPoint *vec3.T) []*vec3.T {
	var lowerHalfTipPoints []*vec3.T
	amountPavilionMainFacets := amountTableCorners
	for pavilionMainFacetIndex := 0; pavilionMainFacetIndex < amountPavilionMainFacets; pavilionMainFacetIndex++ {
		pavilionMainAngleProgress := float64(pavilionMainFacetIndex) / float64(amountPavilionMainFacets)
		pavilionMainAngle := pavilionMainAngleProgress * 2.0 * math.Pi
		pavilionMainSidePointAngle := pavilionMainAngle + math.Pi/8.0

		girdleEdgePoint := girdlePavilionMainPoints[pavilionMainFacetIndex]
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

func calculateGirdlePoints(amountPoints int, girdleDiameter float64, yPosition float64) []*vec3.T {
	var girdlePoints []*vec3.T
	for pointIndex := 0; pointIndex < amountPoints; pointIndex++ {
		girdleEdgePoint := girdlePoint(pointIndex, amountPoints, girdleDiameter, yPosition)
		girdlePoints = append(girdlePoints, &girdleEdgePoint)
	}
	return girdlePoints
}

func girdlePoint(girdlePointIndex int, amountPoints int, girdleDiameter float64, yPosition float64) vec3.T {
	bezelTipAngleProgress := float64(girdlePointIndex) / float64(amountPoints)
	x3 := (girdleDiameter / 2.0) * math.Cos(bezelTipAngleProgress*2.0*math.Pi)
	y3 := yPosition
	z3 := (girdleDiameter / 2.0) * math.Sin(bezelTipAngleProgress*2.0*math.Pi)
	girdleEdgePoint := vec3.T{x3, y3, z3}
	return girdleEdgePoint
}

func calculateGirdleHalfPoints(amountTableCorners int, girdleDiameter float64, yPosition float64) []*vec3.T {
	var girdleUpperHalfPoints []*vec3.T
	amountBezelFacetSets := amountTableCorners
	for girdlePointIndex := 0; girdlePointIndex < amountBezelFacetSets; girdlePointIndex++ {
		bezelTipAngleProgress := float64(girdlePointIndex) / float64(amountBezelFacetSets)
		pointAngle := bezelTipAngleProgress*2.0*math.Pi + math.Pi/8.0
		x3 := (girdleDiameter / 2.0) * math.Cos(pointAngle)
		y3 := yPosition
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
