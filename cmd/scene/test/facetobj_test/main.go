package main

import (
	"fmt"
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/fileformat/wavefront"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/mat3"

	"github.com/ungerik/go3d/float64/vec3"
)

const (
	animationName = "facetobj_test"

	amountImages = 180

	imageWidth    = 1024 / 2
	imageHeight   = 800 / 2
	magnification = 1.0 / 2

	amountSamples = 100

	// objectFilename = "lamppost.obj.3ds.obj"
	// objectFilename = "Diamond.obj"
	// objectFilename = "go_gopher_color.obj"
	// objectFilename = "cube_smooth.obj"
	// objectFilename = "cube.obj"
	// objectFilename = "facet.obj"
	// objectFilename = "triangle.obj"
	// objectFilename = "go_gopher_high.obj"
	objectFilename   = "cube_smooth.obj"
	objectPath       = "/Users/christian/projects/code/go/pathtracer/objects/"
	objectScale      = 2.5
	objectStartAngle = 0.0

	environmentRadius         = 100.0 * 20.0 // 20m (if 1 unit is 1 cm)
	environmentEmissionFactor = 1.0

	useLights = true

	lightRadius         = environmentRadius / 14
	lightEmissionFactor = 2.0

	cameraDistance = 400.0

	viewPlaneDistance = 500.0
)

func main() {
	objectFilepath := objectPath + objectFilename

	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, true, false)

	for imageIndex := range amountImages {
		fmt.Printf("\n\nCostructing frame %d\n", imageIndex)
		fmt.Printf("Reading file: %s\n", objectFilepath)

		facetStructure := wavefront.ReadFacetStructureOrPanic(objectFilepath)

		facetStructure.UpdateBounds()
		fmt.Printf("Object in file \"%s\" has bounds %+v.\n", objectFilepath, facetStructure.Bounds)

		facetStructure.ScaleUniform(&vec3.Zero, objectScale)
		facetStructure.RotateY(&vec3.Zero, objectStartAngle)

		scn := scene.NewSceneNode().FS(facetStructure)
		scn.UpdateBounds()
		scn.Bounds = nil

		var lightColor = color.NewColor(1.0, 0.97, 0.95)
		var lightOrigin = vec3.T{environmentRadius / 8, environmentRadius / 4, -environmentRadius / 1.5}

		var cameraOrigin = vec3.T{cameraDistance / 1.0, cameraDistance * 4 / 5, -cameraDistance}

		if useLights {
			lampMaterial := scene.NewMaterial().C(lightColor).E(lightColor, lightEmissionFactor, true)
			lamp := scene.NewSphere(&lightOrigin, lightRadius, lampMaterial).N("Light")
			scn.S(lamp)
		}

		skyDomeOrigin := vec3.T{0, 0, 0}
		skyDomeMaterial := scene.NewMaterial().
			E(color.White, environmentEmissionFactor, true).
			SP(scene.NewTextureReferenceOrPanic("textures/equirectangular/white room 01 1836x918.png", scene.TextureTypeAlbedo), &skyDomeOrigin, vec3.UnitZ, vec3.UnitY)
		skyDome := scene.NewSphere(&skyDomeOrigin, environmentRadius, skyDomeMaterial).N("sky dome")
		scn.S(skyDome)

		animationProgress := float64(imageIndex) / float64(amountImages)
		heightFactor := math.Sin(2.0 * 2.0 * math.Pi * animationProgress)
		camera := getCamera(&cameraOrigin, facetStructure.Bounds.Center(), 2.0*math.Pi*animationProgress, heightFactor)

		frame := scene.NewFrame(animationName, imageIndex, camera, scn)

		animation.AddFrame(frame)
	}

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}

func getCamera(cameraOrigin *vec3.T, focusPoint *vec3.T, yRotationAngle float64, heightFactor float64) *scene.Camera {
	rotationMatrix := mat3.T{}
	rotationMatrix.AssignYRotation(yRotationAngle)

	newCameraOrigin := vec3.T{cameraOrigin[0], cameraOrigin[1] * heightFactor, cameraOrigin[2]}
	newCameraOrigin = rotationMatrix.MulVec3(&newCameraOrigin)

	return scene.NewCamera(&newCameraOrigin, focusPoint, amountSamples, magnification).V(viewPlaneDistance)
}
