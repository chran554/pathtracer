package main

import (
	"fmt"
	"os"
	dmd "pathtracer/cmd/obj/diamond/pkg/diamond"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/fileformat/wavefront"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"
	"pathtracer/internal/pkg/util"

	"github.com/ungerik/go3d/float64/mat3"
	"github.com/ungerik/go3d/float64/vec3"
)

const (
	animationName = "diamondsR4ever"

	amountAnimationFrames = 360
	frameIndexOffset      = 139 + 44

	imageWidth    = 800
	imageHeight   = 800
	magnification = 1.5

	amountSamples = 1024 * 4

	viewPlaneDistance    = 1000.0
	cameraDistanceFactor = 0.8
	apertureSize         = 1.0

	diamondSize    = 100.0
	recursionDepth = 20

	lampEmission = 4
)

func main() {
	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)

	roomScale := 100.0
	// cornellBox := getCornellBox(cornellBoxFilenamePath, 100.0)

	environmentSphere := scene.NewSphere(&vec3.T{0, 0, 0}, 30*1000, scene.NewMaterial().
		E(color.White, 1, true).
		SP(
			scene.NewTextureReferenceOrPanic("textures/equirectangular/6031885477_bbffaa37d7_o.jpg", scene.TextureTypeAlbedo),
			&vec3.T{0, 0, 0},
			vec3.T{1, 0, 0},
			vec3.T{0, 1, 0},
		)).N("sky dome")
	environmentSphere.RotateY(environmentSphere.Bounds().Center(), util.DegToRad(40))

	lamp1 := scene.NewSphere(&vec3.T{roomScale * 3.0, roomScale * 2.3, 0}, roomScale*0.8, scene.NewMaterial().E(color.White, lampEmission, true)).N("Lamp1")
	lamp2 := scene.NewSphere(&vec3.T{-roomScale * 3.0, roomScale * 2.3, 0}, roomScale*0.8, scene.NewMaterial().E(color.White, lampEmission, true)).N("Lamp2")

	d := dmd.NewDiamond()
	diamondMaterial := dmd.NewMaterial().Attenuate(color.NewColor(0.995, 0.900, 0.880), diamondSize).C(color.NewColor(0.995, 0.900*0.8, 0.880*0.8)).M(0.3, 0.1)
	diamondMaterial.T(0.0, true, scene.RefractionIndex_Diamond)
	diamond := d.RoundBrilliantCutFacetStructure(diamondSize, diamondMaterial)
	diamond.Translate(&vec3.T{0.0, -roomScale * 0.3, 0.0})
	diamondBounds := diamond.Bounds

	// writeDiamondWavefrontFile(diamond)

	floorTexture := scene.NewTextureReferenceOrPanic("textures/marble/white_marble.png", scene.TextureTypeAlbedo)
	//floorTexture := floatimage.LoadOrPanic("textures/floor/7451-diffuse 02 low contrast.png")
	floorMaterial := scene.NewMaterial().
		M(0, 0.05).
		PP(floorTexture, &vec3.T{0, 0, 0}, vec3.T{roomScale * 1.5 * 2, 0, 0}, vec3.T{0, 0, roomScale * 1.5 * 2})
	floor := scene.NewDisc(&vec3.T{0, diamond.Bounds.Ymin, 0}, &vec3.T{0, 1, 0}, roomScale*1.5, floorMaterial).N("Floor")

	for animationFrameIndex := range amountAnimationFrames {
		animationProgress := float64(animationFrameIndex+frameIndexOffset) / float64(amountAnimationFrames)

		cameraOrigin := &vec3.T{0, diamondBounds.Ymax + 50, -200}
		cameraOrigin.Scale(cameraDistanceFactor)
		focusPoint := &vec3.T{diamondBounds.Center()[0], diamondBounds.Ymax, diamondBounds.Center()[2]}

		rotatePoint(util.DegToRad(360*animationProgress), cameraOrigin, focusPoint)

		scn := scene.NewSceneNode().
			S(lamp1, lamp2, environmentSphere).
			D(floor).
			FS(diamond)

		camera := scene.NewCamera(cameraOrigin, focusPoint, amountSamples, magnification).
			V(viewPlaneDistance).
			D(recursionDepth).
			A(apertureSize, nil)

		fi := -1
		if amountAnimationFrames > 1 {
			fi = animationFrameIndex
		}

		frame := scene.NewFrame(animation.AnimationName, fi, camera, scn)
		animation.AddFrame(frame)
	}

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	fmt.Printf("Writing scene file '%s'.\n", filename)
	err := renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}

func rotatePoint(angle float64, point *vec3.T, rotationOrigin *vec3.T) {
	rotationMatrix := mat3.T{}
	rotationMatrix.AssignYRotation(angle)
	p := point.Subed(rotationOrigin)
	p[0], p[2] = p[2], p[0] // Change to right hand coordinate system
	rotationMatrix.TransformVec3(&p)
	p[0], p[2] = p[2], p[0] // Change back to left hand coordinate system
	p.Add(rotationOrigin)

	point[0] = p[0]
	point[1] = p[1]
	point[2] = p[2]
}

func writeDiamondWavefrontFile(diamond *scene.FacetStructure) {
	objFile, err := os.OpenFile("diamond.obj", os.O_WRONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}
	mtlFile, err := os.OpenFile("diamond.mtl", os.O_WRONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}
	wavefront.WriteObjFile(objFile, mtlFile, diamond, nil)
}

func getDiamondConfig() dmd.Diamond {
	return dmd.Diamond{
		GirdleDiameter:                         1.00, // 100.0%
		GirdleHeightRelativeGirdleDiameter:     0.03, //   3.0%
		CrownAngleDegrees:                      34.0, //  34.0°
		TableFacetSizeRelativeGirdle:           0.56, //  56.0%
		StarFacetSizeRelativeCrownSide:         0.55, //  55.0%
		PavilionAngleDegrees:                   41.1, //  41.1°
		LowerHalfFacetSizeRelativeGirdleRadius: 0.77, //  77.0%
	}
}
