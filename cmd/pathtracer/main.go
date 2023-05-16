package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ungerik/go3d/float64/mat3"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/image"
	"pathtracer/internal/pkg/rendermonitor"
	"pathtracer/internal/pkg/renderpass"
	scn "pathtracer/internal/pkg/scene"
	"pathtracer/internal/pkg/sunflower"
	"pathtracer/internal/pkg/util"
	"strings"
	"sync"
	"time"

	progressbar2 "github.com/schollz/progressbar/v3"
	"github.com/ungerik/go3d/float64/vec3"
)

const (
	epsilonDistance = 0.0001
	//epsilonDistance = 0.000001
)

var (
	debugPixel = struct{ x, y int }{x: -1, y: -1} // No debug
	//debugPixel = struct{ x, y int }{x: 450, y: 300}
)

type IntersectionInformation struct {
	intersection         bool
	intersectionPoint    *vec3.T
	shortestDistance     float64
	material             *scn.Material
	normalAtIntersection *vec3.T
	intersectedFacet     *scn.Facet
	intersectedSphere    *scn.Sphere
	intersectedDisc      *scn.Disc
}

type RenderFrameInformation struct {
	frameIndex          int
	animationFrameCount int
	imageFilename       string
	renderAlgorithm     scn.RenderType
	imageWidth          int
	imageHeight         int

	amountFacets  int
	amountSpheres int
	amountDiscs   int

	samplesPerPixel   int
	maxRecursionDepth int

	renderStartTime time.Time
	renderEndTime   time.Time
}

func NewRenderFrameInformation(scene *scn.SceneNode, animation *scn.Animation, frame *scn.Frame) RenderFrameInformation {
	return RenderFrameInformation{
		// frameIndex:          frameIndex,
		animationFrameCount: len(animation.Frames),
		imageFilename:       frame.Filename,
		renderAlgorithm:     frame.Camera.RenderType,
		imageWidth:          animation.Width,
		imageHeight:         animation.Height,
		amountFacets:        scene.GetAmountFacets(),
		amountSpheres:       scene.GetAmountSpheres(),
		amountDiscs:         scene.GetAmountDiscs(),
		samplesPerPixel:     frame.Camera.Samples,
		maxRecursionDepth:   frame.Camera.RecursionDepth,
		// renderStartTime:     time.Now(),
		// renderEndTime:       time.Time{},
		// renderDuration:      0,
	}
}

func NewIntersectionInformation() *IntersectionInformation {
	return &IntersectionInformation{
		intersection:         false,           // Intersection occurred? True/false
		intersectionPoint:    nil,             // Point of intersection
		shortestDistance:     math.MaxFloat64, // At what distance from start point of fired ray
		material:             nil,             // The material of the closest object that was intersected
		normalAtIntersection: nil,             // The normal of the object that was intersected, at intersection point
		intersectedFacet:     nil,
		intersectedSphere:    nil,
		intersectedDisc:      nil,
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: pathtracer <animation filename>")
		os.Exit(1)
	}

	animationFilename := os.Args[1]

	if _, err := os.Stat(animationFilename); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("File '%s' do not exist.", animationFilename)
		fmt.Println("Usage: pathtracer <animation filename>")
		os.Exit(1)
	}

	startTimestamp := time.Now()

	var animationJSON, err = os.ReadFile(animationFilename)
	if err != nil {
		panic(err)
	}

	animation := scn.Animation{}
	err = json.Unmarshal(animationJSON, &animation)
	if err != nil {
		panic(err)
	}

	fmt.Println("-----------------------------------------------")
	fmt.Println("Animation file: ", animationFilename)
	fmt.Println("Animation name: ", animation.AnimationName)
	fmt.Println("Amount frames:  ", len(animation.Frames))
	fmt.Println()

	renderMonitor := rendermonitor.NewRenderMonitor()
	defer renderMonitor.Close()

	for frameIndex, frame := range animation.Frames {
		frameInformation := NewRenderFrameInformation(frame.SceneNode, &animation, frame)
		frameInformation.frameIndex = frameIndex
		frameInformation.renderStartTime = time.Now()

		renderMonitor.Initialize(animation.AnimationName, frame.Filename, animation.Width, animation.Height)
		time.Sleep(50 * time.Millisecond)

		fmt.Println(frameInformationPreRenderText(frameInformation))

		fmt.Println()
		fmt.Println("Initialize scene...")
		scene := frame.SceneNode
		initializeScene(scene)

		renderedPixelData := image.NewFloatImage(animation.AnimationName, animation.Width, animation.Height)

		fmt.Println(frameInformationProgressSummary(frameInformation))
		render(frame.Camera, scene, animation.Width, animation.Height, renderedPixelData, &renderMonitor)

		fmt.Println("Releasing resources...")
		deInitializeScene(scene)
		frame.SceneNode = nil

		frameInformation.renderEndTime = time.Now()
		fmt.Println()
		fmt.Println(frameInformationPostRenderText(frameInformation))

		writeRenderedImage(animation, frame, renderedPixelData, frameInformation)
	}

	fmt.Printf("Total execution time (for %d frames): %s\n", len(animation.Frames), time.Since(startTimestamp))
}

