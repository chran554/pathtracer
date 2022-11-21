package ply

import (
	"bufio"
	"fmt"
	"github.com/ungerik/go3d/float64/vec3"
	"io"
	"os"
	"path/filepath"
	"pathtracer/internal/pkg/scene"
	"strconv"
	"strings"
)

// PropertyDefinition holds the definition from the ply file header of a property of an element.
// This property definition can hold both single property definitions and "reference list" definitions.
type PropertyDefinition struct {
	// Single property definition
	name     string
	dataType string
	index    int

	// Reference list definition
	listProperty             bool   // listProperty holds whether this property definition is a reference list definition or a single property definition
	listCountDataType        string // listCountDataType is data type for the amount of references in the list. Should always be some kind of integer type(?)
	listElementIndexDataType string // listElementIndexDataType is the data type for index references. Should always be some kind of integer type(?)
	listElementType          string // listElementType the type of element that the index references in the list refer to. (Like a facet element with list property refer to index of element type vertex (vertices).
}

type ElementDefinition struct {
	name       string
	count      int
	properties []*PropertyDefinition
}

type PropertyValue struct {
	name       string
	valueType  ValueType
	intValue   int
	floatValue float64
}

type ElementValue struct {
	name   string
	index  int
	values []*PropertyValue
}

type ValueType string

const (
	Unknown        ValueType = "?"
	Int                      = "int"
	Float                    = "float"
	IndexReference           = "indexReference"
)

// ReadPly
// https://web.archive.org/web/20161204152348/http://www.dcs.ed.ac.uk/teaching/cs4/www/graphics/Web/ply.html
// https://www.mathworks.com/help/vision/ug/the-ply-format.html
// http://paulbourke.net/dataformats/ply/
func ReadPly(file *os.File) (*scene.FacetStructure, error) {
	reader := bufio.NewReader(file)
	lines, err := readLines(reader)
	if err != nil {
		return nil, err
	}

	elementDefinitions, dataStartLineIndex, err := parsePlyHeaderSection(lines)
	if err != nil {
		return nil, err
	}

	var elementValues []*ElementValue
	if elementValues, err = parsePlyDataSection(elementDefinitions, dataStartLineIndex, lines); err != nil {
		return nil, fmt.Errorf("could not parse ply file '%s': %w", file.Name(), err)
	}

	var facetStructure *scene.FacetStructure
	if facetStructure, err = convertToFacetStructure(elementValues); err != nil {
		return nil, fmt.Errorf("could not convert ply file '%s' values to facet structure: %w", file.Name(), err)
	}

	facetStructure.UpdateBounds()
	facetStructure.UpdateNormals()

	facetStructure.Name = strings.TrimSuffix(strings.TrimSuffix(filepath.Base(file.Name()), ".ply"), ".PLY")

	return facetStructure, nil
}

func convertToFacetStructure(values []*ElementValue) (*scene.FacetStructure, error) {
	var vertices []*vec3.T
	var facets []*scene.Facet
	var err error

	if vertices, err = extractVertices(values); err != nil {
		return nil, fmt.Errorf("could not extract verticies from ply values: %w", err)
	}

	if facets, err = extractFacets(values, vertices); err != nil {
		return nil, fmt.Errorf("could not extract facets from ply values: %w", err)
	}

	return &scene.FacetStructure{Facets: facets}, nil
}

func extractFacets(values []*ElementValue, vertices []*vec3.T) ([]*scene.Facet, error) {
	var facets []*scene.Facet

	for _, value := range values {
		if (value.name == "facet") || (value.name == "face") {
			if value.index != len(facets) {
				return nil, fmt.Errorf("illegal facet value index (%d), it do not match slice index (%d)", value.index, len(vertices))
			}

			var facet scene.Facet
			for _, propertyValue := range value.values {
				if (propertyValue.valueType == IndexReference) && (propertyValue.name == "vertex") {
					facet.Vertices = append(facet.Vertices, vertices[propertyValue.intValue])
				} else {
					// If facet vertices is not a list of vertex indices in ply file, what is it then?
					return nil, fmt.Errorf("unknown ply format for face data (not a vertex index reference list): %+v", *value)
				}
			}
			facets = append(facets, &facet)
		}
	}

	return facets, nil
}

