package vm

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/iotaledger/hive.go/core/safemath"
	"github.com/iotaledger/hive.go/ierrors"
	iotago "github.com/iotaledger/iota.go/v4"
)

// VirtualMachine executes and validates transactions.
type VirtualMachine interface {
	// Execute executes the given tx in the VM.
	// Pass own ExecFunc(s) to override the VM's default execution function list.
	Execute(t *iotago.SignedTransaction, params *Params, inputs ResolvedInputs, overrideFuncs ...ExecFunc) error
	// ChainSTVF executes the chain state transition validation function.
	ChainSTVF(transType iotago.ChainTransitionType, input *ChainOutputWithCreationSlot, next iotago.ChainOutput, vmParams *Params) error
}

// Params defines the VirtualMachine parameters under which the VM operates.
type Params struct {
	API iotago.API

	// The working set which is auto. populated during the semantic validation.
	WorkingSet *WorkingSet
}

// WorkingSet contains fields which get automatically populated
// by the library during execution of a Transaction.
type WorkingSet struct {
	// The identities which are successfully unlocked from the input side.
	UnlockedIdents UnlockedIdentities
	// The inputs to the transaction.
	UTXOInputs iotago.Outputs[iotago.Output]
	// The UTXO inputs to the transaction with their creation slots.
	UTXOInputsWithCreationSlot InputSet
	// The mapping of inputs' OutputID to the index.
	InputIDToIndex map[iotago.OutputID]uint16
	// The transaction for which this semantic validation happens.
	Tx *iotago.SignedTransaction
	// The message which signatures are signing.
	EssenceMsgToSign []byte
	// The inputs of the transaction mapped by type.
	InputsByType iotago.OutputsByType
	// The ChainOutput(s) at the input side.
	InChains ChainInputSet
	// The sum of NativeTokens at the input side.
	InNativeTokens iotago.NativeTokenSum
	// The Outputs of the transaction mapped by type.
	OutputsByType iotago.OutputsByType
	// The ChainOutput(s) at the output side.
	OutChains iotago.ChainOutputSet
	// The sum of NativeTokens at the output side.
	OutNativeTokens iotago.NativeTokenSum
	// The Unlocks carried by the transaction mapped by type.
	UnlocksByType iotago.UnlocksByType
	// BIC is the block issuance credit for MCA slots prior to the transaction's creation slot (or for the slot to which the block commits)
	// Contains one value for each account output touched in the transaction and empty if no account outputs touched.
	BIC BlockIssuanceCreditInputSet
	// Commitment contains set of commitment inputs necessary for transaction execution. FIXME
	Commitment VMCommitmentInput
	// Rewards contains a set of account or delegation IDs mapped to their rewards amount.
	Rewards RewardsInputSet
}

// UTXOInputAtIndex retrieves the UTXOInput at the given index.
// Caller must ensure that the index is valid.
func (workingSet *WorkingSet) UTXOInputAtIndex(inputIndex uint16) *iotago.UTXOInput {
	//nolint:forcetypeassert // we can safely assume that this is a UTXOInput
	return workingSet.Tx.Transaction.Inputs[inputIndex].(*iotago.UTXOInput)
}

func NewVMParamsWorkingSet(api iotago.API, t *iotago.SignedTransaction, inputs ResolvedInputs) (*WorkingSet, error) {
	var err error
	utxoInputsSet := inputs.InputSet
	workingSet := &WorkingSet{}
	workingSet.Tx = t
	workingSet.UnlockedIdents = make(UnlockedIdentities)
	workingSet.UTXOInputsWithCreationSlot = utxoInputsSet
	workingSet.InputIDToIndex = make(map[iotago.OutputID]uint16)
	for inputIndex, inputRef := range workingSet.Tx.Transaction.Inputs {
		//nolint:forcetypeassert // we can safely assume that this is an UTXOInput
		ref := inputRef.(*iotago.UTXOInput).OutputID()
		workingSet.InputIDToIndex[ref] = uint16(inputIndex)
		input, ok := workingSet.UTXOInputsWithCreationSlot[ref]
		if !ok {
			return nil, ierrors.Wrapf(iotago.ErrMissingUTXO, "utxo for input %d not supplied", inputIndex)
		}
		workingSet.UTXOInputs = append(workingSet.UTXOInputs, input.Output)
	}

	workingSet.EssenceMsgToSign, err = t.Transaction.SigningMessage(api)
	if err != nil {
		return nil, err
	}

	workingSet.InputsByType = func() iotago.OutputsByType {
		slice := make(iotago.Outputs[iotago.Output], len(utxoInputsSet))
		var i int
		for _, output := range utxoInputsSet {
			slice[i] = output.Output
			i++
		}

		return slice.ToOutputsByType()
	}()

	txID, err := workingSet.Tx.ID(api)
	if err != nil {
		return nil, err
	}

	workingSet.InChains = utxoInputsSet.ChainInputSet()
	workingSet.OutputsByType = t.Transaction.Outputs.ToOutputsByType()
	workingSet.OutChains = workingSet.Tx.Transaction.Outputs.ChainOutputSet(txID)

	workingSet.UnlocksByType = t.Unlocks.ToUnlockByType()
	workingSet.BIC = inputs.BlockIssuanceCreditInputSet
	workingSet.Commitment = inputs.CommitmentInput
	workingSet.Rewards = inputs.RewardsInputSet

	return workingSet, nil
}

