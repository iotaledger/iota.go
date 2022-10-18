package tpkg

import (
	"bytes"
	"crypto/ed25519"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/iotaledger/hive.go/serializer/v2"

	legacy "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/trinary"
	iotago "github.com/iotaledger/iota.go/v3"
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

// RandFloat64 returns a random float64.
func RandFloat64(max float64) float64 {
	return rand.Float64() * max
}

// RandTrytes returns length amount of random trytes.
func RandTrytes(length int) trinary.Trytes {
	var trytes strings.Builder
	for i := 0; i < length; i++ {
		trytes.WriteByte(legacy.TryteAlphabet[rand.Intn(len(legacy.TryteAlphabet))])
	}
	return trytes.String()
}

func RandOutputID(index uint16) iotago.OutputID {
	var outputID iotago.OutputID
	_, err := rand.Read(outputID[:iotago.TransactionIDLength])
	if err != nil {
		panic(err)
	}
	binary.LittleEndian.PutUint16(outputID[iotago.TransactionIDLength:], index)
	return outputID
}

func RandOutputIDs(count uint16) iotago.OutputIDs {
	outputIDs := make(iotago.OutputIDs, int(count))
	for i := 0; i < int(count); i++ {
		outputIDs[i] = RandOutputID(count)
	}
	return outputIDs
}

func RandTransactionID() iotago.TransactionID {
	var transactionID iotago.TransactionID
	_, err := rand.Read(transactionID[:iotago.TransactionIDLength])
	if err != nil {
		panic(err)
	}
	return transactionID
}

// RandNativeToken returns a random NativeToken.
func RandNativeToken() *iotago.NativeToken {
	b := RandBytes(iotago.NativeTokenIDLength)
	nt := &iotago.NativeToken{Amount: RandUint256()}
	copy(nt.ID[:], b)
	return nt
}

// RandSortNativeTokens returns count sorted NativeToken.
func RandSortNativeTokens(count int) iotago.NativeTokens {
	var nativeTokens iotago.NativeTokens
	for i := 0; i < count; i++ {
		nativeTokens = append(nativeTokens, RandNativeToken())
	}
	sort.Slice(nativeTokens, func(i, j int) bool {
		return bytes.Compare(nativeTokens[i].ID[:], nativeTokens[j].ID[:]) == -1
	})
	return nativeTokens
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

// SortedRandBlockIDs returned random block IDs.
func SortedRandBlockIDs(count int) iotago.BlockIDs {
	return iotago.BlockIDs(SortedRandMSParents(count))
}

// SortedRandMSParents returns random milestone parents IDs.
func SortedRandMSParents(count int) iotago.MilestoneParentIDs {
	slice := make(iotago.MilestoneParentIDs, count)
	for i, ele := range SortedRand32ByteArray(count) {
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

// RandAliasAddress returns a random AliasAddress.
func RandAliasAddress() *iotago.AliasAddress {
	aliasAddr := &iotago.AliasAddress{}
	addr := RandBytes(iotago.AliasAddressBytesLength)
	copy(aliasAddr[:], addr)
	return aliasAddr
}

// RandNFTAddress returns a random NFTAddress.
func RandNFTAddress() *iotago.NFTAddress {
	nftAddr := &iotago.NFTAddress{}
	addr := RandBytes(iotago.NFTAddressBytesLength)
	copy(nftAddr[:], addr)
	return nftAddr
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

// RandEd25519SignatureUnlock returns a random Ed25519 signature unlock.
func RandEd25519SignatureUnlock() *iotago.SignatureUnlock {
	return &iotago.SignatureUnlock{Signature: RandEd25519Signature()}
}

// RandReferenceUnlock returns a random reference unlock.
func RandReferenceUnlock() *iotago.ReferenceUnlock {
	return ReferenceUnlock(uint16(rand.Intn(1000)))
}

// RandAliasUnlock returns a random alias unlock.
func RandAliasUnlock() *iotago.AliasUnlock {
	return &iotago.AliasUnlock{Reference: uint16(rand.Intn(1000))}
}

// RandNFTUnlock returns a random alias unlock.
func RandNFTUnlock() *iotago.NFTUnlock {
	return &iotago.NFTUnlock{Reference: uint16(rand.Intn(1000))}
}

// ReferenceUnlock returns a reference unlock with the given index.
func ReferenceUnlock(index uint16) *iotago.ReferenceUnlock {
	return &iotago.ReferenceUnlock{Reference: index}
}

// RandTransactionEssence returns a random transaction essence.
func RandTransactionEssence() *iotago.TransactionEssence {
	return RandTransactionEssenceWithInputOutputCount(rand.Intn(iotago.MaxInputsCount)+1, rand.Intn(iotago.MaxOutputsCount)+1)
}

// RandTransactionEssenceWithInputCount returns a random transaction essence with a specific amount of inputs..
func RandTransactionEssenceWithInputCount(inputCount int) *iotago.TransactionEssence {
	return RandTransactionEssenceWithInputOutputCount(inputCount, rand.Intn(iotago.MaxOutputsCount)+1)
}

// RandTransactionEssenceWithOutputCount returns a random transaction essence with a specific amount of outputs.
func RandTransactionEssenceWithOutputCount(outputCount int) *iotago.TransactionEssence {
	return RandTransactionEssenceWithInputOutputCount(rand.Intn(iotago.MaxInputsCount)+1, outputCount)
}

// RandTransactionEssenceWithInputOutputCount returns a random transaction essence with a specific amount of inputs and outputs.
func RandTransactionEssenceWithInputOutputCount(inputCount int, outputCount int) *iotago.TransactionEssence {
	tx := &iotago.TransactionEssence{
		NetworkID: TestNetworkID,
	}

	for i := inputCount; i > 0; i-- {
		tx.Inputs = append(tx.Inputs, RandUTXOInput())
	}

	for i := outputCount; i > 0; i-- {
		tx.Outputs = append(tx.Outputs, RandBasicOutput(iotago.AddressEd25519))
	}

	return tx
}

// RandTransactionEssenceWithInputs returns a random transaction essence with a specific slice of inputs.
func RandTransactionEssenceWithInputs(inputs iotago.Inputs[iotago.TxEssenceInput]) *iotago.TransactionEssence {
	tx := &iotago.TransactionEssence{
		NetworkID: TestNetworkID,
	}

	tx.Inputs = inputs

	outputCount := rand.Intn(iotago.MaxOutputsCount) + 1
	for i := outputCount; i > 0; i-- {
		tx.Outputs = append(tx.Outputs, RandBasicOutput(iotago.AddressEd25519))
	}

	return tx
}

// RandMigratedFundsEntry returns a random migrated funds entry.
func RandMigratedFundsEntry() *iotago.MigratedFundsEntry {
	return &iotago.MigratedFundsEntry{
		TailTransactionHash: Rand49ByteArray(),
		Address:             RandEd25519Address(),
		Deposit:             rand.Uint64(),
	}
}

// RandReceipt returns a random receipt.
func RandReceipt() *iotago.ReceiptMilestoneOpt {
	receipt := &iotago.ReceiptMilestoneOpt{MigratedAt: 1000, Final: true}

	migFundsEntriesCount := rand.Intn(10) + 1
	for i := migFundsEntriesCount; i > 0; i-- {
		receipt.Funds = append(receipt.Funds, RandMigratedFundsEntry())
	}
	receipt.SortFunds()
	receipt.Transaction = RandTreasuryTransaction()

	return receipt
}

// RandMilestone returns a random milestone with the given parent blocks.
func RandMilestone(parents iotago.MilestoneParentIDs) *iotago.Milestone {
	const sigsCount = 3

	if parents == nil {
		parents = SortedRandMSParents(1 + rand.Intn(7))
	}

	msPayload := &iotago.Milestone{
		MilestoneEssence: iotago.MilestoneEssence{
			Index:               iotago.MilestoneIndex(rand.Intn(1000)),
			Timestamp:           uint32(time.Now().Unix()),
			PreviousMilestoneID: Rand32ByteArray(),
			Parents:             iotago.MilestoneParentIDs(parents),
			InclusionMerkleRoot: func() iotago.MilestoneMerkleProof {
				var b iotago.MilestoneMerkleProof
				copy(b[:], RandBytes(iotago.MilestoneMerkleProofLength))
				return b
			}(),
			AppliedMerkleRoot: func() iotago.MilestoneMerkleProof {
				var b iotago.MilestoneMerkleProof
				copy(b[:], RandBytes(iotago.MilestoneMerkleProofLength))
				return b
			}(),
			Metadata: RandBytes(10),
			Opts: iotago.MilestoneOpts{
				&iotago.ProtocolParamsMilestoneOpt{
					TargetMilestoneIndex: 100,
					ProtocolVersion:      2,
					Params:               RandBytes(200),
				},
			},
		},
		Signatures: func() iotago.Signatures[iotago.MilestoneSignature] {
			msSigs := make(iotago.Signatures[iotago.MilestoneSignature], sigsCount)
			for i := 0; i < sigsCount; i++ {
				msSigs[i] = RandEd25519Signature()
			}
			sort.Slice(msSigs, func(i, j int) bool {
				return bytes.Compare(msSigs[i].(*iotago.Ed25519Signature).PublicKey[:], msSigs[j].(*iotago.Ed25519Signature).PublicKey[:]) == -1
			})
			return msSigs
		}(),
	}

	return msPayload
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

// RandBlockID produces a random block ID.
func RandBlockID() iotago.BlockID {
	return Rand32ByteArray()
}

// RandBlock returns a random block with the given inner payload.
func RandBlock(withPayloadType iotago.PayloadType) *iotago.Block {
	var payload iotago.Payload

	parents := SortedRandMSParents(1 + rand.Intn(7))

	switch withPayloadType {
	case iotago.PayloadTransaction:
		payload = RandTransaction()
	case iotago.PayloadTaggedData:
		payload = RandTaggedData([]byte("tag"))
	case iotago.PayloadMilestone:
		payload = RandMilestone(parents)
	}

	return &iotago.Block{
		ProtocolVersion: TestProtocolVersion,
		Parents:         iotago.BlockIDs(parents),
		Payload:         payload,
		Nonce:           uint64(rand.Intn(1000)),
	}
}

// RandTransactionWithEssence returns a random transaction with a specific essence.
func RandTransactionWithEssence(essence *iotago.TransactionEssence) *iotago.Transaction {
	sigTxPayload := &iotago.Transaction{}
	sigTxPayload.Essence = essence

	unlocksCount := len(essence.Inputs)
	for i := unlocksCount; i > 0; i-- {
		sigTxPayload.Unlocks = append(sigTxPayload.Unlocks, RandEd25519SignatureUnlock())
	}

	return sigTxPayload
}

// RandTransaction returns a random transaction.
func RandTransaction() *iotago.Transaction {
	return RandTransactionWithEssence(RandTransactionEssence())
}

// RandTransactionWithInputCount returns a random transaction with a specific amount of inputs.
func RandTransactionWithInputCount(inputCount int) *iotago.Transaction {
	return RandTransactionWithEssence(RandTransactionEssenceWithInputCount(inputCount))
}

// RandTransactionWithOutputCount returns a random transaction with a specific amount of outputs.
func RandTransactionWithOutputCount(outputCount int) *iotago.Transaction {
	return RandTransactionWithEssence(RandTransactionEssenceWithOutputCount(outputCount))
}

// RandTransactionWithInputOutputCount returns a random transaction with a specific amount of inputs and outputs.
func RandTransactionWithInputOutputCount(inputCount int, outputCount int) *iotago.Transaction {
	return RandTransactionWithEssence(RandTransactionEssenceWithInputOutputCount(inputCount, outputCount))
}

// RandTreasuryInput returns a random treasury input.
func RandTreasuryInput() *iotago.TreasuryInput {
	treasuryInput := &iotago.TreasuryInput{}
	input := RandBytes(iotago.TreasuryInputBytesLength)
	copy(treasuryInput[:], input)
	return treasuryInput
}

// RandUTXOInput returns a random UTXO input.
func RandUTXOInput() *iotago.UTXOInput {
	return RandUTXOInputWithIndex(uint16(rand.Intn(iotago.RefUTXOIndexMax)))
}

// RandUTXOInputWithIndex returns a random UTXO input with a specific index.
func RandUTXOInputWithIndex(index uint16) *iotago.UTXOInput {
	utxoInput := &iotago.UTXOInput{}
	txID := RandBytes(iotago.TransactionIDLength)
	copy(utxoInput.TransactionID[:], txID)

	utxoInput.TransactionOutputIndex = index
	return utxoInput
}

// RandTreasuryOutput returns a random treasury output.
func RandTreasuryOutput() *iotago.TreasuryOutput {
	return &iotago.TreasuryOutput{Amount: rand.Uint64()}
}

// RandTreasuryTransaction returns a random treasury transaction.
func RandTreasuryTransaction() *iotago.TreasuryTransaction {
	return &iotago.TreasuryTransaction{
		Input:  RandTreasuryInput(),
		Output: RandTreasuryOutput(),
	}
}

// RandBasicOutput returns a random basic output (with no features).
func RandBasicOutput(addrType iotago.AddressType) *iotago.BasicOutput {
	dep := &iotago.BasicOutput{
		Amount:       0,
		NativeTokens: nil,
		Conditions:   nil,
		Features:     nil,
	}

	switch addrType {
	case iotago.AddressEd25519:
		dep.Conditions = iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{&iotago.AddressUnlockCondition{Address: RandEd25519Address()}}
	default:
		panic(fmt.Sprintf("invalid addr type: %d", addrType))
	}

	amount := uint64(rand.Intn(10000) + 1)
	dep.Amount = amount
	return dep
}

// OneInputOutputTransaction generates a random transaction with one input and output.
func OneInputOutputTransaction() *iotago.Transaction {
	return &iotago.Transaction{
		Essence: &iotago.TransactionEssence{
			NetworkID: 14147312347886322761,
			Inputs: iotago.Inputs[iotago.TxEssenceInput]{
				&iotago.UTXOInput{
					TransactionID: func() [iotago.TransactionIDLength]byte {
						var b [iotago.TransactionIDLength]byte
						copy(b[:], RandBytes(iotago.TransactionIDLength))
						return b
					}(),
					TransactionOutputIndex: 0,
				},
			},
			Outputs: iotago.Outputs[iotago.TxEssenceOutput]{
				&iotago.BasicOutput{
					Amount: 1337,
					Conditions: iotago.UnlockConditions[iotago.BasicOutputUnlockCondition]{
						&iotago.AddressUnlockCondition{Address: RandEd25519Address()},
					},
				},
			},
			Payload: nil,
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

// RandEd25519Seed returns a random Ed25519 seed.
func RandEd25519Seed() [ed25519.SeedSize]byte {
	var b [ed25519.SeedSize]byte
	read, err := rand.Read(b[:])
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
	edAddr := iotago.Ed25519AddressFromPubKey(edSk.Public().(ed25519.PublicKey))
	addrKeys := iotago.NewAddressKeysForEd25519Address(&edAddr, edSk)
	return edSk, &edAddr, addrKeys
}

// RandMilestoneID produces a random milestone ID.
func RandMilestoneID() iotago.MilestoneID {
	return Rand32ByteArray()
}

// RandMilestoneMerkleProof produces a random milestone merkle proof.
func RandMilestoneMerkleProof() iotago.MilestoneMerkleProof {
	return Rand32ByteArray()
}

// RandRentStructure produces random rent structure.
func RandRentStructure() *iotago.RentStructure {
	return &iotago.RentStructure{
		VByteCost:    RandUint32(math.MaxUint32),
		VBFactorData: iotago.VByteCostFactor(RandUint8(math.MaxUint8)),
		VBFactorKey:  iotago.VByteCostFactor(RandUint8(math.MaxUint8)),
	}
}

// RandProtocolParameters produces random protocol parameters.
func RandProtocolParameters() *iotago.ProtocolParameters {
	return &iotago.ProtocolParameters{
		Version:       RandByte(),
		NetworkName:   RandString(255),
		Bech32HRP:     iotago.NetworkPrefix(RandString(255)),
		MinPoWScore:   RandUint32(50000),
		BelowMaxDepth: RandUint8(math.MaxUint8),
		RentStructure: *RandRentStructure(),
		TokenSupply:   RandUint64(math.MaxUint64),
	}
}
