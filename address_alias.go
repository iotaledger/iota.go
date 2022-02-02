package iotago

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer/v2"
	"golang.org/x/crypto/blake2b"
)

const (
	// AliasAddressBytesLength is the length of an Alias address.
	AliasAddressBytesLength = 20
	// AliasAddressSerializedBytesSize is the size of a serialized Alias address with its type denoting byte.
	AliasAddressSerializedBytesSize = serializer.SmallTypeDenotationByteSize + AliasAddressBytesLength
)

var (
	emptyAliasAddress = [20]byte{}
)

// ParseAliasAddressFromHexString parses the given hex string into an AliasAddress.
func ParseAliasAddressFromHexString(hexAddr string) (*AliasAddress, error) {
	addrBytes, err := hex.DecodeString(hexAddr)
	if err != nil {
		return nil, err
	}
	addr := &AliasAddress{}
	copy(addr[:], addrBytes)
	return addr, nil
}

// MustParseAliasAddressFromHexString parses the given hex string into an AliasAddress.
// It panics if the hex address is invalid.
func MustParseAliasAddressFromHexString(hexAddr string) *AliasAddress {
	addr, err := ParseAliasAddressFromHexString(hexAddr)
	if err != nil {
		panic(err)
	}
	return addr
}

// AliasAddress defines an Alias address.
// An AliasAddress is the Blake2b-160 hash of the OutputID which created it.
type AliasAddress [AliasAddressBytesLength]byte

func (aliasAddr *AliasAddress) Clone() Address {
	cpy := &AliasAddress{}
	copy(cpy[:], aliasAddr[:])
	return cpy
}

func (aliasAddr *AliasAddress) VByteCost(costStruct *RentStructure, _ VByteCostFunc) uint64 {
	return costStruct.VBFactorData.Multiply(serializer.SmallTypeDenotationByteSize + AliasAddressBytesLength)
}

func (aliasAddr *AliasAddress) Key() string {
	return string(append([]byte{byte(AddressAlias)}, (*aliasAddr)[:]...))
}

func (aliasAddr *AliasAddress) Chain() ChainID {
	return AliasID(*aliasAddr)
}

func (aliasAddr *AliasAddress) AliasID() AliasID {
	return AliasID(*aliasAddr)
}

func (aliasAddr *AliasAddress) Equal(other Address) bool {
	otherAddr, is := other.(*AliasAddress)
	if !is {
		return false
	}
	return *aliasAddr == *otherAddr
}

func (aliasAddr *AliasAddress) Type() AddressType {
	return AddressAlias
}

func (aliasAddr *AliasAddress) Bech32(hrp NetworkPrefix) string {
	return bech32String(hrp, aliasAddr)
}

func (aliasAddr *AliasAddress) String() string {
	return hex.EncodeToString(aliasAddr[:])
}

func (aliasAddr *AliasAddress) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
		if err := serializer.CheckMinByteLength(AliasAddressSerializedBytesSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid alias address bytes: %w", err)
		}
		if err := serializer.CheckTypeByte(data, byte(AddressAlias)); err != nil {
			return 0, fmt.Errorf("unable to deserialize alias address: %w", err)
		}
	}
	copy(aliasAddr[:], data[serializer.SmallTypeDenotationByteSize:])
	return AliasAddressSerializedBytesSize, nil
}

func (aliasAddr *AliasAddress) Serialize(_ serializer.DeSerializationMode, deSeriCtx interface{}) (data []byte, err error) {
	var b [AliasAddressSerializedBytesSize]byte
	b[0] = byte(AddressAlias)
	copy(b[serializer.SmallTypeDenotationByteSize:], aliasAddr[:])
	return b[:], nil
}

func (aliasAddr *AliasAddress) MarshalJSON() ([]byte, error) {
	jAliasAddress := &jsonAliasAddress{}
	jAliasAddress.Address = hex.EncodeToString(aliasAddr[:])
	jAliasAddress.Type = int(AddressAlias)
	return json.Marshal(jAliasAddress)
}

func (aliasAddr *AliasAddress) UnmarshalJSON(bytes []byte) error {
	jAliasAddress := &jsonAliasAddress{}
	if err := json.Unmarshal(bytes, jAliasAddress); err != nil {
		return err
	}
	seri, err := jAliasAddress.ToSerializable()
	if err != nil {
		return err
	}
	*aliasAddr = *seri.(*AliasAddress)
	return nil
}

// AliasAddressFromOutputID returns the alias address computed from a given OutputID.
func AliasAddressFromOutputID(outputID OutputID) AliasAddress {
	// TODO: maybe use pkg with Sum160 exposed
	blake2b160, _ := blake2b.New(20, nil)
	var aliasAddress AliasAddress
	if _, err := blake2b160.Write(outputID[:]); err != nil {
		panic(err)
	}
	copy(aliasAddress[:], blake2b160.Sum(nil))
	return aliasAddress
}

// jsonAliasAddress defines the json representation of an AliasAddress.
type jsonAliasAddress struct {
	Type    int    `json:"type"`
	Address string `json:"address"`
}

func (j *jsonAliasAddress) ToSerializable() (serializer.Serializable, error) {
	addrBytes, err := hex.DecodeString(j.Address)
	if err != nil {
		return nil, fmt.Errorf("unable to decode address from JSON for alias address: %w", err)
	}
	if err := serializer.CheckExactByteLength(len(addrBytes), AliasAddressBytesLength); err != nil {
		return nil, fmt.Errorf("unable to decode address from JSON for alias address: %w", err)
	}
	addr := &AliasAddress{}
	copy(addr[:], addrBytes)
	return addr, nil
}
