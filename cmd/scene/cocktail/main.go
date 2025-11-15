package main

import (
	"fmt"
	"math"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj"
	anm "pathtracer/internal/pkg/renderfile"
	scn "pathtracer/internal/pkg/scene"

	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "cocktail"

var amountFrames = 1

var amountSamples = 1024 * 2 * 12

var imageWidth = 800
var imageHeight = 600
var magnification = 1.0

func main() {
	ceilingLight := createCeilingLight(16)
	ceilingLight.Scale(&vec3.Zero, &vec3.T{300, 1, 10})
	ceilingLight.Translate(&vec3.T{-150, 70, -10})

	brickWall := createWall(450.0)
	neonSign := createNeonSign(200.0, 15, 15, vec3.T{-80, -50, -6})

	// Posters
	poster0 := createPoster(vec3.T{60, -40, -0.23})
	poster1 := createPoster(vec3.T{-115, -20, -0.22})
	poster2 := createPoster(vec3.T{-115, -80, -0.20})

	poster0.RotateZ(poster1.Bounds.Center(), -math.Pi/14)
	poster1.RotateZ(poster1.Bounds.Center(), math.Pi/12)
	poster2.RotateZ(poster2.Bounds.Center(), -math.Pi/16)

	lightBoxBlinds := obj.NewLightBox(&vec3.T{200, 150, 1600}, color.KelvinTemperatureColor2(2500), 1000.0, "textures/misc/cocktail/blinds_1_2.png")
	lightBoxBlinds.GetFirstObjectByMaterialName("lightpanel").Scale(&vec3.T{0, 0, 0}, &vec3.T{1, 0.75, 1})
	lightBoxBlinds.RotateY(&vec3.Zero, math.Pi)
	lightBoxBlinds.RotateX(&vec3.Zero, -math.Pi*12/180)
	lightBoxBlinds.RotateY(&vec3.Zero, math.Pi*20/180)
	lightBoxBlinds.Translate(&vec3.T{130 + 250 + 30, 0 + 95, -700 - 160})

	lightMaterial2 := scn.NewMaterial().E(color.KelvinTemperatureColor2(3500), 2, true)
	light2 := scn.NewSphere(&vec3.T{100, -150, -125}, 50.0, lightMaterial2).N("light2")

	scene := scn.NewSceneNode().
		S(light2).
		FS(brickWall).
		FS(ceilingLight).
		FS(poster0).
		FS(poster1).
		FS(poster2).
		FS(neonSign).
		FS(lightBoxBlinds)

	animation := scn.NewAnimation(animationName, imageWidth, imageHeight, magnification, true, true)

	for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountFrames)

		camera := getCamera(animationProgress)
		frame := scn.NewFrame(animationName, -1, camera, scene)
		animation.AddFrame(frame)
	}

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := anm.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}

func createCeilingLight(emission float64) *scn.FacetStructure {
	lightPanel := &scn.FacetStructure{Facets: obj.NewSquare(obj.XZPlane, false)}
	lightPanel.Material = scn.NewMaterial().N("light strip").
		E(color.KelvinTemperatureColor2(3000), emission, true).
		PP(floatimage.Load("textures/misc/cocktail/lightstrip_1_2.png"), &vec3.T{0, 0, 0}, vec3.UnitX, vec3.UnitZ)

	shadePanel := &scn.FacetStructure{Facets: obj.NewSquare(obj.XZPlane, false)}
	shadePanel.Material = scn.NewMaterial().N("shade strip").
		PP(floatimage.Load("textures/misc/cocktail/lightstrip_1_2_shade.png"), &vec3.T{0, 0, 0}, vec3.UnitX, vec3.UnitZ)

	lightPanel.Translate(&vec3.T{0, 4, 0})

	panel := &scn.FacetStructure{FacetStructures: []*scn.FacetStructure{lightPanel, shadePanel}}

	return panel
}

