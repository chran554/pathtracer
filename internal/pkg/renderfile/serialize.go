package renderfile

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"os"
	"pathtracer/internal/pkg/scene"
	"pathtracer/internal/pkg/util"

	"github.com/vmihailenco/msgpack/v5"
)

func WriteRenderFile(filename string, animation *scene.Animation) error {
	zipFile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could not create render file '%s': %w", filename, err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer func() {
		err := zipWriter.Close()
		if err != nil {
			return
		}
		fileStat, err := zipFile.Stat()
		if err != nil {
			return
		}
		fmt.Printf("Finished writing render file: %s (%s)\n", zipFile.Name(), util.ByteCountIEC(fileStat.Size()))
	}()

	s := newSerializer(zipWriter)

	animationInformation, _ := s.serializeAnimation(animation)

	animationInformationData, err := json.MarshalIndent(animationInformation, "", "  ")
	if err != nil {
		return err
	}
	err = s.writeToZip("animation.json", animationInformationData)
	if err != nil {
		return err
	}

	return nil
}

func (s *serializer) serializeAnimation(animation *scene.Animation) (*AnimationInformation, error) {
	framesInformation, err := s.serializeFrameFiles(animation.Frames)
	if err != nil {
		return nil, err
	}

	a := &AnimationInformation{
		Name:               animation.AnimationName,
		Width:              animation.Width,
		Height:             animation.Height,
		WriteRawImageFile:  animation.WriteRawImageFile,
		WriteImageInfoFile: animation.WriteImageInfoFile,
		FramesInformation:  framesInformation,
	}

	return a, nil
}

func (s *serializer) serializeFrameFiles(frames []*scene.Frame) ([]*FrameInformation, error) {
	var framesInformation []*FrameInformation
	for _, frame := range frames {
		frameInformation, err := s.serializeFrameFile(frame)
		if err != nil {
			return nil, err
		}
		framesInformation = append(framesInformation, frameInformation)
	}
	return framesInformation, nil
}

func (s *serializer) serializeFrameFile(frame *scene.Frame) (*FrameInformation, error) {
	s.clearFrameCache()

	filePrefix := frame.Filename
	if frame.Index != -1 {
		filePrefix = fmt.Sprintf("%06d", frame.Index)
	}

	vectorsFilename := fmt.Sprintf("frames/%s.vector.msgpack", filePrefix)
	vectors2DFilename := fmt.Sprintf("frames/v%s.vector2d.msgpack", filePrefix)
	materialsFilename := fmt.Sprintf("frames/%s.material.msgpack", filePrefix)
	colorsFilename := fmt.Sprintf("frames/%s.color.msgpack", filePrefix)
	frameFilename := fmt.Sprintf("frames/%s.frame.msgpack", filePrefix)

	camera, err := s.serializeCamera(frame.Camera)
	if err != nil {
		return nil, err
	}

	sceneNode, _ := s.serializeSceneNode(frame.SceneNode)

	f := &Frame{
		Filename:  frame.Filename,
		Index:     frame.Index,
		Camera:    camera,
		SceneNode: sceneNode,
	}

	err = s.writeMarshalledDataToZipEntry(f, frameFilename)
	if err != nil {
		return nil, err
	}

	err = s.writeMarshalledDataToZipEntry(s.vectors, vectorsFilename)
	if err != nil {
		return nil, err
	}

	err = s.writeMarshalledDataToZipEntry(s.vector2Ds, vectors2DFilename)
	if err != nil {
		return nil, err
	}

	err = s.writeMarshalledDataToZipEntry(s.materials, materialsFilename)
	if err != nil {
		return nil, err
	}

	err = s.writeMarshalledDataToZipEntry(s.colors, colorsFilename)
	if err != nil {
		return nil, err
	}

	frameInformation := &FrameInformation{
		Index:        frame.Index,
		Filename:     frame.Filename,
		FrameFile:    frameFilename,
		VectorFile:   vectorsFilename,
		Vector2DFile: vectors2DFilename,
		MaterialFile: materialsFilename,
		ColorFile:    colorsFilename,
	}

	return frameInformation, nil
}

func (s *serializer) writeMarshalledDataToZipEntry(data interface{}, filename string) error {
	marshalledData, err := msgpack.Marshal(data)
	if err != nil {
		return err
	}
	err = s.writeToZip(filename, marshalledData)
	if err != nil {
		return err
	}
	return nil
}

func (s *serializer) serializeCamera(camera *scene.Camera) (*Camera, error) {
	apertureResourceIndex, err := s.fileResourceIndex(camera.ApertureShape)
	if err != nil {
		return nil, err
	}

	return &Camera{
		Origin:            s.vectorIndex(camera.Origin),
		Heading:           s.vectorIndex(camera.Heading),
		ViewUp:            s.vectorIndex(camera.ViewUp),
		ViewPlaneDistance: camera.ViewPlaneDistance,
		ApertureSize:      camera.ApertureSize,
		ApertureShape:     apertureResourceIndex,
		FocusDistance:     camera.FocusDistance,
		Samples:           camera.Samples,
		AntiAlias:         camera.AntiAlias,
		Magnification:     camera.Magnification,
		RenderType:        string(camera.RenderType),
		RecursionDepth:    camera.RecursionDepth,
	}, nil
}