func TotalManaIn(manaDecayProvider *iotago.ManaDecayProvider, rentStructure *iotago.RentStructure, txCreationSlot iotago.SlotIndex, inputSet InputSet) (iotago.Mana, error) {
	var totalIn iotago.Mana
	for outputID, input := range inputSet {
		// stored Mana
		manaStored, err := manaDecayProvider.ManaWithDecay(input.Output.StoredMana(), input.CreationSlot, txCreationSlot)
		if err != nil {
			return 0, ierrors.Wrapf(err, "input %s stored mana calculation failed", outputID)
		}
		totalIn, err = safemath.SafeAdd(totalIn, manaStored)
		if err != nil {
			return 0, ierrors.Wrapf(iotago.ErrManaOverflow, "%w", err)
		}

		// potential Mana
		// the storage deposit does not generate potential mana, so we only use the excess base tokens to calculate the potential mana
		minDeposit := rentStructure.MinDeposit(input.Output)
		if input.Output.BaseTokenAmount() <= minDeposit {
			continue
		}
		excessBaseTokens := input.Output.BaseTokenAmount() - minDeposit
		manaPotential, err := manaDecayProvider.ManaGenerationWithDecay(excessBaseTokens, input.CreationSlot, txCreationSlot)
		if err != nil {
			return 0, ierrors.Wrapf(err, "input %s potential mana calculation failed", outputID)
		}
		totalIn, err = safemath.SafeAdd(totalIn, manaPotential)
		if err != nil {
			return 0, ierrors.Wrapf(iotago.ErrManaOverflow, "%w", err)
		}
	}

	return totalIn, nil
}

func TotalManaOut(outputs iotago.Outputs[iotago.TxEssenceOutput], allotments iotago.Allotments) (iotago.Mana, error) {
	var totalOut iotago.Mana
	var err error

	for _, output := range outputs {
		totalOut, err = safemath.SafeAdd(totalOut, output.StoredMana())
		if err != nil {
			return 0, ierrors.Wrapf(iotago.ErrManaOverflow, "%w", err)
		}
	}
	for _, allotment := range allotments {
		totalOut, err = safemath.SafeAdd(totalOut, allotment.Value)
		if err != nil {
			return 0, ierrors.Wrapf(iotago.ErrManaOverflow, "%w", err)
		}
	}

	return totalOut, nil
}

// PastBoundedSlotIndex calculates the past bounded slot for the given slot.
// Given any slot index of a commitment input, the result of this function is a slot index
// that is at least equal to the slot of the block in which it was issued, or higher.
// That means no commitment input can be chosen such that the index lies behind the slot index of the block,
// hence the past is bounded.
func (params *Params) PastBoundedSlotIndex(commitmentInputSlot iotago.SlotIndex) iotago.SlotIndex {
	return commitmentInputSlot + params.API.ProtocolParameters().MaxCommittableAge()
}

// FutureBoundedSlotIndex calculates the future bounded slot for the given slot.
// Given any slot index of a commitment input, the result of this function is a slot index
// that is at most equal to the slot of the block in which it was issued, or lower.
// That means no commitment input can be chosen such that the index lies ahead of the slot index of the block,
// hence the future is bounded.
func (params *Params) FutureBoundedSlotIndex(commitmentInputSlot iotago.SlotIndex) iotago.SlotIndex {
	return commitmentInputSlot + params.API.ProtocolParameters().MinCommittableAge()
}

// RunVMFuncs runs the given ExecFunc(s) in serial order.
func RunVMFuncs(vm VirtualMachine, vmParams *Params, execFuncs ...ExecFunc) error {
	for _, execFunc := range execFuncs {
		if err := execFunc(vm, vmParams); err != nil {
			return err
		}
	}

	return nil
}

// UnlockedIdentities defines a set of identities which are unlocked from the input side of a Transaction.
// The value represent the index of the unlock which unlocked the identity.
type UnlockedIdentities map[string]*UnlockedIdentity

// SigUnlock performs a signature unlock check and adds the given ident to the set of unlocked identities if
// the signature is valid, otherwise returns an error.
func (unlockedIdents UnlockedIdentities) SigUnlock(ident iotago.DirectUnlockableAddress, essence []byte, sig iotago.Signature, inputIndex uint16, checkUnlockOnly bool) error {
	if err := ident.Unlock(essence, sig); err != nil {
		return ierrors.Wrapf(err, "input %d's address is not unlocked through its signature unlock", inputIndex)
	}

	if checkUnlockOnly {
		return nil
	}

	unlockedIdents[ident.Key()] = &UnlockedIdentity{
		Ident:        ident,
		UnlockedAt:   inputIndex,
		ReferencedBy: map[uint16]struct{}{},
	}

	return nil
}

