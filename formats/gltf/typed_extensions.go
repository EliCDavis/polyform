package gltf

// ExtensionID enumerates standard extensions
type ExtensionID string

const (
	KHRMaterialsPbrSpecularGlossiness ExtensionID = "KHR_materials_pbrSpecularGlossiness"
)

func (et ExtensionID) String() string {
	return string(et)
}

type Extension interface {
	ExtensionID() ExtensionID
	Equal(other Extension) bool
}
