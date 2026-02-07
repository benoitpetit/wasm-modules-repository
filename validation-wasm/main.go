//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"syscall/js"
	"unicode"
)

var silentMode = false

// ============================================================================
// Regular Expressions for Validation
// ============================================================================

var (
	// Email validation (RFC 5322 simplified)
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	// URL validation
	urlRegex = regexp.MustCompile(`^(https?|ftp)://[^\s/$.?#].[^\s]*$`)

	// Phone patterns for various countries
	phonePatterns = map[string]*regexp.Regexp{
		"international": regexp.MustCompile(`^\+[1-9]\d{6,14}$`),
		"FR":            regexp.MustCompile(`^(?:\+33|0033|0)[1-9](?:[0-9]{8})$`),
		"US":            regexp.MustCompile(`^(?:\+1)?[2-9]\d{2}[2-9](?:[02-9]\d|1[02-9])\d{4}$`),
		"UK":            regexp.MustCompile(`^(?:\+44|0044|0)(?:7\d{9}|[1-9]\d{8,9})$`),
		"DE":            regexp.MustCompile(`^(?:\+49|0049|0)[1-9]\d{1,14}$`),
		"ES":            regexp.MustCompile(`^(?:\+34)?[6-9]\d{8}$`),
		"IT":            regexp.MustCompile(`^(?:\+39)?3\d{8,9}$`),
		"BE":            regexp.MustCompile(`^(?:\+32|0032|0)[1-9]\d{7,8}$`),
		"CH":            regexp.MustCompile(`^(?:\+41|0041|0)[1-9]\d{8}$`),
		"CA":            regexp.MustCompile(`^(?:\+1)?[2-9]\d{2}[2-9](?:[02-9]\d|1[02-9])\d{4}$`),
		"JP":            regexp.MustCompile(`^(?:\+81|0081|0)[1-9]\d{8,9}$`),
		"CN":            regexp.MustCompile(`^(?:\+86)?1[3-9]\d{9}$`),
		"BR":            regexp.MustCompile(`^(?:\+55)?[1-9]\d{10,11}$`),
		"IN":            regexp.MustCompile(`^(?:\+91)?[6-9]\d{9}$`),
		"AU":            regexp.MustCompile(`^(?:\+61|0061|0)[2-9]\d{8}$`),
	}

	// IBAN patterns per country (length validation)
	ibanLengths = map[string]int{
		"AL": 28, "AD": 24, "AT": 20, "AZ": 28, "BH": 22, "BY": 28,
		"BE": 16, "BA": 20, "BR": 29, "BG": 22, "CR": 22, "HR": 21,
		"CY": 28, "CZ": 24, "DK": 18, "DO": 28, "TL": 23, "EG": 29,
		"SV": 28, "EE": 20, "FO": 18, "FI": 18, "FR": 27, "GE": 22,
		"DE": 22, "GI": 23, "GR": 27, "GL": 18, "GT": 28, "HU": 28,
		"IS": 26, "IQ": 23, "IE": 22, "IL": 23, "IT": 27, "JO": 30,
		"KZ": 20, "XK": 20, "KW": 30, "LV": 21, "LB": 28, "LI": 21,
		"LT": 20, "LU": 20, "MT": 31, "MR": 27, "MU": 30, "MC": 27,
		"MD": 24, "ME": 22, "NL": 18, "MK": 19, "NO": 15, "PK": 24,
		"PS": 29, "PL": 28, "PT": 25, "QA": 29, "RO": 24, "LC": 32,
		"SM": 27, "ST": 25, "SA": 24, "RS": 22, "SC": 31, "SK": 24,
		"SI": 19, "ES": 24, "SE": 24, "CH": 21, "TN": 24, "TR": 26,
		"UA": 29, "AE": 23, "GB": 22, "VA": 22, "VG": 24,
	}

	// Postal code patterns per country
	postalCodePatterns = map[string]*regexp.Regexp{
		"FR": regexp.MustCompile(`^\d{5}$`),
		"US": regexp.MustCompile(`^\d{5}(-\d{4})?$`),
		"UK": regexp.MustCompile(`^[A-Z]{1,2}\d[A-Z\d]?\s?\d[A-Z]{2}$`),
		"DE": regexp.MustCompile(`^\d{5}$`),
		"ES": regexp.MustCompile(`^\d{5}$`),
		"IT": regexp.MustCompile(`^\d{5}$`),
		"BE": regexp.MustCompile(`^\d{4}$`),
		"CH": regexp.MustCompile(`^\d{4}$`),
		"CA": regexp.MustCompile(`^[A-Z]\d[A-Z]\s?\d[A-Z]\d$`),
		"JP": regexp.MustCompile(`^\d{3}-?\d{4}$`),
		"CN": regexp.MustCompile(`^\d{6}$`),
		"BR": regexp.MustCompile(`^\d{5}-?\d{3}$`),
		"IN": regexp.MustCompile(`^\d{6}$`),
		"AU": regexp.MustCompile(`^\d{4}$`),
		"NL": regexp.MustCompile(`^\d{4}\s?[A-Z]{2}$`),
		"PT": regexp.MustCompile(`^\d{4}-?\d{3}$`),
		"PL": regexp.MustCompile(`^\d{2}-?\d{3}$`),
		"SE": regexp.MustCompile(`^\d{3}\s?\d{2}$`),
		"NO": regexp.MustCompile(`^\d{4}$`),
		"DK": regexp.MustCompile(`^\d{4}$`),
		"AT": regexp.MustCompile(`^\d{4}$`),
		"RU": regexp.MustCompile(`^\d{6}$`),
		"MX": regexp.MustCompile(`^\d{5}$`),
	}

	// Credit card patterns
	creditCardPatterns = map[string]*regexp.Regexp{
		"visa":       regexp.MustCompile(`^4[0-9]{12}(?:[0-9]{3})?$`),
		"mastercard": regexp.MustCompile(`^5[1-5][0-9]{14}$|^2(?:2(?:2[1-9]|[3-9][0-9])|[3-6][0-9][0-9]|7(?:[01][0-9]|20))[0-9]{12}$`),
		"amex":       regexp.MustCompile(`^3[47][0-9]{13}$`),
		"discover":   regexp.MustCompile(`^6(?:011|5[0-9]{2})[0-9]{12}$`),
		"diners":     regexp.MustCompile(`^3(?:0[0-5]|[68][0-9])[0-9]{11}$`),
		"jcb":        regexp.MustCompile(`^(?:2131|1800|35\d{3})\d{11}$`),
		"unionpay":   regexp.MustCompile(`^62[0-9]{14,17}$`),
	}
)

