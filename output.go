package iotago

import (
	"fmt"
	"math"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/hive.go/constraints"
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
	"FoundryOutput",
	"NFTOutput",
	"DelegationOutput",
}

var (
	// ErrTransDepIdentOutputNonUTXOChainID gets returned when a TransDepIdentOutput has a ChainID which is not a UTXOIDChainID.
	ErrTransDepIdentOutputNonUTXOChainID = ierrors.New("transition dependable ident outputs must have UTXO chain IDs")
	// ErrTransDepIdentOutputNextInvalid gets returned when a TransDepIdentOutput's next state is invalid.
	ErrTransDepIdentOutputNextInvalid = ierrors.New("transition dependable ident output's next output is invalid")
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
)

// ChainTransitionError gets returned when a state transition validation fails for a ChainOutput.
type ChainTransitionError struct {
	Inner error
	Msg   string
}

func (i *ChainTransitionError) Error() string {
	var s strings.Builder
	s.WriteString("invalid chain transition")
	if i.Inner != nil {
		s.WriteString(fmt.Sprintf("; inner err: %s", i.Inner))
	}
	if len(i.Msg) > 0 {
		s.WriteString(fmt.Sprintf("; %s", i.Msg))
	}

	return s.String()
}

func (i *ChainTransitionError) Unwrap() error {
	return i.Inner
}

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

// NewAccounts returns an AccountOutputsSet for all AccountOutputs which are new.
func (outputSet OutputSet) NewAccounts() AccountOutputsSet {
	set := make(AccountOutputsSet)
	for utxoInputID, output := range outputSet {
		accountOutput, is := output.(*AccountOutput)
		if !is || !accountOutput.AccountEmpty() {
			continue
		}
		set[AccountIDFromOutputID(utxoInputID)] = accountOutput
	}

	return set
}

// This is a helper function to check if an output is unlockable by a given target.
func outputUnlockableBy(output Output, next TransDepIdentOutput, target Address, pastBoundedSlotIndex SlotIndex, futureBoundedSlotIndex SlotIndex) (bool, error) {
	unlockConds := output.UnlockConditionSet()
	var ownerIdent Address
	switch x := output.(type) {
	case TransIndepIdentOutput:
		ownerIdent = x.Ident()
	case TransDepIdentOutput:
		targetToUnlock, err := x.Ident(next)
		if err != nil {
			return false, err
		}
		ownerIdent = targetToUnlock
	default:
		panic("invalid output type in outputUnlockableBy")
	}

	targetIdentCanUnlock := unlockConds.unlockableBy(target, ownerIdent, pastBoundedSlotIndex, futureBoundedSlotIndex)
	if !targetIdentCanUnlock {
		return false, nil
	}

	return true, nil
}

// TransIndepIdentOutput is a type of Output where the identity to unlock is independent
// of any transition the output does (without considering Feature(s)).
type TransIndepIdentOutput interface {
	Output
	// Ident returns the default identity to which this output is locked to.
	Ident() Address
	// UnlockableBy tells whether the given ident can unlock this Output
	// while also taking into consideration constraints enforced by UnlockConditions(s) within this Output (if any).
	UnlockableBy(ident Address, pastBoundedSlotIndex SlotIndex, futureBoundedSlotIndex SlotIndex) bool
}

// TransDepIdentOutput is a type of Output where the identity to unlock is dependent
// on the transition the output does (without considering UnlockConditions(s)).
type TransDepIdentOutput interface {
	ChainOutput
	// Ident computes the identity to which this output is locked to by examining
	// the transition to the next output state. If next is nil, then this TransDepIdentOutput
	// treats the ident computation as being for ChainTransitionTypeDestroy.
	Ident(next TransDepIdentOutput) (Address, error)
	// UnlockableBy tells whether the given ident can unlock this Output
	// while also taking into consideration constraints enforced by UnlockConditions(s) within this Output
	// and the next state of this TransDepIdentOutput. To indicate that this TransDepIdentOutput
	// is to be destroyed, pass nil as next.
	UnlockableBy(ident Address, next TransDepIdentOutput, pastBoundedSlotIndex SlotIndex, futureBoundedSlotIndex SlotIndex) (bool, error)
}

