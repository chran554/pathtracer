package obj

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/scene"
	"strconv"
	"strings"

	"github.com/ungerik/go3d/float64/vec3"
)

// An Object is the toplevel loadable object
type Object struct {
	Name            string
	Vertices        []*vec3.T
	Normals         []*vec3.T
	TextureVertices []*vec3.T
	Facets          []*scene.Facet
	Custom          map[string][]interface{} // Custom types for custom
}

type Facet struct {
	Vertices        []*vec3.T
	TextureVertices []*vec3.T
	VertexNormals   []*vec3.T
}

func Read(file *os.File) (*scene.FacetStructure, error) {
	var facetStructure *scene.FacetStructure

	reader := bufio.NewReader(file)
	lines, err := readLines(reader)
	if err != nil {
		return nil, err
	}

	if facetStructure, err = parseLines(lines, file); err != nil {
		return nil, err
	}

	return facetStructure, nil
}

func parseLines(lines []string, file *os.File) (*scene.FacetStructure, error) {
	rootFacetStructure := &scene.FacetStructure{}
	currentFacetStructure := rootFacetStructure

	var vertices []*vec3.T
	var normals []*vec3.T
	var textureVertices []*vec3.T

	materialMap := make(map[string]*scene.Material)

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
		var vertex *vec3.T
		var normal *vec3.T
		var face *scene.Facet

		lineType := strings.TrimSpace(tokens[0])

		switch lineType {
		case "v":
			vertex, err = parseVertex(tokens[1:])
			vertices = append(vertices, vertex)
		case "vt":
			vertex, err = parseTextureVertex(tokens[1:])
			textureVertices = append(textureVertices, vertex)
		case "vn":
			normal, err = parseNormal(tokens[1:])
			normals = append(normals, normal)
		case "f":
			face, err = parseFace(tokens[1:], vertices, normals, textureVertices)
			triangleFacets := splitMultiPointFacetToTriangles(face)
			currentFacetStructure.Facets = append(currentFacetStructure.Facets, triangleFacets...)
		case "o":
			fmt.Printf("Object at line %d: %s\n", lineNumber, line) // TODO implement
		case "l":
			fmt.Printf("Line at line %d: %s\n", lineNumber, line) // TODO implement
		case "g":
			fmt.Printf("Group at line %d: %s\n", lineNumber, line)
			newFacetStructure := &scene.FacetStructure{}
			newFacetStructure.Name = strings.Join(tokens[1:], " ") // Group name
			rootFacetStructure.FacetStructures = append(rootFacetStructure.FacetStructures, newFacetStructure)
			currentFacetStructure = newFacetStructure
		case "s":
			// TODO implement
		case "mtllib":
			//fmt.Printf("Mtllib at line %d: %s\n", lineNumber, line)
			materialsFileName := strings.Join(tokens[1:], " ")
			materials, err := readMaterials(materialsFileName, file)
			materialMap = appendMaterialsMap(materialMap, materials)
			if err != nil {
				fmt.Println(err.Error())
			}
		case "usemtl":
			// fmt.Printf("Usemtl at line %d: %s\n", lineNumber, line)
			materialName := strings.Join(tokens[1:], " ")
			newFacetStructure := &scene.FacetStructure{}
			newFacetStructure.Name = materialName // Material name
			newFacetStructure.Material = materialMap[materialName]
			// TODO set material to material from mtl-file
			rootFacetStructure.FacetStructures = append(rootFacetStructure.FacetStructures, newFacetStructure)
			currentFacetStructure = newFacetStructure
		default:
			err = fmt.Errorf("unknown/unexpected line type: '%s'", line)
		}

		if err != nil {
			return nil, fmt.Errorf("%d: %s", lineNumber, err)
		}
	}

	return rootFacetStructure, nil
}

func parseTokens(line string, delimiter rune) []string {
	f := func(c rune) bool {
		return c == delimiter
	}
	return strings.FieldsFunc(line, f)
}

func appendMaterialsMap(materialMap1 map[string]*scene.Material, materialMap2 map[string]*scene.Material) map[string]*scene.Material {
	resultMap := make(map[string]*scene.Material, len(materialMap1)+len(materialMap2))

	for materialName, material := range materialMap1 {
		resultMap[materialName] = material
	}
	for materialName, material := range materialMap2 {
		resultMap[materialName] = material
	}

	return resultMap
}

