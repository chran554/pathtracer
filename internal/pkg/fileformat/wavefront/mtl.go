package wavefront

import (
	"fmt"
	"pathtracer/internal/pkg/util"
	"strings"
)

// parseMaterialFile parses materials from a mtl-file
// http://paulbourke.net/dataformats/mtl/
func parseMaterialFile(lines []string, fileDir string) (map[string]*Material, error) {
	materialMap := make(map[string]*Material)

	var currentMaterial *Material

	for lineIndex, line := range lines {
		lineNumber := lineIndex + 1
		line = strings.TrimSpace(line)

		commentIndex := strings.Index(line, "#")

		// Comment line
		if commentIndex == 0 {
			continue
		}

		// Remove trailing comment
		if commentIndex > -1 {
			line = strings.TrimSpace(line[:commentIndex])
		}

		// Empty line
		if len(line) == 0 {
			continue
		}

		tokens := parseTokens(line, ' ')

		var err error

		lineType := strings.TrimSpace(tokens[0])

		switch lineType {
		case "newmtl":
			materialName := strings.TrimSpace(strings.TrimPrefix(line, lineType))
			//fmt.Printf("New material at line %d: %s\n", lineNumber, line)
			currentMaterial = &Material{
				Name:  materialName,
				Illum: -1,
				D:     1.0, // Fully opaque
				Ns:    -1,  // Default value, no specular exponent value (i.e. "Phong" specular exponent) set
				Ni:    -1.0,
				Refl:  -1.0,
			}
			materialMap[materialName] = currentMaterial

		case "sharpness":
			f, err := parseFloat64(tokens[1])
			if err != nil {
				return nil, fmt.Errorf("error parsing sharpness value at line %d: %s", lineNumber, err)
			}
			currentMaterial.Sharpness = util.ClampFloat64(0.0, 1.0, (1000.0-f)/1000.0)

		case "Ns":
			f, err := parseFloat64(tokens[1])
			if err != nil {
				return nil, fmt.Errorf("error parsing specular exponent value at line %d: %s", lineNumber, err)
			}
			currentMaterial.Ns = f

		case "refl":
			f, err := parseFloat64(tokens[1])
			if err != nil {
				return nil, fmt.Errorf("error parsing refraction value at line %d: %s", lineNumber, err)
			}
			currentMaterial.Refl = f

		case "Ks":
			r, err := parseFloat64(tokens[1])
			if err != nil {
				return nil, fmt.Errorf("error parsing specular color r value at line %d: %s", lineNumber, err)
			}
			g, err := parseFloat64(tokens[2])
			if err != nil {
				return nil, fmt.Errorf("error parsing specular color g value at line %d: %s", lineNumber, err)
			}
			b, err := parseFloat64(tokens[3])
			if err != nil {
				return nil, fmt.Errorf("error parsing specular color b value at line %d: %s", lineNumber, err)
			}
			currentMaterial.Ks = &Color{r, g, b}

		case "Tf":
			r, err := parseFloat64(tokens[1])
			if err != nil {
				return nil, fmt.Errorf("error parsing transmission r value at line %d: %s", lineNumber, err)
			}
			g, err := parseFloat64(tokens[2])
			if err != nil {
				return nil, fmt.Errorf("error parsing transmission g value at line %d: %s", lineNumber, err)
			}
			b, err := parseFloat64(tokens[3])
			if err != nil {
				return nil, fmt.Errorf("error parsing transmission b value at line %d: %s", lineNumber, err)
			}

			currentMaterial.Tf = &Color{r, g, b}

		case "Ke":
			r, err := parseFloat64(tokens[1])
			if err != nil {
				return nil, fmt.Errorf("error parsing emission r value at line %d: %s", lineNumber, err)
			}
			g, err := parseFloat64(tokens[2])
			if err != nil {
				return nil, fmt.Errorf("error parsing emission g value at line %d: %s", lineNumber, err)
			}
			b, err := parseFloat64(tokens[3])
			if err != nil {
				return nil, fmt.Errorf("error parsing emission b value at line %d: %s", lineNumber, err)
			}
			currentMaterial.Ke = &Color{r, g, b}

		case "Ni":
			f, err := parseFloat64(tokens[1])
			if err != nil {
				return nil, fmt.Errorf("error parsing index of refraction value at line %d: %s", lineNumber, err)
			}
			currentMaterial.Ni = f

		case "d":
			f, err := parseFloat64(tokens[1])
			if err != nil {
				return nil, fmt.Errorf("error parsing transparency value at line %d: %s", lineNumber, err)
			}
			currentMaterial.D = f

		case "illum":
			i, err := parseInt(tokens[1])
			if err != nil {
				return nil, fmt.Errorf("error parsing illumination model value at line %d: %s", lineNumber, err)
			}
			currentMaterial.Illum = i

		case "Pr":
			f, err := parseFloat64(tokens[1])
			if err != nil {
				return nil, fmt.Errorf("error parsing reflection value at line %d: %s", lineNumber, err)
			}
			currentMaterial.Pr = f

		case "Ka":
			r, err := parseFloat64(tokens[1])
			if err != nil {
				return nil, fmt.Errorf("error parsing ambient r value at line %d: %s", lineNumber, err)
			}
			g, err := parseFloat64(tokens[2])
			if err != nil {
				return nil, fmt.Errorf("error parsing ambient g value at line %d: %s", lineNumber, err)
			}
			b, err := parseFloat64(tokens[3])
			if err != nil {
				return nil, fmt.Errorf("error parsing ambient b value at line %d: %s", lineNumber, err)
			}

			currentMaterial.Ka = &Color{r, g, b}

		case "Kd":
			r, err := parseFloat64(tokens[1])
			if err != nil {
				return nil, fmt.Errorf("error parsing diffuse r value at line %d: %s", lineNumber, err)
			}
			g, err := parseFloat64(tokens[2])
			if err != nil {
				return nil, fmt.Errorf("error parsing diffuse g value at line %d: %s", lineNumber, err)
			}
			b, err := parseFloat64(tokens[3])
			if err != nil {
				return nil, fmt.Errorf("error parsing diffuse b value at line %d: %s", lineNumber, err)
			}

			currentMaterial.Kd = &Color{r, g, b}

		case "map_Kd":
			currentMaterial.MapKd = strings.Join(tokens[1:], " ")
			currentMaterial.MapBasePath = fileDir

		default:
			err = fmt.Errorf("unknown/unexpected line type: '%s'", line)
		}

		if err != nil {
			return nil, fmt.Errorf("%d: %s", lineNumber, err)
		}
	}

	return materialMap, nil
}
