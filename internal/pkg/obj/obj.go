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

	WriteString(objWriter, fmt.Sprintf("# Original OBJ-file '%s' created at %s\n\n", objFile.Name(), time.Now().String()))
	WriteString(mtlWriter, fmt.Sprintf("# Original MTL-file '%s' created at %s\n\n", mtlFile.Name(), time.Now().String()))
	for _, textLine := range comment {
		if strings.TrimSpace(textLine) != "" {
			WriteString(objWriter, "# "+textLine+"\n")
			WriteString(mtlWriter, "# "+textLine+"\n")
		} else {
			WriteString(objWriter, "\n")
			WriteString(mtlWriter, "\n")
		}
	}
	WriteString(objWriter, "\n")
	WriteString(mtlWriter, "\n")

	vertexIndexHashMap := make(map[*vec3.T]int)
	vertexNormalHashMap := make(map[*vec3.T]int)
	normalHashMap := make(map[*vec3.T]int)

	extractVectors(facetStructure, vertexIndexHashMap, vertexNormalHashMap, normalHashMap)

	serializeVerticesToObjFile(objWriter, vertexIndexHashMap)
	WriteString(objWriter, "\n")
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
		WriteString(objWriter, fmt.Sprintf("v %f %f %f\n", vertex[0], vertex[1], -vertex[2]))
	}
}

func serializeVertexNormalsToObjFile(objWriter *bufio.Writer, vertexNormalToIndexMap map[*vec3.T]int) {
	indexToVertexNormalMap := make(map[int]*vec3.T, 0)
	for vertexNormal, index := range vertexNormalToIndexMap {
		indexToVertexNormalMap[index] = vertexNormal
	}

	indices := make([]int, 0, len(indexToVertexNormalMap))
	for k := range indexToVertexNormalMap {
		indices = append(indices, k)
	}
	sort.Ints(indices)

	for index := range indices {
		vertexNormal := indexToVertexNormalMap[index]
		// OBJ-files require right hand coordinate system (thus convert from left hand coordinate system by inverting z-axis)
		WriteString(objWriter, fmt.Sprintf("vn %f %f %f\n", vertexNormal[0], vertexNormal[1], -vertexNormal[2]))
		// objWriter.WriteString(fmt.Sprintf("vn %f %f %f       # %d\n", vertexNormal[0], vertexNormal[1], -vertexNormal[2], index))
	}
}