// RefUnlock performs a check whether the given ident is unlocked at ref and if so,
// adds the index of the input to the set of unlocked inputs by this identity.
func (unlockedIdents UnlockedIdentities) RefUnlock(identKey string, ref uint16, inputIndex uint16, checkUnlockOnly bool) error {
	ident, has := unlockedIdents[identKey]
	if !has || ident.UnlockedAt != ref {
		return ierrors.Errorf("input %d is not unlocked through input %d's unlock", inputIndex, ref)
	}

	if checkUnlockOnly {
		return nil
	}

	ident.ReferencedBy[inputIndex] = struct{}{}

	return nil
}

// MultiUnlock performs a check whether all given unlocks are valid and if so,
// adds the index of the input to the set of unlocked inputs by this identity.
func (unlockedIdents UnlockedIdentities) MultiUnlock(vmParams *Params, ident *iotago.MultiAddress, multiUnlock *iotago.MultiUnlock, inputIndex uint16) error {
	if len(ident.Addresses) != len(multiUnlock.Unlocks) {
		return ierrors.Wrapf(iotago.ErrMultiAddressAndUnlockLengthDoesNotMatch, "input %d has a multi address (%T) but the amount of addresses does not match the unlocks %d != %d", inputIndex, ident, len(ident.Addresses), len(multiUnlock.Unlocks))
	}

	var cumulativeUnlockedWeight uint16
	for subIndex, unlock := range multiUnlock.Unlocks {
		switch unlock.(type) {
		case *iotago.EmptyUnlock:
			// EmptyUnlocks are simply skipped. They are used to maintain correct index relationship between
			// addresses and signatures if the signer doesn't know the signature of another signer.
			continue

		case *iotago.MultiUnlock:
			return ierrors.Wrapf(iotago.ErrNestedMultiUnlock, "unlock at index %d.%d is invalid", inputIndex, subIndex)

		default:
			// ATTENTION: we perform the checks only, but we do not unlock the input yet.
			if err := unlockIdent(vmParams, ident.Addresses[subIndex].Address, unlock, inputIndex, true); err != nil {
				return err
			}
			// the unlock was successful, add the weight of the address
			cumulativeUnlockedWeight += uint16(ident.Addresses[subIndex].Weight)
		}
	}

	// check if the threshold for a successful unlock was reached
	if cumulativeUnlockedWeight < ident.Threshold {
		return ierrors.Wrapf(iotago.ErrMultiAddressUnlockThresholdNotReached, "input %d has a multi address (%T) but the threshold of valid unlocks was not reached %d < %d", inputIndex, ident, cumulativeUnlockedWeight, ident.Threshold)
	}

	unlockedIdents[ident.Key()] = &UnlockedIdentity{
		Ident:        ident,
		UnlockedAt:   inputIndex,
		ReferencedBy: map[uint16]struct{}{},
	}

	return nil
}

// AddUnlockedChain allocates an UnlockedIdentity for the given chain.
func (unlockedIdents UnlockedIdentities) AddUnlockedChain(chainAddr iotago.ChainAddress, inputIndex uint16) {
	unlockedIdents[chainAddr.Key()] = &UnlockedIdentity{
		Ident:        chainAddr,
		UnlockedAt:   inputIndex,
		ReferencedBy: map[uint16]struct{}{},
	}
}

func (unlockedIdents UnlockedIdentities) String() string {
	var b strings.Builder
	idents := make([]*UnlockedIdentity, 0, len(unlockedIdents))
	for _, ident := range unlockedIdents {
		idents = append(idents, ident)
	}
	sort.Slice(idents, func(i, j int) bool {
		x, y := idents[i].UnlockedAt, idents[j].UnlockedAt
		// prefer to show direct unlockable addresses first in string
		if x == y {
			if _, is := idents[i].Ident.(iotago.ChainAddress); is {
				return false
			}

			return true
		}

		return x < y
	})
	for _, ident := range idents {
		b.WriteString(ident.String() + "\n")
	}

	return b.String()
}

// UnlockedBy checks whether the given input was unlocked either directly by a signature or indirectly
// through a ReferentialUnlock by the given identity.
func (unlockedIdents UnlockedIdentities) UnlockedBy(inputIndex uint16, identKey string) bool {
	unlockedIdent, has := unlockedIdents[identKey]
	if !has {
		return false
	}

	if unlockedIdent.UnlockedAt == inputIndex {
		return true
	}

	_, refUnlocked := unlockedIdent.ReferencedBy[inputIndex]

	return refUnlocked
}

// UnlockedIdentity represents an unlocked identity.
type UnlockedIdentity struct {
	// The source ident which got unlocked.
	Ident iotago.Address
	// The index at which this identity has been unlocked.
	UnlockedAt uint16
	// A set of input/unlock-block indices which referenced this unlocked identity.
	ReferencedBy map[uint16]struct{}
}

func (unlockedIdent *UnlockedIdentity) String() string {
	refs := make([]int, 0, len(unlockedIdent.ReferencedBy))
	for ref := range unlockedIdent.ReferencedBy {
		refs = append(refs, int(ref))
	}
	sort.Ints(refs)

	return fmt.Sprintf("ident %s (%s), unlocked at %d, ref unlocks at %v", unlockedIdent.Ident, unlockedIdent.Ident.Type(),
		unlockedIdent.UnlockedAt, refs)
}

