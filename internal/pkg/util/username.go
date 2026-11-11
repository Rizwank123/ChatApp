package util

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	ErrUsernameTooShort     = errors.New("username must be at least 3 characters long")
	ErrUsernameTooLong      = errors.New("username cannot exceed 30 characters")
	ErrUsernameInvalidChars = errors.New("username can only contain lowercase letters, numbers, underscores, and periods")
	ErrUsernameInvalidEdge  = errors.New("username cannot start or end with a period or underscore")
	ErrUsernameConsecutive  = errors.New("username cannot contain consecutive periods or underscores")
	ErrUsernameReserved     = errors.New("this username is reserved and cannot be used")
)

var (
	validUsernamePattern = regexp.MustCompile(`^[a-z0-9_.]+$`)
	reservedUsernames    = map[string]struct{}{
		"admin":         {},
		"administrator": {},
		"root":          {},
		"support":       {},
		"system":        {},
		"api":           {},
		"null":          {},
		"undefined":     {},
		"moderator":     {},
		"help":          {},
		"vibely":        {},
		"chatapp":       {},
		"bot":           {},
		"security":      {},
		"staff":         {},
		"auth":          {},
		"login":         {},
		"register":      {},
		"signup":        {},
		"explore":       {},
		"messages":      {},
		"profile":       {},
		"settings":      {},
		"feed":          {},
		"status":        {},
	}
)

// NormalizeUsername strips leading/trailing whitespace and converts to lowercase.
func NormalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

// ValidateUsername checks if the given username meets all system constraints.
func ValidateUsername(rawUsername string) (string, error) {
	normalized := NormalizeUsername(rawUsername)

	if len(normalized) < 3 {
		return "", ErrUsernameTooShort
	}
	if len(normalized) > 30 {
		return "", ErrUsernameTooLong
	}

	if strings.HasPrefix(normalized, ".") || strings.HasPrefix(normalized, "_") ||
		strings.HasSuffix(normalized, ".") || strings.HasSuffix(normalized, "_") {
		return "", ErrUsernameInvalidEdge
	}

	if strings.Contains(normalized, "..") || strings.Contains(normalized, "__") ||
		strings.Contains(normalized, "._") || strings.Contains(normalized, "_.") {
		return "", ErrUsernameConsecutive
	}

	if !validUsernamePattern.MatchString(normalized) {
		return "", ErrUsernameInvalidChars
	}

	if _, reserved := reservedUsernames[normalized]; reserved {
		return "", fmt.Errorf("%w: %s", ErrUsernameReserved, normalized)
	}

	return normalized, nil
}
