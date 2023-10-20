package iotago

import (
	"github.com/iotaledger/hive.go/ierrors"
	"github.com/iotaledger/iota.go/v4/hexutil"
)

// ParseImplicitAccountCreationAddressFromHexString parses the given hex string into an ImplicitAccountCreationAddress.
func ParseImplicitAccountCreationAddressFromHexString(hexAddr string) (*ImplicitAccountCreationAddress, error) {
	addrBytes, err := hexutil.DecodeHex(hexAddr)
	if err != nil {
		return nil, err
	}

	if len(addrBytes) < ImplicitAccountCreationAddressBytesLength {
		return nil, ierrors.New("invalid ImplicitAccountCreationAddress length")
	}

	addr := &ImplicitAccountCreationAddress{}
	copy(addr[:], addrBytes)

	return addr, nil
}

// MustParseImplicitAccountCreationAddressFromHexString parses the given hex string into an ImplicitAccountCreationAddress.
// It panics if the hex address is invalid.
func MustParseImplicitAccountCreationAddressFromHexString(hexAddr string) *ImplicitAccountCreationAddress {
	addr, err := ParseImplicitAccountCreationAddressFromHexString(hexAddr)
	if err != nil {
		panic(err)
	}

	return addr
}

func (addr *ImplicitAccountCreationAddress) StorageScore(storageScoreStruct *StorageScoreStructure, _ StorageScoreFunc) StorageScore {
	return storageScoreStruct.OffsetImplicitAccountCreationAddress
}

func (addr *ImplicitAccountCreationAddress) CannotReceiveNativeTokens() bool {
	return true
}

func (addr *ImplicitAccountCreationAddress) CannotReceiveMana() bool {
	return false
}

func (addr *ImplicitAccountCreationAddress) CannotReceiveOutputsWithTimelockUnlockCondition() bool {
	return true
}

func (addr *ImplicitAccountCreationAddress) CannotReceiveOutputsWithExpirationUnlockCondition() bool {
	return true
}

func (addr *ImplicitAccountCreationAddress) CannotReceiveOutputsWithStorageDepositReturnUnlockCondition() bool {
	return true
}

func (addr *ImplicitAccountCreationAddress) CannotReceiveAccountOutputs() bool {
	return true
}

func (addr *ImplicitAccountCreationAddress) CannotReceiveNFTOutputs() bool {
	return true
}

func (addr *ImplicitAccountCreationAddress) CannotReceiveDelegationOutputs() bool {
	return true
}
