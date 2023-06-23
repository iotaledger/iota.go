package iotago

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

// Output defines a unit of output of a transaction.
type Output interface {
	Sizer
	NonEphemeralObject

	// Deposit returns the amount this Output deposits.
	Deposit() uint64

	// StoredMana returns the stored mana held by this output.
	StoredMana() uint64

	// NativeTokenList returns the NativeToken this output defines.
	NativeTokenList() NativeTokens

	// UnlockConditionSet returns the UnlockConditionSet this output defines.
	UnlockConditionSet() UnlockConditionSet

	// FeatureSet returns the FeatureSet this output contains.
	FeatureSet() FeatureSet

	// Type returns the type of the output.
	Type() OutputType

	// Clone clones the Output.
	Clone() Output
}

// OutputType defines the type of outputs.
type OutputType byte

const (
	// OutputTreasury denotes the type of the TreasuryOutput.
	OutputTreasury OutputType = 2
	// OutputBasic denotes an BasicOutput.
	OutputBasic OutputType = 3
	// OutputAccount denotes an AccountOutput.
	OutputAccount OutputType = 4
	// OutputFoundry denotes a FoundryOutput.
	OutputFoundry OutputType = 5
	// OutputNFT denotes an NFTOutput.
	OutputNFT OutputType = 6
)

func (outputType OutputType) String() string {
	if int(outputType) >= len(outputNames) {
		return fmt.Sprintf("unknown output type: %d", outputType)
	}
	return outputNames[outputType]
}

var outputNames = [OutputNFT + 1]string{
	"SigLockedSingleOutput",
	"SigLockedDustAllowanceOutput",
	"TreasuryOutput",
	"BasicOutput",
	"AccountOutput",
	"FoundryOutput",
	"NFTOutput",
}

const (
	// OutputIDLength defines the length of an OutputID.
	OutputIDLength = TransactionIDLength + serializer.UInt16ByteSize
)

var (
	ErrInvalidOutputIDLength = errors.New("Invalid outputID length")

	// ErrTransDepIdentOutputNonUTXOChainID gets returned when a TransDepIdentOutput has a ChainID which is not a UTXOIDChainID.
	ErrTransDepIdentOutputNonUTXOChainID = errors.New("transition dependable ident outputs must have UTXO chain IDs")
	// ErrTransDepIdentOutputNextInvalid gets returned when a TransDepIdentOutput's next state is invalid.
	ErrTransDepIdentOutputNextInvalid = errors.New("transition dependable ident output's next output is invalid")
)

// defines the default offset virtual byte costs for an output.
func outputOffsetVByteCost(rentStruct *RentStructure) VBytes {
	return rentStruct.VBFactorKey.Multiply(OutputIDLength) +
		// included block id, conf ms index, conf ms ts
		rentStruct.VBFactorData.Multiply(BlockIDLength+serializer.UInt32ByteSize+serializer.UInt32ByteSize)
}

// OutputID defines the identifier for an UTXO which consists
// out of the referenced TransactionID and the output's index.
type OutputID [OutputIDLength]byte

// EmptyOutputID is an empty OutputID.
var EmptyOutputID = OutputID{}

// ToHex converts the OutputID to its hex representation.
func (outputID OutputID) ToHex() string {
	return hexutil.EncodeHex(outputID[:])
}

// String converts the OutputID to its human-readable string representation.
func (outputID OutputID) String() string {
	return fmt.Sprintf("OutputID(%s:%d)", outputID.TransactionID().String(), outputID.Index())
}

// Index returns the index of the Output this OutputID references.
func (outputID OutputID) Index() uint16 {
	return binary.LittleEndian.Uint16(outputID[TransactionIDLength:])
}

// TransactionID returns the TransactionID of the Output this OutputID references.
func (outputID OutputID) TransactionID() TransactionID {
	var txID TransactionID
	copy(txID[:], outputID[:TransactionIDLength])
	return txID
}

// UTXOInput creates a UTXOInput from this OutputID.
func (outputID OutputID) UTXOInput() *UTXOInput {
	utxoInput := &UTXOInput{}
	copy(utxoInput.TransactionID[:], outputID[:TransactionIDLength])
	utxoInput.TransactionOutputIndex = binary.LittleEndian.Uint16(outputID[TransactionIDLength:])
	return utxoInput
}

