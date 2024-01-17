//nolint:gosec
package tpkg

import (
	cryptorand "crypto/rand"
	"math"
	"math/big"
	"math/rand"
	"sort"
	"time"

	"github.com/iotaledger/hive.go/serializer/v2"
	iotago "github.com/iotaledger/iota.go/v4"
)

func RandomRead(p []byte) (n int, err error) {
	return cryptorand.Read(p)
}

// RandByte returns a random byte.
func RandByte() byte {
	return byte(RandInt(256))
}

// RandBytes returns length amount random bytes.
func RandBytes(length int) []byte {
	var b []byte
	for i := 0; i < length; i++ {
		b = append(b, RandByte())
	}

	return b
}

func RandString(length int) string {
	var b []byte
	for i := 0; i < length; i++ {
		// Generate random printable ASCII values between 32 and 126 (inclusive)
		b = append(b, byte(RandInt(95)+32)) // 95 printable ASCII characters (126 - 32 + 1)
	}

	return string(b)
}

// RandInt returns a random int.
func RandInt(max int) int {
	return rand.Intn(max)
}

// RandInt8 returns a random int8.
func RandInt8(max int8) int8 {
	return int8(RandInt32(uint32(max)))
}

// RandInt16 returns a random int16.
func RandInt16(max int16) int16 {
	return int16(RandInt32(uint32(max)))
}

// RandInt32 returns a random int32.
func RandInt32(max uint32) int32 {
	return rand.Int31n(int32(max))
}

// RandInt64 returns a random int64.
func RandInt64(max uint64) int64 {
	return rand.Int63n(int64(uint32(max)))
}

// RandUint returns a random uint.
func RandUint(max uint) uint {
	return uint(RandInt(int(max)))
}

// RandUint8 returns a random uint8.
func RandUint8(max uint8) uint8 {
	return uint8(RandInt32(uint32(max)))
}

// RandUint16 returns a random uint16.
func RandUint16(max uint16) uint16 {
	return uint16(RandInt32(uint32(max)))
}

// RandUint32 returns a random uint32.
func RandUint32(max uint32) uint32 {
	return uint32(RandInt64(uint64(max)))
}

// RandUint64 returns a random uint64.
func RandUint64(max uint64) uint64 {
	return uint64(RandInt64(max))
}

// RandFloat64 returns a random float64.
func RandFloat64(max float64) float64 {
	return rand.Float64() * max
}

// RandUTCTime returns a random time from current year until now in UTC.
func RandUTCTime() time.Time {
	now := time.Now()
	beginnigOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	secTillNow := now.Unix() - beginnigOfYear.Unix()

	return time.Unix(beginnigOfYear.Unix()+RandInt64(uint64(secTillNow)), RandInt64(1e9)).UTC()
}

func RandDuration() time.Duration {
	return time.Duration(RandInt64(math.MaxInt64))
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

func RandSlot() iotago.SlotIndex {
	return iotago.SlotIndex(RandUint32(uint32(iotago.MaxSlotIndex)))
}

func RandEpoch() iotago.EpochIndex {
	return iotago.EpochIndex(RandUint32(uint32(iotago.MaxEpochIndex)))
}

// RandBaseToken returns a random amount of base token.
func RandBaseToken(max iotago.BaseToken) iotago.BaseToken {
	return iotago.BaseToken(rand.Int63n(int64(uint32(max))))
}

// RandMana returns a random amount of mana.
func RandMana(max iotago.Mana) iotago.Mana {
	return iotago.Mana(rand.Int63n(int64(uint32(max))))
}

// RandTaggedData returns a random tagged data payload.
func RandTaggedData(tag []byte, dataLength ...int) *iotago.TaggedData {
	var data []byte
	switch {
	case len(dataLength) > 0:
		data = RandBytes(dataLength[0])
	default:
		data = RandBytes(RandInt(200) + 1)
	}

	return &iotago.TaggedData{Tag: tag, Data: data}
}

// RandUTXOInput returns a random UTXO input.
func RandUTXOInput() *iotago.UTXOInput {
	return RandUTXOInputWithIndex(uint16(RandInt(iotago.RefUTXOIndexMax)))
}

func RandCommitmentID() iotago.CommitmentID {
	return Rand36ByteArray()
}

func RandCommitment() *iotago.Commitment {
	return &iotago.Commitment{
		ProtocolVersion:      iotago.LatestProtocolVersion(),
		Slot:                 RandSlot(),
		PreviousCommitmentID: RandCommitmentID(),
		RootsID:              RandIdentifier(),
		CumulativeWeight:     RandUint64(math.MaxUint64),
		ReferenceManaCost:    RandMana(iotago.MaxMana),
	}
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

// RandWorkScore returns a random workscore.
func RandWorkScore(max uint32) iotago.WorkScore {
	return iotago.WorkScore(RandUint32(max))
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
// Some protocol parameters are subject to sanity checks when the protocol parameters are created
// so we use default values here to avoid panics rather than random ones.
func RandProtocolParameters() iotago.ProtocolParameters {
	return iotago.NewV3SnapshotProtocolParameters(
		iotago.WithStorageOptions(
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
	)
}

func RandTokenScheme() iotago.TokenScheme {
	maxSupply := RandInt64(1_000_000_000)
	return &iotago.SimpleTokenScheme{
		MintedTokens:  big.NewInt(maxSupply - 10),
		MaximumSupply: big.NewInt(maxSupply),
		MeltedTokens:  big.NewInt(0),
	}
}
