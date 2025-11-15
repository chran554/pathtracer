package renderfile

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmihailenco/msgpack/v5"
)

func Test(t *testing.T) {
	//msgpack.RegisterExt(1, (*Vector)(nil))

	v1 := Vector{X: 1, Y: 2, Z: 3}
	v2 := Vector{X: 4, Y: 5, Z: 6}
	v3 := Vector{X: 7, Y: 8, Z: 9}

	vectors := []Vector{v1, v2, v3}

	v1data, err := msgpack.Marshal(&v1)
	assert.NoError(t, err)
	fmt.Println(v1data)

	data, err := msgpack.Marshal(&vectors)
	assert.NoError(t, err)

	var vectors2 []Vector
	err = msgpack.Unmarshal(data, &vectors2)
	assert.NoError(t, err)

	assert.Equal(t, vectors, vectors2)
}

func TestReadRenderFile(t *testing.T) {
	t.Skip("Skip test for now. Requires a valid render file to be present.")
	readRenderFile, err := ReadRenderFile("../../../scene/tessellated_sphere.render.zip")
	assert.NoError(t, err)

	assert.NotNil(t, readRenderFile)
}
