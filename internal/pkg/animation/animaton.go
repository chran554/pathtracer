package animation

import (
	"encoding/json"
	"fmt"
	"os"
	"pathtracer/internal/pkg/scene"
	"strconv"
)

func WriteAnimationToFile(animation *scene.Animation, indent bool) {
	var jsonData []byte
	var err error

	if indent {
		jsonData, err = json.MarshalIndent(animation, "", "  ")
	} else {
		jsonData, err = json.Marshal(animation)
	}
	if err != nil {
		message := fmt.Sprintf("Could not marshal animation \"%s\" to json: %s", animation.AnimationName, err.Error())
		panic(message)
	}

	filename := "scene/" + animation.AnimationName + ".animation.json"
	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		panic(fmt.Sprintf("Could not write animation file: %s", filename))
	} else {
		fileSize, err := getFileSize(filename)
		if err != nil {
			panic(fmt.Sprintf("Written animation file seem to be broken: %s", filename))
		}
		fmt.Println("Wrote animation file \"" + filename + "\" of size " + ByteCountIEC(fileSize) + " (" + strconv.FormatInt(fileSize, 10) + " bytes)")
	}
}

func getFileSize(filename string) (size int64, err error) {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return -1, err
	}
	return fileInfo.Size(), nil
}

func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
