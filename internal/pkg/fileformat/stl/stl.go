package stl

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

const (
	asciiFormatKeywordSolid    = "solid"
	asciiFormatKeywordFacet    = "facet"
	asciiFormatKeywordNormal   = "normal"
	asciiFormatKeywordOuter    = "outer"
	asciiFormatKeywordLoop     = "loop"
	asciiFormatKeywordVertex   = "vertex"
	asciiFormatKeywordEndLoop  = "endloop"
	asciiFormatKeywordEndFacet = "endfacet"
	asciiFormatKeywordEndSolid = "endsolid"

	binaryFormatHeaderSize        = 80
	binaryFormatTriangleCountSize = 4
	binaryFormatTriangleDataSize  = 12 + 12 + 12 + 12 + 2
)

/*
STL-information:
https://en.wikipedia.org/wiki/STL_(file_format)
https://www.fabbers.com/tech/STL_Format

ASCII STL-file format:
The numerical data in the facet normal and vertex lines are single precision floats, for example, 1.23456E+789.

solid <name>
   facet normal <nx> <ny> <nz>
      outer loop
         vertex <v1x> <v1y> <v1z>
         vertex <v2x> <v2y> <v2z>
         vertex <v3x> <v3y> <v3z>
      endloop
   endfacet
endsolid <name>

Binary STL-file format:
Floating-point numbers are represented as 32-bit IEEE floating-point numbers and are assumed to be little-endian,
although this is not stated in documentation.

UINT8[80]    – Header                 - 80 bytes
UINT32       – Number of triangles    - 04 bytes
foreach triangle                      - 50 bytes
    REAL32[3] – Normal vector         - 12 bytes
    REAL32[3] – Vertex 1              - 12 bytes
    REAL32[3] – Vertex 2              - 12 bytes
    REAL32[3] – Vertex 3              - 12 bytes
    UINT16    – Attribute byte count  - 02 bytes
end

Last is a 2-byte ("short") unsigned integer that is the "attribute byte count".
In the standard format, this is said to be zero because the file format specification does not mention the format or usage any further.
However, there are vendors that use this field to store additional data like color information for the facet.
*/

func Read(r io.ReadSeeker) (*Stl, error) {
	if checkBinaryFormat(r) {
		return readBinary(r)
	}

	if checkASCIIFormat(r) {
		return readASCII(r)
	}

	return nil, errors.New("unparsable or illegal STL data")
}

func checkASCIIFormat(r io.ReadSeeker) bool {
	// Remember original file position, and restore file position after check
	originalPos, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return false
	}
	defer r.Seek(originalPos, io.SeekStart)

	// We verify by checking the size.
	size, err := r.Seek(0, io.SeekEnd)
	if err != nil {
		return false
	}
	r.Seek(originalPos, io.SeekStart)

	// Read enough data to check if it's an ASCII STL file.'
	stlData := make([]byte, min(400, size))
	_, err = io.ReadFull(r, stlData)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return false
	}

	stlText := strings.ToLower(string(stlData))
	return strings.Contains(stlText, asciiFormatKeywordSolid) &&
		strings.Contains(stlText, asciiFormatKeywordFacet+" "+asciiFormatKeywordNormal) &&
		strings.Contains(stlText, asciiFormatKeywordOuter+" "+asciiFormatKeywordLoop) &&
		strings.Contains(stlText, asciiFormatKeywordVertex) &&
		strings.Contains(stlText, asciiFormatKeywordEndLoop) &&
		strings.Contains(stlText, asciiFormatKeywordEndFacet) &&
		strings.Contains(stlText, asciiFormatKeywordEndSolid)
}

