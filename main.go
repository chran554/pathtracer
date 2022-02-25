package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"sync"
	"time"

	progressbar2 "github.com/schollz/progressbar/v3"
	"github.com/ungerik/go3d/vec3"
)

func main() {
	startTimestamp := time.Now()

	sceneFilename := "scene/three_balls.scene.json"
	imageFilename := "rendered/rendered.png"
	const width = 800 * 1
	const height = 600 * 1

	var sceneJson, err = os.ReadFile(sceneFilename)
	if err != nil {
		panic(err)
	}

	scene := Scene{}
	err = json.Unmarshal(sceneJson, &scene)
	if err != nil {
		panic(err)
	}

	// fmt.Printf("%+v\n", scene)
	fmt.Println("-----------------------------------------------")
	fmt.Println("Scene file:     ", sceneFilename)
	fmt.Println("Image file:     ", imageFilename)
	fmt.Println("Image size:     ", strconv.Itoa(width)+"x"+strconv.Itoa(height))
	fmt.Println("Amount samples: ", scene.Camera.Samples)
	fmt.Println("Amount discs:   ", len(scene.Discs))
	fmt.Println("Amount spheres: ", len(scene.Spheres))
	fmt.Println("-----------------------------------------------")

	pixeldata := make([]Color, width*height)

	render(&scene, width, height, pixeldata)

	writeImage(imageFilename, width, height, pixeldata)

	fmt.Println("Total execution time:", time.Since(startTimestamp))
}

func (c *Color) add(color Color) {
	c.R += color.R
	c.G += color.G
	c.B += color.B
}

func (c *Color) divide(divider float32) {
	c.R /= divider
	c.G /= divider
	c.B /= divider
}

func render(scene *Scene, width int, height int, pixeldata []Color) {
	var wg sync.WaitGroup

	amountSamples := scene.Camera.Samples

	progressbar := progressbar2.NewOptions(height,
		progressbar2.OptionFullWidth(),
		progressbar2.OptionClearOnFinish(),
		progressbar2.OptionSetPredictTime(true),
		progressbar2.OptionEnableColorCodes(true),
		progressbar2.OptionSetDescription("Render progress"),
	)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			wg.Add(1)
			go parallelPixelRendering(pixeldata, scene, width, height, x, y, amountSamples, &wg)
		}
		if err := progressbar.Set(y + 1); err != nil {
			fmt.Println(err)
		}
	}

	wg.Wait()

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			pixelIndex := y*width + x
			pixeldata[pixelIndex].divide(float32(amountSamples))
		}
	}
}

func parallelPixelRendering(pixeldata []Color, scene *Scene, width int, height int, x int, y int, amountSamples int, wg *sync.WaitGroup) {
	defer wg.Done()

	for sampleIndex := 0; sampleIndex < amountSamples; sampleIndex++ {
		cameraRay := createCameraRay(x, y, width, height, scene.Camera, sampleIndex)
		color := tracePath(cameraRay, scene)
		pixeldata[y*width+x].add(color)
	}
}

func tracePath(ray ray, scene *Scene) Color {
	shortestDistance := float32(math.MaxFloat32)
	traceColor := black

	for _, sphere := range scene.Spheres {
		intersectionPoint, intersection := sphereIntersection(ray, sphere)
		if intersection {
			distance := vec3.Distance(&ray.origin, &intersectionPoint)
			if distance < shortestDistance {
				shortestDistance = distance

				sphereOrigin := sphere.Origin
				sphereNormalAtIntersection := intersectionPoint.Sub(&sphereOrigin)
				incomingRay := ray.heading.Inverted()
				cosineIncomingRayAndSphereNormal := vec3.Dot(sphereNormalAtIntersection, &incomingRay) / (sphereNormalAtIntersection.Length() * incomingRay.Length())

				traceColor = Color{
					R: sphere.Material.Color.R * cosineIncomingRayAndSphereNormal,
					G: sphere.Material.Color.G * cosineIncomingRayAndSphereNormal,
					B: sphere.Material.Color.B * cosineIncomingRayAndSphereNormal,
				}
			}
		}
	}

	for _, disc := range scene.Discs {
		intersectionPoint, intersection := discIntersection(ray, disc)
		if intersection {
			distance := vec3.Distance(&ray.origin, &intersectionPoint)
			if distance < shortestDistance {
				shortestDistance = distance

				normalAtIntersection := disc.Normal
				incomingRay := ray.heading.Inverted()
				cosineIncomingRayAndNormal := vec3.Dot(&normalAtIntersection, &incomingRay) / (normalAtIntersection.Length() * incomingRay.Length())

				traceColor = Color{
					R: disc.Material.Color.R * cosineIncomingRayAndNormal,
					G: disc.Material.Color.G * cosineIncomingRayAndNormal,
					B: disc.Material.Color.B * cosineIncomingRayAndNormal,
				}
			}
		}
	}

	return traceColor
}
