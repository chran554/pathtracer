package wavefront

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/floatimage"
	"pathtracer/internal/pkg/scene"
	"pathtracer/internal/pkg/util"
	"slices"
	"strings"

	"github.com/ungerik/go3d/float64/vec2"
	"github.com/ungerik/go3d/float64/vec3"
)

func ReadFacetStructureOrPanic(objFilenamePath string) *scene.FacetStructure {
	objFile, err := os.Open(objFilenamePath)
	if err != nil {
		currentPath, err2 := filepath.Abs(".")
		if err2 != nil {
			currentPath = "[unknown]"
		}
		message := fmt.Sprintf("Could not open wavefront object file '%s' at current path '%s': %s\n", objFilenamePath, currentPath, err.Error())
		panic(message)
	}
	defer objFile.Close()

	facetStructure, err := ReadFacetStructure(objFile)
	if err != nil {
		message := fmt.Sprintf("Could not read wavefront object file '%s': %s\n", objFile.Name(), err.Error())
		panic(message)
	}

	return facetStructure
}

// ReadFacetStructure reads a obj file and returns a facet structure.
// The rightHandCoordinateSystem parameter specifies whether the obj file assumes a right-handed or left-handed coordinate system.
func ReadFacetStructure(objFile *os.File) (*scene.FacetStructure, error) {
	wavefront, err := Read(objFile)
	if err != nil {
		return nil, fmt.Errorf("could not read wavefront file '%s': %v", objFile.Name(), err)
	}

	facetStructure, err := convertToFacetStructure(wavefront, filepath.Base(objFile.Name()))
	if err != nil {
		return nil, fmt.Errorf("could not convert wavefront object to facet structure (wavefront file '%s'): %w\n", objFile.Name(), err)
	}

	facetStructure.UpdateBounds()
	facetStructure.UpdateNormals()

	return facetStructure, nil
}

func convertToFacetStructure(wavefront *Wavefront, filename string) (*scene.FacetStructure, error) {
	var defaultName = filepath.Base(filename)
	defaultName = strings.TrimSuffix(defaultName, filepath.Ext(defaultName))

	materialMap := make(map[string]*scene.Material)
	for name, mtl := range wavefront.Materials {
		materialMap[name] = convertToSceneMaterial(mtl)
	}

	facetCollections := make(map[facetCollectionKey][]*scene.Facet)
	smoothGroups := make(map[string][]*scene.Facet)
	wavefrontVertexToVertexMap := make(map[*vec]*vec3.T)
	wavefrontVertexNormalToVertexNormalMap := make(map[*vec]*vec3.T)
	wavefrontTextureVertexToTextureVertex := make(map[*vec]*vec2.T)

	for _, face := range wavefront.Facets {
		currentMaterialName := ""
		if face.Material != nil {
			currentMaterialName = face.Material.Name
		}

		// We need to convert wavefront.Facet to scene.Facet.

		var texture *scene.Texture
		if face.Material != nil && face.Material.MapKd != "" {
			textureFilename := filepath.Join(face.Material.MapBasePath, face.Material.MapKd)
			textureImage, err := floatimage.EmptyPlaceholderImage(textureFilename)
			if err != nil {
				return nil, err
			}
			texture = &scene.Texture{
				Image:         textureImage,
				Type:          scene.TextureTypeAlbedo,
				Interpolation: floatimage.InterpolationNearestNeighbor,
				Strength:      1.0,
			}
		}
		sceneFacet := convertToSceneFacet(face, wavefrontVertexToVertexMap, wavefrontVertexNormalToVertexNormalMap, wavefrontTextureVertexToTextureVertex, texture)
		faceTriangleFacets := sceneFacet.SplitMultiPointFacet()

		currentObjectName := face.Object
		currentGroups := face.Groups
		currentGroups = slices.DeleteFunc(currentGroups, func(s string) bool { return s == "default" }) // Remove "default" group, which is used in the wavefront specification but not used in the scene.

		groupsName := strings.Join(currentGroups, ",")
		key := facetCollectionKey{objectName: currentObjectName, groupName: groupsName, materialName: currentMaterialName}
		facetCollections[key] = append(facetCollections[key], faceTriangleFacets...)

		if face.SmoothingGroup != 0 {
			group := fmt.Sprintf("%d", face.SmoothingGroup)
			smoothGroups[group] = append(smoothGroups[group], faceTriangleFacets...)
		}
	}

	// Smooth facet groups
	for _, facets := range smoothGroups {
		facetGroup := &scene.FacetStructure{Facets: facets} // Use a temporary facet structure
		facetGroup.UpdateVertexNormals(true)
	}

	objectStructures := getObjectStructures(facetCollections, materialMap)

	for _, objectStructure := range objectStructures {
		optimiseObjectStructures(objectStructure)
	}

	fileFacetStructure := getOrCreateFileTopLevelNode(objectStructures, defaultName)

	return fileFacetStructure, nil
}