func checkBinaryFormat(r io.ReadSeeker) bool {
	// Remember original file position, and restore file position after check
	originalPos, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return false
	}
	defer r.Seek(originalPos, io.SeekStart)

	// We verify by checking the size.
	size, err := r.Seek(0, io.SeekEnd)
	if err != nil {
		return false
	}
	r.Seek(originalPos, io.SeekStart)

	// Attempt to read the full 80+4 byte header.
	headerData := make([]byte, binaryFormatHeaderSize+binaryFormatTriangleCountSize)
	n, err := io.ReadFull(r, headerData)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return false
	}

	if n < binaryFormatHeaderSize+binaryFormatTriangleCountSize {
		return false // too short file size for binary
	}

	triangleCountBytes := headerData[binaryFormatHeaderSize : binaryFormatHeaderSize+binaryFormatTriangleCountSize]
	triangleCount := binary.LittleEndian.Uint32(triangleCountBytes)
	expectedSize := int64(binaryFormatHeaderSize + binaryFormatTriangleCountSize + triangleCount*binaryFormatTriangleDataSize)
	return size == expectedSize
}

func readASCII(r io.Reader) (*Stl, error) {
	s := bufio.NewScanner(r)
	stl := &Stl{IsBinary: false}
	var currentFacet *Facet
	vertexMap := make(map[[3]float32]*Vertex)

	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}

		tokens := strings.Fields(line)
		if len(tokens) == 0 {
			continue
		}

		lowerToken := strings.ToLower(tokens[0])
		switch lowerToken {
		case asciiFormatKeywordSolid:
			if len(tokens) > 1 {
				stl.Header = strings.Join(tokens[1:], " ")
			}
		case asciiFormatKeywordFacet:
			currentFacet = &Facet{}
			if len(tokens) >= 5 && strings.ToLower(tokens[1]) == asciiFormatKeywordNormal {
				nx, errX := strconv.ParseFloat(tokens[2], 32)
				ny, errY := strconv.ParseFloat(tokens[3], 32)
				nz, errZ := strconv.ParseFloat(tokens[4], 32)
				if errX == nil && errY == nil && errZ == nil {
					currentFacet.Normal = Normal{float32(nx), float32(ny), float32(nz)}
				}
			}
		case asciiFormatKeywordVertex:
			if currentFacet != nil && len(tokens) >= 4 {
				vx, errX := strconv.ParseFloat(tokens[1], 32)
				vy, errY := strconv.ParseFloat(tokens[2], 32)
				vz, errZ := strconv.ParseFloat(tokens[3], 32)
				if errX == nil && errY == nil && errZ == nil {
					v3 := [3]float32{float32(vx), float32(vy), float32(vz)}
					v, ok := vertexMap[v3]
					if !ok {
						v = &Vertex{v3[0], v3[1], v3[2]}
						vertexMap[v3] = v
					}

					added := false
					for i := 0; i < 3; i++ {
						if currentFacet.Vertices[i] == nil {
							currentFacet.Vertices[i] = v
							added = true
							break
						}
					}
					if !added {
						// Too many vertices for a facet, ignore extra
					}
				}
			}
		case asciiFormatKeywordEndFacet:
			if currentFacet != nil {
				stl.Facets = append(stl.Facets, currentFacet)
				currentFacet = nil
			}
		case asciiFormatKeywordEndSolid:
			// Done
		}
	}

	if err := s.Err(); err != nil {
		return nil, err
	}

	return stl, nil
}

