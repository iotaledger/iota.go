package tpkg

import (
	"bytes"
	"crypto/ed25519"
	"encoding/binary"
	"fmt"
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

// RandBytes returns length amount random bytes.
func RandBytes(length int) []byte {
	var b []byte
	for i := 0; i < length; i++ {
		b = append(b, byte(rand.Intn(256)))
	}
	return b
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
	seris := nativeTokens.ToSerializables()
	sort.Sort(serializer.SortedSerializables(seris))
	nativeTokens.FromSerializables(seris)
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

// SortedRand32BytArray returns a count length slice of sorted 32 byte arrays.
func SortedRand32BytArray(count int) [][32]byte {
	hashes := make(serializer.LexicalOrdered32ByteArrays, count)
	for i := 0; i < count; i++ {
		hashes[i] = Rand32ByteArray()
	}
	sort.Sort(hashes)
	return hashes
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
	tx := &iotago.TransactionEssence{
		NetworkID: TestNetworkID,
	}

	inputCount := rand.Intn(10) + 1
	for i := inputCount; i > 0; i-- {
		tx.Inputs = append(tx.Inputs, RandUTXOInput())
	}

	outputCount := rand.Intn(10) + 1
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
func RandMilestone(parents iotago.BlockIDs) *iotago.Milestone {
	const sigsCount = 3

	if parents == nil {
		parents = SortedRand32BytArray(1 + rand.Intn(7))
	}

	msPayload := &iotago.Milestone{
		Index:               uint32(rand.Intn(1000)),
		Timestamp:           uint32(time.Now().Unix()),
		PreviousMilestoneID: Rand32ByteArray(),
		Parents:             parents,
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
		Signatures: func() iotago.Signatures {
			msSigs := make(iotago.Signatures, sigsCount)
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

// RandBlock returns a random block with the given inner payload.
func RandBlock(withPayloadType iotago.PayloadType) *iotago.Block {
	var payload iotago.Payload

	parents := SortedRand32BytArray(1 + rand.Intn(7))

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
		Parents:         parents,
		Payload:         payload,
		Nonce:           uint64(rand.Intn(1000)),
	}
}

// RandTransaction returns a random transaction.
func RandTransaction() *iotago.Transaction {
	sigTxPayload := &iotago.Transaction{}
	essence := RandTransactionEssence()
	sigTxPayload.Essence = essence

	unlocksCount := len(essence.Inputs)
	for i := unlocksCount; i > 0; i-- {
		sigTxPayload.Unlocks = append(sigTxPayload.Unlocks, RandEd25519SignatureUnlock())
	}

	return sigTxPayload
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
	utxoInput := &iotago.UTXOInput{}
	txID := RandBytes(iotago.TransactionIDLength)
	copy(utxoInput.TransactionID[:], txID)

	index := uint16(rand.Intn(iotago.RefUTXOIndexMax))
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

// RandBasicOutput returns a random basic output (with no feature blocks).
func RandBasicOutput(addrType iotago.AddressType) *iotago.BasicOutput {
	dep := &iotago.BasicOutput{}

	switch addrType {
	case iotago.AddressEd25519:
		dep.Conditions = iotago.UnlockConditions{&iotago.AddressUnlockCondition{Address: RandEd25519Address()}}
	default:
		panic(fmt.Sprintf("invalid addr type: %d", addrType))
	}

	amount := uint64(rand.Intn(10000))
	dep.Amount = amount
	return dep
}

// OneInputOutputTransaction generates a random transaction with one input and output.
func OneInputOutputTransaction() *iotago.Transaction {
	return &iotago.Transaction{
		Essence: &iotago.TransactionEssence{
			Inputs: iotago.Inputs{
				&iotago.UTXOInput{
					TransactionID: func() [iotago.TransactionIDLength]byte {
						var b [iotago.TransactionIDLength]byte
						copy(b[:], RandBytes(iotago.TransactionIDLength))
						return b
					}(),
					TransactionOutputIndex: 0,
				},
			},
			Outputs: iotago.Outputs{
				&iotago.BasicOutput{
					Amount: 1337,
					Conditions: iotago.UnlockConditions{
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
