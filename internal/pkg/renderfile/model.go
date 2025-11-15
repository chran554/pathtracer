package renderfile

import (
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
)

type AnimationInformation struct {
	Name               string              `json:"name"`
	Width              int                 `json:"width"`
	Height             int                 `json:"height"`
	WriteRawImageFile  bool                `json:"write-raw-image-file"`
	WriteImageInfoFile bool                `json:"write-image-info-file"`
	FramesInformation  []*FrameInformation `json:"framesinformation"`
}

type FrameInformation struct {
	Index        int    `json:"index"`
	Filename     string `json:"filename"`
	FrameFile    string `json:"frame-file"`
	SceneFile    string `json:"scene-file,omitempty"`
	VectorFile   string `json:"vector-file,omitempty"`
	Vector2DFile string `json:"vector2d-file,omitempty"`
	MaterialFile string `json:"material-file,omitempty"`
	ColorFile    string `json:"color-file,omitempty"`
	// Camera       *Camera `json:"camera"`
}

type Camera struct {
	Origin            VectorIndex   `msagpack:"origin"`
	Heading           VectorIndex   `msagpack:"heading"`
	ViewUp            VectorIndex   `msagpack:"view-up"`
	ViewPlaneDistance float64       `msagpack:"view-plane-distance"` // ViewPlaneDistance determine the focal length, the view angle of the camera.
	ApertureSize      float64       `msagpack:"aperture-size"`       // ApertureSize is the size of the aperture opening. The wider the aperture the less focus depth. Value 0.0 is infinite focus depth.
	ApertureShape     ResourceIndex `msagpack:"aperture-shape"`      // ApertureShape is the file path to a black and white image where white define the aperture shape. Aperture size determine the size of the longest side (width or height) of the image. If empty string then a default round aperture shape is used.
	FocusDistance     float64       `msagpack:"focus-distance"`
	Samples           int           `msagpack:"samples"`
	AntiAlias         bool          `msagpack:"anti-alias"`
	Magnification     float64       `msagpack:"magnification"`
	RenderType        string        `msagpack:"render-type"`
	RecursionDepth    int           `msagpack:"recursion-depth"`
}

type Frame struct {
	Index     int        `msgpack:"index"`
	Filename  string     `msgpack:"filename"`
	SceneNode *SceneNode `msgpack:"scene-node"`
	Camera    *Camera    `msgpack:"camera"`
}

type SceneNode struct {
	Spheres         []*Sphere         `msgpack:"spheres,omitempty"`
	Discs           []*Disc           `msgpack:"discs,omitempty"`
	ChildNodes      []*SceneNode      `msgpack:"child-nodes,omitempty"`
	FacetStructures []*FacetStructure `msgpack:"facet-structures,omitempty"`
}

type FacetStructure struct {
	Name             string            `msgpack:"name,omitempty"`
	SubstructureName string            `msgpack:"substructure-name,omitempty"`
	Material         MaterialIndex     `msgpack:"material,omitempty"`
	Facets           []*Facet          `msgpack:"facets,omitempty"`
	FacetStructures  []*FacetStructure `msgpack:"facet-structures,omitempty"`
	IgnoreBounds     bool              `msgpack:"ignore-bounds,omitempty"`
}

type Facet struct {
	Vertices           []VectorIndex   `msgpack:"vertices"`
	VertexNormals      []VectorIndex   `msgpack:"vertex-normals,omitempty"`
	TextureCoordinates []Vector2DIndex `msgpack:"texture-coordinates,omitempty"`
}

type Sphere struct {
	Name     string        `msgpack:"name"`
	Origin   VectorIndex   `msgpack:"origin"`
	Radius   float64       `msgpack:"radius"`
	Material MaterialIndex `msgpack:"material"`
}

type Disc struct {
	Name     string        `msgpack:"name"`
	Origin   VectorIndex   `msgpack:"origin"`
	Normal   VectorIndex   `msgpack:"normal"`
	Radius   float64       `msgpack:"radius"`
	Material MaterialIndex `msgpack:"material"`
}

type ColorIndex uint
type ResourceIndex uint
type VectorIndex uint
type Vector2DIndex uint
type MaterialIndex uint

type Vector struct {
	X float64 `msgpack:"x"`
	Y float64 `msgpack:"y"`
	Z float64 `msgpack:"z"`
}

type Vector2D struct {
	X float64 `msgpack:"x"`
	Y float64 `msgpack:"y"`
}

