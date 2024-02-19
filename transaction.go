package iotago

import (
	"context"
	"crypto"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2/byteutils"
	"github.com/iotaledger/hive.go/stringify"
	"github.com/iotaledger/iota.go/v4/merklehasher"
)

const (
	// MaxOutputsCount defines the maximum amount of outputs within a Transaction.
	MaxOutputsCount = 128
	// MinOutputsCount defines the minimum amount of inputs within a Transaction.
	MinOutputsCount = 1
)

var (
	// ErrTxTypeInvalid gets returned for invalid transaction type.
	ErrTxTypeInvalid = ierrors.New("transaction type is invalid")
	// ErrTxEssenceNetworkIDInvalid gets returned when a network ID within a Transaction is invalid.
	ErrTxEssenceNetworkIDInvalid = ierrors.New("invalid network ID")
	// ErrTxEssenceCapabilitiesInvalid gets returned when the capabilities within a Transaction are invalid.
	ErrTxEssenceCapabilitiesInvalid = ierrors.New("invalid capabilities")
	// ErrInputUTXORefsNotUnique gets returned if multiple inputs reference the same UTXO.
	ErrInputUTXORefsNotUnique = ierrors.New("inputs must each reference a unique UTXO")
	// ErrInputRewardIndexExceedsMaxInputsCount gets returned if a reward input references an index greater than max inputs count.
	ErrInputRewardIndexExceedsMaxInputsCount = ierrors.New("reward input references an index greater than max inputs count")
	// ErrAccountOutputNonEmptyState gets returned if an AccountOutput with zeroed AccountID contains state (counters non-zero etc.).
	ErrAccountOutputNonEmptyState = ierrors.New("account output is not empty state")
	// ErrAccountOutputCyclicAddress gets returned if an AccountOutput's AccountID results into the same address as the address field within the output.
	ErrAccountOutputCyclicAddress = ierrors.New("account output's ID corresponds to address field")
	// ErrAccountOutputAmountLessThanStakedAmount gets returned if an AccountOutput's amount is less than the StakedAmount field of its staking feature.
	ErrAccountOutputAmountLessThanStakedAmount = ierrors.New("account output's amount is less than the staked amount")
	// ErrAnchorOutputNonEmptyState gets returned if an AnchorOutput with zeroed AnchorID contains state (counters non-zero etc.).
	ErrAnchorOutputNonEmptyState = ierrors.New("anchor output is not empty state")
	// ErrAnchorOutputCyclicAddress gets returned if an AnchorOutput's AnchorID results into the same address as the State/Governance controller.
	ErrAnchorOutputCyclicAddress = ierrors.New("anchor output's AnchorID corresponds to state and/or governance controller")
	// ErrNFTOutputCyclicAddress gets returned if an NFTOutput's NFTID results into the same address as the address field within the output.
	ErrNFTOutputCyclicAddress = ierrors.New("NFT output's ID corresponds to address field")
	// ErrOutputsSumExceedsTotalSupply gets returned if the sum of the output deposits exceeds the total supply of tokens.
	ErrOutputsSumExceedsTotalSupply = ierrors.New("accumulated output balance exceeds total supply")
	// ErrStorageDepositLessThanMinReturnOutputStorageDeposit gets returned when the storage deposit condition's amount is less than the min storage deposit for the return output.
	ErrStorageDepositLessThanMinReturnOutputStorageDeposit = ierrors.New("storage deposit return amount is less than the min storage deposit needed for the return output")
	// ErrStorageDepositExceedsTargetOutputAmount gets returned when the storage deposit condition's amount exceeds the target output's base token amount.
	ErrStorageDepositExceedsTargetOutputAmount = ierrors.New("storage deposit return amount exceeds target output's base token amount")
	// ErrMaxManaExceeded gets returned when the sum of stored mana in all outputs or the sum of Mana in all allotments exceeds the maximum Mana value.
	ErrMaxManaExceeded = ierrors.New("max mana value exceeded")
	// ErrManaDecayCreationIndexExceedsTargetIndex gets returned when the creation slot/epoch index
	// exceeds the target slot/epoch index in mana decay.
	ErrManaDecayCreationIndexExceedsTargetIndex = ierrors.New("mana decay creation slot/epoch index exceeds target slot/epoch index")
)

