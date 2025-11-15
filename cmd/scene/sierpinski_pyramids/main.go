package main

import (
	"fmt"
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	anm "pathtracer/internal/pkg/renderfile"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "sierpinski_pyramids"

var skyDomeRadius = 200.0 * 100.0 // radius
var skyDomeEmissionFactor = 1.0

var amountFrames = 360 * 2

var imageWidth = 1280
var imageHeight = 1024
var magnification = 0.5

var amountSamples = 256 // 1024 * 3

var pyramidSize = 4.5 * 100.0
var maxPyramidRecursionDepth = 8

var apertureSize = 3.3

var pyramidMaterial = scn.NewMaterial().C(color.NewColorGrey(0.75)).M(0.60, 0.08)

// Pyramid is represented by four 3D points.
//
//	                  p1
//		              '
//		             /=\ \
//		            /===\ \
//		           /=====\' \
//		          /=======\'' \
//		         /=========\ ' '\
//		        /===========\''   \
//		       /=============\ ' '  \
//		      /===============\   ''  \
//		     /=================\' ' ' ' \
//		    /===================\' ' '  ' \
//		   /=====================\' '   ' ' \  p4
//		  /=======================\  '   ' /
//		 /=========================\   ' /
//		/===========================\'  /
//	   /=============================\/
//	  p2                              p3
type Pyramid struct {
	p1         *vec3.T // top point
	p2, p3, p4 *vec3.T // base points
}

func (p *Pyramid) SierpinskiSubPyramids() []*Pyramid {
	v12 := p.p2.Subed(p.p1)
	v13 := p.p3.Subed(p.p1)
	v14 := p.p4.Subed(p.p1)

	v23 := p.p3.Subed(p.p2)
	v24 := p.p4.Subed(p.p2)
	v34 := p.p4.Subed(p.p3)

	p12 := v12.Scale(0.5).Add(p.p1) // Top split point (halfway from top p1 to base p2)
	p13 := v13.Scale(0.5).Add(p.p1) // Top split point (halfway from top p1 to base p3)
	p14 := v14.Scale(0.5).Add(p.p1) // Top split point (halfway from top p1 to base p4)

	p23 := v23.Scale(0.5).Add(p.p2) // Bottom split point (halfway from base p2 to base p3)
	p24 := v24.Scale(0.5).Add(p.p2) // Bottom split point (halfway from base p2 to base p4)
	p34 := v34.Scale(0.5).Add(p.p3) // Bottom split point (halfway from base p3 to base p4)

	return []*Pyramid{
		{p1: p.p1, p2: p12, p3: p13, p4: p14}, // Top sub pyramid
		{p1: p12, p2: p.p2, p3: p23, p4: p24}, // Bottom sub pyramid
		{p1: p13, p2: p23, p3: p.p3, p4: p34}, // Bottom sub pyramid
		{p1: p14, p2: p24, p3: p34, p4: p.p4}, // Bottom sub pyramid
	}
}

func main() {
	// var environmentEnvironMap = "textures/equirectangular/open_grassfield_sunny_day.jpg"
	// var environmentEnvironMap = "textures/equirectangular/5792766093_8153225334_o.jpg"
	//var environmentEnvironMap = "textures/equirectangular/white room 02 612x612.jpg"

	// var environmentEnvironMap = "textures/equirectangular/nightsky.png"
	environmentEnvironMap := floatimage.Load("textures/equirectangular/sunset horizon 2800x1400.jpg")

	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, true, false)

	for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountFrames)

		sierpinskiPyramid := getSierpinskiPyramid()

		sierpinskiPyramid.Scale(&vec3.T{0.0, 0.0, 0.0}, &vec3.T{pyramidSize, pyramidSize, pyramidSize})
		sierpinskiPyramid.Translate(&vec3.T{0.0, -pyramidSize / 2.0, 0.0})

		pyramidsBounds := sierpinskiPyramid.UpdateBounds()
		sierpinskiPyramid.Translate(&vec3.T{0, -pyramidsBounds.Ymin, 0})
		pyramidsBounds = sierpinskiPyramid.UpdateBounds()
		// fmt.Printf("Pyramid bounds: %+v   (center: %+v)\n", pyramidsBounds, pyramidsBounds.Center())

		a360 := 2.0 * math.Pi
		animationAngle := animationProgress * a360
		// recursivePyramids.RotateY(&vec3.Zero, animationAngle)

		// Sky dome
		skyDomeOrigin := vec3.T{0, 0, 0}
		skyDomeMaterial := scn.NewMaterial().
			E(color.White, skyDomeEmissionFactor, true).
			SP(environmentEnvironMap, &skyDomeOrigin, vec3.T{1, 0, -1}, vec3.T{0, 1, 0})
		skyDome := scn.NewSphere(&skyDomeOrigin, skyDomeRadius, skyDomeMaterial).N("Environment mapping")

		cameraOrigin := pyramidsBounds.Center().Add(&vec3.T{0, 150, -700})
		cameraFocusPoint := pyramidsBounds.Center().Add(&vec3.T{0, 0, -1.0 * (pyramidsBounds.SizeZ() / 2.0) * 0.8})
		camera := scn.NewCamera(cameraOrigin, cameraFocusPoint, amountSamples, magnification).A(apertureSize, nil).V(700)

		scene := scn.NewSceneNode().S(skyDome).SN(sierpinskiPyramid)

		scene.RotateY(pyramidsBounds.Center(), animationAngle)

		frame := scn.NewFrame(animationName, frameIndex, camera, scene)

		animation.Frames = append(animation.Frames, frame)
	}

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := anm.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}

