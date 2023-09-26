package builder

import (
	"github.com/iotaledger/hive.go/ierrors"
	iotago "github.com/iotaledger/iota.go/v4"
)

// ErrTransactionBuilder defines a generic error occurring within the TransactionBuilder.
var ErrTransactionBuilder = ierrors.New("transaction builder error")

// NewTransactionBuilder creates a new TransactionBuilder.
func NewTransactionBuilder(api iotago.API) *TransactionBuilder {
	return &TransactionBuilder{
		api: api,
		essence: &iotago.TransactionEssence{
			TransactionInputEssence: &iotago.TransactionInputEssence{
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

// TransactionBuilder is used to easily build up a Transaction.
type TransactionBuilder struct {
	api              iotago.API
	occurredBuildErr error
	essence          *iotago.TransactionEssence
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
		essence:          b.essence.Clone(),
		inputs:           b.inputs.Clone(),
		inputOwner:       cpyInputOwner,
	}
}

// AddInput adds the given input to the builder.
func (b *TransactionBuilder) AddInput(input *TxInput) *TransactionBuilder {
	b.inputOwner[input.InputID] = input.UnlockTarget
	b.essence.Inputs = append(b.essence.Inputs, input.InputID.UTXOInput())
	b.inputs[input.InputID] = input.Input

	return b
}

// TransactionBuilderInputFilter is a filter function which determines whether
// an input should be used or not. (returning true = pass). The filter can also
// be used to accumulate data over the set of inputs, i.e. the input sum etc.
type TransactionBuilderInputFilter func(outputID iotago.OutputID, input iotago.Output) bool

// AddContextInput adds the given context input to the builder.
func (b *TransactionBuilder) AddContextInput(contextInput iotago.Input) *TransactionBuilder {
	b.essence.ContextInputs = append(b.essence.ContextInputs, contextInput)

	return b
}

// AddAllotment adds the given allotment to the builder.
func (b *TransactionBuilder) AddAllotment(allotment *iotago.Allotment) *TransactionBuilder {
	b.essence.Allotments = append(b.essence.Allotments, allotment)

	return b
}

// AddOutput adds the given output to the builder.
func (b *TransactionBuilder) AddOutput(output iotago.Output) *TransactionBuilder {
	b.essence.Outputs = append(b.essence.Outputs, output)

	return b
}

func (b *TransactionBuilder) SetCreationSlot(creationSlot iotago.SlotIndex) *TransactionBuilder {
	b.essence.CreationSlot = creationSlot

	return b
}

// AddTaggedDataPayload adds the given TaggedData as the inner payload.
func (b *TransactionBuilder) AddTaggedDataPayload(payload *iotago.TaggedData) *TransactionBuilder {
	b.essence.Payload = payload

	return b
}

// TransactionFunc is a function which receives a Transaction as its parameter.
type TransactionFunc func(tx *iotago.Transaction)

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

// Build sings the inputs with the given signer and returns the built payload.
func (b *TransactionBuilder) Build(signer iotago.AddressSigner) (*iotago.Transaction, error) {
	switch {
	case b.occurredBuildErr != nil:
		return nil, b.occurredBuildErr
	case signer == nil:
		return nil, ierrors.Wrap(ErrTransactionBuilder, "must supply signer")
	}

	// prepare the inputs commitment in the same order as the inputs in the essence
	var inputIDs iotago.OutputIDs
	for _, input := range b.essence.Inputs {
		//nolint:forcetypeassert // we can safely assume that this is an UTXOInput
		inputIDs = append(inputIDs, input.(*iotago.UTXOInput).OutputID())
	}

	inputs := inputIDs.OrderedSet(b.inputs)
	commitment, err := inputs.Commitment(b.api)
	if err != nil {
		return nil, ierrors.Wrapf(err, "failed to calculate TX inputs commitment: %s, %s", inputIDs, b.inputs)
	}
	copy(b.essence.InputsCommitment[:], commitment)

	txEssenceData, err := b.essence.SigningMessage(b.api)
	if err != nil {
		return nil, ierrors.Wrap(err, "failed to calculate tx essence for signing message")
	}

	unlockPos := map[string]int{}
	unlocks := iotago.Unlocks{}
	for i, inputRef := range b.essence.Inputs {
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
				return nil, ierrors.Wrapf(err, "failed to sign tx essence: %s", txEssenceData)
			}

			unlocks = append(unlocks, &iotago.SignatureUnlock{Signature: signature})
			addChainAsUnlocked(inputs[i], i, unlockPos)
			unlockPos[addrKey] = i

			continue
		}

		unlocks = addReferentialUnlock(addr, unlocks, pos)
		addChainAsUnlocked(inputs[i], i, unlockPos)
	}

	sigTxPayload := &iotago.Transaction{Essence: b.essence, Unlocks: unlocks}

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
	if chainInput, is := input.(iotago.ChainOutput); is && chainInput.Chain().Addressable() {
		prevUnlocked[chainInput.Chain().ToAddress().Key()] = posUnlocked
	}
}
