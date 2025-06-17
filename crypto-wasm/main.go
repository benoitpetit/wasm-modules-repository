//go:build js && wasm

package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strings"
	"syscall/js"
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"
)

var silentMode = false

// CryptoError represents an error in crypto operations
type CryptoError struct {
	Operation string `json:"operation"`
	Message   string `json:"message"`
}

// KeyPair represents an RSA key pair
type KeyPair struct {
	PublicKey  string `json:"publicKey"`
	PrivateKey string `json:"privateKey"`
}

// JWTClaims represents JWT claims
type JWTClaims struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

// setSilentMode - Set silent mode for operations
func setSilentMode(this js.Value, args []js.Value) interface{} {
	if len(args) == 1 {
		silentMode = args[0].Bool()
	}
	return js.ValueOf(silentMode)
}

// hashSHA256 - Generate SHA256 hash
func hashSHA256(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf(map[string]interface{}{
			"error": "hashSHA256 requires exactly 1 argument (data)",
		})
	}

	data := args[0].String()
	hash := sha256.Sum256([]byte(data))
	result := hex.EncodeToString(hash[:])

	if !silentMode {
		fmt.Printf("Go WASM: SHA256 hash generated for %d bytes\n", len(data))
	}

	return js.ValueOf(map[string]interface{}{
		"hash": result,
		"algorithm": "SHA256",
	})
}

// hashSHA512 - Generate SHA512 hash
func hashSHA512(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf(map[string]interface{}{
			"error": "hashSHA512 requires exactly 1 argument (data)",
		})
	}

	data := args[0].String()
	hash := sha512.Sum512([]byte(data))
	result := hex.EncodeToString(hash[:])

	if !silentMode {
		fmt.Printf("Go WASM: SHA512 hash generated for %d bytes\n", len(data))
	}

	return js.ValueOf(map[string]interface{}{
		"hash": result,
		"algorithm": "SHA512",
	})
}

// hashMD5 - Generate MD5 hash (for legacy support only)
func hashMD5(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf(map[string]interface{}{
			"error": "hashMD5 requires exactly 1 argument (data)",
		})
	}

	data := args[0].String()
	hash := md5.Sum([]byte(data))
	result := hex.EncodeToString(hash[:])

	if !silentMode {
		fmt.Printf("Go WASM: MD5 hash generated (WARNING: MD5 is cryptographically broken)\n")
	}

	return js.ValueOf(map[string]interface{}{
		"hash": result,
		"algorithm": "MD5",
		"warning": "MD5 is cryptographically broken and should not be used for security purposes",
	})
}

// generateAESKey - Generate a random AES key
func generateAESKey(this js.Value, args []js.Value) interface{} {
	keySize := 32 // Default to 256-bit key
	if len(args) > 0 {
		size := args[0].Int()
		if size == 16 || size == 24 || size == 32 {
			keySize = size
		}
	}

	key := make([]byte, keySize)
	_, err := rand.Read(key)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Failed to generate key: %v", err),
		})
	}

	if !silentMode {
		fmt.Printf("Go WASM: Generated %d-bit AES key\n", keySize*8)
	}

	return js.ValueOf(map[string]interface{}{
		"key": base64.StdEncoding.EncodeToString(key),
		"keySize": keySize * 8,
	})
}

// encryptAES - Encrypt data using AES-GCM
func encryptAES(this js.Value, args []js.Value) interface{} {
	if len(args) != 2 {
		return js.ValueOf(map[string]interface{}{
			"error": "encryptAES requires exactly 2 arguments (data, key)",
		})
	}

	data := args[0].String()
	keyStr := args[1].String()

	key, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Invalid key format: %v", err),
		})
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Failed to create cipher: %v", err),
		})
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Failed to create GCM: %v", err),
		})
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Failed to generate nonce: %v", err),
		})
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(data), nil)
	result := base64.StdEncoding.EncodeToString(ciphertext)

	if !silentMode {
		fmt.Printf("Go WASM: Encrypted %d bytes using AES-GCM\n", len(data))
	}

	return js.ValueOf(map[string]interface{}{
		"encryptedData": result,
		"algorithm": "AES-GCM",
	})
}

