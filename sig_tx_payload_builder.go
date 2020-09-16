package iota

import (
	"fmt"
)

// NewSignedTransactionBuilder creates a new SignedTransactionPayloadBuilder.
func NewSignedTransactionBuilder() *SignedTransactionPayloadBuilder {
	return &SignedTransactionPayloadBuilder{
		unsigTx: &UnsignedTransaction{
			Inputs:  Serializables{},
			Outputs: Serializables{},
			Payload: nil,
		},
		inputToAddr: map[UTXOInputID]Serializable{},
	}
}

// SignedTransactionPayloadBuilder is used to easily build up a transaction.
type SignedTransactionPayloadBuilder struct {
	unsigTx     *UnsignedTransaction
	inputToAddr map[UTXOInputID]Serializable
}

// ToBeSignedUTXOInput defines a UTXO input which needs to be signed.
type ToBeSignedUTXOInput struct {
	// The address to which this input belongs to.
	Address Serializable `json:"address"`
	// The actual UTXO input.
	Input *UTXOInput `json:"input"`
}

// AddInput adds the given input to the builder.
func (b *SignedTransactionPayloadBuilder) AddInput(input *ToBeSignedUTXOInput) *SignedTransactionPayloadBuilder {
	b.inputToAddr[input.Input.ID()] = input.Address
	b.unsigTx.Inputs = append(b.unsigTx.Inputs, input.Input)
	return b
}

// AddOutput adds the given output to the builder.
func (b *SignedTransactionPayloadBuilder) AddOutput(output *SigLockedSingleDeposit) *SignedTransactionPayloadBuilder {
	b.unsigTx.Outputs = append(b.unsigTx.Outputs, output)
	return b
}

// AddIndexationPayload adds the given IndexationPayload as the inner payload.
func (b *SignedTransactionPayloadBuilder) AddIndexationPayload(payload *IndexationPayload) *SignedTransactionPayloadBuilder {
	b.unsigTx.Payload = payload
	return b
}

// Build sings the inputs with the given signer and returns the built payload.
func (b *SignedTransactionPayloadBuilder) Build(signer AddressSigner) (*SignedTransactionPayload, error) {

	// sort inputs and outputs by their serialized byte order
	txDataToBeSigned, err := b.unsigTx.Serialize(DeSeriModePerformValidation | DeSeriModePerformLexicalOrdering)
	if err != nil {
		return nil, err
	}

	sigBlockPos := map[string]int{}
	unlockBlocks := Serializables{}
	for i, input := range b.unsigTx.Inputs {
		addr := b.inputToAddr[input.(*UTXOInput).ID()]
		addrStr := addr.(fmt.Stringer).String()

		// check whether a previous signature unlock block
		// already signs inputs for the given address
		pos, alreadySigned := sigBlockPos[addrStr]
		if alreadySigned {
			// create a reference unlock block instead
			unlockBlocks = append(unlockBlocks, &ReferenceUnlockBlock{Reference: uint16(pos)})
			continue
		}

		// create a new signature for the given address
		var signature Serializable
		signature, err = signer.Sign(addr, txDataToBeSigned)
		if err != nil {
			return nil, err
		}

		unlockBlocks = append(unlockBlocks, &SignatureUnlockBlock{Signature: signature})
		sigBlockPos[addrStr] = i
	}

	sigTxPayload := &SignedTransactionPayload{Transaction: b.unsigTx, UnlockBlocks: unlockBlocks}

	return sigTxPayload, nil
}
