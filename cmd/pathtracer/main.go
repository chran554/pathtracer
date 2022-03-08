package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
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

	// animationFilename := "scene/three_balls.scene.json"
	// animationFilename := "scene/sphere_circle_rotation_focaldistance.animation.json"
	// imageFilename := "rendered/rendered.png"
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

		fmt.Println("-----------------------------------------------")
		fmt.Println("Frame number:     ", frameIndex+1, "of", len(animation.Frames))
		fmt.Println("Frame label:      ", frame.FrameIndex)
		fmt.Println("Frame image file: ", frame.Filename)
		fmt.Println()
		fmt.Println("Image size:       ", strconv.Itoa(animation.Width)+"x"+strconv.Itoa(animation.Height))
		fmt.Println("Amount samples:   ", scene.Camera.Samples)
		fmt.Println("Amount discs:     ", len(scene.Discs))
		fmt.Println("Amount spheres:   ", len(scene.Spheres))
		fmt.Println()

		fmt.Println("Initialize scene...")
		initializeScene(&scene)

		renderedPixelData := make([]scn.Color, animation.Width*animation.Height)

		render(&scene, animation.Width, animation.Height, renderedPixelData)

		animationDirectory := filepath.Join(".", "rendered", animation.AnimationName)
		animationFrameFilename := filepath.Join(animationDirectory, frame.Filename+".png")
		os.MkdirAll(animationDirectory, os.ModePerm)
		writeImage(animationFrameFilename, animation.Width, animation.Height, renderedPixelData)

		if animation.WriteRawImageFile {
			animationFrameRawFilename := filepath.Join(animationDirectory, frame.Filename+".praw")
			writeRawImage(animationFrameRawFilename, animation.Width, animation.Height, renderedPixelData)
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
	discs := scene.Discs

	for _, disc := range discs {
		projection := disc.Material.Projection
		if projection != nil {
			projection.InitializeProjection()
		}
	}

	spheres := scene.Spheres

	for _, sphere := range spheres {
		projection := sphere.Material.Projection
		if projection != nil {
			projection.InitializeProjection()
		}
	}
}

func deInitializeScene(scene *scn.Scene) {
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

func render(scene *scn.Scene, width int, height int, pixeldata []scn.Color) {
	var wg sync.WaitGroup

	amountSamples := scene.Camera.Samples

	progressbar := progressbar2.NewOptions(width*height+1, // Stay on 99% until all worker threads are done
		progressbar2.OptionFullWidth(),
		progressbar2.OptionClearOnFinish(),
		progressbar2.OptionSetPredictTime(true),
		progressbar2.OptionEnableColorCodes(true),
		progressbar2.OptionSetDescription("Render progress"),
	)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			wg.Add(1)
			go parallelPixelRendering(pixeldata, scene, width, height, x, y, amountSamples, &wg, progressbar)
		}
	}

	wg.Wait()

	progressbar.Add(1) // Final step to 100% in progress bar
	//progressbar.Clear()

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			pixelIndex := y*width + x
			pixeldata[pixelIndex].Divide(float64(amountSamples))
		}
	}
}

func parallelPixelRendering(pixeldata []scn.Color, scene *scn.Scene, width int, height int, x int, y int, amountSamples int, wg *sync.WaitGroup, progressbar *progressbar2.ProgressBar) {
	defer wg.Done()
	progressbar.Add(1)

	for sampleIndex := 0; sampleIndex < amountSamples; sampleIndex++ {
		cameraRay := CreateCameraRay(x, y, width, height, scene.Camera, sampleIndex)
		color := tracePath(cameraRay, scene)
		pixeldata[y*width+x].Add(color)
	}
}

func tracePath(ray scn.Ray, scene *scn.Scene) scn.Color {
	shortestDistance := float64(math.MaxFloat64)
	traceColor := scn.Black

	for _, sphere := range scene.Spheres {
		intersectionPoint, intersection := SphereIntersection(ray, sphere)
		if intersection {
			distance := vec3.Distance(&ray.Origin, &intersectionPoint)
			if distance < shortestDistance {
				shortestDistance = distance

				sphereOrigin := sphere.Origin
				sphereNormalAtIntersection := intersectionPoint.Sub(&sphereOrigin)
				incomingRay := ray.Heading.Inverted()
				cosineIncomingRayAndNormal := vec3.Dot(sphereNormalAtIntersection, &incomingRay) / (sphereNormalAtIntersection.Length() * incomingRay.Length())

				material := sphere.Material

				projectionColor := scn.White
				if material.Projection != nil {
					projectionColor = material.Projection.GetUV(&intersectionPoint)
				}

				traceColor = scn.Color{
					R: material.Color.R * cosineIncomingRayAndNormal * projectionColor.R,
					G: material.Color.G * cosineIncomingRayAndNormal * projectionColor.G,
					B: material.Color.B * cosineIncomingRayAndNormal * projectionColor.B,
				}
			}
		}
	}

	for _, disc := range scene.Discs {
		intersectionPoint, intersection := DiscIntersection(ray, disc)
		if intersection {
			distance := vec3.Distance(&ray.Origin, &intersectionPoint)
			if distance < shortestDistance {
				shortestDistance = distance

				normalAtIntersection := disc.Normal
				incomingRay := ray.Heading.Inverted()
				cosineIncomingRayAndNormal := vec3.Dot(&normalAtIntersection, &incomingRay) / (normalAtIntersection.Length() * incomingRay.Length())

				material := disc.Material

				projectionColor := scn.White
				if material.Projection != nil {
					projectionColor = material.Projection.GetUV(&intersectionPoint)
				}

				traceColor = scn.Color{
					R: material.Color.R * cosineIncomingRayAndNormal * projectionColor.R,
					G: material.Color.G * cosineIncomingRayAndNormal * projectionColor.G,
					B: material.Color.B * cosineIncomingRayAndNormal * projectionColor.B,
				}
			}
		}
	}

	return traceColor
}
