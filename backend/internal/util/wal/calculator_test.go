package wal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	mb1  = 1 * 1024 * 1024
	mb16 = 16 * 1024 * 1024
	mb64 = 64 * 1024 * 1024
)

// NextSegment — no wrap

func Test_WalCalculator_NextSegment_DefaultSegmentSize_NoWrap(t *testing.T) {
	calc := NewWalCalculator(mb16)

	cases := []struct {
		input    string
		expected string
	}{
		{"000000010000000100000000", "000000010000000100000001"},
		{"000000010000000100000027", "000000010000000100000028"},
		{"0000000100000001000000AB", "0000000100000001000000AC"},
		{"0000000100000001000000FE", "0000000100000001000000FF"},
	}

	for _, tc := range cases {
		result, err := calc.NextSegment(tc.input)
		require.NoError(t, err)
		assert.Equal(t, tc.expected, result, "input=%s", tc.input)
	}
}

// NextSegment — wrap at 0x100 (256) for 16 MB segments

func Test_WalCalculator_NextSegment_DefaultSegmentSize_WrapsAt256(t *testing.T) {
	calc := NewWalCalculator(mb16)

	// SS=0xFF → SS=0x00, LL++
	result, err := calc.NextSegment("0000000100000001000000FF")
	require.NoError(t, err)
	assert.Equal(t, "000000010000000200000000", result)

	result, err = calc.NextSegment("0000000200000005000000FF")
	require.NoError(t, err)
	assert.Equal(t, "000000020000000600000000", result)
}

// NextSegment — wrap at 0x1000 (4096) for 1 MB segments

func Test_WalCalculator_NextSegment_1MbSegmentSize_WrapsAt4096(t *testing.T) {
	calc := NewWalCalculator(mb1)

	// segmentsPerXLogId = 0x100000000 / (1*1024*1024) = 4096 = 0x1000
	// Last valid SS = 0x00000FFF

	result, err := calc.NextSegment("000000010000000100000FFE")
	require.NoError(t, err)
	assert.Equal(t, "000000010000000100000FFF", result)

	// wrap: SS=0x0FFF → SS=0, LL++
	result, err = calc.NextSegment("000000010000000100000FFF")
	require.NoError(t, err)
	assert.Equal(t, "000000010000000200000000", result)
}

// NextSegment — wrap at 0x40 (64) for 64 MB segments

func Test_WalCalculator_NextSegment_64MbSegmentSize_WrapsAt64(t *testing.T) {
	calc := NewWalCalculator(mb64)

	// segmentsPerXLogId = 0x100000000 / (64*1024*1024) = 64 = 0x40
	// Last valid SS = 0x0000003F

	result, err := calc.NextSegment("00000001000000010000003E")
	require.NoError(t, err)
	assert.Equal(t, "00000001000000010000003F", result)

	// wrap: SS=0x3F → SS=0, LL++
	result, err = calc.NextSegment("00000001000000010000003F")
	require.NoError(t, err)
	assert.Equal(t, "000000010000000200000000", result)
}

// NextSegment — log file increment on wrap

func Test_WalCalculator_NextSegment_IncrementsLog_OnSegmentWrap(t *testing.T) {
	calc := NewWalCalculator(mb16)

	// LL=0x00000001, wraps to LL=0x00000002
	result, err := calc.NextSegment("0000000100000001000000FF")
	require.NoError(t, err)
	assert.Equal(t, "000000010000000200000000", result)

	// LL=0x0000000F, wraps to LL=0x00000010
	result, err = calc.NextSegment("00000001000000FF000000FF")
	require.NoError(t, err)
	assert.Equal(t, "000000010000010000000000", result)
}

// NextSegment — timeline is preserved

func Test_WalCalculator_NextSegment_TimelinePreserved(t *testing.T) {
	calc := NewWalCalculator(mb16)

	result, err := calc.NextSegment("000000030000000100000005")
	require.NoError(t, err)
	assert.Equal(t, "000000030000000100000006", result)
}

