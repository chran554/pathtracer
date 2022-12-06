package main

import (
	"fmt"
	"math"
	"os"
	anm "pathtracer/internal/pkg/animation"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/obj"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "lamp_post"

var environmentEnvironMap = "textures/equirectangular/sunset horizon 2800x1400.jpg"
var environmentRadius = 100.0 * 1000.0
var environmentEmissionFactor = float32(1.5)

var amountFrames = 1

var imageWidth = 1280
var imageHeight = 1024
var magnification = 1.0

var renderType = scn.Pathtracing
var amountSamples = 2000 * 2 * 4
var maxRecursion = 4

var viewPlaneDistance = 800.0
var apertureSize = 2.0

func main() {
	animation := getAnimation(int(float64(imageWidth)*magnification), int(float64(imageHeight)*magnification))
	animation.WriteRawImageFile = true

	for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountFrames)

		environmentSphere := addEnvironmentMapping(environmentEnvironMap)

		// Ground

		groundProjection := scn.NewParallelImageProjection("textures/ground/soil-cracked.png", vec3.Zero, vec3.UnitX.Scaled(150), vec3.UnitZ.Scaled(150))
		ground := scn.Disc{
			Name:   "Ground",
			Origin: &vec3.Zero,
			Normal: &vec3.UnitY,
			Radius: environmentRadius,
			Material: &scn.Material{
				Name:       "Ground material",
				Color:      &color.White,
				Emission:   &color.Black,
				Glossiness: 0.0,
				Roughness:  1.0,
				Projection: &groundProjection,
			},
		}

		// Gopher

		gopher := GetGopher(&vec3.T{50, 50, 50})
		gopher.RotateY(&vec3.Zero, math.Pi*5.0/6.0)
		gopher.Translate(&vec3.T{75, 0, 100})
		gopher.UpdateBounds()

		// Lamp post

		lampPostMaterial := scn.Material{
			Name:          "lamp post",
			Color:         &color.Color{R: 0.9, G: 0.4, B: 0.3},
			Emission:      &color.Black,
			Glossiness:    0.2,
			Roughness:     0.3,
			RayTerminator: false,
		}

		lampMaterial := scn.Material{
			Name:          "lamp",
			Color:         &color.Color{R: 1.0, G: 1.0, B: 1.0},
			Emission:      (&color.Color{R: 10.0, G: 10.0, B: 9.0}).Multiply(3.0),
			Glossiness:    0.0,
			Roughness:     1.0,
			RayTerminator: true,
		}

		lampPost := GetLampPost(&vec3.T{200, 200, 200})
		lampPost.ClearMaterials()
		lampPost.Material = &lampPostMaterial

		lampPost.GetFirstObjectByName("lamp_0").Material = &lampMaterial
		lampPost.GetFirstObjectByName("lamp_1").Material = &lampMaterial
		lampPost.GetFirstObjectByName("lamp_2").Material = &lampMaterial
		lampPost.GetFirstObjectByName("lamp_3").Material = &lampMaterial

		gopherBounds := gopher.Bounds
		cameraFocusPoint := gopherBounds.Center().Add(&vec3.T{0.0, lampPost.Bounds.SizeY() * 0.40, 0.0})
		cameraOrigin := gopherBounds.Center().Add(&vec3.T{0, 0.0, -300})
		camera := getCamera(magnification, animationProgress, cameraOrigin, cameraFocusPoint)

		scene := scn.SceneNode{
			Spheres:         []*scn.Sphere{&environmentSphere},
			Discs:           []*scn.Disc{&ground},
			ChildNodes:      []*scn.SceneNode{},
			FacetStructures: []*scn.FacetStructure{gopher, lampPost},
		}

		frame := scn.Frame{
			Filename:   animation.AnimationName + "_" + fmt.Sprintf("%06d", frameIndex),
			FrameIndex: frameIndex,
			Camera:     &camera,
			SceneNode:  &scene,
		}

		animation.Frames = append(animation.Frames, frame)
	}

	anm.WriteAnimationToFile(animation, false)
}

