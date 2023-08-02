package iotago

import (
	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
)

// TransactionEssenceType defines the type of transaction.
type TransactionEssenceType = byte

const (
	// TransactionEssenceNormal denotes a standard transaction essence.
	TransactionEssenceNormal TransactionEssenceType = 2

	// MinContextInputsCount defines the minimum amount of context inputs within a TransactionEssence.
	MinContextInputsCount = 0
	// MaxContextInputsCount defines the maximum amount of context inputs within a TransactionEssence.
	MaxContextInputsCount = 128
	// MaxInputsCount defines the maximum amount of inputs within a TransactionEssence.
	MaxInputsCount = 128
	// MinInputsCount defines the minimum amount of inputs within a TransactionEssence.
	MinInputsCount = 1
	// MaxOutputsCount defines the maximum amount of outputs within a TransactionEssence.
	MaxOutputsCount = 128
	// MinOutputsCount defines the minimum amount of inputs within a TransactionEssence.
	MinOutputsCount = 1
	// MinAllotmentCount defines the minimum amount of allotments within a TransactionEssence.
	MinAllotmentCount = 0
	// MaxAllotmentCount defines the maximum amount of allotments within a TransactionEssence.
	MaxAllotmentCount = 128

	// InputsCommitmentLength defines the length of the inputs commitment hash.
	InputsCommitmentLength = blake2b.Size256
)

var (
	// ErrInvalidInputsCommitment gets returned when the inputs commitment is invalid.
	ErrInvalidInputsCommitment = ierrors.New("invalid inputs commitment")
	// ErrTxEssenceNetworkIDInvalid gets returned when a network ID within a TransactionEssence is invalid.
	ErrTxEssenceNetworkIDInvalid = ierrors.New("invalid network ID")
	// ErrAllotmentsNotUnique gets returned if multiple Allotments reference the same Account.
	ErrAllotmentsNotUnique = ierrors.New("allotments must each reference a unique account")
	// ErrInputUTXORefsNotUnique gets returned if multiple inputs reference the same UTXO.
	ErrInputUTXORefsNotUnique = ierrors.New("inputs must each reference a unique UTXO")
	// ErrInputBICNotUnique gets returned if multiple inputs reference the same BIC.
	ErrInputBICNotUnique = ierrors.New("inputs must each reference a unique BIC")
	// ErrInputRewardInvalid gets returned if multiple reward inputs reference the same input index
	// or if they reference an index greater than max inputs count.
	ErrInputRewardInvalid = ierrors.New("invalid reward input")
	// ErrMultipleInputCommitments gets returned if multiple commitment inputs are provided.
	ErrMultipleInputCommitments = ierrors.New("there are multiple commitment inputs")
	// ErrAccountOutputNonEmptyState gets returned if an AccountOutput with zeroed AccountID contains state (counters non-zero etc.).
	ErrAccountOutputNonEmptyState = ierrors.New("account output is not empty state")
	// ErrAccountOutputCyclicAddress gets returned if an AccountOutput's AccountID results into the same address as the State/Governance controller.
	ErrAccountOutputCyclicAddress = ierrors.New("account output's AccountID corresponds to state and/or governance controller")
	// ErrNFTOutputCyclicAddress gets returned if an NFTOutput's NFTID results into the same address as the address field within the output.
	ErrNFTOutputCyclicAddress = ierrors.New("NFT output's ID corresponds to address field")
	// ErrDelegationValidatorIDZeroed gets returned if a Delegation Output's Validator ID is zeroed out.
	ErrDelegationValidatorIDZeroed = ierrors.New("delegation output's validator ID is zeroed")
	// ErrOutputsSumExceedsTotalSupply gets returned if the sum of the output deposits exceeds the total supply of tokens.
	ErrOutputsSumExceedsTotalSupply = ierrors.New("accumulated output balance exceeds total supply")
	// ErrOutputAmountMoreThanTotalSupply gets returned if an output base token amount is more than the total supply.
	ErrOutputAmountMoreThanTotalSupply = ierrors.New("an output's base token amount cannot exceed the total supply")
	// ErrStorageDepositLessThanMinReturnOutputStorageDeposit gets returned when the storage deposit condition's amount is less than the min storage deposit for the return output.
	ErrStorageDepositLessThanMinReturnOutputStorageDeposit = ierrors.New("storage deposit return amount is less than the min storage deposit needed for the return output")
	// ErrStorageDepositExceedsTargetOutputAmount gets returned when the storage deposit condition's amount exceeds the target output's base token amount.
	ErrStorageDepositExceedsTargetOutputAmount = ierrors.New("storage deposit return amount exceeds target output's base token amount")
	// ErrMaxNativeTokensCountExceeded gets returned if outputs or transactions exceed the MaxNativeTokensCount.
	ErrMaxNativeTokensCountExceeded = ierrors.New("max native tokens count exceeded")
)

// TransactionEssenceSelector implements SerializableSelectorFunc for transaction essence types.
func TransactionEssenceSelector(txType uint32) (*TransactionEssence, error) {
	var seri *TransactionEssence
	switch byte(txType) {
	case TransactionEssenceNormal:
		seri = &TransactionEssence{}
	default:
		return nil, ierrors.Wrapf(ErrUnknownTransactionEssenceType, "type byte %d", txType)
	}

	return seri, nil
}

// InputsCommitment is a commitment to the inputs of a transaction.
type InputsCommitment = [InputsCommitmentLength]byte

