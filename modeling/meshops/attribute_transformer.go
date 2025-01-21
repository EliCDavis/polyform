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

func fallbackAttribute(attribute, fallback string) string {
	if strings.TrimSpace(attribute) == "" {
		return fallback
	}
	return attribute
}