func (s *serializer) serializeSceneNode(sceneNode *scene.SceneNode) (*SceneNode, error) {
	serializedDiscs, err := s.serializeDiscs(sceneNode.Discs)
	if err != nil {
		return nil, err
	}

	serializedSceneNodes, err := s.serializeSceneNodes(sceneNode.ChildNodes)
	if err != nil {
		return nil, err
	}

	serializedSpheres, err := s.serializeSpheres(sceneNode.Spheres)
	if err != nil {
		return nil, err
	}

	serializedFacetStructures, err := s.serializeFacetStructures(sceneNode.FacetStructures)
	if err != nil {
		return nil, err
	}

	return &SceneNode{
		Spheres:         serializedSpheres,
		Discs:           serializedDiscs,
		ChildNodes:      serializedSceneNodes,
		FacetStructures: serializedFacetStructures,
	}, nil
}

func (s *serializer) serializeSceneNodes(sceneNodes []*scene.SceneNode) ([]*SceneNode, error) {
	var nodes []*SceneNode
	for _, sceneNode := range sceneNodes {
		serializedSceneNode, err := s.serializeSceneNode(sceneNode)
		if err != nil {
			return nil, err
		}

		nodes = append(nodes, serializedSceneNode)
	}
	return nodes, nil
}

func (s *serializer) serializeFacetStructures(facetStructures []*scene.FacetStructure) ([]*FacetStructure, error) {
	var structures []*FacetStructure
	for _, facetStructure := range facetStructures {
		serializedFacetStructure, err := s.serializeFacetStructure(facetStructure)
		if err != nil {
			return nil, err
		}

		structures = append(structures, serializedFacetStructure)
	}
	return structures, nil
}

func (s *serializer) serializeFacetStructure(facetStructure *scene.FacetStructure) (*FacetStructure, error) {
	materialIndex, err := s.materialIndex(facetStructure.Material)
	if err != nil {
		return nil, err
	}

	serializedFacetStructures, err := s.serializeFacetStructures(facetStructure.FacetStructures)
	if err != nil {
		return nil, err
	}

	return &FacetStructure{
		Name:             facetStructure.Name,
		SubstructureName: facetStructure.SubstructureName,
		Material:         materialIndex,
		Facets:           s.serializeFacets(facetStructure.Facets),
		FacetStructures:  serializedFacetStructures,
		IgnoreBounds:     facetStructure.IgnoreBounds,
	}, nil
}

func (s *serializer) serializeFacets(facets []*scene.Facet) []*Facet {
	var serializedFacets []*Facet
	for _, facet := range facets {
		serializedFacets = append(serializedFacets, s.serializeFacet(facet))
	}
	return serializedFacets
}

func (s *serializer) serializeFacet(facet *scene.Facet) *Facet {
	return &Facet{
		Vertices:           s.vectorIndices(facet.Vertices),
		VertexNormals:      s.vectorIndices(facet.VertexNormals),
		TextureCoordinates: s.vector2DIndices(facet.TextureCoordinates),
	}
}

func (s *serializer) serializeSpheres(spheres []*scene.Sphere) ([]*Sphere, error) {
	var serializedSpheres []*Sphere
	for _, sphere := range spheres {
		serializedSphere, err := s.serializeSphere(sphere)
		if err != nil {
			return nil, err
		}

		serializedSpheres = append(serializedSpheres, serializedSphere)
	}
	return serializedSpheres, nil
}

func (s *serializer) serializeSphere(sphere *scene.Sphere) (*Sphere, error) {
	materialIndex, err := s.materialIndex(sphere.Material)
	if err != nil {
		return nil, err
	}

	return &Sphere{
		Name:     sphere.Name,
		Origin:   s.vectorIndex(sphere.Origin),
		Radius:   sphere.Radius,
		Material: materialIndex,
	}, nil
}

func (s *serializer) serializeDiscs(discs []*scene.Disc) ([]*Disc, error) {
	var serializedDiscs []*Disc
	for _, disc := range discs {
		serializedDisc, err := s.serializeDisc(disc)
		if err != nil {
			return nil, err
		}
		serializedDiscs = append(serializedDiscs, serializedDisc)
	}
	return serializedDiscs, nil
}

func (s *serializer) serializeDisc(disc *scene.Disc) (*Disc, error) {
	materialIndex, err := s.materialIndex(disc.Material)
	if err != nil {
		return nil, err
	}

	return &Disc{
		Name:     disc.Name,
		Origin:   s.vectorIndex(disc.Origin),
		Normal:   s.vectorIndex(disc.Normal),
		Radius:   disc.Radius,
		Material: materialIndex,
	}, nil
}

func (s *serializer) serializeProjection(projection *scene.ImageProjection) (*Projection, error) {
	if projection == nil {
		return nil, nil
	}

	resourceIndex, err := s.fileResourceIndex(projection.Image)
	if err != nil {
		return nil, err
	}

	return &Projection{
		ProjectionType: string(projection.ProjectionType),
		Image:          resourceIndex,
		Origin:         s.vectorIndex(projection.Origin),
		U:              s.vectorIndex(projection.U),
		V:              s.vectorIndex(projection.V),
		RepeatU:        projection.RepeatU,
		RepeatV:        projection.RepeatV,
		FlipU:          projection.FlipU,
		FlipV:          projection.FlipV,
	}, nil
}
