package vm

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	iotago "github.com/iotaledger/iota.go/v4"
)

// VirtualMachine executes and validates transactions.
type VirtualMachine interface {
	// Execute executes the given tx in the VM.
	// Pass own ExecFunc(s) to override the VM's default execution function list.
	Execute(t *iotago.Transaction, params *Params, inputs iotago.InputSet, overrideFuncs ...ExecFunc) error
	// ChainSTVF executes the chain state transition validation function.
	ChainSTVF(transType iotago.ChainTransitionType, input iotago.ChainOutputWithCreationTime, next iotago.ChainOutput, vmParams *Params) error
}

// Params defines the VirtualMachine parameters under which the VM operates.
type Params struct {
	External *iotago.ExternalUnlockParameters

	// The working set which is auto. populated during the semantic validation.
	WorkingSet *WorkingSet
}

// WorkingSet contains fields which get automatically populated
// by the library during execution of a Transaction.
type WorkingSet struct {
	// The identities which are successfully unlocked from the input side.
	UnlockedIdents UnlockedIdentities
	// The mapping of OutputID to the actual Outputs and their creation times.
	InputSet iotago.InputSet
	// The inputs to the transaction.
	Inputs iotago.Outputs[iotago.Output]
	// The mapping of inputs' OutputID to the index.
	InputIDToIndex map[iotago.OutputID]uint16
	// The transaction for which this semantic validation happens.
	Tx *iotago.Transaction
	// The message which signatures are signing.
	EssenceMsgToSign []byte
	// The inputs of the transaction mapped by type.
	InputsByType iotago.OutputsByType
	// The ChainOutput(s) at the input side.
	InChains iotago.ChainInputSet
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
	// BIC is the block issuance credit for MCA slots prior to the transaction's creation time (or for the slot to which the block commits)
	// Contains one value for each account output touched in the transaction and empty if no acccount outputs touched.
	BIC iotago.BlockIssuanceCredit
}

// UTXOInputAtIndex retrieves the UTXOInput at the given index.
// Caller must ensure that the index is valid.
func (workingSet *WorkingSet) UTXOInputAtIndex(inputIndex uint16) *iotago.UTXOInput {
	return workingSet.Tx.Essence.Inputs[inputIndex].(*iotago.UTXOInput)
}