func (outputID OutputID) Bytes() ([]byte, error) {
	return outputID[:], nil
}

func (outputID *OutputID) FromBytes(bytes []byte) (int, error) {
	var err error
	*outputID, err = OutputIDFromBytes(bytes)
	if err != nil {
		return 0, err
	}
	return OutputIDLength, nil
}

// HexOutputIDs is a slice of hex encoded OutputID strings.
type HexOutputIDs []string

// MustOutputIDs converts the hex strings into OutputIDs.
func (ids HexOutputIDs) MustOutputIDs() OutputIDs {
	vals, err := ids.OutputIDs()
	if err != nil {
		panic(err)
	}
	return vals
}

// OutputIDs converts the hex strings into OutputIDs.
func (ids HexOutputIDs) OutputIDs() (OutputIDs, error) {
	vals := make(OutputIDs, len(ids))
	for i, v := range ids {
		val, err := hexutil.DecodeHex(v)
		if err != nil {
			return nil, err
		}
		copy(vals[i][:], val)
	}
	return vals, nil
}

// OutputIDFromTransactionIDAndIndex creates a OutputID from the given TransactionID and index.
func OutputIDFromTransactionIDAndIndex(txID TransactionID, index uint16) OutputID {
	utxo := UTXOInput{TransactionOutputIndex: index}
	copy(utxo.TransactionID[:], (txID)[:])

	return utxo.ID()
}

// OutputIDFromBytes creates a OutputID from the given bytes.
func OutputIDFromBytes(bytes []byte) (OutputID, error) {
	if len(bytes) != OutputIDLength {
		return OutputID{}, ErrInvalidOutputIDLength
	}

	return OutputID(bytes), nil
}

// OutputIDFromHex creates a OutputID from the given hex encoded OutputID data.
func OutputIDFromHex(hexStr string) (OutputID, error) {
	outputIDData, err := hexutil.DecodeHex(hexStr)
	if err != nil {
		return OutputID{}, err
	}

	return OutputIDFromBytes(outputIDData)
}

// MustOutputIDFromHex works like OutputIDFromHex but panics if an error is encountered.
func MustOutputIDFromHex(hexStr string) OutputID {
	var outputID OutputID
	outputIDData, err := hexutil.DecodeHex(hexStr)
	if err != nil {
		panic(err)
	}
	copy(outputID[:], outputIDData)
	return outputID
}

// OutputSet is a map of the OutputID to Output.
type OutputSet map[OutputID]Output

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

// OutputIDs is a slice of OutputID.
type OutputIDs []OutputID

// ToHex converts all UTXOInput to their hex string representation.
func (outputIDs OutputIDs) ToHex() []string {
	ids := make([]string, len(outputIDs))
	for i := range outputIDs {
		ids[i] = hexutil.EncodeHex(outputIDs[i][:])
	}
	return ids
}

// RemoveDupsAndSort removes duplicated OutputIDs and sorts the slice by the lexical ordering.
func (outputIDs OutputIDs) RemoveDupsAndSort() OutputIDs {
	sorted := append(OutputIDs{}, outputIDs...)
	sort.Slice(sorted, func(i, j int) bool {
		return bytes.Compare(sorted[i][:], sorted[j][:]) == -1
	})

	var result OutputIDs
	var prev OutputID
	for i, id := range sorted {
		if i == 0 || !bytes.Equal(prev[:], id[:]) {
			result = append(result, id)
		}
		prev = id
	}
	return result
}

// UTXOInputs converts the OutputIDs slice to Inputs.
func (outputIDs OutputIDs) UTXOInputs() TxEssenceInputs {
	inputs := make(TxEssenceInputs, 0)
	for _, outputID := range outputIDs {
		inputs = append(inputs, outputID.UTXOInput())
	}
	return inputs
}

// OrderedSet returns an Outputs slice ordered by this OutputIDs slice given an OutputSet.
func (outputIDs OutputIDs) OrderedSet(set OutputSet) Outputs[Output] {
	outputs := make(Outputs[Output], len(outputIDs))
	for i, outputID := range outputIDs {
		outputs[i] = set[outputID]
	}
	return outputs
}

