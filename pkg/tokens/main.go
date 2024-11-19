package tokens

import (
	"log"
	"strings"
	"time"

	config "github.com/devbenho/luka-platform/configs"
	"github.com/devbenho/luka-platform/internal/utils"
	"github.com/golang-jwt/jwt/v5"
)

// JWTPayload represents the payload stored in a JWT token
type JWTPayload struct {
	userId string
	role   string
	email  string
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
func GenerateAccessToken(payload map[string]interface{}) string {
	cfg := config.GetConfig()
	payload["type"] = "access"
	tokenContent := jwt.MapClaims{
		"payload": payload,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	jwtToken := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tokenContent)
	token, err := jwtToken.SignedString([]byte(cfg.JWT.Secret))
	if err != nil {
		log.Fatal("Failed to generate access token: ", err)
		return ""
	}

	return token
}

func ValidateToken(jwtToken string) (map[string]interface{}, error) {
	cfg := config.GetConfig()
	cleanJWT := strings.Replace(jwtToken, "Bearer ", "", -1)
	tokenData := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(cleanJWT, tokenData, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrInvalidKey
	}

	var data map[string]interface{}
	utils.Copy(&data, tokenData["payload"])

	return data, nil
}
