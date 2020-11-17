package iota

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/iotaledger/iota.go/pow"
	"golang.org/x/crypto/blake2b"
)

const (
	// Defines the length of a message ID.
	MessageIDLength = blake2b.Size256
	// Defines the length of the network ID in bytes.
	MessageNetworkIDLength = UInt64ByteSize
	// Defines the minimum size of a message: network ID + 2 msg IDs + uint16 payload length + nonce
	MessageBinSerializedMinSize = MessageNetworkIDLength + 2*MessageIDLength + UInt32ByteSize + UInt64ByteSize
)

// PayloadSelector implements SerializableSelectorFunc for payload types.
func PayloadSelector(payloadType uint32) (Serializable, error) {
	var seri Serializable
	switch payloadType {
	case TransactionPayloadTypeID:
		seri = &Transaction{}
	case MilestonePayloadTypeID:
		seri = &Milestone{}
	case IndexationPayloadTypeID:
		seri = &Indexation{}
	default:
		return nil, fmt.Errorf("%w: type %d", ErrUnknownPayloadType, payloadType)
	}
	return seri, nil
}

// MessageID is the ID of a Message.
type MessageID = [MessageIDLength]byte

// MessageIDs are IDs of messages.
type MessageIDs = []MessageID

// MessageIDFromHexString converts the given message IDs from their hex
// to MessageID representation.
func MessageIDFromHexString(messageIDHex string) (MessageID, error) {
	messageIDBytes, err := hex.DecodeString(messageIDHex)
	if err != nil {
		return MessageID{}, err
	}

	messageID := MessageID{}
	copy(messageID[:], messageIDBytes)

	return messageID, nil
}

// Message can carry a payload and references two other messages.
type Message struct {
	// The network ID for which this message is meant for.
	NetworkID uint64
	// The 1st parent the message references.
	Parent1 [MessageIDLength]byte
	// The 2nd parent the message references.
	Parent2 [MessageIDLength]byte
	// The inner payload of the message. Can be nil.
	Payload Serializable
	// The nonce which lets this message fulfill the PoW requirements.
	Nonce uint64
}

// ID computes the ID of the Message.
func (m *Message) ID() (*MessageID, error) {
	data, err := m.Serialize(DeSeriModeNoValidation)
	if err != nil {
		return nil, fmt.Errorf("can't compute message ID: %w", err)
	}
	h := blake2b.Sum256(data)
	return &h, nil
}

// POW computes the PoW score of the Message.
func (m *Message) POW() (float64, error) {
	data, err := m.Serialize(DeSeriModeNoValidation)
	if err != nil {
		return 0, fmt.Errorf("can't compute message PoW score: %w", err)
	}
	return pow.Score(data), nil
}

func (m *Message) Deserialize(data []byte, deSeriMode DeSerializationMode) (int, error) {
	return NewDeserializer(data).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(DeSeriModePerformValidation) {
				if err := checkMinByteLength(MessageBinSerializedMinSize, len(data)); err != nil {
					return fmt.Errorf("invalid message bytes: %w", err)
				}
			}
			return nil
		}).
		ReadNum(&m.NetworkID, func(err error) error {
			return fmt.Errorf("unable to deserialize message network ID: %w", err)
		}).
		ReadArrayOf32Bytes(&m.Parent1, func(err error) error {
			return fmt.Errorf("unable to deserialize message parent 1: %w", err)
		}).
		ReadArrayOf32Bytes(&m.Parent2, func(err error) error {
			return fmt.Errorf("unable to deserialize message parent 2: %w", err)
		}).
		ReadPayload(func(seri Serializable) { m.Payload = seri }, deSeriMode, func(err error) error {
			return fmt.Errorf("unable to deserialize message's inner payload: %w", err)
		}).
		ReadNum(&m.Nonce, func(err error) error {
			return fmt.Errorf("unable to deserialize message nonce: %w", err)
		}).
		ConsumedAll(func(leftOver int, err error) error {
			return fmt.Errorf("%w: unable to deserialize message: %d bytes are still available", err, leftOver)
		}).
		Done()
}