func NewVMParamsWorkingSet(t *iotago.Transaction, inputsSet iotago.InputSet) (*WorkingSet, error) {
	var err error
	workingSet := &WorkingSet{}
	workingSet.Tx = t
	workingSet.UnlockedIdents = make(UnlockedIdentities)
	workingSet.InputSet = inputsSet
	workingSet.InputIDToIndex = make(map[iotago.OutputID]uint16)
	for inputIndex, inputRef := range workingSet.Tx.Essence.Inputs {
		ref := inputRef.(iotago.IndexedUTXOReferencer).Ref()
		workingSet.InputIDToIndex[ref] = uint16(inputIndex)
		input, ok := workingSet.InputSet[ref]
		if !ok {
			return nil, fmt.Errorf("%w: utxo for input %d not supplied", iotago.ErrMissingUTXO, inputIndex)
		}
		workingSet.Inputs = append(workingSet.Inputs, input)
	}

	workingSet.EssenceMsgToSign, err = t.Essence.SigningMessage()
	if err != nil {
		return nil, err
	}

	workingSet.InputsByType = func() iotago.OutputsByType {
		slice := make(iotago.Outputs[iotago.Output], len(inputsSet))
		var i int
		for _, output := range inputsSet {
			slice[i] = output
			i++
		}
		return slice.ToOutputsByType()
	}()

	txID, err := workingSet.Tx.ID()
	if err != nil {
		return nil, err
	}

	workingSet.InChains = workingSet.InputSet.ChainInputSet()
	workingSet.OutputsByType = t.Essence.Outputs.ToOutputsByType()
	workingSet.OutChains = workingSet.Tx.Essence.Outputs.ChainOutputSet(txID)

	workingSet.UnlocksByType = t.Unlocks.ToUnlockByType()
	return workingSet, nil
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
func (unlockedIdents UnlockedIdentities) SigUnlock(ident iotago.DirectUnlockableAddress, essence []byte, sig iotago.Signature, inputIndex uint16) error {
	if err := ident.Unlock(essence, sig); err != nil {
		return fmt.Errorf("%w: input %d's address is not unlocked through its signature unlock", err, inputIndex)
	}

	unlockedIdents[ident.Key()] = &UnlockedIdentity{
		Ident:      ident,
		UnlockedAt: inputIndex, ReferencedBy: map[uint16]struct{}{},
	}
	return nil
}

// RefUnlock performs a check whether the given ident is unlocked at ref and if so,
// adds the index of the input to the set of unlocked inputs by this identity.
func (unlockedIdents UnlockedIdentities) RefUnlock(identKey string, ref uint16, inputIndex uint16) error {
	ident, has := unlockedIdents[identKey]
	if !has || ident.UnlockedAt != ref {
		return fmt.Errorf("%w: input %d is not unlocked through input %d's unlock", iotago.ErrInvalidInputUnlock, inputIndex, ref)
	}

	ident.ReferencedBy[inputIndex] = struct{}{}
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
func IsIssuerOnOutputUnlocked(output iotago.ChainOutput, unlockedIdents UnlockedIdentities) error {
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
		actualInputCommitment, err := vmParams.WorkingSet.Inputs.Commitment()
		if err != nil {
			return fmt.Errorf("unable to compute hash of inputs: %w", err)
		}

		expectedInputCommitment := vmParams.WorkingSet.Tx.Essence.InputsCommitment[:]
		if !bytes.Equal(expectedInputCommitment, actualInputCommitment) {
			return fmt.Errorf("%w: specified %v but got %v", iotago.ErrInvalidInputsCommitment, expectedInputCommitment, actualInputCommitment)
		}

		for inputIndex, input := range vmParams.WorkingSet.Inputs {
			if err := unlockOutput(vmParams, input, uint16(inputIndex)); err != nil {
				return err
			}

			// since this input is now unlocked, and it is a ChainOutput, the chain's address becomes automatically unlocked
			if chainConstrOutput, is := input.(iotago.ChainOutput); is && chainConstrOutput.Chain().Addressable() {
				// mark this ChainOutput's identity as unlocked by this input
				chainID := chainConstrOutput.Chain()
				if chainID.Empty() {
					chainID = chainID.(iotago.UTXOIDChainID).FromOutputID(vmParams.WorkingSet.UTXOInputAtIndex(uint16(inputIndex)).Ref())
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
			chainID = utxoChainID.FromOutputID(vmParams.WorkingSet.Tx.Essence.Inputs[inputIndex].(iotago.IndexedUTXOReferencer).Ref())
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

func checkExpiredForReceiver(vmParams *Params, output iotago.Output) iotago.Address {
	if ok, returnIdent := output.UnlockConditionSet().ReturnIdentCanUnlock(vmParams.External); ok {
		return returnIdent
	}

	return nil
}

func unlockOutput(vmParams *Params, output iotago.Output, inputIndex uint16) error {
	ownerIdent, err := identToUnlock(vmParams, output, inputIndex)
	if err != nil {
		return fmt.Errorf("unable to retrieve ident to unlock of input %d: %w", inputIndex, err)
	}

	if actualIdentToUnlock := checkExpiredForReceiver(vmParams, output); actualIdentToUnlock != nil {
		ownerIdent = actualIdentToUnlock
	}

	unlock := vmParams.WorkingSet.Tx.Unlocks[inputIndex]

	switch owner := ownerIdent.(type) {
	case iotago.ChainAddress:
		refUnlock, isReferentialUnlock := unlock.(iotago.ReferentialUnlock)
		if !isReferentialUnlock || !refUnlock.Chainable() || !refUnlock.SourceAllowed(ownerIdent) {
			return fmt.Errorf("%w: input %d has a chain address (%T) but its corresponding unlock is of type %T", iotago.ErrInvalidInputUnlock, inputIndex, owner, unlock)
		}

		if err := vmParams.WorkingSet.UnlockedIdents.RefUnlock(owner.Key(), refUnlock.Ref(), inputIndex); err != nil {
			return fmt.Errorf("%w: chain address %s (%T)", err, owner, owner)
		}

	case iotago.DirectUnlockableAddress:
		switch uBlock := unlock.(type) {
		case iotago.ReferentialUnlock:
			if uBlock.Chainable() || !uBlock.SourceAllowed(ownerIdent) {
				return fmt.Errorf("%w: input %d has none chain address of %s but its corresponding unlock is of type %s", iotago.ErrInvalidInputUnlock, inputIndex, owner.Type(), unlock.Type())
			}

			if err := vmParams.WorkingSet.UnlockedIdents.RefUnlock(owner.Key(), uBlock.Ref(), inputIndex); err != nil {
				return fmt.Errorf("%w: direct unlockable address %s (%T)", err, owner, owner)
			}

		case *iotago.SignatureUnlock:
			// owner must not be unlocked already
			if unlockedAtIndex, wasAlreadyUnlocked := vmParams.WorkingSet.UnlockedIdents[owner.Key()]; wasAlreadyUnlocked {
				return fmt.Errorf("%w: input %d's address is already unlocked through input %d's unlock but the input uses a non referential unlock", iotago.ErrInvalidInputUnlock, inputIndex, unlockedAtIndex)
			}

			if err := vmParams.WorkingSet.UnlockedIdents.SigUnlock(owner, vmParams.WorkingSet.EssenceMsgToSign, uBlock.Signature, inputIndex); err != nil {
				return err
			}

		}
	default:
		panic("unknown address in semantic unlocks")
	}

	return nil
}

// ExecFuncSenderUnlocked validates that for SenderFeature occurring on the output side,
// the given identity is unlocked on the input side.
func ExecFuncSenderUnlocked() ExecFunc {
	return func(vm VirtualMachine, vmParams *Params) error {
		for outputIndex, output := range vmParams.WorkingSet.Tx.Essence.Outputs {
			senderFeat := output.FeatureSet().SenderFeature()
			if senderFeat == nil {
				continue
			}

			// check unlocked
			sender := senderFeat.Address
			if _, isUnlocked := vmParams.WorkingSet.UnlockedIdents[sender.Key()]; !isUnlocked {
				return fmt.Errorf("%w: output %d", iotago.ErrSenderFeatureNotUnlocked, outputIndex)
			}
		}
		return nil
	}
}

// ExecFuncBalancedMana validates that Mana is balanced from the input/output side.
// TODO: Return Mana according to StorageDepositReturnUnlockCondition(s)?
func ExecFuncBalancedMana() ExecFunc {
	return func(vm VirtualMachine, vmParams *Params) error {
		var inStored, inPotential, outStored, outAllotted uint64
		txCreationTime := vmParams.WorkingSet.Tx.Essence.CreationTime
		for _, input := range vmParams.WorkingSet.InputSet {
			inStored += input.StoredMana()
			inPotential += input.Deposit() * uint64(vmParams.External.ManaGenerationRate) * uint64(txCreationTime-input.CreationTime)
		}

		// TODO: apply decay to input side Mana here.

		for _, output := range vmParams.WorkingSet.Tx.Essence.Outputs {
			outStored += output.StoredMana()
		}
		for _, allotment := range vmParams.WorkingSet.Tx.Essence.Allotments {
			outAllotted += allotment.Mana
		}

		if inStored+inPotential != outStored+outAllotted {
			return fmt.Errorf("%w: inStored %d, inPotential %d, outStored %d, outAllotted %d", iotago.ErrInputOutputSumMismatch, inStored, inPotential, outStored, outAllotted)
		}

		return nil
	}
}

// ExecFuncBalancedDeposit validates that the IOTA tokens are balanced from the input/output side.
// It additionally also incorporates the check whether return amounts via StorageDepositReturnUnlockCondition(s) for specified identities
// are fulfilled from the output side.
func ExecFuncBalancedDeposit() ExecFunc {
	return func(vm VirtualMachine, vmParams *Params) error {
		// note that due to syntactic validation of outputs, input and output deposit sums
		// are always within bounds of the total token supply
		var in, out uint64
		inputSumReturnAmountPerIdent := make(map[string]uint64)
		for inputID, input := range vmParams.WorkingSet.InputSet {
			in += input.Deposit()

			returnUnlockCond := input.UnlockConditionSet().StorageDepositReturn()
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

		outputSimpleTransfersPerIdent := make(map[string]uint64)
		for _, output := range vmParams.WorkingSet.Tx.Essence.Outputs {
			outDeposit := output.Deposit()
			out += outDeposit

			// accumulate simple transfers for StorageDepositReturnUnlockCondition checks
			if basicOutput, is := output.(*iotago.BasicOutput); is && basicOutput.IsSimpleTransfer() {
				outputSimpleTransfersPerIdent[basicOutput.Ident().Key()] += outDeposit
			}
		}

		if in != out {
			return fmt.Errorf("%w: in %d, out %d", iotago.ErrInputOutputSumMismatch, in, out)
		}

		for ident, returnSum := range inputSumReturnAmountPerIdent {
			outSum, has := outputSimpleTransfersPerIdent[ident]
			if !has {
				return fmt.Errorf("%w: return amount of %d not fulfilled as there is no output for %s", iotago.ErrReturnAmountNotFulFilled, returnSum, ident)
			}
			if outSum < returnSum {
				return fmt.Errorf("%w: return amount of %d not fulfilled as output is only %d for %s", iotago.ErrReturnAmountNotFulFilled, returnSum, outSum, ident)
			}
		}

		return nil
	}
}

// ExecFuncTimelocks validates that the inputs' timelocks are expired.
func ExecFuncTimelocks() ExecFunc {
	return func(vm VirtualMachine, vmParams *Params) error {
		for inputIndex, input := range vmParams.WorkingSet.InputSet {
			if err := input.UnlockConditionSet().TimelocksExpired(vmParams.External); err != nil {
				return fmt.Errorf("%w: input at index %d's timelocks are not expired", err, inputIndex)
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
					return fmt.Errorf("input chain %s (%T) destruction transition failed: %w", chainID, inputChain, err)
				}
				continue
			}
			if err := vm.ChainSTVF(iotago.ChainTransitionTypeStateChange, inputChain, next, vmParams); err != nil {
				return fmt.Errorf("chain %s (%T) state transition failed: %w", chainID, inputChain, err)
			}
		}

		for chainID, outputChain := range vmParams.WorkingSet.OutChains {
			inputChain := vmParams.WorkingSet.InChains[chainID]
			if previousState := inputChain.ChainOutput; previousState != nil {
				continue
			}
			if err := vm.ChainSTVF(iotago.ChainTransitionTypeGenesis, inputChain, outputChain, vmParams); err != nil {
				return fmt.Errorf("new chain %s (%T) state transition failed: %w", chainID, outputChain, err)
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
		vmParams.WorkingSet.InNativeTokens, err = vmParams.WorkingSet.Inputs.NativeTokenSum()
		if err != nil {
			return fmt.Errorf("invalid input native token set: %w", err)
		}
		inNTCount := len(vmParams.WorkingSet.InNativeTokens)

		if inNTCount > iotago.MaxNativeTokensCount {
			return fmt.Errorf("%w: inputs native token count %d exceeds max of %d", iotago.ErrMaxNativeTokensCountExceeded, inNTCount, iotago.MaxNativeTokensCount)
		}

		vmParams.WorkingSet.OutNativeTokens, err = vmParams.WorkingSet.Tx.Essence.Outputs.NativeTokenSum()
		if err != nil {
			return fmt.Errorf("invalid output native token set: %w", err)
		}

		distinctNTCount := make(map[iotago.NativeTokenID]struct{})
		for nt := range vmParams.WorkingSet.InNativeTokens {
			distinctNTCount[nt] = struct{}{}
		}
		for nt := range vmParams.WorkingSet.OutNativeTokens {
			distinctNTCount[nt] = struct{}{}
		}

		if len(distinctNTCount) > iotago.MaxNativeTokensCount {
			return fmt.Errorf("%w: native token count %d exceeds max of %d", iotago.ErrMaxNativeTokensCountExceeded, distinctNTCount, iotago.MaxNativeTokensCount)
		}

		// check invariants for when token foundry is absent

		for nativeTokenID, inSum := range vmParams.WorkingSet.InNativeTokens {
			if _, foundryIsTransitioning := vmParams.WorkingSet.OutChains[nativeTokenID]; foundryIsTransitioning {
				continue
			}

			// input sum must be greater equal the output sum (burning allows it to be greater)
			if outSum := vmParams.WorkingSet.OutNativeTokens[nativeTokenID]; outSum != nil && inSum.Cmp(outSum) == -1 {
				return fmt.Errorf("%w: native token %s is less on input (%d) than output (%d) side but the foundry is absent for minting", iotago.ErrNativeTokenSumUnbalanced, nativeTokenID, inSum, outSum)
			}
		}

		for nativeTokenID := range vmParams.WorkingSet.OutNativeTokens {
			if _, foundryIsTransitioning := vmParams.WorkingSet.OutChains[nativeTokenID]; foundryIsTransitioning {
				continue
			}

			// foundry must be present when native tokens only reside on the output side
			// as they need to get minted by it within the tx
			if vmParams.WorkingSet.InNativeTokens[nativeTokenID] == nil {
				return fmt.Errorf("%w: native token %s is new on the output side but the foundry is not transitioning", iotago.ErrNativeTokenSumUnbalanced, nativeTokenID)
			}
		}

		// from here the native tokens balancing is handled by each foundry's STVF

		return nil
	}
}