type (
	TxEssenceOutputs = Outputs[TxEssenceOutput]
)

// Transaction is the part of a SignedTransaction that contains inputs and outputs.
type Transaction struct {
	API                 API
	*TransactionEssence `serix:",inlined"`
	// The outputs of this transaction.
	Outputs TxEssenceOutputs `serix:""`
}

// ID returns the TransactionID created without the signatures.
func (t *Transaction) ID() (TransactionID, error) {
	transactionCommitment, err := t.TransactionCommitment()
	if err != nil {
		return EmptyTransactionID, ierrors.Errorf("can't compute transaction commitment: %w", err)
	}

	outputCommitment, err := t.OutputCommitment()
	if err != nil {
		return TransactionID{}, ierrors.Errorf("can't compute output commitment: %w", err)
	}

	return TransactionIDFromTransactionCommitmentAndOutputCommitment(t.CreationSlot, transactionCommitment, outputCommitment), nil
}

func TransactionIDFromTransactionCommitmentAndOutputCommitment(slot SlotIndex, transactionCommitment Identifier, outputCommitment Identifier) TransactionID {
	return TransactionIDRepresentingData(slot, byteutils.ConcatBytes(transactionCommitment[:], outputCommitment[:]))
}

// TransactionCommitment returns the transaction commitment hashing the transaction essence.
func (t *Transaction) TransactionCommitment() (Identifier, error) {
	essenceBytes, err := t.API.Encode(t.TransactionEssence)
	if err != nil {
		return EmptyIdentifier, ierrors.Errorf("can't compute essence bytes: %w", err)
	}

	return IdentifierFromData(essenceBytes), nil
}

// OutputCommitment returns the output commitment which is the root of the merkle tree of the outputs.
func (t *Transaction) OutputCommitment() (Identifier, error) {
	//nolint:nosnakecase // false positive
	outputHasher := merklehasher.NewHasher[*APIByter[TxEssenceOutput]](crypto.BLAKE2b_256)
	wrappedOutputs := lo.Map(t.Outputs, APIByterFactory[TxEssenceOutput](t.API))

	root, err := outputHasher.HashValues(wrappedOutputs)
	if err != nil {
		return EmptyIdentifier, err
	}

	return Identifier(root), nil
}

func (t *Transaction) SetDeserializationContext(ctx context.Context) {
	t.API = APIFromContext(ctx)
}

func (t *Transaction) Clone() *Transaction {
	return &Transaction{
		API:                t.API,
		TransactionEssence: t.TransactionEssence.Clone(),
		Outputs:            t.Outputs.Clone(),
	}
}

func (t *Transaction) Inputs() ([]*UTXOInput, error) {
	references := make([]*UTXOInput, 0, len(t.TransactionEssence.Inputs))
	for _, input := range t.TransactionEssence.Inputs {
		switch castInput := input.(type) {
		case *UTXOInput:
			references = append(references, castInput)
		default:
			return nil, ErrUnknownInputType
		}
	}

	return references, nil
}

// OutputsSet returns an OutputSet from the Transaction's outputs, mapped by their OutputID.
func (t *Transaction) OutputsSet() (OutputSet, error) {
	txID, err := t.ID()
	if err != nil {
		return nil, err
	}
	set := make(OutputSet)
	for index, output := range t.Outputs {
		set[OutputIDFromTransactionIDAndIndex(txID, uint16(index))] = output
	}

	return set, nil
}

func (t *Transaction) ContextInputs() (TransactionContextInputs, error) {
	references := make(TransactionContextInputs, 0, len(t.TransactionEssence.ContextInputs))
	for _, input := range t.TransactionEssence.ContextInputs {
		switch castInput := input.(type) {
		case *CommitmentInput, *BlockIssuanceCreditInput, *RewardInput:
			references = append(references, castInput)
		default:
			return nil, ErrUnknownContextInputType
		}
	}

	return references, nil
}

