package iotago

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/serializer/v2"
)

var (
	// ErrMissingProtocolParas is returned when ProtocolParameters are missing for operations which require them.
	ErrMissingProtocolParas = errors.New("missing protocol parameters")
)

var (
	protocolParametersRentStructureGuard = &serializer.SerializableGuard{
		ReadGuard: func(ty uint32) (serializer.Serializable, error) {
			return &RentStructure{}, nil
		},
		WriteGuard: func(seri serializer.Serializable) error {
			if seri == nil {
				return fmt.Errorf("%w: because nil", ErrTypeIsNotSupportedRentStructure)
			}

			return nil
		},
	}
)

// NetworkID defines the ID of the network on which entities operate on.
type NetworkID = uint64

// NetworkIDFromString returns the network ID string's numerical representation.
func NetworkIDFromString(networkIDStr string) NetworkID {
	networkIDBlakeHash := blake2b.Sum256([]byte(networkIDStr))
	return binary.LittleEndian.Uint64(networkIDBlakeHash[:])
}

// ProtocolParameters defines the parameters of the protocol.
type ProtocolParameters struct {
	// The version of the protocol running.
	Version byte
	// The human friendly name of the network.
	NetworkName string
	// The HRP prefix used for Bech32 addresses in the network.
	Bech32HRP NetworkPrefix
	// The minimum pow score of the network.
	MinPoWScore uint32
	// The below max depth parameter of the network.
	BelowMaxDepth uint8
	// The rent structure used by given node/network.
	RentStructure RentStructure
	// TokenSupply defines the current token supply on the network.
	TokenSupply uint64
}

func (p *ProtocolParameters) Deserialize(data []byte, deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) (int, error) {
	var bech32HRP string
	var rentStructure *RentStructure
	return serializer.NewDeserializer(data).
		ReadByte(&p.Version, func(err error) error {
			return fmt.Errorf("%w: unable to deserialize version within protocol parameters", err)
		}).
		ReadString(&p.NetworkName, serializer.SeriLengthPrefixTypeAsByte, func(err error) error {
			return fmt.Errorf("%w: unable to deserialize network name within protocol parameters", err)
		}).
		ReadString(&bech32HRP, serializer.SeriLengthPrefixTypeAsByte, func(err error) error {
			return fmt.Errorf("%w: unable to deserialize Bech32HRP prefix within protocol parameters", err)
		}).
		Do(func() {
			p.Bech32HRP = NetworkPrefix(bech32HRP)
		}).
		ReadNum(&p.MinPoWScore, func(err error) error {
			return fmt.Errorf("%w: unable to deserialize minimum pow score within protocol parameters", err)
		}).
		ReadNum(&p.BelowMaxDepth, func(err error) error {
			return fmt.Errorf("%w: unable to deserialize below max depth within protocol parameters", err)
		}).
		ReadObject(&rentStructure, deSeriMode, deSeriCtx, serializer.TypeDenotationNone, protocolParametersRentStructureGuard.ReadGuard, func(err error) error {
			return fmt.Errorf("%w: unable to deserialize rent structure within protocol parameters", err)
		}).
		Do(func() {
			p.RentStructure = *rentStructure
		}).
		ReadNum(&p.TokenSupply, func(err error) error {
			return fmt.Errorf("%w: unable to deserialize token supply within protocol parameters", err)
		}).
		Done()
}

func (p *ProtocolParameters) Serialize(deSeriMode serializer.DeSerializationMode, deSeriCtx interface{}) ([]byte, error) {
	return serializer.NewSerializer().
		WriteByte(p.Version, func(err error) error {
			return fmt.Errorf("%w: unable to serialize version within protocol parameters", err)
		}).
		WriteString(p.NetworkName, serializer.SeriLengthPrefixTypeAsByte, func(err error) error {
			return fmt.Errorf("%w: unable to serialize network name within protocol parameters", err)
		}).
		WriteString(string(p.Bech32HRP), serializer.SeriLengthPrefixTypeAsByte, func(err error) error {
			return fmt.Errorf("%w: unable to serialize Bech32HRP prefix within protocol parameters", err)
		}).
		WriteNum(p.MinPoWScore, func(err error) error {
			return fmt.Errorf("%w: unable to serialize minimum pow score within protocol parameters", err)
		}).
		WriteNum(p.BelowMaxDepth, func(err error) error {
			return fmt.Errorf("%w: unable to serialize below max depth within protocol parameters", err)
		}).
		WriteObject(&p.RentStructure, deSeriMode, deSeriCtx, protocolParametersRentStructureGuard.WriteGuard, func(err error) error {
			return fmt.Errorf("%w: unable to serialize rent structure within protocol parameters", err)
		}).
		WriteNum(p.TokenSupply, func(err error) error {
			return fmt.Errorf("%w: unable to serialize token supply within protocol parameters", err)
		}).
		Serialize()
}

func (p ProtocolParameters) NetworkID() NetworkID {
	return NetworkIDFromString(p.NetworkName)
}

type jsonProtocolParameters struct {
	Version       byte          `json:"version"`
	NetworkName   string        `json:"networkName"`
	Bech32HRP     NetworkPrefix `json:"bech32HRP"`
	MinPoWScore   uint32        `json:"minPoWScore"`
	BelowMaxDepth uint8         `json:"belowMaxDepth"`
	RentStructure RentStructure `json:"rentStructure"`
	TokenSupply   string        `json:"tokenSupply"`
}

func (p ProtocolParameters) MarshalJSON() ([]byte, error) {
	return json.Marshal(&jsonProtocolParameters{
		Version:       p.Version,
		NetworkName:   p.NetworkName,
		Bech32HRP:     p.Bech32HRP,
		MinPoWScore:   p.MinPoWScore,
		BelowMaxDepth: p.BelowMaxDepth,
		RentStructure: p.RentStructure,
		TokenSupply:   EncodeUint64(p.TokenSupply),
	})
}

func (p *ProtocolParameters) UnmarshalJSON(data []byte) error {
	aux := &jsonProtocolParameters{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	supply, err := DecodeUint64(aux.TokenSupply)
	if err != nil {
		return err
	}

	p.Version = aux.Version
	p.NetworkName = aux.NetworkName
	p.Bech32HRP = aux.Bech32HRP
	p.MinPoWScore = aux.MinPoWScore
	p.BelowMaxDepth = aux.BelowMaxDepth
	p.RentStructure = aux.RentStructure
	p.TokenSupply = supply

	return nil
}