// OutputsSyntacticalValidationFunc which given the index of an output and the output itself, runs syntactical validations and returns an error if any should fail.
type OutputsSyntacticalValidationFunc func(index int, output Output) error

// OutputsSyntacticalDepositAmount returns an OutputsSyntacticalValidationFunc which checks that:
//   - every output has base token amount more than zero
//   - every output has base token amount less than the total supply
//   - the sum of base token amounts does not exceed the total supply
//   - the base token amount fulfills the minimum storage deposit as calculated from the storage score of the output
//   - if the output contains a StorageDepositReturnUnlockCondition, it must "return" bigger equal than the minimum storage deposit
//     required for the sender to send back the tokens.
func OutputsSyntacticalDepositAmount(protoParams ProtocolParameters, storageScoreStructure *StorageScoreStructure) OutputsSyntacticalValidationFunc {
	var sum BaseToken

	return func(index int, output Output) error {
		amount := output.BaseTokenAmount()

		switch {
		case amount == 0:
			return ierrors.Wrapf(ErrAmountMustBeGreaterThanZero, "output %d", index)
		case amount > protoParams.TokenSupply():
			return ierrors.Wrapf(ErrOutputAmountMoreThanTotalSupply, "output %d", index)
		case sum+amount > protoParams.TokenSupply():
			return ierrors.Wrapf(ErrOutputsSumExceedsTotalSupply, "output %d", index)
		}

		// check whether base token amount fulfills the storage deposit cost
		if _, err := storageScoreStructure.CoversMinDeposit(output, amount); err != nil {
			return ierrors.Wrapf(err, "output %d", index)
		}

		// check whether the amount in the return condition allows the receiver to fulfill the storage deposit for the return output
		if storageDep := output.UnlockConditionSet().StorageDepositReturn(); storageDep != nil {
			minStorageDepositForReturnOutput, err := storageScoreStructure.MinStorageDepositForReturnOutput(storageDep.ReturnAddress)
			if err != nil {
				return ierrors.Wrapf(err, "failed to calculate storage deposit for output index %d", index)
			}
			switch {
			case storageDep.Amount < minStorageDepositForReturnOutput:
				return ierrors.Wrapf(ErrStorageDepositLessThanMinReturnOutputStorageDeposit, "output %d, needed %d, have %d", index, minStorageDepositForReturnOutput, storageDep.Amount)
			case storageDep.Amount > amount:
				return ierrors.Wrapf(ErrStorageDepositExceedsTargetOutputAmount, "output %d, target output's base token amount %d < storage deposit %d", index, amount, storageDep.Amount)
			}
		}

		sum += amount

		return nil
	}
}

// OutputsSyntacticalNativeTokens returns an OutputsSyntacticalValidationFunc which checks that:
//   - each native token holds an amount bigger than zero
func OutputsSyntacticalNativeTokens() OutputsSyntacticalValidationFunc {
	return func(index int, output Output) error {
		nativeToken := output.FeatureSet().NativeToken()
		if nativeToken == nil {
			return nil
		}

		if nativeToken.Amount.Cmp(common.Big0) == 0 {
			return ierrors.Wrapf(ErrNativeTokenAmountLessThanEqualZero, "output %d", index)
		}

		return nil
	}
}

// OutputsSyntacticalExpirationAndTimelock returns an OutputsSyntacticalValidationFunc which checks that:
// That ExpirationUnlockCondition and TimelockUnlockCondition does not have its unix criteria set to zero.
func OutputsSyntacticalExpirationAndTimelock() OutputsSyntacticalValidationFunc {
	return func(index int, output Output) error {
		unlockConditionSet := output.UnlockConditionSet()

		if expiration := unlockConditionSet.Expiration(); expiration != nil {
			if expiration.Slot == 0 {
				return ErrExpirationConditionZero
			}
		}

		if timelock := unlockConditionSet.Timelock(); timelock != nil {
			if timelock.Slot == 0 {
				return ErrTimelockConditionZero
			}
		}

		return nil
	}
}

