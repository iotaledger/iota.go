//nolint:gosec
package tpkg

import (
	"bytes"
	"crypto/ed25519"
	cryptorand "crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"slices"
	"sort"
	"time"

	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/hive.go/runtime/options"
	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v4"
)

// Must panics if the given error is not nil.
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

// RandByte returns a random byte.
func RandByte() byte {
	return byte(rand.Intn(256))
}

// RandBytes returns length amount random bytes.
func RandBytes(length int) []byte {
	var b []byte
	for i := 0; i < length; i++ {
		b = append(b, byte(rand.Intn(127)))
	}

	return b
}

func RandString(length int) string {
	return string(RandBytes(length))
}

// RandInt returns a random int.
func RandInt(max int) int {
	return rand.Intn(max)
}

// RandUint8 returns a random uint8.
func RandUint8(max uint8) uint8 {
	return uint8(rand.Int31n(int32(max)))
}

// RandUint16 returns a random uint16.
func RandUint16(max uint16) uint16 {
	return uint16(rand.Int31n(int32(max)))
}

// RandUint32 returns a random uint32.
func RandUint32(max uint32) uint32 {
	return uint32(rand.Int63n(int64(max)))
}

// RandUint64 returns a random uint64.
func RandUint64(max uint64) uint64 {
	return uint64(rand.Int63n(int64(uint32(max))))
}

// RandInt64 returns a random int64.
func RandInt64(max uint64) int64 {
	return rand.Int63n(int64(uint32(max)))
}

// RandUTCTime returns a random time from current year until now in UTC.
func RandUTCTime() time.Time {
	now := time.Now()
	beginnigOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	secTillNow := now.Unix() - beginnigOfYear.Unix()

	return time.Unix(beginnigOfYear.Unix()+RandInt64(uint64(secTillNow)), RandInt64(1e9)).UTC()
}

// RandBaseToken returns a random amount of base token.
func RandBaseToken(max iotago.BaseToken) iotago.BaseToken {
	return iotago.BaseToken(rand.Int63n(int64(uint32(max))))
}

// RandMana returns a random amount of mana.
func RandMana(max iotago.Mana) iotago.Mana {
	return iotago.Mana(rand.Int63n(int64(uint32(max))))
}

// RandFloat64 returns a random float64.
func RandFloat64(max float64) float64 {
	return rand.Float64() * max
}

func RandOutputIDWithCreationSlot(slot iotago.SlotIndex, index uint16) iotago.OutputID {
	txID := RandTransactionIDWithCreationSlot(slot)

	var outputID iotago.OutputID
	copy(outputID[:], txID[:])
	binary.LittleEndian.PutUint16(outputID[iotago.TransactionIDLength:], index)

	return outputID
}

func RandOutputID(index uint16) iotago.OutputID {
	return RandOutputIDWithCreationSlot(0, index)
}

func RandOutputIDsWithCreationSlot(slot iotago.SlotIndex, count uint16) iotago.OutputIDs {
	outputIDs := make(iotago.OutputIDs, int(count))
	for i := 0; i < int(count); i++ {
		outputIDs[i] = RandOutputIDWithCreationSlot(slot, count)
	}

	return outputIDs
}

func RandOutputIDs(count uint16) iotago.OutputIDs {
	outputIDs := make(iotago.OutputIDs, int(count))
	for i := 0; i < int(count); i++ {
		outputIDs[i] = RandOutputID(count)
	}

	return outputIDs
}

func RandSignedTransactionIDWithCreationSlot(slot iotago.SlotIndex) iotago.SignedTransactionID {
	var signedTransactionID iotago.SignedTransactionID
	_, err := cryptorand.Read(signedTransactionID[:iotago.IdentifierLength])
	if err != nil {
		panic(err)
	}
	binary.LittleEndian.PutUint32(signedTransactionID[iotago.IdentifierLength:iotago.TransactionIDLength], uint32(slot))

	return signedTransactionID
}

func RandTransactionIDWithCreationSlot(slot iotago.SlotIndex) iotago.TransactionID {
	var transactionID iotago.TransactionID
	_, err := cryptorand.Read(transactionID[:iotago.IdentifierLength])
	if err != nil {
		panic(err)
	}
	binary.LittleEndian.PutUint32(transactionID[iotago.IdentifierLength:iotago.TransactionIDLength], uint32(slot))

	return transactionID
}

func RandSignedTransactionID() iotago.SignedTransactionID {
	return RandSignedTransactionIDWithCreationSlot(RandSlot())
}

func RandTransactionID() iotago.TransactionID {
	return RandTransactionIDWithCreationSlot(RandSlot())
}

func RandNativeTokenID() iotago.NativeTokenID {
	var nativeTokenID iotago.NativeTokenID
	copy(nativeTokenID[:], RandBytes(iotago.NativeTokenIDLength))

	// the underlying address needs to be an account address
	nativeTokenID[0] = byte(iotago.AddressAccount)

	// set the simple token scheme type
	nativeTokenID[iotago.FoundryIDLength-iotago.FoundryTokenSchemeLength] = byte(iotago.TokenSchemeSimple)

	return nativeTokenID
}

