package main

import (
	"fmt"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"
	"pathtracer/internal/pkg/util"

	"github.com/ungerik/go3d/float64/vec3"
)

const (
	animationName   = "refraction_lens_test"
	amountFrames    = 180
	startFrameIndex = 0

	maxRecursionDepth = 5
	amountSamples     = 1024 * 4 // * 4

	imageWidth    = 400
	imageHeight   = 500
	magnification = 1.0
)

var texture = floatimage.LoadOrPanic("textures/test/christian.png")

const (
	textureEmisssion = 6.0
	textureBaseSize  = 300.0
	textureDistance  = 400.0

	lensRadius      = 50.0
	lensMinDistance = 0.0 // 0.0
	lensMaxDistance = 80.0
)

func main() {
	lensDistance := lensMinDistance

	textureSkyDome := scene.NewTextureReferenceOrPanic("textures/equirectangular/331_PDM_BG1.jpg", scene.TextureTypeAlbedo)
	skyDome := scene.NewSphere(&vec3.T{0, 0, 0}, 2000, scene.NewMaterial().
		E(color.White, 0.8, true).
		SP(textureSkyDome, &vec3.T{0, 0, 0}, vec3.T{1, 0, 0}, vec3.T{0, 1, 0})).N("sky dome")
	skyDome.RotateY(&vec3.Zero, util.DegToRad(-20))

	backplane := &scene.FacetStructure{
		Name:     "backplane",
		Material: scene.NewMaterial().N("backplane").C(color.NewColor(0.95, 0.95, 0.95)),
		Facets:   obj.NewSquare(obj.SquareTypeXYPlane, nil),
	}
	backplane.Scale(&vec3.Zero, &vec3.T{125, 125, 1})
	backplane.Translate(&vec3.T{-backplane.Bounds.SizeX() / 2, -backplane.Bounds.SizeY() / 2, 0})

	// frame := wavefrontobj.ReadOrPanic(filepath.Join(obj.ObjEvaluationFileDir, "Mirror Devon & Devon N100812.obj"))
	// frame.ScaleUniform(&vec3.Zero, 1/frame.Bounds.SizeY())
	// frame.Scale(&vec3.Zero, &vec3.T{125, 125, 1})

	lightPlane := &scene.FacetStructure{
		Name: "light plane",
		Material: scene.NewMaterial().N("light plane").C(color.NewColor(0.95, 0.95, 0.95)).
			E(color.White, textureEmisssion, true),
		Facets: obj.NewSquare(obj.SquareTypeXYPlane, texture),
	}
	lightPlane.Scale(&vec3.Zero, &vec3.T{textureBaseSize, textureBaseSize * 1.5, 1})
	lightPlane.Translate(&vec3.T{-lightPlane.Bounds.SizeX() / 2, -lightPlane.Bounds.SizeY() / 2, -textureDistance})
	lightPlane.RotateZ(&vec3.Zero, util.DegToRad(180))

	cameraOrigin := &vec3.T{200, 0, -150}
	cameraFocusPoint := &vec3.Zero
	camera := scene.NewCamera(cameraOrigin, cameraFocusPoint, amountSamples, magnification).D(maxRecursionDepth).A(0, nil)

	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)

	for frameIndex := startFrameIndex; frameIndex < amountFrames; frameIndex++ {
		progress := float64(frameIndex) / float64(amountFrames)

		lensDistance = lensMinDistance + progress*(lensMaxDistance-lensMinDistance)
		lens := scene.NewSphere(&vec3.T{0, 0, -(lensDistance + lensRadius)}, lensRadius, scene.NewMaterialGlass("glass lens"))

		scn := scene.NewSceneNode().FS(lightPlane).FS(backplane).S(lens).S(skyDome)

		fi := -1
		if amountFrames > 1 {
			fi = frameIndex
		}

		frame := scene.NewFrame(animation.AnimationName, fi, camera, scn)
		animation.AddFrame(frame)
	}

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}
