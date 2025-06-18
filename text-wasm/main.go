//go:build js && wasm

package main

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"net/mail"
	"regexp"
	"strings"
	"syscall/js"
	"unicode/utf8"
)

var silentMode = false

// Regular expressions for pattern extraction
var (
	emailRegex = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	urlRegex   = regexp.MustCompile(`https?://[^\s<>"{}|\\^` + "`" + `\[\]]+`)
	phoneRegex = regexp.MustCompile(`(?:\+\d{1,3}\s?)?\(?(\d{3})\)?[-.\s]?(\d{3})[-.\s]?(\d{4})|\+\d{1,3}[-.\s]?\d{1,4}[-.\s]?\d{1,4}[-.\s]?\d{1,9}`)
)

// Soundex mapping for phonetic matching
var soundexMap = map[rune]rune{
	'B': '1', 'F': '1', 'P': '1', 'V': '1',
	'C': '2', 'G': '2', 'J': '2', 'K': '2', 'Q': '2', 'S': '2', 'X': '2', 'Z': '2',
	'D': '3', 'T': '3',
	'L': '4',
	'M': '5', 'N': '5',
	'R': '6',
}

// Diacritics removal mapping
var diacriticsMap = map[rune]rune{
	'À': 'A', 'Á': 'A', 'Â': 'A', 'Ã': 'A', 'Ä': 'A', 'Å': 'A', 'Æ': 'A',
	'à': 'a', 'á': 'a', 'â': 'a', 'ã': 'a', 'ä': 'a', 'å': 'a', 'æ': 'a',
	'È': 'E', 'É': 'E', 'Ê': 'E', 'Ë': 'E',
	'è': 'e', 'é': 'e', 'ê': 'e', 'ë': 'e',
	'Ì': 'I', 'Í': 'I', 'Î': 'I', 'Ï': 'I',
	'ì': 'i', 'í': 'i', 'î': 'i', 'ï': 'i',
	'Ò': 'O', 'Ó': 'O', 'Ô': 'O', 'Õ': 'O', 'Ö': 'O', 'Ø': 'O',
	'ò': 'o', 'ó': 'o', 'ô': 'o', 'õ': 'o', 'ö': 'o', 'ø': 'o',
	'Ù': 'U', 'Ú': 'U', 'Û': 'U', 'Ü': 'U',
	'ù': 'u', 'ú': 'u', 'û': 'u', 'ü': 'u',
	'Ý': 'Y', 'Ÿ': 'Y', 'ý': 'y', 'ÿ': 'y',
	'Ñ': 'N', 'ñ': 'n',
	'Ç': 'C', 'ç': 'c',
}

// setSilentMode enables/disables silent mode for console logs
func setSilentMode(this js.Value, args []js.Value) interface{} {
	if len(args) == 1 {
		silentMode = args[0].Bool()
	}
	return js.ValueOf(silentMode)
}

// textSimilarity calculates similarity between two texts using Jaro-Winkler distance
func textSimilarity(this js.Value, args []js.Value) interface{} {
	if len(args) != 2 {
		return js.ValueOf("Error: two arguments required for textSimilarity")
	}

	s1 := args[0].String()
	s2 := args[1].String()

	if s1 == s2 {
		return js.ValueOf(1.0)
	}

	// Convert to lowercase for comparison
	s1 = strings.ToLower(s1)
	s2 = strings.ToLower(s2)

	// Calculate Jaro similarity
	jaro := jaroSimilarity(s1, s2)

	// Apply Winkler prefix bonus
	prefix := commonPrefixLength(s1, s2, 4)
	similarity := jaro + (0.1 * float64(prefix) * (1.0 - jaro))

	if !silentMode {
		fmt.Printf("Go WASM: Text similarity between '%s' and '%s' = %.3f\n", args[0].String(), args[1].String(), similarity)
	}

	return js.ValueOf(similarity)
}

// levenshteinDistance calculates the Levenshtein distance between two strings
func levenshteinDistance(this js.Value, args []js.Value) interface{} {
	if len(args) != 2 {
		return js.ValueOf("Error: two arguments required for levenshteinDistance")
	}

	s1 := args[0].String()
	s2 := args[1].String()

	if s1 == s2 {
		return js.ValueOf(0)
	}

	runes1 := []rune(s1)
	runes2 := []rune(s2)

	len1 := len(runes1)
	len2 := len(runes2)

	// Create matrix
	matrix := make([][]int, len1+1)
	for i := range matrix {
		matrix[i] = make([]int, len2+1)
	}

	// Initialize first row and column
	for i := 0; i <= len1; i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len2; j++ {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			cost := 0
			if runes1[i-1] != runes2[j-1] {
				cost = 1
			}

			matrix[i][j] = min(
				min(matrix[i-1][j]+1, matrix[i][j-1]+1), // deletion, insertion
				matrix[i-1][j-1]+cost,                   // substitution
			)
		}
	}

	distance := matrix[len1][len2]

	if !silentMode {
		fmt.Printf("Go WASM: Levenshtein distance between '%s' and '%s' = %d\n", s1, s2, distance)
	}

	return js.ValueOf(distance)
}