// RandNativeTokenFeature returns a random NativeToken feature.
func RandNativeTokenFeature() *iotago.NativeTokenFeature {
	return &iotago.NativeTokenFeature{
		ID:     RandNativeTokenID(),
		Amount: RandUint256(),
	}
}

func RandUint256() *big.Int {
	return new(big.Int).SetUint64(rand.Uint64())
}

// Rand12ByteArray returns an array with 12 random bytes.
func Rand12ByteArray() [12]byte {
	var h [12]byte
	b := RandBytes(12)
	copy(h[:], b)

	return h
}

// Rand32ByteArray returns an array with 32 random bytes.
func Rand32ByteArray() [32]byte {
	var h [32]byte
	b := RandBytes(32)
	copy(h[:], b)

	return h
}

// Rand36ByteArray returns an array with 36 random bytes.
func Rand36ByteArray() [36]byte {
	var h [36]byte
	b := RandBytes(36)
	copy(h[:], b)

	return h
}

// Rand40ByteArray returns an array with 40 random bytes.
func Rand40ByteArray() [40]byte {
	var h [40]byte
	b := RandBytes(40)
	copy(h[:], b)

	return h
}

// Rand50ByteArray returns an array with 38 random bytes.
func Rand50ByteArray() [50]byte {
	var h [50]byte
	b := RandBytes(50)
	copy(h[:], b)

	return h
}

// Rand38ByteArray returns an array with 38 random bytes.
func Rand38ByteArray() [38]byte {
	var h [38]byte
	b := RandBytes(38)
	copy(h[:], b)

	return h
}

// Rand49ByteArray returns an array with 49 random bytes.
func Rand49ByteArray() [49]byte {
	var h [49]byte
	b := RandBytes(49)
	copy(h[:], b)

	return h
}

// Rand64ByteArray returns an array with 64 random bytes.
func Rand64ByteArray() [64]byte {
	var h [64]byte
	b := RandBytes(64)
	copy(h[:], b)

	return h
}

// SortedRand32ByteArray returns a count length slice of sorted 32 byte arrays.
func SortedRand32ByteArray(count int) [][32]byte {
	hashes := make(serializer.LexicalOrdered32ByteArrays, count)
	for i := 0; i < count; i++ {
		hashes[i] = Rand32ByteArray()
	}
	sort.Sort(hashes)

	return hashes
}

// SortedRand36ByteArray returns a count length slice of sorted 36 byte arrays.
func SortedRand36ByteArray(count int) [][36]byte {
	hashes := make(serializer.LexicalOrdered36ByteArrays, count)
	for i := 0; i < count; i++ {
		hashes[i] = Rand36ByteArray()
	}
	sort.Sort(hashes)

	return hashes
}

// SortedRand40ByteArray returns a count length slice of sorted 32 byte arrays.
func SortedRand40ByteArray(count int) [][40]byte {
	hashes := make(serializer.LexicalOrdered40ByteArrays, count)
	for i := 0; i < count; i++ {
		hashes[i] = Rand40ByteArray()
	}
	sort.Sort(hashes)

	return hashes
}

// SortedRandBlockIDs returned random block IDs.
func SortedRandBlockIDs(count int) iotago.BlockIDs {
	slice := make([]iotago.BlockID, count)
	for i, ele := range SortedRand36ByteArray(count) {
		slice[i] = ele
	}

	return slice
}

// RandEd25519Address returns a random Ed25519 address.
func RandEd25519Address() *iotago.Ed25519Address {
	edAddr := &iotago.Ed25519Address{}
	addr := RandBytes(iotago.Ed25519AddressBytesLength)
	copy(edAddr[:], addr)

	return edAddr
}

// RandAccountAddress returns a random AccountAddress.
func RandAccountAddress() *iotago.AccountAddress {
	addr := &iotago.AccountAddress{}
	accountID := RandBytes(iotago.AccountAddressBytesLength)
	copy(addr[:], accountID)

	return addr
}

// RandNFTAddress returns a random NFTAddress.
func RandNFTAddress() *iotago.NFTAddress {
	addr := &iotago.NFTAddress{}
	nftID := RandBytes(iotago.NFTAddressBytesLength)
	copy(addr[:], nftID)

	return addr
}

// RandAnchorAddress returns a random AnchorAddress.
func RandAnchorAddress() *iotago.AnchorAddress {
	addr := &iotago.AnchorAddress{}
	anchorID := RandBytes(iotago.AnchorAddressBytesLength)
	copy(addr[:], anchorID)

	return addr
}

// RandImplicitAccountCreationAddress returns a random ImplicitAccountCreationAddress.
func RandImplicitAccountCreationAddress() *iotago.ImplicitAccountCreationAddress {
	iacAddr := &iotago.ImplicitAccountCreationAddress{}
	addr := RandBytes(iotago.Ed25519AddressBytesLength)
	copy(iacAddr[:], addr)

	return iacAddr
}

