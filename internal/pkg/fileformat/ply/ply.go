package ply

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
)

const (
	PropertyTypeUnknown        PropertyType = "unknown"
	PropertyTypeInt            PropertyType = "int"
	PropertyTypeFloat          PropertyType = "float"
	PropertyTypeIndexReference PropertyType = "indexReference"
)

func (f *TypeVersion) Binary() bool {
	return strings.HasPrefix(strings.ToLower(f.Format), "binary")
}

func (f *TypeVersion) ASCII() bool {
	return strings.HasPrefix(strings.ToLower(f.Format), "ascii")
}

func (f *TypeVersion) ByteOrder() binary.ByteOrder {
	var byteOrder binary.ByteOrder = binary.LittleEndian
	if strings.HasSuffix(strings.ToLower(f.Format), "big_endian") {
		byteOrder = binary.BigEndian
	}

	return byteOrder
}

// String implementation for debugging and logging clarity
func (pd *PropertyDefinition) String() string {
	if pd == nil {
		return "<nil propertyDefinition>"
	}
	return fmt.Sprintf("Property [%d]:'%s' (%s)", pd.Index, pd.Name, pd.DataType)
}

func (rld *ReferenceListDefinition) getReferenceType() string {
	s := rld.ReferencedType

	etr := strings.ToLower(s)
	lastIndex := strings.LastIndex(etr, "_")
	if (lastIndex != -1) && (etr[lastIndex:] == "_index" || etr[lastIndex:] == "_indices") {
		s = s[:lastIndex]
	}

	return s
}

// String implementation for debugging and logging clarity
func (rld *ReferenceListDefinition) String() string {
	if rld == nil {
		return "<nil referenceListDefinition>"
	}
	return fmt.Sprintf("Property list %s %s %s", rld.CountDataType, rld.IDDataType, rld.ReferencedType)
}

func (ed *ElementDefinition) String() string {
	if ed == nil {
		return "<nil elementDefinition>"
	}
	var props []string
	for _, p := range ed.Properties {
		if p == nil {
			props = append(props, "<nil>")
			continue
		}
		props = append(props, p.String())
	}
	return fmt.Sprintf("Element %s (count:%d) { %s }", ed.Name, ed.Count, strings.Join(props, ", "))
}

func (pt PropertyType) String() string { return string(pt) }

func (p *Property) String() string {
	if p == nil {
		return "<nil Property>"
	}
	switch p.Type {
	case PropertyTypeInt, PropertyTypeIndexReference:
		return fmt.Sprintf("%s=%d(%s)", p.Name, p.IntValue, p.Type)
	case PropertyTypeFloat:
		return fmt.Sprintf("%s=%g(%s)", p.Name, p.FloatValue, p.Type)
	default:
		return fmt.Sprintf("%s<?>(%s)", p.Name, p.Type)
	}
}

func (e *Element) String() string {
	if e == nil {
		return "<nil Element>"
	}
	var vals []string
	for _, v := range e.Properties {
		if v == nil {
			vals = append(vals, "<nil>")
			continue
		}
		vals = append(vals, v.String())
	}
	return fmt.Sprintf("%s[%d]{ %s }", e.Name, e.ID, strings.Join(vals, ", "))
}

// Read
//
// https://web.archive.org/web/20161204152348/http://www.dcs.ed.ac.uk/teaching/cs4/www/graphics/Web/ply.html
// https://www.mathworks.com/help/vision/ug/the-ply-format.html
// http://paulbourke.net/dataformats/ply/
func Read(r *bufio.Reader) (ply *Ply, err error) {
	format, elementDefinitions, _, fullLineComments, textureFilename, err := parsePlyHeaderSection(r)
	if err != nil {
		return nil, err
	}

	var elementValues []*Element

	// Parse PLY in ASCII format
	if format.ASCII() && format.Version == "1.0" {
		fmt.Printf("Reading ascii ply: %s v%s\nStructure:\n", format.Format, format.Version)
		if elementValues, err = parsePlyASCIIDataSection(r, elementDefinitions); err != nil {
			return nil, fmt.Errorf("could not parse ascii ply: %w", err)
		}

	} else if format.Binary() && format.Version == "1.0" {
		fmt.Printf("Reading binary ply: %s v%s\nStructure:\n", format.Format, format.Version)
		for _, elementDef := range elementDefinitions {
			fmt.Printf("\t%+v\n", elementDef)
		}

		if elementValues, err = parsePlyBinaryDataSection(r, elementDefinitions, format.ByteOrder()); err != nil {
			return nil, fmt.Errorf("could not parse binary ply: %w", err)
		}

	} else {
		return nil, fmt.Errorf("unsupported ply format: %s v%s", format.Format, format.Version)
	}

	return &Ply{
		Header: &Header{
			Format:             format,
			Comments:           fullLineComments,
			TextureFilename:    textureFilename,
			ElementDefinitions: elementDefinitions,
		},
		Elements: elementValues,
	}, nil
}

