package builder

import (
	"github.com/iotaledger/hive.go/core/safemath"
	"github.com/iotaledger/hive.go/ierrors"
	iotago "github.com/iotaledger/iota.go/v4"
)

// ErrTransactionBuilder defines a generic error occurring within the TransactionBuilder.
var ErrTransactionBuilder = ierrors.New("transaction builder error")

// NewTransactionBuilder creates a new TransactionBuilder.
func NewTransactionBuilder(api iotago.API) *TransactionBuilder {
	return &TransactionBuilder{
		api: api,
		transaction: &iotago.Transaction{
			TransactionEssence: &iotago.TransactionEssence{
				NetworkID:     api.ProtocolParameters().NetworkID(),
				ContextInputs: iotago.TxEssenceContextInputs{},
				Inputs:        iotago.TxEssenceInputs{},
				Allotments:    iotago.Allotments{},
			},
			Outputs: iotago.TxEssenceOutputs{},
		},
		inputOwner: map[iotago.OutputID]iotago.Address{},
		inputs:     iotago.OutputSet{},
	}
}

// TransactionBuilder is used to easily build up a SignedTransaction.
type TransactionBuilder struct {
	api              iotago.API
	occurredBuildErr error
	transaction      *iotago.Transaction
	inputs           iotago.OutputSet
	inputOwner       map[iotago.OutputID]iotago.Address
}

// TxInput defines an input with the address to unlock.
type TxInput struct {
	// The address which needs to be unlocked to spend this input.
	UnlockTarget iotago.Address `json:"address"`
	// The ID of the referenced input.
	InputID iotago.OutputID `json:"inputId"`
	// The output which is used as an input.
	Input iotago.Output `json:"input"`
}

func (b *TransactionBuilder) Clone() *TransactionBuilder {
	cpyInputOwner := make(map[iotago.OutputID]iotago.Address, len(b.inputOwner))
	for outputID, address := range b.inputOwner {
		cpyInputOwner[outputID] = address.Clone()
	}

	return &TransactionBuilder{
		api:              b.api,
		occurredBuildErr: b.occurredBuildErr,
		transaction:      b.transaction.Clone(),
		inputs:           b.inputs.Clone(),
		inputOwner:       cpyInputOwner,
	}
}

// AddInput adds the given input to the builder.
func (b *TransactionBuilder) AddInput(input *TxInput) *TransactionBuilder {
	b.inputOwner[input.InputID] = input.UnlockTarget
	b.transaction.Inputs = append(b.transaction.Inputs, input.InputID.UTXOInput())
	b.inputs[input.InputID] = input.Input

	return b
}

// TransactionBuilderInputFilter is a filter function which determines whether
// an input should be used or not. (returning true = pass). The filter can also
// be used to accumulate data over the set of inputs, i.e. the input sum etc.
type TransactionBuilderInputFilter func(outputID iotago.OutputID, input iotago.Output) bool

// AddContextInput adds the given context input to the builder.
func (b *TransactionBuilder) AddContextInput(contextInput iotago.Input) *TransactionBuilder {
	b.transaction.ContextInputs = append(b.transaction.ContextInputs, contextInput)

	return b
}

// IncreaseAllotment adds or increases the given allotment to the builder.
func (b *TransactionBuilder) IncreaseAllotment(accountID iotago.AccountID, value iotago.Mana) *TransactionBuilder {
	// check if the allotment already exists and add the value on top
	for _, allotment := range b.transaction.Allotments {
		if allotment.AccountID.ToAddress().Equal(accountID.ToAddress()) {
			allotment.Value += value
			return b
		}
	}

	// allotment does not exist yet
	b.transaction.Allotments = append(b.transaction.Allotments, &iotago.Allotment{
		AccountID: accountID,
		Value:     value,
	})

	return b
}

// AddOutput adds the given output to the builder.
func (b *TransactionBuilder) AddOutput(output iotago.Output) *TransactionBuilder {
	b.transaction.Outputs = append(b.transaction.Outputs, output)

	return b
}

func (b *TransactionBuilder) SetCreationSlot(creationSlot iotago.SlotIndex) *TransactionBuilder {
	b.transaction.CreationSlot = creationSlot

	return b
}

// AddTaggedDataPayload adds the given TaggedData as the inner payload.
func (b *TransactionBuilder) AddTaggedDataPayload(payload *iotago.TaggedData) *TransactionBuilder {
	b.transaction.Payload = payload

	return b
}

// TransactionFunc is a function which receives a SignedTransaction as its parameter.
type TransactionFunc func(tx *iotago.SignedTransaction)

