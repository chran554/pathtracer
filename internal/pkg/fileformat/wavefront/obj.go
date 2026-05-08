package wavefront

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"maps"
	"math"
	"os"
	"path/filepath"
	"pathtracer/internal/pkg/scene"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ungerik/go3d/float64/vec3"
)

func Read(objFile *os.File) (*Wavefront, error) {
	reader := bufio.NewReader(objFile)
	lines, err := readLines(reader)
	if err != nil {
		return nil, err
	}

	materialFilenames, err := parseMaterialFilenames(lines, objFile)
	if err != nil {
		return nil, err
	}

	materialsMap, err := parseMaterialFiles(materialFilenames)
	if err != nil {
		return nil, err
	}

	wavefront, err := parseWavefront(lines, objFile, materialsMap)
	if err != nil {
		return nil, err
	}

	return wavefront, nil
}

func ReadOrPanic(objFilenamePath string) *Wavefront {
	objFile, err := os.Open(objFilenamePath)
	if err != nil {
		currentPath, err2 := filepath.Abs(".")
		if err2 != nil {
			currentPath = "[unknown]"
		}
		message := fmt.Sprintf("Could not open obj file: '%s'\n%s\nCurrent path: '%s'\n", objFilenamePath, err.Error(), currentPath)
		panic(message)
	}
	defer objFile.Close()

	obj, err := Read(objFile)
	if err != nil {
		message := fmt.Sprintf("Could not read obj file: '%s'\n%s\n", objFile.Name(), err.Error())
		panic(message)
	}

	return obj
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
		// OBJ-files require right-hand coordinate system (thus convert from left hand coordinate system by inverting z-axis)
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
		// OBJ-files require right-hand coordinate system (thus convert from left hand coordinate system by inverting z-axis)
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

type facetCollectionKey struct {
	objectName   string
	groupName    string
	materialName string
}

func (fck facetCollectionKey) isTopLevelCollection() bool {
	return (fck.objectName == "") && (fck.groupName == "")
}

func parseWavefront(lines []string, file *os.File, materialsMap map[string]*Material) (*Wavefront, error) {
	var defaultName = filepath.Base(file.Name())
	defaultName = strings.TrimSuffix(defaultName, filepath.Ext(defaultName))

	var wavefront = &Wavefront{
		Facets:          nil,
		Materials:       materialsMap,
		Groups:          make(map[string][]*Facet),
		Objects:         make(map[string][]*Facet),
		SmoothingGroups: make(map[int][]*Facet),
	}

	var vertices []*vec
	var normals []*vec
	var textureVertices []*vec

	var currentObjectName = ""
	var currentGroups = []string{"default"}
	var currentSmoothGroup = -1
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

		command := strings.TrimSpace(tokens[0])
		var arguments []string
		if len(tokens) > 1 {
			arguments = tokens[1:]
		}

		switch command {
		case "v":
			vertex, err := parseVertex(arguments)
			if err != nil {
				return nil, fmt.Errorf("error parsing vertex at line %d: %s", lineNumber, err)
			}
			vertices = append(vertices, vertex)
		case "vt":
			vertex, err := parseTextureVertex(arguments)
			if err != nil {
				return nil, fmt.Errorf("error parsing texture vertex at line %d: %s", lineNumber, err)
			}
			textureVertices = append(textureVertices, vertex)
		case "vn":
			normal, err := parseNormal(arguments)
			if err != nil {
				return nil, fmt.Errorf("error parsing normal at line %d: %s", lineNumber, err)
			}
			normals = append(normals, normal)
		case "f":
			face, err := parseFace(arguments, vertices, normals, textureVertices)
			if err != nil {
				return nil, fmt.Errorf("error parsing face at line %d: %s", lineNumber, err)
			}

			face.Groups = currentGroups
			face.SmoothingGroup = currentSmoothGroup
			face.Object = currentObjectName
			face.Material = currentMaterial

			for _, currentGroup := range currentGroups {
				if facets, exist := wavefront.Groups[currentGroup]; exist {
					facets = append(facets, face)
					wavefront.Groups[currentGroup] = facets
				} else {
					wavefront.Groups[currentGroup] = []*Facet{face}
				}
			}

			if currentSmoothGroup != 0 {
				if facets, exist := wavefront.SmoothingGroups[currentSmoothGroup]; exist {
					facets = append(facets, face)
					wavefront.SmoothingGroups[currentSmoothGroup] = facets
				} else {
					wavefront.SmoothingGroups[currentSmoothGroup] = []*Facet{face}
				}
			}

			if currentObjectName != "" {
				if facets, exist := wavefront.Objects[currentObjectName]; exist {
					facets = append(facets, face)
					wavefront.Objects[currentObjectName] = facets
				} else {
					wavefront.Objects[currentObjectName] = []*Facet{face}
				}
			}

			wavefront.Facets = append(wavefront.Facets, face)
		case "o":
			fmt.Printf("Object at line %d: %v\n", lineNumber, arguments)
			if len(arguments) > 0 {
				currentObjectName = strings.Join(arguments, "_")
			} else {
				currentObjectName = "" // Undefined state, this should never occur when there is no name for the object but assume and use "top level object"
			}
		case "l":
			fmt.Printf("Line (not implemented yet) at line %d: %v\n", lineNumber, arguments) // TODO implement
		case "g":
			fmt.Printf("Group at line %d: %v\n", lineNumber, arguments)
			currentGroups = arguments
		case "s":
			// fmt.Printf("Smooth group at line %d: %v\n", lineNumber, arguments)
			if len(arguments) > 1 {
				if (arguments[0] == "off") || (arguments[0] == "0") {
					currentSmoothGroup = -1
				} else {
					var err error
					currentSmoothGroup, err = parseInt(arguments[0])
					return nil, fmt.Errorf("could not parse smooth group at line %d: %v", lineNumber, err)
				}
			}
		case "mtllib":
			// do nothing, materials are parsed and loaded in separate pass
		case "usemtl":
			currentMaterialName := strings.Join(tokens[1:], " ")
			currentMaterialName = strings.TrimSpace(strings.ReplaceAll(currentMaterialName, "\"", ""))
			var exist bool
			if currentMaterial, exist = materialsMap[currentMaterialName]; !exist {
				fmt.Printf("Material '%s' not found in map %+v\n", currentMaterialName, materialsMap)
				return nil, fmt.Errorf("material '%s' not found", currentMaterialName)
			}

		default:
			return nil, fmt.Errorf("unknown/unexpected line type: '%s'", line)
		}
	}

	return wavefront, nil
}

