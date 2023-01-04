package generator

import (
	"fmt"
	"math/rand"
)

func NewCategoryParameter(categories []string) CategoryParameter {
	return CategoryParameter{
		categories: categories,
		setValue:   -1,
	}
}

type CategoryParameter struct {
	categories []string
	setValue   int
}

func (cp CategoryParameter) Value() string {
	if cp.setValue > -1 {
		return cp.categories[cp.setValue]
	}
	return cp.categories[rand.Intn(len(cp.categories))]
}

func (cp CategoryParameter) IsSet() bool {
	return cp.setValue > -1
}

func (cp *CategoryParameter) Set(category int) {
	if category >= len(cp.categories) || category < 0 {
		panic(fmt.Errorf("invalid category index: %d", category))
	}
	cp.setValue = category
}
