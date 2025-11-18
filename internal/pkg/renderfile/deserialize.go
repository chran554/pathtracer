package renderfile

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"pathtracer/internal/pkg/color"
	"pathtracer/internal/pkg/scene"
	"regexp"

	"github.com/ungerik/go3d/float64/vec2"
	"github.com/ungerik/go3d/float64/vec3"
	"github.com/vmihailenco/msgpack/v5"
)

func ReadRenderFile(renderFilename string) (*scene.Animation, error) {
	zipReader, err := zip.OpenReader(renderFilename)
	if err != nil {
		return nil, fmt.Errorf("could not open zip file reader: %w", err)
	}
	defer zipReader.Close()

	s, err := newDeserializer(&zipReader.Reader)
	if err != nil {
		return nil, err
	}

	animationInformation, err := s.deserializeAnimationInformation()
	if err != nil {
		return nil, err
	}

	animation, err := s.deserializeAnimation(animationInformation)
	if err != nil {
		return nil, err
	}

	return animation, nil
}

func (s *serializer) deserializeAnimationInformation() (*AnimationInformation, error) {
	for _, file := range s.zipReader.File {
		if match, _ := regexp.Match("animation\\.json", []byte(file.Name)); match {
			fileData, err := readZipFileEntry(file)
			if err != nil {
				return nil, err
			}

			var animationInformation = &AnimationInformation{}
			err = json.Unmarshal(fileData, animationInformation)
			if err != nil {
				return nil, err
			}

			return animationInformation, nil
		}
	}

	return nil, fmt.Errorf("no animation file found")
}

func (s *serializer) deserializeAnimation(animationInformation *AnimationInformation) (*scene.Animation, error) {
	var animation *scene.Animation

	animation = &scene.Animation{
		AnimationName:      animationInformation.Name,
		Width:              animationInformation.Width,
		Height:             animationInformation.Height,
		WriteRawImageFile:  animationInformation.WriteRawImageFile,
		WriteImageInfoFile: animationInformation.WriteImageInfoFile,
	}

	for _, frameInformation := range animationInformation.FramesInformation {
		frame, err := s.deserializeFrame(frameInformation)
		if err != nil {
			return nil, err
		}

		animation.Frames = append(animation.Frames, frame)
	}

	return animation, nil
}

func (s *serializer) deserializeFrame(frameInformation *FrameInformation) (*scene.Frame, error) {
	err := s.initFrameCache(frameInformation)
	if err != nil {
		return nil, err
	}

	for _, file := range s.zipReader.File {
		if file.Name == frameInformation.FrameFile {
			fmt.Println("Initializing: Reading ", file.Name)
			fileData, err := readZipFileEntry(file)
			if err != nil {
				return nil, err
			}

			var frame = &Frame{}
			err = msgpack.Unmarshal(fileData, &frame)
			if err != nil {
				return nil, fmt.Errorf("could not unmarshal frame from file %s: %w", file.Name, err)
			}

			camera, err := s.deserializeCamera(frame.Camera)
			if err != nil {
				return nil, err
			}

			sceneNode := s.deserializeSceneNode(frame.SceneNode)

			return &scene.Frame{
				Filename:  frame.Filename,
				Index:     frame.Index,
				Camera:    camera,
				SceneNode: sceneNode,
			}, nil
		}
	}

	return nil, fmt.Errorf("could not find frame file %s", frameInformation.FrameFile)
}

