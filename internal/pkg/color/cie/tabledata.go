package cie

import (
	"fmt"
	"math"
)

type TableData struct {
	Name  string
	Data  []float64
	Start int
	End   int
	Step  int
}

func (td *TableData) Copy() TableData {
	dataCopy := make([]float64, len(td.Data))
	copy(dataCopy, td.Data)
	return TableData{Name: td.Name, Data: dataCopy, Start: td.Start, End: td.End, Step: td.Step}
}

func (td *TableData) Value(x int) (value float64, err error) {
	if (x-td.Start)%td.Step > 0 {
		return 0.0, fmt.Errorf("can not get '%s' value, %d is not a valid data point", td.Name, x)
	}

	if !td.withinRange(x) {
		return 0.0, nil
	}

	return td.Data[(x-td.Start)/td.Step], nil
}

func (td *TableData) Integrate() float64 {
	var sum = 0.0
	for _, value := range td.Data {
		sum += value * float64(td.Step)
	}

	return sum
}

func (td *TableData) withinRange(x int) bool {
	return x >= td.Start && x <= td.End
}

func (td *TableData) withinRangeFloat(x float64) bool {
	return x >= float64(td.Start) && x <= float64(td.End)
}

func (td *TableData) InterpolatedValue(x float64) (value float64) {
	if !td.withinRangeFloat(x) {
		return 0.0
	}

	xi := int(math.Floor(x))
	xv := ((xi-td.Start)/td.Step)*td.Step + td.Start

	v, _ := td.Value(xv)
	v_1, _ := td.Value(maxInt(xv-td.Step, td.Start))
	v1, _ := td.Value(minInt(xv+td.Step, td.End))
	v2, _ := td.Value(minInt(xv+2*td.Step, td.End))

	u := (x - float64(xv)) / float64(td.Step)
	return cubicHermiteSplineInterpolateValue(v_1, v, v1, v2, u)
}

func (td *TableData) ClosestValue(x float64) (value float64, err error) {
	if !td.withinRangeFloat(x) {
		return 0.0, fmt.Errorf("can not get value, %f is outside valid range [%d,%d]", x, td.Start, td.End)
	}

	xi := int(math.Floor(x))
	xv := ((xi-td.Start)/td.Step)*td.Step + td.Start
	u := (x - float64(xv)) / float64(td.Step)

	if u < 0.5 {
		return td.Value(xv)
	} else {
		return td.Value(xv + td.Step)
	}
}

func (td *TableData) AddInterpolated(data TableData) error {
	f := func(a, b float64) float64 { return a + b }
	return td.combineInterpolated(data, f)
}

func (td *TableData) SubInterpolated(data TableData) error {
	f := func(a, b float64) float64 { return a - b }
	return td.combineInterpolated(data, f)
}

func (td *TableData) MulInterpolated(data TableData) error {
	f := func(a, b float64) float64 { return a * b }
	return td.combineInterpolated(data, f)
}

func (td *TableData) combineInterpolated(data TableData, f func(a, b float64) float64) error {
	minX := minInt(td.Start, data.Start)
	maxX := maxInt(td.End, data.End)

	amountSteps := (maxX-minX)/td.Step + 1

	newData := make([]float64, amountSteps) // TODO reuse existing slice if possible

	for i := 0; i < amountSteps; i++ {
		x := minX + (i * td.Step)

		value, err := td.Value(x)
		if err != nil {
			return err
		}

		interpolatedValue := data.InterpolatedValue(float64(x))

		newData[i] = f(value, interpolatedValue)
	}

	td.Data = newData
	td.Start = minX
	td.End = maxX

	return nil
}