func (b *TransactionBuilder) AllotRequiredManaAndStoreRemainingManaInOutput(slotIndexTarget iotago.SlotIndex, rmc iotago.Mana, blockIssuerAccountID iotago.AccountID, storedManaOutputIndex int) *TransactionBuilder {
	setBuildError := func(err error) *TransactionBuilder {
		b.occurredBuildErr = err
		return b
	}

	// calculate the available mana on input side
	totalManaIn, err := b.TotalManaInputs(b.api.ProtocolParameters(), b.inputs, slotIndexTarget)
	if err != nil {
		return setBuildError(ierrors.Wrap(err, "failed to calculate the available mana on input side"))
	}

	// substract the already alloted mana
	for _, allotment := range b.transaction.Allotments {
		totalManaIn, err = safemath.SafeSub(totalManaIn, allotment.Value)
		if err != nil {
			return setBuildError(ierrors.Wrap(err, "failed to substract alloted mana"))
		}
	}

	// substract the stored mana on the outputs side
	for _, output := range b.transaction.Outputs {
		totalManaIn, err = safemath.SafeSub(totalManaIn, output.StoredMana())
		if err != nil {
			return setBuildError(ierrors.Wrap(err, "failed to substract the stored mana on the outputs side"))
		}
	}

	// calculate the minimum required mana to issue the block
	minRequiredMana, err := b.MinRequiredAllotedMana(b.api.ProtocolParameters().WorkScoreStructure(), rmc, blockIssuerAccountID)
	if err != nil {
		return setBuildError(ierrors.Wrap(err, "failed to substract the stored mana on the outputs side"))
	}

	// substract the minimum required mana to issue the block
	totalManaIn, err = safemath.SafeSub(totalManaIn, minRequiredMana)
	if err != nil {
		return setBuildError(ierrors.Wrap(err, "failed to substract the minimum required mana to issue the block"))
	}

	// allot the mana to the block issuer account
	b.IncreaseAllotment(blockIssuerAccountID, minRequiredMana)

	if storedManaOutputIndex >= len(b.transaction.Outputs) {
		return setBuildError(ierrors.Wrapf(err, "given storedManaOutputIndex does not exist: %d", storedManaOutputIndex))
	}

	// move the remaining mana to stored mana on the specified output index
	switch output := b.transaction.Outputs[storedManaOutputIndex].(type) {
	case *iotago.BasicOutput:
		output.Mana += totalManaIn
	case *iotago.AccountOutput:
		output.Mana += totalManaIn
	case *iotago.NFTOutput:
		output.Mana += totalManaIn
	default:
		return setBuildError(ierrors.Wrapf(iotago.ErrUnknownOutputType, "output type %T does not support stored mana", output))
	}

	return b
}

// BuildAndSwapToBlockBuilder builds the transaction and then swaps to a BasicBlockBuilder with
// the transaction set as its payload. txFunc can be nil.
func (b *TransactionBuilder) BuildAndSwapToBlockBuilder(signer iotago.AddressSigner, txFunc TransactionFunc) *BasicBlockBuilder {
	blockBuilder := NewBasicBlockBuilder(b.api)
	tx, err := b.Build(signer)
	if err != nil {
		blockBuilder.err = err

		return blockBuilder
	}
	if txFunc != nil {
		txFunc(tx)
	}

	return blockBuilder.Payload(tx)
}

func (b *TransactionBuilder) TotalManaInputs(protoParams iotago.ProtocolParameters, inputSet iotago.OutputSet, slotIndexTarget iotago.SlotIndex) (iotago.Mana, error) {
	var totalMana iotago.Mana

	for inputID, input := range inputSet {
		// calculate the potential mana of the input
		// we need to ignore the storage deposit, because it doesn't generate mana
		rentStructure := iotago.NewRentStructure(protoParams.RentParameters())
		excessBaseTokens := input.BaseTokenAmount() - rentStructure.MinDeposit(input)
		potentialMana, err := protoParams.ManaDecayProvider().ManaGenerationWithDecay(excessBaseTokens, inputID.CreationSlot(), slotIndexTarget)
		if err != nil {
			// todo add error message
			return 0, err
		}

		// calculate the decayed stored mana of the input
		storedMana, err := protoParams.ManaDecayProvider().ManaWithDecay(input.StoredMana(), inputID.CreationSlot(), slotIndexTarget)
		if err != nil {
			// todo add error message
			return 0, err
		}

		totalMana += potentialMana + storedMana
	}

	return totalMana, nil
}

