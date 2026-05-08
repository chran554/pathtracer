package main

import (
	"fmt"
	"math"
	"math/rand/v2"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/obj"
	"pathtracer/internal/pkg/renderfile"
	"pathtracer/internal/pkg/scene"
	"pathtracer/internal/pkg/util"

	"github.com/ungerik/go3d/float64/quaternion"
	"github.com/ungerik/go3d/float64/vec3"
)

var animationName = "mapletreebynight"

var imageWidth = 800
var imageHeight = 800
var magnification = 1.0 // 0.5

var amountSamples = 1024 * 2 // 1024 // 512

var skyDomeEmission = 0.5

var maxRayDepth = 10

var leafScale = 1.15

var rnd *rand.Rand

func init() {
	rnd = rand.New(rand.NewPCG(1, 2))
}

func main() {
	skyDomeRadius := 150.0

	textureGround, err := floatimage.EmptyPlaceholderImage("textures/ground/dry-grass-ground-2048x2048.png")
	if err != nil {
		panic(err)
	}
	textureEnv, err := floatimage.EmptyPlaceholderImage("textures/equirectangular/331_PDM_BG1.jpg")
	if err != nil {
		panic(err)
	}
	textureLeaf, err := floatimage.EmptyPlaceholderImage("textures/tree/Leaves0120_35_S_02.png")
	if err != nil {
		panic(err)
	}
	textureBark, err := floatimage.EmptyPlaceholderImage("textures/tree/BarkDecidious0143_5_S.jpg")
	if err != nil {
		panic(err)
	}

	leafMaterial := scene.NewMaterial().
		C(color.NewColorRGBA(1.0, 1.0, 1.0, 1.0)).
		T(0.10, false, 1.44). // Refraction index green leaves 1.40-1.48
		M(0.15, 0.85)

	skydomeMaterial := scene.NewMaterial().
		E(color.White, skyDomeEmission, true).
		SP(textureEnv, &vec3.T{0, 0, 0}, vec3.T{1, 0, 0}, vec3.T{0, 1, 0})

	skyDome := scene.NewSphere(&vec3.T{0, 0, 0}, skyDomeRadius, skydomeMaterial).N("sky dome")
	skyDome.RotateY(&vec3.Zero, util.DegToRad(-18))
	skyDome.Translate(&vec3.T{0, 2, 0})

	ground := obj.NewSquareFacetStructure(obj.SquareTypeXZPlane, nil, true)
	ground.ScaleUniform(&vec3.T{0, 0, 0}, skyDomeRadius*2)
	groundMaterial := scene.NewMaterial().
		N("ground").
		PP(textureGround, &vec3.Zero, vec3.T{5, 0, 0}, vec3.T{0, 0, 5})
	groundMaterial.FresnelMaxGlossiness = 0.15

	ground.Material = groundMaterial

	// sun := scene.NewSphere(&vec3.T{-130, 130, -130}, 150, scene.NewMaterial().
	// 	N("sun").
	// 	E(color.KelvinTemperatureColor2(6000), 2.0, true))

	var trees []*scene.FacetStructure
	treeCount := 3
	for i := 0; i < treeCount; i++ {
		tree := createTree(int64(1337+i), textureBark, textureLeaf, leafMaterial, mapleTreeParams())
		if i == 0 {
			tree.Translate(&vec3.T{0, 0, 0})
		} else if i == 1 {
			tree.Translate(&vec3.T{-5, 0, 5})
		} else {
			tree.Translate(&vec3.T{6, 0, 6})
		}
		trees = append(trees, tree)
	}

	lamppostMaterial := scene.NewMaterial().N("lamppost").
		C(color.NewColor(0.10, 0.20, 0.08)).
		M(0.2, 0.3)

	var lampposts []*scene.FacetStructure
	lampPostCount := 5
	for lampPostIndex := 0; lampPostIndex < lampPostCount; lampPostIndex++ {
		lamppost := obj.NewLamppost(4.5, 4)
		lamppost.Material = lamppostMaterial
		lamppost.RotateY(&vec3.T{0, 0, 0}, util.DegToRad(float64(25.0*lampPostIndex)))

		lamppostStartPos := &vec3.T{-7.5, 0, -0.5}
		lamppostOffset := (&vec3.T{5, 0, 3}).Scale(float64(lampPostIndex))
		lamppostPosition := lamppostStartPos.Added(lamppostOffset)
		lamppost.Translate(&lamppostPosition)

		lampposts = append(lampposts, lamppost)
	}

	gopher1 := obj.NewGopher(0.4)
	gopher1.ReplaceMaterial("body", scene.NewMaterial().C(color.NewColor(0.8, 0.7, 0.5)))
	gopher1.RotateY(&vec3.Zero, util.DegToRad(180-33))
	gopher1.Translate(&vec3.T{3.5, 0, 4.0})

	gopher2 := obj.NewGopher(0.4)
	// gopher2.ReplaceMaterial("body", scene.NewMaterial().C(color.NewColor(0.3, 0.2, 0.8)))
	gopher2.RotateY(&vec3.Zero, util.DegToRad(180-7))
	gopher2.RotateX(&vec3.Zero, util.DegToRad(-37))
	gopher2.RotateZ(&vec3.Zero, util.DegToRad(20))
	gopher2.Translate(&vec3.T{0.3, 1.43, -14.6})

	scn := scene.NewSceneNode().S(skyDome).FS(ground).FS(lampposts...).FS(gopher1, gopher2).FS(trees...)

	cameraOrigin := &vec3.T{0, 1.8, -15}
	focusPoint := &vec3.T{0, 5, 0}

	viewVector := focusPoint.Subed(cameraOrigin)
	focusDistance := viewVector.Length()

	camera := scene.NewCamera(cameraOrigin, focusPoint, amountSamples, magnification).
		F(focusDistance).
		D(maxRayDepth)

	animation := scene.NewAnimation(animationName, imageWidth, imageHeight, magnification, true, true)
	frame := scene.NewFrame(animation.AnimationName, -1, camera, scn)
	animation.AddFrame(frame)

	filename := fmt.Sprintf("scene/%s.render.zip", animation.AnimationName)
	err = renderfile.WriteRenderFile(filename, animation)
	if err != nil {
		panic(err)
	}
}

