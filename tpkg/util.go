package tpkg

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/iotaledger/hive.go/serializer"
	legacy "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/trinary"
	"github.com/iotaledger/iota.go/v2"
	"github.com/iotaledger/iota.go/v2/ed25519"
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

// Rand32ByteArray returns an array with 32 random bytes.
func Rand32ByteArray() [32]byte {
	var h [32]byte
	b := RandBytes(32)
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
func RandEd25519Address() (*iotago.Ed25519Address, []byte) {
	// type
	edAddr := &iotago.Ed25519Address{}
	addr := RandBytes(iotago.Ed25519AddressBytesLength)
	copy(edAddr[:], addr)
	// serialized
	var b [iotago.Ed25519AddressSerializedBytesSize]byte
	b[0] = iotago.AddressEd25519
	copy(b[serializer.SmallTypeDenotationByteSize:], addr)
	return edAddr, b[:]
}

// RandEd25519Signature returns a random Ed25519 signature.
func RandEd25519Signature() (*iotago.Ed25519Signature, []byte) {
	// type
	edSig := &iotago.Ed25519Signature{}
	pub := RandBytes(ed25519.PublicKeySize)
	sig := RandBytes(ed25519.SignatureSize)
	copy(edSig.PublicKey[:], pub)
	copy(edSig.Signature[:], sig)
	// serialized
	var b [iotago.Ed25519SignatureSerializedBytesSize]byte
	b[0] = iotago.SignatureEd25519
	copy(b[serializer.SmallTypeDenotationByteSize:], pub)
	copy(b[serializer.SmallTypeDenotationByteSize+ed25519.PublicKeySize:], sig)
	return edSig, b[:]
}

// RandEd25519SignatureUnlockBlock returns a random Ed25519 signature unlock block.
func RandEd25519SignatureUnlockBlock() (*iotago.SignatureUnlockBlock, []byte) {
	edSig, edSigData := RandEd25519Signature()
	block := &iotago.SignatureUnlockBlock{Signature: edSig}
	return block, append([]byte{iotago.UnlockBlockSignature}, edSigData...)
}

// RandReferenceUnlockBlock returns a random reference unlock block.
func RandReferenceUnlockBlock() (*iotago.ReferenceUnlockBlock, []byte) {
	return ReferenceUnlockBlock(uint16(rand.Intn(1000)))
}

// ReferenceUnlockBlock returns a reference unlock block with the given index.
func ReferenceUnlockBlock(index uint16) (*iotago.ReferenceUnlockBlock, []byte) {
	var b [iotago.ReferenceUnlockBlockSize]byte
	b[0] = iotago.UnlockBlockReference
	binary.LittleEndian.PutUint16(b[serializer.SmallTypeDenotationByteSize:], index)
	return &iotago.ReferenceUnlockBlock{Reference: index}, b[:]
}

// RandTransactionEssence returns a random transaction essence.
func RandTransactionEssence() (*iotago.TransactionEssence, []byte) {
	var buf bytes.Buffer

	tx := &iotago.TransactionEssence{}
	Must(buf.WriteByte(iotago.TransactionEssenceNormal))

	inputsBytes := serializer.LexicalOrderedByteSlices{}
	inputCount := rand.Intn(10) + 1
	Must(binary.Write(&buf, binary.LittleEndian, uint16(inputCount)))
	for i := inputCount; i > 0; i-- {
		_, inputData := RandUTXOInput()
		inputsBytes = append(inputsBytes, inputData)
	}

	sort.Sort(inputsBytes)

	for _, inputData := range inputsBytes {
		_, err := buf.Write(inputData)
		Must(err)
		input := &iotago.UTXOInput{}
		if _, err := input.Deserialize(inputData, serializer.DeSeriModePerformValidation); err != nil {
			panic(err)
		}
		tx.Inputs = append(tx.Inputs, input)
	}

	outputsBytes := serializer.LexicalOrderedByteSlices{}
	outputCount := rand.Intn(10) + 1
	Must(binary.Write(&buf, binary.LittleEndian, uint16(outputCount)))
	for i := outputCount; i > 0; i-- {
		_, depData := RandSigLockedSingleOutput(iotago.AddressEd25519)
		outputsBytes = append(outputsBytes, depData)
	}

	sort.Sort(outputsBytes)
	for _, outputData := range outputsBytes {
		_, err := buf.Write(outputData)
		Must(err)
		output := &iotago.SigLockedSingleOutput{}
		if _, err := output.Deserialize(outputData, serializer.DeSeriModePerformValidation); err != nil {
			panic(err)
		}
		tx.Outputs = append(tx.Outputs, output)
	}

	// empty payload
	Must(binary.Write(&buf, binary.LittleEndian, uint32(0)))

	return tx, buf.Bytes()
}

// RandMigratedFundsEntry returns a random migrated funds entry.
func RandMigratedFundsEntry() (*iotago.MigratedFundsEntry, []byte) {
	tailTxHash := Rand49ByteArray()
	addr, addrBytes := RandEd25519Address()
	deposit := rand.Uint64()

	var b bytes.Buffer
	_, err := b.Write(tailTxHash[:])
	Must(err)
	_, err = b.Write(addrBytes)
	Must(err)
	Must(binary.Write(&b, binary.LittleEndian, deposit))

	return &iotago.MigratedFundsEntry{
		TailTransactionHash: tailTxHash,
		Address:             addr,
		Deposit:             deposit,
	}, b.Bytes()
}

// RandReceipt returns a random receipt.
func RandReceipt() (*iotago.Receipt, []byte) {
	receipt := &iotago.Receipt{MigratedAt: 1000, Final: true}

	var b bytes.Buffer

	Must(binary.Write(&b, binary.LittleEndian, iotago.ReceiptPayloadTypeID))
	Must(binary.Write(&b, binary.LittleEndian, receipt.MigratedAt))
	Must(b.WriteByte(1))

	migFundsEntriesBytes := serializer.LexicalOrderedByteSlices{}
	migFundsEntriesCount := rand.Intn(10) + 1
	Must(binary.Write(&b, binary.LittleEndian, uint16(migFundsEntriesCount)))
	for i := migFundsEntriesCount; i > 0; i-- {
		_, migFundsEntryBytes := RandMigratedFundsEntry()
		migFundsEntriesBytes = append(migFundsEntriesBytes, migFundsEntryBytes)
	}

	sort.Sort(migFundsEntriesBytes)

	for _, migFundEntryBytes := range migFundsEntriesBytes {
		_, err := b.Write(migFundEntryBytes)
		Must(err)
		migFundsEntry := &iotago.MigratedFundsEntry{}
		if _, err := migFundsEntry.Deserialize(migFundEntryBytes, serializer.DeSeriModePerformValidation); err != nil {
			panic(err)
		}
		receipt.Funds = append(receipt.Funds, migFundsEntry)
	}

	randTreasuryTx, randTreasuryTxBytes := RandTreasuryTransaction()
	receipt.Transaction = randTreasuryTx

	Must(binary.Write(&b, binary.LittleEndian, uint32(len(randTreasuryTxBytes))))
	if _, err := b.Write(randTreasuryTxBytes); err != nil {
		Must(err)
	}

	return receipt, b.Bytes()
}

// RandMilestone returns a random milestone with the given parent messages.
func RandMilestone(parents iotago.MessageIDs) (*iotago.Milestone, []byte) {
	inclusionMerkleProof := RandBytes(iotago.MilestoneInclusionMerkleProofLength)
	const sigsCount = 3

	if parents == nil {
		parents = SortedRand32BytArray(1 + rand.Intn(7))
	}

	msPayload := &iotago.Milestone{
		Index:     uint32(rand.Intn(1000)),
		Timestamp: uint64(time.Now().Unix()),
		Parents:   parents,
		InclusionMerkleProof: func() [iotago.MilestoneInclusionMerkleProofLength]byte {
			b := [iotago.MilestoneInclusionMerkleProofLength]byte{}
			copy(b[:], inclusionMerkleProof)
			return b
		}(),
		PublicKeys: func() [][iotago.MilestonePublicKeyLength]byte {
			msPubKeys := make([][iotago.MilestonePublicKeyLength]byte, sigsCount)
			for i := 0; i < sigsCount; i++ {
				msPubKeys[i] = Rand32ByteArray()
				// ensure lexical ordering
				msPubKeys[i][0] = byte(i)
			}
			return msPubKeys
		}(),
		Signatures: func() [][iotago.MilestoneSignatureLength]byte {
			msSigs := make([][iotago.MilestoneSignatureLength]byte, sigsCount)
			for i := 0; i < sigsCount; i++ {
				msSigs[i] = RandMilestoneSig()
			}
			return msSigs
		}(),
	}

	var b bytes.Buffer
	Must(binary.Write(&b, binary.LittleEndian, iotago.MilestonePayloadTypeID))
	Must(binary.Write(&b, binary.LittleEndian, msPayload.Index))
	Must(binary.Write(&b, binary.LittleEndian, msPayload.Timestamp))
	Must(binary.Write(&b, binary.LittleEndian, byte(len(msPayload.Parents))))
	for _, parent := range msPayload.Parents {
		if _, err := b.Write(parent[:]); err != nil {
			panic(err)
		}
	}
	if _, err := b.Write(msPayload.InclusionMerkleProof[:]); err != nil {
		panic(err)
	}

	Must(binary.Write(&b, binary.LittleEndian, msPayload.NextPoWScore))
	Must(binary.Write(&b, binary.LittleEndian, msPayload.NextPoWScoreMilestoneIndex))

	Must(b.WriteByte(sigsCount))
	for _, pubKey := range msPayload.PublicKeys {
		if _, err := b.Write(pubKey[:]); err != nil {
			panic(err)
		}
	}
	Must(binary.Write(&b, binary.LittleEndian, uint32(0)))
	Must(b.WriteByte(sigsCount))
	for _, sig := range msPayload.Signatures {
		if _, err := b.Write(sig[:]); err != nil {
			panic(err)
		}
	}

	return msPayload, b.Bytes()
}

// RandMilestoneSig returns a random milestone signature.
func RandMilestoneSig() [iotago.MilestoneSignatureLength]byte {
	var sig [iotago.MilestoneSignatureLength]byte
	copy(sig[:], RandBytes(iotago.MilestoneSignatureLength))
	return sig
}

// RandIndexation returns a random indexation payload.
func RandIndexation(dataLength ...int) (*iotago.Indexation, []byte) {
	const index = "寿司を作って"

	var data []byte
	switch {
	case len(dataLength) > 0:
		data = RandBytes(dataLength[0])
	default:
		data = RandBytes(rand.Intn(200) + 1)
	}

	indexationPayload := &iotago.Indexation{Index: []byte(index), Data: data}

	var b bytes.Buffer
	Must(binary.Write(&b, binary.LittleEndian, iotago.IndexationPayloadTypeID))

	strLen := uint16(len(index))
	Must(binary.Write(&b, binary.LittleEndian, strLen))

	if _, err := b.Write([]byte(index)); err != nil {
		panic(err)
	}

	Must(binary.Write(&b, binary.LittleEndian, uint32(len(indexationPayload.Data))))
	if _, err := b.Write(indexationPayload.Data); err != nil {
		panic(err)
	}

	return indexationPayload, b.Bytes()
}

// RandMessage returns a random message with the given inner payload.
func RandMessage(withPayloadType uint32) (*iotago.Message, []byte) {
	var payload serializer.Serializable
	var payloadData []byte

	parents := SortedRand32BytArray(1 + rand.Intn(7))

	switch withPayloadType {
	case iotago.TransactionPayloadTypeID:
		payload, payloadData = RandTransaction()
	case iotago.IndexationPayloadTypeID:
		payload, payloadData = RandIndexation()
	case iotago.MilestonePayloadTypeID:
		payload, payloadData = RandMilestone(parents)
	}

	m := &iotago.Message{}
	m.NetworkID = 1
	m.Payload = payload
	m.Nonce = uint64(rand.Intn(1000))
	m.Parents = parents

	var b bytes.Buffer
	Must(binary.Write(&b, binary.LittleEndian, m.NetworkID))
	Must(binary.Write(&b, binary.LittleEndian, byte(len(m.Parents))))

	for _, parent := range m.Parents {
		if _, err := b.Write(parent[:]); err != nil {
			panic(err)
		}
	}

	switch {
	case payload == nil:
		// zero length payload
		Must(binary.Write(&b, binary.LittleEndian, uint32(0)))
	default:
		Must(binary.Write(&b, binary.LittleEndian, uint32(len(payloadData))))
		if _, err := b.Write(payloadData); err != nil {
			panic(err)
		}
	}

	Must(binary.Write(&b, binary.LittleEndian, m.Nonce))

	return m, b.Bytes()
}

// RandTransaction returns a random transaction.
func RandTransaction() (*iotago.Transaction, []byte) {
	var buf bytes.Buffer
	Must(binary.Write(&buf, binary.LittleEndian, iotago.TransactionPayloadTypeID))

	sigTxPayload := &iotago.Transaction{}
	unTx, unTxData := RandTransactionEssence()
	_, err := buf.Write(unTxData)
	Must(err)
	sigTxPayload.Essence = unTx

	unlockBlocksCount := len(unTx.Inputs)
	Must(binary.Write(&buf, binary.LittleEndian, uint16(unlockBlocksCount)))
	for i := unlockBlocksCount; i > 0; i-- {
		unlockBlock, unlockBlockData := RandEd25519SignatureUnlockBlock()
		_, err := buf.Write(unlockBlockData)
		Must(err)
		sigTxPayload.UnlockBlocks = append(sigTxPayload.UnlockBlocks, unlockBlock)
	}

	return sigTxPayload, buf.Bytes()
}

// RandTreasuryInput returns a random treasury input.
func RandTreasuryInput() (*iotago.TreasuryInput, []byte) {
	treasuryInput := &iotago.TreasuryInput{}
	input := RandBytes(iotago.TreasuryInputBytesLength)
	copy(treasuryInput[:], input)
	var b [iotago.TreasuryInputSerializedBytesSize]byte
	b[0] = iotago.InputTreasury
	copy(b[serializer.SmallTypeDenotationByteSize:], input)
	return treasuryInput, b[:]
}

// RandUTXOInput returns a random UTXO input.
func RandUTXOInput() (*iotago.UTXOInput, []byte) {
	utxoInput := &iotago.UTXOInput{}
	var b [iotago.UTXOInputSize]byte
	b[0] = iotago.InputUTXO

	txID := RandBytes(iotago.TransactionIDLength)
	copy(b[serializer.SmallTypeDenotationByteSize:], txID)
	copy(utxoInput.TransactionID[:], txID)

	index := uint16(rand.Intn(iotago.RefUTXOIndexMax))
	binary.LittleEndian.PutUint16(b[len(b)-serializer.UInt16ByteSize:], index)
	utxoInput.TransactionOutputIndex = index
	return utxoInput, b[:]
}

// RandTreasuryOutput returns a random treasury output.
func RandTreasuryOutput() (*iotago.TreasuryOutput, []byte) {
	var b bytes.Buffer

	deposit := rand.Uint64()
	Must(binary.Write(&b, binary.LittleEndian, iotago.OutputTreasuryOutput))
	Must(binary.Write(&b, binary.LittleEndian, deposit))

	return &iotago.TreasuryOutput{Amount: deposit}, b.Bytes()
}

// RandTreasuryTransaction returns a random treasury transaction.
func RandTreasuryTransaction() (*iotago.TreasuryTransaction, []byte) {
	var b bytes.Buffer

	treasuryInput, treasuryInputBytes := RandTreasuryInput()
	treasuryOutput, treasuryOutputBytes := RandTreasuryOutput()
	Must(binary.Write(&b, binary.LittleEndian, iotago.TreasuryTransactionPayloadTypeID))
	_, err := b.Write(treasuryInputBytes)
	Must(err)
	_, err = b.Write(treasuryOutputBytes)
	Must(err)
	return &iotago.TreasuryTransaction{
		Input:  treasuryInput,
		Output: treasuryOutput,
	}, b.Bytes()
}

// RandSigLockedSingleOutput returns a random signature locked single output.
func RandSigLockedSingleOutput(addrType iotago.AddressType) (*iotago.SigLockedSingleOutput, []byte) {
	var buf bytes.Buffer
	Must(buf.WriteByte(iotago.OutputSigLockedSingleOutput))

	dep := &iotago.SigLockedSingleOutput{}

	var addrData []byte
	switch addrType {
	case iotago.AddressEd25519:
		dep.Address, addrData = RandEd25519Address()
	default:
		panic(fmt.Sprintf("invalid addr type: %d", addrType))
	}

	_, err := buf.Write(addrData)
	Must(err)

	amount := uint64(rand.Intn(10000))
	Must(binary.Write(&buf, binary.LittleEndian, amount))
	dep.Amount = amount

	return dep, buf.Bytes()
}

// OneInputOutputTransaction generates a random transaction with one input and output.
func OneInputOutputTransaction() *iotago.Transaction {
	return &iotago.Transaction{
		Essence: &iotago.TransactionEssence{
			Inputs: []serializer.Serializable{
				&iotago.UTXOInput{
					TransactionID: func() [iotago.TransactionIDLength]byte {
						var b [iotago.TransactionIDLength]byte
						copy(b[:], RandBytes(iotago.TransactionIDLength))
						return b
					}(),
					TransactionOutputIndex: 0,
				},
			},
			Outputs: []serializer.Serializable{
				&iotago.SigLockedSingleOutput{
					Address: func() serializer.Serializable {
						edAddr, _ := RandEd25519Address()
						return edAddr
					}(),
					Amount: 1337,
				},
			},
			Payload: nil,
		},
		UnlockBlocks: []serializer.Serializable{
			&iotago.SignatureUnlockBlock{
				Signature: func() serializer.Serializable {
					edSig, _ := RandEd25519Signature()
					return edSig
				}(),
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
