package util

import (
	"regexp"
	"strings"
)

func CreateSlug(name string) string {
	name = strings.ToLower(name)

	name = strings.ReplaceAll(name, " ", "-")

	reg := regexp.MustCompile(`[^a-z0-9\-]+`)
	name = reg.ReplaceAllString(name, "")

	return name
}
