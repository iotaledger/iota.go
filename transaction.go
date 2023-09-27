package iotago

import (
	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/serializer/v2"
)

// TransactionType defines the type of transaction.
type TransactionType = byte

const (
	// TransactionNormal denotes a standard transaction essence.
	TransactionNormal TransactionType = 2

	// MinContextInputsCount defines the minimum amount of context inputs within a Transaction.
	MinContextInputsCount = 0
	// MaxContextInputsCount defines the maximum amount of context inputs within a Transaction.
	MaxContextInputsCount = 128
	// MaxInputsCount defines the maximum amount of inputs within a Transaction.
	MaxInputsCount = 128
	// MinInputsCount defines the minimum amount of inputs within a Transaction.
	MinInputsCount = 1
	// MaxOutputsCount defines the maximum amount of outputs within a Transaction.
	MaxOutputsCount = 128
	// MinOutputsCount defines the minimum amount of inputs within a Transaction.
	MinOutputsCount = 1
	// MinAllotmentCount defines the minimum amount of allotments within a Transaction.
	MinAllotmentCount = 0
	// MaxAllotmentCount defines the maximum amount of allotments within a Transaction.
	MaxAllotmentCount = 128

	// InputsCommitmentLength defines the length of the inputs commitment hash.
	InputsCommitmentLength = blake2b.Size256
)

var (
	// ErrInvalidInputsCommitment gets returned when the inputs commitment is invalid.
	ErrInvalidInputsCommitment = ierrors.New("invalid inputs commitment")
	// ErrTxEssenceNetworkIDInvalid gets returned when a network ID within a Transaction is invalid.
	ErrTxEssenceNetworkIDInvalid = ierrors.New("invalid network ID")
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
	// ErrDelegationValidatorAddressZeroed gets returned if a Delegation Output's Validator address is zeroed out.
	ErrDelegationValidatorAddressZeroed = ierrors.New("delegation output's validator address is zeroed")
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

// TransactionSelector implements SerializableSelectorFunc for transaction types.
func TransactionSelector(txType uint32) (*Transaction, error) {
	var seri *Transaction
	switch byte(txType) {
	case TransactionNormal:
		seri = &Transaction{}
	default:
		return nil, ierrors.Wrapf(ErrUnknownTransactionEssenceType, "type byte %d", txType)
	}

	return seri, nil
}

// InputsCommitment is a commitment to the inputs of a transaction.
type InputsCommitment = [InputsCommitmentLength]byte

type (
	txEssenceContextInput  interface{ Input }
	txEssenceInput         interface{ Input }
	TxEssenceOutput        interface{ Output }
	TxEssencePayload       interface{ Payload }
	TxEssenceContextInputs = ContextInputs[txEssenceContextInput]
	TxEssenceInputs        = Inputs[txEssenceInput]
	TxEssenceOutputs       = Outputs[TxEssenceOutput]
	TxEssenceAllotments    = Allotments
)

// Transaction is the essence part of a Transaction.
type Transaction struct {
	// TransactionEssence is the essence of a Transaction.
	*TransactionEssence `serix:"0"`
	// The outputs of this transaction.
	Outputs TxEssenceOutputs `serix:"1,mapKey=outputs"`
}

// TransactionEssence is the essence part of a Transaction.
type TransactionEssence struct {
	// The network ID for which this essence is valid for.
	NetworkID NetworkID `serix:"0,mapKey=networkId"`
	// The slot index in which the transaction was created by the client.
	CreationSlot SlotIndex `serix:"1,mapKey=creationSlot"`
	// The commitment references of this transaction.
	ContextInputs TxEssenceContextInputs `serix:"2,mapKey=contextInputs"`
	// The inputs of this transaction.
	Inputs TxEssenceInputs `serix:"3,mapKey=inputs"`
	// The commitment to the referenced inputs.
	InputsCommitment InputsCommitment `serix:"4,mapKey=inputsCommitment"`
	// The optional accounts map with corresponding allotment values.
	Allotments TxEssenceAllotments `serix:"5,mapKey=allotments"`
	// The optional embedded payload.
	Payload TxEssencePayload `serix:"6,optional,mapKey=payload"`
}

// Clone creates a copy of the Transaction.
func (u *Transaction) Clone() *Transaction {
	var payload TxEssencePayload
	if u.Payload != nil {
		payload = u.Payload.Clone()
	}

	return &Transaction{
		TransactionEssence: &TransactionEssence{
			NetworkID:        u.NetworkID,
			CreationSlot:     u.CreationSlot,
			ContextInputs:    u.ContextInputs.Clone(),
			Inputs:           u.Inputs.Clone(),
			InputsCommitment: u.InputsCommitment,
			Allotments:       u.Allotments.Clone(),
			Payload:          payload,
		},
		Outputs: u.Outputs.Clone(),
	}
}

// SigningMessage returns the to be signed message.
func (u *Transaction) SigningMessage(api API) ([]byte, error) {
	essenceBytes, err := api.Encode(u)
	if err != nil {
		return nil, err
	}
	essenceBytesHash := blake2b.Sum256(essenceBytes)

	return essenceBytesHash[:], nil
}

// Sign produces signatures signing the Transaction for every given AddressKeys.
// The produced signatures are in the same order as the AddressKeys.
func (u *Transaction) Sign(api API, inputsCommitment []byte, addrKeys ...AddressKeys) ([]Signature, error) {
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

func (u *Transaction) Size() int {
	payloadSize := serializer.PayloadLengthByteSize
	if u.Payload != nil {
		payloadSize = u.Payload.Size()
	}

	// TransactionType
	return serializer.OneByte +
		// NetworkID
		serializer.UInt64ByteSize +
		// CreationSlot
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
func (u *Transaction) syntacticallyValidate(protoParams ProtocolParameters) error {
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

	return SyntacticallyValidateOutputs(u.Outputs,
		OutputsSyntacticalDepositAmount(protoParams),
		OutputsSyntacticalExpirationAndTimelock(),
		OutputsSyntacticalNativeTokens(),
		OutputsSyntacticalChainConstrainedOutputUniqueness(),
		OutputsSyntacticalFoundry(),
		OutputsSyntacticalAccount(),
		OutputsSyntacticalNFT(),
		OutputsSyntacticalDelegation(),
	)
}

func (u *Transaction) WorkScore(workScoreStructure *WorkScoreStructure) (WorkScore, error) {
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

	var workScorePayload WorkScore
	if u.Payload != nil {
		workScorePayload, err = u.Payload.WorkScore(workScoreStructure)
		if err != nil {
			return 0, err
		}
	}

	return workScoreContextInputs.Add(workScoreInputs, workScoreOutputs, workScoreAllotments, workScorePayload)
}
