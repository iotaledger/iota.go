package iota

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"golang.org/x/crypto/blake2b"
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
	case MilestonePayloadID:
		seri = &MilestonePayload{}
	case IndexationPayloadID:
		seri = &IndexationPayload{}
	default:
		return nil, fmt.Errorf("%w: type %d", ErrUnknownPayloadType, payloadType)
	}
	return seri, nil
}

// MessageHash is the hash of a Message.
type MessageHash = [MessageHashLength]byte

// MessageHashes are hashes of messages.
type MessageHashes = []MessageHash

// Message can carry a payload and references two other messages.
type Message struct {
	// The version of the message.
	Version byte `json:"version"`
	// The 1st parent the message references.
	Parent1 [MessageHashLength]byte `json:"parent_1"`
	// The 2nd parent the message references.
	Parent2 [MessageHashLength]byte `json:"parent_2"`
	// The inner payload of the message. Can be nil.
	Payload Serializable `json:"payload"`
	// The nonce which lets this message fulfill the PoW requirements.
	Nonce uint64 `json:"nonce"`
}

// Hash computes the hash of the Message.
func (m *Message) Hash() (*MessageHash, error) {
	data, err := m.Serialize(DeSeriModeNoValidation)
	if err != nil {
		return nil, fmt.Errorf("can't compute message hash: %w", err)
	}
	h := blake2b.Sum256(data)
	return &h, nil
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
	m.Version = data[0]

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

func (m *Message) MarshalJSON() ([]byte, error) {
	jsonMsg := &jsonmessage{}
	msgHash, err := m.Hash()
	if err != nil {
		return nil, err
	}
	jsonMsg.Version = MessageVersion
	jsonMsg.Hash = hex.EncodeToString(msgHash[:])
	jsonMsg.Parent1 = hex.EncodeToString(m.Parent1[:])
	jsonMsg.Parent2 = hex.EncodeToString(m.Parent2[:])
	jsonMsg.Nonce = int(m.Nonce)
	jsonPayload, err := m.Payload.MarshalJSON()
	if err != nil {
		return nil, err
	}
	rawMsgJsonPayload := json.RawMessage(jsonPayload)
	jsonMsg.Payload = &rawMsgJsonPayload
	return json.Marshal(jsonMsg)
}

func (m *Message) UnmarshalJSON(bytes []byte) error {
	jsonMsg := &jsonmessage{}
	if err := json.Unmarshal(bytes, jsonMsg); err != nil {
		return err
	}
	seri, err := jsonMsg.ToSerializable()
	if err != nil {
		return err
	}
	*m = *seri.(*Message)
	return nil
}

// selects the json object for the given type.
func jsonPayloadSelector(ty int) (JSONSerializable, error) {
	var obj JSONSerializable
	switch uint32(ty) {
	case SignedTransactionPayloadID:
		obj = &jsonsignedtransactionpayload{}
	case MilestonePayloadID:
		obj = &jsonmilestonepayload{}
	case IndexationPayloadID:
		obj = &jsonindexationpayload{}
	default:
		return nil, fmt.Errorf("unable to decode payload type from JSON: %w", ErrUnknownPayloadType)
	}
	return obj, nil
}

// JSONNodeMessages is a slice of jsonmessage.
type JSONNodeMessages []*jsonmessage

// ToMessages converts a slice of jsonmessage to a slice of Message.
func (nm JSONNodeMessages) ToMessages() ([]*Message, error) {
	msgs := make([]*Message, len(nm))
	for i, n := range nm {
		seri, err := n.ToSerializable()
		if err != nil {
			return nil, fmt.Errorf("unable to decode message at pos %d: %w", i, err)
		}
		msgs[i] = seri.(*Message)
	}
	return msgs, nil
}

// jsonmessage defines the JSON representation of a Message.
type jsonmessage struct {
	// The hex encoded hash of the message.
	Hash string `json:"hash"`
	// The version of the message.
	Version int `json:"version"`
	// The hex encoded hash of the first referenced parent.
	Parent1 string `json:"parent1"`
	// The hex encoded hash of the second referenced parent.
	Parent2 string `json:"parent2"`
	// The payload within the message.
	Payload *json.RawMessage `json:"payload"`
	// The nonce the message used to fulfill the PoW requirement.
	Nonce int `json:"nonce"`
}

func (jm *jsonmessage) ToSerializable() (Serializable, error) {
	jsonPayload, err := DeserializeObjectFromJSON(jm.Payload, jsonPayloadSelector)
	if err != nil {
		return nil, err
	}

	payload, err := jsonPayload.ToSerializable()
	if err != nil {
		return nil, err
	}

	parent1, err := hex.DecodeString(jm.Parent1)
	if err != nil {
		return nil, err
	}

	parent2, err := hex.DecodeString(jm.Parent2)
	if err != nil {
		return nil, err
	}

	m := &Message{Version: byte(jm.Version), Nonce: uint64(jm.Nonce), Payload: payload}
	copy(m.Parent1[:], parent1)
	copy(m.Parent2[:], parent2)

	return m, nil
}
