package iotago

import (
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
)

const (
	// MinContextInputsCount defines the minimum amount of context inputs within a Transaction.
	MinContextInputsCount = 0
	// MaxContextInputsCount defines the maximum amount of context inputs within a Transaction.
	MaxContextInputsCount = 128
	// MaxInputsCount defines the maximum amount of inputs within a Transaction.
	MaxInputsCount = 128
	// MinInputsCount defines the minimum amount of inputs within a Transaction.
	MinInputsCount = 1

	// MinAllotmentCount defines the minimum amount of allotments within a Transaction.
	MinAllotmentCount = 0
	// MaxAllotmentCount defines the maximum amount of allotments within a Transaction.
	MaxAllotmentCount = 128
)

type (
	txEssenceContextInput  interface{ ContextInput }
	txEssenceInput         interface{ Input }
	TxEssenceOutput        interface{ Output }
	TxEssencePayload       interface{ Payload }
	TxEssenceContextInputs = ContextInputs[txEssenceContextInput]
	TxEssenceInputs        = Inputs[txEssenceInput]
	TxEssenceAllotments    = Allotments
)

// TransactionEssence is the essence part if a Transaction.
type TransactionEssence struct {
	// The network ID for which this essence is valid for.
	NetworkID NetworkID `serix:""`
	// The slot index in which the transaction was created by the client.
	CreationSlot SlotIndex `serix:""`
	// The commitment references of this transaction.
	ContextInputs TxEssenceContextInputs `serix:",omitempty"`
	// The inputs of this transaction.
	Inputs TxEssenceInputs `serix:""`
	// The optional accounts map with corresponding allotment values.
	Allotments TxEssenceAllotments `serix:",omitempty"`
	// The capabilities of the transaction.
	Capabilities TransactionCapabilitiesBitMask `serix:",omitempty"`
	// The optional embedded payload.
	Payload TxEssencePayload `serix:",optional"`
}

func (u *TransactionEssence) Clone() *TransactionEssence {
	var payload TxEssencePayload
	if u.Payload != nil {
		payload = u.Payload.Clone()
	}

	return &TransactionEssence{
		NetworkID:     u.NetworkID,
		CreationSlot:  u.CreationSlot,
		ContextInputs: u.ContextInputs.Clone(),
		Inputs:        u.Inputs.Clone(),
		Allotments:    u.Allotments.Clone(),
		Capabilities:  u.Capabilities.Clone(),
		Payload:       payload,
	}
}

func (u *TransactionEssence) Size() int {
	payloadSize := serializer.PayloadLengthByteSize
	if u.Payload != nil {
		payloadSize = u.Payload.Size()
	}

	// NetworkID
	return serializer.UInt64ByteSize +
		// CreationSlot
		SlotIndexLength +
		u.ContextInputs.Size() +
		u.Inputs.Size() +
		u.Allotments.Size() +
		u.Capabilities.Size() +
		payloadSize
}

// WorkScore calculates the Work Score of the TransactionEssence.
//
// Does not specifically include the work score of the optional payload because that is already
// included in the Work Score of the SignedTransaction.
func (u *TransactionEssence) WorkScore(workScoreParameters *WorkScoreParameters) (WorkScore, error) {
	workScoreContextInputs, err := u.ContextInputs.WorkScore(workScoreParameters)
	if err != nil {
		return 0, err
	}

	workScoreInputs, err := u.Inputs.WorkScore(workScoreParameters)
	if err != nil {
		return 0, err
	}

	workScoreAllotments, err := u.Allotments.WorkScore(workScoreParameters)
	if err != nil {
		return 0, err
	}

	return workScoreContextInputs.Add(workScoreInputs, workScoreAllotments)
}

// SyntacticallyValidate checks whether the transaction essence is syntactically valid.
// The function does not syntactically validate the input or outputs themselves.
//
// This function includes "essence" in its name to disambiguate it from the "SyntacticallyValidate" function on the Transaction,
// and since the essence is embedded in the Transaction a similar name could easily lead to confusion.
func (u *TransactionEssence) syntacticallyValidateEssence(api API) error {
	if err := BitMaskNonTrailingZeroBytesValidatorFunc(u.Capabilities); err != nil {
		return ierrors.Wrapf(ErrTxEssenceCapabilitiesInvalid, "invalid capabilities bitmask: %w", err)
	}

	protoParams := api.ProtocolParameters()
	expectedNetworkID := protoParams.NetworkID()
	if u.NetworkID != expectedNetworkID {
		return ierrors.Wrapf(ErrTxEssenceNetworkIDInvalid, "got %v, want %v (%s)", u.NetworkID, expectedNetworkID, protoParams.NetworkName())
	}

	err := SliceValidatorMapper(u.Inputs,
		func(input txEssenceInput) Input { return input },
		InputsSyntacticalUnique(),
		InputsSyntacticalIndicesWithinBounds(),
	)
	if err != nil {
		return err
	}

	// cast is safe since we just validated that at most MaxInputsCount inputs are added which is less than what fits into a uint16.
	inputsCount := uint16(len(u.Inputs))
	err = SliceValidatorMapper(u.ContextInputs,
		func(input txEssenceContextInput) ContextInput { return input },
		LexicalOrderAndUniquenessValidator[ContextInput](),
		ContextInputsRewardInputMaxIndex(inputsCount),
		ContextInputsCommitmentInputRequirement(),
	)
	if err != nil {
		return err
	}

	var maxManaValue Mana = (1 << protoParams.ManaParameters().BitsCount) - 1

	return SliceValidator(
		u.Allotments,
		LexicalOrderAndUniquenessValidator[*Allotment](),
		allotmentMaxManaValidator(maxManaValue),
	)
}
