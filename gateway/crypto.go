package gateway

import "encoding/binary"

type Key struct {
	Value uint32
	Buff  [4]byte
}

func DeriveKey(clientSalt, serverSalt uint32) Key {
	value := (clientSalt ^ serverSalt) + clientSalt + 8654
	var buff [4]byte
	binary.LittleEndian.PutUint32(buff[:], value)
	return Key{Value: value, Buff: buff}
}

func (k Key) Bytes() []byte {
	buff := make([]byte, 4)
	binary.LittleEndian.PutUint32(buff, k.Value)
	return buff
}

func CheckKey(clientSalt, serverSalt uint32) uint16 {
	return CRC16(DeriveKey(clientSalt, serverSalt).Bytes())
}

func XORBytes(data []byte, key [4]byte) {
	xorBytes(data, key, 0, len(data))
}

func xorBytes(data []byte, key [4]byte, offset, length int) {
	if offset >= len(data) || length <= 0 {
		return
	}
	end := offset + length
	if end > len(data) {
		end = len(data)
	}
	for i := offset; i < end; i++ {
		data[i] ^= key[i&3]
	}
}
