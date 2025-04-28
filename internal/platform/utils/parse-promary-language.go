package utils

import (
	"fmt"
	"golang.org/x/text/language"
)

func ParsePrimaryLanguage(acceptLang string) (string, error) {
	tags, _, err := language.ParseAcceptLanguage(acceptLang)
	if err != nil {
		return "", err
	}
	if len(tags) == 0 {
		return "", fmt.Errorf("no languages found")
	}
	base, _ := tags[0].Base()
	return base.String(), nil
}
