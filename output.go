package iotago

import (
	"fmt"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/hive.go/constraints"
	"github.com/iotaledger/hive.go/core/safemath"
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/serializer/v2"
)

// BaseToken defines the unit of the base token of the network.
type BaseToken uint64

// BaseTokenSize is the size in bytes that is used by BaseToken.
const BaseTokenSize = 8

const MaxBaseToken = BaseToken(math.MaxUint64)

// Mana defines the type of the consumable resource e.g. used in congestion control.
type Mana uint64

// ManaSize is the size in bytes that is used by Mana.
const ManaSize = 8

const MaxMana = Mana(math.MaxUint64)

// The maximum a metadata map may have (excluding the type byte of the feature).
const MaxMetadataMapSize = 8192

// Output defines a unit of output of a transaction.
type Output interface {
	Sizer
	NonEphemeralObject
	ProcessableObject
	constraints.Cloneable[Output]
	constraints.Equalable[Output]

	// BaseTokenAmount returns the amount of base tokens held by this Output.
	BaseTokenAmount() BaseToken

	// StoredMana returns the stored mana held by this output.
	StoredMana() Mana

	// UnlockConditionSet returns the UnlockConditionSet this output defines.
	UnlockConditionSet() UnlockConditionSet

	// FeatureSet returns the FeatureSet this output contains.
	FeatureSet() FeatureSet

	// Type returns the type of the output.
	Type() OutputType
}

// OutputType defines the type of outputs.
type OutputType byte

const (
	// OutputBasic denotes an BasicOutput.
	OutputBasic OutputType = iota
	// OutputAccount denotes an AccountOutput.
	OutputAccount
	// OutputAnchor denotes an AnchorOuptut.
	OutputAnchor
	// OutputFoundry denotes a FoundryOutput.
	OutputFoundry
	// OutputNFT denotes an NFTOutput.
	OutputNFT
	// OutputDelegation denotes a DelegationOutput.
	OutputDelegation
)

func (outputType OutputType) String() string {
	if int(outputType) >= len(outputNames) {
		return fmt.Sprintf("unknown output type: %d", outputType)
	}

	return outputNames[outputType]
}

var outputNames = [OutputDelegation + 1]string{
	"BasicOutput",
	"AccountOutput",
	"AnchorOutput",
	"FoundryOutput",
	"NFTOutput",
	"DelegationOutput",
}

var (
	// ErrOwnerTransitionDependentOutputNonUTXOChainID gets returned when a OwnerTransitionDependentOutput has a ChainID which is not a UTXOIDChainID.
	ErrOwnerTransitionDependentOutputNonUTXOChainID = ierrors.New("owner transition dependent outputs must have UTXO chain IDs")
	// ErrOwnerTransitionDependentOutputNextInvalid gets returned when a OwnerTransitionDependentOutput's next state is invalid.
	ErrOwnerTransitionDependentOutputNextInvalid = ierrors.New("owner transition dependent output's next output is invalid")
	// ErrArrayValidationOrderViolatesLexicalOrder gets returned if the array elements are not in lexical order.
	ErrArrayValidationOrderViolatesLexicalOrder = ierrors.New("array elements must be in their lexical order")
	// ErrArrayValidationViolatesUniqueness gets returned if the array elements are not unique.
	ErrArrayValidationViolatesUniqueness = ierrors.New("array elements must be unique")
)

// OutputSet is a map of the OutputID to Output.
type OutputSet map[OutputID]Output

// Clone clones the OutputSet.
func (outputSet OutputSet) Clone() OutputSet {
	return lo.CloneMap(outputSet)
}

// Filter creates a new OutputSet with Outputs which pass the filter function f.
func (outputSet OutputSet) Filter(f func(outputID OutputID, output Output) bool) OutputSet {
	m := make(OutputSet)
	for id, output := range outputSet {
		if f(id, output) {
			m[id] = output
		}
	}

	return m
}

var (
	// ErrAmountMustBeGreaterThanZero gets returned if the base token amount of an output is less or equal zero.
	ErrAmountMustBeGreaterThanZero = ierrors.New("base token amount must be greater than zero")
	// ErrChainMissing gets returned when a chain is missing.
	ErrChainMissing = ierrors.New("chain missing")
	// ErrNonUniqueChainOutputs gets returned when multiple ChainOutputs(s) with the same ChainID exist within sets.
	ErrNonUniqueChainOutputs = ierrors.New("non unique chain outputs")
	// ErrNewChainOutputHasNonZeroedID gets returned when a new chain output has a non-zeroed ID.
	ErrNewChainOutputHasNonZeroedID = ierrors.New("new chain output has non-zeroed ID")
	// ErrChainOutputImmutableFeaturesChanged gets returned when a chain output's immutable features are modified in a transition.
	ErrChainOutputImmutableFeaturesChanged = ierrors.New("immutable features in chain output modified during transition")
)