func frameInformationProgressSummary(frameInformation RenderFrameInformation) string {
	animationProgressText := fmt.Sprintf("animation progress, before rendering frame, %.2f%%", (float64(frameInformation.frameIndex)/float64(frameInformation.animationFrameCount))*100.0)
	animationEstimatedRemainingTimeText := ""
	if frameInformation.frameIndex > 0 {
		renderTime := time.Since(frameInformation.renderStartTime)
		timePerFrame := float64(renderTime) / float64(frameInformation.frameIndex)
		animationEstimatedRemainingTime := timePerFrame * float64(frameInformation.animationFrameCount-(frameInformation.frameIndex))
		animationEstimatedRemainingTimeText = fmt.Sprintf(" with %d to go", time.Duration(int64(animationEstimatedRemainingTime)).Round(time.Minute))
	}

	if frameInformation.animationFrameCount > 1 {
		return fmt.Sprintf("Rendering %s, frame %d of %d  (%s%s)...\n", frameInformation.imageFilename, frameInformation.frameIndex+1, frameInformation.animationFrameCount, animationProgressText, animationEstimatedRemainingTimeText)
	} else {
		return fmt.Sprintf("Rendering %s (single frame)...\n", frameInformation.imageFilename)
	}
}

func frameInformationPostRenderText(frameInformation RenderFrameInformation) string {
	stringBuilder := strings.Builder{}
	renderDuration := frameInformation.renderEndTime.Sub(frameInformation.renderStartTime)
	renderDuration.Round(time.Minute)
	stringBuilder.WriteString(fmt.Sprintf("Render date:           %s\n", frameInformation.renderStartTime.Format("2006-01-02")))
	stringBuilder.WriteString(fmt.Sprintf("Render duration:       %s\n", renderDuration))

	return stringBuilder.String()
}

func frameInformationPreRenderText(frameInformation RenderFrameInformation) string {
	stringBuilder := strings.Builder{}

	unevenWidth := (frameInformation.imageWidth%2) == 1 || (frameInformation.imageHeight%2) == 1
	mp4CreationWarning := ""
	if unevenWidth && (frameInformation.animationFrameCount > 1) {
		mp4CreationWarning = " (uneven width or height, no animation can be made)"
	}

	progress := float64(frameInformation.frameIndex+1) / float64(frameInformation.animationFrameCount)
	stringBuilder.WriteString("-----------------------------------------------\n")
	stringBuilder.WriteString("\n")
	stringBuilder.WriteString(fmt.Sprintf("Frame number:          %d of %d   (animation progress %.2f%%)\n", frameInformation.frameIndex+1, frameInformation.animationFrameCount, progress*100.0))
	// stringBuilder.WriteString(fmt.Sprintf("Frame label:           %d\n", frameInformation.frameIndex))
	stringBuilder.WriteString(fmt.Sprintf("Frame image file:      %s\n", frameInformation.imageFilename))
	stringBuilder.WriteString("\n")
	stringBuilder.WriteString(fmt.Sprintf("Render algorithm:      %s\n", frameInformation.renderAlgorithm))
	stringBuilder.WriteString(fmt.Sprintf("Image size:            %dx%d %s\n", frameInformation.imageWidth, frameInformation.imageHeight, mp4CreationWarning))
	stringBuilder.WriteString(fmt.Sprintf("Amount samples/pixel:  %d\n", frameInformation.samplesPerPixel))
	stringBuilder.WriteString(fmt.Sprintf("Max recursion depth:   %d\n", frameInformation.maxRecursionDepth))
	stringBuilder.WriteString("\n")
	if frameInformation.amountFacets > 0 {
		stringBuilder.WriteString(fmt.Sprintf("Amount facets:         %d\n", frameInformation.amountFacets))
	}
	if frameInformation.amountSpheres > 0 {
		stringBuilder.WriteString(fmt.Sprintf("Amount spheres:        %d\n", frameInformation.amountSpheres))
	}
	if frameInformation.amountDiscs > 0 {
		stringBuilder.WriteString(fmt.Sprintf("Amount discs:          %d\n", frameInformation.amountDiscs))
	}

	return stringBuilder.String()
}

