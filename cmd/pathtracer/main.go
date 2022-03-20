package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/image"
	scn "pathtracer/internal/pkg/scene"
	"strconv"
	"sync"
	"time"

	progressbar2 "github.com/schollz/progressbar/v3"
	"github.com/ungerik/go3d/float64/vec3"
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

	var animationJson, err = os.ReadFile(animationFilename)
	if err != nil {
		panic(err)
	}

	animation := scn.Animation{}
	err = json.Unmarshal(animationJson, &animation)
	if err != nil {
		panic(err)
	}

	fmt.Println("-----------------------------------------------")
	fmt.Println("Animation file: ", animationFilename)
	fmt.Println("Animation name: ", animation.AnimationName)
	fmt.Println("Amount frames:  ", len(animation.Frames))
	fmt.Println()

	for frameIndex, frame := range animation.Frames {
		frameStartTimestamp := time.Now()
		scene := frame.Scene

		progress := float64(frameIndex+1) / float64(len(animation.Frames))
		fmt.Println("-----------------------------------------------")
		fmt.Println("Frame number:     ", frameIndex+1, "of", len(animation.Frames), "   (progression "+fmt.Sprintf("%.2f", progress*100.0)+"%)")
		fmt.Println("Frame label:      ", frame.FrameIndex)
		fmt.Println("Frame image file: ", frame.Filename)
		fmt.Println()
		fmt.Println("Render algorithm: ", scene.Camera.RenderType)
		fmt.Println("Image size:       ", strconv.Itoa(animation.Width)+"x"+strconv.Itoa(animation.Height))
		fmt.Println("Amount samples:   ", scene.Camera.Samples)
		fmt.Println("Max recursion:    ", scene.Camera.RecursionDepth)
		fmt.Println("Amount discs:     ", len(scene.Discs))
		fmt.Println("Amount spheres:   ", len(scene.Spheres))
		fmt.Println()

		fmt.Println("Initialize scene...")
		initializeScene(&scene)

		renderedPixelData := image.NewFloatImage(animation.Width, animation.Height)

		fmt.Println("Rendering...")
		render(&scene, animation.Width, animation.Height, renderedPixelData)

		animationDirectory := filepath.Join(".", "rendered", animation.AnimationName)
		animationFrameFilename := filepath.Join(animationDirectory, frame.Filename+".png")
		os.MkdirAll(animationDirectory, os.ModePerm)
		image.WriteImage(animationFrameFilename, animation.Width, animation.Height, renderedPixelData)

		if animation.WriteRawImageFile {
			animationFrameRawFilename := filepath.Join(animationDirectory, frame.Filename+".praw")
			image.WriteImage(animationFrameRawFilename, animation.Width, animation.Height, renderedPixelData)
		}

		deInitializeScene(&scene)
		frame.Scene = scn.Scene{}
		fmt.Println("Releasing resources...")
		fmt.Println()

		fmt.Println("Frame render time:", time.Since(frameStartTimestamp))
	}

	fmt.Println("Total execution time:", time.Since(startTimestamp))
}

func initializeScene(scene *scn.Scene) {
	scene.Initialize()

	discs := scene.Discs
	for _, disc := range discs {
		disc.Initialize(scene)
	}

	spheres := scene.Spheres
	for _, sphere := range spheres {
		sphere.Initialize(scene)
	}
}

func deInitializeScene(scene *scn.Scene) {
	scene.Clear()

	discs := scene.Discs

	for _, disc := range discs {
		projection := disc.Material.Projection
		if projection != nil {
			projection.ClearProjection()
		}
	}

	spheres := scene.Spheres

	for _, sphere := range spheres {
		projection := sphere.Material.Projection
		if projection != nil {
			projection.ClearProjection()
		}
	}
}

