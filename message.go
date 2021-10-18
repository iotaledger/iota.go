package iotago

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/iotaledger/hive.go/serializer"
	"github.com/iotaledger/iota.go/v2/pow"
	"golang.org/x/crypto/blake2b"
)

const (
	// MessageIDLength defines the length of a message ID.
	MessageIDLength = blake2b.Size256
	// MessageNetworkIDLength defines the length of the network ID in bytes.
	MessageNetworkIDLength = serializer.UInt64ByteSize
	// MessageBinSerializedMinSize defines the minimum size of a message: network ID + parent count + 1 parent + uint16 payload length + nonce
	MessageBinSerializedMinSize = MessageNetworkIDLength + serializer.OneByte + MessageIDLength + serializer.UInt32ByteSize + serializer.UInt64ByteSize
	// MessageBinSerializedMaxSize defines the maximum size of a message.
	MessageBinSerializedMaxSize = 32768
	// MinParentsInAMessage defines the minimum amount of parents in a message.
	MinParentsInAMessage = 1
	// MaxParentsInAMessage defines the maximum amount of parents in a message.
	MaxParentsInAMessage = 8
)

var (
	// ErrMessageExceedsMaxSize gets returned when a serialized message exceeds MessageBinSerializedMaxSize.
	ErrMessageExceedsMaxSize = errors.New("message exceeds max size")

	// restrictions around parents within a message.
	messageParentArrayRules = serializer.ArrayRules{
		Min:            MinParentsInAMessage,
		Max:            MaxParentsInAMessage,
		ValidationMode: serializer.ArrayValidationModeNoDuplicates | serializer.ArrayValidationModeLexicalOrdering,
	}
)