// Outputs is a slice of Output.
type Outputs[T Output] []T

func (outputs Outputs[T]) Clone() Outputs[T] {
	cpy := make(Outputs[T], len(outputs))
	for idx, output := range outputs {
		//nolint:forcetypeassert // we can safely assume that this is of type T
		cpy[idx] = output.Clone().(T)
	}

	return cpy
}

func (outputs Outputs[T]) Size() int {
	sum := serializer.UInt16ByteSize
	for _, output := range outputs {
		sum += output.Size()
	}

	return sum
}

func (outputs Outputs[T]) WorkScore(workScoreParameters *WorkScoreParameters) (WorkScore, error) {
	var workScoreOutputs WorkScore
	for _, output := range outputs {
		workScoreOutput, err := output.WorkScore(workScoreParameters)
		if err != nil {
			return 0, err
		}

		workScoreOutputs, err = workScoreOutputs.Add(workScoreOutput)
		if err != nil {
			return 0, err
		}
	}

	return workScoreOutputs, nil
}

// ChainOutputSet returns a ChainOutputSet for all ChainOutputs in Outputs.
func (outputs Outputs[T]) ChainOutputSet(txID TransactionID) ChainOutputSet {
	set := make(ChainOutputSet)
	for outputIndex, output := range outputs {
		chainOutput, is := Output(output).(ChainOutput)
		if !is {
			continue
		}

		chainID := chainOutput.ChainID()
		if chainID.Empty() {
			if utxoIDChainID, is := chainOutput.ChainID().(UTXOIDChainID); is {
				chainID = utxoIDChainID.FromOutputID(OutputIDFromTransactionIDAndIndex(txID, uint16(outputIndex)))
			}
		}

		if chainID.Empty() {
			panic(fmt.Sprintf("output of type %s has empty chain ID but is not utxo dependable", output.Type()))
		}

		set[chainID] = chainOutput
	}

	return set
}

// OutputsFilterFunc is a predicate function operating on an Output.
type OutputsFilterFunc func(output Output) bool

// OutputsFilterByType is an OutputsFilterFunc which filters Outputs by OutputType.
func OutputsFilterByType(ty OutputType) OutputsFilterFunc {
	return func(output Output) bool { return output.Type() == ty }
}

// Filter returns Outputs (retained order) passing the given OutputsFilterFunc.
func (outputs Outputs[T]) Filter(f OutputsFilterFunc) Outputs[T] {
	filtered := make(Outputs[T], 0)
	for _, output := range outputs {
		if !f(output) {
			continue
		}
		filtered = append(filtered, output)
	}

	return filtered
}

// NativeTokenSum sums up the different NativeTokens occurring within the given outputs.
func (outputs Outputs[T]) NativeTokenSum() (NativeTokenSum, error) {
	sum := make(map[NativeTokenID]*big.Int)
	for _, output := range outputs {
		nativeTokenFeature := output.FeatureSet().NativeToken()
		if nativeTokenFeature == nil {
			continue
		}

		if sign := nativeTokenFeature.Amount.Sign(); sign == -1 || sign == 0 {
			return nil, ErrNativeTokenAmountLessThanEqualZero
		}

		val := sum[nativeTokenFeature.ID]
		if val == nil {
			val = new(big.Int)
		}

		if val.Add(val, nativeTokenFeature.Amount).Cmp(abi.MaxUint256) == 1 {
			return nil, ErrNativeTokenSumExceedsUint256
		}
		sum[nativeTokenFeature.ID] = val
	}

	return sum, nil
}

// This is a helper function to check if an output is unlockable by a given target.
func outputUnlockableBy(output Output, next OwnerTransitionDependentOutput, target Address, pastBoundedSlotIndex SlotIndex, futureBoundedSlotIndex SlotIndex) (bool, error) {
	unlockConds := output.UnlockConditionSet()
	var owner Address
	switch x := output.(type) {
	case OwnerTransitionIndependentOutput:
		owner = x.Owner()
	case OwnerTransitionDependentOutput:
		targetToUnlock, err := x.Owner(next)
		if err != nil {
			return false, err
		}
		owner = targetToUnlock
	default:
		panic("invalid output type in outputUnlockableBy")
	}

	targetAddrCanUnlock := unlockConds.unlockableBy(target, owner, pastBoundedSlotIndex, futureBoundedSlotIndex)
	if !targetAddrCanUnlock {
		return false, nil
	}

	return true, nil
}

