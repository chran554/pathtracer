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

func GetOrReadCachedImage(cacheKey, imageName string, r io.Reader, decode func(*floatimage.FloatImage)) (*floatimage.FloatImage, error) {
	globalImageCacheLock.Lock()
	defer globalImageCacheLock.Unlock()

	if img, ok := globalImageCache[cacheKey]; ok {
		return img, nil
	}
	if strings.TrimSpace(imageName) == "" {
		return nil, nil
	}

	fmt.Println("Image cache reading file:", imageName)
	img, err := floatimage.Read(imageName, r)
	if err != nil {
		return nil, err
	}
	if decode != nil {
		decode(img)
	}

	fmt.Println("Image cache reading file:", imageName, "... done", img.String())

	globalImageCache[cacheKey] = img

	return img, nil
}
