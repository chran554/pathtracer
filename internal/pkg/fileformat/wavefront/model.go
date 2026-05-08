package wavefront

type Wavefront struct {
	Facets []*Facet

	Materials       map[string]*Material
	Groups          map[string][]*Facet
	Objects         map[string][]*Facet
	SmoothingGroups map[int][]*Facet
}

type vec [3]float64

type Facet struct {
	Object                   string
	Groups                   []string
	SmoothingGroup           int
	Vertices                 []*vec
	VertexNormals            []*vec
	VertexTextureCoordinates []*vec // Possibility to have texture coordinates per vertex (u, v) with texture depth w.
	Material                 *Material
}

type Color struct {
	R, G, B float64
}

type Material struct {
	Name string

	// Specifies the specular exponent for the current material.
	// This defines the focus of the specular highlight exponent in a Blinn-Phong model.
	// A high value makes the material glossier/sharper.
	//
	// Syntax:
	//   Ns exponent
	//
	// "exponent" is the value for the specular exponent. A high exponent results in a tight, concentrated highlight.
	// Ns values normally range from 0 to 1000.
	//
	// This is most likely the exponent value for "Phong" highlight calculations or similar models.
	//
	// Do not confuse with the "Sharpness" parameter in the material definition.
	//
	// Blender software exports the "Roughness" material parameter as mtl-file parameter "Ns".
	Ns float64

	// Refl appears in Blender software export "Metallic" material parameter as mtl-file parameter "refl".
	// This is NOT part of the mtl-file specification as "refl" is _not_ supposed to be
	// used for scalar values but rather specify reflection maps.
	//
	// Syntax:
	//   Refl float64
	//
	// However, since that feature is very rarely used, and it is more likely to encounter a Blender exported file
	// we adhere to the format used by Blender.
	//
	// Blender software exports the "Metallic" material parameter as mtl-file parameter "refl".
	Refl float64

	// Ks specify the specular reflectivity of the current material
	//
	// Syntax:
	//   Ks r g b
	//   Ks spectral file.rfl factor
	//   Ks xyz x y z
	//
	// "Specularity / Glossiness" [0.0 .. 1.0]
	Ks *Color

	// Tf specify the transmission filter of the current material
	//
	// Syntax:
	//   Tf r g b
	//   Tf spectral file.rfl factor
	//   Tf xyz x y z
	//
	// Any light passing through the object is filtered by the transmission
	// filter, which only allows the specific colors to pass through.
	// For example, Tf 0 1 0 allows all the green to pass through and
	// filters out all the red and blue.
	Tf *Color

	// Ke is a parameter for "emission"
	// "Emission" [0.0 .. 1.0], [0.0 .. 1.0], [0.0 .. 1.0]
	//
	// The Ke parameter is widely used in mtl-files and is a later de-facto addition to the format
	// but not present in mtl-file specification.
	//
	// Syntax:
	//   Ke r g b
	Ke *Color

	// Sharpness specifies the sharpness of the reflections from the local reflection map.
	//
	// Syntax:
	//   Sharpness value
	//
	// Do not confuse with the "Ns" parameter in the material definition.
	//
	// If a material does not have a local reflection map defined in its
	// material definition, sharpness will apply to the global reflection map
	// defined in PreView.
	//
	// "value" can be a number from 0 to 1000.  The default is 60.  A high
	// value results in a clear reflection of objects in the reflection map.
	Sharpness float64

	// Ni is optical density (index of refraction)
	//
	// Syntax:
	//   Ni float
	//
	// Specifies the optical density for the surface.
	Ni float64

	// Proprietary parameter for "roughness" (not present in mtl-file specification)
	// "Roughness" [0.0 .. 1.0]
	Pr float64

	// Illum statement specifies the illumination model to use in the material.
	//
	// Syntax:
	//   illum illum_#
	//
	// Illumination models are mathematical equations that represent
	// various material lighting and shading effects.
	//
	//   Illumination   Properties that are turned on in the
	//   model          Property Editor
	//
	//   0              Color on and Ambient off
	//   1              Color on and Ambient on
	//   2              Highlight on
	//   3              Reflection on and Ray trace on
	//   4              Transparency: Glass on, Reflection: Ray trace on
	//   5              Reflection: Fresnel on and Ray trace on
	//   6              Transparency: Refraction on, Reflection: Fresnel off and Ray trace on
	//   7              Transparency: Refraction on, Reflection: Fresnel on and Ray trace on
	//   8              Reflection on and Ray trace off
	//   9              Transparency: Glass on, Reflection: Ray trace off
	//   10             Casts shadows onto invisible surfaces
	Illum int

	// D specifies the dissolve for the current material.
	//
	// Syntax:
	//   d factor
	//
	// "factor" is the amount this material dissolves into the background.
	//
	// A factor of 1.0 is fully opaque.  This is the default when a new material is created.
	//
	// A factor of 0.0 is fully dissolved (completely transparent).
	//
	// Unlike a real transparent material, the dissolve does not depend upon
	// material thickness nor does it have any spectral character.
	// Dissolve works on all illumination models.
	D float64

	// Ka specify the ambient reflectivity of the current material
	//
	// Syntax:
	//   Ka r g b
	//   Ka spectral file.rfl factor
	//   Ka xyz x y z
	//
	// "Ambient color" [[0.0 .. 1.0] [0.0 .. 1.0] [0.0 .. 1.0]]
	Ka *Color

	// Kd specify the diffuse reflectivity of the current material
	//
	// Syntax:
	//   Kd r g b
	//   Kd spectral file.rfl factor
	//   Kd xyz x y z
	//
	// "Diffuse color" [[0.0 .. 1.0] [0.0 .. 1.0] [0.0 .. 1.0]]
	Kd *Color

	// MapKd specify the diffuse reflectivity of the current material.
	// The diffuse reflectivity is a texture map (image file).
	//
	// Syntax:
	//   map_Kd filename
	MapKd string

	// MapKe specify the emission of the current material.
	// The emission is a texture map (image file).
	//
	// Syntax:
	//   map_Ke filename
	MapKe string

	// MapPr specify the roughness of the current material.
	// The roughness is a texture map (image file).
	//
	// Syntax:
	//   map_Pr filename
	MapPr string

	// MapNs specify the shininess specular-exponent map in a Blinn-Phong model.
	// A bright map_Ns pixel means glossier/sharper of the current material.
	//
	// Syntax:
	//   map_Ns filename
	MapNs string

	// MapBump specify the emission of the current material.
	// The emission is a texture map (image file).
	//
	// Syntax:
	//   map_Bump filename
	MapBump string

	// MapBasePath is the base file path for map_Kx parameters.
	//
	// This path is the base path of the material file.
	// The MapKx image files are relative to this base path.
	MapBasePath string
}