func writeRenderedImage(animation scn.Animation, frame *scn.Frame, renderedPixelData *image.FloatImage, frameInformation RenderFrameInformation) {
	animationDirectory := filepath.Join(".", "rendered", animation.AnimationName)

	animationFrameFilename := filepath.Join(animationDirectory, frame.Filename+".png")
	os.MkdirAll(animationDirectory, os.ModePerm)
	image.WriteImage(animationFrameFilename, renderedPixelData)

	if animation.WriteRawImageFile {
		animationFrameRawFilename := filepath.Join(animationDirectory, frame.Filename+".praw")
		image.WriteRawImage(animationFrameRawFilename, renderedPixelData)
	}

	if animation.WriteImageInfoFile {
		frameInfoTextFilename := filepath.Join(animationDirectory, frame.Filename+".txt")
		frameInfoText := fmt.Sprintf("%s\n%s", frameInformationPreRenderText(frameInformation), frameInformationPostRenderText(frameInformation))
		err := os.WriteFile(frameInfoTextFilename, []byte(frameInfoText), 0644)
		if err != nil {
			panic("could not write text information for frame \"" + frameInformation.imageFilename + "\"")
		}
	}
}

func initializeScene(scene *scn.SceneNode) {
	_initializeScene(scene)
	scene.UpdateBounds()
}

func _initializeScene(scene *scn.SceneNode) {
	// fmt.Printf("Scene: %+v\n", scene)

	discs := scene.GetDiscs()
	for _, disc := range discs {
		disc.Initialize()
	}

	spheres := scene.GetSpheres()
	if len(spheres) < 10 {
		for _, sphere := range spheres {
			sphere.Initialize()
		}
	} else {
		subSceneNodeStructures := subdivideSpheres(spheres)

		if len(subSceneNodeStructures) > 1 {
			scene.ChildNodes = append(scene.ChildNodes, subSceneNodeStructures...)
			scene.Spheres = nil
		}
	}

	facetStructures := scene.GetFacetStructures()

	for _, facetStructure := range facetStructures {
		facetStructure.SplitMultiPointFacets()
	}

	// Subdivide facet structure for performance
	for _, facetStructure := range facetStructures {
		facetStructure.SubdivideFacetStructure(15, 0)
	}

	// Initialize facet structures (calculate bounds etc)
	for _, facetStructure := range facetStructures {
		facetStructure.Initialize()
	}

	for _, sceneNode := range scene.ChildNodes {
		initializeScene(sceneNode)
	}
}

func subdivideSpheres(spheres []*scn.Sphere) []*scn.SceneNode {
	bounds := scn.NewBounds()
	for _, sphere := range spheres {
		bounds.AddBounds(sphere.Bounds())
	}

	center := bounds.Center()
	subSceneNodeStructures := make([]*scn.SceneNode, 8)

	for _, sphere := range spheres {
		substructureIndex := 0

		if sphere.Origin[0] >= center[0] {
			substructureIndex |= 0b001
		}
		if sphere.Origin[1] >= center[1] {
			substructureIndex |= 0b010
		}
		if sphere.Origin[2] >= center[2] {
			substructureIndex |= 0b100
		}

		if subSceneNodeStructures[substructureIndex] == nil {
			subSceneNodeStructures[substructureIndex] = &scn.SceneNode{}
		}

		//fmt.Printf("Substructure: %d   Center: %+v   Bounds:%+v\n", substructureIndex, center, bounds)

		subSceneNodeStructures[substructureIndex].Spheres = append(subSceneNodeStructures[substructureIndex].Spheres, sphere)
	}

	for i := 0; i < len(subSceneNodeStructures); {
		if subSceneNodeStructures[i] == nil {
			subSceneNodeStructures = append(subSceneNodeStructures[:i], subSceneNodeStructures[i+1:]...)
		} else {
			i++
		}
	}

	// for i := 0; i < len(subSceneNodeStructures); i++ {
	// 	fmt.Printf("Substructure %d (of %d) has %d spheres (total amount spheres %d).\n", i+1, len(subSceneNodeStructures), len(subSceneNodeStructures[i].Spheres), len(spheres))
	// }

	return subSceneNodeStructures
}

