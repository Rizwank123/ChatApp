package security

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// TokenMetadata represents the metadata in the auth token
type TokenMetadata struct {
	UserID         string `json:"user_id"`
	OrganizationID string `json:"organization_id"`
	Role           string `json:"role"`
}

// Manager defines the methods that a security manager should implement
type Manager interface {
	// GenerateAuthToken generates an auth token for a user.
	GenerateAuthToken(metadata TokenMetadata) (token string, err error)
}

func GetClaimsForContext(ctx echo.Context) jwt.MapClaims {
	token := ctx.Get("user")
	if token != nil {
		u := token.(*jwt.Token)
		claims := u.Claims.(jwt.MapClaims)
		return claims
	}
	return nil
}
