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
			API: api,
			TransactionEssence: &iotago.TransactionEssence{
				NetworkID:     api.ProtocolParameters().NetworkID(),
				ContextInputs: iotago.TxEssenceContextInputs{},
				Inputs:        iotago.TxEssenceInputs{},
				Allotments:    iotago.Allotments{},
				Capabilities:  iotago.TransactionCapabilitiesBitMask{},
			},
			Outputs: iotago.TxEssenceOutputs{},
		},
		inputOwner: map[iotago.OutputID]iotago.Address{},
		inputs:     iotago.OutputSet{},
		rewards:    iotago.Mana(0),
	}
}

// TransactionBuilder is used to easily build up a SignedTransaction.
type TransactionBuilder struct {
	api              iotago.API
	occurredBuildErr error
	transaction      *iotago.Transaction
	inputs           iotago.OutputSet
	inputOwner       map[iotago.OutputID]iotago.Address
	rewards          iotago.Mana
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
	b.transaction.TransactionEssence.Inputs = append(b.transaction.TransactionEssence.Inputs, input.InputID.UTXOInput())
	b.inputs[input.InputID] = input.Input

	return b
}

// TransactionBuilderInputFilter is a filter function which determines whether
// an input should be used or not. (returning true = pass). The filter can also
// be used to accumulate data over the set of inputs, i.e. the input sum etc.
type TransactionBuilderInputFilter func(outputID iotago.OutputID, input iotago.Output) bool

// AddCommitmentInput adds the given commitment input to the builder.
func (b *TransactionBuilder) AddCommitmentInput(commitmentInput *iotago.CommitmentInput) *TransactionBuilder {
	b.transaction.TransactionEssence.ContextInputs = append(b.transaction.TransactionEssence.ContextInputs, commitmentInput)

	return b
}

// AddBlockIssuanceCreditInput adds the given block issuance credit input to the builder.
func (b *TransactionBuilder) AddBlockIssuanceCreditInput(blockIssuanceCreditInput *iotago.BlockIssuanceCreditInput) *TransactionBuilder {
	b.transaction.TransactionEssence.ContextInputs = append(b.transaction.TransactionEssence.ContextInputs, blockIssuanceCreditInput)

	return b
}

// AddRewardInput adds the given reward input to the builder.
func (b *TransactionBuilder) AddRewardInput(rewardInput *iotago.RewardInput, mana iotago.Mana) *TransactionBuilder {
	b.transaction.TransactionEssence.ContextInputs = append(b.transaction.TransactionEssence.ContextInputs, rewardInput)
	b.rewards += mana

	return b
}

// IncreaseAllotment adds or increases the given allotment to the builder.
func (b *TransactionBuilder) IncreaseAllotment(accountID iotago.AccountID, value iotago.Mana) *TransactionBuilder {
	if value == 0 {
		return b
	}

	// check if the allotment already exists and add the value on top
	for _, allotment := range b.transaction.Allotments {
		if allotment.AccountID == accountID {
			allotment.Mana += value
			return b
		}
	}

	// allotment does not exist yet
	b.transaction.Allotments = append(b.transaction.Allotments, &iotago.Allotment{
		AccountID: accountID,
		Mana:      value,
	})

	return b
}

// AddOutput adds the given output to the builder.
func (b *TransactionBuilder) AddOutput(output iotago.Output) *TransactionBuilder {
	b.transaction.Outputs = append(b.transaction.Outputs, output)

	return b
}

// WithTransactionCapabilities sets the capabilities of the transaction.
func (b *TransactionBuilder) WithTransactionCapabilities(capabilities iotago.TransactionCapabilitiesBitMask) *TransactionBuilder {
	b.transaction.Capabilities = capabilities
	return b
}