// ============================================================================
// System Functions
// ============================================================================

// setSilentMode enables/disables silent mode for console logs
func setSilentMode(this js.Value, args []js.Value) interface{} {
	if len(args) == 1 {
		silentMode = args[0].Bool()
	}
	return js.ValueOf(silentMode)
}

// ============================================================================
// Email Validation
// ============================================================================

// validateEmail validates an email address with detailed results
func validateEmail(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for validateEmail (email string)")
	}

	email := strings.TrimSpace(args[0].String())

	result := map[string]interface{}{
		"input": email,
		"valid": false,
	}

	if email == "" {
		result["error"] = "Email address cannot be empty"
		return js.ValueOf(result)
	}

	// Basic regex check
	if !emailRegex.MatchString(email) {
		result["error"] = "Invalid email format"
		return js.ValueOf(result)
	}

	// Parse with Go's mail package for more thorough validation
	addr, err := mail.ParseAddress(email)
	if err != nil {
		result["error"] = fmt.Sprintf("Invalid email: %v", err)
		return js.ValueOf(result)
	}

	parts := strings.SplitN(addr.Address, "@", 2)
	if len(parts) != 2 {
		result["error"] = "Invalid email structure"
		return js.ValueOf(result)
	}

	localPart := parts[0]
	domain := parts[1]

	// Check domain has at least one dot
	if !strings.Contains(domain, ".") {
		result["error"] = "Invalid domain: must contain at least one dot"
		return js.ValueOf(result)
	}

	// Check TLD length
	tldParts := strings.Split(domain, ".")
	tld := tldParts[len(tldParts)-1]
	if len(tld) < 2 {
		result["error"] = "Invalid TLD: too short"
		return js.ValueOf(result)
	}

	result["valid"] = true
	result["email"] = addr.Address
	result["localPart"] = localPart
	result["domain"] = domain
	result["tld"] = tld

	if addr.Name != "" {
		result["displayName"] = addr.Name
	}

	if !silentMode {
		fmt.Printf("Go WASM: Email validation for '%s': valid\n", email)
	}

	return js.ValueOf(result)
}

// validateEmailBatch validates multiple email addresses
func validateEmailBatch(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf("Error: at least one email argument required for validateEmailBatch")
	}

	results := make([]interface{}, 0, len(args))
	validCount := 0

	for _, arg := range args {
		email := arg.String()
		valid := emailRegex.MatchString(strings.TrimSpace(email))
		if valid {
			_, err := mail.ParseAddress(strings.TrimSpace(email))
			valid = err == nil
		}
		if valid {
			validCount++
		}
		results = append(results, map[string]interface{}{
			"email": email,
			"valid": valid,
		})
	}

	result := map[string]interface{}{
		"results":    results,
		"total":      len(args),
		"valid":      validCount,
		"invalid":    len(args) - validCount,
		"validRatio": fmt.Sprintf("%.1f%%", float64(validCount)/float64(len(args))*100),
	}

	if !silentMode {
		fmt.Printf("Go WASM: Batch email validation: %d/%d valid\n", validCount, len(args))
	}

	return js.ValueOf(result)
}

// ============================================================================
// URL Validation
// ============================================================================

