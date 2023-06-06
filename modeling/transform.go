package modeling

type Transformer interface {
	Transform(m Mesh) (Mesh, error)
}
