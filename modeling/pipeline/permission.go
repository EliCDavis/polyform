package pipeline

import "fmt"

type Permission interface {
	HasPrimitivePermission() bool
	HasAttributePermission() bool
	HasMaterialPermission() bool

	HasFloat3Permission(attr string) bool
	HasFloat2Permission(attr string) bool
	HasFloat1Permission(attr string) bool
}

type Resource struct {
	key string
}

func RequireMeshPrimitive() Resource {
	return Resource{
		key: "primitive",
	}
}

func RequireMeshFloat3Attribute(attribute string) Resource {
	return Resource{
		key: fmt.Sprintf("v3.%s", attribute),
	}
}

func RequireMeshFloat2Attribute(attribute string) Resource {
	return Resource{
		key: fmt.Sprintf("v2.%s", attribute),
	}
}

func RequireMeshFloat1Attribute(attribute string) Resource {
	return Resource{
		key: fmt.Sprintf("v1.%s", attribute),
	}
}

func PermissionForResources(resources ...Resource) Permission {
	return resourcePermission{}
}

type resourcePermission struct {
}

func (resourcePermission) HasAttributePermission() bool {
	return true
}

func (resourcePermission) HasFloat3Permission(attr string) bool {
	return true
}

func (resourcePermission) HasFloat2Permission(attr string) bool {
	return true
}

func (resourcePermission) HasFloat1Permission(attr string) bool {
	return true
}

func (resourcePermission) HasPrimitivePermission() bool {
	return true
}

func (resourcePermission) HasMaterialPermission() bool {
	return true
}

type everythingPermission struct {
}

func (everythingPermission) HasAttributePermission() bool {
	return true
}

func (everythingPermission) HasFloat3Permission(attr string) bool {
	return true
}

func (everythingPermission) HasFloat2Permission(attr string) bool {
	return true
}

func (everythingPermission) HasFloat1Permission(attr string) bool {
	return true
}

func (everythingPermission) HasPrimitivePermission() bool {
	return true
}

func (everythingPermission) HasMaterialPermission() bool {
	return true
}
