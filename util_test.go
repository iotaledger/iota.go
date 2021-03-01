package iotago_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	legacy "github.com/iotaledger/iota.go/consts"
	"github.com/iotaledger/iota.go/trinary"
	"github.com/iotaledger/iota.go/v2"
	"github.com/iotaledger/iota.go/v2/ed25519"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

// returns length amount random bytes
func randBytes(length int) []byte {
	var b []byte
	for i := 0; i < length; i++ {
		b = append(b, byte(rand.Intn(256)))
	}
	return b
}

func randTrytes(length int) trinary.Trytes {
	var trytes strings.Builder
	for i := 0; i < length; i++ {
		trytes.WriteByte(legacy.TryteAlphabet[rand.Intn(len(legacy.TryteAlphabet))])
	}
	return trytes.String()
}

func rand32ByteHash() [32]byte {
	var h [32]byte
	b := randBytes(32)
	copy(h[:], b)
	return h
}

func rand49ByteHash() [49]byte {
	var h [49]byte
	b := randBytes(49)
	copy(h[:], b)
	return h
}

func sortedRand32ByteHashes(count int) [][32]byte {
	hashes := make(iotago.LexicalOrdered32ByteArrays, count)
	for i := 0; i < count; i++ {
		hashes[i] = rand32ByteHash()
	}
	sort.Sort(hashes)
	return hashes
}

func rand64ByteHash() [64]byte {
	var h [64]byte
	b := randBytes(64)
	copy(h[:], b)
	return h
}

func randEd25519Addr() (*iotago.Ed25519Address, []byte) {
	// type
	edAddr := &iotago.Ed25519Address{}
	addr := randBytes(iotago.Ed25519AddressBytesLength)
	copy(edAddr[:], addr)
	// serialized
	var b [iotago.Ed25519AddressSerializedBytesSize]byte
	b[0] = iotago.AddressEd25519
	copy(b[iotago.SmallTypeDenotationByteSize:], addr)
	return edAddr, b[:]
}

func randEd25519Signature() (*iotago.Ed25519Signature, []byte) {
	// type
	edSig := &iotago.Ed25519Signature{}
	pub := randBytes(ed25519.PublicKeySize)
	sig := randBytes(ed25519.SignatureSize)
	copy(edSig.PublicKey[:], pub)
	copy(edSig.Signature[:], sig)
	// serialized
	var b [iotago.Ed25519SignatureSerializedBytesSize]byte
	b[0] = iotago.SignatureEd25519
	copy(b[iotago.SmallTypeDenotationByteSize:], pub)
	copy(b[iotago.SmallTypeDenotationByteSize+ed25519.PublicKeySize:], sig)
	return edSig, b[:]
}

func randEd25519SignatureUnlockBlock() (*iotago.SignatureUnlockBlock, []byte) {
	edSig, edSigData := randEd25519Signature()
	block := &iotago.SignatureUnlockBlock{Signature: edSig}
	return block, append([]byte{iotago.UnlockBlockSignature}, edSigData...)
}

func randReferenceUnlockBlock() (*iotago.ReferenceUnlockBlock, []byte) {
	return referenceUnlockBlock(uint16(rand.Intn(1000)))
}

func referenceUnlockBlock(index uint16) (*iotago.ReferenceUnlockBlock, []byte) {
	var b [iotago.ReferenceUnlockBlockSize]byte
	b[0] = iotago.UnlockBlockReference
	binary.LittleEndian.PutUint16(b[iotago.SmallTypeDenotationByteSize:], index)
	return &iotago.ReferenceUnlockBlock{Reference: index}, b[:]
}

