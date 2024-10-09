package security

import (
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/chatApp/internal/pkg/config"
)

const issuer = "Chat-App-Server"

// jwtSecurityManager represents the JWT security manager
type jwtSecurityManager struct {
	cfg config.ChatApiConfig
}

// authClaims represents the claims in the auth token
type authClaims struct {
	TokenMetadata
	jwt.RegisteredClaims
}

// NewJwtSecurityManager creates a new JWT security manager
func NewJwtSecurityManager(cfg config.ChatApiConfig) Manager {
	return &jwtSecurityManager{
		cfg: cfg,
	}
}

// GenerateAuthToken generates an auth token for a user.
func (s jwtSecurityManager) GenerateAuthToken(metadata TokenMetadata) (token string, err error) {
	claims := &authClaims{
		TokenMetadata: metadata,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(s.cfg.AuthExpiryPeriod))),
			Issuer:    issuer,
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = t.SignedString([]byte(s.cfg.AuthSecret))
	if err != nil {
		return "", err
	}
	return token, nil
}
