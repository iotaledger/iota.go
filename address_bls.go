package iotago

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/crypto/bls"
	"github.com/iotaledger/hive.go/serializer"
	"golang.org/x/crypto/blake2b"
)

const (
	// BLSAddressBytesLength is the length of a BLS address.
	BLSAddressBytesLength = 32
	// BLSAddressSerializedBytesSize is the size of a serialized BLS address with its type denoting byte.
	BLSAddressSerializedBytesSize = serializer.SmallTypeDenotationByteSize + BLSAddressBytesLength
)

// ParseBLSAddressFromHexString parses the given hex string into a BLSAddress.
func ParseBLSAddressFromHexString(hexAddr string) (*BLSAddress, error) {
	addrBytes, err := hex.DecodeString(hexAddr)
	if err != nil {
		return nil, err
	}
	addr := &BLSAddress{}
	copy(addr[:], addrBytes)
	return addr, nil
}

// MustParseBLSAddressFromHexString parses the given hex string into a BLSAddress.
// It panics if the hex address is invalid.
func MustParseBLSAddressFromHexString(hexAddr string) *BLSAddress {
	addr, err := ParseBLSAddressFromHexString(hexAddr)
	if err != nil {
		panic(err)
	}
	return addr
}

// BLSAddress defines a BLS address.
// A BLSAddress is the Blake2b-256 hash of a BLS public key.
type BLSAddress [BLSAddressBytesLength]byte

func (blsAddr *BLSAddress) VByteCost(costStruct *RentStructure, _ VByteCostFunc) uint64 {
	return costStruct.VBFactorKey.With(costStruct.VBFactorData).Multiply(serializer.SmallTypeDenotationByteSize + BLSAddressBytesLength)
}

func (blsAddr *BLSAddress) Key() string {
	return string(append([]byte{AddressBLS}, (*blsAddr)[:]...))
}

func (blsAddr *BLSAddress) Unlock(msg []byte, sig Signature) error {
	blsSig, isBLSSig := sig.(*BLSSignature)
	if !isBLSSig {
		return fmt.Errorf("%w: can not unlock BLS address with signature of type %s", ErrSignatureAndAddrIncompatible, SignatureTypeToString(sig.Type()))
	}
	return blsSig.Valid(msg, blsAddr)
}

func (blsAddr *BLSAddress) Equal(other Address) bool {
	otherAddr, is := other.(*BLSAddress)
	if !is {
		return false
	}
	return *blsAddr == *otherAddr
}

func (blsAddr *BLSAddress) Type() AddressType {
	return AddressBLS
}

func (blsAddr *BLSAddress) Bech32(hrp NetworkPrefix) string {
	return bech32String(hrp, blsAddr)
}

func (blsAddr *BLSAddress) String() string {
	return hex.EncodeToString(blsAddr[:])
}

func (blsAddr *BLSAddress) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
		if err := serializer.CheckMinByteLength(BLSAddressSerializedBytesSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid BLS address bytes: %w", err)
		}
		if err := serializer.CheckTypeByte(data, AddressBLS); err != nil {
			return 0, fmt.Errorf("unable to deserialize BLS address: %w", err)
		}
	}
	copy(blsAddr[:], data[serializer.SmallTypeDenotationByteSize:])
	return BLSAddressSerializedBytesSize, nil
}

func (blsAddr *BLSAddress) Serialize(_ serializer.DeSerializationMode, deSeriCtx interface{}) (data []byte, err error) {
	var b [BLSAddressSerializedBytesSize]byte
	b[0] = AddressBLS
	copy(b[serializer.SmallTypeDenotationByteSize:], blsAddr[:])
	return b[:], nil
}

func (blsAddr *BLSAddress) MarshalJSON() ([]byte, error) {
	jBLSAddress := &jsonBLSAddress{}
	jBLSAddress.Address = hex.EncodeToString(blsAddr[:])
	jBLSAddress.Type = AddressBLS
	return json.Marshal(jBLSAddress)
}

func (blsAddr *BLSAddress) UnmarshalJSON(bytes []byte) error {
	jBLSAddress := &jsonBLSAddress{}
	if err := json.Unmarshal(bytes, jBLSAddress); err != nil {
		return err
	}
	seri, err := jBLSAddress.ToSerializable()
	if err != nil {
		return err
	}
	*blsAddr = *seri.(*BLSAddress)
	return nil
}

// BLSAddressFromPubKey returns the address belonging to the given BLS public key.
func BLSAddressFromPubKey(pubKey bls.PublicKey) BLSAddress {
	return blake2b.Sum256(pubKey.Bytes())
}

// jsonBLSAddress defines the json representation of an BLSAddress.
type jsonBLSAddress struct {
	Type    int    `json:"type"`
	Address string `json:"address"`
}

func (j *jsonBLSAddress) ToSerializable() (serializer.Serializable, error) {
	addrBytes, err := hex.DecodeString(j.Address)
	if err != nil {
		return nil, fmt.Errorf("unable to decode address from JSON for BLS address: %w", err)
	}
	if err := serializer.CheckExactByteLength(len(addrBytes), BLSAddressBytesLength); err != nil {
		return nil, fmt.Errorf("unable to decode address from JSON for BLS address: %w", err)
	}
	addr := &BLSAddress{}
	copy(addr[:], addrBytes)
	return addr, nil
}