// validateURL validates a URL with detailed breakdown
func validateURL(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for validateURL (URL string)")
	}

	urlStr := strings.TrimSpace(args[0].String())

	result := map[string]interface{}{
		"input": urlStr,
		"valid": false,
	}

	if urlStr == "" {
		result["error"] = "URL cannot be empty"
		return js.ValueOf(result)
	}

	// Basic format check
	if !urlRegex.MatchString(urlStr) {
		result["error"] = "Invalid URL format"
		return js.ValueOf(result)
	}

	// Parse URL components
	var protocol, host, path, query, fragment string

	// Extract protocol
	protocolEnd := strings.Index(urlStr, "://")
	if protocolEnd == -1 {
		result["error"] = "Missing protocol (http:// or https://)"
		return js.ValueOf(result)
	}
	protocol = urlStr[:protocolEnd]

	rest := urlStr[protocolEnd+3:]

	// Extract fragment
	if fragIdx := strings.Index(rest, "#"); fragIdx != -1 {
		fragment = rest[fragIdx+1:]
		rest = rest[:fragIdx]
	}

	// Extract query
	if queryIdx := strings.Index(rest, "?"); queryIdx != -1 {
		query = rest[queryIdx+1:]
		rest = rest[:queryIdx]
	}

	// Extract path
	if pathIdx := strings.Index(rest, "/"); pathIdx != -1 {
		path = rest[pathIdx:]
		host = rest[:pathIdx]
	} else {
		host = rest
	}

	// Validate host
	if host == "" {
		result["error"] = "Missing host"
		return js.ValueOf(result)
	}

	// Extract port if present
	port := ""
	if colonIdx := strings.LastIndex(host, ":"); colonIdx != -1 {
		portStr := host[colonIdx+1:]
		host = host[:colonIdx]
		if p, err := strconv.Atoi(portStr); err == nil && p > 0 && p <= 65535 {
			port = portStr
		}
	}

	// Detect if IP or domain
	isIP := false
	ipv4Regex := regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
	if ipv4Regex.MatchString(host) {
		isIP = true
	}

	// Check for HTTPS
	isSecure := protocol == "https"

	result["valid"] = true
	result["protocol"] = protocol
	result["host"] = host
	result["isSecure"] = isSecure
	result["isIP"] = isIP

	if port != "" {
		result["port"] = port
	}
	if path != "" {
		result["path"] = path
	}
	if query != "" {
		result["query"] = query
	}
	if fragment != "" {
		result["fragment"] = fragment
	}

	if !silentMode {
		fmt.Printf("Go WASM: URL validation for '%s': valid (%s)\n", urlStr, protocol)
	}

	return js.ValueOf(result)
}

// ============================================================================
// Phone Number Validation
// ============================================================================

// validatePhoneNumber validates a phone number with optional country code
func validatePhoneNumber(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf("Error: at least one argument required for validatePhoneNumber (phone string, optional country code)")
	}

	phone := strings.TrimSpace(args[0].String())
	// Remove spaces, dashes, dots, parentheses for validation
	cleanPhone := regexp.MustCompile(`[\s\-\.\(\)]+`).ReplaceAllString(phone, "")

	result := map[string]interface{}{
		"input":      phone,
		"cleaned":    cleanPhone,
		"valid":      false,
		"detectedAs": []string{},
	}

	if cleanPhone == "" {
		result["error"] = "Phone number cannot be empty"
		return js.ValueOf(result)
	}

	// If country code specified, validate against that country
	if len(args) >= 2 {
		country := strings.ToUpper(args[1].String())
		if pattern, ok := phonePatterns[country]; ok {
			valid := pattern.MatchString(cleanPhone)
			result["valid"] = valid
			result["country"] = country
			if !valid {
				result["error"] = fmt.Sprintf("Invalid phone number for country %s", country)
			}
			return js.ValueOf(result)
		}
		result["error"] = fmt.Sprintf("Unknown country code: %s", country)
		return js.ValueOf(result)
	}

	// Try to detect country
	matches := make([]string, 0)
	for country, pattern := range phonePatterns {
		if pattern.MatchString(cleanPhone) {
			matches = append(matches, country)
		}
	}

	if len(matches) > 0 {
		result["valid"] = true
		result["detectedAs"] = matches
		result["country"] = matches[0]
	} else {
		// Fallback: basic length and format check
		if len(cleanPhone) >= 7 && len(cleanPhone) <= 15 {
			if cleanPhone[0] == '+' || (cleanPhone[0] >= '0' && cleanPhone[0] <= '9') {
				result["valid"] = true
				result["detectedAs"] = []string{"unknown"}
				result["country"] = "unknown"
			} else {
				result["error"] = "Invalid phone number format"
			}
		} else {
			result["error"] = "Phone number must be between 7 and 15 digits"
		}
	}

	if !silentMode {
		fmt.Printf("Go WASM: Phone validation for '%s': %v\n", phone, result["valid"])
	}

	return js.ValueOf(result)
}

// ============================================================================
// IBAN Validation
// ============================================================================

