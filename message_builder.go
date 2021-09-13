package iotago

import (
	"context"
	"fmt"

	"github.com/iotaledger/hive.go/serializer"
	"github.com/iotaledger/iota.go/v2/pow"
)

// NewMessageBuilder creates a new MessageBuilder.
func NewMessageBuilder() *MessageBuilder {
	return &MessageBuilder{
		msg: &Message{},
	}
}

// MessageBuilder is used to easily build up a Message.
type MessageBuilder struct {
	msg *Message
	err error
}

// Build builds the Message or returns any error which occurred during the build steps.
func (mb *MessageBuilder) Build() (*Message, error) {
	if mb.err != nil {
		return nil, mb.err
	}
	return mb.msg, nil
}

// NetworkID sets the network ID for which this message is meant for.
func (mb *MessageBuilder) NetworkID(networkID uint64) *MessageBuilder {
	if mb.err != nil {
		return mb
	}
	mb.msg.NetworkID = networkID
	return mb
}

// NetworkIDFromString sets the network ID for which this message is meant for.
func (mb *MessageBuilder) NetworkIDFromString(networkIDStr string) *MessageBuilder {
	if mb.err != nil {
		return mb
	}
	mb.msg.NetworkID = NetworkIDFromString(networkIDStr)
	return mb
}

// Payload sets the payload to embed within the message.
func (mb *MessageBuilder) Payload(seri serializer.Serializable) *MessageBuilder {
	if mb.err != nil {
		return mb
	}
	switch seri.(type) {
	case *Indexation:
	case *Milestone:
	case *Transaction:
	case nil:
	default:
		mb.err = fmt.Errorf("%w: unsupported type %T", ErrUnknownPayloadType, seri)
		return mb
	}
	mb.msg.Payload = seri
	return mb
}

// Tips uses the given NodeHTTPAPIClient to query for parents to use.
func (mb *MessageBuilder) Tips(ctx context.Context, nodeAPI *NodeHTTPAPIClient) *MessageBuilder {
	if mb.err != nil {
		return mb
	}

	res, err := nodeAPI.Tips(ctx)
	if err != nil {
		mb.err = fmt.Errorf("unable to fetch tips from node API: %w", err)
		return mb
	}

	parents, err := res.Tips()
	if err != nil {
		mb.err = fmt.Errorf("unable to fetch tips: %w", err)
		return mb
	}

	mb.ParentsMessageIDs(parents)

	return mb
}

// Parents sets the parents of the message.
func (mb *MessageBuilder) Parents(parents [][]byte) *MessageBuilder {
	if mb.err != nil {
		return mb
	}

	pars := make(MessageIDs, len(parents))
	for i, parentBytes := range parents {
		parent := MessageID{}
		copy(parent[:], parentBytes)
		pars[i] = parent
	}
	mb.msg.Parents = serializer.RemoveDupsAndSortByLexicalOrderArrayOf32Bytes(pars)
	return mb
}

// ParentsMessageIDs sets the parents of the message.
func (mb *MessageBuilder) ParentsMessageIDs(parents MessageIDs) *MessageBuilder {
	if mb.err != nil {
		return mb
	}

	mb.msg.Parents = serializer.RemoveDupsAndSortByLexicalOrderArrayOf32Bytes(parents)
	return mb
}

// ProofOfWork does the proof-of-work needed in order to satisfy the given target score.
// It can be cancelled by cancelling the given context. This function should appear
// as the last step before Build.
func (mb *MessageBuilder) ProofOfWork(ctx context.Context, targetScore float64, numWorkers ...int) *MessageBuilder {
	if mb.err != nil {
		return mb
	}
	msgData, err := mb.msg.Serialize(serializer.DeSeriModePerformValidation)
	if err != nil {
		mb.err = err
		return mb
	}

	// cut out the nonce
	powRelevantData := msgData[:len(msgData)-serializer.UInt64ByteSize]
	worker := pow.New(numWorkers...)
	nonce, err := worker.Mine(ctx, powRelevantData, targetScore)
	if err != nil {
		mb.err = fmt.Errorf("unable to complete proof-of-work: %w", err)
		return mb
	}
	mb.msg.Nonce = nonce
	return mb
}