func parsePlyBinaryDataSection(r io.Reader, elementDefinitions []*ElementDefinition, byteOrder binary.ByteOrder) (elements []*Element, error error) {
	for _, elementDef := range elementDefinitions {
		fmt.Printf("\tReading %d %s\n", elementDef.Count, elementDef.Name)

		for elementIndex := 0; elementIndex < elementDef.Count; elementIndex++ {
			element, err := parseBinaryElement(r, elementIndex, elementDef, byteOrder)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return nil, err
			}

			elements = append(elements, element)
		}
	}

	return elements, nil
}

func parseBinaryElement(r io.Reader, elementID int, elementDefinition *ElementDefinition, byteOrder binary.ByteOrder) (*Element, error) {
	element := &Element{Name: elementDefinition.Name, ID: elementID}

	if elementDefinition.ReferenceList != nil {
		expectedAmountReferences, err := readInt(r, elementDefinition.ReferenceList.CountDataType, byteOrder)
		if err != nil {
			return nil, err
		}

		element.ReferenceType = strings.ToLower(elementDefinition.ReferenceList.getReferenceType())

		for referenceIndex := 0; referenceIndex < expectedAmountReferences; referenceIndex++ {
			val, err := readInt(r, elementDefinition.ReferenceList.IDDataType, byteOrder)
			if err != nil {
				return nil, err
			}
			referenceID := IndexReference(val)
			element.References = append(element.References, referenceID)
		}
	}

	if len(elementDefinition.Properties) > 0 {
		for _, elementProperty := range elementDefinition.Properties {

			// name        type        number of bytes
			// ---------------------------------------
			// char       character                 1
			// uchar      unsigned character        1
			// short      short integer             2
			// ushort     unsigned short integer    2
			// int        integer                   4
			// uint       unsigned integer          4
			// float      single-precision float    4
			// double     double-precision float    8

			var iVal int
			var fVal float64

			switch elementProperty.DataType {
			case "char", "uchar", "short", "ushort", "int", "uint":
				val, err := readInt(r, elementProperty.DataType, byteOrder)
				if err != nil {
					return nil, err
				}
				iVal = val
			case "float", "double":
				val, err := readFloat(r, elementProperty.DataType, byteOrder)
				if err != nil {
					return nil, err
				}
				fVal = val

			default:
				return nil, fmt.Errorf("could not recognize binary Property data type '%s' for Property '%s'", elementProperty.DataType, elementProperty.Name)
			}

			element.Properties = append(element.Properties, &Property{
				Name:       elementProperty.Name,
				Type:       PropertyType(elementProperty.DataType),
				IntValue:   iVal,
				FloatValue: fVal,
			})
		}
	}

	return element, nil
}

func parsePlyASCIIDataSection(r *bufio.Reader, elementDefinitions []*ElementDefinition) (elements []*Element, error error) {
	scanner := bufio.NewScanner(r)

	for _, elementDef := range elementDefinitions {
		fmt.Printf("\tReading %d %s\n", elementDef.Count, elementDef.Name)

		for elementIndex := 0; elementIndex < elementDef.Count; elementIndex++ {
			if !scanner.Scan() {
				return nil, fmt.Errorf("ran out of ply lines before expected end of file")
			}
			line := scanner.Text()

			if strings.TrimSpace(line) == "" {
				continue
			}

			element, err := parseAsciiElement(line, elementIndex, elementDef)
			if err != nil {
				return nil, err
			}

			elements = append(elements, element)
		}
	}

	return elements, nil
}

