package floatimage

import (
	"fmt"
	"io"
	"strings"
	"sync"
)

type Cache map[string]*FloatImage

var globalImageCacheLock = &sync.Mutex{}
var globalImageCache = Cache{}

func GetCachedImage(filename string) *FloatImage {
	globalImageCacheLock.Lock()
	defer globalImageCacheLock.Unlock()

	floatImage := globalImageCache[filename]

	if floatImage != nil {
		return floatImage
	}

	if strings.TrimSpace(filename) != "" {
		fmt.Println("Image cache loading file:", filename)
		floatImage = Load(filename)
		fmt.Println("Image cache loading file:", filename, "... done", floatImage.String())
		globalImageCache[filename] = floatImage
	}

	return floatImage
}

func GetOrReadCachedImage(imageName string, r io.Reader) (*FloatImage, error) {
	globalImageCacheLock.Lock()
	defer globalImageCacheLock.Unlock()

	img, exist := globalImageCache[imageName]

	if exist {
		return img, nil
	}

	if strings.TrimSpace(imageName) != "" {
		fmt.Println("Image cache reading file:", imageName)
		floatImage, err := Read(imageName, r)
		if err != nil {
			return nil, err
		}
		fmt.Println("Image cache reading file:", imageName, "... done", floatImage.String())
		globalImageCache[imageName] = floatImage
	}

	img = globalImageCache[imageName]

	return img, nil
}
