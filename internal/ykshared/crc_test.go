package ykshared

import (
	"fmt"
	"testing"
)

func TestCalculateCRC16(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected uint16
	}{
		{"empty data", []byte{}, 0xFFFF},
		{"single zero byte", []byte{0x00}, 0xE1F0},
		{"single byte 0x01", []byte{0x01}, 0xF1D1},
		{"known data sequence", []byte{0x12, 0x34, 0x56}, 0x12FD},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateCRC16(tt.data)
			if result != tt.expected {
				t.Errorf("CalculateCRC16(%v) = 0x%04X, expected 0x%04X", tt.data, result, tt.expected)
			}
		})
	}
}

func TestVerifyCRC16(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		expectedCRC uint16
		shouldMatch bool
	}{
		{"empty data with correct CRC", []byte{}, 0xFFFF, true},
		{"empty data with wrong CRC", []byte{}, 0x0000, false},
		{"single byte with correct CRC", []byte{0x01}, 0xF1D1, true},
		{"single byte with wrong CRC", []byte{0x01}, 0x0000, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VerifyCRC16(tt.data, tt.expectedCRC)
			if result != tt.shouldMatch {
				t.Errorf("VerifyCRC16(%v, 0x%04X) = %v, expected %v", tt.data, tt.expectedCRC, result, tt.shouldMatch)
			}
		})
	}
}

func TestCRCConsistency(t *testing.T) {
	t.Run("calculate and verify should match", func(t *testing.T) {
		testData := [][]byte{
			{},
			{0x00},
			{0x01, 0x02, 0x03},
			{0xFF, 0xFE, 0xFD},
			{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0},
		}

		for i, data := range testData {
			t.Run(fmt.Sprintf("data_set_%d", i), func(t *testing.T) {
				calculated := CalculateCRC16(data)
				verified := VerifyCRC16(data, calculated)
				if !verified {
					t.Errorf("CRC verification failed for data %v: calculated=0x%04X", data, calculated)
				}
			})
		}
	})
}

func TestCRCProperties(t *testing.T) {
	t.Run("different data produces different CRC", func(t *testing.T) {
		data1 := []byte{0x01, 0x02, 0x03}
		data2 := []byte{0x01, 0x02, 0x04}

		crc1 := CalculateCRC16(data1)
		crc2 := CalculateCRC16(data2)

		if crc1 == crc2 {
			t.Errorf("Different data produced same CRC: data1=%v, data2=%v, crc=0x%04X", data1, data2, crc1)
		}
	})

	t.Run("same data produces same CRC", func(t *testing.T) {
		data := []byte{0x01, 0x02, 0x03}

		crc1 := CalculateCRC16(data)
		crc2 := CalculateCRC16(data)

		if crc1 != crc2 {
			t.Errorf("Same data produced different CRC: crc1=0x%04X, crc2=0x%04X", crc1, crc2)
		}
	})

	t.Run("order matters", func(t *testing.T) {
		data1 := []byte{0x01, 0x02}
		data2 := []byte{0x02, 0x01}

		crc1 := CalculateCRC16(data1)
		crc2 := CalculateCRC16(data2)

		if crc1 == crc2 {
			t.Errorf("Different order produced same CRC: data1=%v, data2=%v, crc=0x%04X", data1, data2, crc1)
		}
	})
}
