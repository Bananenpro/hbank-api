package services

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

var supportedTranslations []string

var translations map[string]map[string]string

func Tr(text string, lang string) string {
	langTranslations, ok := translations[lang]
	if ok {
		translation, ok := langTranslations[text]
		if ok {
			return translation
		} else {
			log.Printf("Missing translation of string '%s' for language '%s'", text, lang)
		}
	}

	return text
}

func LoadTranslations() error {
	files, err := os.ReadDir("translations/")
	if err != nil {
		log.Println("Couldn't find translation files:", err)
		return err
	}

	supportedTranslations = make([]string, 0, len(files))
	translations = make(map[string]map[string]string, len(files))

	for _, f := range files {
		bytes, err := os.ReadFile("translations/" + f.Name())
		if err != nil {
			log.Printf("Couldn't open translation file '%s': %s", "translations/"+f.Name(), err)
			continue
		}

		content := string(bytes)

		lang, err := parseTranslationFile(content)
		if err != nil {
			log.Printf("Error while parsing '%s': %s", f.Name(), err.Error())
			continue
		}
		supportedTranslations = append(supportedTranslations, f.Name())
		translations[f.Name()] = lang
	}

	log.Println("Loaded translations: ", supportedTranslations)

	return nil
}

func parseTranslationFile(fileContent string) (map[string]string, error) {
	fileContent = strings.ReplaceAll(fileContent, "\r", "")

	lines := strings.Split(fileContent, "\n")
	lang := make(map[string]string, len(lines))
	for i, l := range lines {
		l := strings.TrimSpace(l)
		if len(l) == 0 {
			continue
		}
		parts := strings.Split(l, `"="`)
		if len(parts) != 2 || !strings.HasPrefix(parts[0], `"`) || !strings.HasSuffix(parts[1], `"`) {
			return nil, errors.New(fmt.Sprintf("Syntax error in line %d: '%s'", i, l))
		}

		lang[strings.TrimPrefix(parts[0], `"`)] = strings.TrimSuffix(parts[1], `"`)
	}

	return lang, nil
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
	for _, l := range supportedTranslations {
		if l == lang {
			return true
		}
	}
	return false
}