func render(scene *scn.Scene, width int, height int, renderedPixelData *image.FloatImage) {
	var wg sync.WaitGroup

	amountSamples := scene.Camera.Samples

	progressbar := progressbar2.NewOptions(width*height*amountSamples+1+1, // Stay on 99% until all worker threads are done
		progressbar2.OptionFullWidth(),
		progressbar2.OptionClearOnFinish(),
		progressbar2.OptionSetRenderBlankState(true),
		progressbar2.OptionSetPredictTime(true),
		progressbar2.OptionEnableColorCodes(true),
		progressbar2.OptionSetDescription("Render progress"),
	)

	progressbar.Add(1) // Indicate start

	for y := 0; y < height; y++ {
		wg.Add(1)
		go parallelPixelRendering(renderedPixelData, scene, width, height, y, amountSamples, &wg, progressbar)
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

func parallelPixelRendering(renderedPixelData *image.FloatImage, scene *scn.Scene, width int, height int, y int, amountSamples int, wg *sync.WaitGroup, progressbar *progressbar2.ProgressBar) {
	defer wg.Done()

	for x := 0; x < width; x++ {
		for sampleIndex := 0; sampleIndex < amountSamples; sampleIndex++ {
			cameraRay := scn.CreateCameraRay(x, y, width, height, &scene.Camera, sampleIndex)
			col := tracePath(cameraRay, scene, 0)
			renderedPixelData.GetPixel(x, y).Add(col)

			progressbar.Add(1)
		}
	}
}

func getRandomHemisphereVector(hemisphereHeading *vec3.T) vec3.T {
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
	// Math: dot_product = a·b / (|a|*|b|) ; thus only the dot part will make the sign of dot product change
	inHemisphere := (vector[0]*hemisphereHeading[0] + vector[1]*hemisphereHeading[1] + vector[2]*hemisphereHeading[2]) >= 0
	if !inHemisphere {
		// If the created vector is not pointing in the hemisphere direction the just flip it around
		vector.Invert()
	}

	vector.Normalize()

	return vector
}

func tracePath(ray *scn.Ray, scene *scn.Scene, currentDepth int) color.Color {
	outgoingEmission := color.Black

	if currentDepth > scene.Camera.RecursionDepth {
		return outgoingEmission
	}

	intersection := false               // Intersection occurred? True/false
	intersectionPoint := vec3.Zero      // Point of intersection
	shortestDistance := math.MaxFloat64 // At what distance from start point of fired ray
	material := scn.Material{}          // The material of the closest object that was intersected
	normalAtIntersection := vec3.Zero   // The normal of the object that was intersected, at intersection point

	for _, sphere := range scene.Spheres {
		tempIntersectionPoint, tempIntersection := scn.SphereIntersection(ray, &sphere)
		if tempIntersection {
			distance := vec3.Distance(&ray.Origin, &tempIntersectionPoint)
			if distance < shortestDistance {
				intersection = tempIntersection           // Set to true, there has been an intersection
				intersectionPoint = tempIntersectionPoint // Save the intersection point of the closest intersection
				shortestDistance = distance               // Save the shortest intersection distance
				material = sphere.Material

				normalAtIntersection = intersectionPoint.Subed(&sphere.Origin)
				normalAtIntersection.Normalize()

				// Flip normal if it is pointing away from the incoming ray
				if vectorCosinePositive(&normalAtIntersection, &ray.Heading) {
					normalAtIntersection.Invert()
				}
			}
		}
	}

	for _, disc := range scene.Discs {
		tempIntersectionPoint, tempIntersection := scn.DiscIntersection(ray, &disc)
		if tempIntersection {
			distance := vec3.Distance(&ray.Origin, &tempIntersectionPoint)
			if distance < shortestDistance {
				intersection = tempIntersection           // Set to true, there has been an intersection
				intersectionPoint = tempIntersectionPoint // Save the intersection point of the closest intersection
				shortestDistance = distance               // Save the shortest intersection distance
				material = disc.Material
				normalAtIntersection = disc.Normal // Should be normalized from initialization

				// Flip normal if it is pointing away from the incoming ray
				if vectorCosinePositive(&normalAtIntersection, &ray.Heading) {
					normalAtIntersection.Invert()
				}
			}
		}
	}

	if intersection {
		projectionColor := &color.White
		if material.Projection != nil {
			projectionColor = material.Projection.GetUV(&intersectionPoint)
		}

		incomingRayInverted := ray.Heading.Inverted()

		if scene.Camera.RenderType == "" || scene.Camera.RenderType == scn.Raycasting {
			cosineIncomingRayAndNormal := vectorCosine(&normalAtIntersection, &incomingRayInverted)

			outgoingEmission = color.Color{
				R: material.Color.R * float32(cosineIncomingRayAndNormal) * projectionColor.R,
				G: material.Color.G * float32(cosineIncomingRayAndNormal) * projectionColor.G,
				B: material.Color.B * float32(cosineIncomingRayAndNormal) * projectionColor.B,
			}

		} else if scene.Camera.RenderType == scn.Pathtracing {

			var newRayHeading vec3.T

			// Reflectiveness / Glossiness
			if material.Reflective == 0.0 {
				// Perfect matte surface
				newRayHeading = getRandomHemisphereVector(&normalAtIntersection)
			} else if material.Reflective == 1.0 {
				// Perfect reflective (mirror)
				newRayHeading = getReflectionVector(&normalAtIntersection, &ray.Heading)
			} else {
				// Glossy surface, somewhat reflective (on a scale from 0 to 1)
				perfectReflectionHeadingVector := getReflectionVector(&normalAtIntersection, &ray.Heading)
				perfectReflectionHeadingVector.Scale(material.Reflective)

				randomHeadingVector := getRandomHemisphereVector(&normalAtIntersection)
				randomHeadingVector.Scale(1.0 - material.Reflective)

				perfectReflectionHeadingVector.Add(&randomHeadingVector)
				perfectReflectionHeadingVector.Normalize()
				newRayHeading = perfectReflectionHeadingVector
			}

			rayStartOffset := newRayHeading.Scaled(0.000001)
			newRay := scn.Ray{
				Origin:  intersectionPoint.Added(&rayStartOffset),
				Heading: newRayHeading,
			}
			incomingEmission := tracePath(&newRay, scene, currentDepth+1)
			cosineNewRayAndNormal := vec3.Dot(&normalAtIntersection, &newRayHeading) / (normalAtIntersection.Length() * newRayHeading.Length())

			outgoingEmission = color.Color{
				R: material.Color.R * float32(cosineNewRayAndNormal) * projectionColor.R * incomingEmission.R,
				G: material.Color.G * float32(cosineNewRayAndNormal) * projectionColor.G * incomingEmission.G,
				B: material.Color.B * float32(cosineNewRayAndNormal) * projectionColor.B * incomingEmission.B,
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

func getReflectionVector(normal *vec3.T, incomingVector *vec3.T) vec3.T {
	tempV := normal.Scaled(2.0 * vec3.Dot(normal, incomingVector))
	return incomingVector.Subed(&tempV)
}

func vectorCosine(normalAtIntersection *vec3.T, incomingRayInverted *vec3.T) float64 {
	return vec3.Dot(normalAtIntersection, incomingRayInverted) / math.Sqrt(normalAtIntersection.LengthSqr()*incomingRayInverted.LengthSqr())
}

func vectorCosinePositive(normalAtIntersection *vec3.T, incomingRayInverted *vec3.T) bool {
	return vec3.Dot(normalAtIntersection, incomingRayInverted) >= 0
}

func vectorCosineNegative(normalAtIntersection *vec3.T, incomingRayInverted *vec3.T) bool {
	return vec3.Dot(normalAtIntersection, incomingRayInverted) < 0
}
