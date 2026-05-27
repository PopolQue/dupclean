package fsutil

import (
	"fmt"
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
