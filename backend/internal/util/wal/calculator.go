package wal

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

const (
	segmentNameLen          = 24
	timelineLen             = 8
	logLen                  = 8
	segLen                  = 8
	defaultSegmentSizeBytes = 16 * 1024 * 1024 // 16 MB
)

// WalCalculator performs WAL segment name arithmetic for a given WAL segment size.
//
// WAL segment name format: TTTTTTTTLLLLLLLLSSSSSSSS (24 hex chars)
//   - TT: timeline (8 hex digits)
//   - LL: log file / XLogId (8 hex digits)
//   - SS: segment within log file (8 hex digits)
//
// segmentsPerXLogId = 0x100000000 / segmentSizeBytes
// Increment SS; if SS >= segmentsPerXLogId → SS = 0, LL++
type WalCalculator struct {
	segmentSizeBytes  int64
	segmentsPerXLogId uint64
}

// NewWalCalculator creates a WalCalculator for the given WAL segment size in bytes.
// Pass 0 or a negative value to use the PostgreSQL default of 16 MB.
func NewWalCalculator(segmentSizeBytes int64) *WalCalculator {
	if segmentSizeBytes <= 0 {
		segmentSizeBytes = defaultSegmentSizeBytes
	}

	return &WalCalculator{
		segmentSizeBytes:  segmentSizeBytes,
		segmentsPerXLogId: uint64(0x100000000) / uint64(segmentSizeBytes),
	}
}

// NextSegment computes the next WAL segment name after current.
// Returns an error if current is not a valid 24-character hex WAL segment name.
func (c *WalCalculator) NextSegment(current string) (string, error) {
	if !c.IsValidSegmentName(current) {
		return "", fmt.Errorf("invalid WAL segment name: %q", current)
	}

	timeline := current[:timelineLen]
	logHex := current[timelineLen : timelineLen+logLen]
	segHex := current[timelineLen+logLen:]

	logVal, err := parseHex32(logHex)
	if err != nil {
		return "", fmt.Errorf("parse log part of %q: %w", current, err)
	}

	segVal, err := parseHex32(segHex)
	if err != nil {
		return "", fmt.Errorf("parse seg part of %q: %w", current, err)
	}

	segVal++
	if uint64(segVal) >= c.segmentsPerXLogId {
		segVal = 0
		logVal++
	}

	return fmt.Sprintf("%s%08X%08X", strings.ToUpper(timeline), logVal, segVal), nil
}

// IsValidSegmentName returns true if name is a 24-character uppercase (or lowercase) hex string
// representing a valid WAL segment name.
func (c *WalCalculator) IsValidSegmentName(name string) bool {
	if len(name) != segmentNameLen {
		return false
	}

	_, err := hex.DecodeString(name)
	return err == nil
}

// Compare compares two WAL segment names a and b by their numeric value.
// Returns -1 if a < b, 0 if a == b, 1 if a > b.
// Both names must be valid; returns an error otherwise.
// Timeline is compared first, then log file, then segment number.
func (c *WalCalculator) Compare(a, b string) (int, error) {
	if !c.IsValidSegmentName(a) {
		return 0, fmt.Errorf("invalid WAL segment name: %q", a)
	}

	if !c.IsValidSegmentName(b) {
		return 0, fmt.Errorf("invalid WAL segment name: %q", b)
	}

	// Fixed-width uppercase hex: lexicographic order equals numeric order.
	aUpper := strings.ToUpper(a)
	bUpper := strings.ToUpper(b)

	if aUpper < bUpper {
		return -1, nil
	}

	if aUpper > bUpper {
		return 1, nil
	}

	return 0, nil
}

func parseHex32(s string) (uint32, error) {
	if len(s) != 8 {
		return 0, errors.New("expected 8 hex characters")
	}

	b, err := hex.DecodeString(s)
	if err != nil {
		return 0, err
	}

	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3]), nil
}
