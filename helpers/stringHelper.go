package helpers

import "strings"

func RemoveSpacesBeforAndAfter(content string) string {
	return strings.TrimSpace(content)
}
