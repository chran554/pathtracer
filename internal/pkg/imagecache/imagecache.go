package imagecache

import (
	"fmt"
	"io"
	"pathtracer/internal/pkg/floatimage"
	"strings"
	"sync"
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

func GetOrReadCachedImage(imageName string, r io.Reader, gammaDecode bool, postProcessFn func(*floatimage.FloatImage)) (*floatimage.FloatImage, error) {
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
			fimg.GammaDecode()
		}

		if postProcessFn != nil {
			postProcessFn(fimg)
		}

		fmt.Println("Image cache reading file:", imageName, "... done", fimg.String())
		globalImageCache[imageName] = fimg
	}

	img = globalImageCache[imageName]

	return img, nil
}
