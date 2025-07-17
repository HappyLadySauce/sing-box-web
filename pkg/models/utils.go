package models

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
)

// generateUUID generates a new UUID string
func generateUUID() string {
	return uuid.New().String()
}

// generateToken generates a random token of specified length
func generateToken(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to UUID if random generation fails
		return uuid.New().String()
	}
	return hex.EncodeToString(bytes)
}

// FormatBytes formats bytes to human readable string
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatSpeed formats bytes per second to human readable string
func FormatSpeed(bytesPerSec int64) string {
	const unit = 1024
	if bytesPerSec < unit {
		return fmt.Sprintf("%d B/s", bytesPerSec)
	}
	div, exp := int64(unit), 0
	for n := bytesPerSec / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB/s", float64(bytesPerSec)/float64(div), "KMGTPE"[exp])
}

// ParseBytes parses human readable bytes string to int64
func ParseBytes(s string) (int64, error) {
	// This is a simplified implementation
	// In production, you might want to use a library like dustin/go-humanize
	var value float64
	var unit string
	
	n, err := fmt.Sscanf(s, "%f %s", &value, &unit)
	if err != nil || n != 2 {
		return 0, fmt.Errorf("invalid format: %s", s)
	}
	
	switch unit {
	case "B":
		return int64(value), nil
	case "KB":
		return int64(value * 1024), nil
	case "MB":
		return int64(value * 1024 * 1024), nil
	case "GB":
		return int64(value * 1024 * 1024 * 1024), nil
	case "TB":
		return int64(value * 1024 * 1024 * 1024 * 1024), nil
	default:
		return 0, fmt.Errorf("unknown unit: %s", unit)
	}
}