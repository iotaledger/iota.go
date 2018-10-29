package consts

import "strings"

const (
	// The current minimum weight magnitude of mainnet.
	DefaultMinWeightMagnitude = 14
)

// Defines the security level used for input transactions.
type SecurityLevel int

const (
	// Low security level. For devices with low performance and only storing a small amount.
	SecurityLevelLow SecurityLevel = 1
	// Standard security level. For wallets and other applications.
	SecurityLevelMedium SecurityLevel = 2
	// High security level. Recommended for exchanges.
	SecurityLevelHigh SecurityLevel = 3
	// The maximum security level.
	MaxSecurityLevel = 3
)

const (
	// Radix basis of the trinary system.
	TrinaryRadix = 3
	// Letters and 9 which represent Tryte values.
	TryteAlphabet = "9ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// Minimum value of a tryte value.
	MinTryteValue = -13
	// Maximum value of a tryte value.
	MaxTryteValue = 13
	// Minimum value of a trit value.
	MinTritValue = -1
	// Maximum value of a trit value.
	MaxTritValue = 1
)

const (
	// Standard Hash size in trits.
	HashTrinarySize = 243
	// Standard Hash size in trytes.
	HashTrytesSize = HashTrinarySize / 3
	// Standard Hash size in bytes.
	HashBytesSize = 48
	// Int length used for bytes-trytes conversion.
	IntLength = HashBytesSize / 4
	// Amount of segments per key fragment.
	KeySegmentsPerFragment = 27
	// Length of key fragment in trits.
	KeyFragmentLength = HashTrinarySize * KeySegmentsPerFragment // 6561
	// Amount of rounds during key segment hashing.
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