// IsIssuerOnOutputUnlocked checks whether the issuer in an IssuerFeature of this new ChainOutput has been unlocked.
// This function is a no-op if the chain output does not contain an IssuerFeature.
func IsIssuerOnOutputUnlocked(output iotago.ChainOutputImmutable, unlockedIdents UnlockedIdentities) error {
	immFeats := output.ImmutableFeatureSet()
	if len(immFeats) == 0 {
		return nil
	}

	issuerFeat := immFeats.Issuer()
	if issuerFeat == nil {
		return nil
	}
	if _, isUnlocked := unlockedIdents[issuerFeat.Address.Key()]; !isUnlocked {
		return iotago.ErrIssuerFeatureNotUnlocked
	}

	return nil
}

// ExecFunc is a function which given the context, input, outputs and
// unlocks runs a specific execution/validation. The function might also modify the Params
// in order to supply information to subsequent ExecFunc(s).
type ExecFunc func(vm VirtualMachine, svCtx *Params) error

// ExecFuncInputUnlocks produces the UnlockedIdentities which will be set into the given Params
// and verifies that inputs are correctly unlocked and that the inputs commitment matches.
func ExecFuncInputUnlocks() ExecFunc {
	return func(vm VirtualMachine, vmParams *Params) error {
		actualInputCommitment, err := vmParams.WorkingSet.UTXOInputs.Commitment(vmParams.API)
		if err != nil {
			return ierrors.Join(err, iotago.ErrInvalidInputsCommitment)
		}

		expectedInputCommitment := vmParams.WorkingSet.Tx.Transaction.InputsCommitment[:]
		if !bytes.Equal(expectedInputCommitment, actualInputCommitment) {
			return ierrors.Wrapf(iotago.ErrInvalidInputsCommitment, "specified %v but got %v", expectedInputCommitment, actualInputCommitment)
		}

		for inputIndex, input := range vmParams.WorkingSet.UTXOInputs {
			if err = unlockOutput(vmParams, input, uint16(inputIndex)); err != nil {
				return err
			}

			// since this input is now unlocked, and it is a ChainOutput, the chain's address becomes automatically unlocked
			if chainConstrOutput, is := input.(iotago.ChainOutput); is && chainConstrOutput.Chain().Addressable() {
				// mark this ChainOutput's identity as unlocked by this input
				chainID := chainConstrOutput.Chain()
				if chainID.Empty() {
					//nolint:forcetypeassert // we can safely assume that this is an UTXOIDChainID
					chainID = chainID.(iotago.UTXOIDChainID).FromOutputID(vmParams.WorkingSet.UTXOInputAtIndex(uint16(inputIndex)).OutputID())
				}

				// for account outputs which are not state transitioning, we do not add it to the set of unlocked chains
				if currentAccount, ok := chainConstrOutput.(*iotago.AccountOutput); ok {
					next, hasNextState := vmParams.WorkingSet.OutChains[chainID]
					if !hasNextState {
						continue
					}
					// note that isAccount should never be false in practice,
					// but we add it anyway as an additional safeguard
					nextAccount, isAccount := next.(*iotago.AccountOutput)
					if !isAccount || (currentAccount.StateIndex+1 != nextAccount.StateIndex) {
						continue
					}
				}

				vmParams.WorkingSet.UnlockedIdents.AddUnlockedChain(chainID.ToAddress(), uint16(inputIndex))
			}

		}

		return nil
	}
}

func identToUnlock(vmParams *Params, input iotago.Output, inputIndex uint16) (iotago.Address, error) {
	switch in := input.(type) {

	case iotago.TransIndepIdentOutput:
		return in.Ident(), nil

	case iotago.TransDepIdentOutput:
		chainID := in.Chain()
		if chainID.Empty() {
			utxoChainID, is := chainID.(iotago.UTXOIDChainID)
			if !is {
				return nil, iotago.ErrTransDepIdentOutputNonUTXOChainID
			}
			//nolint:forcetypeassert // we can safely assume that this is an UTXOInput
			chainID = utxoChainID.FromOutputID(vmParams.WorkingSet.Tx.Transaction.Inputs[inputIndex].(*iotago.UTXOInput).OutputID())
		}

		next := vmParams.WorkingSet.OutChains[chainID]
		if next == nil {
			return in.Ident(nil)
		}

		nextTransDepIdentOutput, ok := next.(iotago.TransDepIdentOutput)
		if !ok {
			return nil, iotago.ErrTransDepIdentOutputNextInvalid
		}

		return in.Ident(nextTransDepIdentOutput)

	default:
		panic("unknown ident output type in semantic unlocks")
	}
}