func randTransactionEssence() (*iotago.TransactionEssence, []byte) {
	var buf bytes.Buffer

	tx := &iotago.TransactionEssence{}
	must(buf.WriteByte(iotago.TransactionEssenceNormal))

	inputsBytes := iotago.LexicalOrderedByteSlices{}
	inputCount := rand.Intn(10) + 1
	must(binary.Write(&buf, binary.LittleEndian, uint16(inputCount)))
	for i := inputCount; i > 0; i-- {
		_, inputData := randUTXOInput()
		inputsBytes = append(inputsBytes, inputData)
	}

	sort.Sort(inputsBytes)

	for _, inputData := range inputsBytes {
		_, err := buf.Write(inputData)
		must(err)
		input := &iotago.UTXOInput{}
		if _, err := input.Deserialize(inputData, iotago.DeSeriModePerformValidation); err != nil {
			panic(err)
		}
		tx.Inputs = append(tx.Inputs, input)
	}

	outputsBytes := iotago.LexicalOrderedByteSlices{}
	outputCount := rand.Intn(10) + 1
	must(binary.Write(&buf, binary.LittleEndian, uint16(outputCount)))
	for i := outputCount; i > 0; i-- {
		_, depData := randSigLockedSingleOutput(iotago.AddressEd25519)
		outputsBytes = append(outputsBytes, depData)
	}

	sort.Sort(outputsBytes)
	for _, outputData := range outputsBytes {
		_, err := buf.Write(outputData)
		must(err)
		output := &iotago.SigLockedSingleOutput{}
		if _, err := output.Deserialize(outputData, iotago.DeSeriModePerformValidation); err != nil {
			panic(err)
		}
		tx.Outputs = append(tx.Outputs, output)
	}

	// empty payload
	must(binary.Write(&buf, binary.LittleEndian, uint32(0)))

	return tx, buf.Bytes()
}

func randMigratedFundsEntry() (*iotago.MigratedFundsEntry, []byte) {
	tailTxHash := rand49ByteHash()
	addr, addrBytes := randEd25519Addr()
	deposit := rand.Uint64()

	var b bytes.Buffer
	_, err := b.Write(tailTxHash[:])
	must(err)
	_, err = b.Write(addrBytes)
	must(err)
	must(binary.Write(&b, binary.LittleEndian, deposit))

	return &iotago.MigratedFundsEntry{
		TailTransactionHash: tailTxHash,
		Address:             addr,
		Deposit:             deposit,
	}, b.Bytes()
}

func randReceipt() (*iotago.Receipt, []byte) {
	receipt := &iotago.Receipt{MigratedAt: 1000, Final: true}

	var b bytes.Buffer

	must(binary.Write(&b, binary.LittleEndian, iotago.ReceiptPayloadTypeID))
	must(binary.Write(&b, binary.LittleEndian, receipt.MigratedAt))
	must(b.WriteByte(1))

	migFundsEntriesBytes := iotago.LexicalOrderedByteSlices{}
	migFundsEntriesCount := rand.Intn(10) + 1
	must(binary.Write(&b, binary.LittleEndian, uint16(migFundsEntriesCount)))
	for i := migFundsEntriesCount; i > 0; i-- {
		_, migFundsEntryBytes := randMigratedFundsEntry()
		migFundsEntriesBytes = append(migFundsEntriesBytes, migFundsEntryBytes)
	}

	sort.Sort(migFundsEntriesBytes)

	for _, migFundEntryBytes := range migFundsEntriesBytes {
		_, err := b.Write(migFundEntryBytes)
		must(err)
		migFundsEntry := &iotago.MigratedFundsEntry{}
		if _, err := migFundsEntry.Deserialize(migFundEntryBytes, iotago.DeSeriModePerformValidation); err != nil {
			panic(err)
		}
		receipt.Funds = append(receipt.Funds, migFundsEntry)
	}

	randTreasuryTx, randTreasuryTxBytes := randTreasuryTransaction()
	receipt.Transaction = randTreasuryTx

	must(binary.Write(&b, binary.LittleEndian, uint32(len(randTreasuryTxBytes))))
	if _, err := b.Write(randTreasuryTxBytes); err != nil {
		must(err)
	}

	return receipt, b.Bytes()
}

