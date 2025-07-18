package auth

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	configv1 "sing-box-web/pkg/config/v1"
)

// Claims represents the JWT claims structure
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	NodeID   string `json:"node_id,omitempty"`
	jwt.RegisteredClaims
}

// UserRepository interface for user operations
type UserRepository interface {
	GetByID(id uint) (*User, error)
}

// User represents basic user info for JWT
type User struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// JWTManager manages JWT token operations
type JWTManager struct {
	config   configv1.AuthConfig
	logger   *zap.Logger
	userRepo UserRepository
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(config configv1.AuthConfig, logger *zap.Logger) *JWTManager {
	return &JWTManager{
		config:   config,
		logger:   logger,
		userRepo: nil, // Will be set via dependency injection
	}
}

// SetUserRepository sets the user repository for database operations
func (j *JWTManager) SetUserRepository(userRepo UserRepository) {
	j.userRepo = userRepo
}

// GenerateToken generates a JWT token for a user
func (j *JWTManager) GenerateToken(userID, username, role string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(j.config.JWTExpiration)

	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "sing-box-web",
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.config.JWTSecret))
	if err != nil {
		j.logger.Error("Failed to sign JWT token", zap.Error(err))
		return "", err
	}

	j.logger.Debug("Generated JWT token", zap.String("user_id", userID), zap.String("username", username))
	return tokenString, nil
}

// GenerateRefreshToken generates a refresh token
func (j *JWTManager) GenerateRefreshToken(userID string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(j.config.RefreshExpiration)

	claims := jwt.RegisteredClaims{
		Issuer:    "sing-box-web",
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		NotBefore: jwt.NewNumericDate(now),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.config.JWTSecret))
	if err != nil {
		j.logger.Error("Failed to sign refresh token", zap.Error(err))
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(j.config.JWTSecret), nil
	})

	if err != nil {
		j.logger.Warn("Failed to parse JWT token", zap.Error(err))
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		j.logger.Warn("Invalid JWT token claims")
		return nil, errors.New("invalid token claims")
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time) {
		j.logger.Warn("JWT token expired", zap.String("user_id", claims.UserID))
		return nil, errors.New("token expired")
	}

	return claims, nil
}

// RefreshToken validates a refresh token and generates a new access token
func (j *JWTManager) RefreshToken(refreshToken string) (string, error) {
	token, err := jwt.ParseWithClaims(refreshToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(j.config.JWTSecret), nil
	})

	if err != nil {
		j.logger.Warn("Failed to parse refresh token", zap.Error(err))
		return "", err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		j.logger.Warn("Invalid refresh token claims")
		return "", errors.New("invalid refresh token")
	}

	// Check if refresh token is expired
	if claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time) {
		j.logger.Warn("Refresh token expired", zap.String("user_id", claims.Subject))
		return "", errors.New("refresh token expired")
	}

	// Get user details from database to generate new token
	userID := claims.Subject
	userIDUint, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		j.logger.Warn("Invalid user ID in refresh token", zap.String("user_id", userID))
		return "", errors.New("invalid user ID")
	}

	if j.userRepo != nil {
		user, err := j.userRepo.GetByID(uint(userIDUint))
		if err != nil {
			j.logger.Warn("User not found for refresh token", zap.String("user_id", userID))
			return "", errors.New("user not found")
		}
		return j.GenerateToken(userID, user.Username, user.Role)
	}

	// Fallback if repository not set
	j.logger.Warn("UserRepository not set, using placeholder values")
	return j.GenerateToken(userID, "user", "user")
}

// RevokeToken adds a token to the revocation list
func (j *JWTManager) RevokeToken(tokenString string) error {
	// TODO: Implement token revocation using Redis or database
	// For now, we'll just log the revocation
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return err
	}

	j.logger.Info("Token revoked", zap.String("user_id", claims.UserID), zap.String("username", claims.Username))
	return nil
}

// IsTokenRevoked checks if a token is revoked
func (j *JWTManager) IsTokenRevoked(tokenString string) bool {
	// TODO: Check token revocation list
	// For now, we'll return false (not revoked)
	return false
}