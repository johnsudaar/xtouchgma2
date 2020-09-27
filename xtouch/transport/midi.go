package transport

type MidiMessageType string

type MidiMessageStatus byte

const (
	MidiMessageStatusNoteOff              MidiMessageStatus = 0x8
	MidiMessageStatusNoteOn               MidiMessageStatus = 0x9
	MidiMessageStatusPolyphonicAftertouch MidiMessageStatus = 0xa
	MidiMessageStatusControlChange        MidiMessageStatus = 0xb
	MidiMessageStatusProgramChange        MidiMessageStatus = 0xc
	MidiMessageStatusChannelAfterTouch    MidiMessageStatus = 0xd
	MidiMessageStatusPitchWheel           MidiMessageStatus = 0xe

	MidiMessageTypeNoteOff              MidiMessageType = "note_off"
	MidiMessageTypeNoteOn               MidiMessageType = "note_on"
	MidiMessageTypePolyphonicAftertouch MidiMessageType = "polyphonic_aftertouch"
	MidiMessageTypeControlChange        MidiMessageType = "control_change"
	MidiMessageTypeProgramChange        MidiMessageType = "program_change"
	MidiMessageTypeChannelAfterTouch    MidiMessageType = "channel_aftertouch"
	MidiMessageTypePitchWheel           MidiMessageType = "pitch_wheel"
)

type MidiMessage struct {
	Type             MidiMessageType
	Channel          byte
	NoteNumber       byte
	Velocity         byte
	Pressure         byte
	ProgamNumber     byte
	ControllerNumber byte
	ControlData      byte
	PitchBend        uint16
}

func (h *MidiMessage) UnmarshalBinary(data []byte) error {
	status := MidiMessageStatus(data[0] & 0xf0 >> 4)
	h.Channel = data[0] & 0x0f

	if status == MidiMessageStatusNoteOff {
		h.Type = MidiMessageTypeNoteOff
		h.NoteNumber = data[1]
		h.Velocity = data[2]
	}

	if status == MidiMessageStatusNoteOn {
		h.Type = MidiMessageTypeNoteOn
		h.NoteNumber = data[1]
		h.Velocity = data[2]
	}

	if status == MidiMessageStatusPolyphonicAftertouch {
		h.Type = MidiMessageTypePolyphonicAftertouch
		h.NoteNumber = data[1]
		h.Pressure = data[2]
	}

	if status == MidiMessageStatusControlChange {
		h.Type = MidiMessageTypeControlChange
		h.ControllerNumber = data[1]
		h.ControlData = data[2]
	}

	if status == MidiMessageStatusProgramChange {
		h.Type = MidiMessageTypeProgramChange
		h.ProgamNumber = data[1]
	}

	if status == MidiMessageStatusChannelAfterTouch {
		h.Type = MidiMessageTypeChannelAfterTouch
		h.Pressure = data[1]
	}

	if status == MidiMessageStatusPitchWheel {
		h.Type = MidiMessageTypePitchWheel
		LSByte := uint16(data[1] & 0b01111111)
		MSByte := uint16(data[2] & 0b01111111)
		h.PitchBend = LSByte + (MSByte << 7)
	}
	return nil
}

func (h *MidiMessage) MarshalBinary() ([]byte, error) {
	res := make([]byte, 3)
	var status MidiMessageStatus
	if h.Type == MidiMessageTypeNoteOff {
		status = MidiMessageStatusNoteOff
		res[1] = h.NoteNumber
		res[2] = h.Velocity
	}

	if h.Type == MidiMessageTypeNoteOn {
		status = MidiMessageStatusNoteOn
		res[1] = h.NoteNumber
		res[2] = h.Velocity
	}

	if h.Type == MidiMessageTypePolyphonicAftertouch {
		status = MidiMessageStatusPolyphonicAftertouch
		res[1] = h.NoteNumber
		res[2] = h.Pressure
	}

	if h.Type == MidiMessageTypeControlChange {
		status = MidiMessageStatusControlChange
		res[1] = h.ControllerNumber
		res[2] = h.ControlData
	}

	if h.Type == MidiMessageTypeProgramChange {
		status = MidiMessageStatusProgramChange
		res[1] = h.ProgamNumber
	}

	if h.Type == MidiMessageTypeChannelAfterTouch {
		status = MidiMessageStatusChannelAfterTouch
		res[1] = h.Pressure
	}

	if h.Type == MidiMessageTypePitchWheel {
		status = MidiMessageStatusPitchWheel
		res[1] = byte(h.PitchBend & 0b1111111)
		res[2] = byte((h.PitchBend & 0b11111110000000) >> 7)
	}

	res[0] = byte(status) << 4
	res[0] += h.Channel

	return res, nil
}
