package main

import (
	"fmt"
	"math"
	"testing"

	"github.com/ungerik/go3d/float64/mat3"
	"github.com/ungerik/go3d/float64/vec3"
)

func Test_InverseMatrixTransform(t *testing.T) {
	sqrt2 := math.Sqrt(2)
	A := mat3.T{
		vec3.T{1, 0, -1},
		vec3.T{0, sqrt2, 0},
		vec3.T{1, 0, 1},
	}
	Ai, _ := A.Inverted()

	v := vec3.T{2, sqrt2, 0}

	vp := Ai.MulVec3(&v)

	fmt.Println("A: ", A)
	fmt.Println("Ai:", Ai)
	fmt.Println("vp:", vp)
}
