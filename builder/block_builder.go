package builder

import (
	"crypto/ed25519"
	"time"

	"github.com/iotaledger/hive.go/ierrors"
	iotago "github.com/iotaledger/iota.go/v4"
)

// NewBasicBlockBuilder creates a new BasicBlockBuilder.
func NewBasicBlockBuilder(api iotago.API) *BasicBlockBuilder {
	// TODO: burn the correct amount of Mana in all cases according to block work and RMC with issue #285
	basicBlock := &iotago.BasicBlockBody{
		API:                api,
		StrongParents:      iotago.BlockIDs{},
		WeakParents:        iotago.BlockIDs{},
		ShallowLikeParents: iotago.BlockIDs{},
	}

	protocolBlock := &iotago.Block{
		API: api,
		Header: iotago.BlockHeader{
			ProtocolVersion:  api.ProtocolParameters().Version(),
			SlotCommitmentID: iotago.EmptyCommitmentID,
			NetworkID:        api.ProtocolParameters().NetworkID(),
			IssuingTime:      time.Now().UTC(),
		},
		Signature: &iotago.Ed25519Signature{},
		Body:      basicBlock,
	}

	return &BasicBlockBuilder{
		protocolBlock: protocolBlock,
		basicBlock:    basicBlock,
	}
}

// BasicBlockBuilder is used to easily build up a Basic Block.
type BasicBlockBuilder struct {
	basicBlock *iotago.BasicBlockBody

	protocolBlock *iotago.Block
	err           error
}

// Build builds the Block or returns any error which occurred during the build steps.
func (b *BasicBlockBuilder) Build() (*iotago.Block, error) {
	b.basicBlock.ShallowLikeParents.Sort()
	b.basicBlock.WeakParents.Sort()
	b.basicBlock.StrongParents.Sort()

	if b.err != nil {
		return nil, b.err
	}

	return b.protocolBlock, nil
}

// ProtocolVersion sets the protocol version.
func (b *BasicBlockBuilder) ProtocolVersion(version iotago.Version) *BasicBlockBuilder {
	if b.err != nil {
		return b
	}

	b.protocolBlock.Header.ProtocolVersion = version

	return b
}

func (b *BasicBlockBuilder) IssuingTime(time time.Time) *BasicBlockBuilder {
	if b.err != nil {
		return b
	}

	b.protocolBlock.Header.IssuingTime = time.UTC()

	return b
}

// SlotCommitmentID sets the slot commitment.
func (b *BasicBlockBuilder) SlotCommitmentID(commitment iotago.CommitmentID) *BasicBlockBuilder {
	if b.err != nil {
		return b
	}

	b.protocolBlock.Header.SlotCommitmentID = commitment

	return b
}

// LatestFinalizedSlot sets the latest finalized slot.
func (b *BasicBlockBuilder) LatestFinalizedSlot(slot iotago.SlotIndex) *BasicBlockBuilder {
	if b.err != nil {
		return b
	}

	b.protocolBlock.Header.LatestFinalizedSlot = slot

	return b
}

func (b *BasicBlockBuilder) Sign(accountID iotago.AccountID, privKey ed25519.PrivateKey) *BasicBlockBuilder {
	pubKey := privKey.Public().(ed25519.PublicKey)
	ed25519Address := iotago.Ed25519AddressFromPubKey(pubKey)

	signer := iotago.NewInMemoryAddressSigner(
		iotago.NewAddressKeysForEd25519Address(ed25519Address, privKey),
	)

	return b.SignWithSigner(accountID, signer, ed25519Address)
}

func (b *BasicBlockBuilder) SignWithSigner(accountID iotago.AccountID, signer iotago.AddressSigner, addr iotago.Address) *BasicBlockBuilder {
	if b.err != nil {
		return b
	}

	b.protocolBlock.Header.IssuerID = accountID

	signature, err := b.protocolBlock.Sign(signer, addr)
	if err != nil {
		b.err = ierrors.Errorf("error signing block: %w", err)

		return b
	}

	edSig, isEdSig := signature.(*iotago.Ed25519Signature)
	if !isEdSig {
		panic("unsupported signature type")
	}

	b.protocolBlock.Signature = edSig

	return b
}