// validateIBAN validates an International Bank Account Number
func validateIBAN(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for validateIBAN (IBAN string)")
	}

	iban := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(args[0].String()), " ", ""))

	result := map[string]interface{}{
		"input":     args[0].String(),
		"formatted": iban,
		"valid":     false,
	}

	if len(iban) < 5 {
		result["error"] = "IBAN too short"
		return js.ValueOf(result)
	}

	// Extract country code
	countryCode := iban[:2]
	checkDigits := iban[2:4]
	bban := iban[4:]

	result["countryCode"] = countryCode
	result["checkDigits"] = checkDigits
	result["bban"] = bban

	// Check country length
	expectedLen, knownCountry := ibanLengths[countryCode]
	if knownCountry {
		if len(iban) != expectedLen {
			result["error"] = fmt.Sprintf("Invalid IBAN length for %s: expected %d, got %d", countryCode, expectedLen, len(iban))
			return js.ValueOf(result)
		}
		result["expectedLength"] = expectedLen
	}

	// Validate characters (only alphanumeric)
	for _, ch := range iban {
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) {
			result["error"] = "IBAN contains invalid characters"
			return js.ValueOf(result)
		}
	}

	// MOD-97 validation (ISO 7064)
	// Move first 4 characters to end
	rearranged := iban[4:] + iban[:4]

	// Convert letters to numbers (A=10, B=11, ..., Z=35)
	var numStr strings.Builder
	for _, ch := range rearranged {
		if unicode.IsLetter(ch) {
			numStr.WriteString(strconv.Itoa(int(ch-'A') + 10))
		} else {
			numStr.WriteRune(ch)
		}
	}

	// Calculate MOD-97 on the large number (using partial computation)
	remainder := 0
	for _, digit := range numStr.String() {
		remainder = (remainder*10 + int(digit-'0')) % 97
	}

	if remainder != 1 {
		result["error"] = "IBAN check digits validation failed (MOD-97)"
		return js.ValueOf(result)
	}

	result["valid"] = true
	result["country"] = countryCode

	// Format IBAN with spaces every 4 characters
	formatted := ""
	for i, ch := range iban {
		if i > 0 && i%4 == 0 {
			formatted += " "
		}
		formatted += string(ch)
	}
	result["displayFormat"] = formatted

	if !silentMode {
		fmt.Printf("Go WASM: IBAN validation for '%s': valid (%s)\n", iban, countryCode)
	}

	return js.ValueOf(result)
}

// ============================================================================
// Credit Card Validation (Luhn Algorithm)
// ============================================================================

// validateCreditCard validates a credit card number using the Luhn algorithm
func validateCreditCard(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for validateCreditCard (card number string)")
	}

	input := strings.TrimSpace(args[0].String())
	// Remove spaces and dashes
	cardNumber := regexp.MustCompile(`[\s\-]+`).ReplaceAllString(input, "")

	result := map[string]interface{}{
		"input":      input,
		"number":     cardNumber,
		"valid":      false,
		"luhnValid":  false,
		"cardType":   "unknown",
		"maskedCard": "",
	}

	// Check that all characters are digits
	for _, ch := range cardNumber {
		if !unicode.IsDigit(ch) {
			result["error"] = "Card number must contain only digits"
			return js.ValueOf(result)
		}
	}

	// Check length (most cards are 13-19 digits)
	if len(cardNumber) < 13 || len(cardNumber) > 19 {
		result["error"] = "Card number must be between 13 and 19 digits"
		return js.ValueOf(result)
	}

	// Luhn algorithm
	luhnValid := luhnCheck(cardNumber)
	result["luhnValid"] = luhnValid

	// Detect card type
	cardType := "unknown"
	for ctype, pattern := range creditCardPatterns {
		if pattern.MatchString(cardNumber) {
			cardType = ctype
			break
		}
	}
	result["cardType"] = cardType

	// Create masked version
	if len(cardNumber) >= 4 {
		masked := strings.Repeat("*", len(cardNumber)-4) + cardNumber[len(cardNumber)-4:]
		result["maskedCard"] = masked
	}

	// Card length validation per type
	validLength := true
	switch cardType {
	case "visa":
		validLength = len(cardNumber) == 13 || len(cardNumber) == 16
	case "mastercard":
		validLength = len(cardNumber) == 16
	case "amex":
		validLength = len(cardNumber) == 15
	case "discover":
		validLength = len(cardNumber) == 16
	case "diners":
		validLength = len(cardNumber) == 14
	case "jcb":
		validLength = len(cardNumber) == 15 || len(cardNumber) == 16
	}

	result["valid"] = luhnValid && validLength
	result["lengthValid"] = validLength

	if !luhnValid {
		result["error"] = "Luhn check failed: invalid card number"
	} else if !validLength {
		result["error"] = fmt.Sprintf("Invalid length for %s card", cardType)
	}

	if !silentMode {
		fmt.Printf("Go WASM: Credit card validation: %s, type: %s, luhn: %v\n", cardNumber[:4]+"...", cardType, luhnValid)
	}

	return js.ValueOf(result)
}

// luhnCheck implements the Luhn algorithm for credit card number validation
func luhnCheck(number string) bool {
	sum := 0
	nDigits := len(number)
	parity := nDigits % 2

	for i := 0; i < nDigits; i++ {
		digit := int(number[i] - '0')

		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
	}

	return sum%10 == 0
}

// luhnGenerate generates the Luhn check digit for a partial number
func luhnGenerate(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for luhnGenerate (partial number string without check digit)")
	}

	partial := strings.TrimSpace(args[0].String())

	// Validate input contains only digits
	for _, ch := range partial {
		if !unicode.IsDigit(ch) {
			return js.ValueOf("Error: input must contain only digits")
		}
	}

	// Try each digit 0-9 as check digit
	for d := 0; d <= 9; d++ {
		candidate := partial + strconv.Itoa(d)
		if luhnCheck(candidate) {
			result := map[string]interface{}{
				"partialNumber":  partial,
				"checkDigit":     d,
				"completeNumber": candidate,
				"valid":          true,
			}

			if !silentMode {
				fmt.Printf("Go WASM: Luhn check digit for '%s' = %d\n", partial, d)
			}

			return js.ValueOf(result)
		}
	}

	return js.ValueOf("Error: could not generate Luhn check digit")
}

// ============================================================================
// Postal Code Validation
// ============================================================================