func checkExpiration(vmParams *Params, output iotago.Output) (iotago.Address, error) {
	if output.UnlockConditionSet().HasExpirationCondition() {
		commitment := vmParams.WorkingSet.Commitment

		if commitment == nil {
			return nil, iotago.ErrExpirationConditionCommitmentInputRequired
		}

		futureBoundedSlotIndex := vmParams.FutureBoundedSlotIndex(commitment.Index)
		if ok, returnIdent := output.UnlockConditionSet().ReturnIdentCanUnlock(futureBoundedSlotIndex); ok {
			return returnIdent, nil
		}

		pastBoundedSlotIndex := vmParams.PastBoundedSlotIndex(commitment.Index)
		if output.UnlockConditionSet().OwnerIdentCanUnlock(pastBoundedSlotIndex) {
			return nil, nil
		}

		return nil, iotago.ErrExpirationConditionUnlockFailed
	}

	return nil, nil
}

func unlockIdent(vmParams *Params, ownerIdent iotago.Address, unlock iotago.Unlock, inputIndex uint16, checkUnlockOnly bool) error {

	switch owner := ownerIdent.(type) {
	case iotago.ChainAddress:
		refUnlock, isReferentialUnlock := unlock.(iotago.ReferentialUnlock)
		if !isReferentialUnlock || !refUnlock.Chainable() || !refUnlock.SourceAllowed(ownerIdent) {
			return ierrors.Wrapf(iotago.ErrInvalidInputUnlock, "input %d has a chain address (%T) but its corresponding unlock is of type %T", inputIndex, owner, unlock)
		}

		if err := vmParams.WorkingSet.UnlockedIdents.RefUnlock(owner.Key(), refUnlock.Ref(), inputIndex, checkUnlockOnly); err != nil {
			return ierrors.Join(iotago.ErrInvalidInputUnlock, ierrors.Wrapf(err, "chain address %s (%T)", owner, owner))
		}

	case iotago.DirectUnlockableAddress:
		switch uBlock := unlock.(type) {
		case iotago.ReferentialUnlock:
			// ReferentialUnlock for DirectUnlockableAddress are only allowed if the unlock is not chainable, and the owner ident is not a ChainAddress.
			if uBlock.Chainable() || !uBlock.SourceAllowed(ownerIdent) {
				return ierrors.Wrapf(iotago.ErrInvalidInputUnlock, "input %d has a non-chain address of %s but its corresponding unlock of type %s is chainable or not allowed", inputIndex, owner.Type(), unlock.Type())
			}

			if err := vmParams.WorkingSet.UnlockedIdents.RefUnlock(owner.Key(), uBlock.Ref(), inputIndex, checkUnlockOnly); err != nil {
				return ierrors.Join(iotago.ErrInvalidInputUnlock, ierrors.Wrapf(err, "direct unlockable address %s (%T)", owner, owner))
			}

		case *iotago.SignatureUnlock:
			// owner must not be unlocked already
			if unlockedAtIndex, wasAlreadyUnlocked := vmParams.WorkingSet.UnlockedIdents[owner.Key()]; wasAlreadyUnlocked {
				return ierrors.Wrapf(iotago.ErrInvalidInputUnlock, "input %d's address is already unlocked through input %d's unlock but the input uses a non referential unlock", inputIndex, unlockedAtIndex)
			}

			if err := vmParams.WorkingSet.UnlockedIdents.SigUnlock(owner, vmParams.WorkingSet.EssenceMsgToSign, uBlock.Signature, inputIndex, checkUnlockOnly); err != nil {
				return ierrors.Join(iotago.ErrUnlockBlockSignatureInvalid, err)
			}

		default:
			return ierrors.Wrapf(iotago.ErrInvalidInputUnlock, "input %d has a direct unlockable address (%T) but its corresponding unlock is of type %T", inputIndex, owner, unlock)
		}

	case *iotago.MultiAddress:
		switch uBlock := unlock.(type) {
		case iotago.ReferentialUnlock:
			if uBlock.Chainable() || !uBlock.SourceAllowed(ownerIdent) {
				return ierrors.Wrapf(iotago.ErrInvalidInputUnlock, "input %d has a non-chain address of %s but its corresponding unlock of type %s is chainable or not allowed", inputIndex, owner.Type(), unlock.Type())
			}

			if err := vmParams.WorkingSet.UnlockedIdents.RefUnlock(owner.Key(), uBlock.Ref(), inputIndex, checkUnlockOnly); err != nil {
				return ierrors.Join(iotago.ErrInvalidInputUnlock, ierrors.Wrapf(err, "multi address %s (%T)", owner, owner))
			}

		case *iotago.MultiUnlock:
			// owner must not be unlocked already
			if unlockedAtIndex, wasAlreadyUnlocked := vmParams.WorkingSet.UnlockedIdents[owner.Key()]; wasAlreadyUnlocked {
				return ierrors.Wrapf(iotago.ErrInvalidInputUnlock, "input %d's address is already unlocked through input %d's unlock but the input uses a non referential unlock", inputIndex, unlockedAtIndex)
			}

			if err := vmParams.WorkingSet.UnlockedIdents.MultiUnlock(vmParams, owner, uBlock, inputIndex); err != nil {
				return ierrors.Join(iotago.ErrInvalidInputUnlock, err)
			}

		default:
			return ierrors.Wrapf(iotago.ErrInvalidInputUnlock, "input %d has a multi address (%T) but its corresponding unlock is of type %T", inputIndex, owner, unlock)
		}

	default:
		panic("unknown address in semantic unlocks")
	}

	return nil
}