// RandMultiAddress returns a random MultiAddress.
func RandMultiAddress() *iotago.MultiAddress {
	addrCnt := RandInt(10) + 1

	cumulativeWeight := 0
	addresses := make([]*iotago.AddressWithWeight, 0, addrCnt)
	for i := 0; i < addrCnt; i++ {
		weight := RandInt(8) + 1
		cumulativeWeight += weight
		addresses = append(addresses, &iotago.AddressWithWeight{
			Address: RandAddress(),
			Weight:  byte(weight),
		})
	}

	slices.SortFunc(addresses, func(a *iotago.AddressWithWeight, b *iotago.AddressWithWeight) int {
		return bytes.Compare(a.Address.ID(), b.Address.ID())
	})

	threshold := RandInt(cumulativeWeight) + 1

	return &iotago.MultiAddress{
		Addresses: addresses,
		Threshold: uint16(threshold),
	}
}

// RandRestrictedEd25519Address returns a random restricted Ed25519 address.
func RandRestrictedEd25519Address(capabilities iotago.AddressCapabilitiesBitMask) *iotago.RestrictedAddress {
	return &iotago.RestrictedAddress{
		Address:             RandEd25519Address(),
		AllowedCapabilities: capabilities,
	}
}

// RandRestrictedAccountAddress returns a random restricted account address.
func RandRestrictedAccountAddress(capabilities iotago.AddressCapabilitiesBitMask) *iotago.RestrictedAddress {
	return &iotago.RestrictedAddress{
		Address:             RandAccountAddress(),
		AllowedCapabilities: capabilities,
	}
}

// RandRestrictedNFTAddress returns a random restricted NFT address.
func RandRestrictedNFTAddress(capabilities iotago.AddressCapabilitiesBitMask) *iotago.RestrictedAddress {
	return &iotago.RestrictedAddress{
		Address:             RandNFTAddress(),
		AllowedCapabilities: capabilities,
	}
}

// RandRestrictedAnchorAddress returns a random restricted anchor address.
func RandRestrictedAnchorAddress(capabilities iotago.AddressCapabilitiesBitMask) *iotago.RestrictedAddress {
	return &iotago.RestrictedAddress{
		Address:             RandAnchorAddress(),
		AllowedCapabilities: capabilities,
	}
}

// RandRestrictedMultiAddress returns a random restricted multi address.
func RandRestrictedMultiAddress(capabilities iotago.AddressCapabilitiesBitMask) *iotago.RestrictedAddress {
	return &iotago.RestrictedAddress{
		Address:             RandMultiAddress(),
		AllowedCapabilities: capabilities,
	}
}

// RandAddress returns a random address (Ed25519, Account, NFT, Anchor).
func RandAddress() iotago.Address {
	addressTypes := []iotago.AddressType{iotago.AddressEd25519, iotago.AddressAccount, iotago.AddressNFT, iotago.AddressAnchor}

	addressType := addressTypes[RandInt(len(addressTypes))]

	//nolint:exhaustive
	switch addressType {
	case iotago.AddressEd25519:
		return RandEd25519Address()
	case iotago.AddressAccount:
		return RandAccountAddress()
	case iotago.AddressNFT:
		return RandNFTAddress()
	case iotago.AddressAnchor:
		return RandAnchorAddress()
	default:
		panic(ierrors.Wrapf(iotago.ErrUnknownAddrType, "type %d", addressType))
	}
}

// RandEd25519Signature returns a random Ed25519 signature.
func RandEd25519Signature() *iotago.Ed25519Signature {
	edSig := &iotago.Ed25519Signature{}
	pub := RandBytes(ed25519.PublicKeySize)
	sig := RandBytes(ed25519.SignatureSize)
	copy(edSig.PublicKey[:], pub)
	copy(edSig.Signature[:], sig)

	return edSig
}

// RandUnlock returns a random unlock (except Signature, Reference, Account, Anchor, NFT).
func RandUnlock(allowEmptyUnlock bool) iotago.Unlock {
	unlockTypes := []iotago.UnlockType{iotago.UnlockSignature, iotago.UnlockReference, iotago.UnlockAccount, iotago.UnlockAnchor, iotago.UnlockNFT}

	if allowEmptyUnlock {
		unlockTypes = append(unlockTypes, iotago.UnlockEmpty)
	}

	unlockType := unlockTypes[RandInt(len(unlockTypes))]

	//nolint:exhaustive
	switch unlockType {
	case iotago.UnlockSignature:
		return RandEd25519SignatureUnlock()
	case iotago.UnlockReference:
		return RandReferenceUnlock()
	case iotago.UnlockAccount:
		return RandAccountUnlock()
	case iotago.UnlockAnchor:
		return RandAnchorUnlock()
	case iotago.UnlockNFT:
		return RandNFTUnlock()
	case iotago.UnlockEmpty:
		return &iotago.EmptyUnlock{}
	default:
		panic(ierrors.Wrapf(iotago.ErrUnknownUnlockType, "type %d", unlockType))
	}
}

// RandEd25519SignatureUnlock returns a random Ed25519 signature unlock.
func RandEd25519SignatureUnlock() *iotago.SignatureUnlock {
	return &iotago.SignatureUnlock{Signature: RandEd25519Signature()}
}

// RandReferenceUnlock returns a random reference unlock.
func RandReferenceUnlock() *iotago.ReferenceUnlock {
	return ReferenceUnlock(uint16(rand.Intn(1000)))
}