// validatePostalCode validates a postal code for a given country
func validatePostalCode(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return js.ValueOf("Error: two arguments required for validatePostalCode (postal code, country code)")
	}

	postalCode := strings.TrimSpace(args[0].String())
	country := strings.ToUpper(strings.TrimSpace(args[1].String()))

	result := map[string]interface{}{
		"input":   postalCode,
		"country": country,
		"valid":   false,
	}

	if postalCode == "" {
		result["error"] = "Postal code cannot be empty"
		return js.ValueOf(result)
	}

	pattern, ok := postalCodePatterns[country]
	if !ok {
		result["error"] = fmt.Sprintf("Unknown country code: %s. Supported: FR, US, UK, DE, ES, IT, BE, CH, CA, JP, CN, BR, IN, AU, NL, PT, PL, SE, NO, DK, AT, RU, MX", country)
		result["supportedCountries"] = getSupportedPostalCountries()
		return js.ValueOf(result)
	}

	// Normalize for case-insensitive matching
	testCode := strings.ToUpper(postalCode)

	valid := pattern.MatchString(testCode)
	result["valid"] = valid

	if !valid {
		result["error"] = fmt.Sprintf("Invalid postal code format for %s", country)
	}

	if !silentMode {
		fmt.Printf("Go WASM: Postal code validation for '%s' (%s): %v\n", postalCode, country, valid)
	}

	return js.ValueOf(result)
}

func getSupportedPostalCountries() []string {
	countries := make([]string, 0, len(postalCodePatterns))
	for k := range postalCodePatterns {
		countries = append(countries, k)
	}
	return countries
}

// ============================================================================
// JSON Schema Validation
// ============================================================================

// validateJSONSchema validates a JSON string against a schema definition
func validateJSONSchema(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return js.ValueOf("Error: two arguments required for validateJSONSchema (JSON data string, schema string)")
	}

	dataStr := args[0].String()
	schemaStr := args[1].String()

	result := map[string]interface{}{
		"valid":  false,
		"errors": []interface{}{},
	}

	// Parse JSON data
	var data interface{}
	if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
		result["error"] = fmt.Sprintf("Invalid JSON data: %v", err)
		return js.ValueOf(result)
	}

	// Parse schema
	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(schemaStr), &schema); err != nil {
		result["error"] = fmt.Sprintf("Invalid JSON schema: %v", err)
		return js.ValueOf(result)
	}

	// Validate
	errors := validateNode(data, schema, "$")
	result["errors"] = errors
	result["valid"] = len(errors) == 0
	result["errorCount"] = len(errors)

	if !silentMode {
		fmt.Printf("Go WASM: JSON schema validation: %d errors\n", len(errors))
	}

	return js.ValueOf(result)
}

