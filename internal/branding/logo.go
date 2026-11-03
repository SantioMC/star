package branding

import (
	"net/http"
	"slices"
)

func DetectLogoType(data []byte) (string, bool) {
	contentType := http.DetectContentType(data)
	allowed := []string{"image/png", "image/jpeg", "image/webp"}

	return contentType, slices.Contains(allowed, contentType)
}