func deInitializeScene(scene *scn.SceneNode) {
	scene.Clear()

	discs := scene.GetDiscs()

	for _, disc := range discs {
		projection := disc.Material.Projection
		if projection != nil {
			projection.ClearProjection()
		}
	}

	spheres := scene.GetSpheres()

	for _, sphere := range spheres {
		projection := sphere.Material.Projection
		if projection != nil {
			projection.ClearProjection()
		}
	}
}

func render(camera *scn.Camera, scene *scn.SceneNode, width int, height int, renderedPixelData *image.FloatImage, rm *rendermonitor.RenderMonitor) {
	var wg sync.WaitGroup

	amountSamples := camera.Samples

	progressbar := progressbar2.NewOptions(width*height*amountSamples+1+1, // Stay on 99% until all worker threads are done
		progressbar2.OptionFullWidth(),
		progressbar2.OptionClearOnFinish(),
		progressbar2.OptionSetRenderBlankState(true),
		progressbar2.OptionSetPredictTime(true),
		progressbar2.OptionEnableColorCodes(true),
		progressbar2.OptionSetDescription("Render progress"),
	)

	progressbar.Add(1) // Indicate start

	renderPasses := renderpass.CreateRenderPasses(20)
	for _, renderPass := range renderPasses.RenderPasses {
		for y := 0; (y + renderPass.Dy) < height; y += renderPasses.MaxPixelHeight {
			wg.Add(1)
			go parallelPixelRendering(renderedPixelData, camera, scene, width, height, y, renderPass, renderPasses.MaxPixelWidth, amountSamples, &wg, progressbar, rm)
		}
		wg.Wait()
	}

	wg.Wait()

	progressbar.Add(1) // Indicate end, final step to 100% in progress bar
	//progressbar.Clear()

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			renderedPixelData.GetPixel(x, y).Divide(float32(amountSamples))
		}
	}
}

func parallelPixelRendering(renderedPixelData *image.FloatImage, camera *scn.Camera, scene *scn.SceneNode, width int, height int,
	y int, renderPass renderpass.RenderPass, maxPixelWidth int, amountSamples int, wg *sync.WaitGroup, progressbar *progressbar2.ProgressBar, rm *rendermonitor.RenderMonitor) {

	defer wg.Done()

	defaultRenderContext := scn.NewMaterial().N("default render context").C(color.White).T(1.0, true, scn.RefractionIndex_Air)
	rayContexts := []*scn.Material{defaultRenderContext}

	// Debug ray at specified pixel
	if debugPixel.y == y && debugPixel.x >= 0 && debugPixel.y >= 0 {
		fmt.Printf("debugging at pixel (%d, %d)...\n", debugPixel.x, debugPixel.y)

		cameraRay := scn.CreateCameraRay(debugPixel.x, debugPixel.y, width, height, camera, 1)
		tracePath(cameraRay, camera, scene, 0, rayContexts)
	}

	for x := 0; (x + renderPass.Dx) < width; x += maxPixelWidth {
		for sampleIndex := 0; sampleIndex < amountSamples; sampleIndex++ {
			cameraRay := scn.CreateCameraRay(x+renderPass.Dx, y+renderPass.Dy, width, height, camera, sampleIndex)
			col := tracePath(cameraRay, camera, scene, 0, rayContexts)
			renderedPixelData.GetPixel(x+renderPass.Dx, y+renderPass.Dy).ChannelAdd(col)

			progressbar.Add(1)
		}

		// "Log" progress to render monitor
		pixelColor := renderedPixelData.GetPixel(x+renderPass.Dx, y+renderPass.Dy)
		rm.SetPixel(x+renderPass.Dx, y+renderPass.Dy, renderPass.PaintWidth, renderPass.PaintHeight, pixelColor, amountSamples)
	}
}

