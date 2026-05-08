package stl

type Vendor int

const (
	VisCAM Vendor = iota
	SolidView
	MaterialiseMagics
)

type Vertex struct {
	X, Y, Z float32
}

type Normal struct {
	X, Y, Z float32
}

type Color struct {
	R, G, B, A uint8
}

type Material struct {
	Diffuse  Color
	Specular Color
	Ambient  Color
}

type Facet struct {
	Normal             Normal
	Vertices           [3]*Vertex
	AttributeByteCount uint16
}

type Stl struct {
	Header   string // Used as name of solid for ASCII STL files, or as header for binary files
	Facets   []*Facet
	Color    *Color
	Material *Material
	IsBinary bool
}
