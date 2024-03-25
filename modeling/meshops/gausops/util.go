package gausops

import "strings"

func getAttribute(attr string, fallback string) string {
	if strings.TrimSpace(attr) == "" {
		return fallback
	}
	return attr
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
