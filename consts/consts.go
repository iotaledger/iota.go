// Package consts contains constants used throughout the entire library and errors.
package consts

import "strings"

const (
	// DefaultMinWeightMagnitude is the default difficulty on mainnet.
	DefaultMinWeightMagnitude = 14
)

// SecurityLevel defines the security level used for input transactions or respectively, how many
// signature fragments will be generated for value transfers.
type SecurityLevel int

const (
	// SecurityLevelLow is for devices with low performance and only storing a small amount of value.
	SecurityLevelLow SecurityLevel = 1
	// SecurityLevelMedium is the standard security level for wallets and other applications.
	SecurityLevelMedium SecurityLevel = 2
	// SecurityLevelHigh is recommended for exchanges and other high security applications.
	SecurityLevelHigh SecurityLevel = 3
	// MaxSecurityLevel is the maximum security level.
	MaxSecurityLevel = 3
)

const (
	// TrinaryRadix defines the base of the trinary system.
	TrinaryRadix = 3
	// TryteAlphabet are letters of the alphabet and the number 9
	// which directly map to decimal values of a single Tryte value.
	TryteAlphabet = "9ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// MinTryteValue is the minimum value of a tryte value.
	MinTryteValue = -13
	// MaxTryteValue is the maximum value of a tryte value.
	MaxTryteValue = 13
	// MinTritValue is the minimum value of a trit value.
	MinTritValue = -1
	// MaxTritValue is the maximum value of a trit value.
	MaxTritValue = 1
)

const (
	// HashTrinarySize is the standard size for hashes from Curl or Kerl.
	HashTrinarySize = 243
	// HashTrytesSize is the trytes size of HashTrinarySize.
	HashTrytesSize = HashTrinarySize / 3
	// HashBytesSize is the bytes size of HashTrytesSize.
	HashBytesSize = 48
	// IntLength is used for bytes-trytes conversion.
	IntLength = HashBytesSize / 4
	// KeySegmentsPerFragment is the amount of segments per key fragment.
	KeySegmentsPerFragment = 27
	// KeyFragmentLength defines the length of a key fragment.
	KeyFragmentLength = HashTrinarySize * KeySegmentsPerFragment // 6561
	// KeySegmentHashRounds is the amount of hashing rounds during key segment hashing.
	KeySegmentHashRounds = 26
)

// Address and checksum constants.
const (
	AddressChecksumTrytesSize     = 9
	AddressWithChecksumTrytesSize = HashTrytesSize + AddressChecksumTrytesSize
	MinChecksumTrytesSize         = 3
)

// Null value constants.
var (
	NullHashTrytes                     = strings.Repeat("9", HashTrytesSize)
	NullTagTrytes                      = strings.Repeat("9", TagTrinarySize/3)
	NullNonceTrytes                    = strings.Repeat("9", NonceTrinarySize/3)
	NullSignatureMessageFragmentTrytes = strings.Repeat("9", SignatureMessageFragmentTrinarySize/3)
	NullAddressWithChecksum            = strings.Repeat("9", HashTrytesSize) + "A9BEONKZW"
)

// Attachment timestamp constants.
const (
	UpperBoundAttachmentTimestamp = (3 ^ 27 - 1) / 2
	LowerBoundAttachmentTimestamp = 0
)

// Transaction elements size and offsets.
const (
	SignatureMessageFragmentTrinaryOffset = 0
	SignatureMessageFragmentTrinarySize   = 6561
	AddressTrinaryOffset                  = SignatureMessageFragmentTrinaryOffset + SignatureMessageFragmentTrinarySize
	AddressTrinarySize                    = 243
	ValueOffsetTrinary                    = AddressTrinaryOffset + AddressTrinarySize
	ValueSizeTrinary                      = 81
	ObsoleteTagTrinaryOffset              = ValueOffsetTrinary + ValueSizeTrinary
	ObsoleteTagTrinarySize                = 81
	TimestampTrinaryOffset                = ObsoleteTagTrinaryOffset + ObsoleteTagTrinarySize
	TimestampTrinarySize                  = 27
	CurrentIndexTrinaryOffset             = TimestampTrinaryOffset + TimestampTrinarySize
	CurrentIndexTrinarySize               = 27
	LastIndexTrinaryOffset                = CurrentIndexTrinaryOffset + CurrentIndexTrinarySize
	LastIndexTrinarySize                  = 27
	BundleTrinaryOffset                   = LastIndexTrinaryOffset + LastIndexTrinarySize
	BundleTrinarySize                     = 243
	TrunkTransactionTrinaryOffset         = BundleTrinaryOffset + BundleTrinarySize
	TrunkTransactionTrinarySize           = 243
	BranchTransactionTrinaryOffset        = TrunkTransactionTrinaryOffset + TrunkTransactionTrinarySize
	BranchTransactionTrinarySize          = 243
	TagTrinaryOffset                      = BranchTransactionTrinaryOffset + BranchTransactionTrinarySize
	TagTrinarySize                        = 81
	AttachmentTimestampTrinaryOffset      = TagTrinaryOffset + TagTrinarySize
	AttachmentTimestampTrinarySize        = 27

	AttachmentTimestampLowerBoundTrinaryOffset = AttachmentTimestampTrinaryOffset + AttachmentTimestampTrinarySize
	AttachmentTimestampLowerBoundTrinarySize   = 27
	AttachmentTimestampUpperBoundTrinaryOffset = AttachmentTimestampLowerBoundTrinaryOffset + AttachmentTimestampLowerBoundTrinarySize
	AttachmentTimestampUpperBoundTrinarySize   = 27
	NonceTrinaryOffset                         = AttachmentTimestampUpperBoundTrinaryOffset + AttachmentTimestampUpperBoundTrinarySize
	NonceTrinarySize                           = 81

	TransactionTrinarySize = SignatureMessageFragmentTrinarySize + AddressTrinarySize +
		ValueSizeTrinary + ObsoleteTagTrinarySize + TimestampTrinarySize +
		CurrentIndexTrinarySize + LastIndexTrinarySize + BundleTrinarySize +
		TrunkTransactionTrinarySize + BranchTransactionTrinarySize +
		TagTrinarySize + AttachmentTimestampTrinarySize +
		AttachmentTimestampLowerBoundTrinarySize + AttachmentTimestampUpperBoundTrinarySize +
		NonceTrinarySize

	SignatureMessageFragmentSizeInTrytes = SignatureMessageFragmentTrinarySize / 3
	TransactionTrytesSize                = TransactionTrinarySize / 3
)

// Trinary conversion constants.
const (
	Radix                int8 = 3
	NumberOfTritsInAByte      = 5
)

// Merkle constants.
const (
	ISSFragments   uint64 = 27
	ISSKeyLength   uint64 = HashTrinarySize * ISSFragments
	ISSChunkLength int    = HashTrinarySize / TrinaryRadix
)