type (
	txEssenceContextInput  interface{ ContextInput }
	txEssenceInput         interface{ Input }
	TxEssenceOutput        interface{ Output }
	TxEssencePayload       interface{ Payload }
	TxEssenceContextInputs = ContextInputs[txEssenceContextInput]
	TxEssenceInputs        = Inputs[txEssenceInput]
	TxEssenceOutputs       = Outputs[TxEssenceOutput]
	TxEssenceAllotments    = Allotments
)

// TransactionEssence is the essence part of a Transaction.
type TransactionEssence struct {
	// The network ID for which this essence is valid for.
	NetworkID NetworkID `serix:"0,mapKey=networkId"`
	// The time at which this transaction was created by the client.
	CreationTime SlotIndex `serix:"1,mapKey=creationTime"`
	// The commitment references of this transaction.
	ContextInputs TxEssenceContextInputs `serix:"2,mapKey=contextInputs"`
	// The inputs of this transaction.
	Inputs TxEssenceInputs `serix:"3,mapKey=inputs"`
	// The commitment to the referenced inputs.
	InputsCommitment InputsCommitment `serix:"4,mapKey=inputsCommitment"`
	// The outputs of this transaction.
	Outputs TxEssenceOutputs `serix:"5,mapKey=outputs"`
	// The optional accounts map with corresponding allotment values.
	Allotments TxEssenceAllotments `serix:"6,mapKey=allotments"`
	// The optional embedded payload.
	Payload TxEssencePayload `serix:"7,optional,mapKey=payload"`
}

// SigningMessage returns the to be signed message.
func (u *TransactionEssence) SigningMessage(api API) ([]byte, error) {
	essenceBytes, err := api.Encode(u)
	if err != nil {
		return nil, err
	}
	essenceBytesHash := blake2b.Sum256(essenceBytes)

	return essenceBytesHash[:], nil
}

// Sign produces signatures signing the essence for every given AddressKeys.
// The produced signatures are in the same order as the AddressKeys.
func (u *TransactionEssence) Sign(api API, inputsCommitment []byte, addrKeys ...AddressKeys) ([]Signature, error) {
	if inputsCommitment == nil || len(inputsCommitment) != InputsCommitmentLength {
		return nil, ErrInvalidInputsCommitment
	}

	copy(u.InputsCommitment[:], inputsCommitment)

	signMsg, err := u.SigningMessage(api)
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
	payloadSize := serializer.UInt32ByteSize
	if u.Payload != nil {
		payloadSize = u.Payload.Size()
	}

	// TransactionEssenceType
	return serializer.OneByte +
		// NetworkID
		serializer.UInt64ByteSize +
		// CreationTime
		SlotIndexLength +
		u.ContextInputs.Size() +
		u.Inputs.Size() +
		// InputsCommitment
		InputsCommitmentLength +
		u.Outputs.Size() +
		u.Allotments.Size() +
		payloadSize
}

// syntacticallyValidate checks whether the transaction essence is syntactically valid.
// The function does not syntactically validate the input or outputs themselves.
func (u *TransactionEssence) syntacticallyValidate(protoParams ProtocolParameters) error {
	expectedNetworkID := protoParams.NetworkID()
	if u.NetworkID != expectedNetworkID {
		return ierrors.Wrapf(ErrTxEssenceNetworkIDInvalid, "got %v, want %v (%s)", u.NetworkID, expectedNetworkID, protoParams.NetworkName())
	}

	if err := SyntacticallyValidateContextInputs(u.ContextInputs,
		ContextInputsSyntacticalUnique(),
	); err != nil {
		return err
	}

	if err := SyntacticallyValidateInputs(u.Inputs,
		InputsSyntacticalUnique(),
		InputsSyntacticalIndicesWithinBounds(),
	); err != nil {
		return err
	}

	if err := SyntacticallyValidateOutputs(u.Outputs,
		OutputsSyntacticalDepositAmount(protoParams),
		OutputsSyntacticalExpirationAndTimelock(),
		OutputsSyntacticalNativeTokens(),
		OutputsSyntacticalChainConstrainedOutputUniqueness(),
		OutputsSyntacticalFoundry(),
		OutputsSyntacticalAccount(),
		OutputsSyntacticalNFT(),
		OutputsSyntacticalDelegation(),
	); err != nil {
		return err
	}

	return SyntacticallyValidateAllotments(u.Allotments,
		AllotmentsSyntacticalUnique(),
	)
}

func (u *TransactionEssence) WorkScore(workScoreStructure *WorkScoreStructure) (WorkScore, error) {
	// TransactionEssenceType + NetworkID + CreationTime + InputsCommitment
	workScoreBytes, err := workScoreStructure.DataByte.Multiply(serializer.OneByte + serializer.UInt64ByteSize + serializer.UInt64ByteSize + InputsCommitmentLength)
	if err != nil {
		return 0, err
	}

	workScoreContextInputs, err := u.ContextInputs.WorkScore(workScoreStructure)
	if err != nil {
		return 0, err
	}

	workScoreInputs, err := u.Inputs.WorkScore(workScoreStructure)
	if err != nil {
		return 0, err
	}

	workScoreOutputs, err := u.Outputs.WorkScore(workScoreStructure)
	if err != nil {
		return 0, err
	}

	workScoreAllotments, err := u.Allotments.WorkScore(workScoreStructure)
	if err != nil {
		return 0, err
	}

	workScorePayload, err := u.Payload.WorkScore(workScoreStructure)
	if err != nil {
		return 0, err
	}

	return workScoreBytes.Add(workScoreContextInputs, workScoreInputs, workScoreOutputs, workScoreAllotments, workScorePayload)
}
