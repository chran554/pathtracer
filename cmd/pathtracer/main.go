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
	"github.com/ungerik/go3d/vec3"
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

		pixeldata := make([]scn.Color, animation.Width*animation.Height)

		render(&scene, animation.Width, animation.Height, pixeldata)

		animationDirectory := filepath.Join(".", "rendered", animation.AnimationName)
		animationFrameFilename := filepath.Join(animationDirectory, frame.Filename+".png")
		os.MkdirAll(animationDirectory, os.ModePerm)
		writeImage(animationFrameFilename, animation.Width, animation.Height, pixeldata)

		fmt.Println("Frame render time:", time.Since(frameStartTimestamp))

	}

	fmt.Println("Total execution time:", time.Since(startTimestamp))
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
			pixeldata[pixelIndex].Divide(float32(amountSamples))
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
	shortestDistance := float32(math.MaxFloat32)
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
				cosineIncomingRayAndSphereNormal := vec3.Dot(sphereNormalAtIntersection, &incomingRay) / (sphereNormalAtIntersection.Length() * incomingRay.Length())

				traceColor = scn.Color{
					R: sphere.Material.Color.R * cosineIncomingRayAndSphereNormal,
					G: sphere.Material.Color.G * cosineIncomingRayAndSphereNormal,
					B: sphere.Material.Color.B * cosineIncomingRayAndSphereNormal,
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

				traceColor = scn.Color{
					R: disc.Material.Color.R * cosineIncomingRayAndNormal,
					G: disc.Material.Color.G * cosineIncomingRayAndNormal,
					B: disc.Material.Color.B * cosineIncomingRayAndNormal,
				}
			}
		}
	}

	return traceColor
}