func readBinary(r io.Reader) (*Stl, error) {
	// Read the full 80-byte header.
	header := make([]byte, binaryFormatHeaderSize)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}

	// Read triangle count.
	var triangleCount uint32
	if err := binary.Read(r, binary.LittleEndian, &triangleCount); err != nil {
		return nil, fmt.Errorf("failed to read triangle count: %v", err)
	}

	// Preserve the full header as a string, but trim null bytes at the end for display.
	stl := &Stl{
		Header:   strings.TrimRight(string(header), "\x00"),
		IsBinary: true,
	}

	parseMagicsTags(stl, header)

	vertexMap := make(map[[3]float32]*Vertex)

	for i := uint32(0); i < triangleCount; i++ {
		var data [50]byte
		_, err := io.ReadFull(r, data[:])
		if err != nil {
			return nil, err
		}

		facet := &Facet{}
		facet.Normal.X = math.Float32frombits(binary.LittleEndian.Uint32(data[0:4]))
		facet.Normal.Y = math.Float32frombits(binary.LittleEndian.Uint32(data[4:8]))
		facet.Normal.Z = math.Float32frombits(binary.LittleEndian.Uint32(data[8:12]))

		for v := 0; v < 3; v++ {
			vx := math.Float32frombits(binary.LittleEndian.Uint32(data[12+v*12 : 16+v*12]))
			vy := math.Float32frombits(binary.LittleEndian.Uint32(data[16+v*12 : 20+v*12]))
			vz := math.Float32frombits(binary.LittleEndian.Uint32(data[20+v*12 : 24+v*12]))

			v3 := [3]float32{vx, vy, vz}
			vPtr, ok := vertexMap[v3]
			if !ok {
				vPtr = &Vertex{vx, vy, vz}
				vertexMap[v3] = vPtr
			}
			facet.Vertices[v] = vPtr
		}

		facet.AttributeByteCount = binary.LittleEndian.Uint16(data[48:50])
		stl.Facets = append(stl.Facets, facet)
	}

	return stl, nil
}

func parseMagicsTags(stl *Stl, header []byte) {
	colorTag := []byte("COLOR=")
	colorIdx := -1
	for i := 0; i <= len(header)-len(colorTag); i++ {
		match := true
		for j := 0; j < len(colorTag); j++ {
			if header[i+j] != colorTag[j] {
				match = false
				break
			}
		}
		if match {
			colorIdx = i
			break
		}
	}

	if colorIdx != -1 && colorIdx+10 <= len(header) {
		stl.Color = &Color{
			R: header[colorIdx+6],
			G: header[colorIdx+7],
			B: header[colorIdx+8],
			A: header[colorIdx+9],
		}
	}

	matTag := []byte("MATERIAL=")
	matIdx := -1
	for i := 0; i <= len(header)-len(matTag); i++ {
		match := true
		for j := 0; j < len(matTag); j++ {
			if header[i+j] != matTag[j] {
				match = false
				break
			}
		}
		if match {
			matIdx = i
			break
		}
	}

	if matIdx != -1 && matIdx+21 <= len(header) {
		// Three colors of 4 bytes each: Diffuse, Specular, Ambient
		stl.Material = &Material{
			Diffuse: Color{
				R: header[matIdx+9],
				G: header[matIdx+10],
				B: header[matIdx+11],
				A: header[matIdx+12],
			},
			Specular: Color{
				R: header[matIdx+13],
				G: header[matIdx+14],
				B: header[matIdx+15],
				A: header[matIdx+16],
			},
			Ambient: Color{
				R: header[matIdx+17],
				G: header[matIdx+18],
				B: header[matIdx+19],
				A: header[matIdx+20],
			},
		}
	}
}

func Write(w io.Writer, stl *Stl) error {
	if stl.IsBinary {
		return writeBinary(w, stl)
	}
	return writeASCII(w, stl)
}

func writeASCII(w io.Writer, stl *Stl) error {
	bw := bufio.NewWriter(w)
	header := stl.Header
	if header == "" {
		header = "Generated STL file"
	}
	_, _ = fmt.Fprintf(bw, "solid %s\n", header)

	for _, f := range stl.Facets {
		_, _ = fmt.Fprintf(bw, "  facet normal %e %e %e\n", f.Normal.X, f.Normal.Y, f.Normal.Z)
		_, _ = fmt.Fprintf(bw, "    outer loop\n")
		for _, v := range f.Vertices {
			_, _ = fmt.Fprintf(bw, "      vertex %e %e %e\n", v.X, v.Y, v.Z)
		}
		_, _ = fmt.Fprintf(bw, "    endloop\n")
		_, _ = fmt.Fprintf(bw, "  endfacet\n")
	}

	_, _ = fmt.Fprintf(bw, "endsolid %s\n", header)
	return bw.Flush()
}

