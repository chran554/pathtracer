package image

import (
	"fmt"
	"strings"
)

var globalImageCache = ImageCache{}

type ImageCache map[string]*FloatImage

func GetCachedImage(filename string, gamma float64) *FloatImage {
	return globalImageCache.GetImage(filename, gamma)
}

func (cache ImageCache) GetImage(filename string, gamma float64) *FloatImage {
	image := cache[filename]

	if image != nil {
		return image
	}

	if strings.TrimSpace(filename) != "" {
		fmt.Println("Scene image cache loading file:", filename)
		image = LoadImageData(filename, gamma)
		fmt.Println("Scene image cache loading file:", filename, "... done")
		cache[filename] = image
	}

	return image
}
