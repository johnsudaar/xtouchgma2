package rtp

import (
	"bytes"
	"encoding/binary"
)

type AppleMidiCommand string

const (
	AppleMidiCommandInvitation              AppleMidiCommand = "IN"
	AppleMidiCommandInvitationAccepted      AppleMidiCommand = "OK"
	AppleMidiCommandInvitationRejected      AppleMidiCommand = "NO"
	AppleMidiCommandInvitationTimestampSync AppleMidiCommand = "CK"
	AppleMidiCommandExit                    AppleMidiCommand = "BY"
)

type AppleMidiExchangePacket struct {
	Preamble       uint16
	Command        AppleMidiCommand
	Version        uint32
	InitiatorToken uint32
	SSRC           uint32
	Name           string
}

func (h *AppleMidiExchangePacket) UnmarshalBinary(data []byte) error {
	if len(data) < 16 {
		return nil
	}
	h.Preamble = binary.BigEndian.Uint16(data[0:2])
	h.Command = AppleMidiCommand(data[2:4])
	h.Version = binary.BigEndian.Uint32(data[4:8])
	h.InitiatorToken = binary.BigEndian.Uint32(data[8:12])
	h.SSRC = binary.BigEndian.Uint32(data[12:16])
	h.Name = string(data[16 : len(data)-1]) // Last byte is 0
	return nil
}

func (h *AppleMidiExchangePacket) MarshalBinary() ([]byte, error) {
	buff := new(bytes.Buffer)

	binary.Write(buff, binary.BigEndian, h.Preamble)
	buff.WriteString(string(h.Command))
	binary.Write(buff, binary.BigEndian, h.Version)
	binary.Write(buff, binary.BigEndian, h.InitiatorToken)
	binary.Write(buff, binary.BigEndian, h.SSRC)
	buff.WriteString(h.Name)
	binary.Write(buff, binary.BigEndian, byte(0))

	return buff.Bytes(), nil
}

type AppleMidiTimestampPacket struct {
	Preamble uint16           // 0,1
	Command  AppleMidiCommand // 2,3
	SSRC     uint32           // 4,5,6,7
	Count    uint8            // 8
	// 3 bytes of padding // 9,10,11
	Timestamp1 uint64
	Timestamp2 uint64
	Timestamp3 uint64
}

func (h *AppleMidiTimestampPacket) UnmarshalBinary(data []byte) error {
	h.Preamble = binary.BigEndian.Uint16(data[0:2])
	h.Command = AppleMidiCommand(data[2:4])
	h.SSRC = binary.BigEndian.Uint32(data[4:8])
	h.Count = uint8(data[8])
	h.Timestamp1 = binary.BigEndian.Uint64(data[12:20])
	h.Timestamp2 = binary.BigEndian.Uint64(data[20:28])
	h.Timestamp3 = binary.BigEndian.Uint64(data[28:36])
	return nil
}

func (h *AppleMidiTimestampPacket) MarshalBinary() ([]byte, error) {
	buff := new(bytes.Buffer)

	binary.Write(buff, binary.BigEndian, h.Preamble)
	buff.WriteString(string(h.Command))
	binary.Write(buff, binary.BigEndian, h.SSRC)
	binary.Write(buff, binary.BigEndian, h.Count)
	buff.Write([]byte{0, 0, 0})
	binary.Write(buff, binary.BigEndian, h.Timestamp1)
	binary.Write(buff, binary.BigEndian, h.Timestamp2)
	binary.Write(buff, binary.BigEndian, h.Timestamp3)

	return buff.Bytes(), nil
}

type AppleMidiDataPacket struct {
	V              byte   // 2 bits
	P              bool   // 1 bit
	X              bool   // 1 bit
	CC             byte   // 4 bits
	M              bool   // 1 bit
	PayloadType    byte   // 7 bits
	SequenceNumber uint16 // 2 bytes
	Timestamp      uint32 // 4 bytes
	SSRC           uint32 // 4 bytes
	Big            bool   // 1 bit
	Journal        bool   // 1 bit
	Z              bool   // 1 bit
	P2             bool   // 1 bit
	Len            uint16 // 4 bits
	MidiMessage    []byte // LEB bytes
}

func (h *AppleMidiDataPacket) UnmarshalBinary(data []byte) error {
	h.V = (data[0] & 0b11000000) >> 6
	h.P = (data[0] & 0b00100000) != 0
	h.X = (data[0] & 0b00010000) != 0
	h.CC = data[0] & 0b00001111

	h.M = (data[1] & 0b10000000) != 0
	h.PayloadType = (data[1] & 0b01111111)
	h.SequenceNumber = binary.BigEndian.Uint16(data[2:4])
	h.Timestamp = binary.BigEndian.Uint32(data[4:8])
	h.SSRC = binary.BigEndian.Uint32(data[8:12])

	h.Big = (data[12] & 0b10000000) != 0
	h.Journal = (data[12] & 0b01000000) != 0
	h.Z = (data[12] & 0b00100000) != 0
	h.P2 = (data[12] & 0b00010000) != 0
	if h.Big {
		h.Len = uint16(data[12]&0b00001111) << 8
		h.Len += uint16(data[13])
		h.MidiMessage = data[14 : 14+h.Len]
	} else {
		h.Len = uint16(data[12] & 0b00001111)
		h.MidiMessage = data[13 : 13+h.Len]
	}
	return nil
}

func (h *AppleMidiDataPacket) MarshalBinary() ([]byte, error) {
	buff := new(bytes.Buffer)

	var cur byte = 0
	cur += (h.V & 0b00000011) << 6
	if h.P {
		cur += 0b00100000
	}

	if h.X {
		cur += 0b00010000
	}

	cur += (h.CC & 0b00001111)
	buff.WriteByte(cur)
	cur = 0
	if h.M {
		cur += 0b10000000
	}

	cur += (h.PayloadType & 0b01111111)
	buff.WriteByte(cur)
	binary.Write(buff, binary.BigEndian, h.SequenceNumber)
	binary.Write(buff, binary.BigEndian, h.Timestamp)
	binary.Write(buff, binary.BigEndian, h.SSRC)

	cur = 0
	if h.Big {
		cur += 0b10000000
	}

	if h.Journal {
		cur += 0b01000000
	}

	if h.Z {
		cur += 0b00100000
	}

	if h.P2 {
		cur += 0b00010000
	}

	if h.Big {
		cur += (byte(h.Len) >> 8) & 0b00001111
		buff.WriteByte(cur)
		buff.WriteByte(byte(h.Len))
	} else {
		cur += byte(h.Len) & 0b00001111
		buff.WriteByte(cur)
	}
	buff.Write(h.MidiMessage)

	return buff.Bytes(), nil
}