func createTree(rndSeed int64, textureBark *floatimage.FloatImage, textureLeaf *floatimage.FloatImage, leafMaterial *scene.Material, treeParams Params) *scene.FacetStructure {
	branchConfigurations := GenerateTreeLines(rndSeed, vec3.Zero, treeParams)

	var branches []*scene.FacetStructure
	for _, branchConfiguration := range branchConfigurations {
		extraLeafCount := max(0, 3-branchConfiguration.Level) // 1 and 2 extra leaves on the branch on level 2 and 1 respectively
		branch := NewBranch(branchConfiguration, textureBark, textureLeaf, leafMaterial, extraLeafCount)
		branches = append(branches, branch)
	}

	tree := &scene.FacetStructure{FacetStructures: branches}
	tree.UpdateBounds()
	fmt.Printf("Tree bounds: %+v\n", tree.Bounds)
	return tree
}

func createLeaf(textureLeaf *floatimage.FloatImage, leafMaterial *scene.Material, minSize, maxSize float64) *scene.FacetStructure {
	leaf := obj.NewSquareFacetStructure(obj.SquareTypeXYPlane, textureLeaf, true)
	leaf.Material = leafMaterial

	// Scale leaf
	// maxLeafWidth := 0.14 * leafScale // 14 cm
	// minLeafWidth := 0.10 * leafScale // 10 cm
	leaf.Scale(&vec3.T{0, 0, 0}, &vec3.T{random(minSize, maxSize), random(minSize, maxSize), random(minSize, maxSize)})

	// Distort leaf facets by vertex distortion
	maxDistortion := 0.10 * maxSize // 2cm max distortion on lminSize
	for _, leafFacet := range leaf.Facets {
		for _, vertex := range leafFacet.Vertices {
			vertex.Add(&vec3.T{random(-maxDistortion, maxDistortion), random(-maxDistortion, maxDistortion), random(-maxDistortion, maxDistortion)})
		}
		leafFacet.GetBounds()
		leafFacet.UpdateNormal()
	}

	// Rotate leaf
	leaf.RotateY(&vec3.T{0, 0, 0}, random(0, math.Pi*2))
	leaf.RotateX(&vec3.T{0, 0, 0}, random(-math.Pi/2, math.Pi/2))
	leaf.RotateZ(&vec3.T{0, 0, 0}, random(-math.Pi/2, math.Pi/2))
	return leaf
}

func mapleTreeParams() Params {
	return Params{
		Levels: 8,

		TrunkLength: 3.0,  // meters to first branches
		TrunkRadius: 0.20, // 22 cm radius (44 cm diameter)

		ChildrenMin: 3,
		ChildrenMax: 5, // 6

		ConeAngleRadians: util.DegToRad(63), // 1.1. // wide crown
		LengthFalloff:    0.70,              // 0.76
		RadiusFalloff:    0.68,              // 0.68

		LengthJitter: 0.20,
		AngleJitter:  util.DegToRad(43), // 0.75

		UpBias: 0.045, // 0.04 // mostly lateral
	}
}

func birchTreeParams() Params {
	return Params{
		Levels: 8,

		// Slender trunk
		TrunkLength: 1.50,
		TrunkRadius: 0.15,

		// Many small twigs
		ChildrenMin: 2,
		ChildrenMax: 4,

		// Fairly narrow branching cone (birch tends to be “upright/airy”)
		ConeAngleRadians: 0.55,

		// Taper + shortening for finer branches
		LengthFalloff: 0.72,
		RadiusFalloff: 0.62,

		// Natural variation
		LengthJitter: 0.22,
		AngleJitter:  0.30,

		// Keep growth biased upward (cut-leaf birch still grows upward overall)
		UpBias: 0.20,
	}
}