func extractVertices(values []*ElementValue) ([]*vec3.T, error) {
	var vertices []*vec3.T

	for _, value := range values {
		if value.name == "vertex" {
			if value.index != len(vertices) {
				return nil, fmt.Errorf("illegal vertex value index (%d), it do not match slice index (%d)", value.index, len(vertices))
			}

			vertex := vec3.T{}
			for _, propertyValue := range value.values {
				switch propertyValue.name {
				case "x":
					vertex[0] = getPropertyValue(propertyValue)
				case "y":
					vertex[1] = getPropertyValue(propertyValue)
				case "z":
					vertex[2] = getPropertyValue(propertyValue)
				}
			}
			vertices = append(vertices, &vertex)
		}
	}

	return vertices, nil
}

func getPropertyValue(propertyValue *PropertyValue) float64 {
	var value float64

	if propertyValue.valueType == Int {
		value = float64(propertyValue.intValue)
	} else if propertyValue.valueType == Float {
		value = propertyValue.floatValue
	}

	return value
}

func parsePlyDataSection(elementDefinitions []*ElementDefinition, dataStartIndex int, lines []string) (elementValues []*ElementValue, error error) {
	for lineIndex := dataStartIndex; lineIndex < len(lines); lineIndex++ {
		line := lines[lineIndex]
		dataLineIndex := lineIndex - dataStartIndex
		dataLineNumber := dataLineIndex + 1

		currentElementDefinition, elementValueIndex := elementDefinitionAtDataLine(elementDefinitions, dataLineNumber)

		elementValue := &ElementValue{name: currentElementDefinition.name, index: elementValueIndex}
		tokens := parseTokens(line, ' ')

		if currentElementDefinition.properties[0].listProperty {
			// Definition: property list <numerical-type> <numerical-type> <property-name>
			expectedAmountReferences, err := strconv.Atoi(tokens[0])
			if err != nil {
				return nil, fmt.Errorf("could not parse '%s' (from %s line '%s') to expected amount references of value reference list: %w", tokens[0], currentElementDefinition.name, line, err)
			}
			if expectedAmountReferences != (len(tokens) - 1) {
				return nil, fmt.Errorf("expected amount references %d did not match actual amount %d references (tokens %+v)", expectedAmountReferences, len(tokens)-1, tokens)
			}
			propertyDefinition := currentElementDefinition.properties[0]
			elementReferenceType := getElementReferenceType(propertyDefinition.listElementType)
			for _, token := range tokens[1:] {
				propertyValue := &PropertyValue{
					name:       strings.ToLower(elementReferenceType),
					valueType:  IndexReference,
					intValue:   getIntValue(IndexReference, token),
					floatValue: 0,
				}
				elementValue.values = append(elementValue.values, propertyValue)
			}
		} else {
			// Definition: property <data-type> <property-name-1>
			for tokenIndex, token := range tokens {
				propertyDefinition := currentElementDefinition.properties[tokenIndex]
				valueType := getValueType(propertyDefinition.dataType)
				propertyValue := &PropertyValue{
					name:       strings.ToLower(propertyDefinition.name),
					valueType:  valueType,
					intValue:   getIntValue(valueType, token),
					floatValue: getFloatValue(valueType, token),
				}
				elementValue.values = append(elementValue.values, propertyValue)
			}
		}

		// TODO handle odd-looking references...

		elementValues = append(elementValues, elementValue)
	}

	return
}

func getElementReferenceType(elementTypeReference string) string {
	s := elementTypeReference

	etr := strings.ToLower(elementTypeReference)
	lastIndex := strings.LastIndex(etr, "_")
	if (lastIndex != -1) && (etr[lastIndex:] == "_index" || etr[lastIndex:] == "_indices") {
		s = elementTypeReference[:lastIndex]
	}

	return s
}

func getIntValue(valueType ValueType, text string) int {
	if (valueType == Int) || (valueType == IndexReference) {
		value, err := strconv.Atoi(text)
		if err != nil {
			fmt.Printf("could not parse text '%s' to int value.\n", text)
			return 0
		}

		return value
	}
	return 0
}

func getFloatValue(valueType ValueType, text string) float64 {
	if valueType == Float {
		value, err := strconv.ParseFloat(text, 64)
		if err != nil {
			fmt.Printf("could not parse text '%s' to float value.\n", text)
			return 0
		}

		return value
	}
	return 0
}