var (
	// ErrDepositAmountMustBeGreaterThanZero returned if the deposit amount of an output is less or equal zero.
	ErrDepositAmountMustBeGreaterThanZero = errors.New("deposit amount must be greater than zero")
	// ErrChainMissing gets returned when a chain is missing.
	ErrChainMissing = errors.New("chain missing")
	// ErrNonUniqueChainOutputs gets returned when multiple ChainOutputs(s) with the same ChainID exist within sets.
	ErrNonUniqueChainOutputs = errors.New("non unique chain outputs")
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

func (outputs Outputs[T]) Size() int {
	sum := serializer.UInt16ByteSize
	for _, output := range outputs {
		sum += output.Size()
	}
	return sum
}

// MustCommitment works like Commitment but panics if there's an error.
func (outputs Outputs[T]) MustCommitment() []byte {
	comm, err := outputs.Commitment()
	if err != nil {
		panic(err)
	}
	return comm
}

// Commitment computes a hash of the outputs slice to be used as a commitment.
func (outputs Outputs[T]) Commitment() ([]byte, error) {
	h, err := blake2b.New256(nil)
	if err != nil {
		return nil, err
	}
	for _, output := range outputs {
		outputBytes, err := internalEncode(output)
		if err != nil {
			return nil, fmt.Errorf("unable to compute commitment hash: %w", err)
		}

		outputHash := blake2b.Sum256(outputBytes)
		if _, err := h.Write(outputHash[:]); err != nil {
			return nil, fmt.Errorf("unable to write output bytes for commitment hash: %w", err)
		}
	}
	return h.Sum(nil), nil
}

// ChainOutputSet returns a ChainOutputSet for all ChainOutputs in Outputs.
func (outputs Outputs[T]) ChainOutputSet(txID TransactionID) ChainOutputSet {
	set := make(ChainOutputSet)
	for outputIndex, output := range outputs {
		chainOutput, is := Output(output).(ChainOutput)
		if !is {
			continue
		}

		chainID := chainOutput.Chain()
		if chainID.Empty() {
			if utxoIDChainID, is := chainOutput.Chain().(UTXOIDChainID); is {
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

// ToOutputsByType converts the Outputs slice to OutputsByType.
func (outputs Outputs[T]) ToOutputsByType() OutputsByType {
	outputsByType := make(OutputsByType)
	for _, output := range outputs {
		slice, has := outputsByType[output.Type()]
		if !has {
			slice = make([]Output, 0)
		}
		outputsByType[output.Type()] = append(slice, output)
	}
	return outputsByType
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
		nativeTokens := output.NativeTokenList()
		for _, nativeToken := range nativeTokens {
			if sign := nativeToken.Amount.Sign(); sign == -1 || sign == 0 {
				return nil, ErrNativeTokenAmountLessThanEqualZero
			}

			val := sum[nativeToken.ID]
			if val == nil {
				val = new(big.Int)
			}

			if val.Add(val, nativeToken.Amount).Cmp(abi.MaxUint256) == 1 {
				return nil, ErrNativeTokenSumExceedsUint256
			}
			sum[nativeToken.ID] = val
		}
	}
	return sum, nil
}

// OutputsByType is a map of OutputType(s) to slice of Output(s).
type OutputsByType map[OutputType][]Output

// BasicOutputs returns a slice of Outputs which are BasicOutput.
func (outputs OutputsByType) BasicOutputs() BasicOutputs {
	extOutputs := make(BasicOutputs, 0)
	for _, output := range outputs[OutputBasic] {
		extOutput, is := output.(*BasicOutput)
		if !is {
			continue
		}
		extOutputs = append(extOutputs, extOutput)
	}
	return extOutputs
}

// FoundryOutputs returns a slice of Outputs which are FoundryOutput.
func (outputs OutputsByType) FoundryOutputs() FoundryOutputs {
	foundryOutputs := make(FoundryOutputs, 0)
	for _, output := range outputs[OutputFoundry] {
		foundryOutput, is := output.(*FoundryOutput)
		if !is {
			continue
		}
		foundryOutputs = append(foundryOutputs, foundryOutput)
	}
	return foundryOutputs
}

// FoundryOutputsSet returns a map of FoundryID to FoundryOutput.
// If multiple FoundryOutput(s) exist for a given FoundryID, an error is returned.
func (outputs OutputsByType) FoundryOutputsSet() (FoundryOutputsSet, error) {
	foundryOutputsSet := make(FoundryOutputsSet)
	for _, output := range outputs[OutputFoundry] {
		foundryOutput, is := output.(*FoundryOutput)
		if !is {
			continue
		}
		foundryID, err := foundryOutput.ID()
		if err != nil {
			return nil, err
		}
		if _, has := foundryOutputsSet[foundryID]; has {
			return nil, ErrNonUniqueFoundryOutputs
		}
		foundryOutputsSet[foundryID] = foundryOutput
	}
	return foundryOutputsSet, nil
}

// AccountOutputs returns a slice of Outputs which are AccountOutput.
func (outputs OutputsByType) AccountOutputs() AccountOutputs {
	accountOutputs := make(AccountOutputs, 0)
	for _, output := range outputs[OutputAccount] {
		accountOutput, is := output.(*AccountOutput)
		if !is {
			continue
		}
		accountOutputs = append(accountOutputs, accountOutput)
	}
	return accountOutputs
}

// NonNewAccountOutputsSet returns a map of AccountID to AccountOutput.
// If multiple AccountOutput(s) exist for a given AccountID, an error is returned.
// The produced set does not include AccountOutputs of which their AccountID are zeroed.
func (outputs OutputsByType) NonNewAccountOutputsSet() (AccountOutputsSet, error) {
	accountOutputsSet := make(AccountOutputsSet)
	for _, output := range outputs[OutputAccount] {
		accountOutput, is := output.(*AccountOutput)
		if !is || accountOutput.AccountEmpty() {
			continue
		}
		if _, has := accountOutputsSet[accountOutput.AccountID]; has {
			return nil, ErrNonUniqueAccountOutputs
		}
		accountOutputsSet[accountOutput.AccountID] = accountOutput
	}
	return accountOutputsSet, nil
}

// ChainOutputSet returns a map of ChainID to ChainOutput.
// If multiple ChainOutput(s) exist for a given ChainID, an error is returned.
func (outputs OutputsByType) ChainOutputSet() (ChainOutputSet, error) {
	chainOutputSet := make(ChainOutputSet)
	for _, ty := range []OutputType{OutputAccount, OutputFoundry, OutputNFT} {
		for _, output := range outputs[ty] {
			chainOutput, is := output.(ChainOutput)
			if !is || chainOutput.Chain().Empty() {
				continue
			}
			if _, has := chainOutputSet[chainOutput.Chain()]; has {
				return nil, ErrNonUniqueChainOutputs
			}
			chainOutputSet[chainOutput.Chain()] = chainOutput
		}
	}
	return chainOutputSet, nil
}

// ChainOutputs returns a slice of Outputs which are ChainOutput.
func (outputs OutputsByType) ChainOutputs() ChainOutputs {
	chainOutputs := make(ChainOutputs, 0)
	for _, ty := range []OutputType{OutputAccount, OutputFoundry, OutputNFT} {
		for _, output := range outputs[ty] {
			chainOutput, is := output.(ChainOutput)
			if !is {
				continue
			}
			chainOutputs = append(chainOutputs, chainOutput)
		}
	}
	return chainOutputs
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

func outputUnlockable(output Output, next TransDepIdentOutput, target Address, txCreationTime SlotIndex) (bool, error) {
	unlockConds := output.UnlockConditionSet()

	checkTargetIdentOfOutput := func() (bool, error) {
		switch x := output.(type) {
		case TransIndepIdentOutput:
			return x.Ident().Equal(target), nil
		case TransDepIdentOutput:
			targetToUnlock, err := x.Ident(next)
			if err != nil {
				return false, err
			}
			return targetToUnlock.Equal(target), nil
		default:
			panic("invalid output type in outputUnlockable")
		}
	}

	if len(unlockConds) == 0 {
		return checkTargetIdentOfOutput()
	}

	targetIdentCanUnlock, returnIdentCanUnlock := unlockConds.unlockableBy(target, txCreationTime)
	if !targetIdentCanUnlock {
		return false, nil
	}

	// the target ident is the return ident which can unlock
	if returnIdentCanUnlock {
		return true, nil
	}

	return checkTargetIdentOfOutput()
}

// ExternalUnlockParameters defines a palette of external system parameters which are used to
// determine whether an Output can be unlocked.
type ExternalUnlockParameters struct {
	DecayProvider      *ManaDecayProvider
	ProtocolParameters *ProtocolParameters
}

// TransIndepIdentOutput is a type of Output where the identity to unlock is independent
// of any transition the output does (without considering Feature(s)).
type TransIndepIdentOutput interface {
	Output
	// Ident returns the default identity to which this output is locked to.
	Ident() Address
	// UnlockableBy tells whether the given ident can unlock this Output
	// while also taking into consideration constraints enforced by UnlockConditions(s) within this Output (if any).
	UnlockableBy(ident Address, txCreationTime SlotIndex) bool
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
	UnlockableBy(ident Address, next TransDepIdentOutput, txCreationTime SlotIndex) (bool, error)
}

// OutputIDHex is the hex representation of an output ID.
type OutputIDHex string

// MustSplitParts returns the transaction ID and output index parts of the hex output ID.
// It panics if the hex output ID is invalid.
func (oih OutputIDHex) MustSplitParts() (*TransactionID, uint16) {
	txID, outputIndex, err := oih.SplitParts()
	if err != nil {
		panic(err)
	}
	return txID, outputIndex
}

// SplitParts returns the transaction ID and output index parts of the hex output ID.
func (oih OutputIDHex) SplitParts() (*TransactionID, uint16, error) {
	outputIDBytes, err := hexutil.DecodeHex(string(oih))
	if err != nil {
		return nil, 0, err
	}
	var txID TransactionID
	copy(txID[:], outputIDBytes[:TransactionIDLength])
	outputIndex := binary.LittleEndian.Uint16(outputIDBytes[TransactionIDLength : TransactionIDLength+serializer.UInt16ByteSize])
	return &txID, outputIndex, nil
}

// MustAsUTXOInput converts the hex output ID to a UTXOInput.
// It panics if the hex output ID is invalid.
func (oih OutputIDHex) MustAsUTXOInput() *UTXOInput {
	utxoInput, err := oih.AsUTXOInput()
	if err != nil {
		panic(err)
	}
	return utxoInput
}

// AsUTXOInput converts the hex output ID to a UTXOInput.
func (oih OutputIDHex) AsUTXOInput() (*UTXOInput, error) {
	var utxoInput UTXOInput
	txID, outputIndex, err := oih.SplitParts()
	if err != nil {
		return nil, err
	}
	copy(utxoInput.TransactionID[:], txID[:])
	utxoInput.TransactionOutputIndex = outputIndex
	return &utxoInput, nil
}

// OutputsSyntacticalValidationFunc which given the index of an output and the output itself, runs syntactical validations and returns an error if any should fail.
type OutputsSyntacticalValidationFunc func(index int, output Output) error

// OutputsSyntacticalDepositAmount returns an OutputsSyntacticalValidationFunc which checks that:
//   - every output deposits more than zero
//   - every output deposits less than the total supply
//   - the sum of deposits does not exceed the total supply
//   - the deposit fulfills the minimum storage deposit as calculated from the virtual byte cost of the output
//   - if the output contains a StorageDepositReturnUnlockCondition, it must "return" bigger equal than the minimum storage deposit
//     required for the sender to send back the tokens.
func OutputsSyntacticalDepositAmount(protoParams *ProtocolParameters) OutputsSyntacticalValidationFunc {
	var sum uint64
	return func(index int, output Output) error {
		deposit := output.Deposit()

		switch {
		case deposit == 0:
			return fmt.Errorf("%w: output %d", ErrDepositAmountMustBeGreaterThanZero, index)
		case deposit > protoParams.TokenSupply:
			return fmt.Errorf("%w: output %d", ErrOutputDepositsMoreThanTotalSupply, index)
		case sum+deposit > protoParams.TokenSupply:
			return fmt.Errorf("%w: output %d", ErrOutputsSumExceedsTotalSupply, index)
		}

		// check whether deposit fulfills the storage deposit cost
		if _, err := protoParams.RentStructure.CoversStateRent(output, deposit); err != nil {
			return fmt.Errorf("%w: output %d", err, index)
		}

		// check whether the amount in the return condition allows the receiver to fulfill the storage deposit for the return output
		if storageDep := output.UnlockConditionSet().StorageDepositReturn(); storageDep != nil {
			minStorageDepositForReturnOutput := protoParams.RentStructure.MinStorageDepositForReturnOutput(storageDep.ReturnAddress)
			switch {
			case storageDep.Amount < minStorageDepositForReturnOutput:
				return fmt.Errorf("%w: output %d, needed %d, have %d", ErrStorageDepositLessThanMinReturnOutputStorageDeposit, index, minStorageDepositForReturnOutput, storageDep.Amount)
			case storageDep.Amount > deposit:
				return fmt.Errorf("%w: output %d, target output's deposit %d < storage deposit %d", ErrStorageDepositExceedsTargetOutputDeposit, index, deposit, storageDep.Amount)
			}
		}

		sum += deposit
		return nil
	}
}

// OutputsSyntacticalNativeTokens returns an OutputsSyntacticalValidationFunc which checks that:
//   - the sum of native tokens count across all outputs does not exceed MaxNativeTokensCount
//   - each native token holds an amount bigger than zero
func OutputsSyntacticalNativeTokens() OutputsSyntacticalValidationFunc {
	distinctNativeTokens := make(map[NativeTokenID]struct{})
	return func(index int, output Output) error {
		nativeTokens := output.NativeTokenList()

		for i, nt := range nativeTokens {
			distinctNativeTokens[nt.ID] = struct{}{}
			if len(distinctNativeTokens) > MaxNativeTokensCount {
				return ErrMaxNativeTokensCountExceeded
			}
			if nt.Amount.Cmp(common.Big0) == 0 {
				return fmt.Errorf("%w: output %d, native token index %d", ErrNativeTokenAmountLessThanEqualZero, index, i)
			}
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
			if expiration.SlotIndex == 0 {
				return ErrExpirationConditionZero
			}
		}

		if timelock := unlockConditionSet.Timelock(); timelock != nil {
			if timelock.SlotIndex == 0 {
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
				return fmt.Errorf("%w: output %d, state index not zero", ErrAccountOutputNonEmptyState, index)
			case accountOutput.FoundryCounter != 0:
				return fmt.Errorf("%w: output %d, foundry counter not zero", ErrAccountOutputNonEmptyState, index)
			}
			// can not be cyclic when the AccountOutput is new
			return nil
		}

		outputAccountAddr := AccountAddress(accountOutput.AccountID)
		if stateCtrlAddr, ok := accountOutput.StateController().(*AccountAddress); ok && outputAccountAddr == *stateCtrlAddr {
			return fmt.Errorf("%w: output %d, AccountID=StateController", ErrAccountOutputCyclicAddress, index)
		}
		if govCtrlAddr, ok := accountOutput.GovernorAddress().(*AccountAddress); ok && outputAccountAddr == *govCtrlAddr {
			return fmt.Errorf("%w: output %d, AccountID=GovernanceController", ErrAccountOutputCyclicAddress, index)
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
			return fmt.Errorf("%w: output %d", err, index)
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
			return fmt.Errorf("%w: output %d", ErrNFTOutputCyclicAddress, index)
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

		chainID := chainConstrainedOutput.Chain()
		if chainID.Empty() {
			// we can ignore newly minted chainConstrainedOutputs
			return nil
		}

		if _, has := chainConstrainedOutputs[chainID]; has {
			return fmt.Errorf("%w: output with chainID %s already exist on the output side", ErrNonUniqueChainOutputs, chainID.ToHex())
		}

		chainConstrainedOutputs[chainID] = chainConstrainedOutput
		return nil
	}
}
