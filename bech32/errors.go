package bech32

import "github.com/iotaledger/hive.go/ierrors"

// Errors reported during bech32 decoding.
var (
	ErrInvalidLength    = ierrors.New("invalid bech32 length")
	ErrMissingSeparator = ierrors.New("missing bech32 separator '" + string(separator) + "'")
	ErrInvalidSeparator = ierrors.New("bech32 separator '" + string(separator) + "' at invalid position")
	ErrMixedCase        = ierrors.New("mixed case in bech32 string")
	ErrInvalidCharacter = ierrors.New("invalid bech32 character")
	ErrInvalidChecksum  = ierrors.New("invalid bech32 checksum")
)

// A SyntaxError is a description of a Bech32 syntax error.
type SyntaxError struct {
	err    error // wrapped error
	Offset int   // error occurred after reading Offset bytes
}

func (e *SyntaxError) Error() string { return e.err.Error() }

func (e *SyntaxError) Unwrap() error { return e.err }
