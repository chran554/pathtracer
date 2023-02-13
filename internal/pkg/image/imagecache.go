package image

import (
	"fmt"
	"strings"
	"sync"
)

type Cache map[string]*FloatImage

var globalImageCacheLock = &sync.Mutex{}
var globalImageCache = Cache{}

func GetCachedImage(filename string) *FloatImage {
	globalImageCacheLock.Lock()
	defer globalImageCacheLock.Unlock()

	image := globalImageCache[filename]

	if image != nil {
		return image
	}

	if strings.TrimSpace(filename) != "" {
		fmt.Println("Scene image cache loading file:", filename)
		image = LoadImageData(filename)
		fmt.Println("Scene image cache loading file:", filename, "... done")
		globalImageCache[filename] = image
	}

	return image
}