// decryptAES - Decrypt data using AES-GCM
func decryptAES(this js.Value, args []js.Value) interface{} {
	if len(args) != 2 {
		return js.ValueOf(map[string]interface{}{
			"error": "decryptAES requires exactly 2 arguments (encryptedData, key)",
		})
	}

	encryptedDataStr := args[0].String()
	keyStr := args[1].String()

	key, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Invalid key format: %v", err),
		})
	}

	encryptedData, err := base64.StdEncoding.DecodeString(encryptedDataStr)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Invalid encrypted data format: %v", err),
		})
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Failed to create cipher: %v", err),
		})
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Failed to create GCM: %v", err),
		})
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return js.ValueOf(map[string]interface{}{
			"error": "Encrypted data too short",
		})
	}

	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Failed to decrypt: %v", err),
		})
	}

	if !silentMode {
		fmt.Printf("Go WASM: Decrypted %d bytes using AES-GCM\n", len(plaintext))
	}

	return js.ValueOf(map[string]interface{}{
		"decryptedData": string(plaintext),
		"algorithm": "AES-GCM",
	})
}

// generateRSAKeyPair - Generate RSA key pair
func generateRSAKeyPair(this js.Value, args []js.Value) interface{} {
	keySize := 2048 // Default key size
	if len(args) > 0 {
		size := args[0].Int()
		if size >= 1024 && size <= 4096 {
			keySize = size
		}
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Failed to generate RSA key pair: %v", err),
		})
	}

	// Encode private key
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	privateKeyStr := string(pem.EncodeToMemory(privateKeyPEM))

	// Encode public key
	publicKeyPKIX, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Failed to marshal public key: %v", err),
		})
	}

	publicKeyPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyPKIX,
	}
	publicKeyStr := string(pem.EncodeToMemory(publicKeyPEM))

	if !silentMode {
		fmt.Printf("Go WASM: Generated %d-bit RSA key pair\n", keySize)
	}

	return js.ValueOf(map[string]interface{}{
		"publicKey":  publicKeyStr,
		"privateKey": privateKeyStr,
		"keySize":    keySize,
	})
}

// encryptRSA - Encrypt data using RSA public key
func encryptRSA(this js.Value, args []js.Value) interface{} {
	if len(args) != 2 {
		return js.ValueOf(map[string]interface{}{
			"error": "encryptRSA requires exactly 2 arguments (data, publicKey)",
		})
	}

	data := args[0].String()
	publicKeyStr := args[1].String()

	block, _ := pem.Decode([]byte(publicKeyStr))
	if block == nil {
		return js.ValueOf(map[string]interface{}{
			"error": "Failed to parse PEM block containing public key",
		})
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Failed to parse public key: %v", err),
		})
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return js.ValueOf(map[string]interface{}{
			"error": "Key is not an RSA public key",
		})
	}

	encryptedData, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPublicKey, []byte(data))
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Failed to encrypt: %v", err),
		})
	}

	result := base64.StdEncoding.EncodeToString(encryptedData)

	if !silentMode {
		fmt.Printf("Go WASM: Encrypted %d bytes using RSA\n", len(data))
	}

	return js.ValueOf(map[string]interface{}{
		"encryptedData": result,
		"algorithm": "RSA-PKCS1v15",
	})
}

// decryptRSA - Decrypt data using RSA private key
func decryptRSA(this js.Value, args []js.Value) interface{} {
	if len(args) != 2 {
		return js.ValueOf(map[string]interface{}{
			"error": "decryptRSA requires exactly 2 arguments (encryptedData, privateKey)",
		})
	}

	encryptedDataStr := args[0].String()
	privateKeyStr := args[1].String()

	encryptedData, err := base64.StdEncoding.DecodeString(encryptedDataStr)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Invalid encrypted data format: %v", err),
		})
	}

	block, _ := pem.Decode([]byte(privateKeyStr))
	if block == nil {
		return js.ValueOf(map[string]interface{}{
			"error": "Failed to parse PEM block containing private key",
		})
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Failed to parse private key: %v", err),
		})
	}

	decryptedData, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, encryptedData)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Failed to decrypt: %v", err),
		})
	}

	if !silentMode {
		fmt.Printf("Go WASM: Decrypted %d bytes using RSA\n", len(decryptedData))
	}

	return js.ValueOf(map[string]interface{}{
		"decryptedData": string(decryptedData),
		"algorithm": "RSA-PKCS1v15",
	})
}