// resolveUnderlyingIdent returns the underlying address in case of a restricted address.
// this way we handle restricted addresses like normal addresses in the unlock logic.
func resolveUnderlyingIdent(ident iotago.Address) iotago.Address {
	switch addr := ident.(type) {
	case *iotago.RestrictedAddress:
		return addr.Address
	default:
		return addr
	}
}

func unlockOutput(vmParams *Params, output iotago.Output, inputIndex uint16) error {
	ownerIdent, err := identToUnlock(vmParams, output, inputIndex)
	if err != nil {
		return ierrors.Errorf("unable to retrieve ident to unlock of input %d: %w", inputIndex, err)
	}

	if actualIdentToUnlock, err := checkExpiration(vmParams, output); err != nil {
		return err
	} else if actualIdentToUnlock != nil {
		ownerIdent = actualIdentToUnlock
	}

	unlock := vmParams.WorkingSet.Tx.Unlocks[inputIndex]

	return unlockIdent(vmParams, resolveUnderlyingIdent(ownerIdent), unlock, inputIndex, false)
}

// ExecFuncSenderUnlocked validates that for SenderFeature occurring on the output side,
// the given identity is unlocked on the input side.
func ExecFuncSenderUnlocked() ExecFunc {
	return func(vm VirtualMachine, vmParams *Params) error {
		for outputIndex, output := range vmParams.WorkingSet.Tx.Transaction.Outputs {
			senderFeat := output.FeatureSet().SenderFeature()
			if senderFeat == nil {
				continue
			}

			// check unlocked
			sender := senderFeat.Address
			if _, isUnlocked := vmParams.WorkingSet.UnlockedIdents[sender.Key()]; !isUnlocked {
				return ierrors.Wrapf(iotago.ErrSenderFeatureNotUnlocked, "output %d", outputIndex)
			}
		}

		return nil
	}
}

// ExecFuncBalancedMana validates that Mana is balanced from the input/output side.
func ExecFuncBalancedMana() ExecFunc {
	return func(vm VirtualMachine, vmParams *Params) error {
		txCreationSlot := vmParams.WorkingSet.Tx.Transaction.CreationSlot
		for outputID, input := range vmParams.WorkingSet.UTXOInputsWithCreationSlot {
			if input.CreationSlot > txCreationSlot {
				return ierrors.Wrapf(iotago.ErrInputCreationAfterTxCreation, "input %s has creation slot %d, tx creation slot %d", outputID, input.CreationSlot, txCreationSlot)
			}
		}
		manaIn, err := TotalManaIn(vmParams.API.ManaDecayProvider(), vmParams.API.ProtocolParameters().RentStructure(), txCreationSlot, vmParams.WorkingSet.UTXOInputsWithCreationSlot)
		if err != nil {
			return ierrors.Join(iotago.ErrManaAmountInvalid, err)
		}

		manaOut, err := TotalManaOut(vmParams.WorkingSet.Tx.Transaction.Outputs, vmParams.WorkingSet.Tx.Transaction.Allotments)
		if err != nil {
			return ierrors.Join(iotago.ErrManaAmountInvalid, err)
		}

		// Whether it's valid to claim rewards is checked in the delegation and staking STVFs.
		for _, reward := range vmParams.WorkingSet.Rewards {
			manaIn += reward
		}

		if manaIn < manaOut {
			return ierrors.Wrapf(iotago.ErrInputOutputManaMismatch, "Mana in %d, Mana out %d", manaIn, manaOut)
		}

		return nil
	}
}

// ExecFuncBalancedBaseTokens validates that the base tokens are balanced from the input/output side.
// It additionally also incorporates the check whether return amounts via StorageDepositReturnUnlockCondition(s) for specified identities
// are fulfilled from the output side.
func ExecFuncBalancedBaseTokens() ExecFunc {
	return func(vm VirtualMachine, vmParams *Params) error {
		// note that due to syntactic validation of outputs, input and output base token amount sums
		// are always within bounds of the total token supply
		var in, out iotago.BaseToken
		inputSumReturnAmountPerIdent := make(map[string]iotago.BaseToken)
		for inputID, input := range vmParams.WorkingSet.UTXOInputsWithCreationSlot {
			in += input.Output.BaseTokenAmount()

			returnUnlockCond := input.Output.UnlockConditionSet().StorageDepositReturn()
			if returnUnlockCond == nil {
				continue
			}

			returnIdent := returnUnlockCond.ReturnAddress.Key()

			// if the return ident unlocked this input, then the return amount does
			// not have to be fulfilled (this can happen implicit through an expiration condition)
			if vmParams.WorkingSet.UnlockedIdents.UnlockedBy(vmParams.WorkingSet.InputIDToIndex[inputID], returnIdent) {
				continue
			}

			inputSumReturnAmountPerIdent[returnIdent] += returnUnlockCond.Amount
		}

		outputSimpleTransfersPerIdent := make(map[string]iotago.BaseToken)
		for _, output := range vmParams.WorkingSet.Tx.Transaction.Outputs {
			outAmount := output.BaseTokenAmount()
			out += outAmount

			// accumulate simple transfers for StorageDepositReturnUnlockCondition checks
			if basicOutput, is := output.(*iotago.BasicOutput); is && basicOutput.IsSimpleTransfer() {
				outputSimpleTransfersPerIdent[basicOutput.Ident().Key()] += outAmount
			}
		}

		if in != out {
			return ierrors.Wrapf(iotago.ErrInputOutputSumMismatch, "in %d, out %d", in, out)
		}

		for ident, returnSum := range inputSumReturnAmountPerIdent {
			outSum, has := outputSimpleTransfersPerIdent[ident]
			if !has {
				return ierrors.Wrapf(iotago.ErrReturnAmountNotFulFilled, "return amount of %d not fulfilled as there is no output for %s", returnSum, ident)
			}
			if outSum < returnSum {
				return ierrors.Wrapf(iotago.ErrReturnAmountNotFulFilled, "return amount of %d not fulfilled as output is only %d for %s", returnSum, outSum, ident)
			}
		}

		return nil
	}
}

