package gobuild

import "fmt"

// BinarySizer formats binary sizes in human-readable format
type BinarySizer struct {
	getBinary func() []byte
	log       func(...any)
}

// NewBinarySizer creates a new BinarySizer instance
func NewBinarySizer(getBinary func() []byte) *BinarySizer {
	return &BinarySizer{
		getBinary: getBinary,
		log:       func(...any) {}, // No-op by default
	}
}

// SetLog sets the logging function
func (b *BinarySizer) SetLog(f func(...any)) {
	if f != nil {
		b.log = f
	}
}

// BinarySize returns the binary size in human-readable format
// Returns format: "10.4 KB", "2.3 MB", "1.5 GB"
// Returns "0.0 KB" if binary is unavailable or empty
func (b *BinarySizer) BinarySize() string {
	if b.getBinary == nil {
		return "0.0 KB"
	}

	bytes := b.getBinary()
	if len(bytes) == 0 {
		return "0.0 KB"
	}

	size := float64(len(bytes))

	// Thresholds
	const (
		KB = 1024
		MB = 1024 * 1024
		GB = 1024 * 1024 * 1024
	)

	// Determine unit based on size
	if size >= GB {
		return fmt.Sprintf("%.1f GB", size/GB)
	} else if size >= MB {
		return fmt.Sprintf("%.1f MB", size/MB)
	} else {
		// Everything below MB is shown in KB (minimum 0.1 KB)
		return fmt.Sprintf("%.1f KB", size/KB)
	}
}
