package scene

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"

	"github.com/ungerik/go3d/float64/mat3"
	"github.com/ungerik/go3d/float64/vec3"
)

var (
	Black = Color{R: 0, G: 0, B: 0}
	White = Color{R: 1, G: 1, B: 1}
)

type Color struct{ R, G, B float64 }

type Frame struct {
	Filename   string
	FrameIndex int
	Scene      Scene
}

type Animation struct {
	AnimationName string
	Frames        []Frame
	Width         int
	Height        int
}

type Scene struct {
	Camera  Camera
	Spheres []Sphere
	Discs   []Disc
}

type Material struct {
	Color      Color
	Emission   *Color                   `json:"Emission,omitempty"`
	Projection *ParallelImageProjection `json:"Projection,omitempty"`
}

type ParallelImageProjection struct {
	ImageFilename                   string  `json:"ImageFilename"`
	_imageData                      []Color `json:"-,omitempty"`
	_imageWidth                     int     `json:"-,omitempty"`
	_imageHeight                    int     `json:"-,omitempty"`
	_invertedCoordinateSystemMatrix mat3.T  `json:"-,omitempty"`
	Origin                          vec3.T  `json:"Origin"`
	U                               vec3.T  `json:"U"`
	V                               vec3.T  `json:"V"`
	RepeatU                         bool    `json:"RepeatU"`
	RepeatV                         bool    `json:"RepeatV"`
	FlipU                           bool    `json:"FlipU"`
	FlipV                           bool    `json:"FlipV"`
}

func NewParallelImageProjection2(textureFilename string, origin vec3.T, u vec3.T, v vec3.T) ParallelImageProjection {
	return NewParallelImageProjection(textureFilename, origin, u, v, true, true, false, false)
}

func NewParallelImageProjection(textureFilename string, origin vec3.T, u vec3.T, v vec3.T, repeatU bool, repeatV bool, flipU bool, flipV bool) ParallelImageProjection {
	return ParallelImageProjection{
		ImageFilename: textureFilename,
		Origin:        origin,
		U:             u,
		V:             v,
		RepeatU:       repeatU,
		RepeatV:       repeatV,
		FlipU:         flipU,
		FlipV:         flipV,
	}
}

func (imageProjection *ParallelImageProjection) GetUV(point *vec3.T) Color {
	imageProjection.InitializeProjection()

	translatedPoint := point.Sub(&imageProjection.Origin)

	pointInUV := imageProjection._invertedCoordinateSystemMatrix.MulVec3(translatedPoint)
	u := pointInUV[0]
	v := pointInUV[1]
	_, fracU := math.Modf(u)
	_, fracV := math.Modf(v)

	if !imageProjection.RepeatU && ((u >= 1.0) || (u < 0.0)) {
		return White
	}

	if !imageProjection.RepeatV && ((v >= 1.0) || (v < 0.0)) {
		return White
	}

	if fracU < 0.0 {
		fracU = fracU + 1.0
	}
	if fracV < 0.0 {
		fracV = fracV + 1.0
	}

	if imageProjection.FlipU && (int(math.Abs(math.Floor(u)))%2 == 1) {
		fracU = 1.0 - fracU
	}
	if imageProjection.FlipV && (int(math.Abs(math.Floor(v)))%2 == 1) {
		fracV = 1.0 - fracV
	}

	textureX := int(math.Abs(fracU) * float64(imageProjection._imageWidth))
	textureY := int(math.Abs(fracV) * float64(imageProjection._imageHeight))

	color := imageProjection._imageData[(imageProjection._imageHeight-1-textureY)*imageProjection._imageWidth+textureX]

	return color
}

func (imageProjection *ParallelImageProjection) ClearProjection() {
	imageProjection._imageData = nil
	imageProjection._invertedCoordinateSystemMatrix = mat3.Zero
	imageProjection._imageWidth = 0
	imageProjection._imageHeight = 0
}

func (imageProjection *ParallelImageProjection) InitializeProjection() {
	if len(imageProjection._imageData) == 0 {
		textureImage, err := getImageFromFilePath(imageProjection.ImageFilename)
		if err != nil {
			fmt.Println("Ouupps, no image file could be loaded for parallel image projection.", imageProjection.ImageFilename)
			os.Exit(1)
		}

		width := textureImage.Bounds().Max.X
		height := textureImage.Bounds().Max.Y

		imageProjection._imageWidth = width
		imageProjection._imageHeight = height

		imageProjection._imageData = make([]Color, width*height)
		colorNormalizationFactor := 1.0 / 0xffff
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				r, g, b, _ := textureImage.At(x, y).RGBA()
				imageProjection._imageData[y*width+x] = Color{
					R: float64(r) * colorNormalizationFactor,
					G: float64(g) * colorNormalizationFactor,
					B: float64(b) * colorNormalizationFactor,
				}
			}
		}
	}

	if imageProjection._invertedCoordinateSystemMatrix == mat3.Zero {
		W := vec3.Cross(&imageProjection.U, &imageProjection.V)
		imageProjection._invertedCoordinateSystemMatrix = mat3.T{imageProjection.U, imageProjection.V, W}
		imageProjection._invertedCoordinateSystemMatrix.Invert()
	}
}

func getImageFromFilePath(filePath string) (image.Image, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	image, _, err := image.Decode(f)

	// fmt.Println("Read image:", filePath)

	return image, err
}

type Sphere struct {
	Origin vec3.T
	Radius float64

	Material Material
}

type Camera struct {
	Origin            vec3.T
	Heading           vec3.T
	ViewUp            vec3.T
	ViewPlaneDistance float64
	_coordinateSystem mat3.T
	LensRadius        float64
	FocalDistance     float64
	Samples           int
	AntiAlias         bool
	Magnification     float64
}
type Line struct {
	Origin  vec3.T
	Heading vec3.T
}

type Ray Line

type Plane struct {
	Origin vec3.T
	Normal vec3.T

	Material Material
}

type Disc struct {
	Origin vec3.T
	Normal vec3.T
	Radius float64

	Material Material
}

type triangle struct {
	p1      vec3.T
	p2      vec3.T
	p3      vec3.T
	_normal vec3.T

	material Material
}

func (camera *Camera) GetCameraCoordinateSystem() mat3.T {
	if camera._coordinateSystem == (mat3.T{}) {
		heading := camera.Heading.Normalized()

		cameraX := vec3.Cross(&camera.ViewUp, &heading)
		cameraX.Normalize()
		cameraY := vec3.Cross(&heading, &cameraX)
		cameraY.Normalize()

		camera._coordinateSystem = mat3.T{cameraX, cameraY, heading}
	}
	return camera._coordinateSystem
}

func (c *Color) Add(color Color) {
	c.R += color.R
	c.G += color.G
	c.B += color.B
}

func (c *Color) Divide(divider float64) {
	c.R /= divider
	c.G /= divider
	c.B /= divider
}
