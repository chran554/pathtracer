package scene

import (
	"fmt"
	"github.com/ungerik/go3d/float64/mat3"
	"github.com/ungerik/go3d/float64/vec3"
	"math"
	"testing"
)

func Test_matrixRotationY(t *testing.T) {

	m0 := mat3.T{}
	m0.AssignYRotation(0.0)

	m90 := mat3.T{}
	m90.AssignYRotation(math.Pi / 2)

	m180 := mat3.T{}
	m180.AssignYRotation(math.Pi)

	v := vec3.T{1.0, 0.0, 0.0}

	nv0 := m0.MulVec3(&v)
	nv90 := m90.MulVec3(&v)
	nv180 := m180.MulVec3(&v)

	fmt.Printf("  0: %+v\n", nv0)
	fmt.Printf(" 90: %+v\n", nv90)
	fmt.Printf("180: %+v\n", nv180)

	v10 := vec3.T{11, 0, 0}
	rotationOrigin := vec3.T{10, 0, 0}
	v10_2 := v10.Subed(&rotationOrigin)
	rotated := m90.MulVec3(&v10_2)
	rotated.Add(&rotationOrigin)

	fmt.Printf("Rotated: %+v\n", v10)
	fmt.Printf("Rotated: %+v\n", rotated)
}
