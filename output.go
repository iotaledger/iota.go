package iotago

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/iotaledger/hive.go/serializer/v2"
	"golang.org/x/crypto/blake2b"
)

const (
	// OutputIDLength defines the length of an OutputID.
	OutputIDLength = TransactionIDLength + serializer.UInt16ByteSize
)

var (
	// ErrTransDepIdentOutputNonUTXOChainID gets returned when a TransDepIdentOutput has a ChainID which is not a UTXOIDChainID.
	ErrTransDepIdentOutputNonUTXOChainID = errors.New("transition dependable ident outputs must have UTXO chain IDs")
	// ErrTransDepIdentOutputNextInvalid gets returned when a TransDepIdentOutput's next state is invalid.
	ErrTransDepIdentOutputNextInvalid = errors.New("transition dependable ident output's next output is invalid")
)

// defines the default offset virtual byte costs for an output.
func outputOffsetVByteCost(rentStruct *RentStructure) uint64 {
	return rentStruct.VBFactorKey.Multiply(OutputIDLength) +
		// included block id, conf ms index, conf ms ts
		rentStruct.VBFactorData.Multiply(BlockIDLength+serializer.UInt32ByteSize+serializer.UInt32ByteSize)
}

// OutputID defines the identifier for an UTXO which consists
// out of the referenced TransactionID and the output's index.
type OutputID [OutputIDLength]byte

// ToHex converts the OutputID to its hex representation.
func (outputID OutputID) ToHex() string {
	return EncodeHex(outputID[:])
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
		val, err := DecodeHex(v)
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

// OutputIDFromHex creates a OutputID from the given hex encoded OututID data.
func OutputIDFromHex(hexStr string) (OutputID, error) {
	var outputID OutputID
	outputIDData, err := DecodeHex(hexStr)
	if err != nil {
		return outputID, err
	}
	copy(outputID[:], outputIDData)
	return outputID, nil
}

// MustOutputIDFromHex works like OutputIDFromHex but panics if an error is encountered.
func MustOutputIDFromHex(hexStr string) (OutputID, error) {
	var outputID OutputID
	outputIDData, err := DecodeHex(hexStr)
	if err != nil {
		return outputID, err
	}
	copy(outputID[:], outputIDData)
	return outputID, nil
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
		ids[i] = EncodeHex(outputIDs[i][:])
	}
	return ids
}

// UTXOInputs converts the OutputIDs slice to Inputs.
func (outputIDs OutputIDs) UTXOInputs() Inputs {
	inputs := make(Inputs, 0)
	for _, outputID := range outputIDs {
		inputs = append(inputs, outputID.UTXOInput())
	}
	return inputs
}

// OrderedSet returns an Outputs slice ordered by this OutputIDs slice given a OutputSet.
func (outputIDs OutputIDs) OrderedSet(set OutputSet) Outputs {
	outputs := make(Outputs, len(outputIDs))
	for i, outputID := range outputIDs {
		outputs[i] = set[outputID]
	}
	return outputs
}

// OutputType defines the type of outputs.
type OutputType byte

