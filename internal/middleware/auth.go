package middleware

import (
	"github.com/cam-boltnote/go-ignite/internal/models"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))

type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func GenerateToken(user *models.User) (string, error) {
	// Create claims with user data and expiration time
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Generate token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func validateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := validateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Set user claims in context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Next()
	}
}

// DecryptPassword decrypts an encrypted password using AES-256 encryption
func DecryptPassword(encryptedPassword string) (string, error) {
	// Get encryption key from environment and decode from base64
	encodedKey := os.Getenv("ENCRYPTION_KEY")
	key, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 key: %v", err)
	}

	if len(key) != 32 {
		return "", fmt.Errorf("encryption key must be 32 bytes for AES-256 (got %d bytes)", len(key))
	}

	// Decode base64 encrypted password
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedPassword)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 string: %v", err)
	}

	// Create cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher block: %v", err)
	}

	// Extract IV from ciphertext
	if len(ciphertext) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	// Create decrypter
	stream := cipher.NewCFBDecrypter(block, iv)

	// Decrypt the ciphertext
	plaintext := make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)

	return string(plaintext), nil
}

// EncryptPassword encrypts a password using AES-256 encryption
func EncryptPassword(password string) (string, error) {
	// Get encryption key from environment and decode from base64
	encodedKey := os.Getenv("ENCRYPTION_KEY")
	key, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 key: %v", err)
	}

	fmt.Printf("Key length: %d bytes\n", len(key)) // Debug line
	if len(key) != 32 {
		return "", fmt.Errorf("encryption key must be 32 bytes for AES-256 (got %d bytes)", len(key))
	}

	// Create cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher block: %v", err)
	}

	// Create IV (Initialization Vector)
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return "", fmt.Errorf("failed to generate IV: %v", err)
	}

	// Create encrypter
	stream := cipher.NewCFBEncrypter(block, iv)

	// Encrypt the password
	ciphertext := make([]byte, len(password))
	stream.XORKeyStream(ciphertext, []byte(password))

	// Combine IV and ciphertext
	fullCiphertext := make([]byte, len(iv)+len(ciphertext))
	copy(fullCiphertext, iv)
	copy(fullCiphertext[len(iv):], ciphertext)

	// Convert to base64
	encodedStr := base64.StdEncoding.EncodeToString(fullCiphertext)
	return encodedStr, nil
}