func getOrCreateFileTopLevelNode(objectFacetStructures []*scene.FacetStructure, filename string) *scene.FacetStructure {
	if len(objectFacetStructures) == 1 {
		if objectFacetStructures[0].Name == "" {
			objectFacetStructures[0].Name = filename
		}
		return objectFacetStructures[0]
	}

	fileTopLevelNode := &scene.FacetStructure{Name: filename}

	// Set file top level node to object facet structure if it exists
	for i, facetStructure := range objectFacetStructures {
		if (facetStructure.Name == "") && (facetStructure.SubstructureName == "") && (facetStructure.Material == nil) {
			fileTopLevelNode = facetStructure
			objectFacetStructures = append(objectFacetStructures[:i], objectFacetStructures[i+1:]...)
			break
		}
	}

	// Add object facet structures as substructures to the file top level node
	fileTopLevelNode.FacetStructures = objectFacetStructures

	return fileTopLevelNode
}

// optimiseObjectStructures optimises loaded obj files in various ways.
// Optimises both with respect to make structure be more like obj-file structure and also removing
// unnecessary or superfluous substructures.
func optimiseObjectStructures(objectStructure *scene.FacetStructure) {
	// TODO implement structure optimization

	optimiseObjectStructureHierarchy(objectStructure)
	optimiseObjectStructureNames(objectStructure)
	optimiseSubstructureMaterials(objectStructure)
	optimiseSubstructureSlices(objectStructure)
}

