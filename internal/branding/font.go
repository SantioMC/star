package branding

import "regexp"

var fontPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9 ]{0,63}$`)

func ValidFont(name string) bool {
	return fontPattern.MatchString(name)
}
