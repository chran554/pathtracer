package cie

import (
	"embed"
	"fmt"
	"strconv"
	"strings"
)

func parseIlluminantData(text string, name string) Illuminant {
	return Illuminant{TableData: parseTableData(text, name)}
}

func parseObserverData(text string, name string) Observer {
	xBar, yBar, zBar := parseObserverTableData(text)
	return Observer{Name: name, xBar: xBar, yBar: yBar, zBar: zBar}
}

func parseTableData(text string, name string) TableData {
	text = strings.ReplaceAll(text, "\r", "")
	lines := strings.Split(text, "\n")

	wavelength0, _ := strconv.Atoi(strings.Split(lines[0], ",")[0])
	wavelength1, _ := strconv.Atoi(strings.Split(lines[1], ",")[0])
	wavelengthEnd := wavelength0

	var data []float64

	for _, line := range lines {
		if !strings.HasPrefix(line, "#") && (strings.TrimSpace(line) != "") {
			wavelength, _ := strconv.Atoi(strings.Split(line, ",")[0])
			value, _ := strconv.ParseFloat(strings.Split(line, ",")[1], 64)

			data = append(data, value)
			wavelengthEnd = wavelength
		}
	}

	return TableData{
		Name:  name,
		Data:  data,
		Start: wavelength0,
		End:   wavelengthEnd,
		Step:  wavelength1 - wavelength0,
	}
}

func parseObserverTableData(text string) (xBar TableData, yBar TableData, zBar TableData) {
	text = strings.ReplaceAll(text, "\r", "")
	lines := strings.Split(text, "\n")

	wavelength0, _ := strconv.Atoi(strings.Split(lines[0], ",")[0])
	wavelength1, _ := strconv.Atoi(strings.Split(lines[1], ",")[0])
	wavelengthEnd := wavelength0

	var xBarData []float64
	var yBarData []float64
	var zBarData []float64

	for _, line := range lines {
		if !strings.HasPrefix(line, "#") && (strings.TrimSpace(line) != "") {
			wavelength, _ := strconv.Atoi(strings.Split(line, ",")[0])
			xBarValue, _ := strconv.ParseFloat(strings.Split(line, ",")[1], 64)
			yBarValue, _ := strconv.ParseFloat(strings.Split(line, ",")[2], 64)
			zBarValue, _ := strconv.ParseFloat(strings.Split(line, ",")[3], 64)

			xBarData = append(xBarData, xBarValue)
			yBarData = append(yBarData, yBarValue)
			zBarData = append(zBarData, zBarValue)

			wavelengthEnd = wavelength
		}
	}

	return TableData{Name: "xBar", Data: xBarData, Start: wavelength0, End: wavelengthEnd, Step: wavelength1 - wavelength0},
		TableData{Name: "yBar", Data: yBarData, Start: wavelength0, End: wavelengthEnd, Step: wavelength1 - wavelength0},
		TableData{Name: "zBar", Data: zBarData, Start: wavelength0, End: wavelengthEnd, Step: wavelength1 - wavelength0}
}

func readEmbeddedTextFile(fs embed.FS, fileName string) string {
	data, err := fs.ReadFile(fileName)
	if err != nil {
		panic(fmt.Sprintf("could not find data for CIE illuminant in embedded file \"%s\": %s", fileName, err))
	}

	return string(data)
}