// generateJWT - Generate a JWT token
func generateJWT(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return js.ValueOf(map[string]interface{}{
			"error": "generateJWT requires at least 2 arguments (payload, secret)",
		})
	}

	payloadStr := args[0].String()
	secret := args[1].String()
	
	expirationHours := 24 // Default 24 hours
	if len(args) > 2 {
		expirationHours = args[2].Int()
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Invalid payload JSON: %v", err),
		})
	}

	// Create claims
	claims := jwt.MapClaims{
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Duration(expirationHours) * time.Hour).Unix(),
		"iss": "crypto-wasm",
	}

	// Add payload claims
	for key, value := range payload {
		claims[key] = value
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Failed to sign token: %v", err),
		})
	}

	if !silentMode {
		fmt.Printf("Go WASM: Generated JWT token (expires in %d hours)\n", expirationHours)
	}

	return js.ValueOf(map[string]interface{}{
		"token": tokenString,
		"expiresIn": expirationHours * 3600, // seconds
		"algorithm": "HS256",
	})
}

// verifyJWT - Verify a JWT token
func verifyJWT(this js.Value, args []js.Value) interface{} {
	if len(args) != 2 {
		return js.ValueOf(map[string]interface{}{
			"error": "verifyJWT requires exactly 2 arguments (token, secret)",
		})
	}

	tokenString := args[0].String()
	secret := args[1].String()

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"valid": false,
			"error": fmt.Sprintf("Failed to parse token: %v", err),
		})
	}

	if !token.Valid {
		return js.ValueOf(map[string]interface{}{
			"valid": false,
			"error": "Token is invalid",
		})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return js.ValueOf(map[string]interface{}{
			"valid": false,
			"error": "Failed to extract claims",
		})
	}

	claimsJSON, _ := json.Marshal(claims)

	if !silentMode {
		fmt.Printf("Go WASM: JWT token verified successfully\n")
	}

	return js.ValueOf(map[string]interface{}{
		"valid": true,
		"claims": string(claimsJSON),
		"algorithm": "HS256",
	})
}

// bcryptHash - Hash password using bcrypt
func bcryptHash(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf(map[string]interface{}{
			"error": "bcryptHash requires at least 1 argument (password)",
		})
	}

	password := args[0].String()
	cost := bcrypt.DefaultCost // Default cost
	if len(args) > 1 {
		userCost := args[1].Int()
		if userCost >= bcrypt.MinCost && userCost <= bcrypt.MaxCost {
			cost = userCost
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Failed to hash password: %v", err),
		})
	}

	if !silentMode {
		fmt.Printf("Go WASM: Password hashed using bcrypt (cost: %d)\n", cost)
	}

	return js.ValueOf(map[string]interface{}{
		"hash": string(hashedPassword),
		"cost": cost,
		"algorithm": "bcrypt",
	})
}

// bcryptVerify - Verify password against bcrypt hash
func bcryptVerify(this js.Value, args []js.Value) interface{} {
	if len(args) != 2 {
		return js.ValueOf(map[string]interface{}{
			"error": "bcryptVerify requires exactly 2 arguments (password, hash)",
		})
	}

	password := args[0].String()
	hash := args[1].String()

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	valid := err == nil

	if !silentMode {
		fmt.Printf("Go WASM: Password verification: %t\n", valid)
	}

	result := map[string]interface{}{
		"valid": valid,
		"algorithm": "bcrypt",
	}

	if !valid && err != nil {
		result["error"] = err.Error()
	}

	return js.ValueOf(result)
}