func parseAsciiElement(line string, elementID int, elementDefinition *ElementDefinition) (*Element, error) {
	element := &Element{Name: elementDefinition.Name, ID: elementID}

	tokens := parseTokens(line, ' ')

	if elementDefinition.ReferenceList != nil {
		expectedAmountReferences, err := strconv.Atoi(tokens[0])
		if err != nil {
			return nil, fmt.Errorf("could not parse '%s' (from %s line '%s') to expected amount references of value reference list: %w", tokens[0], elementDefinition.Name, line, err)
		}
		if expectedAmountReferences != (len(tokens) - 1) {
			return nil, fmt.Errorf("expected amount references %d did not match actual amount %d references (tokens %+v)", expectedAmountReferences, len(tokens)-1, tokens)
		}

		element.ReferenceType = strings.ToLower(elementDefinition.ReferenceList.getReferenceType())
		for _, token := range tokens[1:] {
			referenceID := IndexReference(getIntValue(PropertyTypeIndexReference, token))
			element.References = append(element.References, referenceID)
		}
	}

	if len(elementDefinition.Properties) > 0 {
		for tokenIndex, token := range tokens {
			propertyDef := elementDefinition.Properties[tokenIndex]
			propertyType := getPropertyType(propertyDef.DataType)
			propertyValue := &Property{
				Name:       strings.ToLower(propertyDef.Name),
				Type:       propertyType,
				IntValue:   getIntValue(propertyType, token),
				FloatValue: getFloatValue(propertyType, token),
			}
			element.Properties = append(element.Properties, propertyValue)
		}
	}

	return element, nil
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

func getIntValue(valueType PropertyType, text string) int {
	if valueType == PropertyTypeIndexReference || valueType == PropertyTypeInt {
		value, err := strconv.Atoi(text)
		if err != nil {
			fmt.Printf("could not parse text '%s' to int value.\n", text)
			return 0
		}
		return value
	}
	return 0
}

func getFloatValue(valueType PropertyType, text string) float64 {
	if valueType == PropertyTypeFloat {
		value, err := strconv.ParseFloat(text, 64)
		if err != nil {
			fmt.Printf("could not parse text '%s' to float value.\n", text)
			return 0
		}
		return value
	}

	return 0
}

// getPropertyType gets the value data type to use in runtime data structure based on ply storage format.
// A parsed ply property value is stored (runtime) in a data type that large enough to contain all datatypes of either
// integer or float type.
//
//	name    type       number of bytes
//	----    ------     ---------------
//	char    character                1
//	uchar   unsigned character       1
//	short   short integer            2
//	ushort  unsigned short integer   2
//	int     integer                  4
//	uint    unsigned integer         4
//	float   single-precision float   4
//	double  double-precision float   8
func getPropertyType(dataType string) PropertyType {
	switch dataType {
	case "char", "uchar", "short", "ushort", "int", "uint":
		return PropertyTypeInt // Map data types to int in internal handling

	case "float", "double":
		return PropertyTypeFloat // Map data types to float (float64) in internal handling

	case "int32", "uint32", "int64", "uint64": // Unspecified int data type by specification
		return PropertyTypeInt

	case "float32", "float64": // Unspecified float data types by specification
		return PropertyTypeFloat

	default:
		fmt.Printf("unknown data type: '%s'\n", dataType)
		return PropertyTypeUnknown
	}
}

func elementDefinitionForElementIndex(elementDefinitions []*ElementDefinition, elementIndex int) (definition *ElementDefinition, elementID int, err error) {
	elementDataStartIndex := 0
	for _, elementDef := range elementDefinitions {
		if (elementIndex >= elementDataStartIndex) && (elementIndex < (elementDataStartIndex + elementDef.Count)) {
			return elementDef, elementIndex - elementDataStartIndex + 1, nil
		}

		elementDataStartIndex += elementDef.Count
	}
	return nil, -1, fmt.Errorf("could not find Element definition for Element with index %d", elementIndex)
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

func parsePlyHeaderSection(r *bufio.Reader) (format *TypeVersion, elementDefinitions []*ElementDefinition, dataStartIndex int, fullLineComments []string, textureFilename string, error error) {
	headerLines := true

	firstLine, err := r.ReadString('\n')
	if err != nil {
		return nil, nil, 0, fullLineComments, textureFilename, err
	}
	firstLine = strings.TrimSpace(firstLine)
	if !(firstLine == "ply" || firstLine == "Ply" || firstLine == "PLY") {
		return nil, nil, -1, fullLineComments, textureFilename, fmt.Errorf("can not recognise PLY file as it do not start with magic number 'ply'")
	}

	format = &TypeVersion{}

	lineIndex := 1
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, nil, 0, fullLineComments, textureFilename, err
		}
		line = strings.TrimSpace(line)
		lineIndex++

		commentIndex := strings.Index(line, "comment")

		// Comment line
		if commentIndex == 0 {
			comment := strings.TrimSpace(line[7:])
			fullLineComments = append(fullLineComments, comment)

			commentFirstWord, commentRest, _ := strings.Cut(strings.ToLower(comment), " ")
			if slices.Contains([]string{"texture", "texture:", "texturefile", "texturefile:"}, commentFirstWord) {
				textureFilename = strings.TrimSpace(commentRest)
			}
			continue
		}

		// Remove trailing line comment
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
			format.Format = tokens[1]
			format.Version = tokens[2]
		case "element":
			// Definition: Element <Element-name> <number-in-file>
			count, _ := strconv.Atoi(tokens[2])
			elementSection := ElementDefinition{Name: tokens[1], Count: count, Properties: nil}
			elementDefinitions = append(elementDefinitions, &elementSection)
		case "property":
			if tokens[1] == "list" {
				// Definition: Property list <numerical-type> <numerical-type> <Property-name>
				//
				// Example:
				// Property list uchar int vertex_index
				// This means that the Property "vertex_index" contains first an unsigned char telling how many indices the Property contains,
				// followed by a list containing that many integers. Each integer in this variable-length list is an index to a vertex.
				lastElementSection := elementDefinitions[len(elementDefinitions)-1]
				referenceList := &ReferenceListDefinition{
					CountDataType:  tokens[2],
					IDDataType:     tokens[3],
					ReferencedType: tokens[4],
				}
				lastElementSection.ReferenceList = referenceList
			} else {
				// Definition:
				// Property <data-type> <Property-name-1>
				// Property <data-type> <Property-name-2>
				// Property <data-type> <Property-name-3>
				// ...
				lastElementSection := elementDefinitions[len(elementDefinitions)-1]
				property := PropertyDefinition{
					Name:     tokens[2],
					DataType: tokens[1],
					Index:    len(lastElementSection.Properties),
				}
				lastElementSection.Properties = append(lastElementSection.Properties, &property)
			}

		case "end_header":
			headerLines = false
			dataStartIndex = lineIndex + 1
		default:
			error = fmt.Errorf("unknown/unexpected line type: '%s'", line)
		}

		if error != nil {
			return nil, nil, -1, fullLineComments, textureFilename, fmt.Errorf("encountered parse error on line %d: %s", lineIndex, error)
		}

		if !headerLines {
			break
		}

	}

	return format, elementDefinitions, dataStartIndex, fullLineComments, textureFilename, nil
}