// optimiseSubstructureMaterials moves material sub structure up one level in hierarchy if possible.
func optimiseSubstructureMaterials(structure *scene.FacetStructure) {
	// Recurse down the structure tree.
	// Optimise bottom up from structure leaves up towards the top level structure node
	for _, substructure := range structure.FacetStructures {
		optimiseSubstructureMaterials(substructure)
	}

	parent := structure

	singleChild := len(structure.FacetStructures) == 1
	if singleChild {
		child := structure.FacetStructures[0]

		emptyParent := (len(parent.Facets) == 0) && (parent.Material == nil)
		if emptyParent {
			parent.Material = child.Material
			parent.Facets = child.Facets
			parent.FacetStructures = child.FacetStructures

			if (parent.SubstructureName == "") && (child.SubstructureName != "") {
				parent.SubstructureName = child.SubstructureName
			}
		}
	}
}

// optimiseSubstructureSlices replaces empty substructure lists of length 0 with nil.
func optimiseSubstructureSlices(structure *scene.FacetStructure) {
	amountSubStructures := len(structure.FacetStructures)
	if amountSubStructures == 0 {
		structure.FacetStructures = nil
	} else {
		for _, subStructure := range structure.FacetStructures {
			optimiseSubstructureSlices(subStructure)
		}
	}
}

// optimiseObjectStructureNames removes names from structures set by structures higher in the hierarchy.
func optimiseObjectStructureNames(objectStructure *scene.FacetStructure) {
	for _, groupStructure := range objectStructure.FacetStructures {
		groupStructure.Name = ""

		for _, groupSubStructure := range groupStructure.FacetStructures {
			groupSubStructure.Name = ""
			groupSubStructure.SubstructureName = ""
		}
	}
}

// optimiseObjectStructureHierarchy removes superfluous intermediate structure nodes.
func optimiseObjectStructureHierarchy(objectStructure *scene.FacetStructure) {
	// Create a map from group name to a slice of structures included in the group.
	groupStructureMap := make(map[string][]*scene.FacetStructure, 0)
	for _, substructure := range objectStructure.FacetStructures {
		groupStructureMap[substructure.SubstructureName] = append(groupStructureMap[substructure.SubstructureName], substructure)
	}

	for groupName, groupStructures := range groupStructureMap {
		// If there is only one single structure for the group name then leave it

		if (len(groupStructures) > 1) && (groupName != "") {
			// Find the top level group structure (if it exists)
			groupTopLevelNode := &scene.FacetStructure{Name: objectStructure.Name, SubstructureName: groupName}
			for i, groupStructure := range groupStructures {
				if groupStructure.Material == nil {
					groupTopLevelNode = groupStructure
					groupStructures = append(groupStructures[:i], groupStructures[i+1:]...)
					break
				}
			}

			// Place other group structures under group top level node
			groupTopLevelNode.FacetStructures = groupStructures

			// Remove
			for i := 0; i < len(objectStructure.FacetStructures); {
				substructure := objectStructure.FacetStructures[i]
				if substructure.SubstructureName == groupName {
					objectStructure.FacetStructures = append(objectStructure.FacetStructures[:i], objectStructure.FacetStructures[i+1:]...)
				} else {
					i++
				}
			}

			objectStructure.FacetStructures = append(objectStructure.FacetStructures, groupTopLevelNode)
		}
	}
}

// getObjectStructures creates a list of all the object top level nodes.
// Each object structure can have an unoptimised set of sub structures which consists of materials and group structures.
func getObjectStructures(facetCollections map[facetCollectionKey][]*scene.Facet, materialMap map[string]*scene.Material) []*scene.FacetStructure {
	objectNames, objectNameStructuresMap := getObjects(facetCollections, materialMap)

	var objectStructures []*scene.FacetStructure

	for objectName := range objectNames {
		var objectStructure *scene.FacetStructure
		var found bool
		if objectStructure, found = removeObjectTopLevelStructure(objectName, objectNameStructuresMap); !found {
			// If no top level node was found for the object name
			// then create an empty top level node for the object.
			objectStructure = &scene.FacetStructure{Name: objectName}
		}

		objectStructure.FacetStructures = objectNameStructuresMap[objectName]

		objectStructures = append(objectStructures, objectStructure)
	}

	return objectStructures
}

