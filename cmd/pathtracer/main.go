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
	"pathtracer/internal/pkg/util"
	"strconv"
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
	//debugPixel = struct{ x, y int }{x: 600, y: 300}
)

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
		frameStartTimestamp := time.Now()
		renderMonitor.Initialize(animation.AnimationName, frame.Filename, animation.Width, animation.Height)
		time.Sleep(50 * time.Millisecond)

		var scene scn.SceneNode
		scene = *frame.SceneNode

		unevenWidth := (animation.Width%2) == 1 || (animation.Height%2) == 1
		mp4CreationWarning := ""
		if unevenWidth && (len(animation.Frames) > 1) {
			mp4CreationWarning = " (uneven width or height, no animation can be made)"
		}

		progress := float64(frameIndex+1) / float64(len(animation.Frames))
		fmt.Println("-----------------------------------------------")
		fmt.Println("Frame number:          ", frameIndex+1, "of", len(animation.Frames), "   (animation progress "+fmt.Sprintf("%.2f", progress*100.0)+"%)")
		fmt.Println("Frame label:           ", frame.FrameIndex)
		fmt.Println("Frame image file:      ", frame.Filename)
		fmt.Println()
		fmt.Println("Render algorithm:      ", frame.Camera.RenderType)
		fmt.Println("Image size:            ", strconv.Itoa(animation.Width)+"x"+strconv.Itoa(animation.Height)+mp4CreationWarning)
		fmt.Println("Amount samples/pixel:  ", frame.Camera.Samples)
		fmt.Println("Max recursion depth:   ", frame.Camera.RecursionDepth)
		fmt.Println()
		if scene.GetAmountFacets() > 0 {
			fmt.Println("Amount facets:         ", scene.GetAmountFacets())
		}
		if scene.GetAmountSpheres() > 0 {
			fmt.Println("Amount spheres:        ", scene.GetAmountSpheres())
		}
		if scene.GetAmountDiscs() > 0 {
			fmt.Println("Amount discs:          ", scene.GetAmountDiscs())
		}
		fmt.Println()

		fmt.Println("Initialize scene...")
		initializeScene(&scene)

		renderedPixelData := image.NewFloatImage(animation.AnimationName, animation.Width, animation.Height)

		fmt.Println("Rendering...")
		render(frame.Camera, &scene, animation.Width, animation.Height, renderedPixelData, &renderMonitor)

		writeRenderedImage(animation, frame, renderedPixelData)

		deInitializeScene(&scene)
		frame.SceneNode = nil
		fmt.Println("Releasing resources...")
		fmt.Println()

		fmt.Println("Frame render time:", time.Since(frameStartTimestamp))
	}

	fmt.Printf("Total execution time (for %d frames): %s\n", len(animation.Frames), time.Since(startTimestamp))
}

func writeRenderedImage(animation scn.Animation, frame scn.Frame, renderedPixelData *image.FloatImage) {
	animationDirectory := filepath.Join(".", "rendered", animation.AnimationName)
	animationFrameFilename := filepath.Join(animationDirectory, frame.Filename+".png")
	os.MkdirAll(animationDirectory, os.ModePerm)
	image.WriteImage(animationFrameFilename, renderedPixelData)

	if animation.WriteRawImageFile {
		animationFrameRawFilename := filepath.Join(animationDirectory, frame.Filename+".praw")
		image.WriteRawImage(animationFrameRawFilename, renderedPixelData)
	}
}

