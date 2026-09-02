package helpers

import (
	"strings"
	"unicode"
)

func NormalizarTextoPermiso(texto string) string {
	texto = strings.ToUpper(strings.TrimSpace(texto))
	var builder strings.Builder
	anteriorSeparador := false
	for _, r := range texto {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
			anteriorSeparador = false
		} else if !anteriorSeparador {
			builder.WriteRune('_')
			anteriorSeparador = true
		}
	}
	return strings.Trim(builder.String(), "_")
}
