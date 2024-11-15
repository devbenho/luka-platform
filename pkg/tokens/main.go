package tokens

import (
	"time"

	"github.com/devbenho/luka-platform/internal/user/models"
	"github.com/golang-jwt/jwt/v5"
)

// JWTPayload represents the payload stored in JWT token
type JWTPayload struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

// TokenService provides methods for generating and parsing JWT tokens
type TokenService struct {
	secret string
}

// NewTokenService creates a new TokenService with the given secret key
func NewTokenService(secret string) *TokenService {
	return &TokenService{
		secret: secret,
	}
}

// GenerateToken generates a JWT token for the provided payload
func (s *TokenService) GenerateToken(payload JWTPayload) (models.Token, error) {
	claims := jwt.MapClaims{
		"username": payload.Username,
		"role":     payload.Role,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // Set token expiration
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.secret))
	if err != nil {
		return models.Token{}, err
	}

	return models.Token{Token: signedToken}, nil
}

// ParseToken parses a JWT token and returns the payload
func (s *TokenService) ParseToken(tokenString string) (*JWTPayload, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is correct
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(s.secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrInvalidKey
	}

	return &JWTPayload{
		Username: claims["username"].(string),
		Role:     claims["role"].(string),
	}, nil
}
