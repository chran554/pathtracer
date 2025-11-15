package renderfile

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/scene"
	"regexp"
	"time"

	"github.com/ungerik/go3d/float64/vec2"
	"github.com/ungerik/go3d/float64/vec3"
)

type serializer struct {
	vectors          []*Vector
	vectorToIndexMap map[*vec3.T]VectorIndex

	vector2Ds          []*Vector2D
	vector2DToIndexMap map[*vec2.T]Vector2DIndex

	materials          []*Material
	materialToIndexMap map[*scene.Material]MaterialIndex

	colors          []*Color
	colorToIndexMap map[*color.Color]ColorIndex

	resourceFileMap map[string]ResourceIndex

	sv   []*vec3.T         // sv is scene vectors
	sv2d []*vec2.T         // sv2d is scene 2D vectors
	sc   []*color.Color    // sc is scene colors
	sm   []*scene.Material // sm is scene materials

	zipWriter *zip.Writer
	zipReader *zip.Reader
}

func (s *serializer) clearFrameCache() {
	s.vectors = nil
	s.vector2Ds = nil
	s.materials = nil
	s.colors = nil

	s.vectorToIndexMap = make(map[*vec3.T]VectorIndex)
	s.vector2DToIndexMap = make(map[*vec2.T]Vector2DIndex)
	s.materialToIndexMap = make(map[*scene.Material]MaterialIndex)
	s.colorToIndexMap = make(map[*color.Color]ColorIndex)

	// Not the resource file map. That is shared and reused between frames.
	// resourceFileMap = make(map[string]ResourceIndex)

	s.sv = nil
	s.sv2d = nil
	s.sc = nil
	s.sm = nil
}

func newSerializer(zipWriter *zip.Writer) *serializer {
	return &serializer{
		zipWriter:          zipWriter,
		vectorToIndexMap:   make(map[*vec3.T]VectorIndex),
		vector2DToIndexMap: make(map[*vec2.T]Vector2DIndex),
		resourceFileMap:    make(map[string]ResourceIndex),
		materialToIndexMap: make(map[*scene.Material]MaterialIndex),
		colorToIndexMap:    make(map[*color.Color]ColorIndex),
	}
}

func newDeserializer(reader *zip.Reader) (*serializer, error) {
	s := &serializer{
		zipReader:          reader,
		vectorToIndexMap:   make(map[*vec3.T]VectorIndex),
		vector2DToIndexMap: make(map[*vec2.T]Vector2DIndex),
		resourceFileMap:    make(map[string]ResourceIndex),
		materialToIndexMap: make(map[*scene.Material]MaterialIndex),
		colorToIndexMap:    make(map[*color.Color]ColorIndex),
	}

	return s, nil
}

func (s *serializer) vectorIndices(vectors []*vec3.T) []VectorIndex {
	var indices []VectorIndex
	for _, vector := range vectors {
		indices = append(indices, s.vectorIndex(vector))
	}
	return indices
}

// vector2DIndices returns a slice of indices for a slice of 2D vectors, adding new vectors to the vector2Ds slice if necessary.
func (s *serializer) vector2DIndices(vectors2D []*vec2.T) []Vector2DIndex {
	var indices []Vector2DIndex
	for _, vector2D := range vectors2D {
		indices = append(indices, s.vector2DIndex(vector2D))
	}
	return indices
}

// vector2DIndex returns the index of a 2D vector, adding it to the vector2Ds slice if it hasn't been indexed before.
func (s *serializer) vector2DIndex(vector2D *vec2.T) Vector2DIndex {
	if vector2D == nil {
		return 0
	}

	if index, exists := s.vector2DToIndexMap[vector2D]; !exists {
		index = Vector2DIndex(len(s.vector2Ds) + 1)
		s.vector2Ds = append(s.vector2Ds, &Vector2D{X: vector2D[0], Y: vector2D[1]})
		s.vector2DToIndexMap[vector2D] = index
		return index
	} else {
		return index
	}
}

func (s *serializer) fileResourceIndex(floatImage *floatimage.FloatImage) (ResourceIndex, error) {
	if floatImage == nil {
		return 0, nil
	}

	imageHash := floatImage.Hash()

	// Check if the resource has already been added
	if index, exists := s.resourceFileMap[imageHash]; exists {
		return index, nil // Reuse the existing index
	}

	// Load the binary data from the file
	fileData, err := os.ReadFile(floatImage.Name())
	if err != nil {
		return 0, fmt.Errorf("could not read resource file %s: %w", floatImage.Name(), err)
	}

	// Calculate the new resource index based on the size of the map
	newIndex := ResourceIndex(len(s.resourceFileMap) + 1)

	// Create the zip entry filename
	resourceZipFilename := fmt.Sprintf("resources/%03d_%s", newIndex, filepath.Base(floatImage.Name()))

	// Write the binary data to the zip file
	err = s.writeToZip(resourceZipFilename, fileData)
	if err != nil {
		return 0, fmt.Errorf("could not write file %s to zip entry %s: %w", floatImage.Name(), resourceZipFilename, err)
	}

	// Add the filename and its index to the map
	s.resourceFileMap[imageHash] = newIndex

	// Return the new index
	return newIndex, nil
}