// FresnelReflectAmount
//
// refractionIndex1 is the index of the medium that we come from.
// refractionIndex1 is the index of the medium that we hit.
// incident is the direction vector of the ray, the direction in which we travelled.
// normal is the normal of the surface we hit (pointing more or less towards our incident vector.
//
// https://blog.demofox.org/2020/06/14/casual-shadertoy-path-tracing-3-fresnel-rough-refraction-absorption-orbit-camera/
func FresnelReflectAmount(refractionIndex1 float64, refractionIndex2 float64, normal *vec3.T, incident *vec3.T, minReflection float64, maxReflection float64) float64 {
	// Schlick approximation
	r0 := (refractionIndex1 - refractionIndex2) / (refractionIndex1 + refractionIndex2)
	r0 *= r0

	cosX := -vec3.Dot(normal, incident)
	if refractionIndex1 > refractionIndex2 {
		n := refractionIndex1 / refractionIndex2
		sinT2 := n * n * (1.0 - cosX*cosX)

		// Total internal reflection
		if sinT2 > 1.0 {
			return maxReflection
		}
		cosX = math.Sqrt(1.0 - sinT2)
	}

	x := 1.0 - cosX
	ret := r0 + (1.0-r0)*x*x*x*x*x

	// adjust reflect multiplier for object reflectivity
	return minReflection*(1.0-ret) + (maxReflection * ret)
}

func getRandomHemisphereVector(hemisphereHeading *vec3.T) *vec3.T {
	var vector vec3.T

	for continueLoop := true; continueLoop; continueLoop = vector.LengthSqr() > 1.0 {
		vector = vec3.T{
			rand.Float64()*2.0 - 1.0,
			rand.Float64()*2.0 - 1.0,
			rand.Float64()*2.0 - 1.0,
		}
	}

	// Check with dot product (really just sign check)
	// if created random vector has an angle < 90 deg to the heading vector.
	// Math: dot_product = a·b / (|a|*|b|) ; thus only the dot part will change the sign of dot product
	inHemisphere := (vector[0]*hemisphereHeading[0] + vector[1]*hemisphereHeading[1] + vector[2]*hemisphereHeading[2]) >= 0
	if !inHemisphere {
		// If the created vector is not pointing in the hemisphere direction the just flip it around
		vector.Invert()
	}

	vector.Normalize()

	return &vector
}

// getRandomCosineWeightedHemispherePoint gets a unit vector in a hemisphere "facing" the direction of vector n.
// The hemisphere is cosine weighted i.e. it gives a weighted distribution of vectors towards the "top" of the hemisphere.
//
// https://www.csie.ntu.edu.tw/~cyy/courses/rendering/05fall/lectures/handouts/lec10_mc_4up.pdf (page 12)
func getRandomCosineWeightedHemisphereVector(n *vec3.T) *vec3.T {
	amountPoints := 10000
	x, y := sunflower.Sunflower(amountPoints, 0.0, rand.Intn(amountPoints), true)
	// ret.z = sqrtf(max(0.f,1.f - ret.x*ret.x - ret.y*ret.y));
	z := math.Sqrt(math.Max(0.0, 1.0-x*x-y*y))
	generatedUnitHemisphereVector := vec3.T{x, y, z}

	var t vec3.T
	a := math.Abs(n[0])
	b := math.Abs(n[1])
	c := math.Abs(n[2])

	// Get the unit vector that is "most orthogonal" to the vector n.
	if a <= b && a <= c {
		t = vec3.UnitX
	} else if b <= a && b <= c {
		t = vec3.UnitY
	} else {
		t = vec3.UnitZ
	}

	u := vec3.Cross(&t, n)
	u.Normalize()
	v := vec3.Cross(n, &u)
	v.Normalize()

	m := mat3.T{u, v, *n}
	hemisphereVector := m.MulVec3(&generatedUnitHemisphereVector)
	return &hemisphereVector
}