type Color struct {
	R float32 `msgpack:"r"`
	G float32 `msgpack:"g"`
	B float32 `msgpack:"b"`
	A float32 `msgpack:"a"` // Alpha channel
}

type Projection struct {
	ProjectionType string        `msgpack:"projection-type"`
	Image          ResourceIndex `msgpack:"image-resource-index"`
	Origin         VectorIndex   `msgpack:"origin"`
	U              VectorIndex   `msgpack:"u"`
	V              VectorIndex   `msgpack:"v"`
	RepeatU        bool          `msgpack:"repeat-u,omitempty"`
	RepeatV        bool          `msgpack:"repeat-v,omitempty"`
	FlipU          bool          `msgpack:"flip-u,omitempty"`
	FlipV          bool          `msgpack:"flip-v,omitempty"`
}

type Material struct {
	Name            string      `msgpack:"name,omitempty"`
	Color           ColorIndex  `msgpack:"color,omitempty"`
	Diffuse         float64     `msgpack:"diffuse,omitempty"`
	Emission        ColorIndex  `msgpack:"emission,omitempty"`
	Glossiness      float64     `msgpack:"glossiness,omitempty"` // Glossiness is the percent amount that will make out specular reflection. Values [0.0 .. 1.0] with default 0.0. Lower value the more diffuse color will appear and higher value the more mirror reflection will appear.
	Roughness       float64     `msgpack:"roughness,omitempty"`  // Roughness is the diffuse spread of the specular reflection. Values [0.0 .. 1.0] with default 0.0. Lower is like "brushed metal" or "foggy/hazy reflection" and higher value give a more mirror like reflection. A value of 1.0 is perfect mirror reflection and a value of 0.0 is a perfect diffuse material (no mirror at al).
	RefractionIndex float64     `msgpack:"refraction-index,omitempty"`
	SolidObject     bool        `msgpack:"solid-object,omitempty"`   // SolidObject is if the material denotes a solid object with volume, not a hollow or open object or object nor an object with plane-thin walls. Solid transparent objects can refract light, hollow objects don't.
	Transparency    float64     `msgpack:"transparency,omitempty"`   // Transparency is the amount [0,1.0) of transparency vs diffuse contribution.
	RayTerminator   bool        `msgpack:"ray-terminator,omitempty"` // RayTerminator decide if the ray should terminate after hit with object. Example can be an environment sphere or environment cube where a hit to the wall is the same as "no hit, continue in infinity". Extremely bright lights can also be ray terminators, their appearance will not notably be affected by further tracing.
	Projection      *Projection `msgpack:"projection,omitempty"`
}

func (v *Vector) MarshalMsgpack() ([]byte, error) {
	return msgpack.Marshal([]float64{v.X, v.Y, v.Z})
}

func (v *Vector) UnmarshalMsgpack(b []byte) error {
	var arr []float64
	if err := msgpack.Unmarshal(b, &arr); err != nil {
		return err
	}
	if len(arr) != 3 {
		return fmt.Errorf("vector decode: expected 3 values, got %d", len(arr))
	}
	v.X, v.Y, v.Z = arr[0], arr[1], arr[2]
	return nil
}

func (v *Vector2D) MarshalMsgpack() ([]byte, error) {
	return msgpack.Marshal([]float64{v.X, v.Y})
}

func (v *Vector2D) UnmarshalMsgpack(b []byte) error {
	var arr []float64
	if err := msgpack.Unmarshal(b, &arr); err != nil {
		return err
	}
	if len(arr) != 2 {
		return fmt.Errorf("vector2d decode: expected 2 values, got %d", len(arr))
	}
	v.X, v.Y = arr[0], arr[1]
	return nil
}

// MarshalMsgpack implements the msgpack.Marshaler interface for custom serialization of Color.
func (c *Color) MarshalMsgpack() ([]byte, error) {
	return msgpack.Marshal([]float32{c.R, c.G, c.B, c.A})
}

// UnmarshalMsgpack implements the msgpack.Unmarshaler interface for custom deserialization of Color.
func (c *Color) UnmarshalMsgpack(data []byte) error {
	var arr []float32
	if err := msgpack.Unmarshal(data, &arr); err != nil {
		return err
	}

	if len(arr) != 4 {
		return fmt.Errorf("color decode: expected 4 values, got %d", len(arr))
	}

	c.R, c.G, c.B, c.A = arr[0], arr[1], arr[2], arr[3]
	return nil
}