// If the color is nil, the function returns 0 as the default index.
func (s *serializer) colorIndex(color *color.Color) ColorIndex {
	// Return 0 if the color is nil
	if color == nil {
		return 0
	}

	// Check if the color is already indexed
	if index, exists := s.colorToIndexMap[color]; !exists {
		index = ColorIndex(len(s.colors) + 1)
		s.colors = append(s.colors, &Color{R: color.R, G: color.G, B: color.B, A: color.A})
		s.colorToIndexMap[color] = index
		return index
	} else {
		return index
	}
}

// materialIndex maps a *scene.Material to a Material struct and assigns it a unique index.
func (s *serializer) materialIndex(material *scene.Material) (MaterialIndex, error) {
	if material == nil {
		return 0, nil
	}

	// Check if the material has already been indexed
	if materialIndex, exists := s.materialToIndexMap[material]; exists {
		return materialIndex, nil // Reuse existing index
	}

	projection, err := s.serializeProjection(material.Projection)
	if err != nil {
		return 0, err
	}

	// Map the *scene.Material to a local Material struct
	mappedMaterial := &Material{
		Name:            material.Name,
		Color:           s.colorIndex(material.Color),
		Diffuse:         material.Diffuse,
		Emission:        s.colorIndex(material.Emission),
		Glossiness:      material.Glossiness,
		Roughness:       material.Roughness,
		RefractionIndex: material.RefractionIndex,
		SolidObject:     material.SolidObject,
		Transparency:    material.Transparency,
		RayTerminator:   material.RayTerminator,
		Projection:      projection,
	}

	// Assign a new index for the material
	newIndex := MaterialIndex(len(s.materials) + 1)
	s.materials = append(s.materials, mappedMaterial)
	s.materialToIndexMap[material] = newIndex

	return newIndex, nil // Return the newly assigned material index
}

func (s *serializer) vectorIndex(v *vec3.T) VectorIndex {
	if v == nil {
		return 0
	}

	if index, exist := s.vectorToIndexMap[v]; !exist {
		index = VectorIndex(len(s.vectors) + 1)
		s.vectors = append(s.vectors, &Vector{X: v[0], Y: v[1], Z: v[2]})
		s.vectorToIndexMap[v] = index
		return index
	} else {
		return index
	}
}

// Adds a file entry to the zip via serializer
func (s *serializer) writeToZip(fileName string, data []byte) error {
	// Create a new FileHeader
	header := &zip.FileHeader{
		Name:     fileName,
		Method:   zip.Deflate, // Use compression
		Modified: time.Now(),  // Set current date and time
	}

	// Create file with the modified header
	fileWriter, err := s.zipWriter.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("could not create file %s in zip: %w", fileName, err)
	}

	// Write the given data to the file in zip
	_, err = fileWriter.Write(data)
	if err != nil {
		return fmt.Errorf("could not write data to file %s in zip: %w", fileName, err)
	}

	return nil
}

func (s *serializer) sceneColor(index ColorIndex) *color.Color {
	if index == 0 {
		return nil
	}
	return s.sc[index-1]
}

func (s *serializer) sceneVector(index VectorIndex) *vec3.T {
	if index == 0 {
		return nil
	}
	return s.sv[index-1]
}

func (s *serializer) sceneVector2D(index Vector2DIndex) *vec2.T {
	if index == 0 {
		return nil
	}
	return s.sv2d[index-1]
}

func (s *serializer) sceneVectors(indices []VectorIndex) []*vec3.T {
	var sceneVectors []*vec3.T
	for _, index := range indices {
		sceneVectors = append(sceneVectors, s.sceneVector(index))
	}
	return sceneVectors
}

func (s *serializer) sceneVectors2D(indices []Vector2DIndex) []*vec2.T {
	var sceneVectors []*vec2.T
	for _, index := range indices {
		sceneVectors = append(sceneVectors, s.sceneVector2D(index))
	}
	return sceneVectors
}

func (s *serializer) sceneMaterial(index MaterialIndex) *scene.Material {
	if index == 0 {
		return nil
	}
	return s.sm[index-1]
}

func (s *serializer) resourceImage(resourceIndex ResourceIndex) (*floatimage.FloatImage, error) {
	if resourceIndex == 0 {
		return nil, nil
	}

	for _, file := range s.zipReader.File {
		filename := file.Name

		if match, _ := regexp.Match(fmt.Sprintf("resources/%03d_.*", resourceIndex), []byte(filename)); match {
			fileReader, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("could not open resource image file %s: %w", filename, err)
			}
			defer fileReader.Close()

			floatImage, err := floatimage.GetOrReadCachedImage(filename, fileReader)
			if err != nil {
				return nil, fmt.Errorf("could not decode resource image file %s: %w", filename, err)
			}

			return floatImage, nil
		}
	}

	return nil, fmt.Errorf("could not find resource file for resource index %d", resourceIndex)
}

func readZipFileEntry(file *zip.File) ([]byte, error) {
	fileReader, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("could not open file %s: %w", file.Name, err)
	}
	defer fileReader.Close()
	fileData, err := io.ReadAll(fileReader)
	if err != nil {
		return nil, fmt.Errorf("could not read file %s: %w", file.Name, err)
	}
	return fileData, err
}
