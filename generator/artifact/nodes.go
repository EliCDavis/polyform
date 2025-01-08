package artifact

import "github.com/EliCDavis/polyform/refutil"

func Nodes() *refutil.TypeFactory {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[ImageNode](factory)

	refutil.RegisterType[GltfArtifact](factory)
	refutil.RegisterType[GltfMaterialAnisotropyExtensionNode](factory)
	refutil.RegisterType[GltfMaterialClearcoatExtensionNode](factory)
	refutil.RegisterType[GltfMaterialNode](factory)
	refutil.RegisterType[GltfMaterialTransmissionExtensionNode](factory)
	refutil.RegisterType[GltfMaterialVolumeExtensionNode](factory)
	refutil.RegisterType[GltfModel](factory)

	return factory
}
