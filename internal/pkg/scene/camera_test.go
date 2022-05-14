package scene

import (
	"fmt"
	"math"
	"math/rand"
	"pathtracer/internal/pkg/color"
	img "pathtracer/internal/pkg/image"
	"strconv"
	"testing"
	"time"

	"github.com/ungerik/go3d/float64/mat3"
	"github.com/ungerik/go3d/float64/vec3"
)

func Test_CameraCoordinateSystem(t *testing.T) {
	t.Run("Camera coordinate system", func(t *testing.T) {
		camera := Camera{
			Heading: vec3.T{1, 0, 1},
			ViewUp:  vec3.T{0, 1, 0},
		}

		cameraSystem := camera.GetCameraCoordinateSystem()

		fmt.Println("Camera x", cameraSystem[0])
		fmt.Println("Camera y", cameraSystem[1])
		fmt.Println("Camera h", cameraSystem[2])

		//			if got := createCameraRay(tt.args.x, tt.args.y, tt.args.width, tt.args.height, tt.args.Camera); !reflect.DeepEqual(got, tt.want) {
		//				t.Errorf("createCameraRay() = %v, want %v", got, tt.want)
		//			}
	})
}

func Test_CoordinateSystemChangeForPoint(t *testing.T) {
	t.Run("coordinate system for point", func(t *testing.T) {
		vectorInCameraSpace := vec3.T{0, 0, 1}

		camera := Camera{
			Heading: vec3.T{1, 0, 1},
			ViewUp:  vec3.T{0, 1, 0},
		}

		cameraSystem := camera.GetCameraCoordinateSystem()

		vectorInSceneSystem := cameraSystem.MulVec3(&vectorInCameraSpace)

		fmt.Println("Camera system", cameraSystem)
		fmt.Println("vector in Camera system", vectorInCameraSpace)
		fmt.Println("vector in Scene system ", vectorInSceneSystem)

		//			if got := createCameraRay(tt.args.x, tt.args.y, tt.args.width, tt.args.height, tt.args.Camera); !reflect.DeepEqual(got, tt.want) {
		//				t.Errorf("createCameraRay() = %v, want %v", got, tt.want)
		//			}
	})
}

func Test_Struct(t *testing.T) {
	t.Run("struct", func(t *testing.T) {
		type Address struct {
			zip int8
		}

		type Person struct {
			name    string
			address Address
		}

		test := Person{
			name: "Gurkan",
			//			address: Address{},
		}

		fmt.Printf("%+v\n", test)
	})
}

func Test_InverseMatrixTransform(t *testing.T) {
	sqrt2 := math.Sqrt(2)
	A := mat3.T{
		vec3.T{1, 0, -1},
		vec3.T{0, sqrt2, 0},
		vec3.T{1, 0, 1},
	}
	Ai, _ := A.Inverted()

	v := vec3.T{2, sqrt2, 0}

	vp := Ai.MulVec3(&v)

	fmt.Println("A: ", A)
	fmt.Println("Ai:", Ai)
	fmt.Println("vp:", vp)
}

func Test_sunflower(t *testing.T) {
	t.Run("sunflower", func(t *testing.T) {
		width := 300
		height := 300
		amount := 4000
		randomize := true

		// ------------------------------------

		rand.Seed(time.Now().UnixMicro())

		halfWidth := float64(width / 2)
		halfHeight := float64(height / 2)

		image := img.NewFloatImage("sunflower", width, height)

		for i := 0; i < amount; i++ {
			x, y := sunflower(amount, 2.0, i+1, randomize)
			x2 := int(halfWidth * (1 + x))
			y2 := int(halfHeight * (1 - y))
			image.SetPixel(x2, y2, &color.Color{R: 1, G: 1, B: 1})
		}

		img.WriteImage("sunflower_["+strconv.Itoa(width)+"x"+strconv.Itoa(height)+"]x"+strconv.Itoa(amount)+"_random.png", image)

		//fmt.Printf("%+v\n", test)
	})
}
