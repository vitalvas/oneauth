package ykshared

// CalculateCRC16 calculates CRC-16 checksum using the YubiKey polynomial
// This matches the CRC implementation in the ksm/crypto package
func CalculateCRC16(data []byte) uint16 {
	const poly = 0x1021
	crc := uint16(0xFFFF)

	for _, b := range data {
		crc ^= uint16(b) << 8
		for i := 0; i < 8; i++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ poly
			} else {
				crc <<= 1
			}
		}
	}

	return crc
}

// VerifyCRC16 verifies if the CRC matches the expected value
func VerifyCRC16(data []byte, expectedCRC uint16) bool {
	return CalculateCRC16(data) == expectedCRC
}
