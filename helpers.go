package main

import "strings"

func extractParenthesis(full string) string {
	full = strings.Trim(full, " ")
	parts := strings.Split(full, "(")
	if len(parts) != 2 {
		return full
	}
	return strings.TrimRight(parts[1], ")")
}
