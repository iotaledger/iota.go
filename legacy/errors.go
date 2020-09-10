package legacy

import "github.com/pkg/errors"

var (
	// ErrInvalidSqueezeLength gets returned when the squeeze length is not a multiple of 243 (which it must be for Kerl).
	ErrInvalidSqueezeLength = errors.New("squeeze length must be a multiple of 243")
	// ErrInvalidTritsLength gets returned when the trits length are invalid for the given operation.
	ErrInvalidTritsLength = errors.New("invalid trits length")
	// ErrInvalidTrytesLength gets returned when the trytes length are invalid for the given operation.
	ErrInvalidTrytesLength = errors.New("invalid trytes length")
	// ErrInvalidBytesLength gets returned when the bytes length are invalid for the given operation.
	ErrInvalidBytesLength = errors.New("invalid bytes length")
	// ErrInvalidAddress gets returned for invalid address parameters.
	ErrInvalidAddress = errors.New("invalid address")
	// ErrInvalidSignature gets returned for bundles with invalid signatures.
	ErrInvalidSignature = errors.New("invalid signature")
	// ErrInvalidChecksum gets returned for addresses with invalid checksum.
	ErrInvalidChecksum = errors.New("invalid checksum")
	// ErrInvalidHash gets returned for invalid hash parameters.
	ErrInvalidHash = errors.New("invalid hash")
	// ErrInvalidSecurityLevel gets returned for invalid security level parameters.
	ErrInvalidSecurityLevel = errors.New("invalid security option")
	// ErrInvalidSeed gets returned for invalid seed parameters.
	ErrInvalidSeed = errors.New("invalid seed")
	// ErrInvalidStartEndOptions gets returned for invalid end options.
	ErrInvalidStartEndOptions = errors.New("invalid end option")
	// ErrInvalidTrytes gets returned for invalid trytes.
	ErrInvalidTrytes = errors.New("invalid trytes")
	// ErrInvalidByte gets returned for an invalid byte for a to trits conversion (5 packed trits in 1 byte).
	ErrInvalidByte = errors.New("invalid byte")
	// ErrInvalidTrit gets returned for invalid trit.
	ErrInvalidTrit = errors.New("invalid trit")
	// ErrInvalidURI gets returned for invalid URIs.
	ErrInvalidURI = errors.New("invalid uri")
	// ErrInvalidASCIIInput gets returned for invalid ASCII input for to trytes conversion.
	ErrInvalidASCIIInput = errors.New("conversion to trytes requires type of input to be encoded in ascii")
	// ErrInvalidOddLength gets returned for odd trytes length for to ASCII conversion.
	ErrInvalidOddLength = errors.New("conversion from trytes requires length of trytes to be even")
)
