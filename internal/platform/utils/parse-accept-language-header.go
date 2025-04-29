package utils

func ParseAcceptLanguageHeader(header string) string {
	var parsedLocale string

	if header == "" {
		parsedLocale = "en"
	} else {
		var err error
		parsedLocale, err = ParsePrimaryLanguage(header)
		if err != nil {
			parsedLocale = "en"
		}
	}

	return parsedLocale
}
