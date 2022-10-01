package main

import (
	"fmt"
	"github.com/ungerik/go3d/float64/mat3"
	"math"
	"os"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/scene"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "facetobj_test"

var amountImages = 180

var imageWidth = 1024 / 2
var imageHeight = 800 / 2
var magnification = 1.0 / 2

var renderType = scn.Pathtracing

var amountSamples = 100
var maxRecursion = 2

var objectFilename = "cube_smooth.obj"
var objectScale = 2.5
var objectStartAngle = 0.0

//var objectFilename = "human skull.obj"
//var objectScale = 10000.0 * 1.5
//var objectStartAngle = math.Pi // 180 degrees

//var environmentEnvironMap = "textures/equirectangular/canyon 3200x1600.jpeg"
//var environmentRadius = 100.0 * 10.0 // 10m (if 1 unit is 1 cm)
//var environmentEmissionFactor = float32(2.0)

var environmentEnvironMap = "textures/equirectangular/white room 01 1836x918.png"
var environmentRadius = 100.0 * 20.0 // 20m (if 1 unit is 1 cm)
var environmentEmissionFactor = float32(1.0)

//var environmentEnvironMap = "textures/equirectangular/spruit_sunrise_2400x1200.jpeg"
//var environmentRadius = 100.0 * 400.0 // 200m (if 1 unit is 1 cm)
//var environmentEmissionFactor = float32(3.0)

var useLights = true
var lightColor = color.Color{R: 1.0, G: 0.97, B: 0.95}
var lightOrigin = vec3.T{environmentRadius / 8, environmentRadius / 4, -environmentRadius / 1.5}
var lightRadius = environmentRadius / 14
var lightEmissionFactor = float32(2)

var cameraDistance = 400.0
var cameraOrigin = vec3.T{cameraDistance / 1.0, cameraDistance * 4 / 5, -cameraDistance}
var useAimPoint = true

var viewPlaneDistance = 500.0
var lensRadius = 0.0

func main() {
	// filename := "objects/lamp_post.obj.3ds.obj"
	// filename := "objects/Diamond.obj"
	//filename := "/Users/christian/projects/code/go/pathtracer/objects/go_gopher_color.obj"
	filename := "/Users/christian/projects/code/go/pathtracer/objects/" + objectFilename
	// filename := "/Users/christian/projects/code/go/pathtracer/objects/cube_smooth.obj"
	// filename := "/Users/christian/projects/code/go/pathtracer/objects/unit_cube.obj"
	//filename := "/Users/christian/projects/code/go/pathtracer/objects/facet.obj"
	// filename := "/Users/christian/projects/code/go/pathtracer/objects/triangle.obj"
	// filename := "objects/go_gopher_high.obj"

	frames := []scene.Frame{}
	for imageIndex := 0; imageIndex < amountImages; imageIndex++ {

		fmt.Printf("\n\nCostructing frame %d\n", imageIndex)
		fmt.Printf("Reading file: %s\n", filename)

		var err error

		/*wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		parentPath := filepath.Dir(wd)
		fmt.Println(parentPath)*/

		f, err := os.Open(filename)
		if err != nil {
			fmt.Printf("ouupps, something went wrong loading file: '%s'\n%s\n", filename, err.Error())
		}

		facetStructure, err := obj.Read(f)
		if err != nil {
			fmt.Printf("ouups, something went wrong parsing object. %s\n", err.Error())
		}
		facetStructure.UpdateBounds()
		fmt.Printf("Object in file \"%s\" has bounds %+v.\n", objectFilename, facetStructure.Bounds)

		facetStructure.Scale(&vec3.Zero, objectScale)
		facetStructure.RotateY(&vec3.Zero, objectStartAngle)

		sceneNode := scn.SceneNode{
			Spheres:         nil,
			Discs:           nil,
			ChildNodes:      nil,
			FacetStructures: []*scn.FacetStructure{facetStructure},
			Bounds:          nil,
		}

		updateBoundingBoxes(&sceneNode)
		sceneNode.Bounds = nil

		if useLights {
			addLights(&sceneNode)
		}

		addEnvironmentMapping(environmentEnvironMap, &sceneNode)

		animationProgress := float64(imageIndex) / float64(amountImages)
		heightFactor := math.Sin(2.0 * 2.0 * math.Pi * animationProgress)
		camera := getCamera(facetStructure.Bounds.Center(), 2.0*math.Pi*animationProgress, heightFactor)

		frame := scene.Frame{
			Filename:   "facetobj" + "_" + fmt.Sprintf("%06d", imageIndex),
			FrameIndex: imageIndex,
			Camera:     &camera,
			SceneNode:  &sceneNode,
		}

		frames = append(frames, frame)
	}

	animation := scene.Animation{
		AnimationName:     animationName,
		Frames:            frames,
		Width:             imageWidth,
		Height:            imageHeight,
		WriteRawImageFile: true,
	}

	anm.WriteAnimationToFile(animation, false)
}

func updateBoundingBoxes(s *scene.SceneNode) {
	for _, structure := range s.FacetStructures {
		structure.UpdateBounds()
	}

	for _, childNode := range s.ChildNodes {
		childNode.GetBounds()
	}

	s.Bounds = nil
}

func getCamera(aimPoint *vec3.T, yRotationAngle float64, heightFactor float64) scn.Camera {
	rotationMatrix := mat3.T{}
	rotationMatrix.AssignYRotation(yRotationAngle)

	newCameraOrigin := vec3.T{cameraOrigin[0], cameraOrigin[1] * heightFactor, cameraOrigin[2]}
	newCameraOrigin = rotationMatrix.MulVec3(&newCameraOrigin)

	heading := vec3.T{-newCameraOrigin[0], -newCameraOrigin[1], -newCameraOrigin[2]} // aim for origin
	if useAimPoint {
		heading = aimPoint.Subed(&newCameraOrigin)
	}

	focalDistance := heading.Length()

	return scn.Camera{
		Origin:            &newCameraOrigin,
		Heading:           &heading,
		ViewUp:            &vec3.T{0, 1, 0},
		ViewPlaneDistance: viewPlaneDistance,
		LensRadius:        lensRadius,
		FocalDistance:     focalDistance,
		Samples:           amountSamples,
		AntiAlias:         true,
		Magnification:     magnification,
		RenderType:        renderType,
		RecursionDepth:    maxRecursion,
	}
}

func addLights(scene *scn.SceneNode) {
	sphere := scn.Sphere{
		Name:   "Light",
		Origin: lightOrigin,
		Radius: lightRadius,
		Material: &scn.Material{
			Color:         lightColor,
			Emission:      &color.Color{R: lightColor.R * lightEmissionFactor, G: lightColor.G * lightEmissionFactor, B: lightColor.B * lightEmissionFactor},
			RayTerminator: true,
		},
	}

	scene.Spheres = append(scene.Spheres, &sphere)
}

func addEnvironmentMapping(filename string, scene *scn.SceneNode) {
	origin := vec3.T{0, 0, 0}

	sphere := scn.Sphere{
		Name:   "Environment mapping",
		Origin: origin,
		Radius: environmentRadius,
		Material: &scn.Material{
			Color:         color.Color{R: 1.0, G: 1.0, B: 1.0},
			Emission:      &color.Color{R: 1.0 * environmentEmissionFactor, G: 1.0 * environmentEmissionFactor, B: 1.0 * environmentEmissionFactor},
			RayTerminator: false, // TODO true,
			Projection: &scn.ImageProjection{
				ProjectionType: scn.Spherical,
				ImageFilename:  filename,
				Gamma:          1.5,
				Origin:         origin,
				U:              vec3.T{0, 0, 1},
				V:              vec3.T{0, 1, 0},
				RepeatU:        true,
				RepeatV:        true,
				FlipU:          false,
				FlipV:          false,
			},
		},
	}

	scene.Spheres = append(scene.Spheres, &sphere)
}
