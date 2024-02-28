package builder

import (
	"github.com/iotaledger/hive.go/core/safemath"
	"github.com/iotaledger/hive.go/ierrors"
	iotago "github.com/iotaledger/iota.go/v4"
)

// ErrTransactionBuilder defines a generic error occurring within the TransactionBuilder.
var ErrTransactionBuilder = ierrors.New("transaction builder error")

// NewTransactionBuilder creates a new TransactionBuilder.
func NewTransactionBuilder(api iotago.API, signer iotago.AddressSigner) *TransactionBuilder {
	return &TransactionBuilder{
		api:    api,
		signer: signer,
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
	signer           iotago.AddressSigner
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
		signer:           b.signer,
		occurredBuildErr: b.occurredBuildErr,
		transaction:      b.transaction.Clone(),
		inputs:           b.inputs.Clone(),
		inputOwner:       cpyInputOwner,
		rewards:          b.rewards,
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

// setBuildError sets the build error and returns the builder.
func (b *TransactionBuilder) setBuildError(err error) *TransactionBuilder {
	b.occurredBuildErr = err
	return b
}

// AllotRemainingAccountBoundMana allots all remaining account bound mana to the accounts, except the ignored accounts.
func (b *TransactionBuilder) AllotRemainingAccountBoundMana(targetSlot iotago.SlotIndex, onAllotment func(iotago.AccountID, iotago.Mana), ignoreAccountIDs ...iotago.AccountID) *TransactionBuilder {
	// calculate the remaining mana that was not allotted or stored
	remainingMana, err := b.CalculateAvailableManaRemaining(targetSlot)
	if err != nil {
		return b.setBuildError(err)
	}

	// allot all remaining account bound mana to the accounts, except the ignored accounts
	for accountID, mana := range remainingMana.AccountBoundMana {
		ignoreAccount := false
		for _, ignoreAccountID := range ignoreAccountIDs {
			if accountID == ignoreAccountID {
				ignoreAccount = true
				break
			}
		}

		if ignoreAccount {
			// skip the ignored account
			continue
		}

		// allot the mana to the account
		b.IncreaseAllotment(accountID, mana)

		if onAllotment != nil {
			onAllotment(accountID, mana)
		}
	}

	return b
}

// AllotAllMana allots all remaining account bound mana to the accounts, as well as the remaining unbound mana to the given account.
// It checks if at least the given "minRequiredMana" was allotted to the given account.
func (b *TransactionBuilder) AllotAllMana(targetSlot iotago.SlotIndex, accountID iotago.AccountID, minRequiredMana iotago.Mana) *TransactionBuilder {
	// calculate the remaining mana that was not allotted or stored
	remainingMana, err := b.CalculateAvailableManaRemaining(targetSlot)
	if err != nil {
		return b.setBuildError(err)
	}

	var allottedManaAccountSum iotago.Mana

	// allot all remaining account bound mana to the accounts
	for accID, mana := range remainingMana.AccountBoundMana {
		b.IncreaseAllotment(accID, mana)

		if accID == accountID {
			allottedManaAccountSum += mana
		}
	}

	b.IncreaseAllotment(accountID, remainingMana.UnboundMana)
	allottedManaAccountSum += remainingMana.UnboundMana

	if allottedManaAccountSum < minRequiredMana {
		return b.setBuildError(ierrors.Errorf("not enough mana available to allot to the given account (%s): %d < %d", accountID.String(), allottedManaAccountSum, minRequiredMana))
	}

	return b
}

// getStoredManaOutputAccountID returns the account ID of the output at the given index if it belongs to an account.
// (account output or output with mana lock condition).
func (b *TransactionBuilder) getStoredManaOutputAccountID(storedManaOutputIndex int) (iotago.AccountID, error) {
	if storedManaOutputIndex >= len(b.transaction.Outputs) {
		return iotago.EmptyAccountID, ierrors.Errorf("given storedManaOutputIndex does not exist: %d", storedManaOutputIndex)
	}

	// identify if the stored mana output belongs to an account, so we must not allot remaining account bound mana to that account
	storedManaOutputAccountID := iotago.EmptyAccountID

	switch output := b.transaction.Outputs[storedManaOutputIndex].(type) {
	case *iotago.AccountOutput:
		storedManaOutputAccountID = output.AccountID

	case *iotago.BasicOutput, *iotago.AnchorOutput, *iotago.NFTOutput:
		// check if the output locked mana to a certain account
		if accountID, isManaLocked := b.hasManalockCondition(output); isManaLocked {
			storedManaOutputAccountID = accountID
		}

	default:
		return iotago.EmptyAccountID, ierrors.Wrapf(iotago.ErrUnknownOutputType, "output type %T does not support stored mana", output)
	}

	return storedManaOutputAccountID, nil
}

// storeRemainingManaInOutput moves the remaining unbound mana to stored mana on the specified output index.
func (b *TransactionBuilder) storeRemainingManaInOutput(targetSlot iotago.SlotIndex, storedManaOutputIndex int, storedManaOutputAccountID iotago.AccountID) error {
	if storedManaOutputIndex >= len(b.transaction.Outputs) {
		return ierrors.Errorf("given storedManaOutputIndex does not exist: %d", storedManaOutputIndex)
	}

	// calculate the remaining mana that was not allotted or stored
	remainingMana, err := b.CalculateAvailableManaRemaining(targetSlot)
	if err != nil {
		return err
	}

	remainingManaBalance := remainingMana.UnboundMana

	// check if there is account bound mana that is bound to the same account as the stored mana output
	if !storedManaOutputAccountID.Empty() {
		if accountBoundMana, exists := remainingMana.AccountBoundMana[storedManaOutputAccountID]; exists {
			remainingManaBalance += accountBoundMana
		}
	}

	// move the remaining mana to stored mana on the specified output index
	switch output := b.transaction.Outputs[storedManaOutputIndex].(type) {
	case *iotago.BasicOutput:
		output.Mana += remainingManaBalance
	case *iotago.AccountOutput:
		output.Mana += remainingManaBalance
	case *iotago.AnchorOutput:
		output.Mana += remainingManaBalance
	case *iotago.NFTOutput:
		output.Mana += remainingManaBalance
	}

	return nil
}

// StoreRemainingManaInOutputAndAllotRemainingAccountBoundMana moves the remaining unbound mana to stored mana on the specified output
// index as well as it's account bound mana if available and if the output belongs to an account.
// The remaining account bound mana is allotted to the respective accounts.
func (b *TransactionBuilder) StoreRemainingManaInOutputAndAllotRemainingAccountBoundMana(targetSlot iotago.SlotIndex, storedManaOutputIndex int) *TransactionBuilder {
	storedManaOutputAccountID, err := b.getStoredManaOutputAccountID(storedManaOutputIndex)
	if err != nil {
		return b.setBuildError(err)
	}

	// allot all remaining account bound mana to the accounts, except the account of the stored mana output
	b.AllotRemainingAccountBoundMana(targetSlot, nil, storedManaOutputAccountID)

	// store the remaining mana in the output
	if err = b.storeRemainingManaInOutput(targetSlot, storedManaOutputIndex, storedManaOutputAccountID); err != nil {
		return b.setBuildError(err)
	}

	return b
}

// AllotMinRequiredManaAndStoreRemainingManaInOutput allots the minimum required mana needed to issue a block to the block issuer account
// and moves the remaining unbound mana to stored mana on the specified output index as well as it's account bound mana
// if available and if the output belongs to an account.
// The remaining account bound mana is allotted to the respective accounts.
func (b *TransactionBuilder) AllotMinRequiredManaAndStoreRemainingManaInOutput(targetSlot iotago.SlotIndex, rmc iotago.Mana, blockIssuerAccountID iotago.AccountID, storedManaOutputIndex int) *TransactionBuilder {
	storedManaOutputAccountID, err := b.getStoredManaOutputAccountID(storedManaOutputIndex)
	if err != nil {
		return b.setBuildError(err)
	}

	// allot all remaining account bound mana to the accounts, except the account of the stored mana output
	var allottedManaBlockIssuer iotago.Mana
	b.AllotRemainingAccountBoundMana(targetSlot, func(accountID iotago.AccountID, mana iotago.Mana) {
		if accountID == blockIssuerAccountID {
			// remember the already allotted mana for the block issuer account, so we can subtract it from the minimum required mana later
			allottedManaBlockIssuer = mana
		}
	}, storedManaOutputAccountID)

	// calculate the minimum required mana to issue the block
	minRequiredMana, err := b.MinRequiredAllottedMana(rmc, blockIssuerAccountID)
	if err != nil {
		return b.setBuildError(ierrors.Wrap(err, "failed to calculate the minimum required mana to issue the block"))
	}

	if allottedManaBlockIssuer > 0 {
		// subtract the already allotted mana for the block issuer account from the minimum required mana
		if minRequiredMana < allottedManaBlockIssuer {
			minRequiredMana = 0
		} else {
			minRequiredMana -= allottedManaBlockIssuer
		}
	}

	if minRequiredMana > 0 {
		// allot the mana to the block issuer account (we increase the value, so we don't interfere with the already allotted value).
		// It could be the case that more mana than available is allotted to the block issuer account.
		// This will be checked in the next step.
		b.IncreaseAllotment(blockIssuerAccountID, minRequiredMana)
	}

	if err := b.storeRemainingManaInOutput(targetSlot, storedManaOutputIndex, storedManaOutputAccountID); err != nil {
		return b.setBuildError(err)
	}

	return b
}

// CalculateAvailableManaRemaining calculates the available mana on the input side, subtracts all the mana on the output side
// and on the allotments and returns the remaining mana. It takes the account bound mana into consideration.
// It will return an error if there is not enough mana available.
func (b *TransactionBuilder) CalculateAvailableManaRemaining(targetSlot iotago.SlotIndex) (*AvailableManaResult, error) {
	// calculate the available mana on input side
	availableManaInputs, err := b.CalculateAvailableManaInputs(targetSlot)
	if err != nil {
		return nil, ierrors.Wrap(err, "failed to calculate the available mana on input side")
	}

	// update the account bound mana balances if they exist and/or the onbound mana balance
	updateUnboundAndAccountBoundManaBalances := func(accountID iotago.AccountID, accountBoundManaOut iotago.Mana) error {
		if accountBoundManaOut == 0 {
			return nil
		}

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
				return nil, ierrors.Wrap(err, "failed to subtract the stored mana on the outputs side for account output")
			}

		default:
			// check if the output locked mana to a certain account
			if accountID, isManaLocked := b.hasManalockCondition(output); isManaLocked {
				if err = updateUnboundAndAccountBoundManaBalances(accountID, output.StoredMana()); err != nil {
					return nil, ierrors.Wrap(err, "failed to subtract the stored mana on the outputs side, while checking locked mana")
				}
			} else {
				availableManaInputs.UnboundMana, err = safemath.SafeSub(availableManaInputs.UnboundMana, output.StoredMana())
				if err != nil {
					return nil, ierrors.Wrap(err, "failed to subtract the stored mana on the outputs side")
				}
			}
		}
	}

	// subtract the already allotted mana
	for _, allotment := range b.transaction.Allotments {
		if err = updateUnboundAndAccountBoundManaBalances(allotment.AccountID, allotment.Mana); err != nil {
			return nil, ierrors.Wrap(err, "failed to subtract the already allotted mana")
		}
	}

	return availableManaInputs, nil
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
func (b *TransactionBuilder) BuildAndSwapToBlockBuilder(txFunc TransactionFunc) *BasicBlockBuilder {
	blockBuilder := NewBasicBlockBuilder(b.api)
	tx, err := b.Build()
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

// CalculateAvailableManaInputs calculates the available mana on the input side
// including mana generation and decay and the rewards.
func (b *TransactionBuilder) CalculateAvailableManaInputs(targetSlot iotago.SlotIndex) (*AvailableManaResult, error) {
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

// MinRequiredAllottedMana returns the minimum allotted mana required to issue a Block
// with the transaction payload from the builder and 1 allotment for the block issuer.
func (b *TransactionBuilder) MinRequiredAllottedMana(rmc iotago.Mana, blockIssuerAccountID iotago.AccountID) (iotago.Mana, error) {
	// clone the essence allotments to not modify the original transaction
	allotmentsCpy := b.transaction.Allotments.Clone()

	// undo the changes to the allotments at the end
	defer func() {
		b.transaction.Allotments = allotmentsCpy
	}()

	// add a dummy allotment to account for the later added allotment for the block issuer in case it does not exist yet
	b.IncreaseAllotment(blockIssuerAccountID, 1074)

	// the transaction is "signed" with empty signatures to get the correct workscore.
	// later the transaction needs to be signed with the correct signer, after the allotted mana was set correctly.
	tx, err := b.build(false)
	if err != nil {
		return 0, ierrors.Wrap(err, "failed to build the dummy tx payload for workscore calculation")
	}

	// create a dummy block builder with the transaction payload to get the correct workscore.
	dummyBlockBuilder := NewBasicBlockBuilder(b.api).Payload(tx)

	// normally the block should be build first to sort the parents, but we don't need the block itself, just the workscore
	dummyBlock, err := dummyBlockBuilder.Build()
	if err != nil {
		return 0, ierrors.Wrap(err, "failed to build the dummy block for workscore calculation")
	}

	return dummyBlock.ManaCost(rmc)
}

// Build signs the transaction essence and returns the built payload.
func (b *TransactionBuilder) Build() (*iotago.SignedTransaction, error) {
	return b.build(true)
}

// build adds a signature and returns the built payload.
// Depending on the value of "signEssence" it either signs the essence or adds empty signatures.
func (b *TransactionBuilder) build(signEssence bool) (*iotago.SignedTransaction, error) {
	switch {
	case b.occurredBuildErr != nil:
		return nil, b.occurredBuildErr
	case b.signer == nil:
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

	unlockedSet := &txBuilderUnlockedSet{
		unlocks:        iotago.Unlocks{},
		signerUIDs:     map[iotago.Identifier]int{},
		unlockedChains: map[string]int{},
	}

	for inputIndex, inputRef := range b.transaction.TransactionEssence.Inputs {
		//nolint:forcetypeassert // we can safely assume that this is an UTXOInput
		owner := b.inputOwner[inputRef.(*iotago.UTXOInput).OutputID()]

		chainAddr, isChainAddress := owner.(iotago.ChainAddress)
		if isChainAddress {
			// the inputs's owning chain address must have been unlocked already
			unlockedAtIndex, isUnlocked := unlockedSet.isChainUnlocked(chainAddr)
			if !isUnlocked {
				return nil, ierrors.Errorf("input %d's owning chain is not unlocked, chainID %s, type %s", inputIndex, owner.Bech32(b.api.ProtocolParameters().Bech32HRP()), owner.Type())
			}

			// add a referential unlock to the former unlock position
			unlockedSet.addReferentialUnlock(owner, unlockedAtIndex)

			// always mark the chain as unlocked in case the output is a chain output.
			// e.g. "an NFT owns an NFT".
			unlockedSet.addChainAsUnlocked(inputs[inputIndex], inputIndex)

			// skip the rest of the input processing because we don't need to sign chain inputs
			continue
		}

		if _, isDirectUnlockable := owner.(iotago.DirectUnlockableAddress); !isDirectUnlockable {
			// we only support directly unlockable addresses in the transaction builder for now
			return nil, ierrors.Errorf("input %d's owning address is not directly unlockable, address %s, type %s", inputIndex, owner.Bech32(b.api.ProtocolParameters().Bech32HRP()), owner.Type())
		}

		// get the signer UID for the directly unlockable address
		signerUID, err := b.signer.SignerUIDForAddress(owner)
		if err != nil {
			return nil, ierrors.Wrapf(err, "failed to get signer UID for address %s", owner.Bech32(b.api.ProtocolParameters().Bech32HRP()))
		}

		unlockedAtIndex, alreadyUnlocked := unlockedSet.signerUIDs[signerUID]
		if !alreadyUnlocked {
			var err error
			var signature iotago.Signature
			if signEssence {
				// sign the tx essence data
				signature, err = b.signer.Sign(owner, txEssenceData)
			} else {
				// sign with empty signature.
				// this is used for example to calculate the workscore of the transaction before the actual signing.
				signature, err = b.signer.EmptySignatureForAddress(owner)
			}
			if err != nil {
				return nil, ierrors.Wrapf(err, "failed to sign tx transaction: %s", txEssenceData)
			}

			// add the new signature to the unlocks
			unlockedSet.addUnlock(&iotago.SignatureUnlock{Signature: signature})

			// remember the unlock index for the new signer UID
			unlockedSet.addSignerUID(signerUID, inputIndex)
		} else {
			// add a referential unlock to the former unlock position
			unlockedSet.addReferentialUnlock(owner, unlockedAtIndex)
		}

		// always mark the chain as unlocked in case the output is a chain output
		// e.g. "an NFT owned by an ed25519 address".
		unlockedSet.addChainAsUnlocked(inputs[inputIndex], inputIndex)
	}

	sigTxPayload := &iotago.SignedTransaction{
		API:         b.api,
		Transaction: b.transaction,
		Unlocks:     unlockedSet.unlocks,
	}

	return sigTxPayload, nil
}

// txBuilderUnlockedSet is a helper struct to keep track of the unlocked inputs and their positions.
type txBuilderUnlockedSet struct {
	// unlocks holds the unlocks for the tx inputs.
	unlocks iotago.Unlocks
	// signerUIDs maps unique signer UIDs to the position of the unlock in the unlocks slice.
	signerUIDs map[iotago.Identifier]int
	// unlockedChains maps the chain address key to the position of the unlock in the unlocks slice.
	unlockedChains map[string]int
}

// addUnlock adds the given unlock to the set.
func (u *txBuilderUnlockedSet) addUnlock(unlock iotago.Unlock) {
	u.unlocks = append(u.unlocks, unlock)
}

// addReferentialUnlock adds a referential unlock to the set.
func (u *txBuilderUnlockedSet) addReferentialUnlock(addr iotago.Address, referencedInputIndex int) {
	switch addr.(type) {
	case *iotago.AccountAddress:
		u.addUnlock(&iotago.AccountUnlock{Reference: uint16(referencedInputIndex)})
	case *iotago.AnchorAddress:
		u.addUnlock(&iotago.AnchorUnlock{Reference: uint16(referencedInputIndex)})
	case *iotago.NFTAddress:
		u.addUnlock(&iotago.NFTUnlock{Reference: uint16(referencedInputIndex)})
	default:
		u.addUnlock(&iotago.ReferenceUnlock{Reference: uint16(referencedInputIndex)})
	}
}

// addSignerUID marks the given signer UID as unlocked at "unlockedAtIndex".
func (u *txBuilderUnlockedSet) addSignerUID(signerUID iotago.Identifier, unlockedAtIndex int) {
	u.signerUIDs[signerUID] = unlockedAtIndex
}

// addChainAsUnlocked marks the underlying chain address as unlocked at "unlockedAtIndex".
// This is only done if the output is a chain output and the chain ID is addressable,
// which is only valid for AccountOutput, AnchorOutput and NFTOutput.
func (u *txBuilderUnlockedSet) addChainAsUnlocked(input iotago.Output, unlockedAtIndex int) {
	if chainInput, is := input.(iotago.ChainOutput); is && chainInput.ChainID().Addressable() {
		u.unlockedChains[chainInput.ChainID().ToAddress().Key()] = unlockedAtIndex
	}
}

// isChainUnlocked checks if the underlying chain address is unlocked and returns the unlock index.
func (u *txBuilderUnlockedSet) isChainUnlocked(chainAddr iotago.ChainAddress) (int, bool) {
	unlockedAtIndex, isUnlocked := u.unlockedChains[chainAddr.ChainID().ToAddress().Key()]
	return unlockedAtIndex, isUnlocked
}
