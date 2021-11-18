package iotago

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/iotaledger/hive.go/serializer"
	"golang.org/x/crypto/blake2b"
)

const (
	// NFTAddressBytesLength is the length of an NFT address.
	NFTAddressBytesLength = 20
	// NFTAddressSerializedBytesSize is the size of a serialized NFT address with its type denoting byte.
	NFTAddressSerializedBytesSize = serializer.SmallTypeDenotationByteSize + NFTAddressBytesLength
)

var (
	emptyNFTAddress = [20]byte{}
)

// ParseNFTAddressFromHexString parses the given hex string into an NFTAddress.
func ParseNFTAddressFromHexString(hexAddr string) (*NFTAddress, error) {
	addrBytes, err := hex.DecodeString(hexAddr)
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
// An NFTAddress is the Blake2b-160 hash of the OutputID which created it.
type NFTAddress [NFTAddressBytesLength]byte

func (nftAddr *NFTAddress) VByteCost(costStruct *RentStructure, _ VByteCostFunc) uint64 {
	return costStruct.VBFactorKey.With(costStruct.VBFactorData).Multiply(serializer.SmallTypeDenotationByteSize + NFTAddressBytesLength)
}

func (nftAddr *NFTAddress) Key() string {
	return string(append([]byte{AddressNFT}, (*nftAddr)[:]...))
}

func (nftAddr *NFTAddress) Chain() ChainID {
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
	return hex.EncodeToString(nftAddr[:])
}

func (nftAddr *NFTAddress) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	if deSeriMode.HasMode(serializer.DeSeriModePerformValidation) {
		if err := serializer.CheckMinByteLength(NFTAddressSerializedBytesSize, len(data)); err != nil {
			return 0, fmt.Errorf("invalid nft address bytes: %w", err)
		}
		if err := serializer.CheckTypeByte(data, AddressNFT); err != nil {
			return 0, fmt.Errorf("unable to deserialize nft address: %w", err)
		}
	}
	copy(nftAddr[:], data[serializer.SmallTypeDenotationByteSize:])
	return NFTAddressSerializedBytesSize, nil
}

func (nftAddr *NFTAddress) Serialize(_ serializer.DeSerializationMode, deSeriCtx interface{}) (data []byte, err error) {
	var b [NFTAddressSerializedBytesSize]byte
	b[0] = AddressNFT
	copy(b[serializer.SmallTypeDenotationByteSize:], nftAddr[:])
	return b[:], nil
}

func (nftAddr *NFTAddress) MarshalJSON() ([]byte, error) {
	jNFTAddress := &jsonNFTAddress{}
	jNFTAddress.Address = hex.EncodeToString(nftAddr[:])
	jNFTAddress.Type = AddressNFT
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
	// TODO: maybe use pkg with Sum160 exposed
	blake2b160, _ := blake2b.New(20, nil)
	var nftAddress NFTAddress
	copy(nftAddress[:], blake2b160.Sum(outputID[:]))
	return nftAddress
}

// jsonNFTAddress defines the json representation of an NFTAddress.
type jsonNFTAddress struct {
	Type    int    `json:"type"`
	Address string `json:"address"`
}

func (j *jsonNFTAddress) ToSerializable() (serializer.Serializable, error) {
	addrBytes, err := hex.DecodeString(j.Address)
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
