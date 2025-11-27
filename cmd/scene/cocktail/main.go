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

var animationName = "cocktail"

var amountFrames = 1

var amountSamples = 1024 * 12 * 4

var imageWidth = 800
var imageHeight = 600
var magnification = 2.0

func main() {
	ceilingLight := createCeilingLight(16)
	ceilingLight.Scale(&vec3.Zero, &vec3.T{300, 1, 10})
	ceilingLight.Translate(&vec3.T{-150, 70, -10})

	textureDarkWood := floatimage.Load("textures/wood/darkwood.png")
	ceilingLightShade := &scene.FacetStructure{Facets: obj.NewSquare(obj.XYPlane, false)}
	ceilingLightShade.Scale(&vec3.Zero, &vec3.T{300, 20, 1})
	ceilingLightShade.Material = scene.NewMaterial().N("ceiling light shade").
		C(color.White).
		PP(textureDarkWood, &vec3.T{0, 0, 0}, vec3.UnitY.Scaled(ceilingLightShade.Bounds.SizeY()), vec3.UnitX.Scaled(ceilingLightShade.Bounds.SizeX()/3))
	ceilingLightShade.Translate(&vec3.T{-150, 64, -10})

	brickWall := createWall(450.0)
	neonSign := createNeonSign(200.0, 8, 2, vec3.T{-80, -50, -5})

	// Posters
	poster0 := createPoster(vec3.T{60, -40, -0.23})
	poster1 := createPoster(vec3.T{-115, -20, -0.22})
	poster2 := createPoster(vec3.T{-115, -80, -0.20})

	poster0.RotateZ(poster1.Bounds.Center(), -math.Pi/14)
	poster1.RotateZ(poster1.Bounds.Center(), math.Pi/12)
	poster2.RotateZ(poster2.Bounds.Center(), -math.Pi/16)

	bench := &scene.FacetStructure{Facets: obj.NewSquare(obj.XZPlane, false)}
	bench.Scale(&vec3.Zero, &vec3.T{300, 1, 60})
	bench.Material = scene.NewMaterial().N("ceiling light shade").
		C(color.White).
		M(0.50, 0.50).
		PP(textureDarkWood, &vec3.T{0, 0, 0}, vec3.UnitX.Scaled(bench.Bounds.SizeX()/3), vec3.UnitZ.Scaled(bench.Bounds.SizeZ()))
	bench.Translate(&vec3.T{-150, -80, -60})

	lightBoxBlinds := obj.NewLightBox(&vec3.T{200, 150, 1600}, color.KelvinTemperatureColor2(2500), 1000.0, "textures/misc/cocktail/blinds_1_2.png")
	lightBoxBlinds.GetFirstObjectByMaterialName("lightpanel").Scale(&vec3.T{0, 0, 0}, &vec3.T{1, 0.75, 1})
	lightBoxBlinds.RotateY(&vec3.Zero, util.DegToRad(180)) // let the light face directly at the wall
	lightBoxBlinds.RotateX(&vec3.Zero, -util.DegToRad(16)) // tilt the light down
	lightBoxBlinds.RotateY(&vec3.Zero, util.DegToRad(18))  // rotate the light box yet to the left around the y-axis
	lightBoxBlinds.Translate(&vec3.T{410, 160, -860})

	lightMaterial2 := scene.NewMaterial().E(color.KelvinTemperatureColor2(3500), 1.0, true)
	light2 := scene.NewSphere(&vec3.T{-80, 0, -180}, 50.0, lightMaterial2).N("light2")

	scn := scene.NewSceneNode().
		S(light2).
		FS(brickWall).
		FS(ceilingLight).
		FS(ceilingLightShade).
		FS(bench).
		FS(poster0).
		FS(poster1).
		FS(poster2).
		FS(neonSign).
		FS(lightBoxBlinds)

	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, true, true)

	for frameIndex := 0; frameIndex < amountFrames; frameIndex++ {
		animationProgress := float64(frameIndex) / float64(amountFrames)

		camera := getCamera(animationProgress)
		frame := scene.NewFrame(animationName, -1, camera, scn)
		animation.AddFrame(frame)
	}

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err := renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}

func createCeilingLight(emission float64) *scene.FacetStructure {
	textureLightStrip := floatimage.Load("textures/misc/cocktail/lightstrip_1_2.png")
	textureLightStripShade := floatimage.Load("textures/misc/cocktail/lightstrip_1_2_shade.png")

	lightPanel := &scene.FacetStructure{Facets: obj.NewSquare(obj.XZPlane, false)}
	lightPanel.Material = scene.NewMaterial().N("light strip").
		E(color.KelvinTemperatureColor2(3000), emission, true).
		PP(textureLightStrip, &vec3.T{0, 0, 0}, vec3.UnitX, vec3.UnitZ)

	shadePanel := &scene.FacetStructure{Facets: obj.NewSquare(obj.XZPlane, false)}
	shadePanel.Material = scene.NewMaterial().N("shade strip").
		PP(textureLightStripShade, &vec3.T{0, 0, 0}, vec3.UnitX, vec3.UnitZ)

	lightPanel.Translate(&vec3.T{0, 4, 0})

	panel := &scene.FacetStructure{FacetStructures: []*scene.FacetStructure{lightPanel, shadePanel}}

	return panel
}

