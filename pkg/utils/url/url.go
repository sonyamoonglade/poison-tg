package url

import "regexp"

var r = regexp.MustCompile(`(?m)^https:\/\/dw4\.co\/.*`)

func IsValidDW4URL(url string) bool {
	return r.MatchString(url)
}