// generateUUID - Generate a UUID v4
func generateUUID(this js.Value, args []js.Value) interface{} {
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Failed to generate UUID: %v", err),
		})
	}

	// Set version (4) and variant bits
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant bits

	uuidStr := fmt.Sprintf("%x-%x-%x-%x-%x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])

	if !silentMode {
		fmt.Printf("Go WASM: Generated UUID v4\n")
	}

	return js.ValueOf(map[string]interface{}{
		"uuid": uuidStr,
		"version": 4,
	})
}

// generateRandomBytes - Generate random bytes
func generateRandomBytes(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf(map[string]interface{}{
			"error": "generateRandomBytes requires exactly 1 argument (length)",
		})
	}

	length := args[0].Int()
	if length <= 0 || length > 1024 {
		return js.ValueOf(map[string]interface{}{
			"error": "Length must be between 1 and 1024",
		})
	}

	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Failed to generate random bytes: %v", err),
		})
	}

	result := base64.StdEncoding.EncodeToString(bytes)

	if !silentMode {
		fmt.Printf("Go WASM: Generated %d random bytes\n", length)
	}

	return js.ValueOf(map[string]interface{}{
		"bytes": result,
		"length": length,
		"encoding": "base64",
	})
}

// base64Encode - Encode data to base64
func base64Encode(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf(map[string]interface{}{
			"error": "base64Encode requires exactly 1 argument (data)",
		})
	}

	data := args[0].String()
	encoded := base64.StdEncoding.EncodeToString([]byte(data))

	if !silentMode {
		fmt.Printf("Go WASM: Encoded %d bytes to base64\n", len(data))
	}

	return js.ValueOf(map[string]interface{}{
		"encoded": encoded,
		"originalLength": len(data),
		"encodedLength": len(encoded),
	})
}

// base64Decode - Decode base64 data
func base64Decode(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf(map[string]interface{}{
			"error": "base64Decode requires exactly 1 argument (encodedData)",
		})
	}

	encodedData := args[0].String()
	decoded, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Failed to decode base64: %v", err),
		})
	}

	if !silentMode {
		fmt.Printf("Go WASM: Decoded %d bytes from base64\n", len(decoded))
	}

	return js.ValueOf(map[string]interface{}{
		"decoded": string(decoded),
		"encodedLength": len(encodedData),
		"decodedLength": len(decoded),
	})
}

// validatePasswordStrength - Validate password strength
func validatePasswordStrength(this js.Value, args []js.Value) interface{} {
	if len(args) != 1 {
		return js.ValueOf(map[string]interface{}{
			"error": "validatePasswordStrength requires exactly 1 argument (password)",
		})
	}

	password := args[0].String()
	score := 0
	issues := []string{}

	// Length check
	if len(password) >= 8 {
		score += 25
	} else {
		issues = append(issues, "Password should be at least 8 characters long")
	}

	// Lowercase check
	if strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz") {
		score += 25
	} else {
		issues = append(issues, "Password should contain lowercase letters")
	}

	// Uppercase check
	if strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		score += 25
	} else {
		issues = append(issues, "Password should contain uppercase letters")
	}

	// Number check
	if strings.ContainsAny(password, "0123456789") {
		score += 15
	} else {
		issues = append(issues, "Password should contain numbers")
	}

	// Special character check
	if strings.ContainsAny(password, "!@#$%^&*()_+-=[]{}|;:,.<>?") {
		score += 10
	} else {
		issues = append(issues, "Password should contain special characters")
	}

	var strength string
	switch {
	case score >= 90:
		strength = "very_strong"
	case score >= 70:
		strength = "strong"
	case score >= 50:
		strength = "medium"
	case score >= 30:
		strength = "weak"
	default:
		strength = "very_weak"
	}

	if !silentMode {
		fmt.Printf("Go WASM: Password strength evaluated: %s (%d/100)\n", strength, score)
	}

	return js.ValueOf(map[string]interface{}{
		"score": score,
		"strength": strength,
		"issues": issues,
		"valid": score >= 70,
	})
}

