package ykshared

import (
	"testing"
)

// Benchmark CRC functions
func BenchmarkCalculateCRC16(b *testing.B) {
	data := []byte{0x87, 0x92, 0xeb, 0xfe, 0x26, 0xcc, 0x01, 0x00, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CalculateCRC16(data)
	}
}

func BenchmarkCalculateCRC16Small(b *testing.B) {
	data := []byte{0x01, 0x02, 0x03, 0x04}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CalculateCRC16(data)
	}
}

func BenchmarkCalculateCRC16Large(b *testing.B) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CalculateCRC16(data)
	}
}

func BenchmarkVerifyCRC16(b *testing.B) {
	data := []byte{0x87, 0x92, 0xeb, 0xfe, 0x26, 0xcc, 0x01, 0x00, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc}
	expectedCRC := CalculateCRC16(data)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = VerifyCRC16(data, expectedCRC)
	}
}

// Benchmark modhex validation functions
func BenchmarkIsValidModhex(b *testing.B) {
	modhex := "ccccccccccccdefghijklnrtuvcbdefghijklnrtuvic"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsValidModhex(modhex)
	}
}

func BenchmarkValidateModhexString(b *testing.B) {
	modhex := "ccccccccccccdefghijklnrtuvcbdefghijklnrtuvic"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateModhexString(modhex)
	}
}

func BenchmarkValidateKeyIDFormat(b *testing.B) {
	keyID := "cccccccccccc"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateKeyIDFormat(keyID)
	}
}

// Benchmark modhex conversion functions
func BenchmarkModhexToHex(b *testing.B) {
	modhex := "ccccccccccccdefghijklnrtuvcbdefghijklnrtuvic"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ModhexToHex(modhex)
	}
}

func BenchmarkModhexToHexKeyID(b *testing.B) {
	keyID := "cccccccccccc"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ModhexToHex(keyID)
	}
}

func BenchmarkModhexToInt(b *testing.B) {
	keyID := "cccccccccccc"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ModhexToInt(keyID)
	}
}

func BenchmarkHexToModhex(b *testing.B) {
	hex := "000000000000234567890abcde234567890abcde70"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = HexToModhex(hex)
	}
}

func BenchmarkBytesToModhex(b *testing.B) {
	data := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x23, 0x45, 0x67, 0x89, 0x0a, 0xbc, 0xde, 0x23, 0x45, 0x67, 0x89, 0x0a, 0xbc, 0xde, 0x70}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BytesToModhex(data)
	}
}

func BenchmarkModhexToBytes(b *testing.B) {
	modhex := "ccccccccccccdefghijklnrtuvcbdefghijklnrtuvic"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ModhexToBytes(modhex)
	}
}

// Benchmark OTP validation functions
func BenchmarkValidateOTP(b *testing.B) {
	otp := "ccccccccccccdefghijklnrtuvcbdefghijklnrtuvic"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ValidateOTP(otp)
	}
}

func BenchmarkExtractKeyID(b *testing.B) {
	otp := "ccccccccccccdefghijklnrtuvcbdefghijklnrtuvic"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ExtractKeyID(otp)
	}
}

// Benchmark random generation functions
func BenchmarkGenerateRandomModhex(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GenerateRandomModhex(44)
	}
}

func BenchmarkGenerateKeyID(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GenerateKeyID()
	}
}

// Benchmark different data sizes for CRC
func BenchmarkCalculateCRC16_1Byte(b *testing.B) {
	data := []byte{0x42}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CalculateCRC16(data)
	}
}

func BenchmarkCalculateCRC16_16Bytes(b *testing.B) {
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CalculateCRC16(data)
	}
}

func BenchmarkCalculateCRC16_32Bytes(b *testing.B) {
	data := make([]byte, 32)
	for i := range data {
		data[i] = byte(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CalculateCRC16(data)
	}
}

// Benchmark modhex validation with different string lengths
func BenchmarkIsValidModhex_Short(b *testing.B) {
	modhex := "cccccc"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsValidModhex(modhex)
	}
}

func BenchmarkIsValidModhex_KeyID(b *testing.B) {
	keyID := "cccccccccccc"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsValidModhex(keyID)
	}
}

func BenchmarkIsValidModhex_FullOTP(b *testing.B) {
	otp := "ccccccccccccdefghijklnrtuvcbdefghijklnrtuvic"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsValidModhex(otp)
	}
}

// Benchmark invalid input handling
func BenchmarkIsValidModhex_Invalid(b *testing.B) {
	invalid := "ccccccccccccdefghijklnrtuvcbdefghijklnrtuXXX"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsValidModhex(invalid)
	}
}

func BenchmarkValidateOTP_Invalid(b *testing.B) {
	invalidOTP := "shortOTP"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ValidateOTP(invalidOTP)
	}
}
