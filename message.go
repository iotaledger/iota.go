package iota

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	// Denotes the current message version.
	MessageVersion = 1
	// Defines the length of a message hash.
	MessageHashLength = 32
	// Defines the minimum size of a message: version + 2 msg hashes + uint16 payload length + nonce
	MessageMinSize = MessageVersionByteSize + 2*MessageHashLength + UInt32ByteSize + UInt64ByteSize
)

// PayloadSelector implements SerializableSelectorFunc for payload types.
func PayloadSelector(payloadType uint32) (Serializable, error) {
	var seri Serializable
	switch payloadType {
	case SignedTransactionPayloadID:
		seri = &SignedTransactionPayload{}
	case IndexationPayloadID:
		seri = &IndexationPayload{}
	default:
		return nil, fmt.Errorf("%w: type %d", ErrUnknownPayloadType, payloadType)
	}
	return seri, nil
}

// Message can carry a payload and references two other messages.
type Message struct {
	// The 1st parent the message references.
	Parent1 [MessageHashLength]byte `json:"parent_1"`
	// The 2nd parent the message references.
	Parent2 [MessageHashLength]byte `json:"parent_2"`
	// The inner payload of the message. Can be nil.
	Payload Serializable `json:"payload"`
	// The nonce which lets this message fulfill the PoW requirements.
	Nonce uint64 `json:"nonce"`
}

func (m *Message) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		if err := checkMinByteLength(MessageMinSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid message bytes: %w", err)
		}
		if err := checkTypeByte(data, MessageVersion); err != nil {
			return 0, fmt.Errorf("unable to deserialize message: %w", err)
		}
	}
	l := len(data)

	// read parents
	data = data[MessageVersionByteSize:]
	copy(m.Parent1[:], data[:MessageHashLength])
	data = data[MessageHashLength:]
	copy(m.Parent2[:], data[:MessageHashLength])
	data = data[MessageHashLength:]

	payload, payloadBytesRead, err := ParsePayload(data, deSeriMode)
	if err != nil {
		return 0, fmt.Errorf("%w: can't parse payload within message", err)
	}
	m.Payload = payload

	// must have consumed entire data slice minus the nonce
	data = data[payloadBytesRead:]
	if leftOver := len(data) - UInt64ByteSize; leftOver != 0 {
		return 0, fmt.Errorf("%w: unable to deserialize message: %d are still available", ErrDeserializationNotAllConsumed, leftOver)
	}

	m.Nonce = binary.LittleEndian.Uint64(data)
	return l, nil
}

func (m *Message) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	if m.Payload == nil {
		var b [MessageMinSize]byte
		b[0] = MessageVersion
		copy(b[MessageVersionByteSize:], m.Parent1[:])
		copy(b[MessageVersionByteSize+MessageHashLength:], m.Parent2[:])
		binary.LittleEndian.PutUint32(b[MessageVersionByteSize+MessageHashLength*2:], 0)
		binary.LittleEndian.PutUint64(b[len(b)-UInt64ByteSize:], m.Nonce)
		return b[:], nil
	}

	var b bytes.Buffer
	if err := b.WriteByte(MessageVersion); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize message version", err)
	}

	if _, err := b.Write(m.Parent1[:]); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize parent 1", err)
	}

	if _, err := b.Write(m.Parent2[:]); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize parent 2", err)
	}

	payloadData, err := m.Payload.Serialize(deSeriMode)
	if err != nil {
		return nil, fmt.Errorf("%w: unable to serialize message payload", err)
	}

	payloadLength := uint32(len(payloadData))

	if deSeriMode.HasMode(DeSeriModePerformValidation) {
		// TODO: check payload length
	}

	if err := binary.Write(&b, binary.LittleEndian, payloadLength); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize payload length within message", err)
	}

	if _, err := b.Write(payloadData); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize message payload", err)
	}

	if err := binary.Write(&b, binary.LittleEndian, m.Nonce); err != nil {
		return nil, fmt.Errorf("%w: unable to serialize message nonce", err)
	}

	return b.Bytes(), nil
}