func NewBranch(conf Line, texture *floatimage.FloatImage, textureLeaf *floatimage.FloatImage, leafMaterial *scene.Material, extraLeaves int) *scene.FacetStructure {
	branch := &scene.FacetStructure{Name: fmt.Sprintf("branch %d", conf.Level)}

	branch.Material = scene.NewMaterial().N("branch")

	const sections = 8

	branchVector := conf.End.Subed(&conf.Start)
	branchLength := branchVector.Length()

	axis, angle, ok := RotationFromTo(&vec3.UnitY, &branchVector)
	if !ok {
		panic("Failed to calculate rotation axis")
	}
	rotationQuat := quaternion.FromAxisAngle(axis, angle)

	var lowPoints []*vec3.T
	var highPoints []*vec3.T

	const halfProgress = 0.5 * (1.0 / sections)
	for i := 0; i < sections; i++ {
		progress := float64(i) / sections
		lowPoint := &vec3.T{conf.StartRadius * math.Cos(util.DegToRad(360*progress)), 0, conf.StartRadius * math.Sin(util.DegToRad(360*progress))}
		highPoint := &vec3.T{conf.EndRadius * math.Cos(util.DegToRad(360*(progress-halfProgress))), branchLength, conf.EndRadius * math.Sin(util.DegToRad(360*(progress-halfProgress)))}

		rotationQuat.RotateVec3(lowPoint)
		rotationQuat.RotateVec3(highPoint)

		lowPoints = append(lowPoints, lowPoint)
		highPoints = append(highPoints, highPoint)
	}

	for i := 0; i < sections; i++ {
		f1 := &scene.Facet{Vertices: []*vec3.T{lowPoints[i], highPoints[i], highPoints[(i+1)%sections]}}
		f2 := &scene.Facet{Vertices: []*vec3.T{lowPoints[i], highPoints[(i+1)%sections], lowPoints[(i+1)%sections]}}

		branch.Facets = append(branch.Facets, f1, f2)
	}

	leafMinSize := 0.10
	leafMaxSize := 0.14

	var leaves []*scene.FacetStructure
	leafCount := 1 + extraLeaves
	for leafIndex := 1; leafIndex <= leafCount; leafIndex++ {
		leafPosition := highPoints[int(random(0, sections))].Subed(&branchVector)
		leafPosition.Scale(leafMaxSize / leafPosition.Length()) // leaf position just outside branch
		someWayUpOnBranch := branchVector.Scaled(float64(leafIndex) / float64(leafCount))
		leafPosition.Add(&someWayUpOnBranch)

		leaf := createLeaf(textureLeaf, leafMaterial, leafMinSize*leafScale, leafMaxSize*leafScale)
		leaf.UpdateVertexNormals(false)
		leaf.Translate(&leafPosition)

		leaves = append(leaves, leaf)
	}

	branch.Material.CP(texture, &vec3.T{}, *lowPoints[0], branchVector, true)
	branch.UpdateVertexNormals(false)
	branch.FacetStructures = append(branch.FacetStructures, leaves...) // Add leaves to branch
	branch.Translate(&conf.Start)                                      // Move branch (and leaves) into position in the tree

	return branch
}

func random(min, max float64) float64 {
	return rnd.Float64()*(max-min) + min
}

func UniformOnSphereGaussian() *vec3.T {
	for {
		x := rnd.NormFloat64()
		y := rnd.NormFloat64()
		z := rnd.NormFloat64()
		n2 := x*x + y*y + z*z
		if n2 > 0 {
			inv := 1.0 / math.Sqrt(n2)
			return &vec3.T{x * inv, y * inv, z * inv}
		}
	}
}

func RotationFromTo(a, b *vec3.T) (axis *vec3.T, angle float64, ok bool) {
	if a.Length() == 0 || b.Length() == 0 {
		return &vec3.Zero, 0, false
	}

	u := a.Normalized()
	v := b.Normalized()

	axis_ := vec3.Cross(&u, &v)
	axis = &axis_
	s := axis.Length()
	c := vec3.Dot(&u, &v)

	const eps = 1e-12
	if s < eps {
		if c > 0 {
			// already parallel
			return &vec3.T{0, 1, 0}, 0, true
		}
		// opposite: pick any perpendicular axis
		tmp := &vec3.T{1, 0, 0}
		if math.Abs(u[0]) > 0.9 {
			tmp = &vec3.T{0, 1, 0}
		}
		axis_ = vec3.Cross(&u, tmp)
		axis = &axis_
		axis.Normalize()
		return axis, math.Pi, true
	}

	axis.Scale(1.0 / s)      // normalize axis
	angle = math.Atan2(s, c) // stable angle
	return axis, angle, true
}