// NextSegment — invalid input

func Test_WalCalculator_NextSegment_InvalidName_ReturnsError(t *testing.T) {
	calc := NewWalCalculator(mb16)

	cases := []string{
		"",
		"00000001000000010000000",   // 23 chars
		"0000000100000001000000001", // 25 chars
		"00000001000000010000000G",  // non-hex char
		"short",
	}

	for _, name := range cases {
		_, err := calc.NextSegment(name)
		assert.Error(t, err, "expected error for input %q", name)
	}
}

// IsValidSegmentName

func Test_WalCalculator_IsValidSegmentName_ValidName_ReturnsTrue(t *testing.T) {
	calc := NewWalCalculator(mb16)

	valid := []string{
		"000000010000000100000000",
		"0000000100000001000000FF",
		"FFFFFFFFFFFFFFFFFFFFFFFF",
		"000000010000000100000027",
		"0000000200000005000000AB",
	}

	for _, name := range valid {
		assert.True(t, calc.IsValidSegmentName(name), "expected valid for %q", name)
	}
}

func Test_WalCalculator_IsValidSegmentName_TooShort_ReturnsFalse(t *testing.T) {
	calc := NewWalCalculator(mb16)

	assert.False(t, calc.IsValidSegmentName("00000001000000010000000")) // 23 chars
	assert.False(t, calc.IsValidSegmentName(""))
}

func Test_WalCalculator_IsValidSegmentName_TooLong_ReturnsFalse(t *testing.T) {
	calc := NewWalCalculator(mb16)

	assert.False(t, calc.IsValidSegmentName("0000000100000001000000001")) // 25 chars
}

func Test_WalCalculator_IsValidSegmentName_NonHex_ReturnsFalse(t *testing.T) {
	calc := NewWalCalculator(mb16)

	assert.False(t, calc.IsValidSegmentName("00000001000000010000000G"))
	assert.False(t, calc.IsValidSegmentName("00000001000000010000000Z"))
	assert.False(t, calc.IsValidSegmentName("000000010000000100000 00"))
}

// Compare

func Test_WalCalculator_Compare_ReturnsCorrectOrdering(t *testing.T) {
	calc := NewWalCalculator(mb16)

	cases := []struct {
		a        string
		b        string
		expected int
	}{
		// equal
		{"000000010000000100000001", "000000010000000100000001", 0},
		// segment ordering
		{"000000010000000100000001", "000000010000000100000002", -1},
		{"000000010000000100000002", "000000010000000100000001", 1},
		// log ordering
		{"000000010000000100000000", "000000010000000200000000", -1},
		{"000000010000000200000000", "000000010000000100000000", 1},
		// timeline ordering
		{"000000010000000100000000", "000000020000000100000000", -1},
		{"000000020000000100000000", "000000010000000100000000", 1},
		// across wrap boundary: log 1, seg 255 < log 2, seg 0
		{"0000000100000001000000FF", "000000010000000200000000", -1},
	}

	for _, tc := range cases {
		result, err := calc.Compare(tc.a, tc.b)
		require.NoError(t, err)
		assert.Equal(t, tc.expected, result, "Compare(%s, %s)", tc.a, tc.b)
	}
}

func Test_WalCalculator_Compare_InvalidInput_ReturnsError(t *testing.T) {
	calc := NewWalCalculator(mb16)

	_, err := calc.Compare("invalid", "000000010000000100000001")
	assert.Error(t, err)

	_, err = calc.Compare("000000010000000100000001", "invalid")
	assert.Error(t, err)
}

// Default segment size via zero input

func Test_WalCalculator_NewWalCalculator_ZeroSize_UsesDefault16MB(t *testing.T) {
	calc := NewWalCalculator(0)
	assert.Equal(t, int64(mb16), calc.segmentSizeBytes)
	assert.Equal(t, uint64(256), calc.segmentsPerXLogId)
}