// ExecFuncTimelocks validates that the inputs' timelocks are expired.
func ExecFuncTimelocks() ExecFunc {
	return func(vm VirtualMachine, vmParams *Params) error {
		for inputIndex, input := range vmParams.WorkingSet.UTXOInputsWithCreationSlot {
			if input.Output.UnlockConditionSet().HasTimelockCondition() {
				commitment := vmParams.WorkingSet.Commitment

				if commitment == nil {
					return iotago.ErrTimelockConditionCommitmentInputRequired
				}
				futureBoundedIndex := vmParams.FutureBoundedSlotIndex(commitment.Index)
				if err := input.Output.UnlockConditionSet().TimelocksExpired(futureBoundedIndex); err != nil {
					return ierrors.Wrapf(err, "input at index %d's timelocks are not expired", inputIndex)
				}
			}
		}

		return nil
	}
}

// ExecFuncChainTransitions executes state transition validation functions on ChainOutput(s).
func ExecFuncChainTransitions() ExecFunc {
	return func(vm VirtualMachine, vmParams *Params) error {
		for chainID, inputChain := range vmParams.WorkingSet.InChains {
			next := vmParams.WorkingSet.OutChains[chainID]
			if next == nil {
				if err := vm.ChainSTVF(iotago.ChainTransitionTypeDestroy, inputChain, nil, vmParams); err != nil {
					return ierrors.Join(iotago.ErrChainTransitionInvalid, ierrors.Wrapf(err, "input chain %s (%T) destruction transition failed", chainID, inputChain))
				}

				continue
			}
			if err := vm.ChainSTVF(iotago.ChainTransitionTypeStateChange, inputChain, next, vmParams); err != nil {
				return ierrors.Join(iotago.ErrChainTransitionInvalid, ierrors.Wrapf(err, "chain %s (%T) state transition failed", chainID, inputChain))
			}
		}

		for chainID, outputChain := range vmParams.WorkingSet.OutChains {
			if _, chainPresentInInputs := vmParams.WorkingSet.InChains[chainID]; chainPresentInInputs {
				continue
			}

			if err := vm.ChainSTVF(iotago.ChainTransitionTypeGenesis, nil, outputChain, vmParams); err != nil {
				return ierrors.Join(iotago.ErrChainTransitionInvalid, ierrors.Wrapf(err, "new chain %s (%T) state transition failed", chainID, outputChain))
			}
		}

		return nil
	}
}

