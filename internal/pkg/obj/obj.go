package obj

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/scene"
	"sort"
	"strconv"
	"strings"
	"time"

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
	facetStructure.UpdateBounds()
	facetStructure.UpdateNormals()

	return facetStructure, nil
}

func WriteObjFile(objFile, mtlFile *os.File, facetStructure *scene.FacetStructure, comment []string) error {
	objWriter := bufio.NewWriter(objFile)
	mtlWriter := bufio.NewWriter(mtlFile)
	defer objWriter.Flush()
	defer mtlWriter.Flush()

	objWriter.WriteString(fmt.Sprintf("# Original OBJ-file '%s' created at %s\n\n", objFile.Name(), time.Now().String()))
	mtlWriter.WriteString(fmt.Sprintf("# Original MTL-file '%s' created at %s\n\n", mtlFile.Name(), time.Now().String()))
	for _, textLine := range comment {
		if strings.TrimSpace(textLine) != "" {
			objWriter.WriteString("# " + textLine + "\n")
			mtlWriter.WriteString("# " + textLine + "\n")
		} else {
			objWriter.WriteString("\n")
			mtlWriter.WriteString("\n")
		}
	}
	objWriter.WriteString("\n")
	mtlWriter.WriteString("\n")

	vertexIndexHashMap := make(map[*vec3.T]int)
	vertexNormalHashMap := make(map[*vec3.T]int)
	normalHashMap := make(map[*vec3.T]int)

	extractVectors(facetStructure, vertexIndexHashMap, vertexNormalHashMap, normalHashMap)

	serializeVerticesToObjFile(objWriter, vertexIndexHashMap)
	serializeVertexNormalsToObjFile(objWriter, vertexNormalHashMap)

	if err := serializeToObjFile(objWriter, mtlWriter, vertexIndexHashMap, vertexNormalHashMap, normalHashMap, facetStructure); err != nil {
		return fmt.Errorf("could not write obj/mtl file: %w", err)
	}

	return nil
}

func serializeVerticesToObjFile(objWriter *bufio.Writer, vertices map[*vec3.T]int) {
	keys := make([]*vec3.T, 0, len(vertices))

	for key := range vertices {
		keys = append(keys, key)
	}

	sort.SliceStable(keys, func(i, j int) bool {
		return vertices[keys[i]] < vertices[keys[j]]
	})

	for _, vertex := range keys {
		// OBJ-files require right hand coordinate system (thus convert from left hand coordinate system by inverting z-axis)
		objWriter.WriteString(fmt.Sprintf("v %f %f %f\n", vertex[0], vertex[1], -vertex[2]))
	}
}

func serializeVertexNormalsToObjFile(objWriter *bufio.Writer, vertexNormals map[*vec3.T]int) {
	for vertexNormal, _ := range vertexNormals {
		// OBJ-files require right hand coordinate system (thus convert from left hand coordinate system by inverting z-axis)
		objWriter.WriteString(fmt.Sprintf("vn %f %f %f\n", vertexNormal[0], vertexNormal[1], -vertexNormal[2]))
	}
}

func extractVectors(facetStructure *scene.FacetStructure, vertexIndexHashMap map[*vec3.T]int, vertexNormalHashMap map[*vec3.T]int, normalHashMap map[*vec3.T]int) {

	for _, facet := range facetStructure.Facets {

		for _, vertex := range facet.Vertices {
			if _, ok := vertexIndexHashMap[vertex]; !ok {
				vertexIndexHashMap[vertex] = len(vertexIndexHashMap)
			}
		}

		if *facet.Normal != vec3.Zero {
			if _, ok := normalHashMap[facet.Normal]; !ok {
				normalHashMap[facet.Normal] = len(normalHashMap)
			}
		}

		for _, normal := range facet.VertexNormals {
			if _, ok := vertexNormalHashMap[normal]; !ok {
				vertexNormalHashMap[normal] = len(vertexNormalHashMap)
			}
		}
	}

	for _, structure := range facetStructure.FacetStructures {
		extractVectors(structure, vertexIndexHashMap, vertexNormalHashMap, normalHashMap)
	}

}

func serializeToObjFile(objWriter *bufio.Writer, mtlWriter *bufio.Writer,
	vertexSet map[*vec3.T]int, vertexNormalSet map[*vec3.T]int, normalSet map[*vec3.T]int,
	facetStructure *scene.FacetStructure) error {

	if facetStructure.Name != "" {
		// objWriter.WriteString(fmt.Sprintf("# Object '%s'\n", facetStructure.Name))
		objWriter.WriteString(fmt.Sprintf("\no %s\n", normalizeName(facetStructure.Name)))
	}

	if facetStructure.SubstructureName != "" {
		//objWriter.WriteString(fmt.Sprintf("\n# Object sub structure '%s'\n", normalizeName(facetStructure.SubstructureName)))
		objWriter.WriteString(fmt.Sprintf("\ng %s\n", normalizeName(facetStructure.SubstructureName)))
	}

	if facetStructure.Material != nil {
		objWriter.WriteString(fmt.Sprintf("usemtl %s\n", normalizeName(facetStructure.Material.Name)))
		serializeMaterial(mtlWriter, facetStructure.Material)
	}

	if len(facetStructure.Facets) > 0 {
		objWriter.WriteString("\n")
		for _, facet := range facetStructure.Facets {
			objWriter.WriteString("f")
			for _, facetVertex := range facet.Vertices {
				if vertexIndex, ok := vertexSet[facetVertex]; ok {
					objWriter.WriteString(fmt.Sprintf(" %d", vertexIndex+1))
				} else {
					fmt.Println("could not find index for facet vertex")
				}
			}
			objWriter.WriteString("\n")
		}
	}

	for _, structure := range facetStructure.FacetStructures {
		if err := serializeToObjFile(objWriter, mtlWriter, vertexSet, vertexNormalSet, normalSet, structure); err != nil {
			return err
		}
	}

	return nil
}

func serializeMaterial(mtlWriter *bufio.Writer, material *scene.Material) {
	mtlWriter.WriteString(fmt.Sprintf("newmtl %s\n", normalizeName(material.Name)))

	mtlWriter.WriteString(fmt.Sprintf("illum 7                           # Transparency: Refraction on; Reflection: Fresnel on and Ray trace on\n"))
	mtlWriter.WriteString(fmt.Sprintf("Kd %1.5f %1.5f %1.5f        # diffuse color\n", material.Color.R, material.Color.G, material.Color.B))
	if material.Transparency > 0.0 {
		mtlWriter.WriteString(fmt.Sprintf("Tf %1.5f %1.5f %1.5f        # transparency\n", material.Transparency, material.Transparency, material.Transparency))
	}

	if material.Glossiness > 0.0 {
		mtlWriter.WriteString(fmt.Sprintf("Ks %1.5f %1.5f %1.5f        # glossiness\n", material.Glossiness, material.Glossiness, material.Glossiness))
		mtlWriter.WriteString(fmt.Sprintf("sharpness %d                    # roughness (inverted)\n", int(math.Round((1.0-float64(material.Roughness))*1000.0))))
	}

	if material.RefractionIndex > 0.0 {
		mtlWriter.WriteString(fmt.Sprintf("Ni %1.5f                        # refraction index (for transparency)\n", material.RefractionIndex))
	}

	mtlWriter.WriteString("\n")
}

func normalizeName(name string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(name), " ", "_"), ".", "_"), "#", "_")
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
			triangleFacets := face.SplitMultiPointFacet()
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
	defer f.Close()

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
