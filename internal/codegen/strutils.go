package codegen

import (
	"regexp"
	"strings"
	"unicode"
)

// sanitizeLabel removes special characters, replaces spaces with underscores, and converts to lowercase.
func sanitizeLabel(input string) string {
	if input == "" {
		return "_"
	}
	reg, _ := regexp.Compile(`[^\w\s]`)
	input = reg.ReplaceAllString(input, "")
	input = strings.ReplaceAll(input, " ", "_")
	input = strings.ToLower(input)
	return input
}

// convertSchemaNameToInterfaceName removes special characters, converts to lowercase, capitalizes first letter of each word, and joins them.
func convertSchemaNameToInterfaceName(input string) string {
	reg, _ := regexp.Compile(`[^\w\s]`)
	input = reg.ReplaceAllString(input, "")
	input = strings.ToLower(input)

	words := strings.Split(input, "_")
	for i, word := range words {
		words[i] = strings.Title(word)
	}

	return strings.Join(words, "")
}

// prependUnderscoreToEnum adds underscore to the beginning if the string starts with a digit.
func prependUnderscoreToEnum(e string) string {
	reg, _ := regexp.Compile(`^(\d)`)
	return reg.ReplaceAllString(e, "_$1")
}

// convertLabelToEnumName removes special characters, converts to lowercase, capitalizes first letter of each word, and joins them.
func convertLabelToEnumName(input string) string {
	reg, _ := regexp.Compile(`[^\w\s]`)
	input = reg.ReplaceAllString(input, "")
	input = strings.ToLower(input)

	words := strings.Fields(input)
	for i, word := range words {
		words[i] = strings.Title(word)
	}

	return strings.Join(words, "")
}

// strings.Title is not directly analogous to TypeScript's `charAt(0).toUpperCase()` as it capitalizes each
// word in the string. We define a proper toFirstUpper function.
func toFirstUpper(word string) string {
	for i, v := range word {
		return string(unicode.ToUpper(v)) + word[i+1:]
	}
	return ""
}