// ExecFuncBalancedNativeTokens validates following rules regarding NativeTokens:
//   - The NativeTokens between Inputs / Outputs must be balanced or have a deficit on the output side if there is no foundry state transition for a given NativeToken.
//   - Max MaxNativeTokensCount native tokens within inputs + outputs
func ExecFuncBalancedNativeTokens() ExecFunc {
	return func(vm VirtualMachine, vmParams *Params) error {
		// native token set creates handle overflows
		var err error
		vmParams.WorkingSet.InNativeTokens, err = vmParams.WorkingSet.UTXOInputs.NativeTokenSum()
		if err != nil {
			return ierrors.Join(iotago.ErrNativeTokenSetInvalid, ierrors.Errorf("invalid input native token set: %w", err))
		}
		inNTCount := len(vmParams.WorkingSet.InNativeTokens)

		if inNTCount > iotago.MaxNativeTokensCount {
			return ierrors.Wrapf(iotago.ErrMaxNativeTokensCountExceeded, "inputs native token count %d exceeds max of %d", inNTCount, iotago.MaxNativeTokensCount)
		}

		vmParams.WorkingSet.OutNativeTokens, err = vmParams.WorkingSet.Tx.Transaction.Outputs.NativeTokenSum()
		if err != nil {
			return ierrors.Join(iotago.ErrNativeTokenSetInvalid, err)
		}

		distinctNTCount := make(map[iotago.NativeTokenID]struct{})
		for nt := range vmParams.WorkingSet.InNativeTokens {
			distinctNTCount[nt] = struct{}{}
		}
		for nt := range vmParams.WorkingSet.OutNativeTokens {
			distinctNTCount[nt] = struct{}{}
		}

		if len(distinctNTCount) > iotago.MaxNativeTokensCount {
			return ierrors.Wrapf(iotago.ErrMaxNativeTokensCountExceeded, "native token count %d exceeds max of %d", distinctNTCount, iotago.MaxNativeTokensCount)
		}

		// check invariants for when token foundry is absent

		for nativeTokenID, inSum := range vmParams.WorkingSet.InNativeTokens {
			if _, foundryIsTransitioning := vmParams.WorkingSet.OutChains[nativeTokenID]; foundryIsTransitioning {
				continue
			}

			// input sum must be greater equal the output sum (burning allows it to be greater)
			if outSum := vmParams.WorkingSet.OutNativeTokens[nativeTokenID]; outSum != nil && inSum.Cmp(outSum) == -1 {
				return ierrors.Wrapf(iotago.ErrNativeTokenSumUnbalanced, "native token %s is less on input (%d) than output (%d) side but the foundry is absent for minting", nativeTokenID, inSum, outSum)
			}
		}

		for nativeTokenID := range vmParams.WorkingSet.OutNativeTokens {
			if _, foundryIsTransitioning := vmParams.WorkingSet.OutChains[nativeTokenID]; foundryIsTransitioning {
				continue
			}

			// foundry must be present when native tokens only reside on the output side
			// as they need to get minted by it within the tx
			if vmParams.WorkingSet.InNativeTokens[nativeTokenID] == nil {
				return ierrors.Wrapf(iotago.ErrNativeTokenSumUnbalanced, "native token %s is new on the output side but the foundry is not transitioning", nativeTokenID)
			}
		}

		// from here the native tokens balancing is handled by each foundry's STVF

		return nil
	}
}

func checkAddressRestrictions(output iotago.TxEssenceOutput, address iotago.Address) error {
	addrWithCapabilities, isAddrWithCapabilities := address.(iotago.AddressCapabilities)
	if !isAddrWithCapabilities {
		// no restrictions
		return nil
	}

	if addrWithCapabilities.CannotReceiveNativeTokens() && len(output.NativeTokenList()) != 0 {
		return iotago.ErrAddressCannotReceiveNativeTokens
	}

	if addrWithCapabilities.CannotReceiveMana() && output.StoredMana() != 0 {
		return iotago.ErrAddressCannotReceiveMana
	}

	if addrWithCapabilities.CannotReceiveOutputsWithTimelockUnlockCondition() && output.UnlockConditionSet().HasTimelockCondition() {
		return iotago.ErrAddressCannotReceiveTimelockUnlockCondition
	}

	if addrWithCapabilities.CannotReceiveOutputsWithExpirationUnlockCondition() && output.UnlockConditionSet().HasExpirationCondition() {
		return iotago.ErrAddressCannotReceiveExpirationUnlockCondition
	}

	if addrWithCapabilities.CannotReceiveOutputsWithStorageDepositReturnUnlockCondition() && output.UnlockConditionSet().HasStorageDepositReturnCondition() {
		return iotago.ErrAddressCannotReceiveStorageDepositReturnUnlockCondition
	}

	if addrWithCapabilities.CannotReceiveAccountOutputs() && output.Type() == iotago.OutputAccount {
		return iotago.ErrAddressCannotReceiveAccountOutput
	}

	if addrWithCapabilities.CannotReceiveNFTOutputs() && output.Type() == iotago.OutputNFT {
		return iotago.ErrAddressCannotReceiveNFTOutput
	}

	if addrWithCapabilities.CannotReceiveDelegationOutputs() && output.Type() == iotago.OutputDelegation {
		return iotago.ErrAddressCannotReceiveDelegationOutput
	}

	return nil
}

// Returns a func that validates that the restrictions on an address are adhered to.
//
// Does not validate the Return Address in StorageDepositReturnUnlockCondition because one cannot
// restrict returning a Basic Output with base tokens.
func ExecFuncAddressRestrictions() ExecFunc {
	return func(vm VirtualMachine, vmParams *Params) error {
		for _, output := range vmParams.WorkingSet.Tx.Transaction.Outputs {
			if addressUnlockCondition := output.UnlockConditionSet().Address(); addressUnlockCondition != nil {
				if err := checkAddressRestrictions(output, addressUnlockCondition.Address); err != nil {
					return err
				}
			}
			if stateControllerUnlockCondition := output.UnlockConditionSet().StateControllerAddress(); stateControllerUnlockCondition != nil {
				if err := checkAddressRestrictions(output, stateControllerUnlockCondition.Address); err != nil {
					return err
				}
			}
			if governorUnlockCondition := output.UnlockConditionSet().GovernorAddress(); governorUnlockCondition != nil {
				if err := checkAddressRestrictions(output, governorUnlockCondition.Address); err != nil {
					return err
				}
			}
			if expirationUnlockCondition := output.UnlockConditionSet().Expiration(); expirationUnlockCondition != nil {
				if err := checkAddressRestrictions(output, expirationUnlockCondition.ReturnAddress); err != nil {
					return err
				}
			}
		}

		return nil
	}
}
