package iotago

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/iota.go/v3/util"
)

// TransactionEssenceType defines the type of transaction.
type TransactionEssenceType = byte

const (
	// TransactionEssenceNormal denotes a standard transaction essence.
	TransactionEssenceNormal TransactionEssenceType = 1

	// MaxInputsCount defines the maximum amount of inputs within a TransactionEssence.
	MaxInputsCount = 128
	// MinInputsCount defines the minimum amount of inputs within a TransactionEssence.
	MinInputsCount = 1
	// MaxOutputsCount defines the maximum amount of outputs within a TransactionEssence.
	MaxOutputsCount = 128
	// MinOutputsCount defines the minimum amount of inputs within a TransactionEssence.
	MinOutputsCount = 1

	// InputsCommitmentLength defines the length of the inputs commitment hash.
	InputsCommitmentLength = blake2b.Size256
)

var (
	// ErrInvalidInputsCommitment gets returned when the inputs commitment is invalid.
	ErrInvalidInputsCommitment = errors.New("invalid inputs commitment")
	// ErrTxEssenceNetworkIDInvalid gets returned when a network ID within a TransactionEssence is invalid.
	ErrTxEssenceNetworkIDInvalid = errors.New("invalid network ID")
	// ErrInputUTXORefsNotUnique gets returned if multiple inputs reference the same UTXO.
	ErrInputUTXORefsNotUnique = errors.New("inputs must each reference a unique UTXO")
	// ErrAliasOutputNonEmptyState gets returned if an AliasOutput with zeroed AliasID contains state (counters non-zero etc.).
	ErrAliasOutputNonEmptyState = errors.New("alias output is not empty state")
	// ErrAliasOutputCyclicAddress gets returned if an AliasOutput's AliasID results into the same address as the State/Governance controller.
	ErrAliasOutputCyclicAddress = errors.New("alias output's AliasID corresponds to state and/or governance controller")
	// ErrNFTOutputCyclicAddress gets returned if an NFTOutput's NFTID results into the same address as the address field within the output.
	ErrNFTOutputCyclicAddress = errors.New("NFT output's ID corresponds to address field")
	// ErrOutputsSumExceedsTotalSupply gets returned if the sum of the output deposits exceeds the total supply of tokens.
	ErrOutputsSumExceedsTotalSupply = errors.New("accumulated output balance exceeds total supply")
	// ErrOutputDepositsMoreThanTotalSupply gets returned if an output deposits more than the total supply.
	ErrOutputDepositsMoreThanTotalSupply = errors.New("an output can not deposit more than the total supply")
	// ErrStorageDepositLessThanMinReturnOutputStorageDeposit gets returned when the storage deposit condition's amount is less than the min storage deposit for the return output.
	ErrStorageDepositLessThanMinReturnOutputStorageDeposit = errors.New("storage deposit return amount is less than the min storage deposit needed for the return output")
	// ErrStorageDepositExceedsTargetOutputDeposit gets returned when the storage deposit condition's amount exceeds the target output's deposit.
	ErrStorageDepositExceedsTargetOutputDeposit = errors.New("storage deposit return amount exceeds target output's deposit")
	// ErrMaxNativeTokensCountExceeded gets returned if outputs or transactions exceed the MaxNativeTokensCount.
	ErrMaxNativeTokensCountExceeded = errors.New("max native tokens count exceeded")
)

// TransactionEssenceSelector implements SerializableSelectorFunc for transaction essence types.
func TransactionEssenceSelector(txType uint32) (*TransactionEssence, error) {
	var seri *TransactionEssence
	switch byte(txType) {
	case TransactionEssenceNormal:
		seri = &TransactionEssence{}
	default:
		return nil, fmt.Errorf("%w: type byte %d", ErrUnknownTransactionEssenceType, txType)
	}
	return seri, nil
}

// InputsCommitment is a commitment to the inputs of a transaction.
type InputsCommitment = [InputsCommitmentLength]byte

type TxEssencePayload interface{ Payload }
type TxEssenceInput interface{ Input }
type TxEssenceOutput interface{ Output }

// TransactionEssence is the essence part of a Transaction.
type TransactionEssence struct {
	// The network ID for which this essence is valid for.
	NetworkID NetworkID `serix:"0,mapKey=networkId"`
	// The inputs of this transaction.
	Inputs Inputs[TxEssenceInput] `serix:"1,mapKey=inputs"`
	// The commitment to the referenced inputs.
	InputsCommitment InputsCommitment `serix:"2,mapKey=inputsCommitment"`
	// The outputs of this transaction.
	Outputs Outputs[TxEssenceOutput] `serix:"3,mapKey=outputs"`
	// The optional embedded payload.
	Payload TxEssencePayload `serix:"4,optional,mapKey=payload"`
}

// SigningMessage returns the to be signed message.
func (u *TransactionEssence) SigningMessage() ([]byte, error) {
	essenceBytes, err := internalEncode(u)
	if err != nil {
		return nil, err
	}
	essenceBytesHash := blake2b.Sum256(essenceBytes)
	return essenceBytesHash[:], nil
}

// Sign produces signatures signing the essence for every given AddressKeys.
// The produced signatures are in the same order as the AddressKeys.
func (u *TransactionEssence) Sign(inputsCommitment []byte, addrKeys ...AddressKeys) ([]Signature, error) {
	if inputsCommitment == nil || len(inputsCommitment) != InputsCommitmentLength {
		return nil, ErrInvalidInputsCommitment
	}

	copy(u.InputsCommitment[:], inputsCommitment)

	signMsg, err := u.SigningMessage()
	if err != nil {
		return nil, err
	}

	sigs := make([]Signature, len(addrKeys))
	signer := NewInMemoryAddressSigner(addrKeys...)
	for i, v := range addrKeys {
		sig, err := signer.Sign(v.Address, signMsg)
		if err != nil {
			return nil, err
		}
		sigs[i] = sig
	}

	return sigs, nil
}

func (u *TransactionEssence) Size() int {
	payloadSize := util.NumByteLen(uint32(0))
	if u.Payload != nil {
		payloadSize = u.Payload.Size()
	}
	return util.NumByteLen(TransactionEssenceNormal) +
		util.NumByteLen(u.NetworkID) +
		u.Inputs.Size() +
		InputsCommitmentLength +
		u.Outputs.Size() +
		payloadSize
}

// syntacticallyValidate checks whether the transaction essence is syntactically valid.
// The function does not syntactically validate the input or outputs themselves.
func (u *TransactionEssence) syntacticallyValidate(protoParas *ProtocolParameters) error {

	expectedNetworkID := protoParas.NetworkID()
	if u.NetworkID != expectedNetworkID {
		return fmt.Errorf("%w: got %v, want %v (%s)", ErrTxEssenceNetworkIDInvalid, u.NetworkID, expectedNetworkID, protoParas.NetworkName)
	}

	if err := SyntacticallyValidateInputs(u.Inputs,
		InputsSyntacticalUnique(),
		InputsSyntacticalIndicesWithinBounds(),
	); err != nil {
		return err
	}

	if err := SyntacticallyValidateOutputs(u.Outputs,
		OutputsSyntacticalDepositAmount(protoParas),
		OutputsSyntacticalExpirationAndTimelock(),
		OutputsSyntacticalNativeTokens(),
		OutputsSyntacticalFoundry(),
		OutputsSyntacticalAlias(),
		OutputsSyntacticalNFT(),
	); err != nil {
		return err
	}

	return nil
}