func createNeonSign(neonSignWidth float64, coreEmission, haloEmission float64, lowerLeftCorner vec3.T) *scn.FacetStructure {
	core1 := &scn.FacetStructure{SubstructureName: "core1", Facets: obj.NewSquare(obj.XYPlane, false)}
	core2 := &scn.FacetStructure{SubstructureName: "core2", Facets: obj.NewSquare(obj.XYPlane, false)}
	halo := &scn.FacetStructure{SubstructureName: "halo", Facets: obj.NewSquare(obj.XYPlane, false)}

	core1.UpdateBounds()
	core1.UpdateNormals()
	core1.ScaleUniform(&vec3.Zero, neonSignWidth)
	core1.Translate(&lowerLeftCorner)
	core1.Translate(&vec3.T{0, 0, -0.3})
	core1.Material = scn.NewMaterial().N("core").
		E(color.White, coreEmission, false).
		PP(floatimage.Load("textures/misc/cocktail/cocktails_mod03_core.png"), &lowerLeftCorner, vec3.UnitX.Scaled((neonSignWidth/2)*1.6), vec3.UnitY.Scaled(neonSignWidth/2))
	core1.Material.Projection.RepeatU = false
	core1.Material.Projection.RepeatV = false
	core1.Material.SolidObject = false

	core2.UpdateBounds()
	core2.UpdateNormals()
	core2.ScaleUniform(&vec3.Zero, neonSignWidth)
	core2.Translate(&lowerLeftCorner)
	core2.Translate(&vec3.T{0, 0, +0.3})
	core2.Material = scn.NewMaterial().N("core").
		E(color.White, coreEmission*1.5, false).
		PP(floatimage.Load("textures/misc/cocktail/cocktails_mod03_core.png"), &lowerLeftCorner, vec3.UnitX.Scaled((neonSignWidth/2)*1.6), vec3.UnitY.Scaled(neonSignWidth/2))
	core2.Material.Projection.RepeatU = false
	core2.Material.Projection.RepeatV = false
	core2.Material.SolidObject = false

	halo.UpdateBounds()
	halo.UpdateNormals()
	halo.ScaleUniform(&vec3.Zero, neonSignWidth)
	halo.Translate(&lowerLeftCorner)
	halo.Material = scn.NewMaterial().N("halo").
		E(color.White, haloEmission, false).
		PP(floatimage.Load("textures/misc/cocktail/cocktails_mod03_halo.png"), &lowerLeftCorner, vec3.UnitX.Scaled((neonSignWidth/2)*1.6), vec3.UnitY.Scaled(neonSignWidth/2))
	halo.Material.Projection.RepeatU = false
	halo.Material.Projection.RepeatV = false
	halo.Material.SolidObject = false

	sign := &scn.FacetStructure{Name: "sign", FacetStructures: []*scn.FacetStructure{core1, halo}}

	return sign
}

func createWall(wallScale float64) *scn.FacetStructure {
	wall := &scn.FacetStructure{Name: "wall", Facets: obj.NewSquare(obj.XYPlane, false)}
	wall.UpdateBounds()
	wall.UpdateNormals()
	wall.CenterOn(&vec3.Zero)
	wall.ScaleUniform(&vec3.Zero, wallScale)
	wall.Material = scn.NewMaterial().N("wall").
		M(0.05, 0.8).
		//PP("textures/tapeter 2/WhiteBrickWall_Image_Tile_Item_9459w.jpg", &vec3.Zero, vec3.UnitX.Scaled(75), vec3.UnitY.Scaled(75))
		PP(floatimage.Load("textures/misc/cocktail/bricks.png"), &vec3.Zero, vec3.UnitX.Scaled(75), vec3.UnitY.Scaled(75))
	return wall
}

func createPoster(posterLocation vec3.T) *scn.FacetStructure {
	poster := &scn.FacetStructure{Name: "poster", Facets: obj.NewSquare(obj.XYPlane, false)}
	poster.UpdateBounds()
	poster.UpdateNormals()
	poster.Material = scn.NewMaterial().N("poster").
		M(0.05, 0.4).
		PP(floatimage.Load("textures/misc/cocktail/cocktailposter_worn.png"), &vec3.T{0, 0, 0}, vec3.UnitX, vec3.UnitY)
	poster.Material.Projection.RepeatU = false
	poster.Material.Projection.RepeatV = false
	poster.Material.SolidObject = false

	poster.Scale(&vec3.Zero, &vec3.T{50, 70, 1})
	poster.Translate(&posterLocation)

	return poster
}

func getCamera(animationProgress float64) *scn.Camera {
	cameraOrigin := &vec3.T{-45, 45, -100}
	cameraOrigin.Scale(1.6)
	//cameraOrigin.Scale(2.6)
	focusPoint := &vec3.T{-10, 0, 0}
	// focusPoint := &vec3.T{30, 0, 0}

	// AnimationInformation
	angle := (math.Pi / 2.0) * animationProgress
	scn.RotateY(cameraOrigin, &vec3.Zero, angle)
	scn.RotateY(focusPoint, &vec3.Zero, angle)

	heading := focusPoint.Subed(cameraOrigin)
	focusDistance := heading.Length() - 150.0

	return scn.NewCamera(cameraOrigin, focusPoint, amountSamples, magnification).
		F(focusDistance).
		V(800).
		D(10).
		A(0.05, nil)
}