func extractVectors(facetStructure *scene.FacetStructure, vertexIndexHashMap map[*vec3.T]int, vertexNormalHashMap map[*vec3.T]int, normalHashMap map[*vec3.T]int) {

	for _, facet := range facetStructure.Facets {

		for _, vertex := range facet.Vertices {
			// Add vertex to vertexIndexMap
			if _, ok := vertexIndexHashMap[vertex]; !ok {
				vertexIndex := len(vertexIndexHashMap)
				vertexIndexHashMap[vertex] = vertexIndex
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
		// WriteString(objWriterfmt.Sprintf("# Object '%s'\n", facetStructure.Name))
		WriteString(objWriter, fmt.Sprintf("\no %s\n", normalizeName(facetStructure.Name)))
	}

	if facetStructure.SubstructureName != "" {
		//WriteString(objWriter, fmt.Sprintf("\n# Object sub structure '%s'\n", normalizeName(facetStructure.SubstructureName)))
		WriteString(objWriter, fmt.Sprintf("\ng %s\n", normalizeName(facetStructure.SubstructureName)))
	}

	if facetStructure.Material != nil {
		WriteString(objWriter, fmt.Sprintf("usemtl %s\n", normalizeName(facetStructure.Material.Name)))
		serializeMaterial(mtlWriter, facetStructure.Material)
	}

	if len(facetStructure.Facets) > 0 {
		WriteString(objWriter, "\n")
		for _, facet := range facetStructure.Facets {
			WriteString(objWriter, "f")
			for faceVertexIndex, facetVertex := range facet.Vertices {
				if vertexIndex, ok := vertexSet[facetVertex]; ok {
					WriteString(objWriter, fmt.Sprintf(" %d", vertexIndex+1))
				} else {
					fmt.Println("could not find index for facet vertex")
				}

				if faceVertexIndex < len(facet.VertexNormals) {
					vertexNormal := facet.VertexNormals[faceVertexIndex]
					if vertexNormalIndex, ok := vertexNormalSet[vertexNormal]; ok {
						WriteString(objWriter, fmt.Sprintf("//%d", vertexNormalIndex+1))
					}
				}
			}
			WriteString(objWriter, "\n")
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
	WriteString(mtlWriter, fmt.Sprintf("newmtl %s\n", normalizeName(material.Name)))

	WriteString(mtlWriter, fmt.Sprintf("illum 7                           # Transparency: Refraction on; Reflection: Fresnel on and Ray trace on\n"))
	WriteString(mtlWriter, fmt.Sprintf("Kd %1.5f %1.5f %1.5f        # diffuse color\n", material.Color.R, material.Color.G, material.Color.B))
	if material.Transparency > 0.0 {
		WriteString(mtlWriter, fmt.Sprintf("Tf %1.5f %1.5f %1.5f        # transparency\n", material.Transparency, material.Transparency, material.Transparency))
	}

	if material.Glossiness > 0.0 {
		WriteString(mtlWriter, fmt.Sprintf("Ks %1.5f %1.5f %1.5f        # glossiness\n", material.Glossiness, material.Glossiness, material.Glossiness))
		WriteString(mtlWriter, fmt.Sprintf("sharpness %d                    # roughness (inverted)\n", int(math.Round((1.0-float64(material.Roughness))*1000.0))))
	}

	if material.RefractionIndex > 0.0 {
		WriteString(mtlWriter, fmt.Sprintf("Ni %1.5f                        # refraction index (for transparency)\n", material.RefractionIndex))
	}

	WriteString(mtlWriter, "\n")
}

func WriteString(w *bufio.Writer, s string) {
	w.WriteString(s)
}

func normalizeName(name string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(name), " ", "_"), ".", "_"), "#", "_")
}

func parseLines(lines []string, file *os.File) (*scene.FacetStructure, error) {
	type facetStructureKey struct {
		groupName    string
		materialName string
	}

	const defaultName = "default"

	var vertices []*vec3.T
	var normals []*vec3.T
	var textureVertices []*vec3.T

	materialMap := make(map[string]*scene.Material)

	groups := make(map[facetStructureKey][]*scene.Facet)
	smoothGroups := make(map[string][]*scene.Facet)

	var currentGroups []string
	var currentSmoothGroups []string
	currentMaterial := defaultName
	var name string

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

		command := strings.TrimSpace(tokens[0])
		var arguments []string
		if len(tokens) > 1 {
			arguments = tokens[1:]
		}

		switch command {
		case "v":
			vertex, err = parseVertex(arguments)
			vertices = append(vertices, vertex)
		case "vt":
			vertex, err = parseTextureVertex(arguments)
			textureVertices = append(textureVertices, vertex)
		case "vn":
			normal, err = parseNormal(arguments)
			normals = append(normals, normal)
		case "f":
			face, err = parseFace(arguments, vertices, normals, textureVertices)
			triangleFacets := face.SplitMultiPointFacet()
			if len(currentGroups) > 0 {
				for _, group := range currentGroups {
					key := facetStructureKey{groupName: group, materialName: currentMaterial}
					groups[key] = append(groups[key], triangleFacets...)
				}
			} else {
				key := facetStructureKey{groupName: defaultName, materialName: currentMaterial}
				groups[key] = append(groups[key], triangleFacets...)
			}

			if len(currentSmoothGroups) > 0 {
				for _, group := range currentSmoothGroups {
					smoothGroups[group] = append(smoothGroups[group], triangleFacets...)
				}
			}
		case "o":
			fmt.Printf("Object at line %d: %v\n", lineNumber, arguments)
			if len(arguments) > 0 {
				name = strings.Join(arguments, " ")
			}
		case "l":
			fmt.Printf("Line (not implemented yet) at line %d: %v\n", lineNumber, arguments) // TODO implement
		case "g":
			fmt.Printf("Group at line %d: %v\n", lineNumber, arguments)
			currentGroups = arguments
		case "s":
			// fmt.Printf("Smooth group at line %d: %v\n", lineNumber, arguments)
			if (len(arguments) == 1) && (arguments[0] == "off") {
				currentSmoothGroups = nil
			} else {
				currentSmoothGroups = arguments
			}
		case "mtllib":
			materialsFileName := strings.Join(tokens[1:], " ")
			materials, err := readMaterials(materialsFileName, file)
			materialMap = appendMaterialsMap(materialMap, materials)
			if err != nil {
				fmt.Println(err.Error())
			}
		case "usemtl":
			materialName := strings.Join(tokens[1:], " ")
			currentMaterial = materialName
		default:
			err = fmt.Errorf("unknown/unexpected line type: '%s'", line)
		}

		if err != nil {
			return nil, fmt.Errorf("%d: %s", lineNumber, err)
		}
	}

	// Build a flat facet structure with either one level or two levels.
	// The second level facet structures are per group and material.
	facetStructure := &scene.FacetStructure{Name: name}
	if len(groups) == 1 {
		for _, facets := range groups {
			facetStructure.Facets = facets
		}
	} else if len(groups) > 1 {
		for key, facets := range groups {
			facetStructureName := key.materialName
			if (key.groupName != "") && (key.groupName != defaultName) {
				fmt.Sprintf("%s::%s", key.groupName, key.materialName)
			}

			facetGroup := &scene.FacetStructure{Name: facetStructureName}
			facetGroup.Facets = facets
			facetGroup.Material = materialMap[key.materialName]
			facetStructure.FacetStructures = append(facetStructure.FacetStructures, facetGroup)
		}
	}

	// Smooth facet groups
	for _, facets := range smoothGroups {
		facetGroup := &scene.FacetStructure{Facets: facets} // Use a temporary facet structure
		facetGroup.UpdateVertexNormals(true)
	}

	return facetStructure, nil
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

// readMaterials reads materials from a mtl-file
// http://paulbourke.net/dataformats/mtl/
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
			newMaterial := scene.NewMaterial().N(materialName)
			materialMap[materialName] = newMaterial
			currentMaterial = newMaterial
		case "sharpness":
			// sharpness value
			//
			// Specifies the sharpness of the reflections from the local reflection
			// map.  If a material does not have a local reflection map defined in its
			// material definition, sharpness will apply to the global reflection map
			// defined in PreView.
			//
			// "value" can be a number from 0 to 1000.  The default is 60.  A high
			// value results in a clear reflection of objects in the reflection map.
		case "Ns":
			// Ns exponent
			//
			// Specifies the specular exponent for the current material.  This defines
			// the focus of the specular highlight.
			//
			// "exponent" is the value for the specular exponent.  A high exponent
			// results in a tight, concentrated highlight.  Ns values normally range
			// from 0 to 1000.

			// Blender software export "Roughness" material parameter as mtl-file parameter "Ns".
			currentMaterial.Roughness = parseFloat64(tokens[1])
		case "refl":
			// Blender software export "Metallic" material parameter as mtl-file parameter "refl".
			// This is NOT part of the mtl-file specification as "refl" is not supposed to be used for scalar values but
			// rather specify reflection maps.
			// However since that feature is very rarely used, and it is more likely to encounter a Blender exported file
			// we adhere to the format used by Blender.

			// Blender software export "Metallic" material parameter as mtl-file parameter "refl".
			currentMaterial.Glossiness = parseFloat64(tokens[1])
		case "Ks":
			// To specify the specular reflectivity of the current material
			// Ks r g b
			// Ks spectral file.rfl factor
			// Ks xyz x y z
			//
			// "Specularity / Glossiness" [0.0 .. 1.0]
			currentMaterial.Glossiness = parseFloat64(tokens[1])
		case "Tf":
			// To specify the transmission filter of the current material
			// Tf r g b
			// Tf spectral file.rfl factor
			// Tf xyz x y z
			//
			// Any light passing through the object is filtered by the transmission
			// filter, which only allows the specific colors to pass through.  For
			// example, Tf 0 1 0 allows all the green to pass through and filters out
			// all the red and blue.
		case "Ke":
			// Proprietary parameter for "emission" (not present in mtl-file specification)
			// "emission" [0.0 ..[
			currentMaterial.Emission = &color.Color{
				R: parseFloat32(tokens[1]),
				G: parseFloat32(tokens[2]),
				B: parseFloat32(tokens[3]),
			}
		case "Ni":
			// Ni optical_density
			//
			// Specifies the optical density for the surface.  This is also known as
			// index of refraction.
			currentMaterial.RefractionIndex = parseFloat64(tokens[1])
			currentMaterial.SolidObject = true
		case "d":
			// d factor
			//
			// Specifies the dissolve for the current material.
			//
			// "factor" is the amount this material dissolves into the background.  A
			// factor of 1.0 is fully opaque.  This is the default when a new material
			// is created.  A factor of 0.0 is fully dissolved (completely
			// transparent).
			//
			// Unlike a real transparent material, the dissolve does not depend upon
			// material thickness nor does it have any spectral character.  Dissolve
			// works on all illumination models.
			currentMaterial.Transparency = 1.0 - parseFloat64(tokens[1])
			currentMaterial.SolidObject = false
		case "illum":
			// illum illum_#
			//
			// The "illum" statement specifies the illumination model to use in the
			// material.  Illumination models are mathematical equations that represent
			// various material lighting and shading effects.
			//
			// Illumination    	Properties that are turned on in the
			// model           	Property Editor
			//
			// 0				Color on and Ambient off
			// 1				Color on and Ambient on
			// 2				Highlight on
			// 3				Reflection on and Ray trace on
			// 4				Transparency: Glass on, Reflection: Ray trace on
			// 5				Reflection: Fresnel on and Ray trace on
			// 6				Transparency: Refraction on, Reflection: Fresnel off and Ray trace on
			// 7				Transparency: Refraction on, Reflection: Fresnel on and Ray trace on
			// 8				Reflection on and Ray trace off
			// 9				Transparency: Glass on, Reflection: Ray trace off
			// 10				Casts shadows onto invisible surfaces
			//
		case "Pr":
			// Proprietary parameter for "roughness" (not present in mtl-file specification)
			// "Roughness" [0.0 .. 1.0]
			currentMaterial.Roughness = parseFloat64(tokens[1])
		case "Ka":
			// To specify the ambient reflectivity of the current material
			// Ka r g b
			// Ka spectral file.rfl factor
			// Ka xyz x y z
			//
			// "Ambient color" [[0.0 .. 1.0] [0.0 .. 1.0] [0.0 .. 1.0]]
		case "Kd":
			// To specify the diffuse reflectivity of the current material
			// Kd r g b
			// Kd spectral file.rfl factor
			// Kd xyz x y z
			//
			// "Diffuse color" [[0.0 .. 1.0] [0.0 .. 1.0] [0.0 .. 1.0]]
			c := &color.Color{
				R: parseFloat32(tokens[1]),
				G: parseFloat32(tokens[2]),
				B: parseFloat32(tokens[3]),
			}
			currentMaterial.Color = c
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