func (s *serializer) initFrameCache(frameInformation *FrameInformation) error {
	s.clearFrameCache()

	for _, file := range s.zipReader.File {
		if file.Name == frameInformation.VectorFile {
			fmt.Println("Initializing: Reading ", file.Name)
			fileData, err := readZipFileEntry(file)
			if err != nil {
				return err
			}

			var vectors []Vector
			err = msgpack.Unmarshal(fileData, &vectors)
			if err != nil {
				return fmt.Errorf("could not unmarshal vectors from file %s: %w", file.Name, err)
			}

			for _, vector := range vectors {
				s.sv = append(s.sv, &vec3.T{vector.X, vector.Y, vector.Z})
			}
		}
		if file.Name == frameInformation.Vector2DFile {
			fmt.Println("Initializing: Reading ", file.Name)
			fileData, err := readZipFileEntry(file)
			if err != nil {
				return err
			}

			var vectors []Vector2D
			err = msgpack.Unmarshal(fileData, &vectors)
			if err != nil {
				return fmt.Errorf("could not unmarshal 2D vectors from file %s: %w", file.Name, err)
			}

			for _, vector := range vectors {
				s.sv2d = append(s.sv2d, &vec2.T{vector.X, vector.Y})
			}
		}
		if file.Name == frameInformation.ColorFile {
			fmt.Println("Initializing: Reading ", file.Name)
			fileData, err := readZipFileEntry(file)
			if err != nil {
				return err
			}

			var colors []Color
			err = msgpack.Unmarshal(fileData, &colors)
			if err != nil {
				return fmt.Errorf("could not unmarshal colors from file %s: %w", file.Name, err)
			}

			for _, c := range colors {
				s.sc = append(s.sc, &color.Color{R: c.R, G: c.G, B: c.B, A: c.A})
			}
		}
	}
	for _, file := range s.zipReader.File {
		if file.Name == frameInformation.MaterialFile {
			fmt.Println("Initializing: Reading ", file.Name)
			fileData, err := readZipFileEntry(file)
			if err != nil {
				return err
			}

			var materials []Material
			err = msgpack.Unmarshal(fileData, &materials)
			if err != nil {
				return fmt.Errorf("could not unmarshal materials from file %s: %w", file.Name, err)
			}

			for _, m := range materials {
				projection, err := s.deserializeProjection(m.Projection)
				if err != nil {
					return err
				}

				s.sm = append(s.sm, &scene.Material{
					Name:            m.Name,
					Color:           s.sceneColor(m.Color),
					Diffuse:         m.Diffuse,
					Emission:        s.sceneColor(m.Emission),
					Glossiness:      m.Glossiness,
					Roughness:       m.Roughness,
					Projection:      projection,
					RefractionIndex: m.RefractionIndex,
					SolidObject:     m.SolidObject,
					Transparency:    m.Transparency,
					RayTerminator:   m.RayTerminator,
				})
			}
		}
	}

	return nil
}

func (s *serializer) deserializeCamera(camera *Camera) (*scene.Camera, error) {
	apertureImage, err := s.resourceImage(camera.ApertureShape)
	if err != nil {
		return nil, err
	}

	return &scene.Camera{
		Origin:            s.sceneVector(camera.Origin),
		Heading:           s.sceneVector(camera.Heading),
		ViewUp:            s.sceneVector(camera.ViewUp),
		ViewPlaneDistance: camera.ViewPlaneDistance,
		ApertureSize:      camera.ApertureSize,
		ApertureShape:     apertureImage,
		FocusDistance:     camera.FocusDistance,
		Samples:           camera.Samples,
		AntiAlias:         camera.AntiAlias,
		Magnification:     camera.Magnification,
		RenderType:        scene.RenderType(camera.RenderType),
		RecursionDepth:    camera.RecursionDepth,
	}, nil
}

func (s *serializer) deserializeSceneFile(sceneFilename string) (*scene.SceneNode, error) {
	for _, file := range s.zipReader.File {
		if match, _ := regexp.Match(sceneFilename, []byte(file.Name)); match {
			fileData, err := readZipFileEntry(file)
			if err != nil {
				return nil, err
			}

			var sceneNode SceneNode
			err = msgpack.Unmarshal(fileData, &sceneNode)
			if err != nil {
				return nil, fmt.Errorf("could not unmarshal scene from file %s: %w", file.Name, err)
			}

			return s.deserializeSceneNode(&sceneNode), nil
		}
	}

	return nil, fmt.Errorf("could not find scene file %s", sceneFilename)
}

