package image

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_getImageFromFilePath(t *testing.T) {
	image, err := getImageFromFilePath("../../../textures/dice/dice_1.png")
	assert.NoError(t, err)
	assert.NotNil(t, image)
}