// removeObjectTopLevelStructure finds, removes and returns the top level facet structure
// for an object given the object name.
// A top level facet structure for an object is the facet structure which
// has nu sub structure name (group name) and no material (material name).
// I.e. an object node with merely facets and has no group belonging nor any associated material.
func removeObjectTopLevelStructure(objectName string, objectFacetStructures map[string][]*scene.FacetStructure) (*scene.FacetStructure, bool) {
	facetStructures := objectFacetStructures[objectName]

	for i, facetStructure := range facetStructures {
		if (facetStructure.SubstructureName == "") && (facetStructure.Material == nil) {
			facetStructures = append(facetStructures[:i], facetStructures[i+1:]...) // remove top level structure
			objectFacetStructures[objectName] = facetStructures
			return facetStructure, true
		}
	}

	return nil, false
}

// getObjects creates a set of all object names and
// builds map from all object names to a list of all facet structures tagged with that object name.
func getObjects(facetCollections map[facetCollectionKey][]*scene.Facet, materialMap map[string]*scene.Material) (objectNames map[string]bool, objectFacetStructures map[string][]*scene.FacetStructure) {
	objectNames = make(map[string]bool)
	objectFacetStructures = make(map[string][]*scene.FacetStructure)

	for fck, facets := range facetCollections {

		if len(facets) > 0 {
			facetStructure := &scene.FacetStructure{
				Name:             fck.objectName,
				SubstructureName: fck.groupName,
				Material:         materialMap[fck.materialName],
				Facets:           facets,
			}

			objectNames[fck.objectName] = true
			objectFacetStructures[fck.objectName] = append(objectFacetStructures[fck.objectName], facetStructure)
		}
	}

	return objectNames, objectFacetStructures
}

func parseTokens(line string, delimiter rune) []string {
	f := func(c rune) bool {
		return c == delimiter
	}
	return strings.FieldsFunc(line, f)
}

func parseFloat64(value string) (float64, error) {
	float, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("could not parse expected float value \"%s\": %v", value, err)
	}

	return float, nil
}