// validateNode validates a JSON node against a schema node
func validateNode(data interface{}, schema map[string]interface{}, path string) []interface{} {
	var errors []interface{}

	// Check type
	if expectedType, ok := schema["type"]; ok {
		typeStr, isStr := expectedType.(string)
		if isStr && !checkJSONType(data, typeStr) {
			errors = append(errors, map[string]interface{}{
				"path":     path,
				"message":  fmt.Sprintf("Expected type '%s', got '%s'", typeStr, getJSONType(data)),
				"expected": typeStr,
				"actual":   getJSONType(data),
			})
			return errors // Type mismatch, no further validation needed
		}
	}

	switch d := data.(type) {
	case map[string]interface{}:
		// Object validation
		if required, ok := schema["required"]; ok {
			if reqArr, ok := required.([]interface{}); ok {
				for _, req := range reqArr {
					reqStr, _ := req.(string)
					if _, exists := d[reqStr]; !exists {
						errors = append(errors, map[string]interface{}{
							"path":    path + "." + reqStr,
							"message": fmt.Sprintf("Required property '%s' is missing", reqStr),
						})
					}
				}
			}
		}

		// Validate properties
		if properties, ok := schema["properties"]; ok {
			if propMap, ok := properties.(map[string]interface{}); ok {
				for propName, propSchema := range propMap {
					if propValue, exists := d[propName]; exists {
						if ps, ok := propSchema.(map[string]interface{}); ok {
							propErrors := validateNode(propValue, ps, path+"."+propName)
							errors = append(errors, propErrors...)
						}
					}
				}
			}
		}

		// minProperties / maxProperties
		if minProps, ok := schema["minProperties"]; ok {
			if mp, ok := minProps.(float64); ok && float64(len(d)) < mp {
				errors = append(errors, map[string]interface{}{
					"path":    path,
					"message": fmt.Sprintf("Object must have at least %d properties, has %d", int(mp), len(d)),
				})
			}
		}
		if maxProps, ok := schema["maxProperties"]; ok {
			if mp, ok := maxProps.(float64); ok && float64(len(d)) > mp {
				errors = append(errors, map[string]interface{}{
					"path":    path,
					"message": fmt.Sprintf("Object must have at most %d properties, has %d", int(mp), len(d)),
				})
			}
		}

	case []interface{}:
		// Array validation
		if minItems, ok := schema["minItems"]; ok {
			if mi, ok := minItems.(float64); ok && float64(len(d)) < mi {
				errors = append(errors, map[string]interface{}{
					"path":    path,
					"message": fmt.Sprintf("Array must have at least %d items, has %d", int(mi), len(d)),
				})
			}
		}
		if maxItems, ok := schema["maxItems"]; ok {
			if mi, ok := maxItems.(float64); ok && float64(len(d)) > mi {
				errors = append(errors, map[string]interface{}{
					"path":    path,
					"message": fmt.Sprintf("Array must have at most %d items, has %d", int(mi), len(d)),
				})
			}
		}
		// Validate items
		if items, ok := schema["items"]; ok {
			if itemSchema, ok := items.(map[string]interface{}); ok {
				for i, item := range d {
					itemErrors := validateNode(item, itemSchema, fmt.Sprintf("%s[%d]", path, i))
					errors = append(errors, itemErrors...)
				}
			}
		}

	case float64:
		// Number validation
		if minimum, ok := schema["minimum"]; ok {
			if min, ok := minimum.(float64); ok && d < min {
				errors = append(errors, map[string]interface{}{
					"path":    path,
					"message": fmt.Sprintf("Value %v is less than minimum %v", d, min),
				})
			}
		}
		if maximum, ok := schema["maximum"]; ok {
			if max, ok := maximum.(float64); ok && d > max {
				errors = append(errors, map[string]interface{}{
					"path":    path,
					"message": fmt.Sprintf("Value %v is greater than maximum %v", d, max),
				})
			}
		}

	case string:
		// String validation
		if minLen, ok := schema["minLength"]; ok {
			if ml, ok := minLen.(float64); ok && float64(len(d)) < ml {
				errors = append(errors, map[string]interface{}{
					"path":    path,
					"message": fmt.Sprintf("String length %d is less than minimum %d", len(d), int(ml)),
				})
			}
		}
		if maxLen, ok := schema["maxLength"]; ok {
			if ml, ok := maxLen.(float64); ok && float64(len(d)) > ml {
				errors = append(errors, map[string]interface{}{
					"path":    path,
					"message": fmt.Sprintf("String length %d is greater than maximum %d", len(d), int(ml)),
				})
			}
		}
		// Pattern validation
		if pattern, ok := schema["pattern"]; ok {
			if patStr, ok := pattern.(string); ok {
				if r, err := regexp.Compile(patStr); err == nil {
					if !r.MatchString(d) {
						errors = append(errors, map[string]interface{}{
							"path":    path,
							"message": fmt.Sprintf("String does not match pattern '%s'", patStr),
						})
					}
				}
			}
		}
		// Enum validation
		if enum, ok := schema["enum"]; ok {
			if enumArr, ok := enum.([]interface{}); ok {
				found := false
				for _, v := range enumArr {
					if vs, ok := v.(string); ok && vs == d {
						found = true
						break
					}
				}
				if !found {
					errors = append(errors, map[string]interface{}{
						"path":    path,
						"message": fmt.Sprintf("Value '%s' is not in enum", d),
					})
				}
			}
		}
		// Format validation
		if format, ok := schema["format"]; ok {
			if fmtStr, ok := format.(string); ok {
				if err := validateFormat(d, fmtStr); err != "" {
					errors = append(errors, map[string]interface{}{
						"path":    path,
						"message": err,
					})
				}
			}
		}
	}

	return errors
}

func checkJSONType(data interface{}, expectedType string) bool {
	switch expectedType {
	case "string":
		_, ok := data.(string)
		return ok
	case "number", "integer":
		v, ok := data.(float64)
		if expectedType == "integer" && ok {
			return v == math.Floor(v)
		}
		return ok
	case "boolean":
		_, ok := data.(bool)
		return ok
	case "object":
		_, ok := data.(map[string]interface{})
		return ok
	case "array":
		_, ok := data.([]interface{})
		return ok
	case "null":
		return data == nil
	}
	return true
}

func getJSONType(data interface{}) string {
	switch data.(type) {
	case string:
		return "string"
	case float64:
		return "number"
	case bool:
		return "boolean"
	case map[string]interface{}:
		return "object"
	case []interface{}:
		return "array"
	case nil:
		return "null"
	}
	return "unknown"
}