// StrongParents sets the strong parents.
func (b *BasicBlockBuilder) StrongParents(parents iotago.BlockIDs) *BasicBlockBuilder {
	if b.err != nil {
		return b
	}

	b.basicBlock.StrongParents = parents.RemoveDupsAndSort()

	return b
}

// WeakParents sets the weak parents.
func (b *BasicBlockBuilder) WeakParents(parents iotago.BlockIDs) *BasicBlockBuilder {
	if b.err != nil {
		return b
	}

	b.basicBlock.WeakParents = parents.RemoveDupsAndSort()

	return b
}

// ShallowLikeParents sets the shallow like parents.
func (b *BasicBlockBuilder) ShallowLikeParents(parents iotago.BlockIDs) *BasicBlockBuilder {
	if b.err != nil {
		return b
	}

	b.basicBlock.ShallowLikeParents = parents.RemoveDupsAndSort()

	return b
}

// Payload sets the payload.
func (b *BasicBlockBuilder) Payload(payload iotago.ApplicationPayload) *BasicBlockBuilder {
	if b.err != nil {
		return b
	}

	b.basicBlock.Payload = payload

	return b
}

// MaxBurnedMana sets the maximum amount of mana allowed to be burned by the block.
func (b *BasicBlockBuilder) MaxBurnedMana(maxBurnedMana iotago.Mana) *BasicBlockBuilder {
	if b.err != nil {
		return b
	}

	b.basicBlock.MaxBurnedMana = maxBurnedMana

	return b
}

// CalculateAndSetMaxBurnedMana sets the maximum amount of mana allowed to be burned by the block based on the provided reference mana cost.
func (b *BasicBlockBuilder) CalculateAndSetMaxBurnedMana(rmc iotago.Mana) *BasicBlockBuilder {
	if b.err != nil {
		return b
	}

	// For calculating the correct workscore, we need to know the signature type, but we can only sign the block after we set the MaxBurnedMana first.
	// In the builder we assume that the signature type is always Ed25519, which is the only supported signature type at the moment, so we can
	// simplify the logic here. As a sanity check we still check the signature type here, in case we add support for other signature types in the future.
	_, isEdSig := b.protocolBlock.Signature.(*iotago.Ed25519Signature)
	if !isEdSig {
		b.err = ierrors.Errorf("only ed2519 signatures supported, got %T", b.protocolBlock.Signature)
		return b
	}

	burnedMana, err := b.protocolBlock.ManaCost(rmc)
	if err != nil {
		b.err = ierrors.Wrap(err, "error calculating mana cost")
		return b
	}

	b.basicBlock.MaxBurnedMana = burnedMana

	return b
}

// NewValidationBlockBuilder creates a new ValidationBlockBuilder.
func NewValidationBlockBuilder(api iotago.API) *ValidationBlockBuilder {
	validationBlock := &iotago.ValidationBlockBody{
		API:                api,
		StrongParents:      iotago.BlockIDs{},
		WeakParents:        iotago.BlockIDs{},
		ShallowLikeParents: iotago.BlockIDs{},
	}

	protocolBlock := &iotago.Block{
		API: api,
		Header: iotago.BlockHeader{
			ProtocolVersion:  api.ProtocolParameters().Version(),
			SlotCommitmentID: iotago.NewEmptyCommitment(api).MustID(),
			NetworkID:        api.ProtocolParameters().NetworkID(),
			IssuingTime:      time.Now().UTC(),
		},
		Signature: &iotago.Ed25519Signature{},
		Body:      validationBlock,
	}

	return &ValidationBlockBuilder{
		protocolBlock:   protocolBlock,
		validationBlock: validationBlock,
	}
}

// ValidationBlockBuilder is used to easily build up a Validation Block.
type ValidationBlockBuilder struct {
	validationBlock *iotago.ValidationBlockBody

	protocolBlock *iotago.Block
	err           error
}