func createNeonSign(neonSignWidth float64, coreEmission, haloEmission float64, lowerLeftCorner vec3.T) *scene.FacetStructure {
	textureCocktailsNeonSignCore := floatimage.Load("textures/misc/cocktail/cocktails_mod03_core.png")
	textureCocktailsNeonSignHalo := floatimage.Load("textures/misc/cocktail/cocktails_mod03_halo.png")

	core1 := &scene.FacetStructure{SubstructureName: "core1", Facets: obj.NewSquare(obj.XYPlane, false)}
	core2 := &scene.FacetStructure{SubstructureName: "core2", Facets: obj.NewSquare(obj.XYPlane, false)}
	halo := &scene.FacetStructure{SubstructureName: "halo", Facets: obj.NewSquare(obj.XYPlane, false)}

	core1.UpdateBounds()
	core1.UpdateNormals()
	core1.ScaleUniform(&vec3.Zero, neonSignWidth)
	core1.Translate(&lowerLeftCorner)
	core1.Translate(&vec3.T{0, 0, -0.3})
	core1.Material = scene.NewMaterial().N("core").
		E(color.White, coreEmission, false).
		T(1.0, false, scene.RefractionIndex_Air).
		PP(textureCocktailsNeonSignCore, &lowerLeftCorner, vec3.UnitX.Scaled((neonSignWidth/2)*1.6), vec3.UnitY.Scaled(neonSignWidth/2))
	core1.Material.Projection.RepeatU = false
	core1.Material.Projection.RepeatV = false
	core1.Material.SolidObject = false

	core2.UpdateBounds()
	core2.UpdateNormals()
	core2.ScaleUniform(&vec3.Zero, neonSignWidth)
	core2.Translate(&lowerLeftCorner)
	core2.Translate(&vec3.T{0, 0, +0.3})
	core2.Material = scene.NewMaterial().N("core").
		E(color.White, coreEmission*1.5, false).
		T(1.0, false, scene.RefractionIndex_Air).
		PP(textureCocktailsNeonSignCore, &lowerLeftCorner, vec3.UnitX.Scaled((neonSignWidth/2)*1.6), vec3.UnitY.Scaled(neonSignWidth/2))
	core2.Material.Projection.RepeatU = false
	core2.Material.Projection.RepeatV = false
	core2.Material.SolidObject = false

	halo.UpdateBounds()
	halo.UpdateNormals()
	halo.ScaleUniform(&vec3.Zero, neonSignWidth)
	halo.Translate(&lowerLeftCorner)
	halo.Material = scene.NewMaterial().N("halo").
		E(color.White, haloEmission, false).
		T(1.0, false, scene.RefractionIndex_Air).
		PP(textureCocktailsNeonSignHalo, &lowerLeftCorner, vec3.UnitX.Scaled((neonSignWidth/2)*1.6), vec3.UnitY.Scaled(neonSignWidth/2))
	halo.Material.Projection.RepeatU = false
	halo.Material.Projection.RepeatV = false
	halo.Material.SolidObject = false

	sign := &scene.FacetStructure{Name: "sign", FacetStructures: []*scene.FacetStructure{core1, halo}}

	return sign
}

func createWall(wallScale float64) *scene.FacetStructure {
	textureWallBricks := floatimage.Load("textures/misc/cocktail/bricks.png")

	wall := &scene.FacetStructure{Name: "wall", Facets: obj.NewSquare(obj.XYPlane, false)}
	wall.UpdateBounds()
	wall.UpdateNormals()
	wall.CenterOn(&vec3.Zero)
	wall.ScaleUniform(&vec3.Zero, wallScale)
	wall.Material = scene.NewMaterial().N("wall").
		M(0.05, 0.8).
		PP(textureWallBricks, &vec3.Zero, vec3.UnitX.Scaled(75), vec3.UnitY.Scaled(75))
	return wall
}

func createPoster(posterLocation vec3.T) *scene.FacetStructure {
	textureCocktailPoster := floatimage.Load("textures/misc/cocktail/cocktailposter_worn.png")

	poster := &scene.FacetStructure{Name: "poster", Facets: obj.NewSquare(obj.XYPlane, false)}
	poster.UpdateBounds()
	poster.UpdateNormals()
	poster.Material = scene.NewMaterial().N("poster").
		M(0.10, 0.2).
		PP(textureCocktailPoster, &vec3.T{0, 0, 0}, vec3.UnitX, vec3.UnitY)
	poster.Material.Projection.RepeatU = false
	poster.Material.Projection.RepeatV = false
	poster.Material.SolidObject = false

	poster.Scale(&vec3.Zero, &vec3.T{50, 70, 1})
	poster.Translate(&posterLocation)

	return poster
}

func getCamera(animationProgress float64) *scene.Camera {
	cameraOrigin := &vec3.T{-45, 45, -100}
	// cameraOrigin := &vec3.T{-45, 45, -400} // For overview of the scene during scene test/development
	cameraOrigin.Scale(1.6)
	//cameraOrigin.Scale(2.6)
	focusPoint := &vec3.T{-10, 0, 0}
	// focusPoint := &vec3.T{30, 0, 0}

	// AnimationInformation
	angle := (math.Pi / 2.0) * animationProgress
	scene.RotateY(cameraOrigin, &vec3.Zero, angle)
	scene.RotateY(focusPoint, &vec3.Zero, angle)

	heading := focusPoint.Subed(cameraOrigin)
	focusDistance := heading.Length() - 150.0

	return scene.NewCamera(cameraOrigin, focusPoint, amountSamples, magnification).
		F(focusDistance).
		V(800).
		D(10).
		A(0.1, nil)
}