func randMilestone(parents iotago.MessageIDs) (*iotago.Milestone, []byte) {
	inclusionMerkleProof := randBytes(iotago.MilestoneInclusionMerkleProofLength)
	const sigsCount = 3

	if parents == nil {
		parents = sortedRand32ByteHashes(1 + rand.Intn(7))
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
				msPubKeys[i] = rand32ByteHash()
				// ensure lexical ordering
				msPubKeys[i][0] = byte(i)
			}
			return msPubKeys
		}(),
		Signatures: func() [][iotago.MilestoneSignatureLength]byte {
			msSigs := make([][iotago.MilestoneSignatureLength]byte, sigsCount)
			for i := 0; i < sigsCount; i++ {
				msSigs[i] = randMilestoneSig()
			}
			return msSigs
		}(),
	}

	var b bytes.Buffer
	must(binary.Write(&b, binary.LittleEndian, iotago.MilestonePayloadTypeID))
	must(binary.Write(&b, binary.LittleEndian, msPayload.Index))
	must(binary.Write(&b, binary.LittleEndian, msPayload.Timestamp))
	must(binary.Write(&b, binary.LittleEndian, byte(len(msPayload.Parents))))
	for _, parent := range msPayload.Parents {
		if _, err := b.Write(parent[:]); err != nil {
			panic(err)
		}
	}
	if _, err := b.Write(msPayload.InclusionMerkleProof[:]); err != nil {
		panic(err)
	}
	must(b.WriteByte(sigsCount))
	for _, pubKey := range msPayload.PublicKeys {
		if _, err := b.Write(pubKey[:]); err != nil {
			panic(err)
		}
	}
	must(binary.Write(&b, binary.LittleEndian, uint32(0)))
	must(b.WriteByte(sigsCount))
	for _, sig := range msPayload.Signatures {
		if _, err := b.Write(sig[:]); err != nil {
			panic(err)
		}
	}

	return msPayload, b.Bytes()
}

func randMilestoneSig() [iotago.MilestoneSignatureLength]byte {
	var sig [iotago.MilestoneSignatureLength]byte
	copy(sig[:], randBytes(iotago.MilestoneSignatureLength))
	return sig
}

func randIndexation(dataLength ...int) (*iotago.Indexation, []byte) {
	const index = "寿司を作って"

	var data []byte
	switch {
	case len(dataLength) > 0:
		data = randBytes(dataLength[0])
	default:
		data = randBytes(rand.Intn(200) + 1)
	}

	indexationPayload := &iotago.Indexation{Index: []byte(index), Data: data}

	var b bytes.Buffer
	must(binary.Write(&b, binary.LittleEndian, iotago.IndexationPayloadTypeID))

	strLen := uint16(len(index))
	must(binary.Write(&b, binary.LittleEndian, strLen))

	if _, err := b.Write([]byte(index)); err != nil {
		panic(err)
	}

	must(binary.Write(&b, binary.LittleEndian, uint32(len(indexationPayload.Data))))
	if _, err := b.Write(indexationPayload.Data); err != nil {
		panic(err)
	}

	return indexationPayload, b.Bytes()
}

func randMessage(withPayloadType uint32) (*iotago.Message, []byte) {
	var payload iotago.Serializable
	var payloadData []byte

	parents := sortedRand32ByteHashes(1 + rand.Intn(7))

	switch withPayloadType {
	case iotago.TransactionPayloadTypeID:
		payload, payloadData = randTransaction()
	case iotago.IndexationPayloadTypeID:
		payload, payloadData = randIndexation()
	case iotago.MilestonePayloadTypeID:
		payload, payloadData = randMilestone(parents)
	}

	m := &iotago.Message{}
	m.NetworkID = 1
	m.Payload = payload
	m.Nonce = uint64(rand.Intn(1000))
	m.Parents = parents

	var b bytes.Buffer
	must(binary.Write(&b, binary.LittleEndian, m.NetworkID))
	must(binary.Write(&b, binary.LittleEndian, byte(len(m.Parents))))

	for _, parent := range m.Parents {
		if _, err := b.Write(parent[:]); err != nil {
			panic(err)
		}
	}

	switch {
	case payload == nil:
		// zero length payload
		must(binary.Write(&b, binary.LittleEndian, uint32(0)))
	default:
		must(binary.Write(&b, binary.LittleEndian, uint32(len(payloadData))))
		if _, err := b.Write(payloadData); err != nil {
			panic(err)
		}
	}

	must(binary.Write(&b, binary.LittleEndian, m.Nonce))

	return m, b.Bytes()
}

func randTransaction() (*iotago.Transaction, []byte) {
	var buf bytes.Buffer
	must(binary.Write(&buf, binary.LittleEndian, iotago.TransactionPayloadTypeID))

	sigTxPayload := &iotago.Transaction{}
	unTx, unTxData := randTransactionEssence()
	_, err := buf.Write(unTxData)
	must(err)
	sigTxPayload.Essence = unTx

	unlockBlocksCount := len(unTx.Inputs)
	must(binary.Write(&buf, binary.LittleEndian, uint16(unlockBlocksCount)))
	for i := unlockBlocksCount; i > 0; i-- {
		unlockBlock, unlockBlockData := randEd25519SignatureUnlockBlock()
		_, err := buf.Write(unlockBlockData)
		must(err)
		sigTxPayload.UnlockBlocks = append(sigTxPayload.UnlockBlocks, unlockBlock)
	}

	return sigTxPayload, buf.Bytes()
}

