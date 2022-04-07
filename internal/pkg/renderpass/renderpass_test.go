package renderpass

import (
	"fmt"
	"testing"
)

func Test_CreateRenderPasses(t *testing.T) {
	for size := 1; size < 100; size++ {
		renderPasses := CreateRenderPasses(size)

		setPositions := make(map[string]bool, size*size)
		for _, pass := range renderPasses.RenderPasses {
			positionKey := fmt.Sprintf("%dx%d", pass.Dx, pass.Dy)
			setPositions[positionKey] = true
		}

		amountExpectedUniquePositions := size * size
		amountActualUniquePositions := len(setPositions)

		/*
			if size <= 5 {
				fmt.Println("Size:", size, ", Unique positions:", setPositions)
			} else {
				fmt.Println("Size:", size, ", Amount unique positions:", len(setPositions))
			}
		*/

		if amountActualUniquePositions != amountExpectedUniquePositions {
			t.Errorf("Actual amount unique positions %d differ from expected amount %d for size %d", amountActualUniquePositions, amountExpectedUniquePositions, size)
		}
	}
}