func getSierpinskiPyramid() *scn.SceneNode {
	a360 := 2.0 * math.Pi
	a000 := 0.0
	a120 := a360 * 1.0 / 3.0
	a240 := a360 * 2.0 / 3.0

	v1 := math.Cos(a120)
	r1 := math.Sin(a120)

	p1 := &vec3.T{0.0, 0.0, 0.0} // Top point of the pyramid
	p2 := &vec3.T{r1 * math.Cos(a000), v1 - 1.0, r1 * math.Sin(a000)}
	p3 := &vec3.T{r1 * math.Cos(a120), v1 - 1.0, r1 * math.Sin(a120)}
	p4 := &vec3.T{r1 * math.Cos(a240), v1 - 1.0, r1 * math.Sin(a240)}

	startPyramid := &Pyramid{p1: p1, p2: p2, p3: p3, p4: p4}

	recursivePyramids := getRecursivePyramids(startPyramid, 1, maxPyramidRecursionDepth)

	// Pyramid has its top at origin.
	recursivePyramids.Translate(&vec3.T{0.0, -(v1 - 1.0), 0.0}) // Move pyramid so pyramid baseplate is centered around origin

	return recursivePyramids
}

// getRecursivePyramids gets a recursive sierpinski (3-sided) pyramid from an initial pyramid.
func getRecursivePyramids(pyramid *Pyramid, recursionDepth int, maxRecursionDepth int) *scn.SceneNode {
	scene := scn.SceneNode{}

	if recursionDepth == maxRecursionDepth {
		scene.FacetStructures = append(scene.FacetStructures, getPyramidFacetStructure(pyramid))
	} else {
		sierpinskiSubPyramids := pyramid.SierpinskiSubPyramids()
		for _, subPyramid := range sierpinskiSubPyramids {
			subPyramidsNode := getRecursivePyramids(subPyramid, recursionDepth+1, maxRecursionDepth)
			scene.ChildNodes = append(scene.ChildNodes, subPyramidsNode)
		}
	}

	return &scene
}

func getPyramidFacetStructure(pyramid *Pyramid) *scn.FacetStructure {
	pyramidFacets := &scn.FacetStructure{Name: "pyramid", Material: pyramidMaterial}

	pyramidFacets.Facets = append(pyramidFacets.Facets, &scn.Facet{Vertices: []*vec3.T{pyramid.p1, pyramid.p2, pyramid.p3}})
	pyramidFacets.Facets = append(pyramidFacets.Facets, &scn.Facet{Vertices: []*vec3.T{pyramid.p1, pyramid.p3, pyramid.p4}})
	pyramidFacets.Facets = append(pyramidFacets.Facets, &scn.Facet{Vertices: []*vec3.T{pyramid.p1, pyramid.p4, pyramid.p2}})
	pyramidFacets.Facets = append(pyramidFacets.Facets, &scn.Facet{Vertices: []*vec3.T{pyramid.p2, pyramid.p3, pyramid.p4}}) // Bottom facet

	pyramidFacets.UpdateNormals()
	pyramidFacets.UpdateBounds()

	return pyramidFacets
}