func GetGopher(scale *vec3.T) *scn.FacetStructure {
	var objFilename = "go_gopher_color.obj"
	var objFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/obj/" + objFilename

	objFile, err := os.Open(objFilenamePath)
	if err != nil {
		fmt.Printf("ouupps, something went wrong loading file: '%s'\n%s\n", objFilenamePath, err.Error())
	}
	defer objFile.Close()

	obj, err := obj.Read(objFile)
	ymin := obj.Bounds.Ymin
	ymax := obj.Bounds.Ymax
	obj.Translate(&vec3.T{0.0, -ymin, 0.0})       // feet touch the ground (xz-plane)
	obj.ScaleUniform(&vec3.Zero, 1.0/(ymax-ymin)) // resize to height == 1.0

	obj.Scale(&vec3.Zero, scale)

	obj.UpdateBounds()
	fmt.Printf("Gopher bounds: %+v\n", obj.Bounds)

	return obj
}

func GetLampPost(scale *vec3.T) *scn.FacetStructure {
	var objFilename = "lamp_post.obj"
	var objFilenamePath = "/Users/christian/projects/code/go/pathtracer/objects/obj/" + objFilename

	objFile, err := os.Open(objFilenamePath)
	if err != nil {
		fmt.Printf("ouupps, something went wrong loading file: '%s'\n%s\n", objFilenamePath, err.Error())
	}
	defer objFile.Close()

	obj, err := obj.Read(objFile)
	ymin := obj.Bounds.Ymin
	ymax := obj.Bounds.Ymax
	obj.Translate(&vec3.T{0.0, -ymin, 0.0})       // lamp post base touch the ground (xz-plane)
	obj.ScaleUniform(&vec3.Zero, 1.0/(ymax-ymin)) // resize to height == 1.0

	obj.Scale(&vec3.Zero, scale)

	obj.UpdateBounds()
	fmt.Printf("Lamp post bounds: %+v\n", obj.Bounds)

	return obj
}

func addEnvironmentMapping(filename string) scn.Sphere {
	origin := vec3.T{0, 0, 0}

	projection := scn.ImageProjection{
		ProjectionType: scn.Spherical,
		ImageFilename:  filename,
		Gamma:          1.5,
		Origin:         &origin,
		U:              &vec3.T{-0.2, 0, -1},
		V:              &vec3.T{0, 1, 0},
		RepeatU:        true,
		RepeatV:        true,
		FlipU:          false,
		FlipV:          false,
	}

	material := scn.Material{
		Color:         &color.Color{R: 1.0, G: 1.0, B: 1.0},
		Emission:      (&color.Color{R: 1.0, G: 1.0, B: 1.0}).Multiply(environmentEmissionFactor),
		RayTerminator: true,
		Projection:    &projection,
	}

	sphere := scn.Sphere{
		Name:     "Environment mapping",
		Origin:   &origin,
		Radius:   environmentRadius,
		Material: &material,
	}

	return sphere
}

func getAnimation(width int, height int) scn.Animation {
	animation := scn.Animation{
		AnimationName:     animationName,
		Frames:            []scn.Frame{},
		Width:             width,
		Height:            height,
		WriteRawImageFile: false,
	}
	return animation
}

func getCamera(magnification float64, progress float64, cameraOrigin *vec3.T, cameraFocusPoint *vec3.T) scn.Camera {

	// Point heading towards center of sphere ring (heading vector starts in camera origin)
	heading := cameraFocusPoint.Subed(cameraOrigin)

	focusDistance := heading.Length()

	return scn.Camera{
		Origin:            cameraOrigin,
		Heading:           &heading,
		ViewUp:            &vec3.T{0, 1, 0},
		ViewPlaneDistance: viewPlaneDistance,
		ApertureSize:      apertureSize,
		FocusDistance:     focusDistance,
		Samples:           amountSamples,
		AntiAlias:         true,
		Magnification:     magnification,
		RenderType:        renderType,
		RecursionDepth:    maxRecursion,
	}
}