func parseInt(value string) (int, error) {
	floatValue, err := parseFloat64(value)
	if err != nil {
		return 0, fmt.Errorf("could not parse expected int value \"%s\": %v", value, err)
	}

	intValue := int(floatValue)
	if math.Abs(float64(intValue)-floatValue) > 0.000000000001 {
		return 0, fmt.Errorf("could not parse expected int value \"%s\", delta diff for parsed int (%d) is to high: %v", value, intValue, err)
	}

	return intValue, nil
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

func parseFace(pointTokens []string, vertices []*vec, normals []*vec, textureVertices []*vec) (*Facet, error) {
	face := &Facet{}

	for _, pointToken := range pointTokens {
		vertexIndex, textureVertexIndex, vertexNormalIndex, err := parsePointIndexes(pointToken)
		if err != nil {
			return nil, err
		}

		if vertexIndex < 0 {
			vertexIndex = len(vertices) + vertexIndex + 1
		}
		if vertexIndex > 0 {
			face.Vertices = append(face.Vertices, vertices[vertexIndex-1])
		}

		if textureVertexIndex < 0 {
			textureVertexIndex = len(textureVertices) + textureVertexIndex + 1
		}
		if textureVertexIndex > 0 {
			face.VertexTextureCoordinates = append(face.VertexTextureCoordinates, textureVertices[textureVertexIndex-1])
		}

		if vertexNormalIndex < 0 {
			vertexNormalIndex = len(normals) + vertexNormalIndex + 1
		}
		if vertexNormalIndex > 0 {
			face.VertexNormals = append(face.VertexNormals, normals[vertexNormalIndex-1])
		}
	}

	return face, nil
}

func parsePointIndexes(pointToken string) (vertexIndex int, textureVertexIndex int, vertexNormalIndex int, err error) {
	vertexItems := strings.Split(pointToken, "/")

	vertexIndexValue, err := strconv.ParseInt(vertexItems[0], 10, 64)
	if err != nil {
		return 0, 0, 0, err
	}

	textureVertexIndexValue := int64(0)
	if len(vertexItems) > 1 && len(vertexItems[1]) != 0 {
		if textureVertexIndexValue, err = strconv.ParseInt(vertexItems[1], 10, 64); err != nil {
			return 0, 0, 0, err
		}
	}

	vertexNormalIndexValue := int64(0)
	if len(vertexItems) > 2 && len(vertexItems[2]) != 0 {
		if vertexNormalIndexValue, err = strconv.ParseInt(vertexItems[2], 10, 64); err != nil {
			return 0, 0, 0, err
		}
	}

	return int(vertexIndexValue), int(textureVertexIndexValue), int(vertexNormalIndexValue), nil
}

func parseNormal(tokens []string) (*vec, error) {
	var err error

	if len(tokens) != 3 {
		return nil, errors.New("item length for normal is incorrect")
	}

	var normal vec

	//TODO: check all, merge error types
	if normal[0], err = strconv.ParseFloat(tokens[0], 64); err != nil {
		return nil, errors.New("unable to parse X coordinate")
	}
	if normal[1], err = strconv.ParseFloat(tokens[1], 64); err != nil {
		return nil, errors.New("unable to parse Y coordinate")
	}
	if normal[2], err = strconv.ParseFloat(tokens[2], 64); err != nil {
		return nil, errors.New("unable to parse Z coordinate")
	}

	return &normal, nil
}

func parseTextureVertex(tokens []string) (*vec, error) {
	var err error

	amountTokens := len(tokens)
	if (amountTokens < 2) || (amountTokens > 3) {
		return nil, errors.New("item length for texture vertex is incorrect")
	}

	var vertex vec

	//TODO: merge errors together, check all fields
	if vertex[0], err = strconv.ParseFloat(tokens[0], 64); err != nil {
		return nil, errors.New("unable to parse U coordinate")
	}
	if vertex[1], err = strconv.ParseFloat(tokens[1], 64); err != nil {
		return nil, errors.New("unable to parse V coordinate")
	}
	if len(tokens) == 3 {
		if vertex[2], err = strconv.ParseFloat(tokens[2], 64); err != nil {
			return nil, errors.New("unable to parse W coordinate")
		}
	}

	return &vertex, nil
}

func parseVertex(tokens []string) (*vec, error) {
	var err error

	if len(tokens) != 3 {
		return nil, errors.New("item length for vertex is incorrect")
	}

	var vertex vec

	// TODO: verify each field, merge errors
	if vertex[0], err = strconv.ParseFloat(tokens[0], 64); err != nil {
		return nil, errors.New("unable to parse X coordinate")
	}
	if vertex[1], err = strconv.ParseFloat(tokens[1], 64); err != nil {
		return nil, errors.New("unable to parse Y coordinate")
	}
	if vertex[2], err = strconv.ParseFloat(tokens[2], 64); err != nil {
		return nil, errors.New("unable to parse Z coordinate")
	}

	return &vertex, nil
}

func parseMaterialFiles(materialFileNames []string) (map[string]*Material, error) {
	materialsMap := make(map[string]*Material)
	for _, materialFileName := range materialFileNames {
		materialFile, err := os.Open(materialFileName)
		if err != nil {
			return nil, err
		}
		defer materialFile.Close()

		reader := bufio.NewReader(materialFile)
		materialFileLines, err := readLines(reader)
		if err != nil {
			return nil, err
		}

		materialFileMap, err := parseMaterialFile(materialFileLines, filepath.Dir(materialFileName))
		if err != nil {
			return nil, err
		}

		maps.Insert(materialsMap, maps.All(materialFileMap))
	}
	return materialsMap, nil
}

func parseMaterialFilenames(lines []string, objFile *os.File) ([]string, error) {
	materialFileNames := make([]string, 0)

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

		command := strings.TrimSpace(tokens[0])

		switch command {
		case "mtllib":
			objFileRoot := filepath.Dir(objFile.Name())
			materialFileName := strings.TrimSpace(strings.TrimPrefix(line, command))
			materialFullFileName := filepath.Join(objFileRoot, materialFileName)

			if materialFileName == "" {
				return nil, fmt.Errorf("mtllib: no material file reference (line %d): '%s'", lineNumber, line)
			}

			if _, err := os.Stat(materialFullFileName); os.IsNotExist(err) {
				return nil, fmt.Errorf("mtllib: file '%s' does not exist (line %d): '%s'", materialFullFileName, lineNumber, line)
			}

			materialFileNames = append(materialFileNames, materialFullFileName)
		}
	}

	return materialFileNames, nil
}
