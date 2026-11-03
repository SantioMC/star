package branding

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/tubblelive/star/internal/database"
)

func LogoURL(branding *database.Branding, serverURL string) *string {
	if branding.LogoUpdatedAt == nil {
		return nil
	}

	version := strconv.FormatInt(branding.LogoUpdatedAt.Unix(), 10)
	return new(strings.TrimSuffix(serverURL, "/") + "/api/branding/logo?v=" + version)
}

func FontURL(branding *database.Branding) string {
	query := url.Values{}
	query.Add("family", branding.HeadingFont)

	if branding.BodyFont != branding.HeadingFont {
		query.Add("family", branding.BodyFont)
	}

	query.Set("display", "swap")
	return "https://fonts.googleapis.com/css2?" + query.Encode()
}