// RandAccountUnlock returns a random account unlock.
func RandAccountUnlock() *iotago.AccountUnlock {
	return &iotago.AccountUnlock{Reference: uint16(rand.Intn(1000))}
}

// RandAnchorUnlock returns a random anchor unlock.
func RandAnchorUnlock() *iotago.AnchorUnlock {
	return &iotago.AnchorUnlock{Reference: uint16(rand.Intn(1000))}
}

// RandNFTUnlock returns a random account unlock.
func RandNFTUnlock() *iotago.NFTUnlock {
	return &iotago.NFTUnlock{Reference: uint16(rand.Intn(1000))}
}

// RandMultiUnlock returns a random multi unlock.
func RandMultiUnlock() *iotago.MultiUnlock {
	unlockCnt := RandInt(10) + 1
	unlocks := make([]iotago.Unlock, 0, unlockCnt)

	for i := 0; i < unlockCnt; i++ {
		unlocks = append(unlocks, RandUnlock(true))
	}

	return &iotago.MultiUnlock{
		Unlocks: unlocks,
	}
}

// ReferenceUnlock returns a reference unlock with the given index.
func ReferenceUnlock(index uint16) *iotago.ReferenceUnlock {
	return &iotago.ReferenceUnlock{Reference: index}
}

// RandTransaction returns a random transaction essence.
func RandTransaction(api iotago.API) *iotago.Transaction {
	return RandTransactionWithOptions(
		api,
		WithUTXOInputCount(rand.Intn(iotago.MaxInputsCount)+1),
		WithOutputCount(rand.Intn(iotago.MaxOutputsCount)+1),
		WithAllotmentCount(rand.Intn(iotago.MaxAllotmentCount)+1),
	)
}

// RandTransactionWithInputCount returns a random transaction essence with a specific amount of inputs..
func RandTransactionWithInputCount(api iotago.API, inputCount int) *iotago.Transaction {
	return RandTransactionWithOptions(
		api,
		WithUTXOInputCount(inputCount),
		WithOutputCount(rand.Intn(iotago.MaxOutputsCount)+1),
		WithAllotmentCount(rand.Intn(iotago.MaxAllotmentCount)+1),
	)
}

// RandTransactionWithOutputCount returns a random transaction essence with a specific amount of outputs.
func RandTransactionWithOutputCount(api iotago.API, outputCount int) *iotago.Transaction {
	return RandTransactionWithOptions(
		api,
		WithUTXOInputCount(rand.Intn(iotago.MaxInputsCount)+1),
		WithOutputCount(outputCount),
		WithAllotmentCount(rand.Intn(iotago.MaxAllotmentCount)+1),
	)
}

// RandTransactionWithAllotmentCount returns a random transaction essence with a specific amount of outputs.
func RandTransactionWithAllotmentCount(api iotago.API, allotmentCount int) *iotago.Transaction {
	return RandTransactionWithOptions(
		api,
		WithUTXOInputCount(rand.Intn(iotago.MaxInputsCount)+1),
		WithOutputCount(rand.Intn(iotago.MaxOutputsCount)+1),
		WithAllotmentCount(allotmentCount),
	)
}

// RandTransactionWithOptions returns a random transaction essence with options applied.
func RandTransactionWithOptions(api iotago.API, opts ...options.Option[iotago.Transaction]) *iotago.Transaction {
	tx := &iotago.Transaction{
		API: api,
		TransactionEssence: &iotago.TransactionEssence{
			NetworkID:     TestNetworkID,
			ContextInputs: iotago.TxEssenceContextInputs{},
			Inputs:        iotago.TxEssenceInputs{},
			Allotments:    iotago.Allotments{},
			Capabilities:  iotago.TransactionCapabilitiesBitMask{},
		},
		Outputs: iotago.TxEssenceOutputs{},
	}

	inputCount := 1
	for i := inputCount; i > 0; i-- {
		tx.TransactionEssence.Inputs = append(tx.TransactionEssence.Inputs, RandUTXOInput())
	}

	outputCount := 1
	for i := outputCount; i > 0; i-- {
		tx.Outputs = append(tx.Outputs, RandBasicOutput(iotago.AddressEd25519))
	}

	return options.Apply(tx, opts)
}

func WithBlockIssuanceCreditInputCount(inputCount int) options.Option[iotago.Transaction] {
	return func(tx *iotago.Transaction) {
		for i := inputCount; i > 0; i-- {
			tx.TransactionEssence.ContextInputs = append(tx.TransactionEssence.ContextInputs, RandBlockIssuanceCreditInput())
		}
	}
}

func WithRewardInputCount(inputCount uint16) options.Option[iotago.Transaction] {
	return func(tx *iotago.Transaction) {
		for i := inputCount; i > 0; i-- {
			rewardInput := &iotago.RewardInput{
				Index: i,
			}
			tx.TransactionEssence.ContextInputs = append(tx.TransactionEssence.ContextInputs, rewardInput)
		}
	}
}

func WithCommitmentInput() options.Option[iotago.Transaction] {
	return func(tx *iotago.Transaction) {
		tx.TransactionEssence.ContextInputs = append(tx.TransactionEssence.ContextInputs, RandCommitmentInput())
	}
}