func convertToSceneFacet(f *Facet, wavefrontVertexToVertexMap map[*vec]*vec3.T, wavefrontVertexNormalToVertexNormalMap map[*vec]*vec3.T, wavefrontTextureVertexToTextureVertex map[*vec]*vec2.T, texture *scene.Texture) *scene.Facet {
	sf := &scene.Facet{}
	for _, v := range f.Vertices {
		if _, exist := wavefrontVertexToVertexMap[v]; !exist {
			wavefrontVertexToVertexMap[v] = &vec3.T{v[0], v[1], v[2]}
		}
		sf.Vertices = append(sf.Vertices, wavefrontVertexToVertexMap[v])
	}
	for _, vn := range f.VertexNormals {
		if _, exist := wavefrontVertexNormalToVertexNormalMap[vn]; !exist {
			wavefrontVertexNormalToVertexNormalMap[vn] = &vec3.T{vn[0], vn[1], vn[2]}
		}
		sf.VertexNormals = append(sf.VertexNormals, wavefrontVertexNormalToVertexNormalMap[vn])
	}

	textureCoordinates := make([]*vec2.T, 0, len(f.VertexTextureCoordinates))
	for _, vt := range f.VertexTextureCoordinates {
		if _, exist := wavefrontTextureVertexToTextureVertex[vt]; !exist {
			textureCoordinate := &vec2.T{vt[0], vt[1] /*, vt[2] */}
			wavefrontTextureVertexToTextureVertex[vt] = textureCoordinate
		}
		textureCoordinates = append(textureCoordinates, wavefrontTextureVertexToTextureVertex[vt])
	}
	if len(textureCoordinates) > 0 || texture != nil {
		facetTexture := &scene.FacetTexture{Texture: texture, Coordinates: textureCoordinates}
		sf.Textures = append(sf.Textures, facetTexture)
	}

	return sf
}

func convertToSceneMaterial(mtl *Material) *scene.Material {
	m := scene.NewMaterial()
	m.Name = mtl.Name

	if mtl.Kd != nil {
		m.Color = color.NewColor(mtl.Kd.R, mtl.Kd.G, mtl.Kd.B)
	}
	if mtl.Ke != nil {
		m.Emission = color.NewColor(mtl.Ke.R, mtl.Ke.G, mtl.Ke.B)
	}
	if mtl.Ns != -1.0 {
		// roughness ≈ sqrt(2 / (Ns + 2))
		// Ns ≈ 2 / roughness² - 2
		roughness := math.Sqrt(2 / (mtl.Ns + 2))
		m.Roughness = util.ClampFloat64(0.0, 1.0, roughness)
	}
	if mtl.Pr != -1.0 {
		m.Roughness = util.ClampFloat64(0.0, 1.0, mtl.Pr)
	}
	if mtl.Ni != -1.0 {
		m.RefractionIndex = max(0.0, mtl.Ni)
	}
	if mtl.D != 1.0 {
		m.Transparency = 1.0 - util.ClampFloat64(0.0, 1.0, mtl.D)
	}

	// MapKd and other fields if necessary...

	return m
}
