package bech32

import "errors"

// Errors reported during bech32 decoding.
var (
	ErrInvalidLength    = errors.New("invalid length")
	ErrMissingSeparator = errors.New("missing separator '" + string(separator) + "'")
	ErrInvalidSeparator = errors.New("separator '" + string(separator) + "' at invalid position")
	ErrMixedCase        = errors.New("mixed case")
	ErrInvalidCharacter = errors.New("invalid character")
	ErrInvalidChecksum  = errors.New("invalid checksum")
)

// A SyntaxError is a description of a Bech32 syntax error.
type SyntaxError struct {
	err    error // wrapped error
	Offset int   // error occurred after reading Offset bytes
}

func (e *SyntaxError) Error() string { return e.err.Error() }

func (e *SyntaxError) Unwrap() error { return e.err }