func WithUTXOInputCount(inputCount int) options.Option[iotago.Transaction] {
	return func(tx *iotago.Transaction) {
		tx.TransactionEssence.Inputs = make(iotago.TxEssenceInputs, 0, inputCount)

		for i := inputCount; i > 0; i-- {
			tx.TransactionEssence.Inputs = append(tx.TransactionEssence.Inputs, RandUTXOInput())
		}
	}
}

func WithOutputCount(outputCount int) options.Option[iotago.Transaction] {
	return func(tx *iotago.Transaction) {
		tx.Outputs = make(iotago.TxEssenceOutputs, 0, outputCount)

		for i := outputCount; i > 0; i-- {
			tx.Outputs = append(tx.Outputs, RandBasicOutput(iotago.AddressEd25519))
		}
	}
}

func WithAllotmentCount(allotmentCount int) options.Option[iotago.Transaction] {
	return func(tx *iotago.Transaction) {
		tx.Allotments = RandSortAllotment(allotmentCount)
	}
}

func WithInputs(inputs iotago.TxEssenceInputs) options.Option[iotago.Transaction] {
	return func(tx *iotago.Transaction) {
		tx.TransactionEssence.Inputs = inputs
	}
}

func WithContextInputs(inputs iotago.TxEssenceContextInputs) options.Option[iotago.Transaction] {
	return func(tx *iotago.Transaction) {
		tx.TransactionEssence.ContextInputs = inputs
	}
}

func WithAllotments(allotments iotago.TxEssenceAllotments) options.Option[iotago.Transaction] {
	return func(tx *iotago.Transaction) {
		tx.Allotments = allotments
	}
}

// RandTaggedData returns a random tagged data payload.
func RandTaggedData(tag []byte, dataLength ...int) *iotago.TaggedData {
	var data []byte
	switch {
	case len(dataLength) > 0:
		data = RandBytes(dataLength[0])
	default:
		data = RandBytes(rand.Intn(200) + 1)
	}

	return &iotago.TaggedData{Tag: tag, Data: data}
}

func RandAccountID() iotago.AccountID {
	alias := iotago.AccountID{}
	copy(alias[:], RandBytes(iotago.AccountIDLength))

	return alias
}

func RandDelegationID() iotago.DelegationID {
	delegation := iotago.DelegationID{}
	copy(delegation[:], RandBytes(iotago.DelegationIDLength))

	return delegation
}

func RandSlot() iotago.SlotIndex {
	return iotago.SlotIndex(RandUint32(uint32(iotago.MaxSlotIndex)))
}

func RandEpoch() iotago.EpochIndex {
	return iotago.EpochIndex(RandUint32(uint32(iotago.MaxEpochIndex)))
}

func RandDuration() time.Duration {
	return time.Duration(RandInt64(math.MaxInt64))
}

// RandBlockID produces a random block ID.
func RandBlockID() iotago.BlockID {
	return Rand36ByteArray()
}

// RandProtocolBlock returns a random block with the given inner payload.
func RandProtocolBlock(block iotago.Block, api iotago.API, rmc iotago.Mana) *iotago.ProtocolBlock {
	if basicBlock, isBasic := block.(*iotago.BasicBlock); isBasic {
		burnedMana, err := basicBlock.ManaCost(rmc, api.ProtocolParameters().WorkScoreParameters())
		if err != nil {
			panic(err)
		}
		basicBlock.MaxBurnedMana = burnedMana

		return &iotago.ProtocolBlock{
			API: api,
			BlockHeader: iotago.BlockHeader{
				ProtocolVersion:  TestAPI.Version(),
				IssuingTime:      RandUTCTime(),
				SlotCommitmentID: iotago.NewEmptyCommitment(api.ProtocolParameters().Version()).MustID(),
				IssuerID:         RandAccountID(),
			},
			Block:     basicBlock,
			Signature: RandEd25519Signature(),
		}
	}

	return &iotago.ProtocolBlock{
		API: api,
		BlockHeader: iotago.BlockHeader{
			ProtocolVersion:  TestAPI.Version(),
			IssuingTime:      RandUTCTime(),
			SlotCommitmentID: iotago.NewEmptyCommitment(api.ProtocolParameters().Version()).MustID(),
			IssuerID:         RandAccountID(),
		},
		Block:     block,
		Signature: RandEd25519Signature(),
	}
}

func RandBasicBlock(api iotago.API, withPayloadType iotago.PayloadType) *iotago.BasicBlock {
	var payload iotago.ApplicationPayload

	//nolint:exhaustive
	switch withPayloadType {
	case iotago.PayloadSignedTransaction:
		payload = RandSignedTransaction(api)
	case iotago.PayloadTaggedData:
		payload = RandTaggedData([]byte("tag"))
	case iotago.PayloadCandidacyAnnouncement:
		payload = &iotago.CandidacyAnnouncement{}
	}

	return &iotago.BasicBlock{
		API:                api,
		StrongParents:      SortedRandBlockIDs(1 + rand.Intn(iotago.BlockMaxParents)),
		WeakParents:        iotago.BlockIDs{},
		ShallowLikeParents: iotago.BlockIDs{},
		Payload:            payload,
		MaxBurnedMana:      RandMana(1000),
	}
}