func (b *TransactionBuilder) CreationSlot() iotago.SlotIndex {
	return b.transaction.CreationSlot
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

func (b *TransactionBuilder) StoreRemainingManaInOutput(targetSlot iotago.SlotIndex, blockIssuerAccountID iotago.AccountID, storedManaOutputIndex int) *TransactionBuilder {
	setBuildError := func(err error) *TransactionBuilder {
		b.occurredBuildErr = err
		return b
	}

	if storedManaOutputIndex >= len(b.transaction.Outputs) {
		return setBuildError(ierrors.Errorf("given storedManaOutputIndex does not exist: %d", storedManaOutputIndex))
	}

	unboundManaInputsLeftoverBalance, err := b.calculateAvailableManaLeftover(targetSlot, 0, blockIssuerAccountID)
	if err != nil {
		return setBuildError(err)
	}

	// move the remaining mana to stored mana on the specified output index
	switch output := b.transaction.Outputs[storedManaOutputIndex].(type) {
	case *iotago.BasicOutput:
		output.Mana += unboundManaInputsLeftoverBalance
	case *iotago.AccountOutput:
		output.Mana += unboundManaInputsLeftoverBalance
	case *iotago.NFTOutput:
		output.Mana += unboundManaInputsLeftoverBalance
	default:
		return setBuildError(ierrors.Wrapf(iotago.ErrUnknownOutputType, "output type %T does not support stored mana", output))
	}

	return b
}

func (b *TransactionBuilder) AllotRequiredManaAndStoreRemainingManaInOutput(targetSlot iotago.SlotIndex, rmc iotago.Mana, blockIssuerAccountID iotago.AccountID, storedManaOutputIndex int) *TransactionBuilder {
	setBuildError := func(err error) *TransactionBuilder {
		b.occurredBuildErr = err
		return b
	}

	if storedManaOutputIndex >= len(b.transaction.Outputs) {
		return setBuildError(ierrors.Errorf("given storedManaOutputIndex does not exist: %d", storedManaOutputIndex))
	}

	// calculate the minimum required mana to issue the block
	minRequiredMana, err := b.MinRequiredAllotedMana(b.api.ProtocolParameters().WorkScoreParameters(), rmc, blockIssuerAccountID)
	if err != nil {
		return setBuildError(ierrors.Wrap(err, "failed to calculate the minimum required mana to issue the block"))
	}

	unboundManaInputsLeftoverBalance, err := b.calculateAvailableManaLeftover(targetSlot, minRequiredMana, blockIssuerAccountID)
	if err != nil {
		return setBuildError(err)
	}

	// allot the mana to the block issuer account (we increase the value, so we don't interfere with the already alloted value)
	b.IncreaseAllotment(blockIssuerAccountID, minRequiredMana)

	// move the remaining mana to stored mana on the specified output index
	switch output := b.transaction.Outputs[storedManaOutputIndex].(type) {
	case *iotago.BasicOutput:
		output.Mana += unboundManaInputsLeftoverBalance
	case *iotago.AccountOutput:
		output.Mana += unboundManaInputsLeftoverBalance
	case *iotago.NFTOutput:
		output.Mana += unboundManaInputsLeftoverBalance
	default:
		return setBuildError(ierrors.Wrapf(iotago.ErrUnknownOutputType, "output type %T does not support stored mana", output))
	}

	return b
}

// AllotAllMana allots all available mana to the provided account, even if the alloted value is less than the minimum required mana value to issue the block.
func (b *TransactionBuilder) AllotAllMana(targetSlot iotago.SlotIndex, blockIssuerAccountID iotago.AccountID) *TransactionBuilder {
	setBuildError := func(err error) *TransactionBuilder {
		b.occurredBuildErr = err
		return b
	}

	unboundManaInputsLeftoverBalance, err := b.calculateAvailableManaLeftover(targetSlot, 0, blockIssuerAccountID)
	if err != nil {
		return setBuildError(err)
	}

	// allot the mana to the block issuer account (we increase the value, so we don't interfere with the already alloted value)
	b.IncreaseAllotment(blockIssuerAccountID, unboundManaInputsLeftoverBalance)

	return b
}

func (b *TransactionBuilder) calculateAvailableManaLeftover(targetSlot iotago.SlotIndex, minRequiredMana iotago.Mana, blockIssuerAccountID iotago.AccountID) (iotago.Mana, error) {
	// calculate the available mana on input side
	availableManaInputs, err := b.CalculateAvailableMana(targetSlot)
	if err != nil {
		return 0, ierrors.Wrap(err, "failed to calculate the available mana on input side")
	}

	// update the account bound mana balances if they exist and/or the onbound mana balance
	updateUnboundAndAccountBoundManaBalances := func(accountID iotago.AccountID, accountBoundManaOut iotago.Mana) error {
		// check if there is account bound mana for this account on the input side
		if accountBalance, exists := availableManaInputs.AccountBoundMana[accountID]; exists {
			// check if there is enough account bound mana for this account on the input side
			if accountBalance < accountBoundManaOut {
				// not enough mana for this account on the input side
				// => set the remaining account bound mana for this account to 0
				availableManaInputs.AccountBoundMana[accountID] = 0

				// subtract the remainder from the unbound mana
				availableManaInputs.UnboundMana, err = safemath.SafeSub(availableManaInputs.UnboundMana, accountBoundManaOut-accountBalance)
				if err != nil {
					return ierrors.Wrapf(err, "not enough unbound mana on the input side for account %s while subtracting remainder", accountID.String())
				}

				return nil
			}

			// there is enough account bound mana for this account, subtract it from there
			availableManaInputs.AccountBoundMana[accountID] -= accountBoundManaOut

			return nil
		}

		// no account bound mana available for the given account, subtract it from the unbounded mana
		availableManaInputs.UnboundMana, err = safemath.SafeSub(availableManaInputs.UnboundMana, accountBoundManaOut)
		if err != nil {
			return ierrors.Wrapf(err, "not enough unbound mana on the input side for account %s", accountID.String())
		}

		return nil
	}

	// subtract the stored mana on the outputs side
	for _, o := range b.transaction.Outputs {
		switch output := o.(type) {
		case *iotago.AccountOutput:
			// mana on account outputs is locked to this account
			if err = updateUnboundAndAccountBoundManaBalances(output.AccountID, output.StoredMana()); err != nil {
				return 0, ierrors.Wrap(err, "failed to subtract the stored mana on the outputs side for account output")
			}

		default:
			// check if the output locked mana to a certain account
			if accountID, isManaLocked := b.hasManalockCondition(output); isManaLocked {
				if err = updateUnboundAndAccountBoundManaBalances(accountID, output.StoredMana()); err != nil {
					return 0, ierrors.Wrap(err, "failed to subtract the stored mana on the outputs side, while checking locked mana")
				}
			} else {
				availableManaInputs.UnboundMana, err = safemath.SafeSub(availableManaInputs.UnboundMana, output.StoredMana())
				if err != nil {
					return 0, ierrors.Wrap(err, "failed to subtract the stored mana on the outputs side")
				}
			}
		}
	}

	// subtract the already alloted mana
	for _, allotment := range b.transaction.Allotments {
		if err = updateUnboundAndAccountBoundManaBalances(allotment.AccountID, allotment.Mana); err != nil {
			return 0, ierrors.Wrap(err, "failed to subtract the already alloted mana")
		}
	}

	// subtract the minimum required mana to issue the block
	if err = updateUnboundAndAccountBoundManaBalances(blockIssuerAccountID, minRequiredMana); err != nil {
		return 0, ierrors.Wrap(err, "failed to subtract the minimum required mana to issue the block")
	}

	return availableManaInputs.UnboundMana, nil
}

// hasManalockCondition checks if the output is locked for a certain time to an account.
func (b *TransactionBuilder) hasManalockCondition(output iotago.Output) (iotago.AccountID, bool) {
	minManalockedSlot := b.transaction.CreationSlot + 2*b.api.ProtocolParameters().MaxCommittableAge()

	if !output.UnlockConditionSet().HasTimelockUntil(minManalockedSlot) {
		return iotago.EmptyAccountID, false
	}

	unlockAddress := output.UnlockConditionSet().Address()
	if unlockAddress == nil {
		return iotago.EmptyAccountID, false
	}

	if unlockAddress.Address.Type() != iotago.AddressAccount {
		return iotago.EmptyAccountID, false
	}
	//nolint:forcetypeassert // we can safely assume that this is an AccountAddress
	accountAddress := unlockAddress.Address.(*iotago.AccountAddress)

	return accountAddress.AccountID(), true
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

type AvailableManaResult struct {
	TotalMana            iotago.Mana
	UnboundMana          iotago.Mana
	PotentialMana        iotago.Mana
	StoredMana           iotago.Mana
	UnboundPotentialMana iotago.Mana
	UnboundStoredMana    iotago.Mana
	AccountBoundMana     map[iotago.AccountID]iotago.Mana
	Rewards              iotago.Mana
}

func (a *AvailableManaResult) addTotalMana(value iotago.Mana) error {
	totalMana, err := safemath.SafeAdd(a.TotalMana, value)
	if err != nil {
		return ierrors.Wrap(err, "failed to add total mana")
	}
	a.TotalMana = totalMana

	return nil
}

func (a *AvailableManaResult) addUnboundMana(value iotago.Mana) error {
	unboundMana, err := safemath.SafeAdd(a.UnboundMana, value)
	if err != nil {
		return ierrors.Wrap(err, "failed to add unbound mana")
	}
	a.UnboundMana = unboundMana

	return nil
}

func (a *AvailableManaResult) AddPotentialMana(value iotago.Mana) error {
	potentialMana, err := safemath.SafeAdd(a.PotentialMana, value)
	if err != nil {
		return ierrors.Wrap(err, "failed to add potential mana")
	}
	a.PotentialMana = potentialMana

	return a.addTotalMana(value)
}

func (a *AvailableManaResult) AddStoredMana(value iotago.Mana) error {
	storedMana, err := safemath.SafeAdd(a.StoredMana, value)
	if err != nil {
		return ierrors.Wrap(err, "failed to add stored mana")
	}
	a.StoredMana = storedMana

	return a.addTotalMana(value)
}

func (a *AvailableManaResult) AddUnboundPotentialMana(value iotago.Mana) error {
	unboundPotentialMana, err := safemath.SafeAdd(a.UnboundPotentialMana, value)
	if err != nil {
		return ierrors.Wrap(err, "failed to add unbound potential mana")
	}
	a.UnboundPotentialMana = unboundPotentialMana

	return a.addUnboundMana(value)
}

func (a *AvailableManaResult) AddUnboundStoredMana(value iotago.Mana) error {
	unboundStoredMana, err := safemath.SafeAdd(a.UnboundStoredMana, value)
	if err != nil {
		return ierrors.Wrap(err, "failed to add unbound stored mana")
	}
	a.UnboundStoredMana = unboundStoredMana

	return a.addUnboundMana(value)
}

func (a *AvailableManaResult) AddRewards(value iotago.Mana) error {
	rewards, err := safemath.SafeAdd(a.Rewards, value)
	if err != nil {
		return ierrors.Wrap(err, "failed to add rewards")
	}
	a.Rewards = rewards

	return a.addUnboundMana(value)
}

func (a *AvailableManaResult) AddAccountBoundMana(accountID iotago.AccountID, value iotago.Mana) error {
	accountBoundMana, err := safemath.SafeAdd(a.AccountBoundMana[accountID], value)
	if err != nil {
		return ierrors.Wrapf(err, "failed to add account bound mana to account %s", accountID.ToHex())
	}
	a.AccountBoundMana[accountID] = accountBoundMana

	return nil
}

func (b *TransactionBuilder) CalculateAvailableMana(targetSlot iotago.SlotIndex) (*AvailableManaResult, error) {
	result := &AvailableManaResult{
		AccountBoundMana: make(map[iotago.AccountID]iotago.Mana),
	}

	for inputID, input := range b.inputs {
		// calculate the potential mana of the input
		var inputPotentialMana iotago.Mana

		inputPotentialMana, err := iotago.PotentialMana(b.api.ManaDecayProvider(), b.api.StorageScoreStructure(), input, inputID.CreationSlot(), targetSlot)
		if err != nil {
			return nil, ierrors.Wrap(err, "failed to calculate potential mana")
		}

		if err := result.AddPotentialMana(inputPotentialMana); err != nil {
			return nil, err
		}

		// calculate the decayed stored mana of the input
		inputStoredMana, err := b.api.ManaDecayProvider().DecayManaBySlots(input.StoredMana(), inputID.CreationSlot(), targetSlot)
		if err != nil {
			return nil, ierrors.Wrap(err, "failed to calculate stored mana decay")
		}

		if err := result.AddStoredMana(inputStoredMana); err != nil {
			return nil, err
		}

		if accountOutput, isAccountOutput := input.(*iotago.AccountOutput); isAccountOutput {
			inputTotalMana, err := safemath.SafeAdd(inputPotentialMana, inputStoredMana)
			if err != nil {
				return nil, ierrors.Wrap(err, "failed to add input mana")
			}

			if err := result.AddAccountBoundMana(accountOutput.AccountID, inputTotalMana); err != nil {
				return nil, err
			}
		} else {
			if err := result.AddUnboundPotentialMana(inputPotentialMana); err != nil {
				return nil, err
			}

			if err := result.AddUnboundStoredMana(inputStoredMana); err != nil {
				return nil, err
			}
		}
	}

	// add the rewards (unbound)
	if err := result.AddRewards(b.rewards); err != nil {
		return nil, err
	}

	return result, nil
}

// MinRequiredAllotedMana returns the minimum alloted mana required to issue a Block
// with 4 strong parents, the transaction payload from the builder and 1 allotment for the block issuer.
func (b *TransactionBuilder) MinRequiredAllotedMana(workScoreParameters *iotago.WorkScoreParameters, rmc iotago.Mana, blockIssuerAccountID iotago.AccountID) (iotago.Mana, error) {
	// clone the essence allotments to not modify the original transaction
	allotmentsCpy := b.transaction.Allotments.Clone()

	// undo the changes to the allotments at the end
	defer func() {
		b.transaction.Allotments = allotmentsCpy
	}()

	// add a dummy allotment to account for the later added allotment for the block issuer in case it does not exist yet
	b.IncreaseAllotment(blockIssuerAccountID, 1074)

	// create a signed transaction with a empty signer to get the correct workscore.
	// later the transaction needs to be signed with the correct signer, after the alloted mana was set correctly.
	dummyTxPayload, err := b.Build(&iotago.EmptyAddressSigner{})
	if err != nil {
		return 0, ierrors.Wrap(err, "failed to build the transaction payload")
	}

	payloadWorkScore, err := dummyTxPayload.WorkScore(workScoreParameters)
	if err != nil {
		return 0, ierrors.Wrap(err, "failed to calculate the transaction payload workscore")
	}

	workScore, err := workScoreParameters.Block.Add(payloadWorkScore)
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

	b.transaction.Allotments.Sort()
	b.transaction.TransactionEssence.ContextInputs.Sort()

	// prepare the inputs commitment in the same order as the inputs in the essence
	var inputIDs iotago.OutputIDs
	for _, input := range b.transaction.TransactionEssence.Inputs {
		//nolint:forcetypeassert // we can safely assume that this is an UTXOInput
		inputIDs = append(inputIDs, input.(*iotago.UTXOInput).OutputID())
	}

	inputs := inputIDs.OrderedSet(b.inputs)

	txEssenceData, err := b.transaction.SigningMessage()
	if err != nil {
		return nil, ierrors.Wrap(err, "failed to calculate tx transaction for signing message")
	}

	unlockPos := map[string]int{}
	unlocks := iotago.Unlocks{}
	for i, inputRef := range b.transaction.TransactionEssence.Inputs {
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
	case *iotago.AnchorAddress:
		return append(unlocks, &iotago.AnchorUnlock{Reference: uint16(pos)})
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
