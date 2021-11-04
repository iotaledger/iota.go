package iotago

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/iotaledger/hive.go/serializer"
)

const (
	// MinNativeTokenCountPerOutput min number of different native tokens that can reside in one output.
	MinNativeTokenCountPerOutput = 0
	// MaxNativeTokenCountPerOutput max number of different native tokens that can reside in one output.
	MaxNativeTokenCountPerOutput = 256

	// MaxNativeTokensCount is the max number of native tokens which can occur in a transaction (sum input/output side).
	MaxNativeTokensCount = 256

	TokenTagLength = 12

	// FoundryIDLength is the byte length of a FoundryID consisting out of the alias address, serial number and token scheme.
	FoundryIDLength = AliasAddressSerializedBytesSize + serializer.UInt32ByteSize + serializer.OneByte

	// NativeTokenIDLength is the byte length of a NativeTokenID consisting out of the FoundryID plus TokenTag.
	NativeTokenIDLength = FoundryIDLength + TokenTagLength
)

var (
	nativeTokensArrayRules = &serializer.ArrayRules{
		Min:            MinNativeTokenCountPerOutput,
		Max:            MaxNativeTokenCountPerOutput,
		ValidationMode: serializer.ArrayValidationModeNoDuplicates | serializer.ArrayValidationModeLexicalOrdering,
	}
)

// NativeTokenID is an identifier which uniquely identifies a NativeToken.
type NativeTokenID = [NativeTokenIDLength]byte

// NativeToken represents a token natively
type NativeToken struct {
	NFTID  NativeTokenID
	Amount *big.Int
}

func (n *NativeToken) Deserialize(data []byte, _ serializer.DeSerializationMode) (int, error) {
	return serializer.NewDeserializer(data).
		ReadArrayOf38Bytes(&n.NFTID, func(err error) error {
			return fmt.Errorf("unable to deserialize ID for native token: %w", err)
		}).
		ReadUint256(n.Amount, func(err error) error {
			return fmt.Errorf("unable to deserialize amount for native token: %w", err)
		}).
		Done()
}

func (n *NativeToken) Serialize(_ serializer.DeSerializationMode) ([]byte, error) {
	return serializer.NewSerializer().
		WriteBytes(n.NFTID[:], func(err error) error {
			return fmt.Errorf("unable to serialize native token ID: %w", err)
		}).
		WriteUint256(n.Amount, func(err error) error {
			return fmt.Errorf("unable to serialize native token amount: %w", err)
		}).
		Serialize()
}

func (n *NativeToken) MarshalJSON() ([]byte, error) {
	jNativeToken := &jsonNativeToken{}
	jNativeToken.ID = hex.EncodeToString(n.NFTID[:])
	jNativeToken.Amount = n.Amount.String()
	return json.Marshal(jNativeToken)
}

func (n *NativeToken) UnmarshalJSON(bytes []byte) error {
	jNativeToken := &jsonNativeToken{}
	if err := json.Unmarshal(bytes, jNativeToken); err != nil {
		return err
	}
	seri, err := jNativeToken.ToSerializable()
	if err != nil {
		return err
	}
	*n = *seri.(*NativeToken)
	return nil
}

// jsonNativeToken defines the json representation of a NativeToken.
type jsonNativeToken struct {
	ID     string `json:"id"`
	Amount string `json:"amount"`
}

func (j *jsonNativeToken) ToSerializable() (serializer.Serializable, error) {
	n := &NativeToken{}

	nftIDBytes, err := hex.DecodeString(j.ID)
	if err != nil {
		return nil, err
	}
	copy(n.NFTID[:], nftIDBytes)

	var ok bool
	n.Amount, ok = new(big.Int).SetString(j.Amount, 10)
	if !ok {
		return nil, fmt.Errorf("%w: amount field of native token '%s'", ErrDecodeJSONUint256Str, j.Amount)
	}

	return n, nil
}