func RandValidationBlock(api iotago.API) *iotago.ValidationBlock {
	return &iotago.ValidationBlock{
		API:                     api,
		StrongParents:           SortedRandBlockIDs(1 + rand.Intn(iotago.BlockTypeValidationMaxParents)),
		WeakParents:             iotago.BlockIDs{},
		ShallowLikeParents:      iotago.BlockIDs{},
		HighestSupportedVersion: TestAPI.Version() + 1,
	}
}

func RandBasicBlockWithIssuerAndRMC(api iotago.API, issuerID iotago.AccountID, rmc iotago.Mana) *iotago.ProtocolBlock {
	basicBlock := RandBasicBlock(api, iotago.PayloadSignedTransaction)

	block := RandProtocolBlock(basicBlock, TestAPI, rmc)
	block.IssuerID = issuerID

	return block
}

// RandSignedTransactionWithTransaction returns a random transaction with a specific essence.
func RandSignedTransactionWithTransaction(api iotago.API, transaction *iotago.Transaction) *iotago.SignedTransaction {
	sigTxPayload := &iotago.SignedTransaction{API: api}
	sigTxPayload.Transaction = transaction

	unlocksCount := len(transaction.TransactionEssence.Inputs)
	for i := unlocksCount; i > 0; i-- {
		sigTxPayload.Unlocks = append(sigTxPayload.Unlocks, RandEd25519SignatureUnlock())
	}

	return sigTxPayload
}

// RandSignedTransaction returns a random transaction.
func RandSignedTransaction(api iotago.API) *iotago.SignedTransaction {
	return RandSignedTransactionWithTransaction(api, RandTransaction(api))
}

// RandSignedTransactionWithUTXOInputCount returns a random transaction with a specific amount of inputs.
func RandSignedTransactionWithUTXOInputCount(api iotago.API, inputCount int) *iotago.SignedTransaction {
	return RandSignedTransactionWithTransaction(api, RandTransactionWithInputCount(api, inputCount))
}

// RandSignedTransactionWithOutputCount returns a random transaction with a specific amount of outputs.
func RandSignedTransactionWithOutputCount(api iotago.API, outputCount int) *iotago.SignedTransaction {
	return RandSignedTransactionWithTransaction(api, RandTransactionWithOutputCount(api, outputCount))
}

// RandSignedTransactionWithAllotmentCount returns a random transaction with a specific amount of allotments.
func RandSignedTransactionWithAllotmentCount(api iotago.API, allotmentCount int) *iotago.SignedTransaction {
	return RandSignedTransactionWithTransaction(api, RandTransactionWithAllotmentCount(api, allotmentCount))
}

// RandSignedTransactionWithInputOutputCount returns a random transaction with a specific amount of inputs and outputs.
func RandSignedTransactionWithInputOutputCount(api iotago.API, inputCount int, outputCount int) *iotago.SignedTransaction {
	return RandSignedTransactionWithTransaction(api, RandTransactionWithOptions(api, WithUTXOInputCount(inputCount), WithOutputCount(outputCount)))
}

// RandUTXOInput returns a random UTXO input.
func RandUTXOInput() *iotago.UTXOInput {
	return RandUTXOInputWithIndex(uint16(rand.Intn(iotago.RefUTXOIndexMax)))
}

// RandCommitmentInput returns a random Commitment input.
func RandCommitmentInput() *iotago.CommitmentInput {
	return &iotago.CommitmentInput{
		CommitmentID: Rand36ByteArray(),
	}
}

// RandBlockIssuanceCreditInput returns a random BlockIssuanceCreditInput.
func RandBlockIssuanceCreditInput() *iotago.BlockIssuanceCreditInput {
	return &iotago.BlockIssuanceCreditInput{
		AccountID: RandAccountID(),
	}
}

// RandUTXOInputWithIndex returns a random UTXO input with a specific index.
func RandUTXOInputWithIndex(index uint16) *iotago.UTXOInput {
	utxoInput := &iotago.UTXOInput{}
	txID := RandBytes(iotago.TransactionIDLength)
	copy(utxoInput.TransactionID[:], txID)

	utxoInput.TransactionOutputIndex = index

	return utxoInput
}

// RandBasicOutput returns a random basic output (with no features).
func RandBasicOutput(addrType iotago.AddressType) *iotago.BasicOutput {
	dep := &iotago.BasicOutput{
		Amount:     0,
		Conditions: iotago.BasicOutputUnlockConditions{},
		Features:   iotago.BasicOutputFeatures{},
	}

	//nolint:exhaustive
	switch addrType {
	case iotago.AddressEd25519:
		dep.Conditions = iotago.BasicOutputUnlockConditions{&iotago.AddressUnlockCondition{Address: RandEd25519Address()}}
	default:
		panic(fmt.Sprintf("invalid addr type: %d", addrType))
	}

	dep.Amount = RandBaseToken(10000) + 1

	return dep
}

// RandAllotment returns a random Allotment.
func RandAllotment() *iotago.Allotment {
	return &iotago.Allotment{
		AccountID: RandAccountID(),
		Mana:      RandMana(10000) + 1,
	}
}

