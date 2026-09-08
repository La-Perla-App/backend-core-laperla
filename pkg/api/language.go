package api

import (
	"sort"
	"strconv"
	"strings"
)

var SupportedLanguages = []string{
	"en",
	"es",
}

type Language struct {
	Tag      string
	ShortTag string
	Q        float64
	Original string
}

func ParseAcceptLanguage(header string) []Language {
	var langs []Language
	parts := strings.Split(header, ",")

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		langParts := strings.Split(trimmed, ";")
		tag := strings.TrimSpace(langParts[0])
		tagParts := strings.Split(tag, "-")
		shortTag := strings.TrimSpace(tagParts[0])

		// Q-value por defecto es 1
		q := 1.0
		if len(langParts) > 1 {
			qParam := strings.TrimSpace(langParts[1])
			if strings.HasPrefix(qParam, "q=") {
				qValueStr := qParam[2:]
				if qValue, err := strconv.ParseFloat(qValueStr, 64); err == nil {
					q = qValue
				}
			}
		}

		langs = append(langs, Language{
			Tag:      tag,
			ShortTag: shortTag,
			Q:        q,
			Original: trimmed,
		})
	}

	// Ordenar por Q descendente (y luego por posición original si hay empate)
	sort.Slice(langs, func(i, j int) bool {
		if langs[i].Q == langs[j].Q {
			return i < j // Mantener orden original si la q es igual
		}
		return langs[i].Q > langs[j].Q
	})

	return langs
}