const (
	// OutputTreasury denotes the type of the TreasuryOutput.
	OutputTreasury OutputType = 2
	// OutputBasic denotes an BasicOutput.
	OutputBasic OutputType = 3
	// OutputAlias denotes an AliasOutput.
	OutputAlias OutputType = 4
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

var (
	outputNames = [OutputNFT + 1]string{
		"LegacyOutput",
		"TreasuryOutput",
		"BasicOutput",
		"AliasOutput",
		"FoundryOutput",
		"NFTOutput",
	}
)

var (
	// ErrDepositAmountMustBeGreaterThanZero returned if the deposit amount of an output is less or equal zero.
	ErrDepositAmountMustBeGreaterThanZero = errors.New("deposit amount must be greater than zero")
	// ErrChainMissing gets returned when a chain is missing.
	ErrChainMissing = errors.New("chain missing")
	// ErrNonUniqueChainConstrainedOutputs gets returned when multiple ChainConstrainedOutputs(s) with the same ChainID exist within sets.
	ErrNonUniqueChainConstrainedOutputs = errors.New("non unique chain constrained outputs")
	// ErrTypeIsNotSupportedOutput gets returned when a serializable was found to not be a supported Output.
	ErrTypeIsNotSupportedOutput = errors.New("serializable is not a supported output")
)

// ChainTransitionError gets returned when a state transition validation fails for a ChainConstrainedOutput.
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
type Outputs []Output

func (outputs Outputs) ToSerializables() serializer.Serializables {
	seris := make(serializer.Serializables, len(outputs))
	for i, x := range outputs {
		seris[i] = x.(serializer.Serializable)
	}
	return seris
}

func (outputs *Outputs) FromSerializables(seris serializer.Serializables) {
	*outputs = make(Outputs, len(seris))
	for i, seri := range seris {
		(*outputs)[i] = seri.(Output)
	}
}

func (outputs Outputs) Size() int {
	sum := serializer.UInt16ByteSize
	for _, output := range outputs {
		sum += output.Size()
	}
	return sum
}

// MustCommitment works like Commitment but panics if there's an error.
func (outputs Outputs) MustCommitment() []byte {
	comm, err := outputs.Commitment()
	if err != nil {
		panic(err)
	}
	return comm
}

// Commitment computes a hash of the outputs slice to be used as a commitment.
func (outputs Outputs) Commitment() ([]byte, error) {
	h, err := blake2b.New256(nil)
	if err != nil {
		return nil, err
	}
	for _, output := range outputs {
		outputBytes, err := output.Serialize(serializer.DeSeriModeNoValidation, nil)
		if err != nil {
			return nil, fmt.Errorf("unable to compute commitment hash: %w", err)
		}
		if _, err := h.Write(outputBytes); err != nil {
			return nil, fmt.Errorf("unable to write output bytes for commitment hash: %w", err)
		}
	}
	return h.Sum(nil), nil
}

// ChainConstrainedOutputSet returns a ChainConstrainedOutputsSet for all ChainConstrainedOutputs in Outputs.
func (outputs Outputs) ChainConstrainedOutputSet(txID TransactionID) ChainConstrainedOutputsSet {
	set := make(ChainConstrainedOutputsSet)
	for outputIndex, output := range outputs {
		chainConstrainedOutput, is := output.(ChainConstrainedOutput)
		if !is {
			continue
		}

		chainID := chainConstrainedOutput.Chain()
		if chainID.Empty() {
			if utxoIDChainID, is := chainConstrainedOutput.Chain().(UTXOIDChainID); is {
				chainID = utxoIDChainID.FromOutputID(OutputIDFromTransactionIDAndIndex(txID, uint16(outputIndex)))
			}
		}

		if chainID.Empty() {
			panic(fmt.Sprintf("output of type %s has empty chain ID but is not utxo dependable", output.Type()))
		}

		set[chainID] = chainConstrainedOutput
	}
	return set
}

// ToOutputsByType converts the Outputs slice to OutputsByType.
func (outputs Outputs) ToOutputsByType() OutputsByType {
	outputsByType := make(OutputsByType)
	for _, output := range outputs {
		slice, has := outputsByType[output.Type()]
		if !has {
			slice = make(Outputs, 0)
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
func (outputs Outputs) Filter(f OutputsFilterFunc) Outputs {
	filtered := make(Outputs, 0)
	for _, output := range outputs {
		if !f(output) {
			continue
		}
		filtered = append(filtered, output)
	}
	return filtered
}

// NativeTokenSum sums up the different NativeTokens occurring within the given outputs.
// limit defines the max amount of native tokens which are allowed
func (outputs Outputs) NativeTokenSum() (NativeTokenSum, int, error) {
	sum := make(map[NativeTokenID]*big.Int)
	var ntCount int
	for _, output := range outputs {
		set := output.NativeTokenSet()
		ntCount += len(set)
		for _, nativeToken := range set {
			if sign := nativeToken.Amount.Sign(); sign == -1 || sign == 0 {
				return nil, 0, ErrNativeTokenAmountLessThanEqualZero
			}

			val := sum[nativeToken.ID]
			if val == nil {
				val = new(big.Int)
			}

			if val.Add(val, nativeToken.Amount).Cmp(abi.MaxUint256) == 1 {
				return nil, 0, ErrNativeTokenSumExceedsUint256
			}
			sum[nativeToken.ID] = val
		}
	}
	return sum, ntCount, nil
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

// AliasOutputs returns a slice of Outputs which are AliasOutput.
func (outputs OutputsByType) AliasOutputs() AliasOutputs {
	aliasOutputs := make(AliasOutputs, 0)
	for _, output := range outputs[OutputAlias] {
		aliasOutput, is := output.(*AliasOutput)
		if !is {
			continue
		}
		aliasOutputs = append(aliasOutputs, aliasOutput)
	}
	return aliasOutputs
}

// NonNewAliasOutputsSet returns a map of AliasID to AliasOutput.
// If multiple AliasOutput(s) exist for a given AliasID, an error is returned.
// The produced set does not include AliasOutputs of which their AliasID are zeroed.
func (outputs OutputsByType) NonNewAliasOutputsSet() (AliasOutputsSet, error) {
	aliasOutputsSet := make(AliasOutputsSet)
	for _, output := range outputs[OutputAlias] {
		aliasOutput, is := output.(*AliasOutput)
		if !is || aliasOutput.AliasEmpty() {
			continue
		}
		if _, has := aliasOutputsSet[aliasOutput.AliasID]; has {
			return nil, ErrNonUniqueAliasOutputs
		}
		aliasOutputsSet[aliasOutput.AliasID] = aliasOutput
	}
	return aliasOutputsSet, nil
}

// ChainConstrainedOutputsSet returns a map of ChainID to ChainConstrainedOutput.
// If multiple ChainConstrainedOutput(s) exist for a given ChainID, an error is returned.
func (outputs OutputsByType) ChainConstrainedOutputsSet() (ChainConstrainedOutputsSet, error) {
	chainConstrainedOutputs := make(ChainConstrainedOutputsSet)
	for _, ty := range []OutputType{OutputAlias, OutputFoundry, OutputNFT} {
		for _, output := range outputs[ty] {
			chainConstrainedOutput, is := output.(ChainConstrainedOutput)
			if !is || chainConstrainedOutput.Chain().Empty() {
				continue
			}
			if _, has := chainConstrainedOutputs[chainConstrainedOutput.Chain()]; has {
				return nil, ErrNonUniqueChainConstrainedOutputs
			}
			chainConstrainedOutputs[chainConstrainedOutput.Chain()] = chainConstrainedOutput
		}
	}
	return chainConstrainedOutputs, nil
}

// ChainConstrainedOutputs returns a slice of Outputs which are ChainConstrainedOutput.
func (outputs OutputsByType) ChainConstrainedOutputs() ChainConstrainedOutputs {
	chainConstrainedOutputs := make(ChainConstrainedOutputs, 0)
	for _, ty := range []OutputType{OutputAlias, OutputFoundry, OutputNFT} {
		for _, output := range outputs[ty] {
			chainConstrainedOutput, is := output.(ChainConstrainedOutput)
			if !is {
				continue
			}
			chainConstrainedOutputs = append(chainConstrainedOutputs, chainConstrainedOutput)
		}
	}
	return chainConstrainedOutputs
}

// NewAliases returns an AliasOutputsSet for all AliasOutputs which are new.
func (outputSet OutputSet) NewAliases() AliasOutputsSet {
	set := make(AliasOutputsSet)
	for utxoInputID, output := range outputSet {
		aliasOutput, is := output.(*AliasOutput)
		if !is || !aliasOutput.AliasEmpty() {
			continue
		}
		set[AliasIDFromOutputID(utxoInputID)] = aliasOutput
	}
	return set
}

// ChainConstrainedOutputSet returns a ChainConstrainedOutputsSet for all ChainConstrainedOutputs in the OutputSet.
func (outputSet OutputSet) ChainConstrainedOutputSet() ChainConstrainedOutputsSet {
	set := make(ChainConstrainedOutputsSet)
	for utxoInputID, output := range outputSet {
		chainConstrainedOutput, is := output.(ChainConstrainedOutput)
		if !is {
			continue
		}

		chainID := chainConstrainedOutput.Chain()
		if chainID.Empty() {
			if utxoIDChainID, is := chainConstrainedOutput.Chain().(UTXOIDChainID); is {
				chainID = utxoIDChainID.FromOutputID(utxoInputID)
			}
		}

		if chainID.Empty() {
			panic(fmt.Sprintf("output of type %s has empty chain ID but is not utxo dependable", output.Type()))
		}

		set[chainID] = chainConstrainedOutput
	}
	return set
}

func outputUnlockable(output Output, next TransDepIdentOutput, target Address, extParas *ExternalUnlockParameters) (bool, error) {
	unlockConds := output.UnlockConditions()

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

	if unlockConds == nil {
		return checkTargetIdentOfOutput()
	}

	targetIdentCanUnlock, returnIdentCanUnlock := unlockConds.MustSet().unlockableBy(target, extParas)
	if !targetIdentCanUnlock {
		return false, nil
	}

	// the target ident is the return ident which can unlock
	if returnIdentCanUnlock {
		return true, nil
	}

	return checkTargetIdentOfOutput()
}

// Output defines a unit of output of a transaction.
type Output interface {
	serializer.SerializableWithSize
	NonEphemeralObject

	// Deposit returns the amount this Output deposits.
	Deposit() uint64

	// NativeTokenSet returns the NativeToken this output defines.
	NativeTokenSet() NativeTokens

	// UnlockConditions returns the UnlockConditions this output defines.
	UnlockConditions() UnlockConditions

	// Features returns the Features this output contains.
	Features() Features

	// Type returns the type of the output.
	Type() OutputType

	// Clone clones the Output.
	Clone() Output
}

// ExternalUnlockParameters defines a palette of external system parameters which are used to
// determine whether an Output can be unlocked.
type ExternalUnlockParameters struct {
	// The confirmed milestone index.
	ConfMsIndex uint32
	// The confirmed unix epoch time in seconds.
	ConfUnix uint32
}

// TransIndepIdentOutput is a type of Output where the identity to unlock is independent
// of any transition the output does (without considering Feature(s)).
type TransIndepIdentOutput interface {
	Output
	// Ident returns the default identity to which this output is locked to.
	Ident() Address
	// UnlockableBy tells whether the given ident can unlock this Output
	// while also taking into consideration constraints enforced by UnlockConditions(s) within this Output (if any).
	UnlockableBy(ident Address, extParas *ExternalUnlockParameters) bool
}

// TransDepIdentOutput is a type of Output where the identity to unlock is dependent
// on the transition the output does (without considering UnlockConditions(s)).
type TransDepIdentOutput interface {
	ChainConstrainedOutput
	// Ident computes the identity to which this output is locked to by examining
	// the transition to the next output state. If next is nil, then this TransDepIdentOutput
	// treats the ident computation as being for ChainTransitionTypeDestroy.
	Ident(next TransDepIdentOutput) (Address, error)
	// UnlockableBy tells whether the given ident can unlock this Output
	// while also taking into consideration constraints enforced by UnlockConditions(s) within this Output
	// and the next state of this TransDepIdentOutput. To indicate that this TransDepIdentOutput
	// is to be destroyed, pass nil as next.
	UnlockableBy(ident Address, next TransDepIdentOutput, extParas *ExternalUnlockParameters) (bool, error)
}

// OutputSelector implements SerializableSelectorFunc for output types.
func OutputSelector(outputType uint32) (Output, error) {
	var seri Output
	switch OutputType(outputType) {
	case OutputBasic:
		seri = &BasicOutput{}
	case OutputTreasury:
		seri = &TreasuryOutput{}
	case OutputAlias:
		seri = &AliasOutput{}
	case OutputFoundry:
		seri = &FoundryOutput{}
	case OutputNFT:
		seri = &NFTOutput{}
	default:
		return nil, fmt.Errorf("%w: type %d", ErrUnknownOutputType, outputType)
	}
	return seri, nil
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
	outputIDBytes, err := DecodeHex(string(oih))
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
//	- every output deposits more than zero
//	- every output deposits less than the total supply
//	- the sum of deposits does not exceed the total supply
//	- the deposit fulfils the minimum storage deposit as calculated from the virtual byte cost of the output
//	- if the output contains a StorageDepositReturnUnlockCondition, it must "return" bigger equal than the minimum storage deposit
//	  required for the sender to send back the tokens.
func OutputsSyntacticalDepositAmount(protoParas *ProtocolParameters) OutputsSyntacticalValidationFunc {
	var sum uint64
	return func(index int, output Output) error {
		deposit := output.Deposit()

		switch {
		case deposit == 0:
			return fmt.Errorf("%w: output %d", ErrDepositAmountMustBeGreaterThanZero, index)
		case deposit > protoParas.TokenSupply:
			return fmt.Errorf("%w: output %d", ErrOutputDepositsMoreThanTotalSupply, index)
		case sum+deposit > protoParas.TokenSupply:
			return fmt.Errorf("%w: output %d", ErrOutputsSumExceedsTotalSupply, index)
		}

		// check whether deposit fulfils the storage deposit cost
		if _, err := protoParas.RentStructure.CoversStateRent(output, deposit); err != nil {
			return fmt.Errorf("%w: output %d", err, index)
		}

		unlockConditionsSet, err := output.UnlockConditions().Set()
		if err != nil {
			return fmt.Errorf("unable to compute unlock conditions set in deposit syntactic checks for output %d: %w", index, err)
		}

		// check whether the amount in the return condition allows the receiver to fulfil the storage deposit for the return output
		if storageDep := unlockConditionsSet.StorageDepositReturn(); storageDep != nil {
			minStorageDepositForReturnOutput := protoParas.RentStructure.MinStorageDepositForReturnOutput(storageDep.ReturnAddress)
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
//	- the sum of native tokens count across all outputs does not exceed MaxNativeTokensCount
//	- each native token holds an amount bigger than zero
func OutputsSyntacticalNativeTokens() OutputsSyntacticalValidationFunc {
	var nativeTokensCount int
	return func(index int, output Output) error {
		set := output.NativeTokenSet()
		nativeTokensCount += len(set)
		if nativeTokensCount > MaxNativeTokensCount {
			return ErrMaxNativeTokensCountExceeded
		}

		for i, nt := range set {
			if nt.Amount.Cmp(common.Big0) == 0 {
				return fmt.Errorf("%w: output %d, native token index %d", ErrNativeTokenAmountLessThanEqualZero, index, i)
			}
		}
		return nil
	}
}

// OutputsSyntacticalExpirationAndTimelock returns an OutputsSyntacticalValidationFunc which checks that:
// That ExpirationUnlockCondition and TimelockUnlockCondition does not have both of its milestone and unix criteria set to zero.
func OutputsSyntacticalExpirationAndTimelock() OutputsSyntacticalValidationFunc {
	return func(index int, output Output) error {
		unlockConditionsSet, err := output.UnlockConditions().Set()
		if err != nil {
			return fmt.Errorf("unable to compute unlock conditions set in expiration/timelock syntactic checks for output %d: %w", index, err)
		}

		if expiration := unlockConditionsSet.Expiration(); expiration != nil {
			if expiration.MilestoneIndex == 0 && expiration.UnixTime == 0 {
				return ErrExpirationConditionsZero
			}
		}

		if timelock := unlockConditionsSet.Timelock(); timelock != nil {
			if timelock.MilestoneIndex == 0 && timelock.UnixTime == 0 {
				return ErrTimelockConditionsZero
			}
		}

		return nil
	}
}

// OutputsSyntacticalAlias returns an OutputsSyntacticalValidationFunc which checks that AliasOutput(s)':
//	- StateIndex/FoundryCounter are zero if the AliasID is zeroed
//	- StateController and GovernanceController must be different from AliasAddress derived from AliasID
func OutputsSyntacticalAlias() OutputsSyntacticalValidationFunc {
	return func(index int, output Output) error {
		aliasOutput, is := output.(*AliasOutput)
		if !is {
			return nil
		}

		if aliasOutput.AliasEmpty() {
			switch {
			case aliasOutput.StateIndex != 0:
				return fmt.Errorf("%w: output %d, state index not zero", ErrAliasOutputNonEmptyState, index)
			case aliasOutput.FoundryCounter != 0:
				return fmt.Errorf("%w: output %d, foundry counter not zero", ErrAliasOutputNonEmptyState, index)
			}
			// can not be cyclic when the AliasOutput is new
			return nil
		}

		outputAliasAddr := AliasAddress(aliasOutput.AliasID)
		if stateCtrlAddr, ok := aliasOutput.StateController().(*AliasAddress); ok && outputAliasAddr == *stateCtrlAddr {
			return fmt.Errorf("%w: output %d, AliasID=StateController", ErrAliasOutputCyclicAddress, index)
		}
		if govCtrlAddr, ok := aliasOutput.GovernorAddress().(*AliasAddress); ok && outputAliasAddr == *govCtrlAddr {
			return fmt.Errorf("%w: output %d, AliasID=GovernanceController", ErrAliasOutputCyclicAddress, index)
		}

		return nil
	}
}

// OutputsSyntacticalFoundry returns an OutputsSyntacticalValidationFunc which checks that FoundryOutput(s)':
//	- Minted and melted supply is less equal MaximumSupply
//	- MaximumSupply is not zero
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
//	- Address must be different from NFTAddress derived from NFTID
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
func SyntacticallyValidateOutputs(outputs Outputs, funcs ...OutputsSyntacticalValidationFunc) error {
	for i, output := range outputs {
		for _, f := range funcs {
			if err := f(i, output); err != nil {
				return err
			}
		}
	}
	return nil
}

// JsonOutputSelector selects the json output implementation for the given type.
func JsonOutputSelector(ty int) (JSONSerializable, error) {
	var obj JSONSerializable
	switch OutputType(ty) {
	case OutputBasic:
		obj = &jsonExtendedOutput{}
	case OutputTreasury:
		obj = &jsonTreasuryOutput{}
	case OutputAlias:
		obj = &jsonAliasOutput{}
	case OutputFoundry:
		obj = &jsonFoundryOutput{}
	case OutputNFT:
		obj = &jsonNFTOutput{}
	default:
		return nil, fmt.Errorf("unable to decode output type from JSON: %w", ErrUnknownOutputType)
	}
	return obj, nil
}
