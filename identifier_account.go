package iotago

import "golang.org/x/crypto/blake2b"

func (id AccountID) Addressable() bool {
	return true
}

func (id AccountID) Key() interface{} {
	return id.String()
}

func (id AccountID) FromOutputID(in OutputID) ChainID {
	return AccountIDFromOutputID(in)
}

func (id AccountID) Matches(other ChainID) bool {
	otherAccountID, isAccountID := other.(AccountID)
	if !isAccountID {
		return false
	}

	return id == otherAccountID
}

func (id AccountID) ToAddress() ChainAddress {
	var addr AccountAddress
	copy(addr[:], id[:])

	return &addr
}

// AccountIDFromOutputID returns the AccountID computed from a given OutputID.
func AccountIDFromOutputID(outputID OutputID) AccountID {
	return blake2b.Sum256(outputID[:])
}
