package meshops

import "strings"

type attributeTransformer interface {
	attribute() string
}

func getAttribute(at attributeTransformer, fallback string) string {
	attr := at.attribute()
	if strings.TrimSpace(attr) == "" {
		return fallback
	}
	return attr
}
