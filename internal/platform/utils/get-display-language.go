package utils

import (
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

func GetDisplayName(code string) string {
	tag := language.Make(code)
	return display.English.Tags().Name(tag)
}
