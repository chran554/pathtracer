package stl

import (
	"bytes"
	"reflect"
	"testing"
)

func TestReadWriteASCII(t *testing.T) {
	v1 := &Vertex{0, 0, 0}
	v2 := &Vertex{1, 0, 0}
	v3 := &Vertex{0, 1, 0}

	stl := &Stl{
		Header: "TestASCII",
		Facets: []*Facet{
			{
				Normal:   Normal{0, 0, 1},
				Vertices: [3]*Vertex{v1, v2, v3},
			},
		},
		IsBinary: false,
	}

	var buf bytes.Buffer
	err := Write(&buf, stl)
	if err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	stlRead, err := Read(r)
	if err != nil {
		t.Fatal(err)
	}

	if len(stlRead.Facets) != 1 {
		t.Errorf("expected 1 facet, got %d", len(stlRead.Facets))
	}

	if stlRead.IsBinary {
		t.Error("expected ASCII STL, got Binary")
	}
}

func TestReadASCIICaseInsensitive(t *testing.T) {
	asciiData := `SOLID CaseTest
  FACET NORMAL 0.0 0.0 1.0
    OUTER LOOP
      VERTEX 0.0 0.0 0.0
      VERTEX 1.0 0.0 0.0
      VERTEX 0.0 1.0 0.0
    ENDLOOP
  ENDFACET
ENDSOLID CaseTest`

	r := bytes.NewReader([]byte(asciiData))
	stlRead, err := Read(r)
	if err != nil {
		t.Fatal(err)
	}

	if len(stlRead.Facets) != 1 {
		t.Errorf("expected 1 facet, got %d", len(stlRead.Facets))
	}
	if stlRead.IsBinary {
		t.Error("expected ASCII")
	}
}

func TestReadWriteBinary(t *testing.T) {
	v1 := &Vertex{0, 0, 0}
	v2 := &Vertex{1, 0, 0}
	v3 := &Vertex{0, 1, 0}

	// Use another triangle that shares vertices
	v4 := &Vertex{1, 1, 0}

	stl := &Stl{
		Header: "TestBinary",
		Facets: []*Facet{
			{
				Normal:   Normal{0, 0, 1},
				Vertices: [3]*Vertex{v1, v2, v3},
			},
			{
				Normal:   Normal{0, 0, 1},
				Vertices: [3]*Vertex{v2, v4, v3},
			},
		},
		IsBinary: true,
		Color:    &Color{R: 255, G: 0, B: 0, A: 255},
	}

	var buf bytes.Buffer
	err := Write(&buf, stl)
	if err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	stlRead, err := Read(r)
	if err != nil {
		t.Fatal(err)
	}

	if len(stlRead.Facets) != 2 {
		t.Errorf("expected 2 facets, got %d", len(stlRead.Facets))
	}

	if !stlRead.IsBinary {
		t.Error("expected Binary STL, got ASCII")
	}

	if stlRead.Color == nil || stlRead.Color.R != 255 {
		t.Errorf("expected red color in header, got %+v", stlRead.Color)
	}

	// Check vertex sharing
	f1 := stlRead.Facets[0]
	f2 := stlRead.Facets[1]

	// f1 uses v1, v2, v3
	// f2 uses v2, v4, v3
	// So f1.Vertices[1] should be the same pointer as f2.Vertices[0] (v2)
	// and f1.Vertices[2] should be the same pointer as f2.Vertices[2] (v3)

	if f1.Vertices[1] != f2.Vertices[0] {
		t.Error("Vertex 2 not shared between facets")
	}
	if f1.Vertices[2] != f2.Vertices[2] {
		t.Error("Vertex 3 not shared between facets")
	}
}

func TestMagicsTags(t *testing.T) {
	stl := &Stl{
		Header:   "MagicsTagsTest",
		IsBinary: true,
		Color:    &Color{R: 10, G: 20, B: 30, A: 40},
		Material: &Material{
			Diffuse:  Color{1, 2, 3, 4},
			Specular: Color{5, 6, 7, 8},
			Ambient:  Color{9, 10, 11, 12},
		},
		Facets: []*Facet{
			{
				Normal:   Normal{0, 0, 1},
				Vertices: [3]*Vertex{{X: 0, Y: 0, Z: 0}, {X: 1, Y: 0, Z: 0}, {X: 0, Y: 1, Z: 0}},
			},
		},
	}

	var buf bytes.Buffer
	err := Write(&buf, stl)
	if err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(buf.Bytes())
	stlRead, err := Read(r)
	if err != nil {
		t.Fatal(err)
	}

	if stlRead.Color == nil || stlRead.Color.R != 10 || stlRead.Color.A != 40 {
		t.Errorf("Color mismatch: %+v", stlRead.Color)
	}

	if stlRead.Material == nil || stlRead.Material.Diffuse.R != 1 || stlRead.Material.Ambient.B != 11 {
		t.Errorf("Material mismatch: %+v", stlRead.Material)
	}
}

func TestFacet_Color(t *testing.T) {
	tests := []struct {
		name               string
		attributeByteCount uint16
		vendor             Vendor
		want               *Color
		testDirect         bool
	}{
		{
			name:               "VisCAM Valid Color",
			attributeByteCount: 0x8000 | (31 << 10) | (15 << 5) | 7, // R=31, G=15, B=7, Valid=1
			vendor:             VisCAM,
			want:               &Color{R: 31 << 3, G: 15 << 3, B: 7 << 3, A: 255},
		},
		{
			name:               "VisCAM Invalid Color",
			attributeByteCount: (31 << 10) | (15 << 5) | 7, // Valid=0
			vendor:             VisCAM,
			want:               nil,
		},
		{
			name:               "Materialise Magics Valid Color",
			attributeByteCount: (7 << 10) | (15 << 5) | 31, // B=7, G=15, R=31, Valid=0
			vendor:             MaterialiseMagics,
			want:               &Color{R: 31 << 3, G: 15 << 3, B: 7 << 3, A: 255},
		},
		{
			name:               "Materialise Magics Invalid Color (per-object)",
			attributeByteCount: 0x8000 | (7 << 10) | (15 << 5) | 31, // Valid=1
			vendor:             MaterialiseMagics,
			want:               nil,
		},
		{
			name:               "SolidView Valid Color",
			attributeByteCount: 0x8000 | (31 << 10) | (15 << 5) | 7,
			vendor:             SolidView,
			want:               &Color{R: 31 << 3, G: 15 << 3, B: 7 << 3, A: 255},
		},
		{
			name:               "AttributeByteCount directly",
			attributeByteCount: 0x8000 | (31 << 10) | (15 << 5) | 7,
			vendor:             VisCAM,
			want:               &Color{R: 31 << 3, G: 15 << 3, B: 7 << 3, A: 255},
			testDirect:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.testDirect {
				abc := Facet{AttributeByteCount: tt.attributeByteCount}
				if got := abc.Color(tt.vendor); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("AttributeByteCount.Color() = %v, want %v", got, tt.want)
				}
				return
			}
			f := &Facet{
				AttributeByteCount: tt.attributeByteCount,
			}
			if got := f.Color(tt.vendor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Facet.Color() = %v, want %v", got, tt.want)
			}
		})
	}
}