func writeBinary(w io.Writer, stl *Stl) error {
	header := make([]byte, 80)
	copy(header, stl.Header)

	if stl.Color != nil {
		copy(header[0:], "COLOR=")
		header[6] = stl.Color.R
		header[7] = stl.Color.G
		header[8] = stl.Color.B
		header[9] = stl.Color.A
		if stl.Material != nil {
			copy(header[10:], ",MATERIAL=")
			header[20] = stl.Material.Diffuse.R
			header[21] = stl.Material.Diffuse.G
			header[22] = stl.Material.Diffuse.B
			header[23] = stl.Material.Diffuse.A
			header[24] = stl.Material.Specular.R
			header[25] = stl.Material.Specular.G
			header[26] = stl.Material.Specular.B
			header[27] = stl.Material.Specular.A
			header[28] = stl.Material.Ambient.R
			header[29] = stl.Material.Ambient.G
			header[30] = stl.Material.Ambient.B
			header[31] = stl.Material.Ambient.A
		}
	}

	if _, err := w.Write(header); err != nil {
		return err
	}

	triangleCount := uint32(len(stl.Facets))
	if err := binary.Write(w, binary.LittleEndian, triangleCount); err != nil {
		return err
	}

	for _, f := range stl.Facets {
		if err := binary.Write(w, binary.LittleEndian, f.Normal.X); err != nil {
			return err
		}
		if err := binary.Write(w, binary.LittleEndian, f.Normal.Y); err != nil {
			return err
		}
		if err := binary.Write(w, binary.LittleEndian, f.Normal.Z); err != nil {
			return err
		}
		for _, v := range f.Vertices {
			if err := binary.Write(w, binary.LittleEndian, v.X); err != nil {
				return err
			}
			if err := binary.Write(w, binary.LittleEndian, v.Y); err != nil {
				return err
			}
			if err := binary.Write(w, binary.LittleEndian, v.Z); err != nil {
				return err
			}
		}
		if err := binary.Write(w, binary.LittleEndian, f.AttributeByteCount); err != nil {
			return err
		}
	}

	return nil
}

// Color return facet color as defined by vendors Materialise Magic, VisCAM and SolidView.
// If the color is not defined for the facet, nil is returned.
//
// https://en.wikipedia.org/wiki/STL_(file_format)
func (f *Facet) Color(vendor Vendor) *Color {
	a := f.AttributeByteCount

	if vendor == VisCAM || vendor == SolidView {
		// VisCAM and SolidView:
		// bit 15 is 1 if the color is valid
		// bits 0 to 4 are blue (0 to 31)
		// bits 5 to 9 are green (0 to 31)
		// bits 10 to 14 are red (0 to 31)
		validColor := a&0x8000 != 0
		if validColor {
			return &Color{
				R: uint8((a >> 10 & 0x1F) << 3),
				G: uint8((a >> 5 & 0x1F) << 3),
				B: uint8((a & 0x1F) << 3),
				A: 255,
			}
		}
	} else if vendor == MaterialiseMagics {
		// Materialise Magics:
		// bit 15 is 0 if this facet has its own unique color
		// bits 0 to 4 are red (0 to 31)
		// bits 5 to 9 are green (0 to 31)
		// bits 10 to 14 are blue (0 to 31)
		validColor := a&0x8000 == 0
		if validColor {
			return &Color{
				R: uint8((a & 0x1F) << 3),
				G: uint8((a >> 5 & 0x1F) << 3),
				B: uint8((a >> 10 & 0x1F) << 3),
				A: 255,
			}
		}
	}

	return nil
}