func tracePath(ray *scn.Ray, camera *scn.Camera, scene *scn.SceneNode, currentDepth int, rayContexts []*scn.Material) color.Color {
	outgoingEmission := color.Black

	if currentDepth > camera.RecursionDepth {
		return outgoingEmission
	}

	ii := NewIntersectionInformation() // Information on the closest intersection

	var sceneNodeStack scn.SceneNodeStack
	sceneNodeStack.Push(scene) // Put the root scene node initially onto the scene node stack

	for !sceneNodeStack.IsEmpty() {
		currentSceneNode, _ := sceneNodeStack.Pop()

		traverseCurrentSceneNode := (currentSceneNode.Bounds == nil) || scn.BoundingBoxIntersection1(ray, currentSceneNode.Bounds)
		if traverseCurrentSceneNode {

			if currentSceneNode.HasChildNodes() {
				sceneNodeStack.PushAll(currentSceneNode.GetChildNodes())
			}

			for _, sphere := range currentSceneNode.GetSpheres() {
				processSphereIntersection(ray, sphere, ii)
			}

			for _, disc := range currentSceneNode.GetDiscs() {
				processDiscIntersection(ray, disc, ii)
			}

			for _, facetStructure := range currentSceneNode.GetFacetStructures() {
				processFacetStructureIntersection(ray, facetStructure, ii)
			}

		}
	}

	if ii.intersection {
		if ii.material == nil {
			ii.material = scn.NewMaterial() // Default material, if not specified, is matte diffuse white
		}

		projectionColor := &color.White // Default value if no projection is applied
		if ii.material.Projection != nil {
			projectionColor = ii.material.Projection.GetColor(ii.intersectionPoint)
		}

		if camera.RenderType == scn.Raycasting || camera.RenderType == "" {
			incomingRayInverted := ray.Heading.Inverted()
			cosineIncomingRayAndNormal := util.Cosine(ii.normalAtIntersection, &incomingRayInverted)

			outgoingEmission = color.Color{
				R: ii.material.Color.R * float32(cosineIncomingRayAndNormal) * projectionColor.R,
				G: ii.material.Color.G * float32(cosineIncomingRayAndNormal) * projectionColor.G,
				B: ii.material.Color.B * float32(cosineIncomingRayAndNormal) * projectionColor.B,
			}

		} else if camera.RenderType == scn.Pathtracing {

			if !ii.material.RayTerminator {
				var newRayHeading *vec3.T

				// Flip normal if it is pointing away from the incoming ray on a non-solid object
				if !ii.material.SolidObject && util.CosinePositive(ii.normalAtIntersection, ray.Heading) {
					ii.normalAtIntersection.Invert()
				}
				isIngoingRay := util.CosineNegative(ii.normalAtIntersection, ray.Heading)

				currentRayContext := rayContexts[len(rayContexts)-1]

				var reflectionProbability float64
				if isIngoingRay {
					// Add Fresnel reflection if it is an ingoing array (for now)
					reflectionProbability = FresnelReflectAmount(currentRayContext.RefractionIndex, ii.material.RefractionIndex, ii.normalAtIntersection, ray.Heading, ii.material.Glossiness, 1.0)
				} else {
					// previousRayContext := rayContexts[len(rayContexts)-2]
					// normal := ii.normalAtIntersection.Inverted()
					// reflectionProbability = FresnelReflectAmount(currentRayContext.RefractionIndex, previousRayContext.RefractionIndex, &normal, ray.Heading, ii.material.Glossiness, 1.0)

					reflectionProbability = ii.material.Glossiness
				}

				probabilitySum := reflectionProbability + ii.material.Transparency + ii.material.Diffuse
				probabilityValue := rand.Float64() * probabilitySum

				useReflectionRay := probabilityValue < reflectionProbability
				useTransparencyRay := !useReflectionRay && (probabilityValue < (reflectionProbability + ii.material.Transparency))
				useDiffuseRay := !useReflectionRay && !useTransparencyRay

				diffuseHeading := getRandomCosineWeightedHemisphereVector(ii.normalAtIntersection)
				cosineNewRayAndNormal := 1.0

				// Uniform random hemisphere sampling
				//diffuseHeading := getRandomHemisphereVector(ii.normalAtIntersection)
				//cosineNewRayAndNormal := 1.0

				if useDiffuseRay {
					// Weight for cosine weighted hemisphere sampling
					cosineNewRayAndNormal = 0.5 // remove the cosine factor as it is already included in hemisphere sampling
					newRayHeading = diffuseHeading

					// Uniform random hemisphere sampling
					// cosineNewRayAndNormal = vec3.Dot(ii.normalAtIntersection, newRayHeading) / (ii.normalAtIntersection.Length() * newRayHeading.Length())

				} else if useReflectionRay {
					reflectionHeading := getReflectionVector(ii.normalAtIntersection, ray.Heading)

					interpolationWeight := ii.material.Roughness * ii.material.Roughness
					interpolatedHeading := vec3.Interpolate(reflectionHeading, diffuseHeading, interpolationWeight)
					interpolatedHeading.Normalize()

					// Weight for cosine weighted hemisphere sampling
					cosineNewRayAndNormal = 0.5*interpolationWeight + (1.0 - interpolationWeight) // Interpolated weight diffuse --> specular
					newRayHeading = &interpolatedHeading

				} else if useTransparencyRay {
					if ii.material.SolidObject && (ii.material.RefractionIndex > 0.0) {

						if isIngoingRay {
							// Ingoing ray to a solid object with refraction index
							//fmt.Printf("Ingoing... %s\n", ii.material.Name)

							var totalInternalReflection bool
							newRayHeading, totalInternalReflection = getRefractionVector(ii.normalAtIntersection, ray.Heading, currentRayContext.RefractionIndex, ii.material.RefractionIndex)

							if !totalInternalReflection {
								rayContexts = append(rayContexts, ii.material)
							}
						} else {
							// Outgoing ray from a solid object with refraction index
							//fmt.Printf("Outgoing... %s\n", ii.material.Name)

							// Outgoing ray from a solid object with refraction index
							rayContexts = rayContexts[:len(rayContexts)-1] // Pop off current ray context
							if len(rayContexts) == 0 {
								panic("About to access empty ray context (after popping last context)...")
							}
							previousRayContext := rayContexts[len(rayContexts)-1] // Get previous ray context

							// Flip normal if needed, to face ray
							if util.CosinePositive(ii.normalAtIntersection, ray.Heading) {
								ii.normalAtIntersection.Invert()
							}

							var totalInternalReflection bool
							newRayHeading, totalInternalReflection = getRefractionVector(ii.normalAtIntersection, ray.Heading, currentRayContext.RefractionIndex, previousRayContext.RefractionIndex)

							if totalInternalReflection {
								rayContexts = append(rayContexts, currentRayContext) // We are not leaving current ray context, due to total internal reflection
							} else {
								// previous ray context is already on top of stack
							}
						}

						cosineNewRayAndNormal = 1.0

					} else if !ii.material.SolidObject {
						// Just pass through the object in the same direction as before.
						// The walls of the object are super thin and do not refract the ray.
						newRayHeading = ray.Heading
						cosineNewRayAndNormal = 1.0
					}
				}

				newRayHeading.Normalize() // TODO remove?

				//rayStartOffset := newRayHeading.Scaled(epsilonDistance)
				rayStartOffset := ii.normalAtIntersection.Scaled(epsilonDistance)
				if util.CosineNegative(newRayHeading, ii.normalAtIntersection) {
					(&rayStartOffset).Invert()
				}

				newRayOrigin := ii.intersectionPoint.Added(&rayStartOffset)
				newRay := scn.Ray{Origin: &newRayOrigin, Heading: newRayHeading}

				incomingEmission := tracePath(&newRay, camera, scene, currentDepth+1, rayContexts)
				incomingEmissionOnSurface := incomingEmission
				incomingEmissionOnSurface.Multiply(float32(cosineNewRayAndNormal))

				outgoingEmission = color.Color{
					R: incomingEmissionOnSurface.R * ii.material.Color.R * projectionColor.R,
					G: incomingEmissionOnSurface.G * ii.material.Color.G * projectionColor.G,
					B: incomingEmissionOnSurface.B * ii.material.Color.B * projectionColor.B,
				}
			}

			if ii.material.Emission != nil {
				outgoingEmission.R += ii.material.Emission.R * projectionColor.R
				outgoingEmission.G += ii.material.Emission.G * projectionColor.G
				outgoingEmission.B += ii.material.Emission.B * projectionColor.B
			}
		}
	}

	return outgoingEmission
}

