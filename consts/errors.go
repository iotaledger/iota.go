package consts

import "github.com/pkg/errors"

var (
	// Returned when no settings are supplied to ComposeAPI().
	ErrSettingsNil = errors.New("settings must not be nil")
	// Returned when a non-ok (200) status code is returned by connected nodes.
	ErrNonOKStatusCodeFromAPIRequest = errors.New("received non ok status from backend")
	// Returned when an error is returned from the conencted nodes but status code was ok.
	ErrUnknownErrorFromAPIRequest = errors.New("received unknown error from backend")
	// Returned if the wrong underyling type of Settings were supplied for creating a Provider.
	ErrInvalidSettingsType = errors.New("incompatible settings type supplied")
	// Returned when the tail transaction is not consistent during promotion.
	ErrInconsistentSubtangle = errors.New("inconsistent subtangle")
	// Kerl's Squeeze() requires a squeeze length which is a multiple of 243.
	ErrInvalidSqueezeLength = errors.New("squeeze length must be a multiple of 243")
	// Returned when the trits length are invalid for the given operation.
	ErrInvalidTritsLength = errors.New("invalid trits length")
	// Returned when the bytes length are invalid for the given operation.
	ErrInvalidBytesLength = errors.New("invalid bytes length")
	// Returned when an operation needs a certain amount of balance to fulfill the operation.
	ErrInsufficientBalance = errors.New("insufficient balance")
	// Returned for invalid address parameters.
	ErrInvalidAddress = errors.New("invalid address")
	// Returned for invalid remainder address parameters.
	ErrInvalidRemainderAddress = errors.New("invalid remainder address")
	// Returned for invalid branch transaction parameters.
	ErrInvalidBranchTransaction = errors.New("invalid branch transaction")
	// Returned for Bundles which are schematically wrong or/and don't pass validation.
	ErrInvalidBundle = errors.New("invalid bundle")
	// Returned for invalid bundle hash parameters.
	ErrInvalidBundleHash = errors.New("invalid bundle hash")
	// Returned for bundles with invalid signatures.
	ErrInvalidSignature = errors.New("invalid signature")
	// Returned for addresses with invalid checksum.
	ErrInvalidChecksum = errors.New("invalid checksum")
	// Returned for invalid hash parameters.
	ErrInvalidHash = errors.New("invalid hash")
	// Returned for invalid index parameters.
	ErrInvalidIndex = errors.New("invalid index option")
	// Returned for invalid total option parameters.
	ErrInvalidTotalOption = errors.New("invalid total option")
	// Returned for invalid input parameters.
	ErrInvalidInput = errors.New("invalid input")
	// Returned for invalid security level parameters.
	ErrInvalidSecurityLevel = errors.New("invalid security option")
	// Returned for invalid seed parameters.
	ErrInvalidSeed = errors.New("invalid seed")
	// Returned for invalid end options.
	ErrInvalidStartEndOptions = errors.New("invalid end option")
	// Returned for invalid tags.
	ErrInvalidTag = errors.New("invalid tag")
	// Returned when transactions trits don't make up a valid transaction.
	ErrInvalidTransaction = errors.New("invalid transaction")
	// Returned for invalid transaction trytes.
	ErrInvalidTransactionTrytes = errors.New("invalid transaction trytes")
	// Returned for invalid attached transaction trytes.
	ErrInvalidAttachedTrytes = errors.New("invalid attached trytes")
	// Returned for invalid transaction hash parameters.
	ErrInvalidTransactionHash = errors.New("invalid transaction hash")
	// Returned for invalid tail transaction hashes.
	ErrInvalidTailTransaction = errors.New("invalid tail transaction")
	// Returned for invalid thresholds used in GetBalances().
	ErrInvalidThreshold = errors.New("invalid threshold option")
	// Returned for invalid transfer parameters.
	ErrInvalidTransfer = errors.New("invalid transfer object")
	// Returned for invalid trunk transaction parameters.
	ErrInvalidTrunkTransaction = errors.New("invalid trunk transaction")
	// Returned for invalid reference hashes.
	ErrInvalidReferenceHash = errors.New("invalid reference hash")
	// Returned for invalid trytes.
	ErrInvalidTrytes = errors.New("invalid trytes")
	// Returned for invalid trit.
	ErrInvalidTrit = errors.New("invalid trit")
	// Returned for invalid URIs.
	ErrInvalidURI = errors.New("invalid uri")
	// Returned for invalid ASCII input for to trytes conversion.
	ErrInvalidASCIIInput = errors.New("conversion to trytes requires type of input to be encoded in ascii")
	// Returned for odd trytes length for to ASCII conversion.
	ErrInvalidOddLength = errors.New("conversion from trytes requires length of trytes to be even")
	// Returned for invalid tryte encoded JSON messages.
	ErrInvalidTryteEncodedJSON = errors.New("invalid tryte encoded JSON message")
	// Returned when a transfer sends back to an Input.
	ErrSendingBackToInputs = errors.New("one of the transaction inputs is used as output")
	// Returned when no remainder was specified for certain type of operations.
	ErrNoRemainderSpecified = errors.New("remainder address is needed on a transfer with remainder")
)