// Computes the Potential Mana that the output generates between creationSlot and targetSlot,
// while deducting the minimum deposit of the output which does not generate Mana.
//
// Returns 0 if the output does not have the minimum storage deposit covered.
func PotentialMana(manaDecayProvider *ManaDecayProvider, storageScoreStructure *StorageScoreStructure, output Output, creationSlot, targetSlot SlotIndex) (Mana, error) {
	minDeposit, err := storageScoreStructure.MinDeposit(output)
	if err != nil {
		return 0, ierrors.Wrap(err, "failed to calculate min deposit for potential mana calculation")
	}

	excessBaseTokens, err := safemath.SafeSub(output.BaseTokenAmount(), minDeposit)
	if err != nil {
		// nolint:nilerr // An underflow means no potential mana is generated and hence no error is returned.
		return 0, nil
	}

	return manaDecayProvider.GenerateManaAndDecayBySlots(excessBaseTokens, creationSlot, targetSlot)
}

// OwnerTransitionIndependentOutput is a type of Output where the address to unlock is independent
// of any transition the output does (without considering Feature(s)).
type OwnerTransitionIndependentOutput interface {
	Output
	// Owner returns the default address to which this output is locked to.
	Owner() Address
	// UnlockableBy tells whether the given address can unlock this Output
	// while also taking into consideration constraints enforced by UnlockConditions(s) within this Output (if any).
	UnlockableBy(addr Address, pastBoundedSlotIndex SlotIndex, futureBoundedSlotIndex SlotIndex) bool
}

// OwnerTransitionDependentOutput is a type of Output where the address to unlock is dependent
// on the transition the output does (without considering UnlockConditions(s)).
type OwnerTransitionDependentOutput interface {
	ChainOutput
	// Owner computes the address to which this output is locked to by examining
	// the transition to the next output state. If next is nil, then this OwnerTransitionDependentOutput
	// treats the owner computation as being for ChainTransitionTypeDestroy.
	Owner(next OwnerTransitionDependentOutput) (Address, error)
	// UnlockableBy tells whether the given address can unlock this Output
	// while also taking into consideration constraints enforced by UnlockConditions(s) within this Output
	// and the next state of this OwnerTransitionDependentOutput. To indicate that this OwnerTransitionDependentOutput
	// is to be destroyed, pass nil as next.
	UnlockableBy(addr Address, next OwnerTransitionDependentOutput, pastBoundedSlotIndex SlotIndex, futureBoundedSlotIndex SlotIndex) (bool, error)
}

// OutputsSyntacticalDepositAmount returns an ElementValidationFunc[Output] which checks that:
//   - every output has base token amount more than zero
//   - the sum of base token amounts does not exceed the total supply
//   - the base token amount fulfills the minimum storage deposit as calculated from the storage score of the output
//   - if the output contains a StorageDepositReturnUnlockCondition, it must "return" bigger equal than the minimum storage deposit
//     required for the sender to send back the tokens.
func OutputsSyntacticalDepositAmount(protoParams ProtocolParameters, storageScoreStructure *StorageScoreStructure) ElementValidationFunc[Output] {
	var sum BaseToken

	return func(index int, output Output) error {
		amount := output.BaseTokenAmount()

		if amount == 0 {
			return ierrors.WithMessagef(ErrAmountMustBeGreaterThanZero, "output %d", index)
		}

		var err error
		sum, err = safemath.SafeAdd(sum, amount)
		if err != nil {
			return ierrors.Join(ErrOutputsSumExceedsTotalSupply, ierrors.WithMessagef(err, "output %d", index))
		}
		if sum > protoParams.TokenSupply() {
			return ierrors.WithMessagef(ErrOutputsSumExceedsTotalSupply, "output %d", index)
		}

		// check whether base token amount fulfills the storage deposit cost
		if _, err := storageScoreStructure.CoversMinDeposit(output, amount); err != nil {
			return ierrors.WithMessagef(err, "output %d", index)
		}

		// check whether the amount in the return condition allows the receiver to fulfill the storage deposit for the return output
		if storageDep := output.UnlockConditionSet().StorageDepositReturn(); storageDep != nil {
			minStorageDepositForReturnOutput, err := storageScoreStructure.MinStorageDepositForReturnOutput(storageDep.ReturnAddress)
			if err != nil {
				return ierrors.WithMessagef(err, "failed to calculate storage deposit for output index %d", index)
			}
			switch {
			case storageDep.Amount < minStorageDepositForReturnOutput:
				return ierrors.WithMessagef(ErrStorageDepositLessThanMinReturnOutputStorageDeposit, "output %d, needed %d, have %d", index, minStorageDepositForReturnOutput, storageDep.Amount)
			case storageDep.Amount > amount:
				return ierrors.WithMessagef(ErrStorageDepositExceedsTargetOutputAmount, "output %d, target output's base token amount %d < storage deposit %d", index, amount, storageDep.Amount)
			}
		}

		return nil
	}
}

