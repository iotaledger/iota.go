package iotago

import "golang.org/x/crypto/blake2b"

func (a AccountID) Addressable() bool {
	return true
}

func (a AccountID) Key() interface{} {
	return a.String()
}

func (a AccountID) FromOutputID(in OutputID) ChainID {
	return AccountIDFromOutputID(in)
}

func (a AccountID) Matches(other ChainID) bool {
	otherAccountID, isAccountID := other.(AccountID)
	if !isAccountID {
		return false
	}

	return a == otherAccountID
}

func (a AccountID) ToAddress() ChainAddress {
	var addr AccountAddress
	copy(addr[:], a[:])

	return &addr
}

// AccountIDFromOutputID returns the AccountID computed from a given OutputID.
func AccountIDFromOutputID(outputID OutputID) AccountID {
	return blake2b.Sum256(outputID[:])
}