func (m *Message) Serialize(deSeriMode DeSerializationMode) ([]byte, error) {
	return NewSerializer().
		WriteNum(m.NetworkID, func(err error) error {
			return fmt.Errorf("unable to serialize message network ID: %w", err)
		}).
		WriteBytes(m.Parent1[:], func(err error) error {
			return fmt.Errorf("unable to serialize message parent 1: %w", err)
		}).
		WriteBytes(m.Parent2[:], func(err error) error {
			return fmt.Errorf("unable to serialize message parent 2: %w", err)
		}).
		WritePayload(m.Payload, deSeriMode, func(err error) error {
			return fmt.Errorf("unable to serialize message inner payload: %w", err)
		}).
		WriteNum(m.Nonce, func(err error) error {
			return fmt.Errorf("unable to serialize message nonce: %w", err)
		}).
		Serialize()
}

func (m *Message) MarshalJSON() ([]byte, error) {
	jsonMsg := &jsonmessage{}
	jsonMsg.NetworkID = strconv.FormatUint(m.NetworkID, 10)
	jsonMsg.Parent1 = hex.EncodeToString(m.Parent1[:])
	jsonMsg.Parent2 = hex.EncodeToString(m.Parent2[:])
	jsonMsg.Nonce = strconv.FormatUint(m.Nonce, 10)
	if m.Payload != nil {
		jsonPayload, err := m.Payload.MarshalJSON()
		if err != nil {
			return nil, err
		}
		rawMsgJsonPayload := json.RawMessage(jsonPayload)
		jsonMsg.Payload = &rawMsgJsonPayload
	}
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
func jsonpayloadselector(ty int) (JSONSerializable, error) {
	var obj JSONSerializable
	switch uint32(ty) {
	case TransactionPayloadTypeID:
		obj = &jsontransaction{}
	case MilestonePayloadTypeID:
		obj = &jsonmilestonepayload{}
	case IndexationPayloadTypeID:
		obj = &jsonindexation{}
	default:
		return nil, fmt.Errorf("unable to decode payload type from JSON: %w", ErrUnknownPayloadType)
	}
	return obj, nil
}

// jsonmessage defines the JSON representation of a Message.
type jsonmessage struct {
	// The network ID identifying the network for this message.
	NetworkID string `json:"networkId"`
	// The hex encoded message ID of the first referenced parent.
	Parent1 string `json:"parent1MessageId"`
	// The hex encoded message ID of the second referenced parent.
	Parent2 string `json:"parent2MessageId"`
	// The payload within the message.
	Payload *json.RawMessage `json:"payload"`
	// The nonce the message used to fulfill the PoW requirement.
	Nonce string `json:"nonce"`
}

func (jm *jsonmessage) ToSerializable() (Serializable, error) {
	parent1, err := hex.DecodeString(jm.Parent1)
	if err != nil {
		return nil, fmt.Errorf("unable to decode hex parent 1 from JSON: %w", err)
	}

	parent2, err := hex.DecodeString(jm.Parent2)
	if err != nil {
		return nil, fmt.Errorf("unable to decode hex parent 2 from JSON: %w", err)
	}

	var parsedNonce uint64
	if len(jm.Nonce) != 0 {
		parsedNonce, err = strconv.ParseUint(jm.Nonce, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse message nonce from JSON: %w", err)
		}
	}

	var parsedNetworkID uint64
	if len(jm.NetworkID) != 0 {
		parsedNetworkID, err = strconv.ParseUint(jm.NetworkID, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse message network ID from JSON: %w", err)
		}
	}

	m := &Message{NetworkID: parsedNetworkID, Nonce: parsedNonce}

	if jm.Payload != nil {
		jsonPayload, err := DeserializeObjectFromJSON(jm.Payload, jsonpayloadselector)
		if err != nil {
			return nil, err
		}

		m.Payload, err = jsonPayload.ToSerializable()
		if err != nil {
			return nil, err
		}
	}

	copy(m.Parent1[:], parent1)
	copy(m.Parent2[:], parent2)

	return m, nil
}
