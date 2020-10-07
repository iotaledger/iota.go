package iota

import (
	"fmt"
)

// NewTransactionBuilder creates a new TransactionBuilder.
func NewTransactionBuilder() *TransactionBuilder {
	return &TransactionBuilder{
		essence: &TransactionEssence{
			Inputs:  Serializables{},
			Outputs: Serializables{},
			Payload: nil,
		},
		inputToAddr: map[UTXOInputID]Serializable{},
	}
}

// TransactionBuilder is used to easily build up a transaction.
type TransactionBuilder struct {
	essence     *TransactionEssence
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
func (b *TransactionBuilder) AddInput(input *ToBeSignedUTXOInput) *TransactionBuilder {
	b.inputToAddr[input.Input.ID()] = input.Address
	b.essence.Inputs = append(b.essence.Inputs, input.Input)
	return b
}

// AddOutput adds the given output to the builder.
func (b *TransactionBuilder) AddOutput(output *SigLockedSingleOutput) *TransactionBuilder {
	b.essence.Outputs = append(b.essence.Outputs, output)
	return b
}

// AddIndexationPayload adds the given Indexation as the inner payload.
func (b *TransactionBuilder) AddIndexationPayload(payload *Indexation) *TransactionBuilder {
	b.essence.Payload = payload
	return b
}

// Build sings the inputs with the given signer and returns the built payload.
func (b *TransactionBuilder) Build(signer AddressSigner) (*Transaction, error) {

	// sort inputs and outputs by their serialized byte order
	txEssenceData, err := b.essence.Serialize(DeSeriModePerformValidation | DeSeriModePerformLexicalOrdering)
	if err != nil {
		return nil, err
	}

	sigBlockPos := map[string]int{}
	unlockBlocks := Serializables{}
	for i, input := range b.essence.Inputs {
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
		signature, err = signer.Sign(addr, txEssenceData)
		if err != nil {
			return nil, err
		}

		unlockBlocks = append(unlockBlocks, &SignatureUnlockBlock{Signature: signature})
		sigBlockPos[addrStr] = i
	}

	sigTxPayload := &Transaction{Essence: b.essence, UnlockBlocks: unlockBlocks}

	return sigTxPayload, nil
}
