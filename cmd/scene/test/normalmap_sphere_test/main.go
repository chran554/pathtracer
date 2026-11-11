package main

import (
	"fmt"
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"
	"pathtracer/internal/pkg/util"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "normalmap_sphere_test"

var amountAnimationFrames = 120 // 72 * 2

var imageWidth = 640
var imageHeight = 480
var magnification = 0.5

var amountSamples = 512 * 2 // * 8 //  8 * 12 // * 2 * 15 // * 8 // * 15
var maxRecursionDepth = 4

var cameraAperture = 0.0
var cameraZoom = 1.0

var lampIntensity = 13.0
var lightIntensityFactor = 0.2
var sphereRadius = 1.5

func main() {
	//textureGround, err := floatimage.EmptyPlaceholderImage("textures/floor/7451-diffuse 02 low contrast.png")
	//textureGround, err := floatimage.EmptyPlaceholderImage("textures/floor/7451-diffuse 02.png")
	textureGround, err := floatimage.EmptyPlaceholderImage("textures/floortiles/dirty-tiles_02/Dirty_tile_Diffuse.jpg")
	if err != nil {
		panic(err)
	}
	textureGroundNM, err := floatimage.EmptyPlaceholderImage("textures/floortiles/dirty-tiles_02/Dirty_tile_Normal.jpg")
	if err != nil {
		panic(err)
	}

	//textureWall, err := floatimage.EmptyPlaceholderImage("textures/tapeter/Camille_Image_Flatshot_Item_7209.jpg")
	//textureWall, err := floatimage.EmptyPlaceholderImage("textures/wallpaper/Borosanpapper2021_Image_Flatshot_Item_8620.jpg")

	textureWall, err := floatimage.EmptyPlaceholderImage("textures/wallpaper/sapphirebirds_image_tile_item_2279.jpg")
	if err != nil {
		panic(err)
	}

	// Cornell box
	cornellBox := obj.NewWhiteCornellBox(&vec3.T{16, 6, 16}, true, lightIntensityFactor) // cm, as units. I.e. a 5x3x5m room
	cornellBox.RemoveObjectsByName("Lamp")

	normalGrayColor := color.NewColorGrey(1.0)

	cornellBox.GetFirstMaterialByName("Left").C(normalGrayColor)
	cornellBox.GetFirstMaterialByName("Right").C(normalGrayColor)
	cornellBox.GetFirstMaterialByName("Back").C(normalGrayColor)

	floor := cornellBox.GetFirstObjectBySubstructureName("Floor")
	if floor == nil {
		panic("floor is nil")
	}
	floor.RotateZ(floor.Bounds.Center(), math.Pi)

	floorMaterial := cornellBox.GetFirstMaterialByName("Floor")
	if floorMaterial == nil {
		panic("floor material is nil")
	}
	floorMaterial.C(color.NewColorGrey(1.0)).M(0.5, 0.5).
		PP(textureGround, scene.TextureTypeAlbedo, &vec3.Zero, vec3.T{30, 0, 0}, vec3.T{0, 0, 30}).
		PP(textureGroundNM, scene.TextureTypeNormalMap, &vec3.Zero, vec3.T{30, 0, 0}, vec3.T{0, 0, 30})
	floorMaterial.Projections[len(floorMaterial.Projections)-1].Texture.Strength = 0.2
	floorMaterial.FresnelMaxGlossiness = 0.15

	wallpaperZoomFactor := 1.0
	wallpaperSize := cornellBox.Bounds.SizeY() / wallpaperZoomFactor

	leftWallMaterial := cornellBox.GetFirstMaterialByName("Left")
	leftWallMaterial.PP(textureWall, scene.TextureTypeAlbedo, &vec3.T{0, cornellBox.Bounds.Ymin, cornellBox.Bounds.Zmin}, vec3.T{0, 0, wallpaperSize}, vec3.T{0, wallpaperSize, 0})
	leftWallMaterial.FresnelMaxGlossiness = 0.15

	rightWallMaterial := cornellBox.GetFirstMaterialByName("Right")
	rightWallMaterial.PP(textureWall, scene.TextureTypeAlbedo, &vec3.T{0, cornellBox.Bounds.Ymin, cornellBox.Bounds.Zmin}, vec3.T{0, 0, wallpaperSize}, vec3.T{0, wallpaperSize, 0})
	rightWallMaterial.FresnelMaxGlossiness = 0.15

	backWallMaterial := cornellBox.GetFirstMaterialByName("Back")
	backWallMaterial.PP(textureWall, scene.TextureTypeAlbedo, &vec3.T{cornellBox.Bounds.Xmin, cornellBox.Bounds.Ymin, 0}, vec3.T{wallpaperSize, 0, 0}, vec3.T{0, wallpaperSize, 0})
	backWallMaterial.FresnelMaxGlossiness = 0.15

	ceilingMaterial := cornellBox.GetFirstMaterialByName("Ceiling").C(color.NewColor(1.0, 1.0, 1.0).Multiply(0.30))
	ceilingMaterial.FresnelMaxGlossiness = 0.15

	behindWall := obj.NewSquareFacetStructure(obj.SquareTypeXYPlane, nil, true)
	behindWall.Name = "Behind"
	behindWall.Scale(&vec3.Zero, &vec3.T{cornellBox.Bounds.SizeX(), cornellBox.Bounds.SizeY(), 1})
	behindWall.Translate(&vec3.T{0, -behindWall.Bounds.Ymin, cornellBox.Bounds.Zmin})
	behindWall.Material = backWallMaterial

	cornellBox.FacetStructures = append(cornellBox.FacetStructures, behindWall)

	//textureSphere, err := floatimage.EmptyPlaceholderImage("textures/equirectangular/earth/light.jpg")
	//textureSphere, err := floatimage.EmptyPlaceholderImage("textures/planets/earth/earth_grey_50.png")
	//textureSphere, err := floatimage.EmptyPlaceholderImage("textures/planets/earth/earth_daymap_dark.jpg")
	//textureSphere, err := floatimage.EmptyPlaceholderImage("textures/football/football-diffuse-512_02.png")
	textureSphere, err := floatimage.EmptyPlaceholderImage("textures/concrete/concrete_flat_lightgray.png")
	if err != nil {
		panic(err)
	}

	//textureNormals, err := floatimage.EmptyPlaceholderImage("textures/football/football-normals-512.png")
	textureNormals, err := floatimage.EmptyPlaceholderImage("textures/normal maps/golfball.jpg")
	if err != nil {
		panic(err)
	}

	//goldColor := color.NewColor(1.00, 0.85, 0.58).Multiply(0.9)
	//silverColor := color.NewColor(0.85, 0.85, 0.9).Multiply(1.0)
	sphereMaterial := scene.NewMaterial().N("sphere").
		M(0.5, 0.5).
		C(color.White)
	sphereMaterial.FresnelMaxGlossiness = 0.10

	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, false, false)

	for frameIndex := range amountAnimationFrames {
		animationProgress := float64(frameIndex) / float64(amountAnimationFrames)

		sphereMat := sphereMaterial.Copy()
		sphereMat.SP(textureSphere, scene.TextureTypeAlbedo, &vec3.T{0, 0, 0}, vec3.T{1, 0, 0}, vec3.T{0, 1, 0})
		sphereMat.SP(textureNormals, scene.TextureTypeNormalMap, &vec3.T{0, 0, 0}, vec3.T{1, 0, 0}, vec3.T{0, 1, 0})
		sphereMat.Projections[len(sphereMat.Projections)-1].Texture.Strength = 1.0
		sphere := scene.NewSphere(&vec3.T{0, 0, 0}, sphereRadius, sphereMat)

		sphere.RotateX(sphere.Origin, util.DegToRad(23.4))
		sphere.Translate(&vec3.T{0, sphereRadius, 0})

		sphere.RotateY(sphere.Origin, util.DegToRad(animationProgress*360))
		sphere.RotateX(sphere.Origin, util.DegToRad(animationProgress*360*2))
		sphere.RotateZ(sphere.Origin, util.DegToRad(-animationProgress*360))

		sphereBounds := sphere.Bounds()
		sphereCenter := sphereBounds.Center()

		lampPosition := vec3.Zero.Added(&vec3.T{-5, sphereBounds.Ymax + sphereRadius/2, -5})
		lamp := scene.NewSphere(&lampPosition, 1.5, scene.NewMaterial().
			N("lamp").
			E(color.KelvinTemperatureColor2(5500), lampIntensity, true))
		lamp.RotateY(sphereCenter, util.DegToRad(animationProgress*360))

		// cameraPosition := &vec3.T{0, 3, -6}
		cameraPosition := &vec3.T{0, sphereBounds.Ymax * 1.5, -4}

		cameraAim := sphereCenter
		focusPosition := cameraAim.Added(&vec3.T{0, 0, -sphereRadius / 2})
		cameraFocusVector := focusPosition.Subed(cameraPosition)

		camera := scene.NewCamera(cameraPosition, cameraAim, amountSamples, magnification).
			D(maxRecursionDepth).
			A(cameraAperture, nil).
			F(cameraFocusVector.Length()).
			V(550 * cameraZoom)

		scn := scene.NewSceneNode().FS(cornellBox).S(lamp).S(sphere)

		fi := -1
		if amountAnimationFrames > 1 {
			fi = frameIndex
		}

		frame := scene.NewFrame(animationName, fi, camera, scn)
		animation.AddFrame(frame)
	}

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err = renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}