// RandSortAllotment returns count sorted Allotments.
func RandSortAllotment(count int) iotago.Allotments {
	var allotments iotago.Allotments
	for i := 0; i < count; i++ {
		allotments = append(allotments, RandAllotment())
	}
	allotments.Sort()

	return allotments
}

// OneInputOutputTransaction generates a random transaction with one input and output.
func OneInputOutputTransaction() *iotago.SignedTransaction {
	return &iotago.SignedTransaction{
		API: TestAPI,
		Transaction: &iotago.Transaction{
			API: TestAPI,
			TransactionEssence: &iotago.TransactionEssence{
				NetworkID:     14147312347886322761,
				ContextInputs: iotago.TxEssenceContextInputs{},
				Inputs: iotago.TxEssenceInputs{
					&iotago.UTXOInput{
						TransactionID: func() iotago.TransactionID {
							var b iotago.TransactionID
							copy(b[:], RandBytes(iotago.TransactionIDLength))

							return b
						}(),
						TransactionOutputIndex: 0,
					},
				},
				Allotments:   iotago.Allotments{},
				Capabilities: iotago.TransactionCapabilitiesBitMask{},
				Payload:      nil,
			},
			Outputs: iotago.TxEssenceOutputs{
				&iotago.BasicOutput{
					Amount: 1337,
					Conditions: iotago.BasicOutputUnlockConditions{
						&iotago.AddressUnlockCondition{Address: RandEd25519Address()},
					},
				},
			},
		},
		Unlocks: iotago.Unlocks{
			&iotago.SignatureUnlock{
				Signature: RandEd25519Signature(),
			},
		},
	}
}

// RandEd25519PrivateKey returns a random Ed25519 private key.
func RandEd25519PrivateKey() ed25519.PrivateKey {
	seed := RandEd25519Seed()

	return ed25519.NewKeyFromSeed(seed[:])
}

// RandomBlockIsssuerKeysEd25519 returns count random block issuer keys.
func RandomBlockIsssuerKeysEd25519(count int) iotago.BlockIssuerKeys {
	blockIssuerKeys := make(iotago.BlockIssuerKeys, 0, count)
	for i := 0; i < count; i++ {
		blockIssuerKeys = append(blockIssuerKeys, iotago.Ed25519PublicKeyBlockIssuerKeyFromPublicKey(Rand32ByteArray()))
	}
	blockIssuerKeys.Sort()

	return blockIssuerKeys
}

// RandEd25519Seed returns a random Ed25519 seed.
func RandEd25519Seed() [ed25519.SeedSize]byte {
	var b [ed25519.SeedSize]byte
	read, err := cryptorand.Read(b[:])
	if read != ed25519.SeedSize {
		panic(fmt.Sprintf("could not read %d required bytes from secure RNG", ed25519.SeedSize))
	}
	if err != nil {
		panic(err)
	}

	return b
}

// RandEd25519Identity produces a random Ed25519 identity.
func RandEd25519Identity() (ed25519.PrivateKey, *iotago.Ed25519Address, iotago.AddressKeys) {
	edSk := RandEd25519PrivateKey()
	//nolint:forcetypeassert // we can safely assume that this is an ed25519.PublicKey
	edAddr := iotago.Ed25519AddressFromPubKey(edSk.Public().(ed25519.PublicKey))
	addrKeys := iotago.NewAddressKeysForEd25519Address(edAddr, edSk)

	return edSk, edAddr, addrKeys
}

// RandImplicitAccountIdentity produces a random Implicit Account identity.
func RandImplicitAccountIdentity() (ed25519.PrivateKey, *iotago.ImplicitAccountCreationAddress, iotago.AddressKeys) {
	edSk := RandEd25519PrivateKey()
	//nolint:forcetypeassert // we can safely assume that this is an ed25519.PublicKey
	implicitAccAddr := iotago.ImplicitAccountCreationAddressFromPubKey(edSk.Public().(ed25519.PublicKey))
	addrKeys := iotago.NewAddressKeysForImplicitAccountCreationAddress(implicitAccAddr, edSk)

	return edSk, implicitAccAddr, addrKeys
}

// RandEd25519IdentitiesSortedByAddress returns random Ed25519 identities and keys lexically sorted by the address.
func RandEd25519IdentitiesSortedByAddress(count int) ([]iotago.Address, []iotago.AddressKeys) {
	addresses := make([]iotago.Address, count)
	addressKeys := make([]iotago.AddressKeys, count)
	for i := 0; i < count; i++ {
		_, addresses[i], addressKeys[i] = RandEd25519Identity()
	}

	// addressses need to be lexically ordered in the MultiAddress
	slices.SortFunc(addresses, func(a iotago.Address, b iotago.Address) int {
		return bytes.Compare(a.ID(), b.ID())
	})

	// addressses need to be lexically ordered in the MultiAddress
	slices.SortFunc(addressKeys, func(a iotago.AddressKeys, b iotago.AddressKeys) int {
		return bytes.Compare(a.Address.ID(), b.Address.ID())
	})

	return addresses, addressKeys
}

