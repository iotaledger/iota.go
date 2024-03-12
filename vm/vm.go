package vm

import (
	"fmt"
	"sort"
	"strings"

	"github.com/iotaledger/hive.go/core/safemath"
	"github.com/iotaledger/hive.go/ierrors"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

// VirtualMachine executes and validates transactions.
type VirtualMachine interface {
	// ValidateUnlocks validates the unlocks of the given SignedTransaction and returns the unlocked addresses.
	ValidateUnlocks(signedTransaction *iotago.SignedTransaction, inputs ResolvedInputs) (unlockedAddrs UnlockedAddresses, err error)
	// Execute executes the given tx in the VM. It is possible to optionally override the default execution functions.
	Execute(transaction *iotago.Transaction, inputs ResolvedInputs, unlockedAddrs UnlockedAddresses, execFunctions ...ExecFunc) (outputs []iotago.Output, err error)
	// ChainSTVF executes the chain state transition validation function.
	ChainSTVF(vmParams *Params, transType iotago.ChainTransitionType, input *ChainOutputWithIDs, next iotago.ChainOutput) error
}

// Params defines the VirtualMachine parameters under which the VM operates.
type Params struct {
	API iotago.API

	// The working set which is auto. populated during the semantic validation.
	WorkingSet *WorkingSet
}

// WorkingSet contains fields which get automatically populated
// by the library during execution of a SignedTransaction.
type WorkingSet struct {
	// The addresses which are successfully unlocked from the input side.
	UnlockedAddrs UnlockedAddresses
	// The UTXO inputs to the transaction.
	UTXOInputs iotago.Outputs[iotago.Output]
	// The mapping of OutputID to the actual inputs.
	UTXOInputsSet InputSet
	// The mapping of inputs' OutputID to the index.
	InputIDToInputIndex map[iotago.OutputID]uint16
	// The transaction for which this semantic validation happens.
	Tx *iotago.Transaction
	// The ChainOutput(s) at the input side.
	InChains ChainInputSet
	// The sum of NativeTokens at the input side.
	InNativeTokens iotago.NativeTokenSum
	// The ChainOutput(s) at the output side.
	OutChains iotago.ChainOutputSet
	// The sum of NativeTokens at the output side.
	OutNativeTokens iotago.NativeTokenSum
	// BIC is the block issuance credit for MCA slots prior to the transaction's creation slot (or for the slot to which the block commits)
	// Contains one value for each account output touched in the transaction and empty if no account outputs touched.
	BIC BlockIssuanceCreditInputSet
	// Commitment contains set of commitment inputs necessary for transaction execution. FIXME
	Commitment VMCommitmentInput
	// Rewards contains a set of account or delegation IDs mapped to their rewards amount.
	Rewards RewardsInputSet
	// TotalManaIn is the total decayed potential and stored Mana from the input side.
	TotalManaIn iotago.Mana
	// TotalManaOut is the total stored and allotted Mana from the output side.
	TotalManaOut iotago.Mana
}

// UTXOInputAtIndex retrieves the UTXOInput at the given index.
// Caller must ensure that the index is valid.
func (workingSet *WorkingSet) UTXOInputAtIndex(inputIndex uint16) *iotago.UTXOInput {
	//nolint:forcetypeassert // we can safely assume that this is a UTXOInput
	return workingSet.Tx.TransactionEssence.Inputs[inputIndex].(*iotago.UTXOInput)
}

func TotalManaIn(manaDecayProvider *iotago.ManaDecayProvider, storageScoreStructure *iotago.StorageScoreStructure, txCreationSlot iotago.SlotIndex, inputSet InputSet, rewards RewardsInputSet) (iotago.Mana, error) {
	var totalIn iotago.Mana

	for outputID, input := range inputSet {
		// stored Mana
		manaStored, err := manaDecayProvider.DecayManaBySlots(input.StoredMana(), outputID.CreationSlot(), txCreationSlot)
		if err != nil {
			return 0, ierrors.Wrapf(err, "stored mana calculation failed for input %s", outputID)
		}
		totalIn, err = safemath.SafeAdd(totalIn, manaStored)
		if err != nil {
			return 0, ierrors.WithMessagef(iotago.ErrManaOverflow, "%w", err)
		}
		manaPotential, err := iotago.PotentialMana(manaDecayProvider, storageScoreStructure, input, outputID.CreationSlot(), txCreationSlot)
		if err != nil {
			return 0, ierrors.Wrapf(err, "input %s potential mana calculation failed", outputID)
		}
		totalIn, err = safemath.SafeAdd(totalIn, manaPotential)
		if err != nil {
			return 0, ierrors.WithMessagef(iotago.ErrManaOverflow, "%w", err)
		}
	}

	// whether it's valid to claim rewards is checked in the delegation and staking STVFs.
	for _, reward := range rewards {
		var err error
		totalIn, err = safemath.SafeAdd(totalIn, reward)
		if err != nil {
			return 0, ierrors.WithMessagef(iotago.ErrManaOverflow, "%w", err)
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
			return 0, ierrors.WithMessagef(iotago.ErrManaOverflow, "%w", err)
		}
	}
	for _, allotment := range allotments {
		totalOut, err = safemath.SafeAdd(totalOut, allotment.Mana)
		if err != nil {
			return 0, ierrors.WithMessagef(iotago.ErrManaOverflow, "%w", err)
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

// unlockedAddressesSet holds a set of unlocked addresses.
type unlockedAddressesSet struct {
	// SignatureUnlockedAddrsByIndex contains direct unlockable addresses only which are unlocked by a signature,
	// indexed by the index of the input/unlock index.
	SignatureUnlockedAddrsByIndex map[uint16]*unlockedAddressWithSignature
	// UnlockedAddrsByAddrKey contains all unlocked addresses indexed by their address key.
	UnlockedAddrsByAddrKey UnlockedAddresses
}

// UnlockedAddresses defines a set of addresses which are unlocked from the input side of a SignedTransaction.
// The value represent the index of the unlock which unlocked the address.
type UnlockedAddresses map[string]*UnlockedAddress

// SignatureUnlock performs a signature unlock check and adds the given address to the set of unlocked addresses if
// the signature is valid, otherwise returns an error.
func (s *unlockedAddressesSet) SignatureUnlock(addr iotago.DirectUnlockableAddress, essenceMsg []byte, sig iotago.Signature, inputIndex uint16, checkUnlockOnly bool) error {
	if err := addr.Unlock(essenceMsg, sig); err != nil {
		return ierrors.Wrapf(err, "input %d's address is not unlocked through its signature unlock", inputIndex)
	}

	if checkUnlockOnly {
		return nil
	}

	unlockedAddr := &UnlockedAddress{
		Address:                addr,
		UnlockedAtInputIndex:   inputIndex,
		ReferencedByInputIndex: map[uint16]struct{}{},
	}

	// we "unlock" the signature here, so it can be used for "ReferentialUnlockDirect" referential unlocks
	s.SignatureUnlockedAddrsByIndex[inputIndex] = &unlockedAddressWithSignature{
		UnlockedAddress: unlockedAddr,
		Signature:       sig,
	}
	s.UnlockedAddrsByAddrKey[addr.Key()] = unlockedAddr

	return nil
}

// ReferentialUnlockNonDirectlyUnlockable expects a non-directly unlockable address and performs a check whether the given address
// is unlocked at referencedInputIndex and if so, it adds the input index to the set of unlocked inputs by this address.
func (s *unlockedAddressesSet) ReferentialUnlockNonDirectlyUnlockable(owner iotago.Address, inputIndex uint16, referencedInputIndex uint16, checkUnlockOnly bool) error {
	if _, isDirectUnlockable := owner.(iotago.DirectUnlockableAddress); isDirectUnlockable {
		return ierrors.Errorf("input %d's address is a directly unlockable address, but a non-directly unlockable address was expected", inputIndex)
	}

	referencedAddr, referenceExists := s.UnlockedAddrsByAddrKey[owner.Key()]
	if !referenceExists {
		return ierrors.Errorf("input %d's address was not previously unlocked by unlock %d", inputIndex, referencedInputIndex)
	}

	if referencedAddr.UnlockedAtInputIndex != referencedInputIndex {
		return ierrors.Errorf("input %d references unlock %d but its address was unlocked by unlock %d instead", inputIndex, referencedInputIndex, referencedAddr.UnlockedAtInputIndex)
	}

	if checkUnlockOnly {
		return nil
	}

	referencedAddr.ReferencedByInputIndex[inputIndex] = struct{}{}

	return nil
}

// ReferentialUnlockDirectlyUnlockable expects a directly unlockable address and performs a check whether the given address
// is unlocked at referencedInputIndex and if the signature of the referenced unlock matches the given address.
// If all checks are successful, it adds the input index to the set of unlocked inputs by this address.
// In case the given address is not yet unlocked, it is added to the set of unlocked addresses.
// This is necessary if for example the signature was used before to unlock another address (e.g. derived from the same public key).
func (s *unlockedAddressesSet) ReferentialUnlockDirectlyUnlockable(owner iotago.DirectUnlockableAddress, inputIndex uint16, referencedInputIndex uint16, checkUnlockOnly bool) error {
	referencedAddrWithSignature, referenceExists := s.SignatureUnlockedAddrsByIndex[referencedInputIndex]
	if !referenceExists {
		return ierrors.Errorf("input %d's address was not previously unlocked by unlock %d", inputIndex, referencedInputIndex)
	}

	if referencedAddrWithSignature.UnlockedAtInputIndex != referencedInputIndex {
		return ierrors.Errorf("input %d references unlock %d but its address was unlocked by unlock %d instead", inputIndex, referencedInputIndex, referencedAddrWithSignature.UnlockedAtInputIndex)
	}

	// the signature was already verified in another unlock, so we don't need to check it again,
	// but we need to make sure that the signature fits the address.
	if !referencedAddrWithSignature.Signature.MatchesAddress(owner) {
		return ierrors.Errorf("input %d's address is not unlocked through unlock %d", inputIndex, referencedInputIndex)
	}

	if checkUnlockOnly {
		return nil
	}

	referencedAddrWithSignature.ReferencedByInputIndex[inputIndex] = struct{}{}

	// we need to add the address to the unlocked addresses map, in case it didn't exist yet.
	// this is necessary if for example the same signature was used before to unlock another address
	// (e.g. derived from the same public key) and is now used in the referential unlock.
	//
	// Attention: for referential unlocks of DirectUnlockableAddress, we don't "unlock" the signature,
	// only the referential unlock address itself. This way the newly unlocked address can't be used
	// for further "ReferentialUnlockDirect" referential unlocks which is correct
	// because the original one should be used instead.
	ownerKey := owner.Key()
	if _, contains := s.UnlockedAddrsByAddrKey[ownerKey]; !contains {
		s.UnlockedAddrsByAddrKey[ownerKey] = &UnlockedAddress{
			Address:                owner,
			UnlockedAtInputIndex:   inputIndex,
			ReferencedByInputIndex: map[uint16]struct{}{},
		}
	}

	return nil
}

// MultiUnlock performs a check whether all given unlocks are valid and if so,
// adds the index of the input to the set of unlocked inputs by this address.
func (s *unlockedAddressesSet) MultiUnlock(addr *iotago.MultiAddress, multiUnlock *iotago.MultiUnlock, essenceMsg []byte, inputIndex uint16) error {
	if len(addr.Addresses) != len(multiUnlock.Unlocks) {
		return ierrors.WithMessagef(iotago.ErrMultiAddressLengthUnlockLengthMismatch, "input %d has a multi address (%T) but the amount of addresses does not match the unlocks %d != %d", inputIndex, addr, len(addr.Addresses), len(multiUnlock.Unlocks))
	}

	var cumulativeUnlockedWeight uint16
	for subIndex, unlock := range multiUnlock.Unlocks {
		switch unlock.(type) {
		case *iotago.EmptyUnlock:
			// EmptyUnlocks are simply skipped. They are used to maintain correct index relationship between
			// addresses and signatures if the signer doesn't know the signature of another signer.
			continue

		case *iotago.MultiUnlock:
			return ierrors.WithMessagef(iotago.ErrNestedMultiUnlock, "unlock at index %d.%d is invalid", inputIndex, subIndex)

		default:
			// ATTENTION: we perform the checks only, but we do not unlock the input yet.
			if err := unlockAddress(addr.Addresses[subIndex].Address, unlock, inputIndex, s, essenceMsg, true); err != nil {
				return err
			}
			// the unlock was successful, add the weight of the address
			cumulativeUnlockedWeight += uint16(addr.Addresses[subIndex].Weight)
		}
	}

	// check if the threshold for a successful unlock was reached
	if cumulativeUnlockedWeight < addr.Threshold {
		return ierrors.WithMessagef(iotago.ErrMultiAddressUnlockThresholdNotReached, "input %d has a multi address but the threshold of valid unlocks was not reached %d < %d", inputIndex, cumulativeUnlockedWeight, addr.Threshold)
	}

	// for multi addresses we don't "unlock" the signatures, only the multi address itself so it can be used for a referential unlock.
	//
	// Attention: the single signatures in the multi unlock must not be able to be referenced by other inputs/unlocks.
	// In case a signature in a multi unlock also exists in a non-multi unlock, the unlock in the multi unlock
	// should be a reference unlock to the non-multi unlock signature.
	s.UnlockedAddrsByAddrKey[addr.Key()] = &UnlockedAddress{
		Address:                addr,
		UnlockedAtInputIndex:   inputIndex,
		ReferencedByInputIndex: map[uint16]struct{}{},
	}

	return nil
}

// AddUnlockedChain unlocks the given chain.
func (s *unlockedAddressesSet) AddUnlockedChain(chainAddr iotago.ChainAddress, inputIndex uint16) {
	s.UnlockedAddrsByAddrKey[chainAddr.Key()] = &UnlockedAddress{
		Address:                chainAddr,
		UnlockedAtInputIndex:   inputIndex,
		ReferencedByInputIndex: map[uint16]struct{}{},
	}
}

func (unlockedAddrs UnlockedAddresses) String() string {
	var b strings.Builder
	addrs := make([]*UnlockedAddress, 0, len(unlockedAddrs))
	for _, addr := range unlockedAddrs {
		addrs = append(addrs, addr)
	}
	sort.Slice(addrs, func(i, j int) bool {
		x, y := addrs[i].UnlockedAtInputIndex, addrs[j].UnlockedAtInputIndex
		// prefer to show direct unlockable addresses first in string
		if x == y {
			if _, is := addrs[i].Address.(iotago.ChainAddress); is {
				return false
			}

			return true
		}

		return x < y
	})
	for _, addr := range addrs {
		b.WriteString(addr.String() + "\n")
	}

	return b.String()
}

// UnlockedBy checks whether the given input was unlocked either directly by a signature or indirectly
// through a ReferentialUnlock by the given address.
func (unlockedAddrs UnlockedAddresses) UnlockedBy(inputIndex uint16, addrKey string) bool {
	unlockedAddr, has := unlockedAddrs[addrKey]
	if !has {
		return false
	}

	if unlockedAddr.UnlockedAtInputIndex == inputIndex {
		return true
	}

	_, refUnlocked := unlockedAddr.ReferencedByInputIndex[inputIndex]

	return refUnlocked
}

// UnlockedAddress represents an unlocked address.
type UnlockedAddress struct {
	// The source address which got unlocked.
	Address iotago.Address
	// The index of the input/unlock by which this address has been unlocked.
	UnlockedAtInputIndex uint16
	// A set of input/unlock indexes which referenced this address.
	ReferencedByInputIndex map[uint16]struct{}
}

type unlockedAddressWithSignature struct {
	*UnlockedAddress

	// Signature is the signature which unlocked the address.
	Signature iotago.Signature
}

func (unlockedAddr *UnlockedAddress) String() string {
	inputIndexes := make([]int, 0, len(unlockedAddr.ReferencedByInputIndex))
	for inputIndex := range unlockedAddr.ReferencedByInputIndex {
		inputIndexes = append(inputIndexes, int(inputIndex))
	}
	sort.Ints(inputIndexes)

	return fmt.Sprintf("address %s (%s), unlocked at %d, referenced by unlocks at %v", unlockedAddr.Address, unlockedAddr.Address.Type(),
		unlockedAddr.UnlockedAtInputIndex, inputIndexes)
}

// IsIssuerOnOutputUnlocked checks whether the issuer in an IssuerFeature of this new ChainOutput has been unlocked.
// This function is a no-op if the chain output does not contain an IssuerFeature.
func IsIssuerOnOutputUnlocked(output iotago.ChainOutputImmutable, unlockedAddrs UnlockedAddresses) error {
	immFeats := output.ImmutableFeatureSet()
	if len(immFeats) == 0 {
		return nil
	}

	issuerFeat := immFeats.Issuer()
	if issuerFeat == nil {
		return nil
	}

	issuer := resolveUnderlyingAddress(issuerFeat.Address)
	if _, isUnlocked := unlockedAddrs[issuer.Key()]; !isUnlocked {
		return iotago.ErrIssuerFeatureNotUnlocked
	}

	return nil
}

// ExecFunc is a function which given the context, input, outputs and
// unlocks runs a specific execution/validation. The function might also modify the Params
// in order to supply information to subsequent ExecFunc(s).
//
//nolint:revive
type ExecFunc func(vm VirtualMachine, svCtx *Params) error

// ValidateUnlocks produces the UnlockedAddresses which will be set into the given Params and verifies that inputs are
// correctly unlocked and that the inputs commitment matches.
func ValidateUnlocks(signedTransaction *iotago.SignedTransaction, resolvedInputs ResolvedInputs) (unlockedAddrs UnlockedAddresses, err error) {
	utxoInputs := signedTransaction.Transaction.Inputs()

	var inputs iotago.Outputs[iotago.Output]
	for _, input := range utxoInputs {
		inputs = append(inputs, resolvedInputs.InputSet[input.OutputID()])
	}

	txID, err := signedTransaction.Transaction.ID()
	if err != nil {
		panic(fmt.Sprintf("transaction ID computation should have succeeded: %s", err.Error()))
	}

	essenceMsgToSign, err := signedTransaction.Transaction.SigningMessage()
	if err != nil {
		panic(fmt.Sprintf("signing message computation should have succeeded: %s", err.Error()))
	}

	unlockedAddrsSet := &unlockedAddressesSet{
		SignatureUnlockedAddrsByIndex: make(map[uint16]*unlockedAddressWithSignature),
		UnlockedAddrsByAddrKey:        make(UnlockedAddresses),
	}

	outChains := signedTransaction.Transaction.Outputs.ChainOutputSet(txID)
	for inputIndex, input := range inputs {
		if err = unlockOutput(signedTransaction.Transaction, resolvedInputs.CommitmentInput, input, signedTransaction.Unlocks[inputIndex], uint16(inputIndex), unlockedAddrsSet, outChains, essenceMsgToSign); err != nil {
			return nil, err
		}

		// since this input is now unlocked, and it is a ChainOutput, the chain's address becomes automatically unlocked
		if chainConstrOutput, is := input.(iotago.ChainOutput); is && chainConstrOutput.ChainID().Addressable() {
			// mark this ChainOutput's address as unlocked by this input
			chainID := chainConstrOutput.ChainID()
			if chainID.Empty() {
				//nolint:forcetypeassert // we can safely assume that this is an UTXOIDChainID
				chainID = chainID.(iotago.UTXOIDChainID).FromOutputID(signedTransaction.Transaction.TransactionEssence.Inputs[inputIndex].(*iotago.UTXOInput).OutputID())
			}

			// for anchor outputs which are not state transitioning, we do not add it to the set of unlocked chains
			if currentAnchor, ok := chainConstrOutput.(*iotago.AnchorOutput); ok {
				next, hasNextState := outChains[chainID]
				if !hasNextState {
					continue
				}
				// note that isAnchor should never be false in practice,
				// but we add it anyway as an additional safeguard
				nextAnchor, isAnchor := next.(*iotago.AnchorOutput)
				if !isAnchor || (currentAnchor.StateIndex+1 != nextAnchor.StateIndex) {
					continue
				}
			}

			unlockedAddrsSet.AddUnlockedChain(chainID.ToAddress(), uint16(inputIndex))
		}
	}

	return unlockedAddrsSet.UnlockedAddrsByAddrKey, err
}

func addressToUnlock(transaction *iotago.Transaction, input iotago.Output, inputIndex uint16, outChains iotago.ChainOutputSet) (iotago.Address, error) {
	switch in := input.(type) {

	case iotago.OwnerTransitionIndependentOutput:
		return in.Owner(), nil

	case iotago.OwnerTransitionDependentOutput:
		chainID := in.ChainID()
		if chainID.Empty() {
			utxoChainID, is := chainID.(iotago.UTXOIDChainID)
			if !is {
				return nil, iotago.ErrOwnerTransitionDependentOutputNonUTXOChainID
			}
			//nolint:forcetypeassert // we can safely assume that this is an UTXOInput
			chainID = utxoChainID.FromOutputID(transaction.TransactionEssence.Inputs[inputIndex].(*iotago.UTXOInput).OutputID())
		}

		next := outChains[chainID]
		if next == nil {
			return in.Owner(nil)
		}

		nextOwnerTransitionDependentOutput, ok := next.(iotago.OwnerTransitionDependentOutput)
		if !ok {
			return nil, iotago.ErrOwnerTransitionDependentOutputNextInvalid
		}

		return in.Owner(nextOwnerTransitionDependentOutput)

	default:
		panic(fmt.Sprintf("unknown address output type in semantic unlocks: %T", in))
	}
}

func checkExpiration(output iotago.Output, commitmentInput VMCommitmentInput, protocolParameters iotago.ProtocolParameters) (iotago.Address, error) {
	if output.UnlockConditionSet().HasExpirationCondition() {
		if commitmentInput == nil {
			return nil, iotago.ErrExpirationCommitmentInputMissing
		}

		pastBoundedSlotIndex := commitmentInput.Slot + protocolParameters.MaxCommittableAge()
		futureBoundedSlotIndex := commitmentInput.Slot + protocolParameters.MinCommittableAge()

		return output.UnlockConditionSet().CheckExpirationCondition(futureBoundedSlotIndex, pastBoundedSlotIndex)
	}

	return nil, nil
}

func unlockAddress(ownerAddr iotago.Address, unlock iotago.Unlock, inputIndex uint16, unlockedAddrsSet *unlockedAddressesSet, essenceMsgToSign []byte, checkUnlockOnly bool) error {
	switch owner := ownerAddr.(type) {
	case iotago.ChainAddress:
		// ChainAddress can either be AccountAddress, AnchorAddress or an NFTAddress.
		// The ChainAddress itself must have been unlocked before (e.g. by a signature unlock of a DirectUnlockableAddress),
		// that's why we only expect a ReferentialUnlock here.
		//
		// Why are other types of "ownerAddr" impossible here:
		// 	Ed25519Address or ImplicitAccountCreationAddress are impossible because the switch statement over the "ownerAddr" type prevents it and the SourceAllowed check would have failed.
		// 	RestrictedAddress is impossible because it is resolved to the underlying "ownerAddr" before this function is called.
		// 	MultiAddress is impossible because the switch statement over the "ownerAddr" type prevents it and the SourceAllowed check would have failed.
		//
		// Therefore the "ReferentialUnlock" must be an "AccountUnlock", "AnchorUnlock" or "NFTUnlock".
		referentialUnlock, isReferentialUnlock := unlock.(iotago.ReferentialUnlock)
		if !isReferentialUnlock || !referentialUnlock.Chainable() || !referentialUnlock.SourceAllowed(ownerAddr) {
			return ierrors.WithMessagef(
				iotago.ErrChainAddressUnlockInvalid,
				"input %d has a chain address of type %s but its corresponding unlock is of type %s", inputIndex, owner.Type(), unlock.Type(),
			)
		}

		if err := unlockedAddrsSet.ReferentialUnlockNonDirectlyUnlockable(owner, inputIndex, referentialUnlock.ReferencedInputIndex(), checkUnlockOnly); err != nil {
			return ierrors.Errorf("%w %s (%s): %w", iotago.ErrChainAddressUnlockInvalid, owner, owner.Type(), err)
		}

	case iotago.DirectUnlockableAddress:
		// DirectUnlockableAddress can either be an Ed25519Address or an ImplicitAccountCreationAddress.
		// The DirectUnlockableAddress can be unlocked by a SignatureUnlock or a ReferentialUnlock.
		switch uBlock := unlock.(type) {
		case iotago.ReferentialUnlock:
			// ReferentialUnlock for DirectUnlockableAddress are only allowed if the unlock is not chainable, and the owner address is not a ChainAddress.
			// This basically means that the ReferentialUnlock must be a ReferenceUnlock and the "ownerAddr" is an Ed25519Address or an ImplicitAccountCreationAddress.
			//
			// Why are other types of "ownerAddr" impossible here:
			// 	AccountAddress, AnchorAddress or NFTAddress are impossible because the switch statement over the "ownerAddr" type prevents it and the SourceAllowed check would have failed.
			// 	RestrictedAddress is impossible because it is resolved to the underlying "ownerAddr" before this function is called.
			// 	MultiAddress is impossible because the switch statement over the "ownerAddr" type prevents it.
			if uBlock.Chainable() || !uBlock.SourceAllowed(ownerAddr) {
				return ierrors.WithMessagef(
					iotago.ErrDirectUnlockableAddressUnlockInvalid,
					"input %d has a non-chain address of type %s but its corresponding unlock of type %s is chainable or not allowed", inputIndex, owner.Type(), unlock.Type(),
				)
			}

			if err := unlockedAddrsSet.ReferentialUnlockDirectlyUnlockable(owner, inputIndex, uBlock.ReferencedInputIndex(), checkUnlockOnly); err != nil {
				return ierrors.Errorf("%w %s (%s): %w", iotago.ErrDirectUnlockableAddressUnlockInvalid, owner, owner.Type(), err)
			}

		case *iotago.SignatureUnlock:
			// owner must not be unlocked already
			if unlockedAddr, wasAlreadyUnlocked := unlockedAddrsSet.UnlockedAddrsByAddrKey[owner.Key()]; wasAlreadyUnlocked {
				return ierrors.WithMessagef(
					iotago.ErrDirectUnlockableAddressUnlockInvalid,
					"input %d's address is already unlocked through input %d's unlock but the input uses a non referential unlock of type %s", inputIndex, unlockedAddr.UnlockedAtInputIndex, unlock.Type(),
				)
			}

			if err := unlockedAddrsSet.SignatureUnlock(owner, essenceMsgToSign, uBlock.Signature, inputIndex, checkUnlockOnly); err != nil {
				return ierrors.Join(iotago.ErrDirectUnlockableAddressUnlockInvalid, iotago.ErrUnlockSignatureInvalid, err)
			}

		default:
			return ierrors.WithMessagef(iotago.ErrDirectUnlockableAddressUnlockInvalid, "input %d has a direct unlockable address of type %s but its corresponding unlock is of type %s", inputIndex, owner.Type(), unlock.Type())
		}

	case *iotago.MultiAddress:
		switch uBlock := unlock.(type) {
		// The MultiAddress can be unlocked by a MultiUnlock or a ReferentialUnlock.
		case iotago.ReferentialUnlock:
			// ReferentialUnlock for MultiAddress are only allowed if the unlock is not chainable, and the owner address is not a ChainAddress.
			// This basically means that the ReferentialUnlock must be a ReferenceUnlock and the "ownerAddr" is a MultiAddress.
			//
			// "ownerAddr" can only be a MultiAddress here, because the switch statement over the "ownerAddr" type prevents other types.
			if uBlock.Chainable() || !uBlock.SourceAllowed(ownerAddr) {
				return ierrors.WithMessagef(iotago.ErrMultiAddressUnlockInvalid,
					"input %d has a non-chain address of %s but its corresponding unlock of type %s is chainable or not allowed",
					inputIndex, owner.Type(), unlock.Type(),
				)
			}

			if err := unlockedAddrsSet.ReferentialUnlockNonDirectlyUnlockable(owner, inputIndex, uBlock.ReferencedInputIndex(), checkUnlockOnly); err != nil {
				return ierrors.Errorf("%w %s (%s): %w", iotago.ErrMultiAddressUnlockInvalid, owner, owner.Type(), err)
			}

		case *iotago.MultiUnlock:
			// owner must not be unlocked already
			if unlockedAddr, wasAlreadyUnlocked := unlockedAddrsSet.UnlockedAddrsByAddrKey[owner.Key()]; wasAlreadyUnlocked {
				return ierrors.WithMessagef(iotago.ErrMultiAddressUnlockInvalid, "input %d's address is already unlocked through input %d's unlock but the input uses a non referential unlock", inputIndex, unlockedAddr.UnlockedAtInputIndex)
			}

			if err := unlockedAddrsSet.MultiUnlock(owner, uBlock, essenceMsgToSign, inputIndex); err != nil {
				return ierrors.WithMessagef(iotago.ErrMultiAddressUnlockInvalid, "%w", err)
			}

		default:
			return ierrors.WithMessagef(iotago.ErrMultiAddressUnlockInvalid, "input %d has a multi address but its corresponding unlock is of type %s", inputIndex, unlock.Type())
		}

	default:
		panic("unknown address in semantic unlocks")
	}

	return nil
}

// resolveUnderlyingAddress returns the underlying address in case of a restricted address.
// this way we handle restricted addresses like normal addresses in the unlock logic.
func resolveUnderlyingAddress(addr iotago.Address) iotago.Address {
	switch addr := addr.(type) {
	case *iotago.RestrictedAddress:
		return addr.Address
	default:
		return addr
	}
}

func unlockOutput(transaction *iotago.Transaction, commitmentInput VMCommitmentInput, input iotago.Output, unlock iotago.Unlock, inputIndex uint16, unlockedAddrsSet *unlockedAddressesSet, outChains iotago.ChainOutputSet, essenceMsgToSign []byte) error {
	ownerAddr, err := addressToUnlock(transaction, input, inputIndex, outChains)
	if err != nil {
		return ierrors.Wrapf(err, "unable to retrieve address to unlock of input %d", inputIndex)
	}

	if actualAddrToUnlock, err := checkExpiration(input, commitmentInput, transaction.API.ProtocolParameters()); err != nil {
		return err
	} else if actualAddrToUnlock != nil {
		ownerAddr = actualAddrToUnlock
	}

	return unlockAddress(resolveUnderlyingAddress(ownerAddr), unlock, inputIndex, unlockedAddrsSet, essenceMsgToSign, false)
}

// ExecFuncSenderUnlocked validates that for SenderFeature occurring on the output side,
// the given address is unlocked on the input side.
func ExecFuncSenderUnlocked() ExecFunc {
	return func(_ VirtualMachine, vmParams *Params) error {
		for outputIndex, output := range vmParams.WorkingSet.Tx.Outputs {
			senderFeat := output.FeatureSet().SenderFeature()
			if senderFeat == nil {
				continue
			}

			// check unlocked
			sender := resolveUnderlyingAddress(senderFeat.Address)
			if _, isUnlocked := vmParams.WorkingSet.UnlockedAddrs[sender.Key()]; !isUnlocked {
				return ierrors.WithMessagef(iotago.ErrSenderFeatureNotUnlocked, "output %d", outputIndex)
			}
		}

		return nil
	}
}

// ExecFuncBalancedMana validates that Mana is balanced from the input/output side.
func ExecFuncBalancedMana() ExecFunc {
	return func(_ VirtualMachine, vmParams *Params) error {
		txCreationSlot := vmParams.WorkingSet.Tx.CreationSlot
		for outputID := range vmParams.WorkingSet.UTXOInputsSet {
			if outputID.CreationSlot() > txCreationSlot {
				return ierrors.WithMessagef(iotago.ErrInputCreationAfterTxCreation, "input %s has creation slot %d, tx creation slot %d", outputID, outputID.CreationSlot(), txCreationSlot)
			}
		}
		manaIn := vmParams.WorkingSet.TotalManaIn
		manaOut := vmParams.WorkingSet.TotalManaOut

		if manaIn < manaOut {
			// less mana on input side than on output side => not allowed
			return ierrors.WithMessagef(iotago.ErrInputOutputManaMismatch, "Mana in %d, Mana out %d", manaIn, manaOut)
		} else if manaIn > manaOut {
			// less mana on output side than on input side => check if mana burning is allowed
			if vmParams.WorkingSet.Tx.Capabilities.CannotBurnMana() {
				return ierrors.WithMessagef(iotago.ErrInputOutputManaMismatch, "%w", iotago.ErrTxCapabilitiesManaBurningNotAllowed)
			}
		}

		return nil
	}
}

// ExecFuncBalancedBaseTokens validates that the base tokens are balanced from the input/output side.
// It additionally also incorporates the check whether return amounts via StorageDepositReturnUnlockCondition(s) for specified addresses
// are fulfilled from the output side.
func ExecFuncBalancedBaseTokens() ExecFunc {
	return func(_ VirtualMachine, vmParams *Params) error {
		// note that due to syntactic validation of outputs, input and output base token amount sums
		// are always within bounds of the total token supply
		var in, out iotago.BaseToken
		inputSumReturnAmountPerAddress := make(map[string]iotago.BaseToken)
		for inputID, input := range vmParams.WorkingSet.UTXOInputsSet {
			in += input.BaseTokenAmount()

			returnUnlockCond := input.UnlockConditionSet().StorageDepositReturn()
			if returnUnlockCond == nil {
				continue
			}

			returnAddr := returnUnlockCond.ReturnAddress.Key()

			// if the return address unlocked this input, then the return amount does
			// not have to be fulfilled (this can happen implicit through an expiration condition)
			if vmParams.WorkingSet.UnlockedAddrs.UnlockedBy(vmParams.WorkingSet.InputIDToInputIndex[inputID], returnAddr) {
				continue
			}

			inputSumReturnAmountPerAddress[returnAddr] += returnUnlockCond.Amount
		}

		outputSimpleTransfersPerAddr := make(map[string]iotago.BaseToken)
		for _, output := range vmParams.WorkingSet.Tx.Outputs {
			outAmount := output.BaseTokenAmount()
			out += outAmount

			// accumulate simple transfers for StorageDepositReturnUnlockCondition checks
			if basicOutput, is := output.(*iotago.BasicOutput); is && basicOutput.IsSimpleTransfer() {
				outputSimpleTransfersPerAddr[basicOutput.Owner().Key()] += outAmount
			}
		}

		if in != out {
			return ierrors.WithMessagef(iotago.ErrInputOutputBaseTokenMismatch, "in %d, out %d", in, out)
		}

		for addr, returnSum := range inputSumReturnAmountPerAddress {
			outSum, has := outputSimpleTransfersPerAddr[addr]
			if !has {
				return ierrors.WithMessagef(iotago.ErrReturnAmountNotFulFilled, "return amount of %d not fulfilled as there is no output for address (serialized) %s", returnSum, hexutil.EncodeHex([]byte(addr)))
			}
			if outSum < returnSum {
				return ierrors.WithMessagef(iotago.ErrReturnAmountNotFulFilled, "return amount of %d not fulfilled as output is only %d for address (serialized) %s", returnSum, outSum, hexutil.EncodeHex([]byte(addr)))
			}
		}

		return nil
	}
}

// ExecFuncTimelocks validates that the inputs' timelocks are expired.
func ExecFuncTimelocks() ExecFunc {
	return func(_ VirtualMachine, vmParams *Params) error {
		for inputIndex, input := range vmParams.WorkingSet.UTXOInputsSet {
			if input.UnlockConditionSet().HasTimelockCondition() {
				commitment := vmParams.WorkingSet.Commitment

				if commitment == nil {
					return iotago.ErrTimelockCommitmentInputMissing
				}
				futureBoundedIndex := vmParams.FutureBoundedSlotIndex(commitment.Slot)
				if err := input.UnlockConditionSet().TimelocksExpired(futureBoundedIndex); err != nil {
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
				if err := vm.ChainSTVF(vmParams, iotago.ChainTransitionTypeDestroy, inputChain, nil); err != nil {
					return ierrors.Wrapf(err, "invalid destruction for %s %s", inputChain.Output.Type(), chainID)
				}

				continue
			}
			if err := vm.ChainSTVF(vmParams, iotago.ChainTransitionTypeStateChange, inputChain, next); err != nil {
				return ierrors.Wrapf(err, "invalid transition for %s %s", inputChain.Output.Type(), chainID)
			}
		}

		for chainID, outputChain := range vmParams.WorkingSet.OutChains {
			if _, chainPresentInInputs := vmParams.WorkingSet.InChains[chainID]; chainPresentInInputs {
				continue
			}

			if err := vm.ChainSTVF(vmParams, iotago.ChainTransitionTypeGenesis, nil, outputChain); err != nil {
				return ierrors.Wrapf(err, "invalid creation of %s %s", outputChain.Type(), chainID)
			}
		}

		return nil
	}
}

// ExecFuncBalancedNativeTokens validates following rules regarding NativeTokens:
//   - The NativeTokens between Inputs / Outputs must be balanced or have a deficit on the output side if there is no foundry state transition for a given NativeToken.
//   - Max MaxNativeTokensCount native tokens within inputs + outputs
func ExecFuncBalancedNativeTokens() ExecFunc {
	return func(_ VirtualMachine, vmParams *Params) error {
		// native token set creates handle overflows
		var err error
		vmParams.WorkingSet.InNativeTokens, err = vmParams.WorkingSet.UTXOInputs.NativeTokenSum()
		if err != nil {
			return ierrors.WithMessagef(iotago.ErrNativeTokenSetInvalid, "invalid input native token set: %w", err)
		}

		vmParams.WorkingSet.OutNativeTokens, err = vmParams.WorkingSet.Tx.Outputs.NativeTokenSum()
		if err != nil {
			return ierrors.WithMessagef(iotago.ErrNativeTokenSetInvalid, "%w", err)
		}

		// check invariants for when token foundry is absent

		for nativeTokenID, inSum := range vmParams.WorkingSet.InNativeTokens {
			if _, foundryIsTransitioning := vmParams.WorkingSet.OutChains[nativeTokenID]; foundryIsTransitioning {
				continue
			}

			outSum := vmParams.WorkingSet.OutNativeTokens[nativeTokenID]

			if vmParams.WorkingSet.Tx.Capabilities.CannotBurnNativeTokens() && (outSum == nil || inSum.Cmp(outSum) != 0) {
				// if burning is not allowed, the input sum must be equal to the output sum
				return ierrors.WithMessagef(iotago.ErrTxCapabilitiesNativeTokenBurningNotAllowed, "%w: native token %s is less on output (%d) than input (%d) side but burning is not allowed in the transaction and the foundry is absent for melting", iotago.ErrNativeTokenSumUnbalanced, nativeTokenID, outSum, inSum)
			} else if (outSum != nil) && (inSum.Cmp(outSum) == -1) {
				// input sum must be greater equal the output sum (burning allows it to be greater)
				return ierrors.WithMessagef(iotago.ErrNativeTokenSumUnbalanced, "native token %s is less on input (%d) than output (%d) side but the foundry is absent for minting", nativeTokenID, inSum, outSum)
			}
		}

		for nativeTokenID := range vmParams.WorkingSet.OutNativeTokens {
			if _, foundryIsTransitioning := vmParams.WorkingSet.OutChains[nativeTokenID]; foundryIsTransitioning {
				continue
			}

			// foundry must be present when native tokens only reside on the output side
			// as they need to get minted by it within the tx
			if vmParams.WorkingSet.InNativeTokens[nativeTokenID] == nil {
				return ierrors.WithMessagef(iotago.ErrNativeTokenSumUnbalanced, "native token %s is new on the output side but the foundry is not transitioning", nativeTokenID)
			}
		}

		// from here the native tokens balancing is handled by each foundry's STVF

		return nil
	}
}

// Returns a func that checks that no more than one Implicit Account Creation Address
// is on the input side of a transaction.
func ExecFuncAtMostOneImplicitAccountCreationAddress() ExecFunc {
	return func(_ VirtualMachine, vmParams *Params) error {
		transactionHasImplicitAccountCreationAddress := false
		for _, input := range vmParams.WorkingSet.UTXOInputs {
			addressUnlockCondition := input.UnlockConditionSet().Address()
			if input.Type() == iotago.OutputBasic && addressUnlockCondition != nil {
				if addressUnlockCondition.Address.Type() == iotago.AddressImplicitAccountCreation {
					if transactionHasImplicitAccountCreationAddress {
						return iotago.ErrMultipleImplicitAccountCreationAddresses
					}
					transactionHasImplicitAccountCreationAddress = true
				}
			}
		}

		return nil
	}
}
