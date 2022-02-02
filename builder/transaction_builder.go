package builder

import (
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v3"
)

var (
	// ErrTransactionBuilder defines a generic error occurring within the TransactionBuilder.
	ErrTransactionBuilder = errors.New("transaction builder error")
)

// NewTransactionBuilder creates a new TransactionBuilder.
func NewTransactionBuilder() *TransactionBuilder {
	return &TransactionBuilder{
		essence: &iotago.TransactionEssence{
			Inputs:  iotago.Inputs{},
			Outputs: iotago.Outputs{},
			Payload: nil,
		},
		inputToAddr: map[iotago.OutputID]iotago.Address{},
	}
}

// TransactionBuilder is used to easily build up a Transaction.
type TransactionBuilder struct {
	occurredBuildErr error
	essence          *iotago.TransactionEssence
	inputToAddr      map[iotago.OutputID]iotago.Address
}

// ToBeSignedUTXOInput defines a UTXO input which needs to be signed.
type ToBeSignedUTXOInput struct {
	// The address to which this input belongs to.
	Address iotago.Address `json:"address"`
	// The actual UTXO input.
	Input *iotago.UTXOInput `json:"input"`
}

// AddInput adds the given input to the builder.
func (b *TransactionBuilder) AddInput(input *ToBeSignedUTXOInput) *TransactionBuilder {
	b.inputToAddr[input.Input.ID()] = input.Address
	b.essence.Inputs = append(b.essence.Inputs, input.Input)
	return b
}

// TransactionBuilderInputFilter is a filter function which determines whether
// an input should be used or not. (returning true = pass). The filter can also
// be used to accumulate data over the set of inputs, i.e. the input sum etc.
type TransactionBuilderInputFilter func(outputID iotago.OutputID, input iotago.Output) bool

// AddOutput adds the given output to the builder.
func (b *TransactionBuilder) AddOutput(output iotago.Output) *TransactionBuilder {
	b.essence.Outputs = append(b.essence.Outputs, output)
	return b
}

// AddTaggedDataPayload adds the given TaggedData as the inner payload.
func (b *TransactionBuilder) AddTaggedDataPayload(payload *iotago.TaggedData) *TransactionBuilder {
	b.essence.Payload = payload
	return b
}

// TransactionFunc is a function which receives a Transaction as its parameter.
type TransactionFunc func(tx *iotago.Transaction)

// BuildAndSwapToMessageBuilder builds the transaction and then swaps to a MessageBuilder with
// the transaction set as its payload. txFunc can be nil.
func (b *TransactionBuilder) BuildAndSwapToMessageBuilder(deSeriParas *iotago.DeSerializationParameters, signer iotago.AddressSigner, txFunc TransactionFunc) *MessageBuilder {
	msgBuilder := NewMessageBuilder()
	tx, err := b.Build(deSeriParas, signer)
	if err != nil {
		msgBuilder.err = err
		return msgBuilder
	}
	if txFunc != nil {
		txFunc(tx)
	}
	return msgBuilder.Payload(tx)
}

// Build sings the inputs with the given signer and returns the built payload.
func (b *TransactionBuilder) Build(deSeriParas *iotago.DeSerializationParameters, signer iotago.AddressSigner) (*iotago.Transaction, error) {
	switch {
	case b.occurredBuildErr != nil:
		return nil, b.occurredBuildErr
	case deSeriParas == nil:
		return nil, fmt.Errorf("%w: must supply de/serialization parameters", ErrTransactionBuilder)
	case signer == nil:
		return nil, fmt.Errorf("%w: must supply signer", ErrTransactionBuilder)
	}

	// sort inputs and outputs by their serialized byte order
	txEssenceData, err := b.essence.SigningMessage()
	if err != nil {
		return nil, err
	}

	sigBlockPos := map[string]int{}
	unlockBlocks := iotago.UnlockBlocks{}
	for i, input := range b.essence.Inputs {
		addr := b.inputToAddr[input.(*iotago.UTXOInput).ID()]
		addrStr := addr.(fmt.Stringer).String()

		// check whether a previous signature unlock block
		// already signs inputs for the given address
		pos, alreadySigned := sigBlockPos[addrStr]
		if alreadySigned {
			// create a reference unlock block instead
			unlockBlocks = append(unlockBlocks, &iotago.ReferenceUnlockBlock{Reference: uint16(pos)})
			continue
		}

		// create a new signature for the given address
		var signature iotago.Signature
		signature, err = signer.Sign(addr, txEssenceData)
		if err != nil {
			return nil, err
		}

		unlockBlocks = append(unlockBlocks, &iotago.SignatureUnlockBlock{Signature: signature})
		sigBlockPos[addrStr] = i
	}

	sigTxPayload := &iotago.Transaction{Essence: b.essence, UnlockBlocks: unlockBlocks}

	if _, err := sigTxPayload.Serialize(serializer.DeSeriModePerformValidation, deSeriParas); err != nil {
		return nil, err
	}

	return sigTxPayload, nil
}