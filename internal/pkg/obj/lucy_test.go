package obj

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_LoadLucy(t *testing.T) {
	setTestResourcesRoot()
	lucy := loadLucy(1.0)
	require.NotNil(t, lucy)

	assert.Equal(t, 28055742, lucy.GetAmountFacets())
}
