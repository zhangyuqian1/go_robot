package gateway

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
)

const (
	HeaderLen          = 16
	mrRobotHeaderMagic = 0xCCEE
)

var (
	ErrHeaderTooShort     = errors.New("gateway packet header too short")
	ErrBodyLengthMismatch = errors.New("gateway packet body length mismatch")
	ErrHeaderMagic        = errors.New("gateway packet header magic mismatch")
	ErrBodyCRC            = errors.New("gateway packet body crc mismatch")
	ErrBodyTooShort       = errors.New("gateway packet body too short")
	ErrProtoLength        = errors.New("gateway packet proto length mismatch")
)

type Packet struct {
	Sys        byte
	MsgID      byte
	ProtoBytes []byte
	SeqID      uint16
}

type Codec struct {
	keyBuff [4]byte
	seqID   uint16
}

func NewCodec(keyBuff [4]byte) *Codec {
	return &Codec{keyBuff: keyBuff}
}

func (c *Codec) KeyBuff() [4]byte {
	return c.keyBuff
}

func (c *Codec) SetKeyBuff(keyBuff [4]byte) {
	c.keyBuff = keyBuff
}

func (c *Codec) SetSeqID(seqID uint16) {
	c.seqID = seqID
}

func (c *Codec) SeqID() uint16 {
	return c.seqID
}

func (c *Codec) MarshalRequest(sys, msgID byte, protoBytes []byte) ([]byte, error) {
	if len(protoBytes) > math.MaxUint16 {
		return nil, fmt.Errorf("%w: proto bytes length %d exceeds uint16", ErrProtoLength, len(protoBytes))
	}

	body := make([]byte, 2+2+len(protoBytes))
	body[0] = sys
	body[1] = msgID
	binary.LittleEndian.PutUint16(body[2:4], uint16(len(protoBytes)))
	copy(body[4:], protoBytes)

	packet := make([]byte, HeaderLen+len(body))
	binary.LittleEndian.PutUint16(packet[0:2], mrRobotHeaderMagic)
	binary.LittleEndian.PutUint16(packet[2:4], uint16(len(body)))
	binary.LittleEndian.PutUint16(packet[4:6], CRC16(body))
	binary.LittleEndian.PutUint16(packet[6:8], c.seqID)
	c.seqID++

	headerCRC := CRC16(packet[:HeaderLen])
	binary.LittleEndian.PutUint16(packet[6:8], headerCRC)
	xorBytes(packet, c.keyBuff, 4, 4)

	copy(packet[HeaderLen:], body)
	return packet, nil
}

func decodeGatewayBody(body []byte) (byte, byte, []byte, error) {
	if len(body) < 4 {
		return 0, 0, nil, fmt.Errorf("%w: got %d bytes", ErrBodyTooShort, len(body))
	}
	protoLen := int(binary.LittleEndian.Uint16(body[2:4]))
	if protoLen != len(body)-4 {
		return 0, 0, nil, fmt.Errorf("%w: declared %d but got %d bytes", ErrProtoLength, protoLen, len(body)-4)
	}
	protoBytes := make([]byte, protoLen)
	copy(protoBytes, body[4:])
	return body[0], body[1], protoBytes, nil
}

func (c *Codec) UnmarshalResponse(packetBytes []byte) (*Packet, error) {
	if len(packetBytes) < HeaderLen {
		return nil, fmt.Errorf("%w: got %d bytes", ErrHeaderTooShort, len(packetBytes))
	}

	header := append([]byte(nil), packetBytes[:HeaderLen]...)
	if magic := binary.LittleEndian.Uint16(header[0:2]); magic != mrRobotHeaderMagic {
		return nil, fmt.Errorf("%w: got 0x%04x want 0x%04x", ErrHeaderMagic, magic, mrRobotHeaderMagic)
	}

	bodyLen := int(binary.LittleEndian.Uint16(header[2:4]))
	if len(packetBytes) != HeaderLen+bodyLen {
		return nil, fmt.Errorf("%w: header bodyLen %d but packet has %d body bytes", ErrBodyLengthMismatch, bodyLen, len(packetBytes)-HeaderLen)
	}
	if bodyLen < 4 {
		return nil, fmt.Errorf("%w: got %d bytes", ErrBodyTooShort, bodyLen)
	}

	body := append([]byte(nil), packetBytes[HeaderLen:]...)
	sys, msgID, protoBytes, err := decodeGatewayBody(body)
	if err != nil {
		decodedBody := append([]byte(nil), body...)
		xorBytes(decodedBody, c.keyBuff, 0, len(decodedBody))
		sys, msgID, protoBytes, err = decodeGatewayBody(decodedBody)
		if err != nil {
			return nil, err
		}
	}

	return &Packet{
		Sys:        sys,
		MsgID:      msgID,
		ProtoBytes: protoBytes,
		SeqID:      binary.LittleEndian.Uint16(header[6:8]),
	}, nil
}
