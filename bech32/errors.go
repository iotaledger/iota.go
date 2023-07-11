package bech32

import "github.com/iotaledger/hive.go/ierrors"

// Errors reported during bech32 decoding.
var (
	ErrInvalidLength    = ierrors.New("invalid length")
	ErrMissingSeparator = ierrors.New("missing separator '" + string(separator) + "'")
	ErrInvalidSeparator = ierrors.New("separator '" + string(separator) + "' at invalid position")
	ErrMixedCase        = ierrors.New("mixed case")
	ErrInvalidCharacter = ierrors.New("invalid character")
	ErrInvalidChecksum  = ierrors.New("invalid checksum")
)

// A SyntaxError is a description of a Bech32 syntax error.
type SyntaxError struct {
	err    error // wrapped error
	Offset int   // error occurred after reading Offset bytes
}

func (e *SyntaxError) Error() string { return e.err.Error() }

func (e *SyntaxError) Unwrap() error { return e.err }