// PayloadSelector implements SerializableSelectorFunc for payload types.
func PayloadSelector(payloadType uint32) (serializer.Serializable, error) {
	var seri serializer.Serializable
	switch payloadType {
	case TransactionPayloadTypeID:
		seri = &Transaction{}
	case MilestonePayloadTypeID:
		seri = &Milestone{}
	case IndexationPayloadTypeID:
		seri = &Indexation{}
	case ReceiptPayloadTypeID:
		seri = &Receipt{}
	case TreasuryTransactionPayloadTypeID:
		seri = &TreasuryTransaction{}
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

// MessageIDToHexString converts the given message ID to their hex representation.
func MessageIDToHexString(msgID MessageID) string {
	return hex.EncodeToString(msgID[:])
}

// MustMessageIDFromHexString converts the given message IDs from their hex
// to MessageID representation.
func MustMessageIDFromHexString(messageIDHex string) MessageID {
	msgID, err := MessageIDFromHexString(messageIDHex)
	if err != nil {
		panic(err)
	}
	return msgID
}

// Message can carry a payload and references two other messages.
type Message struct {
	// The network ID for which this message is meant for.
	NetworkID uint64
	// The parents the message references.
	Parents MessageIDs
	// The inner payload of the message. Can be nil.
	Payload serializer.Serializable
	// The nonce which lets this message fulfill the PoW requirements.
	Nonce uint64
}

// ID computes the ID of the Message.
func (m *Message) ID() (*MessageID, error) {
	data, err := m.Serialize(serializer.DeSeriModeNoValidation)
	if err != nil {
		return nil, fmt.Errorf("can't compute message ID: %w", err)
	}
	h := blake2b.Sum256(data)
	return &h, nil
}

// MustID works like ID but panics if the MessageID can't be computed.
func (m *Message) MustID() MessageID {
	msgID, err := m.ID()
	if err != nil {
		panic(err)
	}
	return *msgID
}

// POW computes the PoW score of the Message.
func (m *Message) POW() (float64, error) {
	data, err := m.Serialize(serializer.DeSeriModeNoValidation)
	if err != nil {
		return 0, fmt.Errorf("can't compute message PoW score: %w", err)
	}
	return pow.Score(data), nil
}

func (m *Message) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode) (int, error) {
	if len(data) > MessageBinSerializedMaxSize {
		return 0, fmt.Errorf("%w: size %d bytes", ErrMessageExceedsMaxSize, len(data))
	}
	return serializer.NewDeserializer(data).
		AbortIf(func(err error) error {
			if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
				if err := serializer.CheckMinByteLength(MessageBinSerializedMinSize, len(data)); err != nil {
					return fmt.Errorf("invalid message bytes: %w", err)
				}
			}
			return nil
		}).
		ReadNum(&m.NetworkID, func(err error) error {
			return fmt.Errorf("unable to deserialize message network ID: %w", err)
		}).
		ReadSliceOfArraysOf32Bytes(&m.Parents, deSeriMode, serializer.SeriLengthPrefixTypeAsByte, &messageParentArrayRules, func(err error) error {
			return fmt.Errorf("unable to deserialize message parents: %w", err)
		}).
		ReadPayload(func(seri serializer.Serializable) { m.Payload = seri }, deSeriMode, func(ty uint32) (serializer.Serializable, error) {
			switch ty {
			case TransactionPayloadTypeID:
			case IndexationPayloadTypeID:
			case MilestonePayloadTypeID:
			default:
				return nil, fmt.Errorf("a message can only contain a transaction, indexation or milestone but got type ID %d: %w", ty, ErrUnsupportedPayloadType)
			}
			return PayloadSelector(ty)
		}, func(err error) error {
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

func (m *Message) Serialize(deSeriMode serializer.DeSerializationMode) ([]byte, error) {
	data, err := serializer.NewSerializer().
		Do(func() {
			if deSeriMode.HasMode(serializer.DeSeriModePerformLexicalOrdering) {
				m.Parents = serializer.RemoveDupsAndSortByLexicalOrderArrayOf32Bytes(m.Parents)
			}
		}).
		WriteNum(m.NetworkID, func(err error) error {
			return fmt.Errorf("unable to serialize message network ID: %w", err)
		}).
		Write32BytesArraySlice(m.Parents, deSeriMode, serializer.SeriLengthPrefixTypeAsByte, &messageParentArrayRules, func(err error) error {
			return fmt.Errorf("unable to serialize message parents: %w", err)
		}).
		WritePayload(m.Payload, deSeriMode, func(err error) error {
			return fmt.Errorf("unable to serialize message inner payload: %w", err)
		}).
		WriteNum(m.Nonce, func(err error) error {
			return fmt.Errorf("unable to serialize message nonce: %w", err)
		}).
		Serialize()
	if err != nil {
		return nil, err
	}
	if len(data) > MessageBinSerializedMaxSize {
		return nil, fmt.Errorf("%w: size %d bytes", ErrMessageExceedsMaxSize, len(data))
	}
	return data, nil
}

func (m *Message) MarshalJSON() ([]byte, error) {
	jMessage := &jsonMessage{}
	jMessage.NetworkID = strconv.FormatUint(m.NetworkID, 10)
	jMessage.Parents = make([]string, len(m.Parents))
	for i, parent := range m.Parents {
		jMessage.Parents[i] = hex.EncodeToString(parent[:])
	}
	jMessage.Nonce = strconv.FormatUint(m.Nonce, 10)
	if m.Payload != nil {
		jsonPayload, err := m.Payload.MarshalJSON()
		if err != nil {
			return nil, err
		}
		rawMsgJsonPayload := json.RawMessage(jsonPayload)
		jMessage.Payload = &rawMsgJsonPayload
	}
	return json.Marshal(jMessage)
}

func (m *Message) UnmarshalJSON(bytes []byte) error {
	jMessage := &jsonMessage{}
	if err := json.Unmarshal(bytes, jMessage); err != nil {
		return err
	}
	seri, err := jMessage.ToSerializable()
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
	case TransactionPayloadTypeID:
		obj = &jsonTransaction{}
	case MilestonePayloadTypeID:
		obj = &jsonMilestone{}
	case IndexationPayloadTypeID:
		obj = &jsonIndexation{}
	default:
		return nil, fmt.Errorf("unable to decode payload type from JSON: %w", ErrUnknownPayloadType)
	}
	return obj, nil
}

// jsonMessage defines the JSON representation of a Message.
type jsonMessage struct {
	// The network ID identifying the network for this message.
	NetworkID string `json:"networkId"`
	// The hex encoded message IDs of the referenced parents.
	Parents []string `json:"parentMessageIds"`
	// The payload within the message.
	Payload *json.RawMessage `json:"payload"`
	// The nonce the message used to fulfill the PoW requirement.
	Nonce string `json:"nonce"`
}

func (jm *jsonMessage) ToSerializable() (serializer.Serializable, error) {
	var err error

	m := &Message{}

	var parsedNetworkID uint64
	if len(jm.NetworkID) != 0 {
		parsedNetworkID, err = strconv.ParseUint(jm.NetworkID, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse message network ID from JSON: %w", err)
		}
	}
	m.NetworkID = parsedNetworkID

	var parsedNonce uint64
	if len(jm.Nonce) != 0 {
		parsedNonce, err = strconv.ParseUint(jm.Nonce, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse message nonce from JSON: %w", err)
		}
	}
	m.Nonce = parsedNonce

	m.Parents = make(MessageIDs, len(jm.Parents))
	for i, jparent := range jm.Parents {
		parentBytes, err := hex.DecodeString(jparent)
		if err != nil {
			return nil, fmt.Errorf("unable to decode hex parent %d from JSON: %w", i+1, err)
		}

		copy(m.Parents[i][:], parentBytes)
	}

	if jm.Payload != nil {
		jsonPayload, err := DeserializeObjectFromJSON(jm.Payload, jsonPayloadSelector)
		if err != nil {
			return nil, err
		}

		m.Payload, err = jsonPayload.ToSerializable()
		if err != nil {
			return nil, err
		}
	}

	return m, nil
}
