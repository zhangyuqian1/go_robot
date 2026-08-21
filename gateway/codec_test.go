package gateway

import (
	"encoding/binary"
	"errors"
	"testing"
)

func makeGatewayBody(sys, msgID byte, protoBytes []byte) []byte {
	body := make([]byte, 2+2+len(protoBytes))
	body[0] = sys
	body[1] = msgID
	binary.LittleEndian.PutUint16(body[2:4], uint16(len(protoBytes)))
	copy(body[4:], protoBytes)
	return body
}

func makeMrRobotPacket(key [4]byte, body []byte, headerSeq uint16) []byte {
	return makeMrRobotPacketWithModes(key, body, headerSeq, true, false)
}

func makeMrRobotPacketWithBodyMode(key [4]byte, body []byte, headerSeq uint16, encryptBody bool) []byte {
	return makeMrRobotPacketWithModes(key, body, headerSeq, true, encryptBody)
}

func makeMrRobotPacketWithModes(key [4]byte, body []byte, headerSeq uint16, encryptHeader bool, encryptBody bool) []byte {
	return makeMrRobotPacketWithBodyCRC(key, body, headerSeq, CRC16(body), encryptHeader, encryptBody)
}
func makeMrRobotPacketWithBodyCRC(key [4]byte, body []byte, headerSeq uint16, bodyCRC uint16, encryptHeader bool, encryptBody bool) []byte {
	packet := make([]byte, HeaderLen+len(body))
	binary.LittleEndian.PutUint16(packet[0:2], mrRobotHeaderMagic)
	binary.LittleEndian.PutUint16(packet[2:4], uint16(len(body)))
	binary.LittleEndian.PutUint16(packet[4:6], bodyCRC)
	binary.LittleEndian.PutUint16(packet[6:8], headerSeq)
	binary.LittleEndian.PutUint16(packet[6:8], CRC16(packet[:HeaderLen]))
	if encryptHeader {
		xorBytes(packet, key, 4, 4)
	}
	copy(packet[HeaderLen:], body)
	if encryptBody {
		xorBytes(packet, key, HeaderLen, len(body))
	}
	return packet
}
func TestDeriveKeyMatchesMrRobotReference(t *testing.T) {
	key := DeriveKey(0x01020304, 0x05060708)
	if key.Value != 0x050628DE {
		t.Fatalf("key value = %#x, want %#x", key.Value, uint32(0x050628DE))
	}
	wantBuff := [4]byte{0xDE, 0x28, 0x06, 0x05}
	if key.Buff != wantBuff {
		t.Fatalf("key buff = %#v, want %#v", key.Buff, wantBuff)
	}
	if got, want := CRC16(key.Bytes()), uint16(0x90BE); got != want {
		t.Fatalf("check crc = %#x, want %#x", got, want)
	}
}

func TestCodecMarshalUnmarshalRoundTrip(t *testing.T) {
	codec := NewCodec([4]byte{0xDE, 0x28, 0x06, 0x05})
	protoBytes := []byte{0x08, 0x96, 0x01, 0x12, 0x03, 'b', 'o', 'b'}

	packet, err := codec.MarshalRequest(255, 1, protoBytes)
	if err != nil {
		t.Fatalf("MarshalRequest returned error: %v", err)
	}
	if len(packet) != HeaderLen+2+len(protoBytes)+2 {
		t.Fatalf("packet len = %d, want %d", len(packet), HeaderLen+2+len(protoBytes)+2)
	}
	if got, want := binary.LittleEndian.Uint16(packet[2:4]), uint16(2+2+len(protoBytes)); got != want {
		t.Fatalf("bodyLen = %d, want %d", got, want)
	}
	if got := binary.LittleEndian.Uint16(packet[6:8]); got == 0 {
		t.Fatalf("seqId was not XOR-obfuscated")
	}

	decoded, err := codec.UnmarshalResponse(packet)
	if err != nil {
		t.Fatalf("UnmarshalResponse returned error: %v", err)
	}
	if decoded.Sys != 255 || decoded.MsgID != 1 {
		t.Fatalf("route = %d/%d, want 255/1", decoded.Sys, decoded.MsgID)
	}
	if string(decoded.ProtoBytes) != string(protoBytes) {
		t.Fatalf("proto bytes = %x, want %x", decoded.ProtoBytes, protoBytes)
	}
	if decoded.SeqID == 0 {
		t.Fatalf("decoded header crc should be recorded in SeqID field")
	}
}

