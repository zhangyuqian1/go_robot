package gateway

const crc16Polynomial uint32 = 0x1021

var crc16Table = makeCRC16Table()

func CRC16(data []byte) uint16 {
	var crc uint32
	for _, b := range data {
		index := (reflectBits(uint32(b), 8) ^ (crc >> 8)) & 0xFF
		crc = crc16Table[index] ^ ((crc << 8) & 0xFFFFFF00)
	}
	return uint16(reflectBits(crc, 16) & 0xFFFF)
}

func makeCRC16Table() [256]uint32 {
	var table [256]uint32
	for i := uint32(0); i < 256; i++ {
		crc := (i << 8) & 0xFFFFFF00
		for j := 0; j < 8; j++ {
			if crc&0x8000 != 0 {
				crc = ((crc << 1) & 0xFFFFFFFE) ^ crc16Polynomial
			} else {
				crc = (crc << 1) & 0xFFFFFFFE
			}
		}
		table[i] = crc
	}
	return table
}

func reflectBits(input uint32, bits uint) uint32 {
	var result uint32
	for i := uint(0); i < bits; i++ {
		if input&1 != 0 {
			result |= 1 << (bits - 1 - i)
		}
		input >>= 1
	}
	return result
}
