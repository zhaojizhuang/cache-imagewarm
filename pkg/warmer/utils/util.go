package utils

import "strings"

// ParseRegistry return the registry of image
func ParseRegistry(imageName string) string {
	idx := strings.Index(imageName, "/")
	if idx == -1 {
		return imageName
	}
	return imageName[0:idx]
}
