package ply

type Ply struct {
	Header   *Header
	Elements []*Element
}

type Property struct {
	Name       string
	Type       PropertyType
	IntValue   int
	FloatValue float64
}

type Element struct {
	Name          string
	ID            int
	Properties    []*Property
	References    []IndexReference
	ReferenceType string
}

type PropertyType string
type IndexReference int

type Header struct {
	Format             *TypeVersion
	Comments           []string
	TextureFilename    string
	ElementDefinitions []*ElementDefinition
}

type TypeVersion struct {
	Format  string
	Version string
}

// PropertyDefinition holds the definition from the ply file header of a Property of an Element.
// This Property definition can hold both single Property definitions and "reference list" definitions.
//
// Valid data types for a scalar Property, field dataType (according to https://paulbourke.net/dataformats/ply/):
//
//	name        type        number of bytes
//	---------------------------------------
//	char       character                 1
//	uchar      unsigned character        1
//	short      short integer             2
//	ushort     unsigned short integer    2
//	int        integer                   4
//	uint       unsigned integer          4
//	float      single-precision float    4
//	double     double-precision float    8
type PropertyDefinition struct {
	// Single Property definition
	Name     string
	DataType string
	Index    int
}

type ReferenceListDefinition struct {
	CountDataType  string // referenceCountDataType is data type for the amount of references in the list. Should always be some kind of integer type(?)
	ReferencedType string // ReferencedType the type of Element that the index references in the list refer to. (Like a facet Element with list Property refer to index of Element type vertex (vertices).)
	IDDataType     string // referenceIDDataType is the data type for index references. Should always be some kind of integer type(?)
}

type ElementDefinition struct {
	Name          string
	Count         int
	Properties    []*PropertyDefinition
	ReferenceList *ReferenceListDefinition
}