func randTreasuryInput() (*iotago.TreasuryInput, []byte) {
	treasuryInput := &iotago.TreasuryInput{}
	input := randBytes(iotago.TreasuryInputBytesLength)
	copy(treasuryInput[:], input)
	var b [iotago.TreasuryInputSerializedBytesSize]byte
	b[0] = iotago.InputTreasury
	copy(b[iotago.SmallTypeDenotationByteSize:], input)
	return treasuryInput, b[:]
}

func randUTXOInput() (*iotago.UTXOInput, []byte) {
	utxoInput := &iotago.UTXOInput{}
	var b [iotago.UTXOInputSize]byte
	b[0] = iotago.InputUTXO

	txID := randBytes(iotago.TransactionIDLength)
	copy(b[iotago.SmallTypeDenotationByteSize:], txID)
	copy(utxoInput.TransactionID[:], txID)

	index := uint16(rand.Intn(iotago.RefUTXOIndexMax))
	binary.LittleEndian.PutUint16(b[len(b)-iotago.UInt16ByteSize:], index)
	utxoInput.TransactionOutputIndex = index
	return utxoInput, b[:]
}

func randTreasuryOutput() (*iotago.TreasuryOutput, []byte) {
	var b bytes.Buffer

	deposit := rand.Uint64()
	must(binary.Write(&b, binary.LittleEndian, iotago.OutputTreasuryOutput))
	must(binary.Write(&b, binary.LittleEndian, deposit))

	return &iotago.TreasuryOutput{Amount: deposit}, b.Bytes()
}

func randTreasuryTransaction() (*iotago.TreasuryTransaction, []byte) {
	var b bytes.Buffer

	treasuryInput, treasuryInputBytes := randTreasuryInput()
	treasuryOutput, treasuryOutputBytes := randTreasuryOutput()
	must(binary.Write(&b, binary.LittleEndian, iotago.TreasuryTransactionPayloadTypeID))
	_, err := b.Write(treasuryInputBytes)
	must(err)
	_, err = b.Write(treasuryOutputBytes)
	must(err)
	return &iotago.TreasuryTransaction{
		Input:  treasuryInput,
		Output: treasuryOutput,
	}, b.Bytes()
}

func randSigLockedSingleOutput(addrType iotago.AddressType) (*iotago.SigLockedSingleOutput, []byte) {
	var buf bytes.Buffer
	must(buf.WriteByte(iotago.OutputSigLockedSingleOutput))

	dep := &iotago.SigLockedSingleOutput{}

	var addrData []byte
	switch addrType {
	case iotago.AddressEd25519:
		dep.Address, addrData = randEd25519Addr()
	default:
		panic(fmt.Sprintf("invalid addr type: %d", addrType))
	}

	_, err := buf.Write(addrData)
	must(err)

	amount := uint64(rand.Intn(10000))
	must(binary.Write(&buf, binary.LittleEndian, amount))
	dep.Amount = amount

	return dep, buf.Bytes()
}

func oneInputOutputTransaction() *iotago.Transaction {
	return &iotago.Transaction{
		Essence: &iotago.TransactionEssence{
			Inputs: []iotago.Serializable{
				&iotago.UTXOInput{
					TransactionID: func() [iotago.TransactionIDLength]byte {
						var b [iotago.TransactionIDLength]byte
						copy(b[:], randBytes(iotago.TransactionIDLength))
						return b
					}(),
					TransactionOutputIndex: 0,
				},
			},
			Outputs: []iotago.Serializable{
				&iotago.SigLockedSingleOutput{
					Address: func() iotago.Serializable {
						edAddr, _ := randEd25519Addr()
						return edAddr
					}(),
					Amount: 1337,
				},
			},
			Payload: nil,
		},
		UnlockBlocks: []iotago.Serializable{
			&iotago.SignatureUnlockBlock{
				Signature: func() iotago.Serializable {
					edSig, _ := randEd25519Signature()
					return edSig
				}(),
			},
		},
	}
}

func randEd25519PrivateKey() ed25519.PrivateKey {
	seed := randEd25519Seed()
	return ed25519.NewKeyFromSeed(seed[:])
}

func randEd25519Seed() [ed25519.SeedSize]byte {
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