func validateFormat(value string, format string) string {
	switch format {
	case "email":
		if !emailRegex.MatchString(value) {
			return fmt.Sprintf("String '%s' is not a valid email", value)
		}
	case "uri", "url":
		if !urlRegex.MatchString(value) {
			return fmt.Sprintf("String '%s' is not a valid URL", value)
		}
	case "date":
		dateRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
		if !dateRegex.MatchString(value) {
			return fmt.Sprintf("String '%s' is not a valid date (YYYY-MM-DD)", value)
		}
	case "date-time":
		dtRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`)
		if !dtRegex.MatchString(value) {
			return fmt.Sprintf("String '%s' is not a valid date-time", value)
		}
	case "ipv4":
		ipv4Regex := regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
		if !ipv4Regex.MatchString(value) {
			return fmt.Sprintf("String '%s' is not a valid IPv4 address", value)
		}
		// Validate each octet
		parts := strings.Split(value, ".")
		for _, part := range parts {
			n, _ := strconv.Atoi(part)
			if n < 0 || n > 255 {
				return fmt.Sprintf("String '%s' has invalid IPv4 octet", value)
			}
		}
	case "ipv6":
		ipv6Regex := regexp.MustCompile(`^([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$|^::$|^([0-9a-fA-F]{1,4}:)*::([0-9a-fA-F]{1,4}:)*[0-9a-fA-F]{1,4}$`)
		if !ipv6Regex.MatchString(value) {
			return fmt.Sprintf("String '%s' is not a valid IPv6 address", value)
		}
	case "uuid":
		uuidRegex := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
		if !uuidRegex.MatchString(value) {
			return fmt.Sprintf("String '%s' is not a valid UUID", value)
		}
	}
	return ""
}

// ============================================================================
// Regex Pattern Validation
// ============================================================================

// validateRegex tests a string against a regex pattern
func validateRegex(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return js.ValueOf("Error: two arguments required for validateRegex (string, regex pattern)")
	}

	input := args[0].String()
	pattern := args[1].String()

	result := map[string]interface{}{
		"input":   input,
		"pattern": pattern,
		"valid":   false,
	}

	r, err := regexp.Compile(pattern)
	if err != nil {
		result["error"] = fmt.Sprintf("Invalid regex pattern: %v", err)
		return js.ValueOf(result)
	}

	matches := r.MatchString(input)
	result["valid"] = matches

	if matches {
		// Find all matches
		allMatches := r.FindAllString(input, -1)
		result["matches"] = allMatches
		result["matchCount"] = len(allMatches)

		// Find first match with groups
		submatches := r.FindStringSubmatch(input)
		if len(submatches) > 1 {
			result["groups"] = submatches[1:]
		}

		// Find match locations
		loc := r.FindStringIndex(input)
		if loc != nil {
			result["firstMatchStart"] = loc[0]
			result["firstMatchEnd"] = loc[1]
			result["firstMatch"] = input[loc[0]:loc[1]]
		}
	}

	if !silentMode {
		fmt.Printf("Go WASM: Regex validation '%s' against '%s': %v\n", input, pattern, matches)
	}

	return js.ValueOf(result)
}

// testRegexPattern compiles and tests a regex pattern without a target string
func testRegexPattern(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for testRegexPattern (regex pattern string)")
	}

	pattern := args[0].String()

	result := map[string]interface{}{
		"pattern":    pattern,
		"valid":      false,
		"hasGroups":  false,
		"groupCount": 0,
	}

	r, err := regexp.Compile(pattern)
	if err != nil {
		result["error"] = fmt.Sprintf("Invalid regex pattern: %v", err)
		return js.ValueOf(result)
	}

	result["valid"] = true
	result["groupCount"] = r.NumSubexp()
	result["hasGroups"] = r.NumSubexp() > 0
	result["subexpNames"] = r.SubexpNames()

	if !silentMode {
		fmt.Printf("Go WASM: Regex pattern '%s' is valid (%d groups)\n", pattern, r.NumSubexp())
	}

	return js.ValueOf(result)
}

// replaceWithRegex replaces matches in a string using a regex pattern
func replaceWithRegex(this js.Value, args []js.Value) interface{} {
	if len(args) < 3 {
		return js.ValueOf("Error: three arguments required for replaceWithRegex (input, pattern, replacement)")
	}

	input := args[0].String()
	pattern := args[1].String()
	replacement := args[2].String()

	r, err := regexp.Compile(pattern)
	if err != nil {
		return js.ValueOf(fmt.Sprintf("Error: invalid regex pattern: %v", err))
	}

	replaced := r.ReplaceAllString(input, replacement)
	matchCount := len(r.FindAllString(input, -1))

	result := map[string]interface{}{
		"input":        input,
		"pattern":      pattern,
		"replacement":  replacement,
		"result":       replaced,
		"matchCount":   matchCount,
		"replacements": matchCount,
	}

	if !silentMode {
		fmt.Printf("Go WASM: Regex replace: %d replacements made\n", matchCount)
	}

	return js.ValueOf(result)
}

// ============================================================================
// Comprehensive Validation
// ============================================================================

// validateAll performs multiple validations on a single input
func validateAll(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf("Error: one argument required for validateAll (string to validate)")
	}

	input := strings.TrimSpace(args[0].String())

	results := map[string]interface{}{
		"input": input,
	}

	// Test as email
	emailValid := emailRegex.MatchString(input)
	results["isEmail"] = emailValid

	// Test as URL
	urlValid := urlRegex.MatchString(input)
	results["isURL"] = urlValid

	// Test as phone (cleaned)
	cleanPhone := regexp.MustCompile(`[\s\-\.\(\)]+`).ReplaceAllString(input, "")
	phoneValid := false
	for _, pattern := range phonePatterns {
		if pattern.MatchString(cleanPhone) {
			phoneValid = true
			break
		}
	}
	results["isPhone"] = phoneValid

	// Test as IBAN
	cleanIBAN := strings.ToUpper(strings.ReplaceAll(input, " ", ""))
	ibanValid := false
	if len(cleanIBAN) >= 15 && len(cleanIBAN) <= 34 {
		cc := cleanIBAN[:2]
		if _, ok := ibanLengths[cc]; ok {
			ibanValid = len(cleanIBAN) == ibanLengths[cc]
		}
	}
	results["isIBAN"] = ibanValid

	// Test as credit card
	cleanCard := regexp.MustCompile(`[\s\-]+`).ReplaceAllString(input, "")
	ccValid := false
	if len(cleanCard) >= 13 && len(cleanCard) <= 19 {
		allDigits := true
		for _, ch := range cleanCard {
			if !unicode.IsDigit(ch) {
				allDigits = false
				break
			}
		}
		if allDigits {
			ccValid = luhnCheck(cleanCard)
		}
	}
	results["isCreditCard"] = ccValid

	// Detect best match
	detected := "unknown"
	if emailValid {
		detected = "email"
	} else if urlValid {
		detected = "url"
	} else if ccValid {
		detected = "credit_card"
	} else if ibanValid {
		detected = "iban"
	} else if phoneValid {
		detected = "phone"
	}
	results["detectedType"] = detected

	if !silentMode {
		fmt.Printf("Go WASM: validateAll for '%s': detected as %s\n", input, detected)
	}

	return js.ValueOf(results)
}

// ============================================================================
// Module Info & System Functions
// ============================================================================

// getModuleInfo returns comprehensive module information
func getModuleInfo(this js.Value, args []js.Value) interface{} {
	info := map[string]interface{}{
		"name":        "validation-wasm",
		"version":     "0.1.0",
		"description": "Comprehensive data validation module for emails, URLs, phones, IBAN, credit cards, postal codes, JSON schema, and regex patterns",
		"author":      "Ben",
		"language":    "Go",
		"target":      "WebAssembly",
		"functions":   16,
		"categories": []string{
			"Email Validation",
			"URL Validation",
			"Phone Validation",
			"IBAN Validation",
			"Credit Card Validation",
			"Postal Code Validation",
			"JSON Schema Validation",
			"Regex Pattern Validation",
			"Comprehensive Validation",
		},
		"features": []string{
			"RFC 5322 email validation",
			"Multi-country phone number validation",
			"IBAN MOD-97 validation (ISO 7064)",
			"Luhn algorithm for credit card validation",
			"International postal code validation (23 countries)",
			"JSON schema validation with nested objects",
			"Regex pattern matching and replacement",
			"Auto-detection of input type",
			"Batch email validation",
		},
		"buildInfo": map[string]interface{}{
			"goVersion":    "1.21+",
			"dependencies": []string{},
			"optimized":    true,
			"compressed":   true,
		},
	}

	if !silentMode {
		fmt.Printf("Go WASM: Module info retrieved for validation-wasm v0.1.0\n")
	}

	// Convert to JSON and back to avoid nested map issues with js.ValueOf
	jsonBytes, err := json.Marshal(info)
	if err != nil {
		return fmt.Sprintf("Error: Failed to marshal module info: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return fmt.Sprintf("Error: Failed to unmarshal module info: %v", err)
	}

	return js.ValueOf(result)
}

// getAvailableFunctions returns list of all available functions
func getAvailableFunctions(this js.Value, args []js.Value) interface{} {
	functions := []string{
		// Email
		"validateEmail", "validateEmailBatch",
		// URL
		"validateURL",
		// Phone
		"validatePhoneNumber",
		// IBAN
		"validateIBAN",
		// Credit Card
		"validateCreditCard", "luhnGenerate",
		// Postal Code
		"validatePostalCode",
		// JSON Schema
		"validateJSONSchema",
		// Regex
		"validateRegex", "testRegexPattern", "replaceWithRegex",
		// Comprehensive
		"validateAll",
		// System
		"setSilentMode", "getAvailableFunctions", "getModuleInfo",
	}

	if !silentMode {
		fmt.Printf("Go WASM: Available functions: %d\n", len(functions))
	}

	return js.ValueOf(functions)
}

func main() {
	c := make(chan struct{})

	fmt.Println("Go WASM Validation Module initializing...")

	// Register Email validation functions
	js.Global().Set("validateEmail", js.FuncOf(validateEmail))
	js.Global().Set("validateEmailBatch", js.FuncOf(validateEmailBatch))

	// Register URL validation
	js.Global().Set("validateURL", js.FuncOf(validateURL))

	// Register Phone validation
	js.Global().Set("validatePhoneNumber", js.FuncOf(validatePhoneNumber))

	// Register IBAN validation
	js.Global().Set("validateIBAN", js.FuncOf(validateIBAN))

	// Register Credit Card validation
	js.Global().Set("validateCreditCard", js.FuncOf(validateCreditCard))
	js.Global().Set("luhnGenerate", js.FuncOf(luhnGenerate))

	// Register Postal Code validation
	js.Global().Set("validatePostalCode", js.FuncOf(validatePostalCode))

	// Register JSON Schema validation
	js.Global().Set("validateJSONSchema", js.FuncOf(validateJSONSchema))

	// Register Regex validation
	js.Global().Set("validateRegex", js.FuncOf(validateRegex))
	js.Global().Set("testRegexPattern", js.FuncOf(testRegexPattern))
	js.Global().Set("replaceWithRegex", js.FuncOf(replaceWithRegex))

	// Register Comprehensive validation
	js.Global().Set("validateAll", js.FuncOf(validateAll))

	// Register System functions
	js.Global().Set("setSilentMode", js.FuncOf(setSilentMode))
	js.Global().Set("getAvailableFunctions", js.FuncOf(getAvailableFunctions))
	js.Global().Set("getModuleInfo", js.FuncOf(getModuleInfo))

	// Signal readiness for GoWM
	js.Global().Set("__gowm_ready", js.ValueOf(true))

	fmt.Println("Go WASM Validation Module ready")
	<-c
}