// OutputsSyntacticalNativeTokens returns an ElementValidationFunc[Output] which checks that:
//   - each native token holds an amount bigger than zero
func OutputsSyntacticalNativeTokens() ElementValidationFunc[Output] {
	return func(index int, output Output) error {
		nativeToken := output.FeatureSet().NativeToken()
		if nativeToken == nil {
			return nil
		}

		if nativeToken.Amount.Cmp(common.Big0) == 0 {
			return ierrors.WithMessagef(ErrNativeTokenAmountLessThanEqualZero, "output %d", index)
		}

		return nil
	}
}

// OutputsSyntacticalStoredMana returns an ElementValidationFunc[Output] which checks that:
//   - the sum of all stored mana fields does not exceed 2^(Mana Bits Count) - 1.
func OutputsSyntacticalStoredMana(maxManaValue Mana) ElementValidationFunc[Output] {
	var sum Mana

	return func(index int, output Output) error {
		storedMana := output.StoredMana()

		var err error
		sum, err = safemath.SafeAdd(sum, storedMana)
		if err != nil {
			return ierrors.Join(ierrors.Wrapf(ErrMaxManaExceeded, "stored mana sum calculation failed at output %d", index), err)
		}

		if sum > maxManaValue {
			return ierrors.WithMessagef(ErrMaxManaExceeded, "sum of stored mana exceeds max value with output %d", index)
		}

		return nil
	}
}

// OutputsSyntacticalExpirationAndTimelock returns an ElementValidationFunc[Output] which checks that:
// That ExpirationUnlockCondition and TimelockUnlockCondition does not have its unix criteria set to zero.
func OutputsSyntacticalExpirationAndTimelock() ElementValidationFunc[Output] {
	return func(index int, output Output) error {
		unlockConditionSet := output.UnlockConditionSet()

		if expiration := unlockConditionSet.Expiration(); expiration != nil {
			if expiration.Slot == 0 {
				return ierrors.WithMessagef(ErrExpirationConditionZero, "output %d", index)
			}
		}

		if timelock := unlockConditionSet.Timelock(); timelock != nil {
			if timelock.Slot == 0 {
				return ierrors.WithMessagef(ErrTimelockConditionZero, "output %d", index)
			}
		}

		return nil
	}
}

// OutputsSyntacticalAccount returns an ElementValidationFunc[Output] which checks that AccountOutput(s)':
//   - FoundryCounter is zero if the AccountID is zeroed
//   - Address must be different from AccountAddress derived from AccountID
//   - Amount must be greater than or equal to StakedAmount of staking feature if it is present
func OutputsSyntacticalAccount() ElementValidationFunc[Output] {
	return func(index int, output Output) error {
		accountOutput, is := output.(*AccountOutput)
		if !is {
			return nil
		}

		if accountOutput.AccountID.Empty() {
			if accountOutput.FoundryCounter != 0 {
				return ierrors.WithMessagef(ErrAccountOutputNonEmptyState, "output %d, foundry counter not zero", index)
			}
		}

		if addr, ok := accountOutput.Owner().(*AccountAddress); ok && AccountAddress(accountOutput.AccountID) == *addr {
			return ierrors.WithMessagef(ErrAccountOutputCyclicAddress, "output %d", index)
		}

		accountFeatures := accountOutput.FeatureSet()
		if stakingFeat := accountFeatures.Staking(); stakingFeat != nil {
			if accountOutput.Amount < stakingFeat.StakedAmount {
				return ierrors.WithMessagef(ErrAccountOutputAmountLessThanStakedAmount, "output %d", index)
			}

			if accountFeatures.BlockIssuer() == nil {
				return ierrors.WithMessagef(ErrStakingBlockIssuerFeatureMissing, "output %d", index)
			}
		}

		return nil
	}
}