func (s *serializer) deserializeSceneNode(sceneNode *SceneNode) *scene.SceneNode {
	return &scene.SceneNode{
		Spheres:         s.deserializeSpheres(sceneNode.Spheres),
		Discs:           s.deserializeDiscs(sceneNode.Discs),
		ChildNodes:      s.deserializeSceneNodes(sceneNode.ChildNodes),
		FacetStructures: s.deserializeFacetStructures(sceneNode.FacetStructures),
		//Bounds:          nil,
	}
}

func (s *serializer) deserializeSceneNodes(sceneNodes []*SceneNode) []*scene.SceneNode {
	var nodes []*scene.SceneNode
	for _, sceneNode := range sceneNodes {
		nodes = append(nodes, s.deserializeSceneNode(sceneNode))
	}

	return nodes
}

func (s *serializer) deserializeFacetStructures(facetStructures []*FacetStructure) []*scene.FacetStructure {
	var structures []*scene.FacetStructure
	for _, structure := range facetStructures {
		structures = append(structures, &scene.FacetStructure{
			Name:             structure.Name,
			SubstructureName: structure.SubstructureName,
			Material:         s.sceneMaterial(structure.Material),
			Facets:           s.deserializeFacets(structure.Facets),
			FacetStructures:  s.deserializeFacetStructures(structure.FacetStructures),
			IgnoreBounds:     structure.IgnoreBounds,
			//Bounds:           nil,
		})
	}
	return structures
}

func (s *serializer) deserializeFacets(facets []*Facet) []*scene.Facet {
	var sceneFacets []*scene.Facet
	for _, facet := range facets {
		sceneFacets = append(sceneFacets, &scene.Facet{
			Vertices:           s.sceneVectors(facet.Vertices),
			VertexNormals:      s.sceneVectors(facet.VertexNormals),
			TextureCoordinates: s.sceneVectors2D(facet.TextureCoordinates),
			//Normal:             nil,
			//Bounds:             nil,
		})
	}
	return sceneFacets
}

func (s *serializer) deserializeDiscs(discs []*Disc) []*scene.Disc {
	var sceneDiscs []*scene.Disc
	for _, disc := range discs {
		sceneDiscs = append(sceneDiscs, &scene.Disc{
			Name:     disc.Name,
			Origin:   s.sceneVector(disc.Origin),
			Normal:   s.sceneVector(disc.Normal),
			Radius:   disc.Radius,
			Material: s.sceneMaterial(disc.Material),
		})
	}
	return sceneDiscs
}

func (s *serializer) deserializeSpheres(spheres []*Sphere) []*scene.Sphere {
	var sceneSpheres []*scene.Sphere
	for _, sphere := range spheres {
		sceneSpheres = append(sceneSpheres, &scene.Sphere{
			Name:     sphere.Name,
			Origin:   s.sceneVector(sphere.Origin),
			Radius:   sphere.Radius,
			Material: s.sceneMaterial(sphere.Material),
		})
	}
	return sceneSpheres
}

func (s *serializer) deserializeProjection(projection *Projection) (*scene.ImageProjection, error) {
	if projection == nil {
		return nil, nil
	}

	img, err := s.resourceImage(projection.Image)
	if err != nil {
		return nil, err
	}

	return &scene.ImageProjection{
		ProjectionType: scene.ProjectionType(projection.ProjectionType),
		Image:          img,
		Origin:         s.sceneVector(projection.Origin),
		U:              s.sceneVector(projection.U),
		V:              s.sceneVector(projection.V),
		RepeatU:        projection.RepeatU,
		RepeatV:        projection.RepeatV,
		FlipU:          projection.FlipU,
		FlipV:          projection.FlipV,
	}, nil
}