func processFacetStructureIntersection(ray *scn.Ray, facetStructure *scn.FacetStructure, ii *IntersectionInformation) {
	tempIntersection, tmpIntersectionFacet, tempIntersectionPoint, tempIntersectionNormal, tempMaterial := scn.FacetStructureIntersection(ray, facetStructure, nil)

	if tempIntersection {
		distance := vec3.Distance(ray.Origin, tempIntersectionPoint)
		if distance < ii.shortestDistance && distance > epsilonDistance {
			ii.shortestDistance = distance               // Save the shortest intersection distance
			ii.intersection = tempIntersection           // Set to true, there has been an intersection
			ii.intersectionPoint = tempIntersectionPoint // Save the intersection point of the closest intersection
			ii.material = tempMaterial
			ii.normalAtIntersection = tempIntersectionNormal // Should be normalized from initialization

			ii.intersectedFacet = tmpIntersectionFacet
			ii.intersectedSphere = nil
			ii.intersectedDisc = nil

			if ii.material == nil {
				ii.material = scn.NewMaterial() // If, for some erroneous reason there is an intersection without any material, use default (diffuse white).
				fmt.Printf("Warning: Could not find any material for intersection point on facet structure '%s' ('%s').\n", facetStructure.Name, facetStructure.SubstructureName)
			}

			if vec3.Dot(ii.normalAtIntersection, tmpIntersectionFacet.Normal) < 0 {
				fmt.Println("rotten normals")
			}
		}
	}
}

