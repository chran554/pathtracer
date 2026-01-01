package imagecache

import (
	"fmt"
	"io"
	"pathtracer/internal/pkg/floatimage"
	"strings"
	"sync"

	"github.com/ungerik/go3d/float64/vec3"
)

type Cache map[string]*floatimage.FloatImage

var globalImageCacheLock = &sync.Mutex{}
var globalImageCache = Cache{}

func GetCachedImage(filename string) *floatimage.FloatImage {
	globalImageCacheLock.Lock()
	defer globalImageCacheLock.Unlock()

	floatImage := globalImageCache[filename]

	if floatImage != nil {
		return floatImage
	}

	if strings.TrimSpace(filename) != "" {
		fmt.Println("Image cache loading file:", filename)
		floatImage = floatimage.LoadOrPanic(filename)
		fmt.Println("Image cache loading file:", filename, "... done", floatImage.String())
		globalImageCache[filename] = floatImage
	}

	return floatImage
}

func GetOrReadCachedImage(imageName string, r io.Reader, gammaDecode bool, normalMap bool) (*floatimage.FloatImage, error) {
	globalImageCacheLock.Lock()
	defer globalImageCacheLock.Unlock()

	img, exist := globalImageCache[imageName]

	if exist {
		return img, nil
	}

	if strings.TrimSpace(imageName) != "" {
		fmt.Println("Image cache reading file:", imageName)
		fimg, err := floatimage.Read(imageName, r)
		if err != nil {
			return nil, err
		}

		if gammaDecode {
			fimg.GammaDecode(floatimage.GammaDefault)
		}

		// Convert normal map r,g,b to vector x,y,z values
		if fimg != nil && normalMap {
			for y := 0; y < fimg.Height; y++ {
				for x := 0; x < fimg.Width; x++ {
					c := fimg.GetPixel(x, y)
					normal := &vec3.T{float64(c.R) - 0.5, float64(c.G) - 0.5, float64(c.B) - 0.5}
					normal.Normalize()
					c.R = float32(normal[0])
					c.G = float32(normal[1])
					c.B = float32(normal[2])
				}
			}
		}

		fmt.Println("Image cache reading file:", imageName, "... done", fimg.String())
		globalImageCache[imageName] = fimg
	}

	img = globalImageCache[imageName]

	return img, nil
}