func (t *Transaction) BICInputs() ([]*BlockIssuanceCreditInput, error) {
	references := make([]*BlockIssuanceCreditInput, 0, len(t.TransactionEssence.ContextInputs))
	for _, input := range t.TransactionEssence.ContextInputs {
		switch castInput := input.(type) {
		case *BlockIssuanceCreditInput:
			references = append(references, castInput)
		case *CommitmentInput, *RewardInput:
			// ignore this type
		default:
			return nil, ErrUnknownContextInputType
		}
	}

	return references, nil
}

func (t *Transaction) RewardInputs() ([]*RewardInput, error) {
	references := make([]*RewardInput, 0, len(t.TransactionEssence.ContextInputs))
	for _, input := range t.TransactionEssence.ContextInputs {
		switch castInput := input.(type) {
		case *RewardInput:
			references = append(references, castInput)
		case *CommitmentInput, *BlockIssuanceCreditInput:
			// ignore this type
		default:
			return nil, ErrUnknownContextInputType
		}
	}

	return references, nil
}

// Returns the first commitment input in the transaction if it exists or nil.
func (t *Transaction) CommitmentInput() *CommitmentInput {
	for _, input := range t.TransactionEssence.ContextInputs {
		switch castInput := input.(type) {
		case *BlockIssuanceCreditInput, *RewardInput:
			// ignore this type
		case *CommitmentInput:
			return castInput
		default:
			return nil
		}
	}

	return nil
}

// SigningMessage returns the to be signed message.
func (t *Transaction) SigningMessage() ([]byte, error) {
	essenceBytes, err := t.API.Encode(t)
	if err != nil {
		return nil, err
	}
	essenceBytesHash := blake2b.Sum256(essenceBytes)

	return essenceBytesHash[:], nil
}

// Sign produces signatures signing the essence for every given AddressKeys.
// The produced signatures are in the same order as the AddressKeys.
func (t *Transaction) Sign(addrKeys ...AddressKeys) ([]Signature, error) {
	signMsg, err := t.SigningMessage()
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

func (t *Transaction) Size() int {
	return t.TransactionEssence.Size() + t.Outputs.Size()
}

// syntacticallyValidate checks whether the transaction essence is syntactically valid.
// The function does not syntactically validate the input or outputs themselves.
func (t *Transaction) SyntacticallyValidate(api API) error {
	protoParams := api.ProtocolParameters()

	if err := t.TransactionEssence.syntacticallyValidateEssence(api); err != nil {
		return err
	}

	var maxManaValue Mana = (1 << protoParams.ManaParameters().BitsCount) - 1

	return SyntacticallyValidateOutputs(t.Outputs,
		OutputsSyntacticalUnlockConditionLexicalOrderAndUniqueness(),
		OutputsSyntacticalFeaturesLexicalOrderAndUniqueness(),
		OutputsSyntacticalMetadataFeatureMaxSize(),
		OutputsSyntacticalDepositAmount(protoParams, api.StorageScoreStructure()),
		OutputsSyntacticalExpirationAndTimelock(),
		OutputsSyntacticalNativeTokens(),
		OutputsSyntacticalStoredMana(maxManaValue),
		OutputsSyntacticalChainConstrainedOutputUniqueness(),
		OutputsSyntacticalFoundry(),
		OutputsSyntacticalAccount(),
		OutputsSyntacticalAnchor(),
		OutputsSyntacticalNFT(),
		OutputsSyntacticalDelegation(),
		OutputsSyntacticalAddressRestrictions(),
		OutputsSyntacticalImplicitAccountCreationAddress(),
	)
}

// WorkScore calculates the Work Score of the Transaction.
func (t *Transaction) WorkScore(workScoreParameters *WorkScoreParameters) (WorkScore, error) {
	workscoreTransactionEssence, err := t.TransactionEssence.WorkScore(workScoreParameters)
	if err != nil {
		return 0, err
	}

	workScoreOutputs, err := t.Outputs.WorkScore(workScoreParameters)
	if err != nil {
		return 0, err
	}

	return workscoreTransactionEssence.Add(workScoreOutputs)
}

// String returns a human readable version of the Transaction.
func (t *Transaction) String() string {
	return stringify.Struct("Transaction",
		stringify.NewStructField("TransactionEssence", t.TransactionEssence),
		stringify.NewStructField("Outputs", t.Outputs),
	)
}
