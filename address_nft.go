package iotago

import (
	"encoding/json"
	"fmt"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/serializer/v2"
)

const (
	// NFTAddressBytesLength is the length of an NFT address.
	NFTAddressBytesLength = blake2b.Size256
	// NFTAddressSerializedBytesSize is the size of a serialized NFT address with its type denoting byte.
	NFTAddressSerializedBytesSize = serializer.SmallTypeDenotationByteSize + NFTAddressBytesLength
)

// ParseNFTAddressFromHexString parses the given hex string into an NFTAddress.
func ParseNFTAddressFromHexString(hexAddr string) (*NFTAddress, error) {
	addrBytes, err := DecodeHex(hexAddr)
	if err != nil {
		return nil, err
	}
	addr := &NFTAddress{}
	copy(addr[:], addrBytes)

	return addr, nil
}

// MustParseNFTAddressFromHexString parses the given hex string into an NFTAddress.
// It panics if the hex address is invalid.
func MustParseNFTAddressFromHexString(hexAddr string) *NFTAddress {
	addr, err := ParseNFTAddressFromHexString(hexAddr)
	if err != nil {
		panic(err)
	}

	return addr
}

// NFTAddress defines an NFT address.
// An NFTAddress is the Blake2b-256 hash of the OutputID which created it.
type NFTAddress [NFTAddressBytesLength]byte

func (nftAddr *NFTAddress) Clone() Address {
	cpy := &NFTAddress{}
	copy(cpy[:], nftAddr[:])

	return cpy
}

func (nftAddr *NFTAddress) VBytes(rentStruct *RentStructure, _ VBytesFunc) uint64 {
	return rentStruct.VBFactorData.Multiply(NFTAddressSerializedBytesSize)
}

func (nftAddr *NFTAddress) Key() string {
	return string(append([]byte{byte(AddressNFT)}, (*nftAddr)[:]...))
}

func (nftAddr *NFTAddress) Chain() ChainID {
	return NFTID(*nftAddr)
}

func (nftAddr *NFTAddress) NFTID() NFTID {
	return NFTID(*nftAddr)
}

func (nftAddr *NFTAddress) Equal(other Address) bool {
	otherAddr, is := other.(*NFTAddress)
	if !is {
		return false
	}

	return *nftAddr == *otherAddr
}

func (nftAddr *NFTAddress) Type() AddressType {
	return AddressNFT
}

func (nftAddr *NFTAddress) Bech32(hrp NetworkPrefix) string {
	return bech32String(hrp, nftAddr)
}

func (nftAddr *NFTAddress) String() string {
	return EncodeHex(nftAddr[:])
}

func (nftAddr *NFTAddress) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
		if err := serializer.CheckMinByteLength(NFTAddressSerializedBytesSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid NFT address bytes: %w", err)
		}
		if err := serializer.CheckTypeByte(data, byte(AddressNFT)); err != nil {
			return 0, fmt.Errorf("unable to deserialize NFT address: %w", err)
		}
	}
	copy(nftAddr[:], data[serializer.SmallTypeDenotationByteSize:])

	return NFTAddressSerializedBytesSize, nil
}

func (nftAddr *NFTAddress) Serialize(_ serializer.DeSerializationMode, deSeriCtx interface{}) (data []byte, err error) {
	var b [NFTAddressSerializedBytesSize]byte
	b[0] = byte(AddressNFT)
	copy(b[serializer.SmallTypeDenotationByteSize:], nftAddr[:])

	return b[:], nil
}

func (nftAddr *NFTAddress) Size() int {
	return NFTAddressSerializedBytesSize
}

func (nftAddr *NFTAddress) MarshalJSON() ([]byte, error) {
	jNFTAddress := &jsonNFTAddress{}
	jNFTAddress.NFTId = EncodeHex(nftAddr[:])
	jNFTAddress.Type = int(AddressNFT)

	return json.Marshal(jNFTAddress)
}

func (nftAddr *NFTAddress) UnmarshalJSON(bytes []byte) error {
	jNFTAddress := &jsonNFTAddress{}
	if err := json.Unmarshal(bytes, jNFTAddress); err != nil {
		return err
	}
	seri, err := jNFTAddress.ToSerializable()
	if err != nil {
		return err
	}
	*nftAddr = *seri.(*NFTAddress)

	return nil
}

// NFTAddressFromOutputID returns the NFT address computed from a given OutputID.
func NFTAddressFromOutputID(outputID OutputID) NFTAddress {
	return blake2b.Sum256(outputID[:])
}

// jsonNFTAddress defines the json representation of an NFTAddress.
type jsonNFTAddress struct {
	Type  int    `json:"type"`
	NFTId string `json:"nftId"`
}

func (j *jsonNFTAddress) ToSerializable() (serializer.Serializable, error) {
	addrBytes, err := DecodeHex(j.NFTId)
	if err != nil {
		return nil, fmt.Errorf("unable to decode address from JSON for NFT address: %w", err)
	}
	if err := serializer.CheckExactByteLength(len(addrBytes), NFTAddressBytesLength); err != nil {
		return nil, fmt.Errorf("unable to decode address from JSON for NFT address: %w", err)
	}
	addr := &NFTAddress{}
	copy(addr[:], addrBytes)

	return addr, nil
}