// soundex generates Soundex code for phonetic matching
func soundex(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for soundex")
	}

	str := strings.ToUpper(args[0].String())
	if len(str) == 0 {
		return js.ValueOf("")
	}

	// Start with the first letter
	result := string(str[0])

	// Convert remaining letters
	for i := 1; i < len(str) && len(result) < 4; i++ {
		char := rune(str[i])
		if code, exists := soundexMap[char]; exists {
			// Don't add if same as previous code
			lastCode := result[len(result)-1]
			if rune(lastCode) != code {
				result += string(code)
			}
		}
	}

	// Pad with zeros if needed
	for len(result) < 4 {
		result += "0"
	}

	if !silentMode {
		fmt.Printf("Go WASM: Soundex for '%s' = %s\n", args[0].String(), result)
	}

	return js.ValueOf(result)
}

// slugify converts a string to a URL-friendly slug
func slugify(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for slugify")
	}

	str := args[0].String()

	// Remove diacritics
	str = removeDiacriticsFromString(str)

	// Convert to lowercase
	str = strings.ToLower(str)

	// Replace non-alphanumeric with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	str = reg.ReplaceAllString(str, "-")

	// Remove leading/trailing hyphens
	str = strings.Trim(str, "-")

	if !silentMode {
		fmt.Printf("Go WASM: Slugified '%s' to '%s'\n", args[0].String(), str)
	}

	return js.ValueOf(str)
}

// camelCase converts string to camelCase
func camelCase(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for camelCase")
	}

	str := args[0].String()
	words := regexp.MustCompile(`\W+`).Split(str, -1)

	var result strings.Builder
	for i, word := range words {
		if word == "" {
			continue
		}
		if i == 0 {
			result.WriteString(strings.ToLower(word))
		} else {
			result.WriteString(strings.Title(strings.ToLower(word)))
		}
	}

	resultStr := result.String()

	if !silentMode {
		fmt.Printf("Go WASM: Converted '%s' to camelCase: '%s'\n", str, resultStr)
	}

	return js.ValueOf(resultStr)
}

// kebabCase converts string to kebab-case
func kebabCase(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for kebabCase")
	}

	str := args[0].String()

	// Handle camelCase/PascalCase
	reg := regexp.MustCompile(`([a-z])([A-Z])`)
	str = reg.ReplaceAllString(str, "$1-$2")

	// Replace non-alphanumeric with hyphens
	reg = regexp.MustCompile(`[^a-zA-Z0-9]+`)
	str = reg.ReplaceAllString(str, "-")

	// Convert to lowercase and trim
	str = strings.ToLower(strings.Trim(str, "-"))

	if !silentMode {
		fmt.Printf("Go WASM: Converted '%s' to kebab-case: '%s'\n", args[0].String(), str)
	}

	return js.ValueOf(str)
}

// snakeCase converts string to snake_case
func snakeCase(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for snakeCase")
	}

	str := args[0].String()

	// Handle camelCase/PascalCase
	reg := regexp.MustCompile(`([a-z])([A-Z])`)
	str = reg.ReplaceAllString(str, "$1_$2")

	// Replace non-alphanumeric with underscores
	reg = regexp.MustCompile(`[^a-zA-Z0-9]+`)
	str = reg.ReplaceAllString(str, "_")

	// Convert to lowercase and trim
	str = strings.ToLower(strings.Trim(str, "_"))

	if !silentMode {
		fmt.Printf("Go WASM: Converted '%s' to snake_case: '%s'\n", args[0].String(), str)
	}

	return js.ValueOf(str)
}

// extractEmails finds all email addresses in the text
func extractEmails(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for extractEmails")
	}

	text := args[0].String()
	matches := emailRegex.FindAllString(text, -1)

	// Validate emails
	var validEmails []interface{}
	for _, email := range matches {
		if _, err := mail.ParseAddress(email); err == nil {
			validEmails = append(validEmails, email)
		}
	}

	if !silentMode {
		fmt.Printf("Go WASM: Found %d valid emails in text\n", len(validEmails))
	}

	return js.ValueOf(validEmails)
}

// extractURLs finds all URLs in the text
func extractURLs(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for extractURLs")
	}

	text := args[0].String()
	matches := urlRegex.FindAllString(text, -1)

	var urls []interface{}
	for _, url := range matches {
		urls = append(urls, url)
	}

	if !silentMode {
		fmt.Printf("Go WASM: Found %d URLs in text\n", len(urls))
	}

	return js.ValueOf(urls)
}

