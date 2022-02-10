package builder

import (
	"context"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/nodeclient"
	"github.com/iotaledger/iota.go/v3/pow"
)

// NewMessageBuilder creates a new MessageBuilder.
func NewMessageBuilder() *MessageBuilder {
	return &MessageBuilder{
		msg: &iotago.Message{
			ProtocolVersion: iotago.ProtocolVersion,
		},
	}
}

// MessageBuilder is used to easily build up a Message.
type MessageBuilder struct {
	msg *iotago.Message
	err error
}

// Build builds the Message or returns any error which occurred during the build steps.
func (mb *MessageBuilder) Build() (*iotago.Message, error) {
	if mb.err != nil {
		return nil, mb.err
	}
	return mb.msg, nil
}

// Payload sets the payload to embed within the message.
func (mb *MessageBuilder) Payload(payload iotago.Payload) *MessageBuilder {
	if mb.err != nil {
		return mb
	}
	mb.msg.Payload = payload
	return mb
}

// Tips uses the given Client to query for parents to use.
func (mb *MessageBuilder) Tips(ctx context.Context, nodeAPI *nodeclient.Client) *MessageBuilder {
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

	pars := make(iotago.MessageIDs, len(parents))
	for i, parentBytes := range parents {
		parent := iotago.MessageID{}
		copy(parent[:], parentBytes)
		pars[i] = parent
	}
	mb.msg.Parents = serializer.RemoveDupsAndSortByLexicalOrderArrayOf32Bytes(pars)
	return mb
}

// ParentsMessageIDs sets the parents of the message.
func (mb *MessageBuilder) ParentsMessageIDs(parents iotago.MessageIDs) *MessageBuilder {
	if mb.err != nil {
		return mb
	}

	mb.msg.Parents = serializer.RemoveDupsAndSortByLexicalOrderArrayOf32Bytes(parents)
	return mb
}

// ProofOfWork does the proof-of-work needed in order to satisfy the given target score.
// It can be cancelled by cancelling the given context. This function should appear
// as the last step before Build.
func (mb *MessageBuilder) ProofOfWork(ctx context.Context, deSeriPara *iotago.DeSerializationParameters, targetScore float64, numWorkers ...int) *MessageBuilder {
	if mb.err != nil {
		return mb
	}

	msgData, err := mb.msg.Serialize(serializer.DeSeriModePerformValidation, deSeriPara)
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
