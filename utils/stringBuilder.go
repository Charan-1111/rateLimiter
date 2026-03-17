package utils

import "strings"

func StringBuilder(rateLimit, algo, scope, identifier string) string {
	var sb strings.Builder

	sb.WriteString(rateLimit)
	sb.WriteString(":")
	sb.WriteString(algo)
	sb.WriteString(":")
	sb.WriteString(scope)
	sb.WriteString(":")
	sb.WriteString(identifier)

	return sb.String()
}