// getAvailableFunctions - Get list of available functions
func getAvailableFunctions(this js.Value, args []js.Value) interface{} {
	functions := []interface{}{
		"hashSHA256", "hashSHA512", "hashMD5",
		"generateAESKey", "encryptAES", "decryptAES",
		"generateRSAKeyPair", "encryptRSA", "decryptRSA",
		"generateJWT", "verifyJWT",
		"bcryptHash", "bcryptVerify",
		"generateUUID", "generateRandomBytes",
		"base64Encode", "base64Decode",
		"validatePasswordStrength",
		"getAvailableFunctions", "setSilentMode",
	}
	return js.ValueOf(functions)
}

func main() {
	// Create the crypto object
	crypto := js.Global().Get("Object").New()

	// Hash functions
	js.Global().Set("hashSHA256", js.FuncOf(hashSHA256))
	js.Global().Set("hashSHA512", js.FuncOf(hashSHA512))
	js.Global().Set("hashMD5", js.FuncOf(hashMD5))
	crypto.Set("hashSHA256", js.FuncOf(hashSHA256))
	crypto.Set("hashSHA512", js.FuncOf(hashSHA512))
	crypto.Set("hashMD5", js.FuncOf(hashMD5))

	// AES encryption
	js.Global().Set("generateAESKey", js.FuncOf(generateAESKey))
	js.Global().Set("encryptAES", js.FuncOf(encryptAES))
	js.Global().Set("decryptAES", js.FuncOf(decryptAES))
	crypto.Set("generateAESKey", js.FuncOf(generateAESKey))
	crypto.Set("encryptAES", js.FuncOf(encryptAES))
	crypto.Set("decryptAES", js.FuncOf(decryptAES))

	// RSA encryption
	js.Global().Set("generateRSAKeyPair", js.FuncOf(generateRSAKeyPair))
	js.Global().Set("encryptRSA", js.FuncOf(encryptRSA))
	js.Global().Set("decryptRSA", js.FuncOf(decryptRSA))
	crypto.Set("generateRSAKeyPair", js.FuncOf(generateRSAKeyPair))
	crypto.Set("encryptRSA", js.FuncOf(encryptRSA))
	crypto.Set("decryptRSA", js.FuncOf(decryptRSA))

	// JWT
	js.Global().Set("generateJWT", js.FuncOf(generateJWT))
	js.Global().Set("verifyJWT", js.FuncOf(verifyJWT))
	crypto.Set("generateJWT", js.FuncOf(generateJWT))
	crypto.Set("verifyJWT", js.FuncOf(verifyJWT))

	// Password hashing
	js.Global().Set("bcryptHash", js.FuncOf(bcryptHash))
	js.Global().Set("bcryptVerify", js.FuncOf(bcryptVerify))
	crypto.Set("bcryptHash", js.FuncOf(bcryptHash))
	crypto.Set("bcryptVerify", js.FuncOf(bcryptVerify))

	// Utilities
	js.Global().Set("generateUUID", js.FuncOf(generateUUID))
	js.Global().Set("generateRandomBytes", js.FuncOf(generateRandomBytes))
	js.Global().Set("base64Encode", js.FuncOf(base64Encode))
	js.Global().Set("base64Decode", js.FuncOf(base64Decode))
	js.Global().Set("validatePasswordStrength", js.FuncOf(validatePasswordStrength))
	crypto.Set("generateUUID", js.FuncOf(generateUUID))
	crypto.Set("generateRandomBytes", js.FuncOf(generateRandomBytes))
	crypto.Set("base64Encode", js.FuncOf(base64Encode))
	crypto.Set("base64Decode", js.FuncOf(base64Decode))
	crypto.Set("validatePasswordStrength", js.FuncOf(validatePasswordStrength))

	// Standard functions
	js.Global().Set("getAvailableFunctions", js.FuncOf(getAvailableFunctions))
	js.Global().Set("setSilentMode", js.FuncOf(setSilentMode))
	crypto.Set("getAvailableFunctions", js.FuncOf(getAvailableFunctions))
	crypto.Set("setSilentMode", js.FuncOf(setSilentMode))

	// Expose the crypto object globally
	js.Global().Set("crypto", crypto)

	// Signal that the module is ready
	fmt.Println("Go WASM Crypto module initialized")
	
	// Set a ready flag that can be checked by the loader (consistent with other modules)
	js.Global().Set("__gowm_ready", js.ValueOf(true))
	
	select {}
}