// OutputsSyntacticalAnchor returns an ElementValidationFunc[Output] which checks that AnchorOutput(s)':
//   - StateIndex is zero if the AnchorID is zeroed
//   - StateController and GovernanceController must be different from AnchorAddress derived from AnchorID
func OutputsSyntacticalAnchor() ElementValidationFunc[Output] {
	return func(index int, output Output) error {
		anchorOutput, is := output.(*AnchorOutput)
		if !is {
			return nil
		}

		if anchorOutput.AnchorID.Empty() {
			if anchorOutput.StateIndex != 0 {
				return ierrors.WithMessagef(ErrAnchorOutputNonEmptyState, "output %d, state index not zero", index)
			}

			// can not be cyclic when the AnchorOutput is new
			return nil
		}

		outputAnchorAddr := AnchorAddress(anchorOutput.AnchorID)
		if stateCtrlAddr, ok := anchorOutput.StateController().(*AnchorAddress); ok && outputAnchorAddr == *stateCtrlAddr {
			return ierrors.WithMessagef(ErrAnchorOutputCyclicAddress, "output %d, AnchorID=StateController", index)
		}
		if govCtrlAddr, ok := anchorOutput.GovernorAddress().(*AnchorAddress); ok && outputAnchorAddr == *govCtrlAddr {
			return ierrors.WithMessagef(ErrAnchorOutputCyclicAddress, "output %d, AnchorID=GovernanceController", index)
		}

		return nil
	}
}

// OutputsSyntacticalFoundry returns an ElementValidationFunc[Output] which checks that FoundryOutput(s)':
//   - Minted and melted supply is less equal MaximumSupply
//   - MaximumSupply is not zero
func OutputsSyntacticalFoundry() ElementValidationFunc[Output] {
	return func(index int, output Output) error {
		foundryOutput, is := output.(*FoundryOutput)
		if !is {
			return nil
		}

		if err := foundryOutput.TokenScheme.SyntacticalValidation(); err != nil {
			return ierrors.WithMessagef(err, "output %d", index)
		}

		nativeTokenFeature := foundryOutput.FeatureSet().NativeToken()
		if nativeTokenFeature == nil {
			return nil
		}

		foundryID, err := foundryOutput.FoundryID()
		if err != nil {
			return err
		}

		// NativeTokenFeature ID should have the same ID as the foundry
		if !foundryID.Matches(nativeTokenFeature.ID) {
			return ierrors.WithMessagef(ErrFoundryIDNativeTokenIDMismatch, "output %d, FoundryID: %s, NativeTokenID: %s", index, foundryID, nativeTokenFeature.ID)
		}

		return nil
	}
}

// OutputsSyntacticalNFT returns an ElementValidationFunc[Output] which checks that NFTOutput(s)':
//   - Address must be different from NFTAddress derived from NFTID
func OutputsSyntacticalNFT() ElementValidationFunc[Output] {
	return func(index int, output Output) error {
		nftOutput, is := output.(*NFTOutput)
		if !is {
			return nil
		}

		if nftOutput.NFTID.Empty() {
			// can not be cyclic when the NFTOutput is new
			return nil
		}

		if addr, ok := nftOutput.Owner().(*NFTAddress); ok && NFTAddress(nftOutput.NFTID) == *addr {
			return ierrors.WithMessagef(ErrNFTOutputCyclicAddress, "output %d", index)
		}

		return nil
	}
}

// OutputsSyntacticalDelegation returns an ElementValidationFunc[Output] which checks that DelegationOutput(s)':
//   - Validator ID is not zeroed out.
func OutputsSyntacticalDelegation() ElementValidationFunc[Output] {
	return func(index int, output Output) error {
		delegationOutput, is := output.(*DelegationOutput)
		if !is {
			return nil
		}

		if delegationOutput.ValidatorAddress.AccountID().Empty() {
			return ierrors.WithMessagef(ErrDelegationValidatorAddressEmpty, "output %d", index)
		}

		return nil
	}
}

