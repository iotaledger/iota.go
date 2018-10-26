package consts

import "strings"

const (
	DefaultMinWeightMagnitude = 14
)

type SecurityLevel int

const (
	SecurityLevelLow    SecurityLevel = 1
	SecurityLevelMedium SecurityLevel = 2
	SecurityLevelHigh   SecurityLevel = 3
)

const (
	TrinaryRadix  = 3
	TryteAlphabet = "9ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	MinTryteValue = -13
	MaxTryteValue = 13
	MinTritValue  = -1
	MaxTritValue  = 1
)

const (
	HashTrinarySize                      = 243
	HashTrytesSize                       = HashTrinarySize / 3
	HashBytesSize                        = 48
	IntLength                            = HashBytesSize / 4
	KeySegmentsPerFragment               = 27
	KeyFragmentLength                    = HashTrinarySize * KeySegmentsPerFragment // 6561
	KeySegmentHashRounds                 = 26
	SignatureMessageFragmentSizeInTrytes = SignatureMessageFragmentTrinarySize / 3
	MaxSecurityLevel                     = 3
)

const (
	AddressChecksumTrytesSize     = 9
	AddressWithChecksumTrytesSize = HashTrytesSize + AddressChecksumTrytesSize
	MinChecksumTrytesSize         = 3
)

var (
	NullHashTrytes                     = strings.Repeat("9", HashTrytesSize)
	NullTagTrytes                      = strings.Repeat("9", TagTrinarySize/3)
	NullNonceTrytes                    = strings.Repeat("9", NonceTrinarySize/3)
	NullSignatureMessageFragmentTrytes = strings.Repeat("9", SignatureMessageFragmentTrinarySize/3)
)

const (
	// (3^27-1)/2
	UpperBoundAttachmentTimestamp = (3 ^ 27 - 1) / 2
	LowerBoundAttachmentTimestamp = 0
)

// transaction elements size and offsets
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

	TransactionTrytesSize = TransactionTrinarySize / 3
)