func parseTokens(line string, delimiter rune) []string {
	f := func(c rune) bool {
		return c == delimiter
	}
	return strings.FieldsFunc(line, f)
}

func readFloat(r io.Reader, floatType string, byteOrder binary.ByteOrder) (float64, error) {
	var fVal float64

	switch floatType {
	case "float":
		var val float32
		err := binary.Read(r, byteOrder, &val)
		if err != nil {
			return 0, err
		}
		fVal = float64(val)
	case "double":
		var val float64
		err := binary.Read(r, byteOrder, &val)
		if err != nil {
			return 0, err
		}
		fVal = val
	default:
		return 0, fmt.Errorf("could not recognize float type '%s' when reading binary value", floatType)
	}

	return fVal, nil
}

func readInt(r io.Reader, intType string, byteOrder binary.ByteOrder) (int, error) {
	var iVal int

	switch intType {
	case "char":
		var val int8
		err := binary.Read(r, byteOrder, &val)
		if err != nil {
			return 0, err
		}
		iVal = int(val)
	case "uchar":
		var val uint8
		err := binary.Read(r, byteOrder, &val)
		if err != nil {
			return 0, err
		}
		iVal = int(val)
	case "short":
		var val int16
		err := binary.Read(r, byteOrder, &val)
		if err != nil {
			return 0, err
		}
		iVal = int(val)
	case "ushort":
		var val uint16
		err := binary.Read(r, byteOrder, &val)
		if err != nil {
			return 0, err
		}
		iVal = int(val)
	case "int":
		var val int32
		err := binary.Read(r, byteOrder, &val)
		if err != nil {
			return 0, err
		}
		iVal = int(val)
	case "uint":
		var val uint32
		err := binary.Read(r, byteOrder, &val)
		if err != nil {
			return 0, err
		}
		iVal = int(val)
	default:
		return 0, fmt.Errorf("could not recognize integer type '%s' when reading binary value", intType)
	}

	return iVal, nil
}
