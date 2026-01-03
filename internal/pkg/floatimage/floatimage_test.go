package floatimage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getImageFromFilePath(t *testing.T) {
	image, err := Load("../../../textures/dice/dice_1.png")
	assert.NoError(t, err)
	assert.NotNil(t, image)
}