func processDiscIntersection(ray *scn.Ray, disc *scn.Disc, ii *IntersectionInformation) {
	tempIntersection, tempIntersectionPoint, tempIntersectionNormal := scn.DiscIntersection(ray, disc)

	if tempIntersection {
		distance := vec3.Distance(ray.Origin, tempIntersectionPoint)
		if distance < ii.shortestDistance && distance > epsilonDistance {
			ii.intersection = tempIntersection           // Set to true, there has been an intersection
			ii.intersectionPoint = tempIntersectionPoint // Save the intersection point of the closest intersection
			ii.shortestDistance = distance               // Save the shortest intersection distance
			ii.material = disc.Material
			ii.normalAtIntersection = tempIntersectionNormal // Should be normalized from initialization

			ii.intersectedFacet = nil
			ii.intersectedSphere = nil
			ii.intersectedDisc = disc

			// Flip normal if it is pointing away from the incoming ray
			if util.CosinePositive(ii.normalAtIntersection, ray.Heading) {
				ii.normalAtIntersection.Invert()
			}
		}
	}
}

func processSphereIntersection(ray *scn.Ray, sphere *scn.Sphere, ii *IntersectionInformation) {
	tempIntersectionPoint, tempIntersection := scn.SphereIntersection(ray, sphere)

	if tempIntersection {
		distance := vec3.Distance(ray.Origin, tempIntersectionPoint)

		// TODO Remove
		//if distance < 0.000001 {
		//	panic(fmt.Sprintf("Not a legal intersection point, no ray has less length than 0.0001. Length: %f", distance))
		//}

		if distance < ii.shortestDistance && distance > epsilonDistance {
			ii.intersection = tempIntersection           // Set to true, there has been an intersection
			ii.intersectionPoint = tempIntersectionPoint // Save the intersection point of the closest intersection
			ii.shortestDistance = distance               // Save the shortest intersection distance
			ii.material = sphere.Material

			ii.normalAtIntersection = sphere.Normal(ii.intersectionPoint)

			ii.intersectedFacet = nil
			ii.intersectedSphere = sphere
			ii.intersectedDisc = nil

			// Flip normal if it is pointing away from the incoming ray
			//if vectorCosinePositive(normalAtIntersection, ray.Heading) {
			//	normalAtIntersection.Invert()
			//}
		}
	}
}

func getReflectionVector(normal *vec3.T, incomingVector *vec3.T) *vec3.T {
	tempV := normal.Scaled(2.0 * vec3.Dot(normal, incomingVector))
	reflectionVector := incomingVector.Subed(&tempV)
	return &reflectionVector
}

// getRefractionVector according to
// https://graphics.stanford.edu/courses/cs148-10-summer/docs/2006--degreve--reflection_refraction.pdf
func getRefractionVector(normal *vec3.T, incomingVector *vec3.T, leavingRefractionIndex float64, enteringRefractionIndex float64) (outgoingVector *vec3.T, totalInternalReflection bool) {
	outgoingVector = incomingVector // No refraction

	refractionRatio := leavingRefractionIndex / enteringRefractionIndex
	cosi := -vec3.Dot(incomingVector, normal)                        // Cosine for angle of incoming vector and surface normal
	sinlsqr := refractionRatio * refractionRatio * (1.0 - cosi*cosi) // Squared sinus for angle between refraction (leaving) vector and inverted normal

	// If the incoming vector angle is to flat to the surface of an optically lighter material then
	// total reflection occur. (Like the mirror effect on the water surface when you are diving and looking up.)
	// Calculate the reflection vector instead.
	if sinlsqr > 1.0 {
		return getReflectionVector(normal, incomingVector), true
	}

	cosl := math.Sqrt(1.0 - sinlsqr) // Need to verify that this part actually is "cosine" of angle

	io := incomingVector.Scaled(refractionRatio)     // Incoming vector part of outgoing (refraction) vector
	no := normal.Scaled(refractionRatio*cosi - cosl) // Normal vector part of outgoing (refraction) vector

	io.Add(&no)
	outgoingVector = &io

	return outgoingVector, false
}