// extractPhoneNumbers finds all phone numbers in the text
func extractPhoneNumbers(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for extractPhoneNumbers")
	}

	text := args[0].String()
	matches := phoneRegex.FindAllString(text, -1)

	var phones []interface{}
	for _, phone := range matches {
		phones = append(phones, phone)
	}

	if !silentMode {
		fmt.Printf("Go WASM: Found %d phone numbers in text\n", len(phones))
	}

	return js.ValueOf(phones)
}

// wordCount counts words in the text
func wordCount(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for wordCount")
	}

	text := args[0].String()
	words := regexp.MustCompile(`\S+`).FindAllString(text, -1)
	count := len(words)

	if !silentMode {
		fmt.Printf("Go WASM: Word count for text: %d\n", count)
	}

	return js.ValueOf(count)
}

// characterCount counts characters in the text
func characterCount(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for characterCount")
	}

	text := args[0].String()

	totalChars := utf8.RuneCountInString(text)
	totalBytes := len(text)
	withoutSpaces := utf8.RuneCountInString(regexp.MustCompile(`\s`).ReplaceAllString(text, ""))

	result := map[string]interface{}{
		"characters":         totalChars,
		"charactersNoSpaces": withoutSpaces,
		"bytes":              totalBytes,
	}

	if !silentMode {
		fmt.Printf("Go WASM: Character count - Total: %d, No spaces: %d, Bytes: %d\n",
			totalChars, withoutSpaces, totalBytes)
	}

	return js.ValueOf(result)
}

// readingTime estimates reading time in minutes
func readingTime(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 || len(args) > 2 {
		return js.ValueOf("Error: one or two arguments required for readingTime")
	}

	text := args[0].String()
	wordsPerMinute := 200 // Default reading speed

	if len(args) == 2 {
		wpm := args[1].Int()
		if wpm > 0 {
			wordsPerMinute = wpm
		}
	}

	words := regexp.MustCompile(`\S+`).FindAllString(text, -1)
	wordCount := len(words)

	minutes := math.Ceil(float64(wordCount) / float64(wordsPerMinute))

	result := map[string]interface{}{
		"minutes": int(minutes),
		"words":   wordCount,
		"wpm":     wordsPerMinute,
	}

	if !silentMode {
		fmt.Printf("Go WASM: Reading time - %d minutes for %d words at %d WPM\n",
			int(minutes), wordCount, wordsPerMinute)
	}

	return js.ValueOf(result)
}

// removeDiacritics removes accents and diacritics from text
func removeDiacritics(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for removeDiacritics")
	}

	text := args[0].String()
	result := removeDiacriticsFromString(text)

	if !silentMode {
		fmt.Printf("Go WASM: Removed diacritics from '%s' -> '%s'\n", text, result)
	}

	return js.ValueOf(result)
}

// transliterate converts text to ASCII equivalent
func transliterate(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for transliterate")
	}

	text := args[0].String()

	// First remove diacritics
	result := removeDiacriticsFromString(text)

	// Additional transliterations
	transliterations := map[string]string{
		"œ": "oe", "Œ": "OE",
		"æ": "ae", "Æ": "AE",
		"ß": "ss",
		"ł": "l", "Ł": "L",
	}

	for from, to := range transliterations {
		result = strings.ReplaceAll(result, from, to)
	}

	if !silentMode {
		fmt.Printf("Go WASM: Transliterated '%s' -> '%s'\n", text, result)
	}

	return js.ValueOf(result)
}

// generatePassword generates a secure password
func generatePassword(this js.Value, args []js.Value) interface{} {
	length := 12 // Default length
	includeSymbols := true

	if len(args) > 0 {
		length = args[0].Int()
		if length < 4 {
			length = 4
		}
		if length > 128 {
			length = 128
		}
	}

	if len(args) > 1 {
		includeSymbols = args[1].Bool()
	}

	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	if includeSymbols {
		charset += "!@#$%^&*()_+-=[]{}|;:,.<>?"
	}

	password := make([]byte, length)
	for i := range password {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		password[i] = charset[num.Int64()]
	}

	passwordStr := string(password)

	if !silentMode {
		fmt.Printf("Go WASM: Generated password of length %d\n", length)
	}

	return js.ValueOf(map[string]interface{}{
		"password":       passwordStr,
		"length":         length,
		"includeSymbols": includeSymbols,
	})
}

