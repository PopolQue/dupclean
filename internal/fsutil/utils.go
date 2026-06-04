package fsutil

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// FormatBytes returns a human-readable string representation of bytes.
func FormatBytes(b int64) string {
	if b < 0 {
		return "n/a"
	}
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}

	units := []string{"KB", "MB", "GB", "TB", "PB", "EB"}
	div := float64(unit)
	exp := 0
	for n := float64(b) / div; n >= unit && exp < len(units)-1; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %s", float64(b)/div, units[exp])
}

// ParseDuration wraps time.ParseDuration to support days ('d').
func ParseDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, nil
	}
	if strings.HasSuffix(s, "d") {
		daysStr := strings.TrimSuffix(s, "d")
		days, err := strconv.Atoi(daysStr)
		if err != nil {
			return 0, fmt.Errorf("parsing days duration %q: %w", daysStr, err)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}