// Build builds the Block or returns any error which occurred during the build steps.
func (v *ValidationBlockBuilder) Build() (*iotago.Block, error) {
	v.validationBlock.ShallowLikeParents.Sort()
	v.validationBlock.WeakParents.Sort()
	v.validationBlock.StrongParents.Sort()

	if v.err != nil {
		return nil, v.err
	}

	return v.protocolBlock, nil
}

// ProtocolVersion sets the protocol version.
func (v *ValidationBlockBuilder) ProtocolVersion(version iotago.Version) *ValidationBlockBuilder {
	if v.err != nil {
		return v
	}

	v.protocolBlock.Header.ProtocolVersion = version

	return v
}

func (v *ValidationBlockBuilder) IssuingTime(time time.Time) *ValidationBlockBuilder {
	if v.err != nil {
		return v
	}

	v.protocolBlock.Header.IssuingTime = time.UTC()

	return v
}

// SlotCommitmentID sets the slot commitment.
func (v *ValidationBlockBuilder) SlotCommitmentID(commitmentID iotago.CommitmentID) *ValidationBlockBuilder {
	if v.err != nil {
		return v
	}

	v.protocolBlock.Header.SlotCommitmentID = commitmentID

	return v
}

// LatestFinalizedSlot sets the latest finalized slot.
func (v *ValidationBlockBuilder) LatestFinalizedSlot(slot iotago.SlotIndex) *ValidationBlockBuilder {
	if v.err != nil {
		return v
	}

	v.protocolBlock.Header.LatestFinalizedSlot = slot

	return v
}

func (v *ValidationBlockBuilder) Sign(accountID iotago.AccountID, privKey ed25519.PrivateKey) *ValidationBlockBuilder {
	pubKey := privKey.Public().(ed25519.PublicKey)
	ed25519Address := iotago.Ed25519AddressFromPubKey(pubKey)

	signer := iotago.NewInMemoryAddressSigner(
		iotago.NewAddressKeysForEd25519Address(ed25519Address, privKey),
	)

	return v.SignWithSigner(accountID, signer, ed25519Address)
}

func (v *ValidationBlockBuilder) SignWithSigner(accountID iotago.AccountID, signer iotago.AddressSigner, addr iotago.Address) *ValidationBlockBuilder {
	if v.err != nil {
		return v
	}

	v.protocolBlock.Header.IssuerID = accountID

	signature, err := v.protocolBlock.Sign(signer, addr)
	if err != nil {
		v.err = ierrors.Errorf("error signing block: %w", err)

		return v
	}

	edSig, isEdSig := signature.(*iotago.Ed25519Signature)
	if !isEdSig {
		panic("unsupported signature type")
	}

	v.protocolBlock.Signature = edSig

	return v
}

// StrongParents sets the strong parents.
func (v *ValidationBlockBuilder) StrongParents(parents iotago.BlockIDs) *ValidationBlockBuilder {
	if v.err != nil {
		return v
	}

	v.validationBlock.StrongParents = parents.RemoveDupsAndSort()

	return v
}

// WeakParents sets the weak parents.
func (v *ValidationBlockBuilder) WeakParents(parents iotago.BlockIDs) *ValidationBlockBuilder {
	if v.err != nil {
		return v
	}

	v.validationBlock.WeakParents = parents.RemoveDupsAndSort()

	return v
}

// ShallowLikeParents sets the shallow like parents.
func (v *ValidationBlockBuilder) ShallowLikeParents(parents iotago.BlockIDs) *ValidationBlockBuilder {
	if v.err != nil {
		return v
	}

	v.validationBlock.ShallowLikeParents = parents.RemoveDupsAndSort()

	return v
}

// HighestSupportedVersion sets the highest supported version.
func (v *ValidationBlockBuilder) HighestSupportedVersion(highestSupportedVersion iotago.Version) *ValidationBlockBuilder {
	if v.err != nil {
		return v
	}

	v.validationBlock.HighestSupportedVersion = highestSupportedVersion

	return v
}

// ProtocolParametersHash sets the ProtocolParametersHash of the highest supported version.
func (v *ValidationBlockBuilder) ProtocolParametersHash(hash iotago.Identifier) *ValidationBlockBuilder {
	if v.err != nil {
		return v
	}

	v.validationBlock.ProtocolParametersHash = hash

	return v
}