func checkAddressRestrictions(output TxEssenceOutput, address Address) error {
	addrWithCapabilities, isAddrWithCapabilities := address.(AddressCapabilities)
	if !isAddrWithCapabilities {
		// no restrictions
		return nil
	}

	if addrWithCapabilities.CannotReceiveNativeTokens() && output.FeatureSet().HasNativeTokenFeature() {
		return ErrAddressCannotReceiveNativeTokens
	}

	if addrWithCapabilities.CannotReceiveMana() && output.StoredMana() != 0 {
		return ErrAddressCannotReceiveMana
	}

	if addrWithCapabilities.CannotReceiveOutputsWithTimelockUnlockCondition() && output.UnlockConditionSet().HasTimelockCondition() {
		return ErrAddressCannotReceiveTimelockUnlockCondition
	}

	if addrWithCapabilities.CannotReceiveOutputsWithExpirationUnlockCondition() && output.UnlockConditionSet().HasExpirationCondition() {
		return ErrAddressCannotReceiveExpirationUnlockCondition
	}

	if addrWithCapabilities.CannotReceiveOutputsWithStorageDepositReturnUnlockCondition() && output.UnlockConditionSet().HasStorageDepositReturnCondition() {
		return ErrAddressCannotReceiveStorageDepositReturnUnlockCondition
	}

	if addrWithCapabilities.CannotReceiveAccountOutputs() && output.Type() == OutputAccount {
		return ErrAddressCannotReceiveAccountOutput
	}

	if addrWithCapabilities.CannotReceiveAnchorOutputs() && output.Type() == OutputAnchor {
		return ErrAddressCannotReceiveAnchorOutput
	}

	if addrWithCapabilities.CannotReceiveNFTOutputs() && output.Type() == OutputNFT {
		return ErrAddressCannotReceiveNFTOutput
	}

	if addrWithCapabilities.CannotReceiveDelegationOutputs() && output.Type() == OutputDelegation {
		return ErrAddressCannotReceiveDelegationOutput
	}

	return nil
}

// OutputsSyntacticalAddressRestrictions returns a func that checks the capability flag restrictions on addresses.
//
// Does not validate the Return Address in StorageDepositReturnUnlockCondition because such a Return Address
// already is as restricted as the most restricted address.
func OutputsSyntacticalAddressRestrictions() ElementValidationFunc[Output] {
	return func(index int, output Output) error {
		if addressUnlockCondition := output.UnlockConditionSet().Address(); addressUnlockCondition != nil {
			if err := checkAddressRestrictions(output, addressUnlockCondition.Address); err != nil {
				return ierrors.WithMessagef(err, "output %d", index)
			}
		}
		if stateControllerUnlockCondition := output.UnlockConditionSet().StateControllerAddress(); stateControllerUnlockCondition != nil {
			if err := checkAddressRestrictions(output, stateControllerUnlockCondition.Address); err != nil {
				return ierrors.WithMessagef(err, "output %d", index)
			}
		}
		if governorUnlockCondition := output.UnlockConditionSet().GovernorAddress(); governorUnlockCondition != nil {
			if err := checkAddressRestrictions(output, governorUnlockCondition.Address); err != nil {
				return ierrors.WithMessagef(err, "output %d", index)
			}
		}
		if expirationUnlockCondition := output.UnlockConditionSet().Expiration(); expirationUnlockCondition != nil {
			if err := checkAddressRestrictions(output, expirationUnlockCondition.ReturnAddress); err != nil {
				return ierrors.WithMessagef(err, "output %d", index)
			}
		}

		return nil
	}
}

func OutputsSyntacticalImplicitAccountCreationAddress() ElementValidationFunc[Output] {
	return func(index int, output Output) error {
		switch typedOutput := output.(type) {
		case *BasicOutput, *FoundryOutput:
			// - Implicit Account Creation Addresses are allowed in Basic Outputs.
			// - Foundry Outputs cannot contain non-Account Addresses in the first place,
			// so they don't have to be checked.
			return nil
		case *AccountOutput, *NFTOutput, *DelegationOutput:
			// The serialization rules enforce that these output types always have an address unlock condition set.
			if output.UnlockConditionSet().Address().Address.Type() == AddressImplicitAccountCreation {
				return ierrors.WithMessagef(ErrImplicitAccountCreationAddressInInvalidOutput, "output %d", index)
			}
		case *AnchorOutput:
			// The serialization rules enforce that these addresses are always set.
			stateControllerAddress := typedOutput.UnlockConditions.MustSet().StateControllerAddress().Address
			governorAddress := typedOutput.UnlockConditions.MustSet().GovernorAddress().Address

			if (stateControllerAddress.Type() == AddressImplicitAccountCreation) ||
				(governorAddress.Type() == AddressImplicitAccountCreation) {
				return ierrors.WithMessagef(ErrImplicitAccountCreationAddressInInvalidOutput, "output %d", index)
			}
		default:
			// We're switching on the Go output type here, so we can only run into the default case
			// if we added a new output type and have not handled it above or a user constructed a type
			// implementing the interface (only possible when iota.go is used as a library).
			// In both cases we want to panic.
			panic("all supported output types should be handled above")
		}

		return nil
	}
}