func readMaterials(materialFilename string, objectFile *os.File) (map[string]*scene.Material, error) {
	materialMap := make(map[string]*scene.Material)

	objectFileDirectory := filepath.Dir(objectFile.Name())
	materialFile := filepath.Join(objectFileDirectory, materialFilename)
	f, err := os.Open(materialFile)
	if err != nil {
		fmt.Printf("ouupps, something went wrong opening material file: '%s'\n%s\n", materialFile, err.Error())
	}

	reader := bufio.NewReader(f)
	lines, err := readLines(reader)
	if err != nil {
		return nil, err
	}

	var currentMaterial *scene.Material

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
			materialName := strings.Join(tokens[1:], " ")
			//fmt.Printf("New material at line %d: %s\n", lineNumber, line)
			newMaterial := &scene.Material{
				Color:           color.White,
				Emission:        nil,
				Glossiness:      0.0,
				Roughness:       0.0,
				Projection:      nil,
				RefractionIndex: 1.0,
				Transparency:    0.0,
				RayTerminator:   false,
			}

			materialMap[materialName] = newMaterial
			currentMaterial = newMaterial
		case "Ns":
		case "Ks":
			// "Specularity / Glossiness" [0.0 .. 1.0]
			currentMaterial.Glossiness = parseFloat32(tokens[1])
		case "Ke":
		case "Ni":
		case "d":
		case "illum":
		case "Pr":
			// "Roughness" [0.0 .. 1.0]
			currentMaterial.Roughness = parseFloat32(tokens[1])
		case "Ka":
			// "Ambient color" [[0.0 .. 1.0] [0.0 .. 1.0] [0.0 .. 1.0]]
		case "Kd":
			// "Diffuse color" [[0.0 .. 1.0] [0.0 .. 1.0] [0.0 .. 1.0]]
			color := color.Color{
				R: parseFloat32(tokens[1]),
				G: parseFloat32(tokens[2]),
				B: parseFloat32(tokens[3]),
			}
			currentMaterial.Color = color
		default:
			err = fmt.Errorf("unknown/unexpected line type: '%s'", line)
		}

		if err != nil {
			return nil, fmt.Errorf("%d: %s", lineNumber, err)
		}

	}

	// for i, line := range lines {
	// 	fmt.Printf("%d: %s\n", i, line)
	// }

	return materialMap, nil
}

func parseFloat32(value string) float32 {
	float, err := strconv.ParseFloat(value, 32)
	if err != nil {
		fmt.Printf("could not parse expected float value \"%s\".\n", value)
		return 1.0
	}

	return float32(float)
}