func initializeScene(scene *scn.SceneNode) {
	// fmt.Printf("Scene: %+v\n", scene)

	discs := scene.GetDiscs()
	for _, disc := range discs {
		disc.Initialize()
	}

	spheres := scene.GetSpheres()
	if len(spheres) < 25 {
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
	(*scene).Clear()

	discs := (*scene).GetDiscs()

	for _, disc := range discs {
		projection := disc.Material.Projection
		if projection != nil {
			projection.ClearProjection()
		}
	}

	spheres := (*scene).GetSpheres()

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

	defaultRenderContext := &scn.Material{Name: "default render context", Color: &color.White, Transparency: 1.0, SolidObject: true, RefractionIndex: scn.RefractionIndex_Air}
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
			renderedPixelData.GetPixel(x+renderPass.Dx, y+renderPass.Dy).ChannelAdd(&col)

			progressbar.Add(1)
		}

		// "Log" progress to render monitor
		pixelColor := renderedPixelData.GetPixel(x+renderPass.Dx, y+renderPass.Dy)
		rm.SetPixel(x+renderPass.Dx, y+renderPass.Dy, renderPass.PaintWidth, renderPass.PaintHeight, pixelColor, amountSamples)
	}
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
	x, y := util.Sunflower(amountPoints, 0.0, rand.Intn(amountPoints), true)
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

	intersection := false               // Intersection occurred? True/false
	intersectionPoint := &vec3.Zero     // Point of intersection
	shortestDistance := math.MaxFloat64 // At what distance from start point of fired ray
	material := &scn.Material{}         // The material of the closest object that was intersected
	normalAtIntersection := &vec3.Zero  // The normal of the object that was intersected, at intersection point

	var sceneNodeStack scn.SceneNodeStack
	sceneNodeStack.Push(scene) // Put the root scene node initially onto the scene node stack

	for !sceneNodeStack.IsEmpty() {
		currentSceneNode, _ := sceneNodeStack.Pop()

		traverseCurrentSceneNode := (currentSceneNode.Bounds == nil) || scn.BoundingBoxIntersection1(ray, currentSceneNode.Bounds)
		if traverseCurrentSceneNode {

			if currentSceneNode.HasChildNodes() {
				sceneNodeStack.PushAll(currentSceneNode.GetChildNodes())
			}

			for _, sphere := range (*currentSceneNode).GetSpheres() {
				tempIntersectionPoint, tempIntersection := scn.SphereIntersection(ray, sphere)
				if tempIntersection {
					distance := vec3.Distance(ray.Origin, tempIntersectionPoint)

					// TODO Remove
					//if distance < 0.000001 {
					//	panic(fmt.Sprintf("Not a legal intersection point, no ray has less length than 0.0001. Length: %f", distance))
					//}

					if distance < shortestDistance && distance > epsilonDistance {
						intersection = tempIntersection           // Set to true, there has been an intersection
						intersectionPoint = tempIntersectionPoint // Save the intersection point of the closest intersection
						shortestDistance = distance               // Save the shortest intersection distance
						material = sphere.Material

						normalAtIntersection = sphere.Normal(intersectionPoint)

						// Flip normal if it is pointing away from the incoming ray
						//if vectorCosinePositive(normalAtIntersection, ray.Heading) {
						//	normalAtIntersection.Invert()
						//}
					}
				}
			}

			for _, disc := range (*currentSceneNode).GetDiscs() {
				tempIntersectionPoint, tempIntersection := scn.DiscIntersection(ray, disc)
				if tempIntersection {
					distance := vec3.Distance(ray.Origin, tempIntersectionPoint)
					if distance < shortestDistance && distance > epsilonDistance {
						intersection = tempIntersection           // Set to true, there has been an intersection
						intersectionPoint = tempIntersectionPoint // Save the intersection point of the closest intersection
						shortestDistance = distance               // Save the shortest intersection distance
						material = disc.Material
						*normalAtIntersection = *disc.Normal // Should be normalized from initialization

						// Flip normal if it is pointing away from the incoming ray
						if vectorCosinePositive(normalAtIntersection, ray.Heading) {
							normalAtIntersection.Invert()
						}
					}
				}
			}

			for _, facetStructure := range (*currentSceneNode).GetFacetStructures() {
				tempIntersection, tempIntersectionPoint, tempIntersectionNormal, tempMaterial := scn.FacetStructureIntersection(ray, facetStructure)

				if tempIntersection {
					distance := vec3.Distance(ray.Origin, tempIntersectionPoint)
					if distance < shortestDistance && distance > epsilonDistance {
						shortestDistance = distance               // Save the shortest intersection distance
						intersection = tempIntersection           // Set to true, there has been an intersection
						intersectionPoint = tempIntersectionPoint // Save the intersection point of the closest intersection
						material = tempMaterial
						normalAtIntersection = tempIntersectionNormal // Should be normalized from initialization

						// Flip normal if it is pointing away from the incoming ray
						//if vectorCosinePositive(normalAtIntersection, ray.Heading) {
						//	normalAtIntersection.Invert()
						//}
					}
				}
			}

		}
	}

	if intersection {
		if material == nil {
			material = scn.NewMaterial() // Default material if not specified is matte diffuse white
		}

		projectionColor := &color.White // Default value if no projection is applied
		if material.Projection != nil {
			projectionColor = material.Projection.GetColor(intersectionPoint)
		}

		if camera.RenderType == scn.Raycasting || camera.RenderType == "" {
			incomingRayInverted := ray.Heading.Inverted()
			cosineIncomingRayAndNormal := vectorCosine(normalAtIntersection, &incomingRayInverted)

			outgoingEmission = color.Color{
				R: material.Color.R * float32(cosineIncomingRayAndNormal) * projectionColor.R,
				G: material.Color.G * float32(cosineIncomingRayAndNormal) * projectionColor.G,
				B: material.Color.B * float32(cosineIncomingRayAndNormal) * projectionColor.B,
			}

		} else if camera.RenderType == scn.Pathtracing {

			if !material.RayTerminator {
				var newRayHeading *vec3.T

				useTransparencyRay := (material.Transparency > 0.0) && (rand.Float64() < material.Transparency)

				cosineNewRayAndNormal := 1.0

				if useTransparencyRay && material.SolidObject && (material.RefractionIndex > 0.0) {
					isIngoingRay := vectorCosineNegative(normalAtIntersection, ray.Heading)

					if isIngoingRay {
						// Ingoing ray to a solid object with refraction index
						currentRayContext := rayContexts[len(rayContexts)-1]

						var totalInternalReflection bool
						newRayHeading, totalInternalReflection = getRefractionVector(normalAtIntersection, ray.Heading, currentRayContext.RefractionIndex, material.RefractionIndex)

						if !totalInternalReflection {
							rayContexts = append(rayContexts, material)
						}
					} else {
						// Outgoing ray from a solid object with refraction index
						currentRayContext := rayContexts[len(rayContexts)-1] // Get current ray context
						rayContexts = rayContexts[:len(rayContexts)-1]       // Pop off current ray context
						if len(rayContexts) == 0 {
							panic("About to access empty ray context (after popping last context)...")
						}
						previousRayContext := rayContexts[len(rayContexts)-1] // Get previous ray context

						// Flip normal if needed, to face ray
						if vectorCosinePositive(normalAtIntersection, ray.Heading) {
							normalAtIntersection.Invert()
						}

						var totalInternalReflection bool
						newRayHeading, totalInternalReflection = getRefractionVector(normalAtIntersection, ray.Heading, currentRayContext.RefractionIndex, previousRayContext.RefractionIndex)

						if totalInternalReflection {
							rayContexts = append(rayContexts, currentRayContext) // We are not leaving current ray context, due to total internal reflection
						} else {
							// previous ray context is already on top of stack
						}
					}

					cosineNewRayAndNormal = 1.0

				} else if useTransparencyRay && !material.SolidObject {
					newRayHeading = ray.Heading

					cosineNewRayAndNormal = 1.0

				} else {
					// Flip normal if it is pointing away from the incoming ray
					if vectorCosinePositive(normalAtIntersection, ray.Heading) {
						normalAtIntersection.Invert()
					}

					// diffuseHeading := getRandomHemisphereVector(normalAtIntersection)
					diffuseHeading := getRandomCosineWeightedHemisphereVector(normalAtIntersection)

					useReflectionRay := rand.Float64() < material.Glossiness

					if useReflectionRay {
						reflectionHeading := getReflectionVector(normalAtIntersection, ray.Heading)

						interpolationWeight := material.Roughness * material.Roughness
						interpolatedHeading := vec3.Interpolate(reflectionHeading, diffuseHeading, interpolationWeight)
						interpolatedHeading.Normalize()

						newRayHeading = &interpolatedHeading

						// Weight for random hemisphere sampling
						//cosineNewRayAndNormal = vec3.Dot(normalAtIntersection, newRayHeading) / (normalAtIntersection.Length() * newRayHeading.Length())
						//cosineNewRayAndNormal = cosineNewRayAndNormal*interpolationWeight + (1.0 - interpolationWeight) // Interpolated weight diffuse --> specular

						// Weight for cosine weighted hemisphere sampling
						cosineNewRayAndNormal = 0.5*interpolationWeight + (1.0 - interpolationWeight) // Interpolated weight diffuse --> specular

					} else {
						// cosine weighted hemisphere sampling
						//diffuseHeading.Add(normalAtIntersection)
						//diffuseHeading.Normalize()

						newRayHeading = diffuseHeading

						// Weight for cosine weighted hemisphere sampling
						cosineNewRayAndNormal = 0.5 // remove the cosine factor as it is already included in hemisphere sampling

						// Weight for random hemisphere sampling
						// cosineNewRayAndNormal = vec3.Dot(normalAtIntersection, newRayHeading) / (normalAtIntersection.Length() * newRayHeading.Length())
					}
				}

				newRayHeading.Normalize() // TODO remove?

				//rayStartOffset := newRayHeading.Scaled(epsilonDistance)
				rayStartOffset := normalAtIntersection.Scaled(epsilonDistance)
				if vectorCosineNegative(newRayHeading, normalAtIntersection) {
					(&rayStartOffset).Invert()
				}

				newRayOrigin := intersectionPoint.Added(&rayStartOffset)
				newRay := scn.Ray{Origin: &newRayOrigin, Heading: newRayHeading}

				incomingEmission := tracePath(&newRay, camera, scene, currentDepth+1, rayContexts)
				incomingEmissionOnSurface := incomingEmission
				incomingEmissionOnSurface.Multiply(float32(cosineNewRayAndNormal))

				outgoingEmission = color.Color{
					//R: material.Color.R * projectionColor.R * (incomingEmission.R * float32(cosineNewRayAndNormal)*material.Glossiness),
					//G: material.Color.G * projectionColor.G * (incomingEmission.G * float32(cosineNewRayAndNormal)*material.Glossiness),
					//B: material.Color.B * projectionColor.B * (incomingEmission.B * float32(cosineNewRayAndNormal)*material.Glossiness),

					R: incomingEmissionOnSurface.R*float32(material.Glossiness) + incomingEmissionOnSurface.R*material.Color.R*projectionColor.R*float32(1.0-material.Glossiness),
					G: incomingEmissionOnSurface.G*float32(material.Glossiness) + incomingEmissionOnSurface.G*material.Color.G*projectionColor.G*float32(1.0-material.Glossiness),
					B: incomingEmissionOnSurface.B*float32(material.Glossiness) + incomingEmissionOnSurface.B*material.Color.B*projectionColor.B*float32(1.0-material.Glossiness),

					// The multiplication by 0.5 is actually a division by 2, to normalize the added light as there are two light rays fired and added together.
					// R: 0.5 * (float32(cosineNewRayAndNormalDiffuse)*incomingEmissionDiffuse.R*material.Color.R*projectionColor.R + float32(cosineNewRayAndNormal)*incomingEmission.R*material.Glossiness),
					// G: 0.5 * (float32(cosineNewRayAndNormalDiffuse)*incomingEmissionDiffuse.G*material.Color.G*projectionColor.G + float32(cosineNewRayAndNormal)*incomingEmission.G*material.Glossiness),
					// B: 0.5 * (float32(cosineNewRayAndNormalDiffuse)*incomingEmissionDiffuse.B*material.Color.B*projectionColor.B + float32(cosineNewRayAndNormal)*incomingEmission.B*material.Glossiness),
				}
			}

			if material.Emission != nil {
				outgoingEmission.R += material.Emission.R * projectionColor.R
				outgoingEmission.G += material.Emission.G * projectionColor.G
				outgoingEmission.B += material.Emission.B * projectionColor.B
			}
		}
	}

	return outgoingEmission
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

func vectorCosine(a *vec3.T, b *vec3.T) float64 {
	return vec3.Dot(a, b) / math.Sqrt(a.LengthSqr()*b.LengthSqr())
}

func vectorCosinePositive(a *vec3.T, b *vec3.T) bool {
	return vec3.Dot(a, b) >= 0
}

func vectorCosineNegative(a *vec3.T, b *vec3.T) bool {
	return vec3.Dot(a, b) < 0
}