// Checks lexical order and uniqueness of the output's unlock conditions.
func OutputsSyntacticalUnlockConditionLexicalOrderAndUniqueness() ElementValidationFunc[Output] {
	return func(index int, output Output) error {
		lexicalOrderUniquenessValidator := LexicalOrderAndUniquenessValidator[UnlockCondition]()
		switch typedOutput := output.(type) {
		case *BasicOutput:
			for idx, uc := range typedOutput.UnlockConditions {
				if err := lexicalOrderUniquenessValidator(idx, uc); err != nil {
					return ierrors.WithMessagef(err, "output %d, unlock condition index: %d", index, idx)
				}
			}
		case *AccountOutput:
			for idx, uc := range typedOutput.UnlockConditions {
				if err := lexicalOrderUniquenessValidator(idx, uc); err != nil {
					return ierrors.WithMessagef(err, "output %d, unlock condition index: %d", index, idx)
				}
			}
		case *AnchorOutput:
			for idx, uc := range typedOutput.UnlockConditions {
				if err := lexicalOrderUniquenessValidator(idx, uc); err != nil {
					return ierrors.WithMessagef(err, "output %d, unlock condition index: %d", index, idx)
				}
			}
		case *FoundryOutput:
			for idx, uc := range typedOutput.UnlockConditions {
				if err := lexicalOrderUniquenessValidator(idx, uc); err != nil {
					return ierrors.WithMessagef(err, "output %d, unlock condition index: %d", index, idx)
				}
			}
		case *NFTOutput:
			for idx, uc := range typedOutput.UnlockConditions {
				if err := lexicalOrderUniquenessValidator(idx, uc); err != nil {
					return ierrors.WithMessagef(err, "output %d, unlock condition index: %d", index, idx)
				}
			}
		case *DelegationOutput:
			for idx, uc := range typedOutput.UnlockConditions {
				if err := lexicalOrderUniquenessValidator(idx, uc); err != nil {
					return ierrors.WithMessagef(err, "output %d, unlock condition index: %d", index, idx)
				}
			}
		default:
			// We're switching on the Go output type here, so we can only run into the default case
			// if we added a new output type and have not handled it above or a user constructed a type
			// implementing the interface (only possible when iota.go is used as a library).
			// In both cases we want to panic.
			panic("all supported output types should be handled above")
		}

		return nil
	}
}

// Checks lexical order and uniqueness of the output's features and immutable features.
func OutputsSyntacticalFeaturesLexicalOrderAndUniqueness() ElementValidationFunc[Output] {
	return func(index int, output Output) error {
		featureValidationFunc := LexicalOrderAndUniquenessValidator[Feature]()
		immutableFeatureValidationFunc := LexicalOrderAndUniquenessValidator[Feature]()

		switch typedOutput := output.(type) {
		case *BasicOutput:
			for idx, uc := range typedOutput.Features {
				if err := featureValidationFunc(idx, uc); err != nil {
					return ierrors.WithMessagef(err, "output %d, feature index: %d", index, idx)
				}
			}
			// This output does not have immutable features.
		case *AccountOutput:
			for idx, uc := range typedOutput.Features {
				if err := featureValidationFunc(idx, uc); err != nil {
					return ierrors.WithMessagef(err, "output %d, feature index: %d", index, idx)
				}
			}
			for idx, uc := range typedOutput.ImmutableFeatures {
				if err := immutableFeatureValidationFunc(idx, uc); err != nil {
					return ierrors.WithMessagef(err, "output %d, immutable feature index: %d", index, idx)
				}
			}
		case *AnchorOutput:
			for idx, uc := range typedOutput.Features {
				if err := featureValidationFunc(idx, uc); err != nil {
					return ierrors.WithMessagef(err, "output %d, feature index: %d", index, idx)
				}
			}
			for idx, uc := range typedOutput.ImmutableFeatures {
				if err := immutableFeatureValidationFunc(idx, uc); err != nil {
					return ierrors.WithMessagef(err, "output %d, immutable feature index: %d", index, idx)
				}
			}
		case *FoundryOutput:
			for idx, uc := range typedOutput.Features {
				if err := featureValidationFunc(idx, uc); err != nil {
					return ierrors.WithMessagef(err, "output %d, feature index: %d", index, idx)
				}
			}
			for idx, uc := range typedOutput.ImmutableFeatures {
				if err := immutableFeatureValidationFunc(idx, uc); err != nil {
					return ierrors.WithMessagef(err, "output %d, immutable feature index: %d", index, idx)
				}
			}
		case *NFTOutput:
			for idx, uc := range typedOutput.Features {
				if err := featureValidationFunc(idx, uc); err != nil {
					return ierrors.WithMessagef(err, "output %d, feature index: %d", index, idx)
				}
			}
			for idx, uc := range typedOutput.ImmutableFeatures {
				if err := immutableFeatureValidationFunc(idx, uc); err != nil {
					return ierrors.WithMessagef(err, "output %d, immutable feature index: %d", index, idx)
				}
			}
		case *DelegationOutput:
			// This output does not have features.
			return nil
		default:
			// We're switching on the Go output type here, so we can only run into the default case
			// if we added a new output type and have not handled it above or a user constructed a type
			// implementing the interface (only possible when iota.go is used as a library).
			// In both cases we want to panic.
			panic("all supported output types should be handled above")
		}

		return nil
	}
}