// MinRequiredAllotedMana returns the minimum alloted mana required to issue a ProtocolBlock
// with 4 strong parents, the transaction payload from the builder and 1 allotment for the block issuer.
func (b *TransactionBuilder) MinRequiredAllotedMana(workScoreStructure *iotago.WorkScoreStructure, rmc iotago.Mana, blockIssuerAccountID iotago.AccountID) (iotago.Mana, error) {
	// clone the essence allotments to not modify the original transaction
	allotmentsCpy := b.transaction.Allotments.Clone()

	// undo the changes to the allotments at the end
	defer func() {
		b.transaction.Allotments = allotmentsCpy
	}()

	// add an empty allotment to account for the later added allotment for the block issuer in case it does not exist yet
	b.IncreaseAllotment(blockIssuerAccountID, 0)

	// create a signed transaction with a empty signer to get the correct workscore.
	// later the transaction needs to be signed with the correct signer, after the alloted mana was set correctly.
	dummyTxPayload, err := b.Build(&iotago.EmptyAddressSigner{})
	if err != nil {
		return 0, ierrors.Wrap(err, "failed to build the transaction payload")
	}

	payloadWorkScore, err := dummyTxPayload.WorkScore(workScoreStructure)
	if err != nil {
		return 0, ierrors.Wrap(err, "failed to calculate the transaction payload workscore")
	}

	workScore, err := workScoreStructure.Block.Add(payloadWorkScore)
	if err != nil {
		return 0, ierrors.Wrap(err, "failed to add the block workscore")
	}

	manaCost, err := iotago.ManaCost(rmc, workScore)
	if err != nil {
		return 0, ierrors.Wrap(err, "failed to calculate the mana cost")
	}

	return manaCost, nil
}

// Build sings the inputs with the given signer and returns the built payload.
func (b *TransactionBuilder) Build(signer iotago.AddressSigner) (*iotago.SignedTransaction, error) {
	switch {
	case b.occurredBuildErr != nil:
		return nil, b.occurredBuildErr
	case signer == nil:
		return nil, ierrors.Wrap(ErrTransactionBuilder, "must supply signer")
	}

	// prepare the inputs commitment in the same order as the inputs in the essence
	var inputIDs iotago.OutputIDs
	for _, input := range b.transaction.Inputs {
		//nolint:forcetypeassert // we can safely assume that this is an UTXOInput
		inputIDs = append(inputIDs, input.(*iotago.UTXOInput).OutputID())
	}

	inputs := inputIDs.OrderedSet(b.inputs)
	commitment, err := inputs.Commitment(b.api)
	if err != nil {
		return nil, ierrors.Wrapf(err, "failed to calculate TX inputs commitment: %s, %s", inputIDs, b.inputs)
	}
	copy(b.transaction.InputsCommitment[:], commitment)

	txEssenceData, err := b.transaction.SigningMessage(b.api)
	if err != nil {
		return nil, ierrors.Wrap(err, "failed to calculate tx transaction for signing message")
	}

	unlockPos := map[string]int{}
	unlocks := iotago.Unlocks{}
	for i, inputRef := range b.transaction.Inputs {
		//nolint:forcetypeassert // we can safely assume that this is an UTXOInput
		addr := b.inputOwner[inputRef.(*iotago.UTXOInput).OutputID()]
		addrKey := addr.Key()

		pos, unlocked := unlockPos[addrKey]
		if !unlocked {
			// the output's owning chain address must have been unlocked already
			if _, is := addr.(iotago.ChainAddress); is {
				return nil, ierrors.Errorf("input %d's owning chain is not unlocked, chainID %s, type %s", i, addr, addr.Type())
			}

			// produce signature
			var signature iotago.Signature
			signature, err = signer.Sign(addr, txEssenceData)
			if err != nil {
				return nil, ierrors.Wrapf(err, "failed to sign tx transaction: %s", txEssenceData)
			}

			unlocks = append(unlocks, &iotago.SignatureUnlock{Signature: signature})
			addChainAsUnlocked(inputs[i], i, unlockPos)
			unlockPos[addrKey] = i

			continue
		}

		unlocks = addReferentialUnlock(addr, unlocks, pos)
		addChainAsUnlocked(inputs[i], i, unlockPos)
	}

	sigTxPayload := &iotago.SignedTransaction{
		API:         b.api,
		Transaction: b.transaction,
		Unlocks:     unlocks,
	}

	return sigTxPayload, nil
}

func addReferentialUnlock(addr iotago.Address, unlocks iotago.Unlocks, pos int) iotago.Unlocks {
	switch addr.(type) {
	case *iotago.AccountAddress:
		return append(unlocks, &iotago.AccountUnlock{Reference: uint16(pos)})
	case *iotago.NFTAddress:
		return append(unlocks, &iotago.NFTUnlock{Reference: uint16(pos)})
	default:
		return append(unlocks, &iotago.ReferenceUnlock{Reference: uint16(pos)})
	}
}

func addChainAsUnlocked(input iotago.Output, posUnlocked int, prevUnlocked map[string]int) {
	if chainInput, is := input.(iotago.ChainOutput); is && chainInput.ChainID().Addressable() {
		prevUnlocked[chainInput.ChainID().ToAddress().Key()] = posUnlocked
	}
}