func parseFloat64(value string) float64 {
	float, err := strconv.ParseFloat(value, 64)
	if err != nil {
		fmt.Printf("could not parse expected float value \"%s\".\n", value)
		return 1.0
	}

	return float
}

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(r io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// splitMultiPointFacetToTriangles maps a multipoint (> 3 points) face into a list of triangles.
// The supplied face must have at least 3 points and be a convex face.
func splitMultiPointFacetToTriangles(facet *scene.Facet) (facets []*scene.Facet) {
	if len(facet.Vertices) > 3 {

		// Add first triangle of facet
		var textureVertices []*vec3.T
		var vertexNormals []*vec3.T

		if len(facet.TextureVertices) >= 3 {
			textureVertices = []*vec3.T{facet.TextureVertices[0], facet.TextureVertices[1], facet.TextureVertices[2]}
		}

		if len(facet.VertexNormals) >= 3 {
			vertexNormals = []*vec3.T{facet.VertexNormals[0], facet.VertexNormals[1], facet.VertexNormals[2]}
		}

		newFacet := scene.Facet{
			Vertices:        []*vec3.T{facet.Vertices[0], facet.Vertices[1], facet.Vertices[2]},
			TextureVertices: textureVertices,
			VertexNormals:   vertexNormals,
		}
		facets = append(facets, &newFacet)

		// Add consecutive triangles of facet
		for i := 3; i < len(facet.Vertices); i++ {
			newVertices := []*vec3.T{facet.Vertices[0], facet.Vertices[i-1], facet.Vertices[i]}

			var newTextureVertices []*vec3.T
			if len(facet.TextureVertices) > 0 {
				newTextureVertices = []*vec3.T{facet.TextureVertices[0], facet.TextureVertices[i-1], facet.TextureVertices[i]}
			}

			var newVertexNormals []*vec3.T
			if len(facet.VertexNormals) > 0 {
				newVertexNormals = []*vec3.T{facet.VertexNormals[0], facet.VertexNormals[i-1], facet.VertexNormals[i]}
			}

			newFace := scene.Facet{
				Vertices:        newVertices,
				TextureVertices: newTextureVertices,
				VertexNormals:   newVertexNormals,
			}
			facets = append(facets, &newFace)
		}
	} else {
		facets = append(facets, facet)
	}

	return facets
}

func parseFace(pointTokens []string, vertices []*vec3.T, normals []*vec3.T, textureVertices []*vec3.T) (*scene.Facet, error) {
	var face scene.Facet

	for _, pointToken := range pointTokens {
		vertexIndex, textureVertexIndex, vertexNormalIndex, err := parsePointIndexes(pointToken)
		if err != nil {
			return nil, err
		}

		if vertexIndex > 0 {
			face.Vertices = append(face.Vertices, vertices[vertexIndex-1])
		}

		if textureVertexIndex > 0 {
			face.TextureVertices = append(face.TextureVertices, textureVertices[textureVertexIndex-1])
		}

		if vertexNormalIndex > 0 {
			face.VertexNormals = append(face.VertexNormals, normals[vertexNormalIndex-1])
		}
	}

	return &face, nil
}

func parsePointIndexes(pointToken string) (vertexIndex int64, textureVertexIndex int64, vertexNormalIndex int64, err error) {
	vertexItems := strings.Split(pointToken, "/")

	vertexIndex, err = strconv.ParseInt(vertexItems[0], 10, 64)
	if err != nil {
		return 0, 0, 0, err
	}

	if len(vertexItems) > 1 && len(vertexItems[1]) != 0 {
		if textureVertexIndex, err = strconv.ParseInt(vertexItems[1], 10, 64); err != nil {
			return 0, 0, 0, err
		}
	}

	if len(vertexItems) > 2 && len(vertexItems[2]) != 0 {
		if vertexNormalIndex, err = strconv.ParseInt(vertexItems[2], 10, 64); err != nil {
			return 0, 0, 0, err
		}
	}

	return
}

func parseNormal(tokens []string) (*vec3.T, error) {
	var err error

	if len(tokens) != 3 {
		err = errors.New("item length for normal is incorrect")
		return nil, err
	}

	var normal vec3.T

	//TODO: check all, merge error types
	if normal[0], err = strconv.ParseFloat(tokens[0], 64); err != nil {
		err = errors.New("unable to parse X coordinate")
		return nil, err
	}
	if normal[1], err = strconv.ParseFloat(tokens[1], 64); err != nil {
		err = errors.New("unable to parse Y coordinate")
		return nil, err
	}
	if normal[2], err = strconv.ParseFloat(tokens[2], 64); err != nil {
		err = errors.New("unable to parse Z coordinate")
		return nil, err
	}

	return &normal, nil
}

func parseTextureVertex(tokens []string) (*vec3.T, error) {
	var err error

	amountTokens := len(tokens)
	if (amountTokens < 2) || (amountTokens > 3) {
		err = errors.New("item length for texture vertex is incorrect")
		return nil, err
	}

	var vertex vec3.T

	//TODO: merge errors together, check all fields
	if vertex[0], err = strconv.ParseFloat(tokens[0], 64); err != nil {
		err = errors.New("unable to parse U coordinate")
		return nil, err
	}
	if vertex[1], err = strconv.ParseFloat(tokens[1], 64); err != nil {
		err = errors.New("unable to parse V coordinate")
		return nil, err
	}
	if len(tokens) == 3 {
		if vertex[2], err = strconv.ParseFloat(tokens[2], 64); err != nil {
			err = errors.New("unable to parse W coordinate")
			return nil, err
		}
	}

	return &vertex, nil
}

func parseVertex(tokens []string) (*vec3.T, error) {
	var err error

	if len(tokens) != 3 {
		err = errors.New("item length for vertex is incorrect")
		return nil, err
	}

	var vertex vec3.T

	// TODO: verify each field, merge errors
	if vertex[0], err = strconv.ParseFloat(tokens[0], 64); err != nil {
		err = errors.New("unable to parse X coordinate")
		return nil, err
	}
	if vertex[1], err = strconv.ParseFloat(tokens[1], 64); err != nil {
		err = errors.New("unable to parse Y coordinate")
		return nil, err
	}
	if vertex[2], err = strconv.ParseFloat(tokens[2], 64); err != nil {
		err = errors.New("unable to parse Z coordinate")
		return nil, err
	}

	return &vertex, nil
}