// validateEmail validates email format
func validateEmail(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for validateEmail")
	}

	email := args[0].String()

	// Basic format check
	if !emailRegex.MatchString(email) {
		return js.ValueOf(map[string]interface{}{
			"valid": false,
			"error": "Invalid email format",
		})
	}

	// More thorough validation
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"valid": false,
			"error": err.Error(),
		})
	}

	result := map[string]interface{}{
		"valid": true,
		"email": addr.Address,
		"name":  addr.Name,
	}

	if !silentMode {
		fmt.Printf("Go WASM: Email validation for '%s': valid\n", email)
	}

	return js.ValueOf(result)
}

// Helper functions

func removeDiacriticsFromString(text string) string {
	var result strings.Builder
	for _, r := range text {
		if replacement, exists := diacriticsMap[r]; exists {
			result.WriteRune(replacement)
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func jaroSimilarity(s1, s2 string) float64 {
	runes1 := []rune(s1)
	runes2 := []rune(s2)

	len1 := len(runes1)
	len2 := len(runes2)

	if len1 == 0 && len2 == 0 {
		return 1.0
	}
	if len1 == 0 || len2 == 0 {
		return 0.0
	}

	matchWindow := max(len1, len2)/2 - 1
	if matchWindow < 0 {
		matchWindow = 0
	}

	s1Matches := make([]bool, len1)
	s2Matches := make([]bool, len2)

	matches := 0
	transpositions := 0

	// Find matches
	for i := 0; i < len1; i++ {
		start := max(0, i-matchWindow)
		end := min(i+matchWindow+1, len2)

		for j := start; j < end; j++ {
			if s2Matches[j] || runes1[i] != runes2[j] {
				continue
			}
			s1Matches[i] = true
			s2Matches[j] = true
			matches++
			break
		}
	}

	if matches == 0 {
		return 0.0
	}

	// Count transpositions
	k := 0
	for i := 0; i < len1; i++ {
		if !s1Matches[i] {
			continue
		}
		for !s2Matches[k] {
			k++
		}
		if runes1[i] != runes2[k] {
			transpositions++
		}
		k++
	}

	return (float64(matches)/float64(len1) +
		float64(matches)/float64(len2) +
		float64(matches-transpositions/2)/float64(matches)) / 3.0
}

func commonPrefixLength(s1, s2 string, maxLen int) int {
	runes1 := []rune(s1)
	runes2 := []rune(s2)

	length := min(min(len(runes1), len(runes2)), maxLen)
	for i := 0; i < length; i++ {
		if runes1[i] != runes2[i] {
			return i
		}
	}
	return length
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// getAvailableFunctions returns all available functions
func getAvailableFunctions(this js.Value, args []js.Value) interface{} {
	functions := []string{
		"setSilentMode",
		"textSimilarity",
		"levenshteinDistance",
		"soundex",
		"slugify",
		"camelCase",
		"kebabCase",
		"snakeCase",
		"extractEmails",
		"extractURLs",
		"extractPhoneNumbers",
		"wordCount",
		"characterCount",
		"readingTime",
		"removeDiacritics",
		"transliterate",
		"generatePassword",
		"validateEmail",
		"getAvailableFunctions",
	}

	if !silentMode {
		fmt.Printf("Go WASM: Available functions: %d\n", len(functions))
	}

	return js.ValueOf(functions)
}

func main() {
	c := make(chan struct{}, 0)

	// Register functions
	js.Global().Set("setSilentMode", js.FuncOf(setSilentMode))
	js.Global().Set("textSimilarity", js.FuncOf(textSimilarity))
	js.Global().Set("levenshteinDistance", js.FuncOf(levenshteinDistance))
	js.Global().Set("soundex", js.FuncOf(soundex))
	js.Global().Set("slugify", js.FuncOf(slugify))
	js.Global().Set("camelCase", js.FuncOf(camelCase))
	js.Global().Set("kebabCase", js.FuncOf(kebabCase))
	js.Global().Set("snakeCase", js.FuncOf(snakeCase))
	js.Global().Set("extractEmails", js.FuncOf(extractEmails))
	js.Global().Set("extractURLs", js.FuncOf(extractURLs))
	js.Global().Set("extractPhoneNumbers", js.FuncOf(extractPhoneNumbers))
	js.Global().Set("wordCount", js.FuncOf(wordCount))
	js.Global().Set("characterCount", js.FuncOf(characterCount))
	js.Global().Set("readingTime", js.FuncOf(readingTime))
	js.Global().Set("removeDiacritics", js.FuncOf(removeDiacritics))
	js.Global().Set("transliterate", js.FuncOf(transliterate))
	js.Global().Set("generatePassword", js.FuncOf(generatePassword))
	js.Global().Set("validateEmail", js.FuncOf(validateEmail))
	js.Global().Set("getAvailableFunctions", js.FuncOf(getAvailableFunctions))

	fmt.Println("Go Text Processing WASM Module Loaded")
	<-c
}