// OutputsSyntacticalAccount returns an OutputsSyntacticalValidationFunc which checks that AccountOutput(s)':
//   - StateIndex/FoundryCounter are zero if the AccountID is zeroed
//   - StateController and GovernanceController must be different from AccountAddress derived from AccountID
func OutputsSyntacticalAccount() OutputsSyntacticalValidationFunc {
	return func(index int, output Output) error {
		accountOutput, is := output.(*AccountOutput)
		if !is {
			return nil
		}

		if accountOutput.AccountEmpty() {
			switch {
			case accountOutput.StateIndex != 0:
				return ierrors.Wrapf(ErrAccountOutputNonEmptyState, "output %d, state index not zero", index)
			case accountOutput.FoundryCounter != 0:
				return ierrors.Wrapf(ErrAccountOutputNonEmptyState, "output %d, foundry counter not zero", index)
			}
			// can not be cyclic when the AccountOutput is new
			return nil
		}

		outputAccountAddr := AccountAddress(accountOutput.AccountID)
		if stateCtrlAddr, ok := accountOutput.StateController().(*AccountAddress); ok && outputAccountAddr == *stateCtrlAddr {
			return ierrors.Wrapf(ErrAccountOutputCyclicAddress, "output %d, AccountID=StateController", index)
		}
		if govCtrlAddr, ok := accountOutput.GovernorAddress().(*AccountAddress); ok && outputAccountAddr == *govCtrlAddr {
			return ierrors.Wrapf(ErrAccountOutputCyclicAddress, "output %d, AccountID=GovernanceController", index)
		}

		return nil
	}
}

// OutputsSyntacticalFoundry returns an OutputsSyntacticalValidationFunc which checks that FoundryOutput(s)':
//   - Minted and melted supply is less equal MaximumSupply
//   - MaximumSupply is not zero
func OutputsSyntacticalFoundry() OutputsSyntacticalValidationFunc {
	return func(index int, output Output) error {
		foundryOutput, is := output.(*FoundryOutput)
		if !is {
			return nil
		}

		if err := foundryOutput.TokenScheme.SyntacticalValidation(); err != nil {
			return ierrors.Wrapf(err, "output %d", index)
		}

		return nil
	}
}

// OutputsSyntacticalNFT returns an OutputsSyntacticalValidationFunc which checks that NFTOutput(s)':
//   - Address must be different from NFTAddress derived from NFTID
func OutputsSyntacticalNFT() OutputsSyntacticalValidationFunc {
	return func(index int, output Output) error {
		nftOutput, is := output.(*NFTOutput)
		if !is {
			return nil
		}

		if nftOutput.NFTID.Empty() {
			// can not be cyclic when the NFTOutput is new
			return nil
		}

		if addr, ok := nftOutput.Ident().(*NFTAddress); ok && NFTAddress(nftOutput.NFTID) == *addr {
			return ierrors.Wrapf(ErrNFTOutputCyclicAddress, "output %d", index)
		}

		return nil
	}
}

// OutputsSyntacticalDelegation returns an OutputsSyntacticalValidationFunc which checks that DelegationOutput(s)':
//   - Validator ID is not zeroed out.
func OutputsSyntacticalDelegation() OutputsSyntacticalValidationFunc {
	return func(index int, output Output) error {
		delegationOutput, is := output.(*DelegationOutput)
		if !is {
			return nil
		}

		if delegationOutput.ValidatorAddress.AccountID().Empty() {
			return ierrors.Wrapf(ErrDelegationValidatorAddressZeroed, "output %d", index)
		}

		return nil
	}
}

// SyntacticallyValidateOutputs validates the outputs by running them against the given OutputsSyntacticalValidationFunc(s).
func SyntacticallyValidateOutputs(outputs TxEssenceOutputs, funcs ...OutputsSyntacticalValidationFunc) error {
	for i, output := range outputs {
		for _, f := range funcs {
			if err := f(i, output); err != nil {
				return err
			}
		}
	}

	return nil
}

func OutputsSyntacticalChainConstrainedOutputUniqueness() OutputsSyntacticalValidationFunc {
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
			return ierrors.Wrapf(ErrNonUniqueChainOutputs, "output with chainID %s already exist on the output side", chainID.ToHex())
		}

		chainConstrainedOutputs[chainID] = chainConstrainedOutput

		return nil
	}
}