func TestCodecMatchesMrRobotHeaderPacket(t *testing.T) {
	key := [4]byte{0xDE, 0x28, 0x06, 0x05}
	codec := NewCodec(key)
	protoBytes := []byte{0x08, 0x96, 0x01, 0x12, 0x03, 'b', 'o', 'b'}
	body := makeGatewayBody(255, 2, protoBytes)

	packet := makeMrRobotPacket(key, body, 7)
	decoded, err := codec.UnmarshalResponse(packet)
	if err != nil {
		t.Fatalf("UnmarshalResponse returned error: %v", err)
	}
	if decoded.Sys != 255 || decoded.MsgID != 2 {
		t.Fatalf("decoded packet = sys:%d msg:%d", decoded.Sys, decoded.MsgID)
	}
	if string(decoded.ProtoBytes) != string(protoBytes) {
		t.Fatalf("proto bytes = %x, want %x", decoded.ProtoBytes, protoBytes)
	}
}

func TestUnmarshalResponseAcceptsPlainResponseHeaderPacket(t *testing.T) {
	key := [4]byte{0xDA, 0xFD, 0x12, 0x34}
	codec := NewCodec(key)
	protoBytes := []byte{0x08, 0x96, 0x01, 0x12, 0x03, 'b', 'o', 'b'}
	body := makeGatewayBody(255, 2, protoBytes)

	packet := makeMrRobotPacketWithModes(key, body, 7, false, false)
	decoded, err := codec.UnmarshalResponse(packet)
	if err != nil {
		t.Fatalf("UnmarshalResponse returned error: %v", err)
	}
	if decoded.Sys != 255 || decoded.MsgID != 2 {
		t.Fatalf("decoded packet = sys:%d msg:%d", decoded.Sys, decoded.MsgID)
	}
	if string(decoded.ProtoBytes) != string(protoBytes) {
		t.Fatalf("proto bytes = %x, want %x", decoded.ProtoBytes, protoBytes)
	}
}
func TestUnmarshalResponseAcceptsZeroBodyCRCPacket(t *testing.T) {
	key := [4]byte{0xDA, 0xFD, 0x12, 0x34}
	codec := NewCodec(key)
	protoBytes := []byte{0x08, 0x96, 0x01, 0x12, 0x03, 'b', 'o', 'b'}
	body := makeGatewayBody(255, 2, protoBytes)

	packet := makeMrRobotPacketWithBodyCRC(key, body, 7, 0, false, false)
	decoded, err := codec.UnmarshalResponse(packet)
	if err != nil {
		t.Fatalf("UnmarshalResponse returned error: %v", err)
	}
	if decoded.Sys != 255 || decoded.MsgID != 2 {
		t.Fatalf("decoded packet = sys:%d msg:%d", decoded.Sys, decoded.MsgID)
	}
	if string(decoded.ProtoBytes) != string(protoBytes) {
		t.Fatalf("proto bytes = %x, want %x", decoded.ProtoBytes, protoBytes)
	}
}
func TestUnmarshalResponseIgnoresUnreliableBodyCRC(t *testing.T) {
	key := [4]byte{0xDA, 0xFD, 0x12, 0x34}
	codec := NewCodec(key)
	protoBytes := []byte{0x08, 0x96, 0x01, 0x12, 0x03, 'b', 'o', 'b'}
	body := makeGatewayBody(255, 2, protoBytes)

	packet := makeMrRobotPacketWithBodyCRC(key, body, 7, 0x7D50, false, false)
	decoded, err := codec.UnmarshalResponse(packet)
	if err != nil {
		t.Fatalf("UnmarshalResponse returned error: %v", err)
	}
	if decoded.Sys != 255 || decoded.MsgID != 2 {
		t.Fatalf("decoded packet = sys:%d msg:%d", decoded.Sys, decoded.MsgID)
	}
	if string(decoded.ProtoBytes) != string(protoBytes) {
		t.Fatalf("proto bytes = %x, want %x", decoded.ProtoBytes, protoBytes)
	}
}