// RandStorageScoreParameters produces random set of  parameters.
func RandStorageScoreParameters() *iotago.StorageScoreParameters {
	return &iotago.StorageScoreParameters{
		StorageCost:                 RandBaseToken(iotago.MaxBaseToken),
		FactorData:                  iotago.StorageScoreFactor(RandUint8(math.MaxUint8)),
		OffsetOutputOverhead:        iotago.StorageScore(RandUint64(math.MaxUint64)),
		OffsetEd25519BlockIssuerKey: iotago.StorageScore(RandUint64(math.MaxUint64)),
		OffsetStakingFeature:        iotago.StorageScore(RandUint64(math.MaxUint64)),
	}
}

// RandWorkScore returns a random workscore.
func RandWorkScore(max uint32) iotago.WorkScore {
	return iotago.WorkScore(RandUint32(max))
}

// RandWorkScoreParameters produces random workscore structure.
func RandWorkScoreParameters() *iotago.WorkScoreParameters {
	return &iotago.WorkScoreParameters{
		DataByte:         RandWorkScore(math.MaxUint32),
		Block:            RandWorkScore(math.MaxUint32),
		Input:            RandWorkScore(math.MaxUint32),
		ContextInput:     RandWorkScore(math.MaxUint32),
		Output:           RandWorkScore(math.MaxUint32),
		NativeToken:      RandWorkScore(math.MaxUint32),
		Staking:          RandWorkScore(math.MaxUint32),
		BlockIssuer:      RandWorkScore(math.MaxUint32),
		Allotment:        RandWorkScore(math.MaxUint32),
		SignatureEd25519: RandWorkScore(math.MaxUint32),
	}
}

// RandProtocolParameters produces random protocol parameters.
func RandProtocolParameters() iotago.ProtocolParameters {
	return iotago.NewV3ProtocolParameters(
		iotago.WithNetworkOptions(
			RandString(255),
			iotago.NetworkPrefix(RandString(255)),
		),
		iotago.WithSupplyOptions(
			RandBaseToken(iotago.MaxBaseToken),
			RandBaseToken(iotago.MaxBaseToken),
			iotago.StorageScoreFactor(RandUint8(math.MaxUint8)),
			iotago.StorageScore(RandUint64(math.MaxUint64)),
			iotago.StorageScore(RandUint64(math.MaxUint64)),
			iotago.StorageScore(RandUint64(math.MaxUint64)),
			iotago.StorageScore(RandUint64(math.MaxUint64)),
		),
		iotago.WithWorkScoreOptions(
			RandWorkScore(math.MaxUint32),
			RandWorkScore(math.MaxUint32),
			RandWorkScore(math.MaxUint32),
			RandWorkScore(math.MaxUint32),
			RandWorkScore(math.MaxUint32),
			RandWorkScore(math.MaxUint32),
			RandWorkScore(math.MaxUint32),
			RandWorkScore(math.MaxUint32),
			RandWorkScore(math.MaxUint32),
			RandWorkScore(math.MaxUint32),
		),
		iotago.WithTimeProviderOptions(time.Now().Unix(), RandUint8(math.MaxUint8), RandUint8(math.MaxUint8)),
		iotago.WithLivenessOptions(RandUint16(math.MaxUint16), RandUint16(math.MaxUint16), RandSlot(), RandSlot(), RandSlot()),
		iotago.WithCongestionControlOptions(
			RandMana(iotago.MaxMana),
			RandMana(iotago.MaxMana),
			RandMana(iotago.MaxMana),
			RandWorkScore(math.MaxUint32),
			RandWorkScore(math.MaxUint32),
			RandWorkScore(math.MaxUint32),
			RandUint32(math.MaxUint32),
			RandUint32(math.MaxUint32),
		),
	)
}

// ManaDecayFactors calculates mana decay factors that can be used in the tests.
func ManaDecayFactors(betaPerYear float64, slotsPerEpoch int, slotTimeSeconds int, decayFactorsExponent uint64) []uint32 {
	epochsPerYear := ((365.0 * 24.0 * 60.0 * 60.0) / float64(slotTimeSeconds)) / float64(slotsPerEpoch)
	decayFactors := make([]uint32, int(epochsPerYear))

	betaPerEpochIndex := betaPerYear / epochsPerYear

	for epoch := 1; epoch <= int(epochsPerYear); epoch++ {
		decayFactor := math.Exp(-betaPerEpochIndex*float64(epoch)) * (math.Pow(2, float64(decayFactorsExponent)))
		decayFactors[epoch-1] = uint32(decayFactor)
	}

	return decayFactors
}

// ManaDecayFactorEpochsSum calculates mana decay factor epochs sum parameter that can be used in the tests.
func ManaDecayFactorEpochsSum(betaPerYear float64, slotsPerEpoch int, slotTimeSeconds int, decayFactorEpochsSumExponent uint64) uint32 {
	delta := float64(slotsPerEpoch) * (1.0 / (365.0 * 24.0 * 60.0 * 60.0)) * float64(slotTimeSeconds)

	return uint32((math.Exp(-betaPerYear*delta) / (1 - math.Exp(-betaPerYear*delta)) * (math.Pow(2, float64(decayFactorEpochsSumExponent)))))
}