/*
name	type	   number of bytes
----    ------     ---------------
char    character                1
uchar   unsigned character       1
short   short integer            2
ushort  unsigned short integer	 2
int     integer                  4
uint    unsigned integer         4
float   single-precision float   4
double  double-precision float   8
*/
func getValueType(dataType string) ValueType {
	switch dataType {
	case "char":
		fallthrough
	case "uchar":
		fallthrough
	case "short":
		fallthrough
	case "ushort":
		fallthrough
	case "int":
		fallthrough
	case "uint":
		return Int // Map data types to int in internal handling

	case "float":
		fallthrough
	case "double":
		return Float // Map data types to float (float64) in internal handling

	default:
		fmt.Printf("unknown data type: '%s'\n", dataType)
		return Unknown
	}
}

func elementDefinitionAtDataLine(elementSections []*ElementDefinition, lineNumber int) (currentElementSection *ElementDefinition, elementValueIndex int) {
	elementSectionLineStart := 1
	for _, elementSection := range elementSections {
		if (lineNumber >= elementSectionLineStart) && (lineNumber < (elementSectionLineStart + elementSection.count)) {
			currentElementSection = elementSection
			elementValueIndex = lineNumber - elementSectionLineStart
			break
		}

		elementSectionLineStart += elementSection.count
	}

	return currentElementSection, elementValueIndex
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

func parsePlyHeaderSection(lines []string) (elementDefinitions []*ElementDefinition, dataStartIndex int, error error) {
	if (len(lines) > 0) && !(lines[0] == "ply" || lines[0] == "Ply" || lines[0] == "PLY") {
		return nil, -1, fmt.Errorf("can notrecognise PLY file as it do not start with magic number 'ply'")
	}

	headerLines := true

	for lineIndex, line := range lines {
		if !headerLines {
			break
		}

		line = strings.TrimSpace(line)

		commentIndex := strings.Index(line, "comment")

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

		lineType := strings.TrimSpace(tokens[0])

		switch lineType {
		case "PLY":
			fallthrough
		case "Ply":
			fallthrough
		case "ply":
			// Nothing by intention, already handled "magic number" before parsing
		case "format":
			if (tokens[1] != "ascii") || (tokens[2] != "1.0") {
				error = fmt.Errorf("can not parse unknown ply file format '%s %s'", tokens[1], tokens[2])
			}
		case "element":
			// Definition: element <element-name> <number-in-file>
			count, _ := strconv.Atoi(tokens[2])
			elementSection := ElementDefinition{name: tokens[1], count: count, properties: nil}
			elementDefinitions = append(elementDefinitions, &elementSection)
		case "property":
			if tokens[1] == "list" {
				// Definition: property list <numerical-type> <numerical-type> <property-name>
				//
				// Example:
				// property list uchar int vertex_index
				// This means that the property "vertex_index" contains first an unsigned char telling how many indices the property contains,
				// followed by a list containing that many integers. Each integer in this variable-length list is an index to a vertex.
				lastElementSection := elementDefinitions[len(elementDefinitions)-1]
				property := PropertyDefinition{
					listProperty:             true,
					listCountDataType:        tokens[2],
					listElementIndexDataType: tokens[3],
					listElementType:          tokens[4],
				}
				lastElementSection.properties = append(lastElementSection.properties, &property)
			} else {
				// Definition:
				// property <data-type> <property-name-1>
				// property <data-type> <property-name-2>
				// property <data-type> <property-name-3>
				// ...
				lastElementSection := elementDefinitions[len(elementDefinitions)-1]
				property := PropertyDefinition{
					name:     tokens[2],
					dataType: tokens[1],
					index:    len(lastElementSection.properties),
				}
				lastElementSection.properties = append(lastElementSection.properties, &property)
			}

		case "end_header":
			headerLines = false
			dataStartIndex = lineIndex + 1
		default:
			error = fmt.Errorf("unknown/unexpected line type: '%s'", line)
		}

		if error != nil {
			return nil, -1, fmt.Errorf("encountered parse error on line %d: %s", lineIndex, error)
		}
	}

	return elementDefinitions, dataStartIndex, nil
}

func parseTokens(line string, delimiter rune) []string {
	f := func(c rune) bool {
		return c == delimiter
	}
	return strings.FieldsFunc(line, f)
}