func TestUnmarshalResponseAcceptsEncryptedBodyPacket(t *testing.T) {
	key := [4]byte{0xDE, 0x28, 0x06, 0x05}
	codec := NewCodec(key)
	protoBytes := []byte{0x08, 0x96, 0x01, 0x12, 0x03, 'b', 'o', 'b'}
	body := makeGatewayBody(255, 2, protoBytes)

	packet := makeMrRobotPacketWithBodyMode(key, body, 7, true)
	decoded, err := codec.UnmarshalResponse(packet)
	if err != nil {
		t.Fatalf("UnmarshalResponse returned error: %v", err)
	}
	if decoded.Sys != 255 || decoded.MsgID != 2 {
		t.Fatalf("decoded packet = sys:%d msg:%d", decoded.Sys, decoded.MsgID)
	}
	if string(decoded.ProtoBytes) != string(protoBytes) {
		t.Fatalf("proto bytes = %x, want %x", decoded.ProtoBytes, protoBytes)
	}
}
func TestMarshalRequestMatchesMrRobotPlainBodyPacket(t *testing.T) {
	key := [4]byte{0xDE, 0x28, 0x06, 0x05}
	codec := NewCodec(key)
	protoBytes := []byte{0x08, 0x96, 0x01, 0x12, 0x03, 'b', 'o', 'b'}

	packet, err := codec.MarshalRequest(255, 1, protoBytes)
	if err != nil {
		t.Fatalf("MarshalRequest returned error: %v", err)
	}
	body := packet[HeaderLen:]
	if body[0] != 255 || body[1] != 1 {
		t.Fatalf("plain route bytes = %d/%d, want 255/1", body[0], body[1])
	}
	if got := binary.LittleEndian.Uint16(body[2:4]); got != uint16(len(protoBytes)) {
		t.Fatalf("proto length = %d, want %d", got, len(protoBytes))
	}
	if string(body[4:]) != string(protoBytes) {
		t.Fatalf("plain proto bytes = %x, want %x", body[4:], protoBytes)
	}
}

func TestCodecSeqIDIncrementsAndWraps(t *testing.T) {
	codec := NewCodec([4]byte{1, 2, 3, 4})
	codec.SetSeqID(0xFFFF)

	if _, err := codec.MarshalRequest(1, 2, nil); err != nil {
		t.Fatalf("first marshal error: %v", err)
	}
	if _, err := codec.MarshalRequest(1, 2, nil); err != nil {
		t.Fatalf("second marshal error: %v", err)
	}
	if codec.SeqID() != 1 {
		t.Fatalf("next seq = %d, want 1", codec.SeqID())
	}
}

func TestUnmarshalResponseErrors(t *testing.T) {
	codec := NewCodec([4]byte{0xDE, 0x28, 0x06, 0x05})
	packet, err := codec.MarshalRequest(255, 1, []byte{1, 2, 3})
	if err != nil {
		t.Fatalf("MarshalRequest error: %v", err)
	}

	t.Run("header too short", func(t *testing.T) {
		_, err := codec.UnmarshalResponse(packet[:HeaderLen-1])
		if !errors.Is(err, ErrHeaderTooShort) {
			t.Fatalf("err = %v, want %v", err, ErrHeaderTooShort)
		}
	})

	t.Run("body length mismatch", func(t *testing.T) {
		bad := append([]byte(nil), packet...)
		bad = bad[:len(bad)-1]
		_, err := codec.UnmarshalResponse(bad)
		if !errors.Is(err, ErrBodyLengthMismatch) {
			t.Fatalf("err = %v, want %v", err, ErrBodyLengthMismatch)
		}
	})

	t.Run("header magic", func(t *testing.T) {
		bad := append([]byte(nil), packet...)
		bad[0] ^= 0xFF
		_, err := codec.UnmarshalResponse(bad)
		if !errors.Is(err, ErrHeaderMagic) {
			t.Fatalf("err = %v, want %v", err, ErrHeaderMagic)
		}
	})

	t.Run("proto length", func(t *testing.T) {
		bad := append([]byte(nil), packet...)
		binary.LittleEndian.PutUint16(bad[HeaderLen+2:HeaderLen+4], 0xFFFF)
		_, err := codec.UnmarshalResponse(bad)
		if !errors.Is(err, ErrProtoLength) {
			t.Fatalf("err = %v, want %v", err, ErrProtoLength)
		}
	})
}
