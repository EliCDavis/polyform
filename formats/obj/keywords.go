package obj

type Keyword string

const (
	Vertex            Keyword = "v"
	TextureCoordinate Keyword = "vt"
	Normal            Keyword = "vn"
	Face              Keyword = "f"
	Group             Keyword = "g"
	ObjectName        Keyword = "o"
	MaterialUsage     Keyword = "usemtl"
	MaterialLibrary   Keyword = "mtllib"
	SmoothingGroup    Keyword = "s"
	Comment           Keyword = "#"
)
