package obj

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_LoadSkull(t *testing.T) {
	setTestResourcesRoot()
	skull := loadSkull(1.0)
	require.NotNil(t, skull)

	assert.Equal(t, 248999, skull.GetAmountFacets())
}