// SyntacticallyValidateOutputs validates the outputs by running them against the given ElementValidationFunc(s).
func SyntacticallyValidateOutputs(outputs TxEssenceOutputs, funcs ...ElementValidationFunc[Output]) error {
	for i, output := range outputs {
		for _, f := range funcs {
			if err := f(i, output); err != nil {
				return err
			}
		}
	}

	return nil
}

// Checks that a chain-constrained output with a certain ChainID is unique on the output side.
func OutputsSyntacticalChainConstrainedOutputUniqueness() ElementValidationFunc[Output] {
	chainConstrainedOutputs := make(ChainOutputSet)

	return func(index int, output Output) error {
		chainConstrainedOutput, is := output.(ChainOutput)
		if !is {
			return nil
		}

		chainID := chainConstrainedOutput.ChainID()
		if chainID.Empty() {
			// we can ignore newly minted chainConstrainedOutputs
			return nil
		}

		if _, has := chainConstrainedOutputs[chainID]; has {
			return ierrors.WithMessagef(ErrNonUniqueChainOutputs, "output %d with chainID %s already exist on the output side", index, chainID.ToHex())
		}

		chainConstrainedOutputs[chainID] = chainConstrainedOutput

		return nil
	}
}

// Checks that the (state) metadata feature in outputs do not exceed the max allowed size.
func OutputsSyntacticalMetadataFeatureMaxSize() ElementValidationFunc[Output] {
	checkMaxSize := func(index int, featType FeatureType, mapSize int) error {
		if mapSize > MaxMetadataMapSize {
			return ierrors.WithMessagef(ErrMetadataExceedsMaxSize,
				"the %s of the output at index %d has size %d; max allowed: %d",
				featType, index, mapSize, MaxMetadataMapSize,
			)
		}

		return nil
	}

	return func(index int, output Output) error {
		stateMetadataFeat := output.FeatureSet().StateMetadata()
		if stateMetadataFeat != nil {
			mapSize := stateMetadataFeat.mapSize()
			if err := checkMaxSize(index, stateMetadataFeat.Type(), mapSize); err != nil {
				return ierrors.WithMessagef(err, "output %d", index)
			}
		}

		metadataFeat := output.FeatureSet().Metadata()
		if metadataFeat != nil {
			mapSize := metadataFeat.mapSize()
			if err := checkMaxSize(index, metadataFeat.Type(), mapSize); err != nil {
				return ierrors.WithMessagef(err, "output %d", index)
			}
		}

		return nil
	}
}

// Checks that a Commitment Input is present for
//   - Accounts with a Staking Feature.
//   - Accounts with a Block Issuer Feature.
//   - Delegation Outputs.
func OutputsSyntacticalCommitmentInput(hasCommitmentInput bool) ElementValidationFunc[Output] {
	return func(index int, output Output) error {
		hasStakingFeature := output.FeatureSet().Staking() != nil
		if hasStakingFeature && !hasCommitmentInput {
			return ierrors.WithMessagef(ErrStakingCommitmentInputMissing, "output %d", index)
		}

		hasBlockIssuerFeature := output.FeatureSet().BlockIssuer() != nil
		if hasBlockIssuerFeature && !hasCommitmentInput {
			return ierrors.WithMessagef(ErrBlockIssuerCommitmentInputMissing, "output %d", index)
		}

		if output.Type() == OutputDelegation && !hasCommitmentInput {
			return ierrors.WithMessagef(ErrDelegationCommitmentInputMissing, "output %d", index)
		}

		return nil
	}
}
