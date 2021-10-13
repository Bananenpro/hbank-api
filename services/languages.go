package services

import (
	"strconv"
	"strings"
)

var supportedLanguages = []string{
	"en", // English
	"de", // German
}

func GetLanguageFromAcceptLanguageHeader(headerValue string) string {
	lang := "en"
	quality := float64(0)

	strs := strings.Split(headerValue, ",")
	for _, s := range strs {
		parts := strings.Split(s, ";")
		q := float64(1)
		if len(parts) > 1 {
			qStr := parts[1]
			qStr = strings.ReplaceAll(qStr, "q", "")
			qStr = strings.ReplaceAll(qStr, "=", "")
			qStr = strings.TrimSpace(qStr)
			_, err := strconv.ParseFloat(qStr, 64)
			if err == nil {
				q, _ = strconv.ParseFloat(qStr, 64)
			}
		}

		if q > quality {
			l := strings.TrimSpace(strings.Split(parts[0], "-")[0])
			if isSupportedLanguage(l) {
				lang = l
				quality = q
			}
		}
	}

	return lang
}

func isSupportedLanguage(lang string) bool {
	for _, l := range supportedLanguages {
		if l == lang {
			return true
		}
	}
	return false
}
